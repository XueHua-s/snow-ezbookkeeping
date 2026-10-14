package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

func TestGetAIAssistantPreferredReplyLanguage(t *testing.T) {
	assert.Equal(t, "Simplified Chinese", getAIAssistantPreferredReplyLanguage("zh-CN"))
	assert.Equal(t, "Traditional Chinese", getAIAssistantPreferredReplyLanguage("zh-Hant"))
	assert.Equal(t, "Japanese", getAIAssistantPreferredReplyLanguage("ja-JP"))
	assert.Equal(t, "English", getAIAssistantPreferredReplyLanguage("en-US"))
	assert.Equal(t, "English", getAIAssistantPreferredReplyLanguage(""))
}

func TestBuildAIAssistantUserPrompt_SummaryUsesChinesePromptForChineseLocale(t *testing.T) {
	api := &LargeLanguageModelsApi{}
	request := &models.AIAssistantChatRequest{
		Message: "重点看餐饮支出",
	}

	prompt := api.buildAIAssistantUserPrompt(request, models.AIAssistantModeSummary, "zh-CN")
	assert.Contains(t, prompt, "请基于我的账单数据，用中文给我一份个人财务总结")
	assert.Contains(t, prompt, "额外关注点：重点看餐饮支出")
}

func TestBuildAIAssistantUserPrompt_SummaryUsesEnglishPromptForEnglishLocale(t *testing.T) {
	api := &LargeLanguageModelsApi{}
	request := &models.AIAssistantChatRequest{
		Message: "Focus on food spending",
	}

	prompt := api.buildAIAssistantUserPrompt(request, models.AIAssistantModeSummary, "en-US")
	assert.Contains(t, prompt, "Please provide a personal finance summary")
	assert.Contains(t, prompt, "Additional focus: Focus on food spending")
}

func TestBuildAIAssistantEmbeddingQueryText_SummaryUsesChineseQueryForChineseLocale(t *testing.T) {
	api := &LargeLanguageModelsApi{}
	request := &models.AIAssistantChatRequest{
		Message: "重点看餐饮支出",
	}

	queryText := api.buildAIAssistantEmbeddingQueryText(request, models.AIAssistantModeSummary, "zh-CN")
	assert.Equal(t, "个人财务总结与记账建议，重点关注：重点看餐饮支出", queryText)
}

func TestBuildAIAssistantEmbeddingQueryText_SummaryUsesEnglishQueryForEnglishLocale(t *testing.T) {
	api := &LargeLanguageModelsApi{}
	request := &models.AIAssistantChatRequest{}

	queryText := api.buildAIAssistantEmbeddingQueryText(request, models.AIAssistantModeSummary, "en-US")
	assert.Equal(t, "summarize recent personal finance trends, spending, risks, and bookkeeping suggestions", queryText)
}

func TestGetOpenAIEmbeddings_RespectsTwentyItemLimit(t *testing.T) {
	sizes := []int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/compatible-mode/v1/embeddings", r.URL.Path)
		assert.Equal(t, "Bearer embedding-key", r.Header.Get("Authorization"))
		var request openAIEmbeddingsRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		sizes = append(sizes, len(request.Input))
		assert.Equal(t, "qwen3.7-text-embedding-flash", request.Model)
		if len(request.Input) > 20 {
			w.WriteHeader(400)
			return
		}
		response := &openAIEmbeddingsResponse{}
		for i := len(request.Input) - 1; i >= 0; i-- {
			v, _ := strconv.Atoi(request.Input[i])
			response.Data = append(response.Data, &openAIEmbeddingsResponseItem{Index: i, Embedding: []float64{float64(v), 1}})
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	inputs := make([]string, 41)
	for i := range inputs {
		inputs[i] = strconv.Itoa(i)
	}
	api := &LargeLanguageModelsApi{}
	vectors, err := api.getOpenAIEmbeddings(core.NewNullContext(), 1, &settings.LLMConfig{
		OpenAIBaseURL: server.URL + "/compatible-mode/v1/", OpenAIAPIKey: "embedding-key",
		OpenAIEmbeddingModelID: "qwen3.7-text-embedding-flash", LargeLanguageModelAPIRequestTimeout: 5000, LargeLanguageModelAPIProxy: "none",
	}, inputs)
	require.NoError(t, err)
	require.Len(t, vectors, 41)
	assert.Equal(t, []int{20, 20, 1}, sizes)
	for i, vector := range vectors {
		assert.Equal(t, float64(i), vector[0])
	}
}

func TestGetOpenAIEmbeddings_RejectsMalformedResponses(t *testing.T) {
	for name, body := range map[string]string{
		"null item":            `{"data":[null,{"index":1,"embedding":[1]}]}`,
		"duplicate index":      `{"data":[{"index":0,"embedding":[1]},{"index":0,"embedding":[2]}]}`,
		"out of range index":   `{"data":[{"index":0,"embedding":[1]},{"index":2,"embedding":[2]}]}`,
		"negative index":       `{"data":[{"index":-1,"embedding":[1]},{"index":1,"embedding":[2]}]}`,
		"empty vector":         `{"data":[{"index":0,"embedding":[]},{"index":1,"embedding":[2]}]}`,
		"different dimensions": `{"data":[{"index":0,"embedding":[1]},{"index":1,"embedding":[2,3]}]}`,
		"wrong count":          `{"data":[{"index":0,"embedding":[1]}]}`,
		"invalid JSON":         `{`,
	} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) }))
			defer server.Close()
			api := &LargeLanguageModelsApi{}
			vectors, err := api.getOpenAIEmbeddings(core.NewNullContext(), 1, &settings.LLMConfig{OpenAIBaseURL: server.URL, OpenAIAPIKey: "test", OpenAIEmbeddingModelID: "test", LargeLanguageModelAPIRequestTimeout: 5000, LargeLanguageModelAPIProxy: "none"}, []string{"one", "two"})
			require.Error(t, err)
			require.Nil(t, vectors)
		})
	}
}

func TestGetOpenAIEmbeddings_RejectsDimensionChangeAcrossBatches(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var request openAIEmbeddingsRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		response := &openAIEmbeddingsResponse{}
		for i := range request.Input {
			response.Data = append(response.Data, &openAIEmbeddingsResponseItem{Index: i, Embedding: make([]float64, calls)})
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	api := &LargeLanguageModelsApi{}
	vectors, err := api.getOpenAIEmbeddings(core.NewNullContext(), 1, &settings.LLMConfig{OpenAIBaseURL: server.URL, OpenAIAPIKey: "test", OpenAIEmbeddingModelID: "test", LargeLanguageModelAPIRequestTimeout: 5000, LargeLanguageModelAPIProxy: "none"}, make([]string, 21))
	require.Error(t, err)
	require.Nil(t, vectors)
}
