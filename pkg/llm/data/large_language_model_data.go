package data

import (
	"encoding/json"
	"reflect"
	"time"
)

type LargeLanguageModelRequestPromptType byte

// Large Language Model Request Prompt Type
const (
	LARGE_LANGUAGE_MODEL_REQUEST_PROMPT_TYPE_TEXT      LargeLanguageModelRequestPromptType = 0
	LARGE_LANGUAGE_MODEL_REQUEST_PROMPT_TYPE_IMAGE_URL LargeLanguageModelRequestPromptType = 1
)

type LargeLanguageModelResponseFormat byte

// Large Language Model Response Format
const (
	LARGE_LANGUAGE_MODEL_RESPONSE_FORMAT_TEXT LargeLanguageModelResponseFormat = 0
	LARGE_LANGUAGE_MODEL_RESPONSE_FORMAT_JSON LargeLanguageModelResponseFormat = 1
)

// LargeLanguageModelRequest represents a request to a large language model
type LargeLanguageModelRequest struct {
	Stream                 bool
	SystemPrompt           string
	UserPrompt             []byte
	UserPromptType         LargeLanguageModelRequestPromptType
	UserPromptContentType  string
	ResponseJsonObjectType reflect.Type
}

// LargeLanguageModelTextualResponse represents a textual response from a large language model
type LargeLanguageModelTextualResponse struct {
	Content string
}

type LargeLanguageModelStreamDeltaType byte

const (
	LARGE_LANGUAGE_MODEL_STREAM_DELTA_TYPE_REPLY LargeLanguageModelStreamDeltaType = iota
	LARGE_LANGUAGE_MODEL_STREAM_DELTA_TYPE_THINKING
)

// LargeLanguageModelStreamResponse represents the complete content collected
// while a provider streams deltas to the caller.
type LargeLanguageModelStreamResponse struct {
	Content  string
	Thinking string
}

// LargeLanguageModelStreamCallback receives each non-empty streamed delta.
type LargeLanguageModelStreamCallback func(deltaType LargeLanguageModelStreamDeltaType, delta string)

// ToolName identifies the assistant's closed read-only tool set. The executor
// validates the name; accepting a string here does not grant tool authority.
type ToolName string

const (
	ToolListAccounts          ToolName = "list_accounts"
	ToolListCategories        ToolName = "list_categories"
	ToolListTags              ToolName = "list_tags"
	ToolQueryTransactions     ToolName = "query_transactions"
	ToolGetTransaction        ToolName = "get_transaction"
	ToolAggregateTransactions ToolName = "aggregate_transactions"
	ToolGetAccountBalances    ToolName = "get_account_balances"
)

// ToolDefinition carries a domain-generated JSON schema. Provider-specific
// strict/nullable transformations belong to the adapter, not the domain schema.
type ToolDefinition struct {
	Name        ToolName        `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// ToolCall is executable only after its entire model turn completes. Arguments
// stay raw so execution can reject duplicate keys, unknown fields and nulls.
type ToolCall struct {
	ID        string          `json:"id"`
	Name      ToolName        `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ToolErrorCode string

const (
	ToolErrorInvalidArgument         ToolErrorCode = "INVALID_ARGUMENT"
	ToolErrorNotFoundOrNotAccessible ToolErrorCode = "NOT_FOUND_OR_NOT_ACCESSIBLE"
	ToolErrorQueryTimeout            ToolErrorCode = "QUERY_TIMEOUT"
	ToolErrorCanceled                ToolErrorCode = "CANCELED"
	ToolErrorBudgetExhausted         ToolErrorCode = "BUDGET_EXHAUSTED"
	ToolErrorProviderUnsupported     ToolErrorCode = "PROVIDER_UNSUPPORTED"
	ToolErrorUpstreamFailed          ToolErrorCode = "UPSTREAM_FAILED"
	ToolErrorResultTooLarge          ToolErrorCode = "RESULT_TOO_LARGE"
	ToolErrorInvalidCursor           ToolErrorCode = "INVALID_CURSOR"
	ToolErrorDataInconsistent        ToolErrorCode = "DATA_INCONSISTENT"
	ToolErrorProtocol                ToolErrorCode = "PROTOCOL_ERROR"
)

// ToolError contains controlled, safe-to-display descriptions, never raw SQL,
// driver/provider errors, credentials or information about another user's IDs.
type ToolError struct {
	Code      ToolErrorCode `json:"code"`
	Message   string        `json:"message"`
	Retryable bool          `json:"retryable"`
	Field     string        `json:"field,omitempty"`
}

// Error allows provider failures to carry the same controlled error code to
// the runner without exposing transport-specific errors or parsing their text.
func (e *ToolError) Error() string {
	return string(e.Code) + ": " + e.Message
}

type QueryCoverage string

const (
	QueryCoveragePage        QueryCoverage = "page"
	QueryCoverageAllMatching QueryCoverage = "all_matching"
)

type QueryPeriod struct {
	Name         string `json:"name"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	MatchedCount string `json:"matched_count,omitempty"`
}

// QueryScope describes effective filters. Nil time bounds mean unbounded,
// including dates in the future. Account/category IDs are expanded and deduped.
type QueryScope struct {
	StartTime       *string       `json:"start_time"`
	EndTime         *string       `json:"end_time"`
	Timezone        string        `json:"timezone"`
	Currency        string        `json:"currency,omitempty"`
	Types           []string      `json:"types,omitempty"`
	AccountIDs      []string      `json:"account_ids,omitempty"`
	CategoryIDs     []string      `json:"category_ids,omitempty"`
	TagIDs          []string      `json:"tag_ids,omitempty"`
	TagMatch        string        `json:"tag_match,omitempty"`
	NoTags          bool          `json:"no_tags,omitempty"`
	TransferSide    string        `json:"transfer_side,omitempty"`
	AmountMin       *string       `json:"amount_min,omitempty"`
	AmountMax       *string       `json:"amount_max,omitempty"`
	Keyword         string        `json:"keyword,omitempty"`
	Periods         []QueryPeriod `json:"periods,omitempty"`
	IncludesHidden  bool          `json:"includes_hidden"`
	ExcludesDeleted bool          `json:"excludes_deleted"`
}

// TransactionReference is a service-confirmed logical transaction projection.
// Amounts use exact decimal strings; IDs must never pass through JSON numbers.
type TransactionReference struct {
	ID                     string  `json:"transaction_id"`
	Time                   string  `json:"time"`
	TimezoneOffset         int16   `json:"timezone_offset"`
	Type                   string  `json:"type"`
	CategoryID             *string `json:"category_id"`
	CategoryName           string  `json:"category_name,omitempty"`
	SourceAccountID        string  `json:"source_account_id"`
	SourceAccountName      string  `json:"source_account_name"`
	DestinationAccountID   string  `json:"destination_account_id,omitempty"`
	DestinationAccountName string  `json:"destination_account_name,omitempty"`
	SourceAmount           string  `json:"source_amount"`
	DestinationAmount      string  `json:"destination_amount,omitempty"`
	Currency               string  `json:"currency"`
	DestinationCurrency    string  `json:"destination_currency,omitempty"`
	Comment                string  `json:"comment,omitempty"`
	HideAmount             bool    `json:"hide_amount"`
}

// ToolExecutionMetrics is internal measurement, not a claim about optimizer
// scan counts. RowsRead counts rows actually consumed by application code.
type ToolExecutionMetrics struct {
	RowsRead        int64
	ResultRecords   int64
	SerializedBytes int64
	Duration        time.Duration
}

// ToolResult separates full-set totals from a page of records/groups. Data,
// Totals and Comparison are domain-owned business projections, not ORM models.
// An error has OK=false and Error set; it must never become a successful empty
// result. References and metrics are trusted request-local side channels only.
type ToolResult struct {
	OK             bool                   `json:"ok"`
	QueryID        string                 `json:"query_id,omitempty"`
	Scope          *QueryScope            `json:"scope,omitempty"`
	MatchedCount   string                 `json:"matched_count,omitempty"`
	ReturnedCount  int                    `json:"returned_count"`
	HasMore        bool                   `json:"has_more"`
	NextCursor     string                 `json:"next_cursor,omitempty"`
	Coverage       QueryCoverage          `json:"coverage,omitempty"`
	QueriedAt      string                 `json:"queried_at,omitempty"`
	Data           json.RawMessage        `json:"data,omitempty"`
	Totals         json.RawMessage        `json:"totals,omitempty"`
	TotalsCoverage QueryCoverage          `json:"totals_coverage,omitempty"`
	GroupsHasMore  *bool                  `json:"groups_has_more,omitempty"`
	CountsAdditive *bool                  `json:"counts_additive,omitempty"`
	Periods        []QueryPeriod          `json:"periods,omitempty"`
	Comparison     json.RawMessage        `json:"comparison,omitempty"`
	Warnings       []string               `json:"warnings"`
	Error          *ToolError             `json:"error,omitempty"`
	References     []TransactionReference `json:"-"`
	Metrics        ToolExecutionMetrics   `json:"-"`
}

type ConversationRole string

const (
	ConversationRoleSystem    ConversationRole = "system"
	ConversationRoleUser      ConversationRole = "user"
	ConversationRoleAssistant ConversationRole = "assistant"
	ConversationRoleTool      ConversationRole = "tool"
)

// ConversationMessage preserves completed call/result pairs during fallback.
// Only the server creates system/tool messages; client history is untrusted.
type ConversationMessage struct {
	Role       ConversationRole `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []ToolCall       `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
	ToolResult *ToolResult      `json:"tool_result,omitempty"`
}

// ToolTurnInput is not a wire format. State belongs only to the original
// adapter and current request. A fallback must discard it and rebuild Messages.
// The adapter must include required private state when measuring the actual
// serialized upstream request against MaxRequestBytes, despite json:"-" here.
type ToolTurnInput struct {
	MaxInputTokens  int                   `json:"-"`
	MaxOutputTokens int                   `json:"-"`
	Messages        []ConversationMessage `json:"messages"`
	Tools           []ToolDefinition      `json:"tools"`
	State           any                   `json:"-"`
	MaxRequestBytes int                   `json:"-"`
}

// TurnResponse only marks Completed after the protocol's valid end marker.
// Content from tool-bearing turns must not be emitted as a final answer.
type TurnResponse struct {
	OutputLimited bool       `json:"-"`
	Content       string     `json:"content"`
	ToolCalls     []ToolCall `json:"tool_calls,omitempty"`
	Completed     bool       `json:"completed"`
	State         any        `json:"-"`
}

// DisplayEvent is limited to a provider's explicitly displayable summary.
// Raw reasoning, private output items and partial tool arguments are forbidden.
type DisplayEvent struct {
	ThinkingDelta string
}

// DisplayEventSink is synchronous; a send failure must stop upstream work.
type DisplayEventSink func(event DisplayEvent) error

// QueryEvidence is the compact request-local projection of a successful result.
// HTTP maps this to camelCase instead of serializing the tool wire format.
type QueryEvidence struct {
	QueryID   string
	Tool      ToolName
	Scope     QueryScope
	Coverage  QueryCoverage
	QueriedAt string
	Warnings  []string
}

// ToolContinuation is a candidate for an explicit follow-up, never an instruction
// to execute automatically. Only a verified cursor may enter runner input.
type ToolContinuation struct {
	Tool   ToolName `json:"tool"`
	Cursor string   `json:"cursor"`
}
