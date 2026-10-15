package api

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
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
	"github.com/stretchr/testify/require"
)

func snapshotFixture(t *testing.T) models.AssistantRunSnapshot {
	t.Helper()
	raw, err := os.ReadFile("../../testdata/assistant_snapshot_v1.json")
	require.NoError(t, err)
	var snapshot models.AssistantRunSnapshot
	require.NoError(t, json.Unmarshal(raw, &snapshot))
	return snapshot
}
func TestAssistantSnapshotFrameExactBytes(t *testing.T) {
	snapshot := snapshotFixture(t)
	frame, err := assistantSnapshotFrame(snapshot, 262144)
	require.NoError(t, err)
	require.Contains(t, string(frame), `"sourceAmountMinor":"-9223372036854775808"`)
	require.Contains(t, string(frame), `"destinationAmountMinor":"9223372036854775807"`)
	require.Contains(t, string(frame), "id: 9223372036854775807:12\n")
	_, err = assistantSnapshotFrame(snapshot, len(frame)-1)
	require.ErrorIs(t, err, models.AssistantReasonOutputLimit)
	_, err = assistantSnapshotFrame(snapshot, len(frame))
	require.NoError(t, err)
	snapshot.Content = strings.Repeat("中", 90000)
	_, err = assistantSnapshotFrame(snapshot, 262144)
	require.ErrorIs(t, err, models.AssistantReasonOutputLimit)
}
func TestAssistantSnapshotHTTPReconnectAndDisconnect(t *testing.T) {
	snapshot := snapshotFixture(t)
	var lookups atomic.Int32
	var returned atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gc, _ := gin.CreateTestContext(w)
		gc.Request = r
		c := &core.WebContext{Context: gc}
		current := snapshot
		if r.URL.Path == "/running" {
			current.State = models.AssistantRunRunning
		}
		frame, err := assistantSnapshotFrame(current, 262144)
		require.NoError(t, err)
		c.Header("Content-Type", "text/event-stream")
		streamAssistantSnapshots(c, current, frame, func() (models.AssistantRunSnapshot, error) {
			lookups.Add(1)
			return current, nil
		}, 262144, time.Millisecond, 5*time.Millisecond, 50*time.Millisecond, time.Now().Add(time.Minute), nil)
		returned.Add(1)
	}))
	defer server.Close()
	for i := 0; i < 2; i++ {
		response, err := http.Get(server.URL)
		require.NoError(t, err)
		raw, err := io.ReadAll(response.Body)
		response.Body.Close()
		require.NoError(t, err)
		require.Equal(t, 1, strings.Count(string(raw), "event: run.terminal"))
		require.Contains(t, string(raw), snapshot.Content[:1])
	}
	require.Equal(t, int32(0), lookups.Load())
	ctx, cancel := context.WithCancel(context.Background())
	request, err := http.NewRequestWithContext(ctx, "GET", server.URL+"/running", nil)
	require.NoError(t, err)
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	line, err := bufio.NewReader(response.Body).ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "event: message.snapshot\n", line)
	cancel()
	response.Body.Close()
	require.Eventually(t, func() bool { return returned.Load() == 3 }, time.Second, time.Millisecond)
}

// A real socket that never drains the body eventually reaches its write deadline.
// The independent durable state can still advance while this reader is stalled.
func TestAssistantSnapshotHTTPSlowReaderBounded(t *testing.T) {
	snapshot := snapshotFixture(t)
	snapshot.State = models.AssistantRunRunning
	snapshot.Content = strings.Repeat("中", 20000)
	var returned atomic.Bool
	var durableRevision atomic.Int64
	durableRevision.Store(snapshot.Revision)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gc, _ := gin.CreateTestContext(w)
		gc.Request = r
		c := &core.WebContext{Context: gc}
		frame, err := assistantSnapshotFrame(snapshot, 262144)
		require.NoError(t, err)
		streamAssistantSnapshots(c, snapshot, frame, func() (models.AssistantRunSnapshot, error) {
			next := snapshot
			next.Revision = durableRevision.Add(1)
			return next, nil
		}, 262144, time.Millisecond, time.Second, 20*time.Millisecond, time.Now().Add(time.Minute), nil)
		returned.Store(true)
	}))
	defer server.Close()
	connection, err := net.Dial("tcp", server.Listener.Addr().String())
	require.NoError(t, err)
	defer connection.Close()
	_, err = io.WriteString(connection, "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n")
	require.NoError(t, err)
	require.Eventually(t, returned.Load, 5*time.Second, 10*time.Millisecond)
	before := durableRevision.Load()
	durableRevision.Add(1)
	require.Equal(t, before+1, durableRevision.Load())
}

func TestAssistantSnapshotHTTPLeaseQuotaAndAuthorization(t *testing.T) {
	api, engine := harnessAPIFixture(t)
	require.NoError(t, engine.Sync2(&models.AssistantUserState{}, &models.AssistantConversation{}, &models.AssistantMessage{}))
	storage := services.NewAssistantStorageService(datastore.Container)
	api.experience = &assistantExperienceDependencies{storage: storage, runs: services.NewAssistantRunService(storage, services.AssistantRunOptions{})}
	_, err := engine.Insert(&models.AssistantConversation{Id: 111, Uid: 7}, &models.AssistantMessage{Id: 112, UserMessageId: 113, CreatedAt: time.Now().UnixMilli(), Uid: 7, ConversationId: 111, Role: "assistant", RunState: models.AssistantRunRunning, SchemaVersion: 1, SnapshotRevision: 1, AttemptNumber: 1})
	require.NoError(t, err)
	var leaseIDs []string
	for i := 0; i < 4; i++ {
		id, err := storage.AcquireSubscription(core.NewNullContext(), 7, "112", time.Minute)
		leaseIDs = append(leaseIDs, id)
		require.NoError(t, err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gc, _ := gin.CreateTestContext(w)
		gc.Request = r
		c := &core.WebContext{Context: gc}
		uid, expiry := int64(7), time.Now().Add(time.Hour).Unix()
		if r.Header.Get("X-Test-User") == "other" {
			uid = 8
		}
		if r.Header.Get("X-Test-User") == "expired" {
			expiry = time.Now().Add(-time.Hour).Unix()
		}
		c.SetTokenClaims(&core.UserTokenClaims{Uid: uid, ExpiresAt: expiry})
		if apiErr := api.AssistantRunStreamHandler(c); apiErr != nil {
			c.JSON(int(apiErr.HttpStatusCode), utils.GetJsonErrorResult(apiErr, r.URL.Path))
		}
	}))
	defer server.Close()
	for _, scenario := range []struct {
		user   string
		status int
	}{{"", 429}, {"other", 404}, {"expired", 401}} {
		request, err := http.NewRequest("GET", server.URL+"?runId=112&afterRevision=1", nil)
		require.NoError(t, err)
		request.Header.Set("X-Test-User", scenario.user)
		response, err := http.DefaultClient.Do(request)
		require.NoError(t, err)
		raw, err := io.ReadAll(response.Body)
		response.Body.Close()
		require.NoError(t, err)
		require.Equal(t, scenario.status, response.StatusCode)
		require.Contains(t, response.Header.Get("Content-Type"), "application/json")
		require.Equal(t, "no-store", response.Header.Get("Cache-Control"))
		require.NotContains(t, string(raw), "event:")
		if scenario.status == 429 {
			require.Equal(t, "15", response.Header.Get("Retry-After"))
		}
	}
	require.NoError(t, storage.ReleaseSubscription(core.NewNullContext(), 7, "112", leaseIDs[0]))
	_, err = engine.Where("id = ?", 112).Cols("run_state", "content").Update(&models.AssistantMessage{RunState: models.AssistantRunCompleted, Content: "saved answer"})
	require.NoError(t, err)
	response, err := http.Get(server.URL + "?runId=112&afterRevision=1")
	require.NoError(t, err)
	raw, err := io.ReadAll(response.Body)
	response.Body.Close()
	require.NoError(t, err)
	require.Equal(t, 200, response.StatusCode)
	require.Equal(t, "no", response.Header.Get("X-Accel-Buffering"))
	require.Contains(t, string(raw), "event: run.terminal")
	require.Contains(t, string(raw), "saved answer")
	require.Equal(t, 1, strings.Count(string(raw), "event:"))

}
