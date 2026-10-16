package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/stretchr/testify/require"
)

type assistantSlowExportWriter struct {
	gin.ResponseWriter
	ctx context.Context
}

func (w *assistantSlowExportWriter) Write(p []byte) (int, error) {
	select {
	case <-w.ctx.Done():
		return 0, w.ctx.Err()
	case <-time.After(5 * time.Millisecond):
	}
	n, err := w.ResponseWriter.Write(p)
	w.ResponseWriter.Flush()
	return n, err
}
func (w *assistantSlowExportWriter) WriteString(s string) (int, error) { return w.Write([]byte(s)) }
func (w *assistantSlowExportWriter) Unwrap() http.ResponseWriter       { return w.ResponseWriter }

func assistantHTTPExportFixture(t *testing.T, large bool) (*LargeLanguageModelsApi, *httptest.Server, chan error) {
	t.Helper()
	a, engine := harnessAPIFixture(t)
	require.NoError(t, engine.Sync2(&models.AssistantConversation{}, &models.AssistantMessage{}, &models.AssistantMemory{}, &models.AssistantUserState{}))
	config := a.CurrentConfig()
	config.AIAssistantExperience = settings.DefaultAIAssistantExperienceConfig()
	storage := services.NewAssistantStorageService(datastore.Container)
	a.experience = &assistantExperienceDependencies{storage: storage, conversations: services.NewAssistantConversationService(storage, services.AssistantConversationOptions{})}
	count, turns, size := 1, 100, 8000
	if large {
		count, turns, size = 50, 200, 4300
	}
	now := time.Now().UnixMilli()
	for c := 1; c <= count; c++ {
		_, err := engine.Insert(&models.AssistantConversation{Id: int64(c), Uid: 7, Title: "Export test", CreatedAt: now, ActivityAt: now, TurnCount: turns, NextTurnSequence: int64(turns + 1), MetadataVersion: 1, ContextVersion: 1})
		require.NoError(t, err)
		rows := make([]models.AssistantMessage, 0, turns*3)
		for turn := 1; turn <= turns; turn++ {
			id := int64(c*10000 + turn*3)
			rows = append(rows, models.AssistantMessage{Id: id, Uid: 7, ConversationId: int64(c), TurnSequence: int64(turn), Role: "user", UserMessageId: id, Content: strings.Repeat("q", 200), SelectedAssistantMessageId: id + 2, SchemaVersion: 1, CreatedAt: now})
			for v := 1; v <= 2; v++ {
				rows = append(rows, models.AssistantMessage{Id: id + int64(v), Uid: 7, ConversationId: int64(c), TurnSequence: int64(turn), Role: "assistant", UserMessageId: id, Content: strings.Repeat("a", size), AttemptNumber: v, RunState: models.AssistantRunCompleted, SnapshotRevision: 1, SchemaVersion: 1, CreatedAt: now, CompletedAt: now, QueriesJSON: "[]", ReferencesJSON: "[]"})
			}
		}
		for start := 0; start < len(rows); start += 30 {
			_, err = engine.Insert(rows[start:min(start+30, len(rows))])
			require.NoError(t, err)
		}
	}
	done := make(chan error, 4)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/export", func(gc *gin.Context) {
		c := core.WrapWebContext(gc, nil)
		c.SetTokenClaims(&core.UserTokenClaims{Uid: 7})
		if gc.GetHeader("X-Slow-Export") == "true" {
			c.Writer = &assistantSlowExportWriter{ResponseWriter: c.Writer, ctx: c.Request.Context()}
		}
		if err := a.AssistantConversationExportHandler(c); err != nil {
			utils.PrintJsonErrorResult(c, err)
		}
		done <- c.Request.Context().Err()
	})
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	return a, server, done
}

func TestAssistantExperienceHTTPExportCancellationAndDeadline(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		name := "client_cancel"
		if deadline {
			name = "deadline"
		}
		t.Run(name, func(t *testing.T) {
			a, server, done := assistantHTTPExportFixture(t, false)
			if deadline {
				a.CurrentConfig().AIAssistantExperience.Limits.ExportTimeoutSeconds = 1
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, "POST", server.URL+"/export", strings.NewReader(`{"all":true,"format":"json","includeVersions":true}`))
			require.NoError(t, err)
			req.Header.Set("X-Slow-Export", "true")
			started := time.Now()
			response, err := server.Client().Do(req)
			require.NoError(t, err)
			defer response.Body.Close()
			require.Equal(t, 200, response.StatusCode)
			require.Equal(t, "no-store", response.Header.Get("Cache-Control"))
			if !deadline {
				var prefix [32]byte
				_, err = io.ReadFull(response.Body, prefix[:])
				require.NoError(t, err)
				cancel()
			}
			select {
			case cause := <-done:
				require.Error(t, cause)
				if deadline {
					// The socket write deadline and context timer intentionally share
					// the same budget. A write timeout may end the handler first,
					// whose deferred cancel then wins the context timer race.
					require.True(t, errors.Is(cause, context.DeadlineExceeded) || errors.Is(cause, context.Canceled))
					require.GreaterOrEqual(t, time.Since(started), time.Second, "export must reach the configured deadline, not stop early")
				}
			case <-time.After(3 * time.Second):
				t.Fatal("export retained HTTP handler after cancellation/deadline")
			}
			require.Less(t, time.Since(started), 3*time.Second)
			if !deadline {
				require.Less(t, time.Since(started), time.Second, "client cancellation must stop before slow export completes")
			}
			a.experience.limiter.mu.Lock()
			heavy := a.experience.limiter.heavy
			a.experience.limiter.mu.Unlock()
			require.Zero(t, heavy, "export slot must be released")
		})
	}
}

// Complements the three-database service test with a real TCP HTTP response and
// the production handler's default 60-second context and socket deadlines.
func TestAssistantExperienceHTTPLargeExport(t *testing.T) {
	if os.Getenv("EBK_ASSISTANT_HISTORY_PERF") != "1" {
		t.Skip("set EBK_ASSISTANT_HISTORY_PERF=1 for large HTTP export")
	}
	a, server, done := assistantHTTPExportFixture(t, true)
	require.Equal(t, 60, a.CurrentConfig().GetAIAssistantExperienceConfig().Limits.ExportTimeoutSeconds)
	start := time.Now()
	response, err := server.Client().Post(server.URL+"/export", "application/json", strings.NewReader(`{"all":true,"format":"json","includeVersions":true}`))
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, 200, response.StatusCode)
	// Discard through a bounded tail writer: no response-sized allocation, but
	// still prove the final JSON delimiter reached the client before EOF.
	sink := &assistantExportTail{}
	n, err := io.Copy(sink, response.Body)
	require.NoError(t, err)
	require.Greater(t, n, int64(90<<20))
	require.True(t, strings.HasSuffix(sink.tail, `],"memories":[]}`))
	require.Less(t, time.Since(start), 60*time.Second)
	<-done
	t.Logf("HTTP export bytes=%d duration=%s", n, time.Since(start))
}

type assistantExportTail struct{ tail string }

func (w *assistantExportTail) Write(p []byte) (int, error) {
	w.tail += string(p)
	if len(w.tail) > 64 {
		w.tail = w.tail[len(w.tail)-64:]
	}
	return len(p), nil
}
