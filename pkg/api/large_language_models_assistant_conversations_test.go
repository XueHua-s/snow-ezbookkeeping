package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
	"github.com/stretchr/testify/require"
)

func TestAssistantExperienceHTTPContracts(t *testing.T) {
	a, engine := harnessAPIFixture(t)
	require.NoError(t, engine.Sync2(&models.AssistantConversation{}, &models.AssistantMessage{}, &models.AssistantMemory{}, &models.AssistantUserState{}, &models.AssistantCommandReceipt{}))
	storage := services.NewAssistantStorageService(datastore.Container)
	var next atomic.Int64
	next.Store(100)
	conversations := services.NewAssistantConversationService(storage, services.AssistantConversationOptions{Secret: []byte("isolated-secret"), GenerateID: func(uuid.UuidType) int64 { return next.Add(1) }, GenerationAvailable: func(core.Context, int64) bool { return true }})
	a.experience = &assistantExperienceDependencies{storage: storage, conversations: conversations, generationAvailable: func(core.Context, int64) bool { return false }}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handlers := map[string]core.ApiHandlerFunc{"/send": a.AssistantMessageSendHandler, "/get": a.AssistantConversationGetHandler, "/capabilities": a.AssistantCapabilitiesHandler, "/command": a.AssistantCommandGetHandler, "/list": a.AssistantConversationListHandler}
	for path, handler := range handlers {
		router.Any(path, func(gc *gin.Context) {
			c := core.WrapWebContext(gc, nil)
			uid, _ := strconv.ParseInt(gc.GetHeader("X-Test-Uid"), 10, 64)
			c.SetTokenClaims(&core.UserTokenClaims{Uid: uid})
			result, err := handler(c)
			if err != nil {
				utils.PrintJsonErrorResult(c, err)
			} else {
				utils.PrintJsonSuccessResult(c, result)
			}
		})
	}
	router.POST("/export", func(gc *gin.Context) {
		c := core.WrapWebContext(gc, nil)
		uid, _ := strconv.ParseInt(gc.GetHeader("X-Test-Uid"), 10, 64)
		c.SetTokenClaims(&core.UserTokenClaims{Uid: uid})
		if err := a.AssistantConversationExportHandler(c); err != nil {
			utils.PrintJsonErrorResult(c, err)
		}
	})
	server := httptest.NewServer(router)
	defer server.Close()
	request := func(method, path, body string, uid int64) (int, map[string]any, string) {
		t.Helper()
		req, err := http.NewRequest(method, server.URL+path, strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-Uid", strconv.FormatInt(uid, 10))
		response, err := server.Client().Do(req)
		require.NoError(t, err)
		defer response.Body.Close()
		require.Equal(t, "no-store", response.Header.Get("Cache-Control"))
		raw, err := io.ReadAll(response.Body)
		require.NoError(t, err)
		var value map[string]any
		_ = json.Unmarshal(raw, &value)
		return response.StatusCode, value, string(raw)
	}
	command := models.AssistantSendRequest{AssistantCommand: models.AssistantCommand{ClientRequestId: "send-1", RequestIssuedAt: time.Now().UTC().Format(time.RFC3339Nano)}, Mode: "chat", Content: "中文 private content", Timezone: "UTC"}
	body, err := json.Marshal(command)
	require.NoError(t, err)
	status, response, _ := request("POST", "/send", string(body), 7)
	require.Equal(t, 200, status)
	result := response["result"].(map[string]any)
	id := result["conversationId"].(string)
	status, replay, _ := request("POST", "/send", string(body), 7)
	require.Equal(t, 200, status)
	require.Equal(t, result, replay["result"])
	for _, suffix := range []string{`,"uid":8}`, `,"history":[]}`, `,"systemPrompt":"inject"}`} {
		bad := strings.TrimSuffix(string(body), "}") + suffix
		status, _, _ = request("POST", "/send", bad, 7)
		require.Equal(t, 400, status)
	}
	for _, bad := range []string{"null", `{"mode":"chat","mode":"summary"}`, `{"Mode":"chat"}`} {
		status, _, _ = request("POST", "/send", bad, 7)
		require.Equal(t, 400, status)
	}
	status, _, _ = request("POST", "/send", string(body)+" {}", 7)
	require.Equal(t, 400, status)
	status, _, _ = request("POST", "/send", strings.Repeat("x", a.CurrentConfig().GetAIAssistantExperienceConfig().Limits.RequestBytes+1), 7)
	require.Equal(t, 413, status)
	status, _, raw := request("GET", "/get?id="+id, "", 8)
	require.Equal(t, 404, status)
	require.NotContains(t, raw, "private content")
	status, _, _ = request("GET", "/get?id="+id, "", 7)
	require.Equal(t, 200, status)
	for _, query := range []string{"limit=-1", "limit=abc", "limit=101", "archived=1", "limit=1&limit=2", "uid=8"} {
		status, _, _ = request("GET", "/list?"+query, "", 7)
		require.Equal(t, 400, status)
	}
	status, response, _ = request("GET", "/command?clientRequestId=unknown", "", 7)
	require.Equal(t, 200, status)
	require.Equal(t, "unknown", response["result"].(map[string]any)["status"])
	status, response, _ = request("GET", "/capabilities", "", 7)
	require.Equal(t, 200, status)
	capabilities := response["result"].(map[string]any)
	require.Equal(t, false, capabilities["generationAvailable"])
	require.Equal(t, true, capabilities["historyAvailable"])
	_, response, _ = request("GET", "/capabilities", "", 8)
	require.NotEqual(t, capabilities["ownerScope"], response["result"].(map[string]any)["ownerScope"])
	status, _, raw = request("POST", "/export", `{"id":"`+id+`","format":"json"}`, 8)
	require.Equal(t, 404, status)
	require.NotContains(t, raw, "private content")
	status, _, raw = request("POST", "/export", `{"id":"`+id+`","format":"json"}`, 7)
	require.Equal(t, 200, status)
	require.Contains(t, raw, "private content")
	require.NotContains(t, raw, "CanonicalRequestHash")
	require.NotContains(t, raw, `"continuation"`)
}
