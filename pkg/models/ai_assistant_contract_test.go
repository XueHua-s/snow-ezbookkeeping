package models

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAssistantContractInput(t *testing.T) {
	require.NoError(t, ValidateAssistantInput(strings.Repeat("😀", 4000), 4000, 16384))
	require.Error(t, ValidateAssistantInput(strings.Repeat("a", 4001), 4000, 16384))
	require.Error(t, ValidateAssistantInput("😀", 4000, 3))
	require.Error(t, ValidateAssistantInput(" \n", 4000, 16384))
	require.Error(t, ValidateAssistantInput(string([]byte{255}), 4000, 16384))
}
func TestAssistantContractCommandTimeAndHash(t *testing.T) {
	now := time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC)
	command := AssistantCommand{ClientRequestId: "stable-id", ExpectedDeletionEpoch: 3}
	for _, delta := range []time.Duration{-24 * time.Hour, 5 * time.Minute, 0} {
		command.RequestIssuedAt = now.Add(delta).Format(time.RFC3339Nano)
		require.NoError(t, command.ValidateTime(now, 24*time.Hour, 5*time.Minute))
	}
	for _, delta := range []time.Duration{-24*time.Hour - time.Nanosecond, 5*time.Minute + time.Nanosecond} {
		command.RequestIssuedAt = now.Add(delta).Format(time.RFC3339Nano)
		require.Error(t, command.ValidateTime(now, 24*time.Hour, 5*time.Minute))
	}
	a, err := AssistantCanonicalRequestHash("clear", command)
	require.NoError(t, err)
	b, err := AssistantCanonicalRequestHash("delete", command)
	require.NoError(t, err)
	require.NotEqual(t, a, b)
	command.ExpectedDeletionEpoch++
	c, err := AssistantCanonicalRequestHash("clear", command)
	require.NoError(t, err)
	require.NotEqual(t, a, c)
}
func TestAssistantContractTerminalFencing(t *testing.T) {
	for _, s := range []AssistantRunState{AssistantRunCompleted, AssistantRunFailed, AssistantRunIncomplete, AssistantRunCanceled, AssistantRunInterrupted} {
		require.True(t, s.IsTerminal())
		require.False(t, s.CanTransitionTo(AssistantRunRunning))
		require.False(t, s.CanTransitionTo(AssistantRunCompleted))
	}
	require.False(t, AssistantRunCancelRequested.CanTransitionTo(AssistantRunCompleted))
	require.True(t, AssistantRunCancelRequested.CanTransitionTo(AssistantRunCanceled))
	require.True(t, AssistantRunQueued.CanTransitionTo(AssistantRunPreparing))
}
func TestAssistantContractWirePrecisionAndNull(t *testing.T) {
	snapshot := AssistantRunSnapshot{SchemaVersion: 1, RunId: "9223372036854775807", ReasonCode: nil, References: []AssistantReferenceDTO{{Id: "9", SourceAmountMinor: "-9223372036854775808", DestinationAmountMinor: "9223372036854775807"}}}
	b, err := json.Marshal(snapshot)
	require.NoError(t, err)
	require.Contains(t, string(b), `"sourceAmountMinor":"-9223372036854775808"`)
	require.Contains(t, string(b), `"runId":"9223372036854775807"`)
	require.Contains(t, string(b), `"reasonCode":null`)
	require.NotContains(t, string(b), `"sourceAmount":`)
	var copy AssistantRunSnapshot
	require.NoError(t, json.Unmarshal(b, &copy))
	require.Equal(t, snapshot, copy)
}

func TestAssistantSnapshotRejectsFalseCompletion(t *testing.T) {
	snapshot := AssistantRunSnapshot{SchemaVersion: 1, RunId: "2", AssistantMessageId: "2", ConversationId: "1", UserMessageId: "3", AttemptNumber: 1, Revision: 1, State: AssistantRunCompleted, SavedAt: "2026-10-15T08:00:00Z", Content: " \n"}
	require.ErrorIs(t, snapshot.Validate(), AssistantReasonEmptyResponse)
	snapshot.Content = "Saved answer"
	require.NoError(t, snapshot.Validate())
	snapshot.AssistantMessageId = "other"
	require.Error(t, snapshot.Validate())
	snapshot.AssistantMessageId = "2"
	snapshot.SchemaVersion = 2
	require.Error(t, snapshot.Validate())
}

// This fixture is shared with browser transport tests; amounts must never pass through float64.
func TestAssistantSharedSnapshotFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/assistant_snapshot_v1.json")
	require.NoError(t, err)
	var snapshot AssistantRunSnapshot
	require.NoError(t, json.Unmarshal(data, &snapshot))
	require.NoError(t, snapshot.Validate())
	require.Equal(t, "-9223372036854775808", snapshot.References[0].SourceAmountMinor)
	require.Equal(t, "9223372036854775807", snapshot.References[0].DestinationAmountMinor)
	require.True(t, snapshot.References[0].HideAmount)
	encoded, err := json.Marshal(snapshot)
	require.NoError(t, err)
	require.JSONEq(t, string(data), string(encoded))
}

func TestAssistantExplicitQueryRestart(t *testing.T) {
	require.Error(t, (AssistantSendRequest{Mode: "chat", RestartQuery: true, SourceMessageId: "1"}).ValidateQueryAction())
	require.Error(t, (AssistantSendRequest{Mode: "next_page", RestartQuery: true}).ValidateQueryAction())
	request := AssistantSendRequest{Mode: "next_page", RestartQuery: true, SourceMessageId: "1"}
	require.NoError(t, request.ValidateQueryAction())
	encoded, err := json.Marshal(request)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"restartQuery":true`)
	restartHash, err := AssistantCanonicalRequestHash("send", request)
	require.NoError(t, err)
	request.RestartQuery = false
	nextHash, err := AssistantCanonicalRequestHash("send", request)
	require.NoError(t, err)
	require.NotEqual(t, restartHash, nextHash)
}

func TestAssistantRunFailureReasons(t *testing.T) {
	for _, reason := range []AssistantReasonCode{AssistantReasonQueueTimeout, AssistantReasonRunTimeout, AssistantReasonGenerationUnavailable} {
		require.Equal(t, 503, reason.HTTPStatus())
	}
	require.Equal(t, 409, AssistantReasonExecutionLimit.HTTPStatus())
	require.NotEqual(t, AssistantReasonQueueTimeout, AssistantReasonQueueFull)
	require.NotEqual(t, AssistantReasonExecutionLimit, AssistantReasonProviderUnavailable)
}

func TestAssistantExportViewedVersionWire(t *testing.T) {
	request := AssistantExportRequest{Id: "10", Format: "markdown", ViewedVersionIds: []string{"21", "31"}}
	encoded, err := json.Marshal(request)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"viewedVersionIds":["21","31"]`)
	var decoded AssistantExportRequest
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	require.Equal(t, request, decoded)
}

func TestAssistantClearCompletionWire(t *testing.T) {
	ordinary := AssistantCommandResult{CommandType: "send"}
	encoded, err := json.Marshal(ordinary)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "operationCompleted")
	completed := false
	ordinary.CommandType = "user_data.clear_assistant"
	ordinary.OperationCompleted = &completed
	encoded, err = json.Marshal(ordinary)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"operationCompleted":false`)
	completed = true
	encoded, err = json.Marshal(ordinary)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"operationCompleted":true`)
}
