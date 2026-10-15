package api

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
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
	"html"
	"io"
	"strings"
	"time"
	"unicode/utf8"
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
