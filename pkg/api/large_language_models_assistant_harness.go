package api

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"html"
	"io"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/templates"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// Dependencies are immutable for a handler's lifetime. Tests supply isolated
// stores and providers; production resolves the same application containers.
type assistantDependencies struct {
	store             *datastore.DataStore
	primary, fallback provider.LargeLanguageModelToolProvider
	now               func() time.Time
	observe           llm.AssistantObserver
	heartbeat         time.Duration
}

type assistantRequest struct {
	context     core.Context
	cancel      context.CancelFunc
	cancelCause context.CancelCauseFunc
	runner      *llm.AssistantRunner
	input       llm.AssistantRunInput
	now         func() time.Time
	heartbeat   time.Duration
}

func (a *LargeLanguageModelsApi) prepareAssistantRun(c *core.WebContext) (*assistantRequest, *errs.Error) {
	config := a.CurrentConfig()
	if !config.EnableAIAssistant || config.AIAssistantLLMConfig == nil || config.AIAssistantLLMConfig.LLMProvider == "" {
		return nil, errs.ErrAIAssistantNotEnabled
	}
	request, err := readAssistantRequest(c.Request.Body)
	if err != nil {
		return nil, err
	}
	location, timezoneErr := c.GetClientTimezone()
	if timezoneErr != nil {
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}
	// The core fallback uses one generic label for every UTC offset. Give the
	// fixed zone an offset-specific identity so scopes and cursor signatures do
	// not silently equate UTC+08:00 with UTC-05:00.
	if location.String() == "Client Fixed Timezone" {
		_, offset := time.Unix(0, 0).In(location).Zone()
		sign, absolute := "+", offset
		if absolute < 0 {
			sign, absolute = "-", -absolute
		}
		location = time.FixedZone(fmt.Sprintf("UTC%s%02d:%02d", sign, absolute/3600, absolute/60%60), offset)
	}
	uid := c.GetCurrentUid()
	if uid <= 0 {
		return nil, errs.ErrUserNotFound
	}
	runtime := config.GetAIAssistantRuntimeConfig()
	causeParent, cancelCause := context.WithCancelCause(c.Request.Context())
	parent, cancelTimeout := context.WithTimeout(causeParent, time.Duration(runtime.RunTimeout)*time.Millisecond)
	cancel := func() { cancelTimeout(); cancelCause(nil) }
	ctx := core.NewAssistantContext(parent, c.GetContextId(), c.GetClientLocale(), c.ClientIP())
	fail := func(e *errs.Error) (*assistantRequest, *errs.Error) { cancel(); return nil, e }
	if _, userErr := a.users.GetUserById(ctx, uid); userErr != nil {
		return fail(errs.ErrUserNotFound)
	}
	store, now, heartbeat := datastore.Container.UserDataStore, time.Now, 15*time.Second
	observe := llm.AssistantObserver(func(o llm.AssistantObservation) {
		// Only the allowlisted observation DTO is logged, never raw errors or messages.
		encoded, _ := json.Marshal(o)
		log.Infof(ctx, "[assistant] %s", encoded)
	})
	var primary, fallback provider.LargeLanguageModelToolProvider
	if a.assistant != nil {
		store, primary, fallback = a.assistant.store, a.assistant.primary, a.assistant.fallback
		if a.assistant.now != nil {
			now = a.assistant.now
		}
		if a.assistant.observe != nil {
			observe = a.assistant.observe
		}
		if a.assistant.heartbeat > 0 {
			heartbeat = a.assistant.heartbeat
		}
	} else {
		var providerErr error
		primary, fallback, providerErr = llm.Container.GetAIAssistantToolProviders()
		if providerErr != nil {
			return fail(assistantRunError(llm.AssistantRunResult{}, providerErr))
		}
	}
	query, queryErr := services.NewAssistantQueryService(store, uid, services.AssistantQueryOptions{Timezone: location, Secret: []byte(config.SecretKey), Now: now, QueryTimeout: time.Duration(runtime.QueryTimeout) * time.Millisecond})
	if queryErr != nil {
		return fail(errs.ErrAIAssistantFailed)
	}
	var continuation *data.ToolContinuation
	var scope *data.QueryScope
	if request.Continuation != nil {
		continuation = &data.ToolContinuation{Tool: data.ToolName(request.Continuation.Tool), Cursor: request.Continuation.Cursor}
		scope, queryErr = query.DescribeContinuation(ctx, *continuation)
		if queryErr != nil {
			return fail(assistantRunError(llm.AssistantRunResult{}, queryErr))
		}
	}
	messages, promptErr := buildAssistantMessages(request, now().In(location), ctx.GetClientLocale(), continuation, scope)
	if promptErr != nil {
		return fail(errs.ErrAIAssistantFailed)
	}
	return &assistantRequest{context: ctx, cancel: cancel, cancelCause: cancelCause, runner: llm.NewAssistantRunner(primary, fallback, query, now), now: now, heartbeat: heartbeat,
		input: llm.AssistantRunInput{UID: uid, Mode: request.Mode, Messages: messages, Continuation: continuation, MaxRounds: runtime.MaxRounds, RunTimeout: time.Duration(runtime.RunTimeout) * time.Millisecond, QueryTimeout: time.Duration(runtime.QueryTimeout) * time.Millisecond, Observe: observe}}, nil
}

func readAssistantRequest(body io.Reader) (*models.AIAssistantChatRequest, *errs.Error) {
	raw, err := io.ReadAll(io.LimitReader(body, 256*1024+1))
	if err != nil || len(raw) > 256*1024 {
		return nil, errs.ErrParameterInvalid
	}
	var request *models.AIAssistantChatRequest
	if err := json.Unmarshal(raw, &request, json.RejectUnknownMembers(true)); err != nil || request == nil {
		return nil, errs.ErrParameterInvalid
	}
	mode, err := normalizeAIAssistantMode(request.Mode)
	if err != nil {
		return nil, errs.ErrAIAssistantInvalidMode
	}
	request.Mode = mode
	if utf8.RuneCountInString(request.Message) > 2048 || len(request.History) > 20 {
		return nil, errs.ErrParameterInvalid
	}
	if mode == models.AIAssistantModeChat && strings.TrimSpace(request.Message) == "" {
		return nil, errs.ErrAIAssistantMessageIsEmpty
	}
	for _, item := range request.History {
		if item == nil || (item.Role != "user" && item.Role != "assistant") || strings.TrimSpace(item.Content) == "" || utf8.RuneCountInString(item.Content) > 2048 {
			return nil, errs.ErrParameterInvalid
		}
	}
	if request.Continuation != nil && (request.Continuation.Cursor == "" || len(request.Continuation.Cursor) > 65536) {
		return nil, errs.ErrAIAssistantInvalidCursor
	}
	return request, nil
}

// Only complete user/assistant exchanges enter history. In particular trimming
// cannot leave an assistant statement detached from its initiating question.
func assistantHistory(history []*models.AIAssistantHistoryItem) []data.ConversationMessage {
	var pairs []data.ConversationMessage
	for i := 0; i+1 < len(history); i++ {
		if history[i].Role == "user" && history[i+1].Role == "assistant" {
			pairs = append(pairs, data.ConversationMessage{Role: data.ConversationRoleUser, Content: history[i].Content}, data.ConversationMessage{Role: data.ConversationRoleAssistant, Content: history[i+1].Content})
			i++
		}
	}
	if len(pairs) > llm.AssistantMaxHistoryMessages {
		pairs = pairs[len(pairs)-llm.AssistantMaxHistoryMessages:]
	}
	return pairs
}

func assistantCalendarShift(now time.Time, years, months int) time.Time {
	month := time.Date(now.Year()+years, now.Month()+time.Month(months), 1, now.Hour(), now.Minute(), now.Second(), 0, now.Location())
	last := time.Date(month.Year(), month.Month()+1, 0, now.Hour(), now.Minute(), now.Second(), 0, now.Location()).Day()
	day := now.Day()
	if day > last {
		day = last
	}
	return time.Date(month.Year(), month.Month(), day, now.Hour(), now.Minute(), now.Second(), 0, now.Location())
}

func buildAssistantMessages(request *models.AIAssistantChatRequest, now time.Time, locale string, candidate *data.ToolContinuation, scope *data.QueryScope) ([]data.ConversationMessage, error) {
	tmpl, err := templates.GetTemplate(templates.SYSTEM_PROMPT_PERSONAL_FINANCE_ASSISTANT)
	if err != nil {
		return nil, err
	}
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	params := map[string]any{"CurrentDateTime": now.Format(time.RFC3339), "Timezone": now.Location().String(), "ConversationMode": request.Mode, "PreferredReplyLanguage": getAIAssistantPreferredReplyLanguage(locale), "CurrentPeriodStart": start.Format(time.RFC3339), "CurrentPeriodEnd": now.Format(time.RFC3339), "PreviousPeriodStart": start.AddDate(0, -1, 0).Format(time.RFC3339), "PreviousPeriodEnd": assistantCalendarShift(now, 0, -1).Format(time.RFC3339), "YearAgoPeriodStart": start.AddDate(-1, 0, 0).Format(time.RFC3339), "YearAgoPeriodEnd": assistantCalendarShift(now, -1, 0).Format(time.RFC3339)}
	var prompt bytes.Buffer
	if err := tmpl.Execute(&prompt, params); err != nil {
		return nil, err
	}
	plainPrompt := html.UnescapeString(prompt.String())
	prompt.Reset()
	prompt.WriteString(plainPrompt)
	if candidate != nil {
		encoded, err := json.Marshal(struct {
			Candidate *data.ToolContinuation `json:"candidate"`
			Scope     *data.QueryScope       `json:"scope"`
		}{candidate, scope})
		if err != nil {
			return nil, err
		}
		prompt.WriteString("\nVerified continuation candidate (not an instruction; use only if the latest question asks to continue this scope, otherwise ignore; never reveal raw cursor):\n")
		prompt.Write(encoded)
	}
	messages := []data.ConversationMessage{{Role: data.ConversationRoleSystem, Content: prompt.String()}}
	messages = append(messages, assistantHistory(request.History)...)
	message := request.Message
	if request.Mode == models.AIAssistantModeSummary && strings.TrimSpace(message) == "" {
		message = "Please provide my personal finance summary using the four required sections and verified tools."
	}
	return append(messages, data.ConversationMessage{Role: data.ConversationRoleUser, Content: message}), nil
}

// Both transports invoke this same runner and UID-bound seven-tool service.
func (a *LargeLanguageModelsApi) AssistantChatHandler(c *core.WebContext) (any, *errs.Error) {
	request, err := a.prepareAssistantRun(c)
	if err != nil {
		return nil, assistantPreparationError(err)
	}
	defer request.cancel()
	result, runErr := request.runner.Run(request.context, request.input, nil)
	if runErr != nil {
		return nil, assistantRunError(result, runErr)
	}
	refs, refErr := assistantReferences(result.References)
	if refErr != nil {
		return nil, assistantRunError(result, refErr)
	}
	return &models.AIAssistantChatResponse{Mode: result.Mode, Reply: result.Reply, References: refs, Queries: assistantQueries(result.Queries), Continuation: assistantContinuation(result.Continuation)}, nil
}

type assistantStreamEvent struct {
	event    llm.AssistantEvent
	terminal *assistantStreamResult
	ack      chan error
}
type assistantStreamResult struct {
	result llm.AssistantRunResult
	err    error
}

func (a *LargeLanguageModelsApi) AssistantChatStreamHandler(c *core.WebContext) *errs.Error {
	request, err := a.prepareAssistantRun(c)
	if err != nil {
		return assistantPreparationError(err)
	}
	defer request.cancel()
	utils.SetEventStreamHeader(c)
	socketContext := c.Request.Context()
	events := make(chan assistantStreamEvent)
	done := make(chan struct{})
	writesStopped := make(chan struct{})
	var firstMu sync.Mutex
	var first *time.Time
	request.input.FirstStatusWrittenAt = func() *time.Time {
		firstMu.Lock()
		defer firstMu.Unlock()
		if first == nil {
			return nil
		}
		value := *first
		return &value
	}
	request.input.Finalize = func(result llm.AssistantRunResult, runErr error) error {
		if socketContext.Err() != nil || errors.Is(context.Cause(request.context), llm.ErrAssistantWriteFailed) {
			return nil
		}
		delivery := assistantStreamEvent{terminal: &assistantStreamResult{result, runErr}, ack: make(chan error, 1)}
		select {
		case events <- delivery:
			return <-delivery.ack
		case <-socketContext.Done():
			return nil
		case <-writesStopped:
			return nil
		}
	}
	go func() {
		defer close(done)
		_, _ = request.runner.Run(request.context, request.input, func(event llm.AssistantEvent) error {
			delivery := assistantStreamEvent{event: event, ack: make(chan error, 1)}
			select {
			case events <- delivery:
			case <-request.context.Done():
				return request.context.Err()
			}
			// Every delivered event is acknowledged by the sole writer, even on failure.
			return <-delivery.ack
		})
	}()
	ticker := time.NewTicker(request.heartbeat)
	defer ticker.Stop()
	stop := func() {
		close(writesStopped)
		request.cancelCause(llm.ErrAssistantWriteFailed)
		request.cancel()
		<-done
	}
	for {
		select {
		case delivery := <-events:
			var writeErr error
			if delivery.terminal != nil {
				writeErr = writeAssistantTerminal(c, *delivery.terminal)
				// No heartbeat or second frame can follow a terminal, including a failed
				// terminal write. Its acknowledgment lets the runner finalize telemetry.
				ticker.Stop()
				delivery.ack <- writeErr
				if writeErr != nil {
					stop()
				} else {
					<-done
				}
				return nil
			}
			event := delivery.event
			refs, projectionErr := assistantReferences(event.References)
			writeErr = projectionErr
			if writeErr == nil {
				writeErr = utils.WriteEventStreamJSON(c, &models.AIAssistantChatStreamChunk{Type: string(event.Type), Status: models.AIAssistantStatus(event.Status), Delta: event.Delta, References: refs})
			}
			if writeErr == nil && event.Type == llm.AssistantEventStatus {
				firstMu.Lock()
				if first == nil {
					at := request.now()
					first = &at
				}
				firstMu.Unlock()
			}
			delivery.ack <- writeErr
			if writeErr != nil {
				stop()
				return nil
			}
		case <-ticker.C:
			if writeErr := utils.WriteEventStreamComment(c, "heartbeat"); writeErr != nil {
				stop()
				return nil
			}
		case <-socketContext.Done():
			close(writesStopped)
			request.cancel()
			<-done
			return nil
		case <-done:
			return nil
		}
	}
}

func writeAssistantTerminal(c *core.WebContext, completed assistantStreamResult) error {
	result := completed.result
	refs, projectionErr := assistantReferences(result.References)
	if completed.err == nil {
		completed.err = projectionErr
	}
	if completed.err != nil {
		apiErr := assistantRunError(result, completed.err)
		c.SetResponseError(apiErr)
		envelope := utils.GetJsonErrorResult(apiErr, c.Request.URL.Path)
		envelope["context"] = apiErr.Context
		return utils.WriteEventStreamJSON(c, envelope)
	}
	return utils.WriteEventStreamJSON(c, &models.AIAssistantChatStreamChunk{Type: models.AIAssistantChatStreamChunkTypeDone, Mode: result.Mode, Reply: result.Reply, Thinking: result.Thinking, References: refs, Queries: assistantQueries(result.Queries), Continuation: assistantContinuation(result.Continuation)})
}

func assistantContinuation(c *data.ToolContinuation) *models.AIAssistantContinuation {
	if c == nil {
		return nil
	}
	return &models.AIAssistantContinuation{Tool: string(c.Tool), Cursor: c.Cursor}
}
func assistantQueries(queries []data.QueryEvidence) []models.AIAssistantQuery {
	result := make([]models.AIAssistantQuery, 0, len(queries))
	for _, q := range queries {
		s := q.Scope
		scope := models.AIAssistantQueryScope{StartTime: s.StartTime, EndTime: s.EndTime, Timezone: s.Timezone, Currency: s.Currency, Types: s.Types, AccountIDs: s.AccountIDs, CategoryIDs: s.CategoryIDs, TagIDs: s.TagIDs, TagMatch: s.TagMatch, NoTags: s.NoTags, TransferSide: s.TransferSide, AmountMin: s.AmountMin, AmountMax: s.AmountMax, Keyword: s.Keyword, IncludesHidden: s.IncludesHidden, ExcludesDeleted: s.ExcludesDeleted}
		for _, p := range s.Periods {
			scope.Periods = append(scope.Periods, models.AIAssistantQueryPeriod{Name: p.Name, StartTime: p.StartTime, EndTime: p.EndTime, MatchedCount: p.MatchedCount})
		}
		warnings := q.Warnings
		if warnings == nil {
			warnings = []string{}
		}
		result = append(result, models.AIAssistantQuery{QueryID: q.QueryID, Tool: string(q.Tool), Scope: scope, Coverage: string(q.Coverage), QueriedAt: q.QueriedAt, Warnings: warnings})
	}
	return result
}

// Tool amounts are exact two-decimal strings. Do not use float64 or ParseAmount
// (whose multiplication can overflow) at the legacy minor-unit boundary.
func assistantMinorAmount(amount string) (int64, error) {
	negative := strings.HasPrefix(amount, "-")
	digits := strings.TrimPrefix(amount, "-")
	parts := strings.Split(digits, ".")
	if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) != 2 {
		return 0, errs.ErrAIAssistantFailed
	}
	for _, r := range parts[0] + parts[1] {
		if r < '0' || r > '9' {
			return 0, errs.ErrAIAssistantFailed
		}
	}
	number, ok := new(big.Int).SetString(parts[0]+parts[1], 10)
	if !ok {
		return 0, errs.ErrAIAssistantFailed
	}
	if negative {
		number.Neg(number)
	}
	if !number.IsInt64() {
		return 0, errs.ErrAIAssistantFailed
	}
	return number.Int64(), nil
}
func assistantReferences(refs []data.TransactionReference) ([]*models.AIAssistantReferencedTransaction, error) {
	result := make([]*models.AIAssistantReferencedTransaction, 0, len(refs))
	for _, ref := range refs {
		id, err := strconv.ParseInt(ref.ID, 10, 64)
		if err != nil || id <= 0 {
			return nil, errs.ErrAIAssistantFailed
		}
		at, err := time.Parse(time.RFC3339, ref.Time)
		if err != nil {
			return nil, errs.ErrAIAssistantFailed
		}
		amount, err := assistantMinorAmount(ref.SourceAmount)
		if err != nil {
			return nil, err
		}
		var destination int64
		if ref.DestinationAmount != "" {
			destination, err = assistantMinorAmount(ref.DestinationAmount)
			if err != nil {
				return nil, err
			}
		}
		var kind models.TransactionType
		switch ref.Type {
		case "income":
			kind = models.TRANSACTION_TYPE_INCOME
		case "expense":
			kind = models.TRANSACTION_TYPE_EXPENSE
		case "transfer":
			kind = models.TRANSACTION_TYPE_TRANSFER
		case "balance_modification":
			kind = models.TRANSACTION_TYPE_MODIFY_BALANCE
		default:
			return nil, errs.ErrAIAssistantFailed
		}
		result = append(result, &models.AIAssistantReferencedTransaction{Id: id, Time: at.Unix(), TimeText: at.In(time.FixedZone("", int(ref.TimezoneOffset)*60)).Format("2006-01-02 15:04:05"), Type: kind, CategoryName: ref.CategoryName, SourceAccountName: ref.SourceAccountName, DestinationAccountName: ref.DestinationAccountName, SourceAmount: amount, DestinationAmount: destination, Currency: ref.Currency, DestinationCurrency: ref.DestinationCurrency, Comment: ref.Comment, HideAmount: ref.HideAmount})
	}
	return result, nil
}

func assistantPreparationError(err *errs.Error) *errs.Error {
	if err.Context != nil {
		return err
	}
	return errs.NewErrorWithContext(err, core.O{"assistant": models.AIAssistantErrorContext{State: models.AIAssistantStateFailed, Queries: []models.AIAssistantQuery{}}})
}
func assistantRunError(result llm.AssistantRunResult, err error) *errs.Error {
	base := errs.ErrAIAssistantFailed
	var toolErr *data.ToolError
	if errors.As(err, &toolErr) {
		switch toolErr.Code {
		case data.ToolErrorInvalidArgument:
			base = errs.ErrAIAssistantInvalidArgument
		case data.ToolErrorInvalidCursor:
			base = errs.ErrAIAssistantInvalidCursor
		case data.ToolErrorNotFoundOrNotAccessible:
			base = errs.ErrAIAssistantNotAccessible
		case data.ToolErrorProviderUnsupported:
			base = errs.ErrAIAssistantProviderUnsupported
		case data.ToolErrorBudgetExhausted, data.ToolErrorResultTooLarge:
			base = errs.ErrAIAssistantIncomplete
		case data.ToolErrorQueryTimeout:
			base = errs.ErrAIAssistantTimeout
		}
	}
	state := models.AIAssistantStateFailed
	if result.State == llm.AssistantStateIncomplete || errors.Is(err, context.DeadlineExceeded) {
		state = models.AIAssistantStateIncomplete
	}
	if errors.Is(err, context.DeadlineExceeded) {
		base = errs.ErrAIAssistantTimeout
	}
	if errors.Is(err, context.Canceled) {
		base = errs.ErrAIAssistantCanceled
	}
	return errs.NewErrorWithContext(base, core.O{"assistant": models.AIAssistantErrorContext{State: state, Queries: assistantQueries(result.Queries), Continuation: assistantContinuation(result.Continuation)}})
}
