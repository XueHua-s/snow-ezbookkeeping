package llm

import (
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
)

// These protocol budgets bound one request, never the searchable ledger. The
// configurable round limit and timeouts come from settings via AssistantRunInput.
const (
	AssistantMaxToolCalls         = 24
	AssistantMaxToolResultBytes   = 64 * 1024
	AssistantMaxModelRequestBytes = data.AssistantMaxRequestBytes
	AssistantMaxReferences        = 8
	AssistantMaxHistoryMessages   = 12
)

// ToolExecutor is implemented by a request-scoped service bound to the logged-in
// UID. The runner cannot choose a database or pass an alternate UID to a tool.
type ToolExecutor interface {
	Definitions() []data.ToolDefinition
	Execute(c core.Context, call data.ToolCall) data.ToolResult
}

type AssistantState string

const (
	AssistantStateRequesting AssistantState = "requesting"
	AssistantStateCompleted  AssistantState = "completed"
	AssistantStateIncomplete AssistantState = "incomplete"
	AssistantStateFailed     AssistantState = "failed"
	AssistantStateCanceled   AssistantState = "canceled"
)

type AssistantStatus string

const (
	AssistantStatusQuerying    AssistantStatus = "querying"
	AssistantStatusAggregating AssistantStatus = "aggregating"
	AssistantStatusAnswering   AssistantStatus = "answering"
)

type AssistantEventType string

const (
	AssistantEventStatus        AssistantEventType = "status"
	AssistantEventThinkingDelta AssistantEventType = "thinking_delta"
	AssistantEventReplyDelta    AssistantEventType = "reply_delta"
	AssistantEventReferences    AssistantEventType = "references"
)

// AssistantEvent has no terminal event: the API serializes exactly one terminal
// response from Run's return value. Tool names/arguments are never status text.
type AssistantEvent struct {
	Type       AssistantEventType
	Status     AssistantStatus
	Delta      string
	References []data.TransactionReference
}

// EventSink is called synchronously and serially. A returned write error stops
// the run; callers must not launch unbounded or unjoined callback goroutines.
// A nil sink selects JSON mode without fabricating a first-status timestamp.
type EventSink func(event AssistantEvent) error

type AssistantRunInput struct {
	// PriorModelCalls accounts for preparing calls within this same foreground run.
	PriorModelCalls uint32
	MaxInputTokens  int
	MaxRequestBytes int
	MaxOutputTokens int

	UID          int64
	Mode         string
	Messages     []data.ConversationMessage
	Continuation *data.ToolContinuation
	MaxRounds    uint32
	RunTimeout   time.Duration
	QueryTimeout time.Duration
	Observe      AssistantObserver
	// The API reports the first successfully written status only. The runner
	// reads this once when finalizing its observation. JSON returns nil; heartbeats
	// never count. A concurrently written timestamp must be synchronized by API.
	FirstStatusWrittenAt func() *time.Time
	// Finalize delivers the sole terminal response before the observation is
	// recorded. A transport error changes the outcome to write_failed. It must
	// return promptly on request cancellation; JSON callers may leave it nil.
	Finalize func(AssistantRunResult, error) error
}

// AssistantRunResult retains confirmed evidence even on error. Continuation
// refers to the last successful pageable query and never resumes automatically.
type AssistantRunResult struct {
	StopReason   AssistantStopReason
	Mode         string
	State        AssistantState
	Reply        string
	Thinking     string
	Queries      []data.QueryEvidence
	References   []data.TransactionReference
	Continuation *data.ToolContinuation
}

type AssistantObservationKind string

const (
	AssistantObservationTool AssistantObservationKind = "tool"
	AssistantObservationRun  AssistantObservationKind = "run"
)

// AssistantStopReason is a controlled vocabulary. Never cast error.Error() to
// this type: driver/provider messages may contain financial data or secrets.
type AssistantStopReason string

const (
	AssistantReasonCompleted           AssistantStopReason = "completed"
	AssistantReasonCanceled            AssistantStopReason = "canceled"
	AssistantReasonRunTimeout          AssistantStopReason = "run_timeout"
	AssistantReasonQueryTimeout        AssistantStopReason = "query_timeout"
	AssistantReasonRoundBudget         AssistantStopReason = "round_budget"
	AssistantReasonCallBudget          AssistantStopReason = "call_budget"
	AssistantReasonContextBudget       AssistantStopReason = "context_budget"
	AssistantReasonOutputBudget        AssistantStopReason = "output_limit"
	AssistantReasonResultBudget        AssistantStopReason = "result_budget"
	AssistantReasonNoProgress          AssistantStopReason = "no_progress"
	AssistantReasonRepeatedError       AssistantStopReason = "repeated_error"
	AssistantReasonInvalidArgument     AssistantStopReason = "invalid_argument"
	AssistantReasonNotAccessible       AssistantStopReason = "not_accessible"
	AssistantReasonInvalidCursor       AssistantStopReason = "invalid_cursor"
	AssistantReasonDataInconsistent    AssistantStopReason = "data_inconsistent"
	AssistantReasonProviderUnsupported AssistantStopReason = "provider_unsupported"
	AssistantReasonUpstreamFailed      AssistantStopReason = "upstream_failed"
	AssistantReasonProtocolError       AssistantStopReason = "protocol_error"
	AssistantReasonWriteFailed         AssistantStopReason = "write_failed"
)

// AssistantObservation is an internal, allowlisted telemetry record, never an
// HTTP query DTO. RowsRead is application-consumed rows, not optimizer scans.
// Rounds counts actual model requests (including fallback); Executions counts
// real tool executions, while CacheHits counts reused results separately.
// The runner owns aggregation and emits one terminal run observation; tools
// supply only their measurement via ToolResult.Metrics. No SQL, arguments,
// financial text, cursors, provider state or arbitrary error string belongs here.
type AssistantObservation struct {
	Kind                 AssistantObservationKind `json:"kind"`
	RequestID            string                   `json:"requestId"`
	Tool                 data.ToolName            `json:"tool,omitempty"`
	RowsRead             int64                    `json:"rowsRead"`
	ResultRecords        int64                    `json:"resultRecords"`
	SerializedBytes      int64                    `json:"serializedBytes"`
	DurationMilliseconds int64                    `json:"durationMilliseconds"`
	Rounds               uint32                   `json:"rounds"`
	Executions           int                      `json:"executions"`
	CacheHits            int                      `json:"cacheHits"`
	Reason               AssistantStopReason      `json:"reason"`
	// nil is "not applicable/no status written", including every JSON run;
	// zero is an actual first status written in less than one millisecond.
	FirstStatusMilliseconds *int64 `json:"firstStatusMilliseconds"`
}

type AssistantObserver func(observation AssistantObservation)
