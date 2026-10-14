package api

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

//go:embed testdata/ai-receipt.png
var aiDeploymentReceiptImage []byte

// Run explicitly against an isolated test deployment. These routes only recognize
// the synthetic receipt and read the dedicated test ledger; they never create bills.
func TestAIDeploymentIntegration(t *testing.T) {
	if os.Getenv("EBK_AI_DEPLOYMENT_CHECK") != "1" {
		t.Skip("deployment integration is opt-in")
	}
	require.Equal(t, "1", os.Getenv("EBK_AI_DEPLOYMENT_ISOLATED"), "use a dedicated isolated deployment and test user, never production")
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
	require.True(t, config.EnableAIAssistant, "enable the assistant in the isolated test configuration")
	forceFallback := os.Getenv("EBK_AI_DEPLOYMENT_FORCE_FALLBACK") == "1"
	if forceFallback {
		require.NotNil(t, config.AIAssistantFallbackLLMConfig, "a real fallback provider is required")
		// Fail only the isolated primary endpoint. The configured fallback must
		// perform the actual tool conversation; this is not model certification.
		failedPrimary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		t.Cleanup(failedPrimary.Close)
		config.AIAssistantLLMConfig.OpenAIBaseURL = failedPrimary.URL
		config.AIAssistantLLMConfig.OpenAICompatibleBaseURL = failedPrimary.URL
		config.AIAssistantLLMConfig.LargeLanguageModelAPIProxy = "none"
	}
	settings.SetCurrentConfig(config)
	require.NoError(t, datastore.InitializeDataStore(config))
	require.NoError(t, llm.InitializeLargeLanguageModelProvider(config))
	primary, fallback, err := llm.Container.GetAIAssistantToolProviders()
	require.NoError(t, err, "configured Chat/forced-stream/Responses provider must support tools")
	require.NotNil(t, primary)
	if forceFallback {
		require.NotNil(t, fallback)
	}
	uid, err := strconv.ParseInt(os.Getenv("EBK_AI_DEPLOYMENT_UID"), 10, 64)
	require.NoError(t, err)
	require.Positive(t, uid, "provide a dedicated test-user UID")
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
		for _, stream := range []bool{false, true} {
			name := "json"
			if stream {
				name = "sse"
			}
			t.Run(name, func(t *testing.T) {
				c, rec := contextFor(strings.NewReader(`{"mode":"chat","message":"请调用 list_accounts 和 get_account_balances，列出我的当前测试账户余额，必须查询后再回答。"}`), "application/json")
				if stream {
					apiErr := LargeLanguageModels.AssistantChatStreamHandler(c)
					require.Nil(t, apiErr)
					require.Contains(t, rec.Body.String(), `"type":"done"`, "assistant stream must finish")
					require.Contains(t, rec.Body.String(), `"type":"reply_delta"`, "assistant must emit a reply")
				} else {
					result, apiErr := LargeLanguageModels.AssistantChatHandler(c)
					require.Nil(t, apiErr)
					response, ok := result.(*models.AIAssistantChatResponse)
					require.True(t, ok)
					require.NotEmpty(t, response.Reply)
					body, err := json.Marshal(response)
					require.NoError(t, err)
					_, err = rec.Write(body)
					require.NoError(t, err)
				}
				require.Contains(t, rec.Body.String(), `"tool":"list_accounts"`, "deployment provider must execute account discovery")
				require.Contains(t, rec.Body.String(), `"tool":"get_account_balances"`, "deployment provider must execute balance query")
				require.NotContains(t, rec.Body.String(), "similarityScore")
				t.Log("Assistant completed read-only tools and replied without embedding")
			})
		}
	})
}

// A cache table may still contain rollback data, but both current routes must
// work even when SQLite forbids every read/write of that table.
func TestAIAssistantRetiredCacheNeverAccessed(t *testing.T) {
	a, engine := harnessAPIFixture(t)
	harnessInsert(t, engine, 100, "2026-03-01T00:00:00+08:00", 10000, 30)
	require.NoError(t, engine.Sync2(&models.AIAssistantEmbedding{}))
	legacy := &models.AIAssistantEmbedding{Uid: 7, TransactionId: 100, EmbeddingModel: "old-model", ContentHash: "old-hash", VectorData: "[1,2]", CreatedUnixTime: 123, UpdatedUnixTime: 456}
	_, err := engine.Insert(legacy)
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	engine.SetMaxIdleConns(1)
	legacyTable := engine.TableName(&models.AIAssistantEmbedding{})
	cacheAccesses := 0
	setAuthorizer := func(enabled bool) {
		conn, err := engine.DB().DB.Conn(context.Background())
		require.NoError(t, err)
		require.NoError(t, conn.Raw(func(raw any) error {
			sqliteConn := raw.(*sqlite3.SQLiteConn)
			if !enabled {
				sqliteConn.RegisterAuthorizer(nil)
				return nil
			}
			sqliteConn.RegisterAuthorizer(func(op int, table, column, database string) int {
				if table == legacyTable {
					cacheAccesses++
					return sqlite3.SQLITE_DENY
				}
				return sqlite3.SQLITE_OK
			})
			return nil
		}))
		require.NoError(t, conn.Close())
	}
	setAuthorizer(true)
	defer setAuthorizer(false)
	// Prove the guard targets the ORM's actual table before trusting a zero
	// count. Acronym mapping produces a_i_assistant_embedding, not its filename.
	var forbiddenRows []models.AIAssistantEmbedding
	require.Error(t, engine.Find(&forbiddenRows))
	require.Positive(t, cacheAccesses, "positive read control must reach the cache guard")
	readDenials := cacheAccesses
	_, err = engine.Where("uid = ?", 7).Cols("content_hash").Update(&models.AIAssistantEmbedding{ContentHash: "must-be-denied"})
	require.Error(t, err)
	require.Greater(t, cacheAccesses, readDenials, "positive write control must reach the cache guard")
	cacheAccesses = 0
	// Use the actual container and HTTP provider so no dependency override can
	// hide a call to the retired embedding endpoint.
	requests := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		require.Equal(t, "/chat/completions", r.URL.Path)
		var input struct {
			Messages []struct{ Role string } `json:"messages"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&input))
		if input.Messages[len(input.Messages)-1].Role == "user" {
			io.WriteString(w, `{"choices":[{"index":0,"finish_reason":"tool_calls","message":{"role":"assistant","tool_calls":[{"id":"balances","type":"function","function":{"name":"get_account_balances","arguments":"{}"}}]}}]}`)
		} else {
			io.WriteString(w, `{"choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"余额查询完成。"}}]}`)
		}
	}))
	defer upstream.Close()
	config := settings.Container.GetCurrentConfig()
	config.AIAssistantLLMConfig = &settings.LLMConfig{LLMProvider: settings.OpenAILLMProvider, OpenAIBaseURL: upstream.URL, OpenAIModelID: "isolated-no-embedding", LargeLanguageModelAPIRequestTimeout: 1000, LargeLanguageModelAPIProxy: "none"}
	previous := *llm.Container
	t.Cleanup(func() { *llm.Container = previous })
	require.NoError(t, llm.InitializeLargeLanguageModelProvider(config))
	a.assistant = nil
	for _, stream := range []bool{false, true} {
		rec, apiErr := harnessRequest(t, a, `{"message":"当前余额"}`, stream)
		require.Nil(t, apiErr)
		require.Contains(t, rec.Body.String(), `"tool":"get_account_balances"`)
		require.Contains(t, rec.Body.String(), "余额查询完成")
	}
	require.Equal(t, 4, requests)
	require.Zero(t, cacheAccesses)
	setAuthorizer(false)
	var retained models.AIAssistantEmbedding
	found, err := engine.ID([]any{int64(7), int64(100), "old-model"}).Get(&retained)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, *legacy, retained)
}
