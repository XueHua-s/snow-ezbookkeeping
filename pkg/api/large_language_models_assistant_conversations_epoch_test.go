package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestAssistantExperienceHTTPDelayedSendCannotCrossClearEpoch(t *testing.T) {
	a, engine := harnessAPIFixture(t)
	require.NoError(t, engine.Sync2(&models.AssistantConversation{}, &models.AssistantMessage{}, &models.AssistantMemory{}, &models.AssistantUserState{}, &models.AssistantCommandReceipt{}))
	storage := services.NewAssistantStorageService(datastore.Container)
	var ids, admissions atomic.Int64
	ids.Store(100)
	conversations := services.NewAssistantConversationService(storage, services.AssistantConversationOptions{Secret: []byte("epoch-http-fixture"), GenerateID: func(uuid.UuidType) int64 { return ids.Add(1) }, GenerationAvailable: func(core.Context, int64) bool { admissions.Add(1); return true }})
	a.experience = &assistantExperienceDependencies{storage: storage, conversations: conversations}
	arrived, release := make(chan struct{}), make(chan struct{})
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if c.GetHeader("X-Test-Pause") == "true" {
			close(arrived)
			select {
			case <-release:
			case <-c.Request.Context().Done():
				return
			}
		}
		c.Next()
	})
	bind := func(handler core.ApiHandlerFunc) gin.HandlerFunc {
		return func(gc *gin.Context) {
			c := core.WrapWebContext(gc, nil)
			c.SetTokenClaims(&core.UserTokenClaims{Uid: 7})
			result, err := handler(c)
			if err != nil {
				utils.PrintJsonErrorResult(c, err)
			} else {
				utils.PrintJsonSuccessResult(c, result)
			}
		}
	}
	router.POST("/api/v1/llm/assistant/messages/send.json", bind(a.AssistantMessageSendHandler))
	router.POST("/api/v1/llm/assistant/conversations/clear.json", bind(a.AssistantConversationClearHandler))
	router.GET("/api/v1/llm/assistant/conversations/get.json", bind(a.AssistantConversationGetHandler))
	server := httptest.NewServer(router)
	defer server.Close()
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	type response struct {
		status int
		body   map[string]any
		err    error
	}
	request := func(path string, payload any, pause bool) response {
		raw, err := json.Marshal(payload)
		if err != nil {
			return response{err: err}
		}
		req, err := http.NewRequest("POST", server.URL+path, bytes.NewReader(raw))
		if err != nil {
			return response{err: err}
		}
		req.Header.Set("Content-Type", "application/json")
		if pause {
			req.Header.Set("X-Test-Pause", "true")
		}
		res, err := server.Client().Do(req)
		if err != nil {
			return response{err: err}
		}
		defer res.Body.Close()
		var body map[string]any
		err = json.NewDecoder(res.Body).Decode(&body)
		return response{status: res.StatusCode, body: body, err: err}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	stale := models.AssistantSendRequest{AssistantCommand: models.AssistantCommand{ClientRequestId: "delayed-never-accepted", RequestIssuedAt: now, ExpectedDeletionEpoch: 0}, Mode: "chat", Content: "must never be resurrected", Timezone: "UTC"}
	finished := make(chan response, 1)
	go func() { finished <- request("/api/v1/llm/assistant/messages/send.json", stale, true) }()
	select {
	case <-arrived:
	case <-time.After(3 * time.Second):
		t.Fatal("send did not reach HTTP barrier")
	}
	// The held request has not reached decoding, command admission or ID allocation.
	require.Equal(t, int64(100), ids.Load())
	require.Zero(t, admissions.Load())
	clear := models.AssistantClearRequest{AssistantCommand: models.AssistantCommand{ClientRequestId: "clear-once", RequestIssuedAt: now, ExpectedDeletionEpoch: 0}}
	cleared := request("/api/v1/llm/assistant/conversations/clear.json", clear, false)
	require.NoError(t, cleared.err)
	require.Equal(t, 200, cleared.status)
	require.Equal(t, float64(1), cleared.body["result"].(map[string]any)["deletionEpoch"])
	close(release)
	var rejected response
	select {
	case rejected = <-finished:
	case <-time.After(3 * time.Second):
		t.Fatal("released send did not finish")
	}
	require.NoError(t, rejected.err)
	require.Equal(t, 409, rejected.status)
	require.Equal(t, "data_epoch_conflict", rejected.body["context"].(map[string]any)["reasonCode"])
	require.Equal(t, int64(100), ids.Load())
	require.Zero(t, admissions.Load(), "stale command must not reach generation admission")
	for _, table := range []any{&models.AssistantConversation{}, &models.AssistantMessage{}} {
		count, err := engine.Count(table)
		require.NoError(t, err)
		require.Zero(t, count, "rejected late send created an object")
	}
	receipt, err := storage.GetCommand(core.NewNullContext(), 7, stale.ClientRequestId)
	require.NoError(t, err)
	require.Equal(t, "unknown", receipt.Status)
	fresh := stale
	fresh.ClientRequestId = "fresh-after-clear"
	fresh.ExpectedDeletionEpoch = 1
	fresh.Content = "keep this new conversation"
	created := request("/api/v1/llm/assistant/messages/send.json", fresh, false)
	require.NoError(t, created.err)
	require.Equal(t, 200, created.status)
	id := created.body["result"].(map[string]any)["conversationId"].(string)
	replay := request("/api/v1/llm/assistant/conversations/clear.json", clear, false)
	require.NoError(t, replay.err)
	require.Equal(t, 200, replay.status)
	require.Equal(t, cleared.body, replay.body)
	// A retry of the already accepted clear is its old receipt, not a new clear.
	res, err := server.Client().Get(server.URL + "/api/v1/llm/assistant/conversations/get.json?id=" + id)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, 200, res.StatusCode)
	var detail map[string]any
	require.NoError(t, json.NewDecoder(res.Body).Decode(&detail))
	messages := detail["result"].(map[string]any)["messages"].([]any)
	require.Equal(t, fresh.Content, messages[0].(map[string]any)["content"])
	state, err := storage.GetUserState(core.NewNullContext(), 7)
	require.NoError(t, err)
	require.Equal(t, int64(1), state.DeletionEpoch)
	require.Equal(t, 1, state.ConversationCount)
}
