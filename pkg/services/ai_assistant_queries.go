package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
)

// AssistantQueryOptions contains only trusted server/request configuration.
type AssistantQueryOptions struct {
	Timezone       *time.Location
	Secret         []byte
	Now            func() time.Time
	QueryTimeout   time.Duration
	MaxResultBytes int
	Observe        func(AssistantQueryObservation)
}

// AssistantQueryObservation deliberately excludes SQL, arguments, cursor text,
// financial payloads and arbitrary error strings.
type AssistantQueryObservation struct {
	RequestID string
	Tool      data.ToolName
	Metrics   data.ToolExecutionMetrics
	ErrorCode data.ToolErrorCode
}
type AssistantQueryService struct {
	database       *datastore.Database
	uid            int64
	uidString      string
	location       *time.Location
	now            func() time.Time
	cursorKey      []byte
	timeout        time.Duration
	maxResultBytes int
	resultFits     func(data.ToolResult) bool
	observe        func(AssistantQueryObservation)
	handlers       map[data.ToolName]assistantToolHandler
}
type assistantToolHandler func(*assistantQueryExecution) (data.ToolResult, error)
type assistantQueryExecution struct {
	QueryID         string
	QueriedAt       string
	Service         *AssistantQueryService
	Context         core.Context
	View            datastore.AssistantReadView
	Tool            data.ToolName
	Args            assistantQueryArgs
	Cursor          *assistantCursor
	Metadata        *assistantMetadata
	AccountIDs      []string
	CategoryIDs     []string
	HierarchyDigest string
	Scope           *data.QueryScope
}

var assistantToolNames = [...]data.ToolName{data.ToolListAccounts, data.ToolListCategories, data.ToolListTags, data.ToolQueryTransactions, data.ToolGetTransaction, data.ToolAggregateTransactions, data.ToolGetAccountBalances}

func NewAssistantQueryService(store *datastore.DataStore, uid int64, options AssistantQueryOptions) (*AssistantQueryService, error) {
	if store == nil || store.Count() == 0 || uid <= 0 || len(options.Secret) == 0 {
		return nil, fmt.Errorf("assistant query service requires datastore, authenticated UID and server secret")
	}
	if options.Timezone == nil {
		options.Timezone = time.UTC
	}
	if options.Timezone.String() == "" {
		_, offset := time.Unix(0, 0).In(options.Timezone).Zone()
		sign := "+"
		absolute := offset
		if absolute < 0 {
			sign = "-"
			absolute = -absolute
		}
		options.Timezone = time.FixedZone(fmt.Sprintf("UTC%s%02d:%02d:%02d", sign, absolute/3600, absolute/60%60, absolute%60), offset)
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.QueryTimeout == 0 {
		options.QueryTimeout = 30 * time.Second
	}
	if options.MaxResultBytes == 0 {
		options.MaxResultBytes = 64 * 1024
	}
	if options.QueryTimeout < 0 || options.MaxResultBytes < 1024 {
		return nil, fmt.Errorf("invalid assistant query limits")
	}
	s := &AssistantQueryService{database: store.Choose(uid), uid: uid, uidString: strconv.FormatInt(uid, 10), location: options.Timezone, now: options.Now, cursorKey: assistantCursorKey(options.Secret, assistantCursorPurpose), timeout: options.QueryTimeout, maxResultBytes: options.MaxResultBytes, observe: options.Observe, handlers: map[data.ToolName]assistantToolHandler{}}
	for _, factory := range []func(*AssistantQueryService) map[data.ToolName]assistantToolHandler{newAssistantTransactionTools, newAssistantAggregateTools, newAssistantMetadataTools} {
		for name, handler := range factory(s) {
			if assistantToolDomain(name) == nil || s.handlers[name] != nil || handler == nil {
				return nil, fmt.Errorf("invalid assistant tool registry")
			}
			s.handlers[name] = handler
		}
	}
	if len(s.handlers) != len(assistantToolNames) {
		return nil, fmt.Errorf("incomplete assistant tool registry")
	}
	return s, nil
}
func (s *AssistantQueryService) Definitions() []data.ToolDefinition {
	descriptions := map[data.ToolName]string{
		data.ToolListAccounts:          "Discover account IDs/names/hierarchy/hidden state, ordered by ID. Paginated; resolve ambiguous names before filtering.",
		data.ToolListCategories:        "Discover income/expense/transfer category IDs/names/hierarchy, including hidden categories. Names may repeat. Paginated.",
		data.ToolListTags:              "Discover tag IDs/names/groups, including hidden tags. Keyword searches names. Paginated.",
		data.ToolQueryTransactions:     "Search logical transactions, newest first. Paginated details, NOT totals. Transfers appear once with both sides. Amount order requires currency and explicit transfer side.",
		data.ToolGetTransaction:        "Read full effective transaction by string ID. Incoming-transfer IDs resolve to canonical outgoing records with both sides. Inaccessible/deleted IDs fail.",
		data.ToolAggregateTransactions: "Exact full-scope statistics and named-period comparisons. Complete totals with paginated groups; separate currencies/flows. Overlapping group counts are not additive.",
		data.ToolGetAccountBalances:    "CURRENT signed leaf-account balances (including hidden); paginated details, complete currency totals. No parent double-counting, historical reconstruction or conversion.",
	}
	definitions := make([]data.ToolDefinition, 0, len(assistantToolNames))
	for _, name := range assistantToolNames {
		d := assistantToolDomain(name)
		d.Required = true
		schema, _ := json.Marshal(d.schema())
		definitions = append(definitions, data.ToolDefinition{Name: name, Description: descriptions[name], Parameters: schema})
	}
	return definitions
}
func assistantSafeError(err error) *data.ToolError {
	var controlled *data.ToolError
	if errors.As(err, &controlled) {
		return controlled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return &data.ToolError{Code: data.ToolErrorQueryTimeout, Message: "Query deadline exceeded", Retryable: true}
	}
	if errors.Is(err, context.Canceled) {
		return &data.ToolError{Code: data.ToolErrorCanceled, Message: "Query canceled"}
	}
	return &data.ToolError{Code: data.ToolErrorDataInconsistent, Message: "Unable to read a consistent ledger"}
}
func assistantFailure(err error) data.ToolResult {
	return data.ToolResult{OK: false, Error: assistantSafeError(err), Warnings: []string{}}
}

// NormalizeToolArguments is an optional request-local cache-key capability.
// It performs no reads and confers no authorization. Cursor position is part
// of the key: equal filters on two different pages must never reuse a result.
func (s *AssistantQueryService) NormalizeToolArguments(call data.ToolCall) (json.RawMessage, error) {
	args, cursor, err := s.parseArguments(call.Name, call.Arguments)
	if err != nil {
		return nil, err
	}
	return assistantCanonicalJSON(struct {
		Tool   data.ToolName      `json:"tool"`
		Args   assistantQueryArgs `json:"args"`
		Cursor *assistantCursor   `json:"cursor,omitempty"`
	}{call.Name, args, cursor})
}

func (s *AssistantQueryService) Execute(c core.Context, call data.ToolCall) (result data.ToolResult) {
	started := time.Now()
	var read int64
	defer func() {
		result.Metrics.RowsRead = read
		result.Metrics.ResultRecords = int64(result.ReturnedCount)
		serialized, _ := json.Marshal(result)
		result.Metrics.SerializedBytes = int64(len(serialized))
		result.Metrics.Duration = time.Since(started)
		if s.observe != nil {
			observation := AssistantQueryObservation{Tool: call.Name, Metrics: result.Metrics}
			if c != nil {
				observation.RequestID = c.GetContextId()
			}
			if result.Error != nil {
				observation.ErrorCode = result.Error.Code
			}
			s.observe(observation)
		}
	}()
	handler := s.handlers[call.Name]
	if handler == nil {
		return assistantFailure(assistantInvalid("tool"))
	}
	args, cursor, err := s.parseArguments(call.Name, call.Arguments)
	if err != nil {
		return assistantFailure(err)
	}
	if c == nil {
		c = core.NewNullContext()
	}
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()
	qc := core.NewAssistantContext(ctx, c.GetContextId(), c.GetClientLocale(), c.ClientIP())
	queriedAt := s.now().UTC().Format(time.RFC3339)
	queryID := make([]byte, 16)
	if _, err := rand.Read(queryID); err != nil {
		return assistantFailure(err)
	}
	queryIDText := hex.EncodeToString(queryID)
	err = s.database.WithAssistantReadView(qc, func(view datastore.AssistantReadView) error {
		defer func() { read = view.Observations().RowsRead }()
		execution := &assistantQueryExecution{Service: s, Context: qc, View: view, Tool: call.Name, Args: args, Cursor: cursor, QueryID: queryIDText, QueriedAt: queriedAt}
		if err := execution.prepare(); err != nil {
			return err
		}
		var err error
		result, err = handler(execution)
		if err != nil {
			return err
		}
		if !result.OK {
			return nil
		}
		// Count the actual business projection, independently of handler claims.
		if len(result.Data) > 0 {
			var projection any
			if err := json.Unmarshal(result.Data, &projection); err != nil {
				return assistantInconsistent()
			}
			switch value := projection.(type) {
			case []any:
				result.ReturnedCount = len(value)
			case map[string]any:
				result.ReturnedCount = 1
			case nil:
				result.ReturnedCount = 0
			default:
				return assistantInconsistent()
			}
		} else {
			result.ReturnedCount = 0
		}
		result.Scope = execution.Scope
		result.QueriedAt = queriedAt
		if result.QueryID == "" {
			result.QueryID = queryIDText
		}
		if result.Warnings == nil {
			result.Warnings = []string{}
		}
		raw, err := json.Marshal(result)
		if err != nil {
			return err
		}
		if len(raw) > s.maxResultBytes || (s.resultFits != nil && !s.resultFits(result)) {
			return &data.ToolError{Code: data.ToolErrorResultTooLarge, Message: "Tool result exceeds the result limit"}
		}
		return nil
	})
	if err != nil {
		return assistantFailure(err)
	}
	return result
}

// ValidateContinuation authenticates and re-authorizes a candidate, without
// executing the tool or automatically applying it to a new question.
func (s *AssistantQueryService) ValidateContinuation(c core.Context, candidate data.ToolContinuation) error {
	_, err := s.DescribeContinuation(c, candidate)
	return err
}

// DescribeContinuation returns only the current verified scope, for candidate
// context in a subsequent model request. It never invokes a tool handler.
func (s *AssistantQueryService) DescribeContinuation(c core.Context, candidate data.ToolContinuation) (*data.QueryScope, error) {
	if s.handlers[candidate.Tool] == nil || candidate.Cursor == "" {
		return nil, assistantCursorError()
	}
	raw, _ := json.Marshal(map[string]any{"cursor": candidate.Cursor})
	args, cursor, err := s.parseArguments(candidate.Tool, raw)
	if err != nil {
		return nil, err
	}
	if cursor == nil {
		return nil, assistantCursorError()
	}
	if c == nil {
		c = core.NewNullContext()
	}
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()
	qc := core.NewAssistantContext(ctx, c.GetContextId(), c.GetClientLocale(), c.ClientIP())
	var scope *data.QueryScope
	err = s.database.WithAssistantReadView(qc, func(v datastore.AssistantReadView) error {
		execution := &assistantQueryExecution{Service: s, Context: qc, View: v, Tool: candidate.Tool, Args: args, Cursor: cursor}
		if err := execution.prepare(); err != nil {
			return err
		}
		scope = execution.Scope
		return nil
	})
	if err != nil {
		return nil, assistantSafeError(err)
	}
	return scope, nil
}
func assistantToolNotImplemented(*assistantQueryExecution) (data.ToolResult, error) {
	return data.ToolResult{}, &data.ToolError{Code: data.ToolErrorProviderUnsupported, Message: "Internal tool implementation is not installed"}
}
