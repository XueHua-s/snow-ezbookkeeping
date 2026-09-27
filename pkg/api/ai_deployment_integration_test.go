package api

import (
	"bytes"
	_ "embed"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

//go:embed testdata/ai-receipt.png
var aiDeploymentReceiptImage []byte

// Run explicitly against a configured deployment; only recognition and embedding cache operations are performed.
func TestAIDeploymentIntegration(t *testing.T) {
	if os.Getenv("EBK_AI_DEPLOYMENT_CHECK") != "1" {
		t.Skip("deployment integration is opt-in")
	}
	workDir := os.Getenv("EBK_AI_DEPLOYMENT_WORK_DIR")
	if workDir == "" {
		workDir = "/ezbookkeeping"
	}
	configPath := os.Getenv("EBK_AI_DEPLOYMENT_CONFIG")
	if configPath == "" {
		configPath = filepath.Join(workDir, "conf", "ezbookkeeping.ini")
	}
	t.Chdir(workDir)
	t.Setenv("EBK_WORK_DIR", workDir)
	config, err := settings.LoadConfiguration(configPath)
	require.NoError(t, err)
	require.NotEmpty(t, config.AIAssistantLLMConfig.OpenAIModelID)
	require.Equal(t, settings.OpenAILLMProvider, config.AIAssistantLLMConfig.LLMProvider)
	require.NotEmpty(t, config.GetAIAssistantEmbeddingLLMConfig().OpenAIEmbeddingModelID)
	require.NotEmpty(t, config.AIAssistantLLMConfig.OpenAIAPIKey)
	require.NotEmpty(t, config.GetAIAssistantEmbeddingLLMConfig().OpenAIAPIKey)
	settings.SetCurrentConfig(config)
	require.NoError(t, datastore.InitializeDataStore(config))
	require.NoError(t, llm.InitializeLargeLanguageModelProvider(config))
	uid, err := strconv.ParseInt(os.Getenv("EBK_AI_DEPLOYMENT_UID"), 10, 64)
	require.NoError(t, err)
	gin.SetMode(gin.TestMode)
	contextFor := func(body io.Reader, contentType string) (*core.WebContext, *httptest.ResponseRecorder) {
		rec := httptest.NewRecorder()
		gc, _ := gin.CreateTestContext(rec)
		gc.Request = httptest.NewRequest("POST", "/", body)
		gc.Request.Header.Set("Content-Type", contentType)
		gc.Request.Header.Set("Accept-Language", "zh-CN")
		gc.Request.Header.Set("X-Timezone-Offset", "480")
		gc.Request.Header.Set("X-Timezone-Name", "Asia/Shanghai")
		c := &core.WebContext{Context: gc}
		c.SetTokenClaims(&core.UserTokenClaims{Uid: uid})
		return c, rec
	}
	t.Run("receipt", func(t *testing.T) {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("image", "test-receipt.png")
		require.NoError(t, err)
		_, err = part.Write(aiDeploymentReceiptImage)
		require.NoError(t, err)
		require.NoError(t, writer.Close())
		c, _ := contextFor(&body, writer.FormDataContentType())
		result, apiErr := LargeLanguageModels.RecognizeReceiptImageHandler(c)
		require.Nil(t, apiErr)
		receipt, ok := result.(*models.RecognizedTransactionResponse)
		require.True(t, ok)
		require.Equal(t, int64(1234), receipt.SourceAmount)
		t.Log("Receipt recognition returned the expected CNY 12.34; no transaction created")
	})
	t.Run("assistant", func(t *testing.T) {
		c, rec := contextFor(strings.NewReader(`{"mode":"chat","message":"这是连通性检查，请仅回复：连接正常。"}`), "application/json")
		apiErr := LargeLanguageModels.AssistantChatStreamHandler(c)
		require.Nil(t, apiErr)
		require.True(t, strings.Contains(rec.Body.String(), `"type":"done"`), "assistant stream must finish")
		require.True(t, strings.Contains(rec.Body.String(), `"type":"reply_delta"`), "assistant must emit a reply")
		t.Log("Assistant completed embedding retrieval and streamed a reply")
	})
}
