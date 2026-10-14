package llm

import (
	"encoding/json"
	"errors"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
	"github.com/stretchr/testify/assert"
)

type fallbackTestProvider struct {
	response       *data.LargeLanguageModelTextualResponse
	err            error
	calls          int
	lastRequest    *data.LargeLanguageModelRequest
	streamResponse *data.LargeLanguageModelStreamResponse
	streamErr      error
	streamDeltas   []string
	streamCalls    int
}

func (p *fallbackTestProvider) GetJsonResponse(c core.Context, uid int64, request *data.LargeLanguageModelRequest) (*data.LargeLanguageModelTextualResponse, error) {
	p.calls++
	p.lastRequest = request
	return p.response, p.err
}

func (p *fallbackTestProvider) StreamTextResponse(c core.Context, uid int64, request *data.LargeLanguageModelRequest, callback data.LargeLanguageModelStreamCallback) (*data.LargeLanguageModelStreamResponse, error) {
	p.streamCalls++
	for _, delta := range p.streamDeltas {
		if callback != nil {
			callback(data.LARGE_LANGUAGE_MODEL_STREAM_DELTA_TYPE_REPLY, delta)
		}
	}
	return p.streamResponse, p.streamErr
}

func TestGetJsonResponseByReceiptImageRecognitionModel_UsesFallbackAfterPrimaryFailure(t *testing.T) {
	primary := &fallbackTestProvider{err: errors.New("primary unavailable")}
	fallback := &fallbackTestProvider{response: &data.LargeLanguageModelTextualResponse{Content: "fallback"}}
	container := &LargeLanguageModelProviderContainer{
		receiptImageRecognitionProvider: &fallbackLargeLanguageModelProvider{
			usage:            "receipt image recognition",
			primaryProvider:  primary,
			fallbackProvider: fallback,
		},
	}
	request := &data.LargeLanguageModelRequest{}

	response, err := container.GetJsonResponseByReceiptImageRecognitionModel(nil, 1, request)

	assert.NoError(t, err)
	assert.Equal(t, "fallback", response.Content)
	assert.Equal(t, 1, primary.calls)
	assert.Equal(t, 1, fallback.calls)
	assert.Same(t, request, primary.lastRequest)
	assert.Same(t, request, fallback.lastRequest)
}

func TestGetJsonResponseByAIAssistantModel_ReturnsFallbackFailure(t *testing.T) {
	primary := &fallbackTestProvider{err: errors.New("primary unavailable")}
	fallbackErr := errors.New("fallback unavailable")
	fallback := &fallbackTestProvider{err: fallbackErr}
	container := &LargeLanguageModelProviderContainer{
		aiAssistantProvider: &fallbackLargeLanguageModelProvider{
			usage:            "ai assistant",
			primaryProvider:  primary,
			fallbackProvider: fallback,
		},
	}

	response, err := container.GetJsonResponseByAIAssistantModel(nil, 1, &data.LargeLanguageModelRequest{})

	assert.Nil(t, response)
	assert.ErrorIs(t, err, fallbackErr)
	assert.Equal(t, 1, primary.calls)
	assert.Equal(t, 1, fallback.calls)
}

func TestGetJsonResponseByAIAssistantModel_DoesNotUseFallbackAfterPrimarySuccess(t *testing.T) {
	primary := &fallbackTestProvider{response: &data.LargeLanguageModelTextualResponse{Content: "primary"}}
	fallback := &fallbackTestProvider{response: &data.LargeLanguageModelTextualResponse{Content: "fallback"}}
	container := &LargeLanguageModelProviderContainer{
		aiAssistantProvider: &fallbackLargeLanguageModelProvider{
			usage:            "ai assistant",
			primaryProvider:  primary,
			fallbackProvider: fallback,
		},
	}

	response, err := container.GetJsonResponseByAIAssistantModel(nil, 1, &data.LargeLanguageModelRequest{})

	assert.NoError(t, err)
	assert.Equal(t, "primary", response.Content)
	assert.Equal(t, 1, primary.calls)
	assert.Equal(t, 0, fallback.calls)
}

func TestStreamTextResponseByAIAssistantModel_UsesFallbackBeforeFirstDelta(t *testing.T) {
	primary := &fallbackTestProvider{streamErr: errors.New("primary unavailable")}
	fallback := &fallbackTestProvider{
		streamResponse: &data.LargeLanguageModelStreamResponse{Content: "fallback"},
		streamDeltas:   []string{"fallback"},
	}
	container := &LargeLanguageModelProviderContainer{
		aiAssistantProvider: &fallbackLargeLanguageModelProvider{
			usage:            "ai assistant",
			primaryProvider:  primary,
			fallbackProvider: fallback,
		},
	}

	deltas := make([]string, 0, 1)
	response, err := container.StreamTextResponseByAIAssistantModel(nil, 1, &data.LargeLanguageModelRequest{}, func(deltaType data.LargeLanguageModelStreamDeltaType, delta string) {
		deltas = append(deltas, delta)
	})

	assert.NoError(t, err)
	assert.Equal(t, "fallback", response.Content)
	assert.Equal(t, []string{"fallback"}, deltas)
	assert.Equal(t, 1, primary.streamCalls)
	assert.Equal(t, 1, fallback.streamCalls)
}

func TestStreamTextResponseByAIAssistantModel_DoesNotFallbackAfterFirstDelta(t *testing.T) {
	primaryErr := errors.New("primary interrupted")
	primary := &fallbackTestProvider{
		streamErr:    primaryErr,
		streamDeltas: []string{"partial"},
	}
	fallback := &fallbackTestProvider{
		streamResponse: &data.LargeLanguageModelStreamResponse{Content: "fallback"},
	}
	container := &LargeLanguageModelProviderContainer{
		aiAssistantProvider: &fallbackLargeLanguageModelProvider{
			usage:            "ai assistant",
			primaryProvider:  primary,
			fallbackProvider: fallback,
		},
	}

	response, err := container.StreamTextResponseByAIAssistantModel(nil, 1, &data.LargeLanguageModelRequest{}, func(deltaType data.LargeLanguageModelStreamDeltaType, delta string) {})

	assert.Nil(t, response)
	assert.ErrorIs(t, err, primaryErr)
	assert.Equal(t, 1, primary.streamCalls)
	assert.Equal(t, 0, fallback.streamCalls)
}

var _ provider.LargeLanguageModelProvider = (*fallbackTestProvider)(nil)
var _ provider.LargeLanguageModelStreamingProvider = (*fallbackTestProvider)(nil)
var _ provider.LargeLanguageModelProvider = (*fallbackLargeLanguageModelProvider)(nil)

// Exercise the merged provider wiring over HTTP: upstream Responses requests and
// fork Chat Completions fallbacks must coexist without sharing configuration.
func TestInitializeProviders_PreservesResponsesChatFallbackAndStreaming(t *testing.T) {
	original := *Container
	defer func() { *Container = original }()
	calls := make(chan string, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Model  string `json:"model"`
			Stream bool   `json:"stream"`
		}
		if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&body)) {
			w.WriteHeader(400)
			return
		}
		calls <- body.Model + ":" + r.URL.Path
		assert.Equal(t, "Bearer "+body.Model+"-key", r.Header.Get("Authorization"))
		if body.Model == "receipt-primary" {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		if body.Stream {
			assert.Equal(t, "/v1/responses", r.URL.Path)
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = io.WriteString(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"stream reply\"}\n\ndata: {\"type\":\"response.completed\",\"response\":{}}\n\n")
		} else if r.URL.Path == "/v1/responses" {
			_, _ = io.WriteString(w, `{"output":[{"type":"message","content":[{"type":"output_text","text":"{\"reply\":\"ok\"}"}]}]}`)
		} else {
			assert.Equal(t, "/v1/chat/completions", r.URL.Path)
			_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"{\"reply\":\"ok\"}"}}]}`)
		}
	}))
	defer server.Close()
	configFor := func(model string) *settings.LLMConfig {
		return &settings.LLMConfig{LLMProvider: settings.OpenAILLMProvider, OpenAIBaseURL: server.URL + "/v1/", OpenAIModelID: model, OpenAIAPIKey: model + "-key", LargeLanguageModelAPIProxy: "none", LargeLanguageModelAPIRequestTimeout: 5000}
	}
	require.NoError(t, InitializeLargeLanguageModelProvider(&settings.Config{
		EnableAIAssistant:                        true,
		TextRecognitionLLMConfig:                 configFor("text"),
		CodingAssistantLLMConfig:                 configFor("coding"),
		ReceiptImageRecognitionLLMConfig:         configFor("receipt-primary"),
		ReceiptImageRecognitionFallbackLLMConfig: configFor("receipt-fallback"),
		AIAssistantLLMConfig:                     configFor("assistant"),
	}))
	ctx := core.NewNullContext()
	request := &data.LargeLanguageModelRequest{UserPrompt: []byte("Return json")}
	for _, call := range []func(core.Context, int64, *data.LargeLanguageModelRequest) (*data.LargeLanguageModelTextualResponse, error){
		Container.GetJsonResponseByTextRecognitionModel,
		Container.GetJsonResponseByCodingAssistantModel,
		Container.GetJsonResponseByReceiptImageRecognitionModel,
		Container.GetJsonResponseByAIAssistantModel,
	} {
		response, err := call(ctx, 1, request)
		require.NoError(t, err)
		assert.JSONEq(t, `{"reply":"ok"}`, response.Content)
	}
	var reply string
	_, err := Container.StreamTextResponseByAIAssistantModel(ctx, 1, request, func(_ data.LargeLanguageModelStreamDeltaType, delta string) { reply += delta })
	require.NoError(t, err)
	assert.Equal(t, "stream reply", reply)
	var actualCalls []string
	for len(calls) > 0 {
		actualCalls = append(actualCalls, <-calls)
	}
	assert.Equal(t, []string{"text:/v1/responses", "coding:/v1/responses", "receipt-primary:/v1/chat/completions", "receipt-fallback:/v1/chat/completions", "assistant:/v1/chat/completions", "assistant:/v1/responses"}, actualCalls)
	require.NoError(t, InitializeLargeLanguageModelProvider(&settings.Config{}))
	assert.Nil(t, Container.textRecognitionProvider)
	assert.Nil(t, Container.codingAssistantProvider)
	assert.Nil(t, Container.receiptImageRecognitionProvider)
	assert.Nil(t, Container.aiAssistantProvider)
}
