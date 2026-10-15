package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const AssistantSchemaVersion = 1

type AssistantRunState string

const (
	AssistantRunQueued          AssistantRunState = "queued"
	AssistantRunPreparing       AssistantRunState = "preparing"
	AssistantRunRunning         AssistantRunState = "running"
	AssistantRunCancelRequested AssistantRunState = "cancel_requested"
	AssistantRunCompleted       AssistantRunState = "completed"
	AssistantRunIncomplete      AssistantRunState = "incomplete"
	AssistantRunFailed          AssistantRunState = "failed"
	AssistantRunCanceled        AssistantRunState = "canceled"
	AssistantRunInterrupted     AssistantRunState = "interrupted"
)

func (s AssistantRunState) IsTerminal() bool {
	switch s {
	case AssistantRunCompleted, AssistantRunIncomplete, AssistantRunFailed, AssistantRunCanceled, AssistantRunInterrupted:
		return true
	}
	return false
}

// CanTransitionTo rejects late callbacks, including completion after cancellation.
func (s AssistantRunState) CanTransitionTo(next AssistantRunState) bool {
	switch s {
	case AssistantRunQueued:
		return next == AssistantRunPreparing || next == AssistantRunCanceled || next == AssistantRunFailed
	case AssistantRunPreparing:
		return next == AssistantRunRunning || next == AssistantRunCancelRequested || next == AssistantRunFailed || next == AssistantRunInterrupted
	case AssistantRunRunning:
		return next == AssistantRunCancelRequested || next == AssistantRunCompleted || next == AssistantRunIncomplete || next == AssistantRunFailed || next == AssistantRunInterrupted
	case AssistantRunCancelRequested:
		return next == AssistantRunCanceled || next == AssistantRunInterrupted
	}
	return false
}

type AssistantReasonCode string

const (
	AssistantReasonQueueTimeout          AssistantReasonCode = "queue_timeout"
	AssistantReasonRunTimeout            AssistantReasonCode = "run_timeout"
	AssistantReasonExecutionLimit        AssistantReasonCode = "execution_limit"
	AssistantReasonGenerationUnavailable AssistantReasonCode = "generation_unavailable"
)

const (
	AssistantReasonInputTooLong         AssistantReasonCode = "input_too_long"
	AssistantReasonConversationConflict AssistantReasonCode = "conversation_conflict"
	AssistantReasonDataEpochConflict    AssistantReasonCode = "data_epoch_conflict"
	AssistantReasonMemoryConflict       AssistantReasonCode = "memory_conflict"
	AssistantReasonContextChanged       AssistantReasonCode = "context_changed"
	AssistantReasonAlreadyRunning       AssistantReasonCode = "already_running"
	AssistantReasonQuotaExceeded        AssistantReasonCode = "quota_exceeded"
	AssistantReasonQueueFull            AssistantReasonCode = "queue_full"
	AssistantReasonSubscriptionLimit    AssistantReasonCode = "subscription_limit"
	AssistantReasonCursorExpired        AssistantReasonCode = "cursor_expired"
	AssistantReasonContextLimit         AssistantReasonCode = "context_limit"
	AssistantReasonOutputLimit          AssistantReasonCode = "output_limit"
	AssistantReasonEmptyResponse        AssistantReasonCode = "empty_response"
	AssistantReasonStorageUnavailable   AssistantReasonCode = "storage_unavailable"
	AssistantReasonProviderUnavailable  AssistantReasonCode = "provider_unavailable"
	AssistantReasonInterrupted          AssistantReasonCode = "interrupted"
	AssistantReasonNotAccessible        AssistantReasonCode = "not_accessible"
)

func (reason AssistantReasonCode) HTTPStatus() int {
	switch reason {
	case AssistantReasonNotAccessible:
		return 404
	case AssistantReasonExecutionLimit, AssistantReasonConversationConflict, AssistantReasonDataEpochConflict, AssistantReasonMemoryConflict, AssistantReasonContextChanged, AssistantReasonAlreadyRunning, AssistantReasonQuotaExceeded:
		return 409
	case AssistantReasonQueueFull, AssistantReasonSubscriptionLimit:
		return 429
	case AssistantReasonStorageUnavailable, AssistantReasonProviderUnavailable, AssistantReasonInterrupted, AssistantReasonQueueTimeout, AssistantReasonRunTimeout, AssistantReasonGenerationUnavailable:
		return 503
	default:
		return 400
	}
}

// ValidateAssistantInput counts Unicode code points, never UTF-16 units, without modifying the original text.
func ValidateAssistantInput(content string, characters, bytes int) error {
	if !utf8.ValidString(content) || strings.TrimSpace(content) == "" {
		return fmt.Errorf("invalid assistant input")
	}
	if utf8.RuneCountInString(content) > characters || len(content) > bytes {
		return fmt.Errorf("%s", AssistantReasonInputTooLong)
	}
	return nil
}

// ValidateTime is only for unknown commands. Existing receipts must be checked first.
func (command AssistantCommand) ValidateTime(now time.Time, maxAge, futureSkew time.Duration) error {
	issued, err := time.Parse(time.RFC3339Nano, command.RequestIssuedAt)
	if err != nil || issued.Before(now.Add(-maxAge)) || issued.After(now.Add(futureSkew)) {
		return fmt.Errorf("invalid command time")
	}
	if strings.TrimSpace(command.ClientRequestId) == "" || len(command.ClientRequestId) > 128 || command.ExpectedDeletionEpoch < 0 {
		return fmt.Errorf("invalid command identity")
	}
	return nil
}

// AssistantCanonicalRequestHash includes the operation as well as every submitted typed field.
// Handlers must reject unknown JSON fields before calling this function.
func AssistantCanonicalRequestHash(commandType string, request any) (string, error) {
	payload, err := json.Marshal(struct {
		Type    string `json:"type"`
		Request any    `json:"request"`
	}{commandType, request})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

// Error exposes only the controlled business reason, never upstream or SQL errors.
func (reason AssistantReasonCode) Error() string { return string(reason) }

// Validate verifies identity and state invariants before a persisted snapshot is
// published. It does not turn malformed or empty output into a successful answer.
func (snapshot AssistantRunSnapshot) Validate() error {
	if snapshot.SchemaVersion != AssistantSchemaVersion || snapshot.RunId == "" || snapshot.ConversationId == "" || snapshot.UserMessageId == "" || snapshot.AssistantMessageId != snapshot.RunId || snapshot.Revision < 1 || snapshot.AttemptNumber < 1 {
		return fmt.Errorf("invalid assistant snapshot identity")
	}
	switch snapshot.State {
	case AssistantRunQueued, AssistantRunPreparing, AssistantRunRunning, AssistantRunCancelRequested, AssistantRunCompleted, AssistantRunIncomplete, AssistantRunFailed, AssistantRunCanceled, AssistantRunInterrupted:
	default:
		return fmt.Errorf("invalid assistant snapshot state")
	}
	if snapshot.State == AssistantRunCompleted && strings.TrimSpace(snapshot.Content) == "" {
		return AssistantReasonEmptyResponse
	}
	if _, err := time.Parse(time.RFC3339Nano, snapshot.SavedAt); err != nil {
		return fmt.Errorf("invalid snapshot save time")
	}
	return nil
}

// ValidateQueryAction checks the public paging shape. The conversation service
// must additionally authorize the source and enforce its frozen-prefix boundary.
func (request AssistantSendRequest) ValidateQueryAction() error {
	if request.RestartQuery && request.Mode != "next_page" {
		return fmt.Errorf("restartQuery requires next_page")
	}
	if request.Mode == "next_page" && strings.TrimSpace(request.SourceMessageId) == "" {
		return fmt.Errorf("next_page requires sourceMessageId")
	}
	return nil
}
