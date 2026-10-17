package services

import (
	"encoding/json"
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider/openai"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

type contextSummaryProvider struct {
	calls   int
	before  func()
	sources []string
}

func (p *contextSummaryProvider) RunToolTurn(_ core.Context, _ int64, input data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
	p.calls++
	raw, _ := json.Marshal(input.Messages)
	p.sources = append(p.sources, string(raw))
	if p.before != nil {
		p.before()
	}
	return data.TurnResponse{Completed: true, Content: `{"topic":"budget","confirmedConstraints":["user: dining budget CNY 3000; checking account 42; date 2026-09; corrected CNY 2500"],"resolvedReferences":[],"openQuestions":[]}`}, nil
}
func TestAssistantContextThirtyTurnsTwoCompactions(t *testing.T) {
	conv := assistantConversationFixture(t, settings.Sqlite3DbType)
	ctx := core.NewNullContext()
	cfg := &settings.Config{EnableAIAssistant: true}
	service := NewAssistantContextService(conv.storage, func() *settings.Config { return cfg })
	provider := &contextSummaryProvider{}
	var last models.AssistantCommandResult
	for i := 0; i < 32; i++ {
		question := fmt.Sprintf("question %d", i)
		if i == 0 {
			question = "dining budget CNY 3000, account 42, date 2026-09"
		}
		if i == 8 {
			question = "correct budget to CNY 2500"
		}
		last = conversationSend(t, conv, fmt.Sprintf("turn-%d", i), last.ConversationId, last.ContextVersion, question)
		if i == 15 || i == 31 {
			var run *models.AssistantMessage
			e := conv.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
				rid, _ := assistantParseID(last.RunId)
				var e error
				run, e = assistantConversationMessage(tx.Session, 7, rid)
				return e
			})
			if e != nil {
				t.Fatal(e)
			}
			snap, e := service.snapshot(ctx, run)
			if e != nil {
				t.Fatal(e)
			}
			if len(snap.Turns) != i {
				t.Fatalf("lost chain: %d wanted %d", len(snap.Turns), i)
			}
			changed, _, e := service.compact(ctx, run, snap, provider, 16000, false)
			if e != nil || !changed {
				t.Fatal(changed, e)
			}
			if !strings.Contains(snap.Conversation.Summary, "CNY 2500") {
				t.Fatal("correction missing")
			}
		}
		conversationComplete(t, conv, last, "verified answer")
	}
	if provider.calls != 2 {
		t.Fatal(provider.calls)
	}
	if !strings.Contains(provider.sources[1], "CNY 2500") {
		t.Fatal("second compaction lost old summary")
	}
	var count int64
	_ = conv.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		count, _ = tx.Session.Where("uid = ?", 7).Count(&models.AssistantMessage{})
		return nil
	})
	if count != 64 {
		t.Fatal("original history changed", count)
	}
}
func TestAssistantContextCompactionCASRejectsChangedVersion(t *testing.T) {
	conv := assistantConversationFixture(t, settings.Sqlite3DbType)
	ctx := core.NewNullContext()
	var last models.AssistantCommandResult
	for i := 0; i < 3; i++ {
		last = conversationSend(t, conv, fmt.Sprint(i), last.ConversationId, last.ContextVersion, "question")
		conversationComplete(t, conv, last, "answer")
	}
	last = conversationSend(t, conv, "pending", last.ConversationId, last.ContextVersion, "next")
	var run *models.AssistantMessage
	_ = conv.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		id, _ := assistantParseID(last.RunId)
		run, _ = assistantConversationMessage(tx.Session, 7, id)
		return nil
	})
	service := NewAssistantContextService(conv.storage, func() *settings.Config { return &settings.Config{} })
	snap, e := service.snapshot(ctx, run)
	if e != nil {
		t.Fatal(e)
	}
	provider := &contextSummaryProvider{before: func() {
		_ = conv.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
			c, _ := assistantConversationLoad(tx.Session, 7, run.ConversationId)
			c.ContextVersion++
			return assistantConversationSave(tx, c)
		})
	}}
	if _, _, e = service.compact(ctx, run, snap, provider, 16000, false); e != models.AssistantReasonContextChanged {
		t.Fatal(e)
	}
}
func TestAssistantContextConservativeWindow(t *testing.T) {
	cfg := settings.DefaultAIAssistantExperienceConfig()
	cfg.FallbackContextTokens = 8192
	if got := assistantContextBudget(cfg); got != 3072 {
		t.Fatal(got)
	}
	if !assistantContextFits([]data.ConversationMessage{{Role: data.ConversationRoleUser, Content: strings.Repeat("中", 1100)}}, nil, 3072) {
		t.Fatal("UTF8 bytes must not be treated as tokens")
	}
}

func TestAssistantPrepareUsesAcceptedPolicyAndLatestQuestion(t *testing.T) {
	t.Chdir("../..")
	conv := assistantConversationFixture(t, settings.Sqlite3DbType)
	last := conversationSend(t, conv, "prepare", "", 0, "请详细解释，而不是使用旧的简短偏好")
	var run *models.AssistantMessage
	_ = conv.storage.WithUserTransaction(core.NewNullContext(), 7, func(tx *AssistantStorageTx) error {
		id, _ := assistantParseID(last.RunId)
		run, _ = assistantConversationMessage(tx.Session, 7, id)
		run.Deadline = time.Now().Add(time.Minute).UnixMilli()
		return nil
	})
	cfg := &settings.Config{EnableAIAssistant: true, SecretKey: "test"}
	service := NewAssistantContextService(conv.storage, func() *settings.Config { return cfg })
	p := &contextAnswerProvider{}
	service.providers = func() (provider.LargeLanguageModelToolProvider, provider.LargeLanguageModelToolProvider, error) {
		return p, nil, nil
	}
	prepared, e := service.Prepare(core.NewNullContext(), run)
	if e != nil {
		t.Fatal(e)
	}
	if prepared.Input.MaxInputTokens != 16000 || prepared.Input.MaxOutputTokens != 4096 {
		t.Fatal(prepared.Input)
	}
	result, e := prepared.Runner.Run(core.NewNullContext(), prepared.Input, nil)
	if e != nil || result.State != llm.AssistantStateCompleted {
		t.Fatal(result, e)
	}
	if p.last[len(p.last)-1].Content != "请详细解释，而不是使用旧的简短偏好" {
		t.Fatal("latest user instruction not final", p.last)
	}
}

type contextAnswerProvider struct{ last []data.ConversationMessage }

func (p *contextAnswerProvider) RunToolTurn(_ core.Context, _ int64, input data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
	p.last = input.Messages
	return data.TurnResponse{Completed: true, Content: "A detailed supported explanation."}, nil
}

func TestAssistantPrepareActualWireFitsDefaultBudget(t *testing.T) {
	t.Chdir("../..")
	conv := assistantConversationFixture(t, settings.Sqlite3DbType)
	current := conversationSend(t, conv, "wire", "", 0, "你好，请解释预算规划")
	var run *models.AssistantMessage
	_ = conv.storage.WithUserTransaction(core.NewNullContext(), 7, func(tx *AssistantStorageTx) error {
		id, _ := assistantParseID(current.RunId)
		run, _ = assistantConversationMessage(tx.Session, 7, id)
		run.Deadline = time.Now().Add(time.Minute).UnixMilli()
		return nil
	})
	wireBytes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		wireBytes = len(raw)
		var body map[string]any
		if json.Unmarshal(raw, &body) != nil {
			t.Error("invalid body")
		}
		if body["max_completion_tokens"] != float64(4096) {
			t.Error("output budget missing")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"A general budgeting explanation."}}]}`)
	}))
	defer server.Close()
	cfg := &settings.Config{EnableAIAssistant: true, SecretKey: "test"}
	service := NewAssistantContextService(conv.storage, func() *settings.Config { return cfg })
	p := openai.NewOpenAICompatibleLargeLanguageModelProvider(&settings.LLMConfig{OpenAICompatibleBaseURL: server.URL, OpenAICompatibleModelID: "fixture", LargeLanguageModelAPIRequestTimeout: 2000, LargeLanguageModelAPIProxy: "none"}, false).(provider.LargeLanguageModelToolProvider)
	service.providers = func() (provider.LargeLanguageModelToolProvider, provider.LargeLanguageModelToolProvider, error) {
		return p, nil, nil
	}
	prepared, e := service.Prepare(core.NewNullContext(), run)
	if e != nil {
		t.Fatal(e)
	}
	result, e := prepared.Runner.Run(core.NewNullContext(), prepared.Input, nil)
	if e != nil || result.State != llm.AssistantStateCompleted {
		t.Fatal(result, e)
	}
	if wireBytes <= 0 || wireBytes > prepared.Input.MaxInputTokens {
		t.Fatal("default request unavailable", wireBytes)
	}
	t.Log("actual request UTF8 bytes", wireBytes)
}
func TestAssistantCompactToolSchemaPreservesEveryConstraint(t *testing.T) {
	conv := assistantConversationFixture(t, settings.Sqlite3DbType)
	query, _ := NewAssistantQueryService(conv.storage.container.UserDataStore, 7, AssistantQueryOptions{Secret: []byte("test")})
	original, compact := query.Definitions(), (assistantContextTools{AssistantQueryService: query}).Definitions()
	if len(compact) != 7 {
		t.Fatal(len(compact))
	}
	for i := range original {
		if original[i].Name != compact[i].Name || original[i].Description != compact[i].Description {
			t.Fatal("tool semantics changed")
		}
		var a, b any
		_ = json.Unmarshal(original[i].Parameters, &a)
		_ = json.Unmarshal(compact[i].Parameters, &b)
		assistantRemoveSchemaDescriptions(a)
		if !reflect.DeepEqual(a, b) {
			t.Fatal("validation schema changed", original[i].Name)
		}
	}
}

func TestAssistantPrepareThirtyTwoTurnsAutomaticCompaction(t *testing.T) {
	t.Chdir("../..")
	conv := assistantConversationFixture(t, settings.Sqlite3DbType)
	var nextID int64 = 7000000000000000000
	conv.options.GenerateID = func(uuid.UuidType) int64 { nextID++; return nextID }
	cfg := &settings.Config{EnableAIAssistant: true, SecretKey: "test"}
	service := NewAssistantContextService(conv.storage, func() *settings.Config { return cfg })
	p := &contextConversationProvider{}
	service.providers = func() (provider.LargeLanguageModelToolProvider, provider.LargeLanguageModelToolProvider, error) {
		return p, nil, nil
	}
	var last models.AssistantCommandResult
	for i := 0; i < 32; i++ {
		question := fmt.Sprintf("turn %d: account 42; dining budget CNY 2500; date 2026-09. %s", i, strings.Repeat("please preserve this explicit condition. ", 80))
		last = conversationSend(t, conv, fmt.Sprint("auto-", i), last.ConversationId, last.ContextVersion, question)
		var run *models.AssistantMessage
		_ = conv.storage.WithUserTransaction(core.NewNullContext(), 7, func(tx *AssistantStorageTx) error {
			id, _ := assistantParseID(last.RunId)
			run, _ = assistantConversationMessage(tx.Session, 7, id)
			run.Deadline = time.Now().Add(time.Minute).UnixMilli()
			return nil
		})
		prepared, e := service.Prepare(core.NewNullContext(), run)
		if e != nil {
			t.Fatalf("turn %d: %v", i, e)
		}
		result, e := prepared.Runner.Run(core.NewNullContext(), prepared.Input, nil)
		if e != nil || result.State != llm.AssistantStateCompleted {
			t.Fatalf("turn %d: %v", i, e)
		}
		conversationComplete(t, conv, last, result.Reply)
	}
	if p.summaries < 2 || p.answers != 32 {
		t.Fatal("automatic compaction not exercised", p.summaries, p.answers)
	}
}

type contextConversationProvider struct{ summaries, answers int }

func (p *contextConversationProvider) RunToolTurn(_ core.Context, _ int64, input data.ToolTurnInput, _ data.DisplayEventSink) (data.TurnResponse, error) {
	if input.Messages[0].Content == assistantCompactionInstruction {
		p.summaries++
		return data.TurnResponse{Completed: true, Content: `{"topic":"budget","confirmedConstraints":["user: account 42; dining budget CNY 2500; date 2026-09"],"resolvedReferences":[],"openQuestions":[]}`}, nil
	}
	p.answers++
	return data.TurnResponse{Completed: true, Content: "Your conditions are account 42; dining budget CNY 2500; date 2026-09. Fresh tools are required for actual financial facts."}, nil
}

func TestAssistantPrepareAcceptsChineseQueryWithinTokenBudget(t *testing.T) {
	t.Chdir("../..")
	conv := assistantConversationFixture(t, settings.Sqlite3DbType)
	accepted := conversationSend(t, conv, "large-context", "", 0, strings.Repeat("中", 4000))
	var run *models.AssistantMessage
	_ = conv.storage.WithUserTransaction(core.NewNullContext(), 7, func(tx *AssistantStorageTx) error {
		id, _ := assistantParseID(accepted.RunId)
		run, _ = assistantConversationMessage(tx.Session, 7, id)
		run.Deadline = time.Now().Add(time.Minute).UnixMilli()
		return nil
	})
	service := NewAssistantContextService(conv.storage, func() *settings.Config { return &settings.Config{EnableAIAssistant: true, SecretKey: "test"} })
	p := &contextConversationProvider{}
	service.providers = func() (provider.LargeLanguageModelToolProvider, provider.LargeLanguageModelToolProvider, error) {
		return p, nil, nil
	}
	if _, e := service.Prepare(core.NewNullContext(), run); e != nil {
		t.Fatal(e)
	}
	if p.summaries != 0 || p.answers != 0 {
		t.Fatal("charged before budget rejection")
	}
}

func TestAssistantSavedCursorHandleIsRunLocalAndNeverPersisted(t *testing.T) {
	query, _ := assistantServiceFixture(t)
	first := assistantCall(query, data.ToolQueryTransactions, `{"page_size":1}`)
	if !first.OK || first.NextCursor == "" {
		t.Fatal(first)
	}
	tools := assistantContextTools{AssistantQueryService: query, continuation: &data.ToolContinuation{Tool: data.ToolQueryTransactions, Cursor: first.NextCursor}}
	call := data.ToolCall{Name: data.ToolQueryTransactions, Arguments: json.RawMessage(`{"cursor":"@saved_page"}`)}
	result := tools.ExecuteForContext(core.NewNullContext(), call, func(data.ToolResult) bool { return true })
	if !result.OK {
		t.Fatal(result)
	}
	if strings.HasPrefix(result.NextCursor, "@") {
		t.Fatal("handle leaked into persistent evidence")
	}
	fresh := assistantContextTools{AssistantQueryService: query}
	if r := fresh.Execute(core.NewNullContext(), call); r.OK || r.Error.Code != data.ToolErrorInvalidCursor {
		t.Fatal("run-local handle restored without server cursor", r)
	}
	wrong := *query
	wrong.uidString = "8"
	other := assistantContextTools{AssistantQueryService: &wrong, continuation: tools.continuation}
	if r := other.Execute(core.NewNullContext(), call); r.OK || r.Error.Code != data.ToolErrorInvalidCursor {
		t.Fatal("handle bypassed UID validation", r)
	}
	call.Arguments = json.RawMessage(`{"cursor":"@unknown"}`)
	if r := tools.Execute(core.NewNullContext(), call); r.OK || r.Error.Code != data.ToolErrorInvalidCursor {
		t.Fatal("unknown handle accepted", r)
	}
}
