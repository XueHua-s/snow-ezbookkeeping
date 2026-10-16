package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider/openai"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

func contextBudgetRun(t *testing.T, conv *AssistantConversationService, accepted models.AssistantCommandResult) *models.AssistantMessage {
	t.Helper()
	var run *models.AssistantMessage
	err := conv.storage.WithUserTransaction(core.NewNullContext(), 7, func(tx *AssistantStorageTx) error {
		id, _ := assistantParseID(accepted.RunId)
		var err error
		run, err = assistantConversationMessage(tx.Session, 7, id)
		if err == nil {
			run.Deadline = time.Now().Add(time.Minute).UnixMilli()
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return run
}

func contextBudgetService(conv *AssistantConversationService, p provider.LargeLanguageModelToolProvider) *AssistantContextService {
	s := NewAssistantContextService(conv.storage, func() *settings.Config { return &settings.Config{EnableAIAssistant: true, SecretKey: "test"} })
	s.providers = func() (provider.LargeLanguageModelToolProvider, provider.LargeLanguageModelToolProvider, error) {
		return p, nil, nil
	}
	return s
}

func TestAssistantPrepareCompactsOnlyLongHistoricalTurn(t *testing.T) {
	t.Chdir("../..")
	conv := assistantConversationFixture(t, settings.Sqlite3DbType)
	old := conversationSend(t, conv, "old", "", 0, "preserve account 42")
	conversationComplete(t, conv, old, strings.Repeat("中", 11000))
	current := conversationSend(t, conv, "current", old.ConversationId, old.ContextVersion, "current question must remain exact")
	p := &contextSummaryProvider{}
	s := contextBudgetService(conv, p)
	prepared, err := s.Prepare(core.NewNullContext(), contextBudgetRun(t, conv, current))
	if err != nil {
		t.Fatal(err)
	}
	if p.calls != 1 || prepared.Input.PriorModelCalls != 1 {
		t.Fatalf("summary calls not accounted: %d/%d", p.calls, prepared.Input.PriorModelCalls)
	}
	if got := prepared.Input.Messages[len(prepared.Input.Messages)-1].Content; got != "current question must remain exact" {
		t.Fatal(got)
	}
	if strings.Contains(p.sources[0], "current question must remain exact") || !strings.Contains(p.sources[0], "preserve account 42") {
		t.Fatal("compacted wrong turn")
	}
}

func TestAssistantCompactionBatchesWholeTurnsAndPreservesFrozenRemainder(t *testing.T) {
	for _, frozen := range []bool{false, true} {
		t.Run(fmt.Sprint("frozen=", frozen), func(t *testing.T) {
			conv := assistantConversationFixture(t, settings.Sqlite3DbType)
			var last models.AssistantCommandResult
			for i := 0; i < 5; i++ {
				last = conversationSend(t, conv, fmt.Sprint("history-", i), last.ConversationId, last.ContextVersion, fmt.Sprint("condition-", i))
				conversationComplete(t, conv, last, strings.Repeat("中", 6000))
			}
			if frozen {
				branch, err := conv.Branch(core.NewNullContext(), 7, models.AssistantBranchRequest{AssistantCommand: storageCommand(conv.storage, "branch", 0), SourceConversationId: last.ConversationId, SourceMessageId: last.UserMessageId, ExpectedContextVersion: last.ContextVersion, Kind: "continue"})
				if err != nil {
					t.Fatal(err)
				}
				last = branch
			}
			current := conversationSend(t, conv, "current", last.ConversationId, last.ContextVersion, "latest")
			run := contextBudgetRun(t, conv, current)
			p := &contextSummaryProvider{}
			s := contextBudgetService(conv, p)
			snap, err := s.snapshot(core.NewNullContext(), run)
			if err != nil {
				t.Fatal(err)
			}
			changed, calls, err := s.compact(core.NewNullContext(), run, snap, p, 16000)
			if err != nil || !changed || calls != 1 {
				t.Fatal(changed, calls, err)
			}
			if !strings.Contains(p.sources[0], "condition-0") || !strings.Contains(p.sources[0], "condition-1") || strings.Contains(p.sources[0], "condition-2") {
				t.Fatal("summary did not select fitting whole prefix")
			}
			fresh, err := s.snapshot(core.NewNullContext(), run)
			if err != nil {
				t.Fatal(err)
			}
			if frozen {
				if len(fresh.Frozen) != 3 || fresh.Frozen[0].Question.Content != "condition-2" {
					t.Fatal("uncovered frozen turns lost", len(fresh.Frozen))
				}
			} else if fresh.Conversation.SummaryThrough != 2 {
				t.Fatal("wrong summary boundary", fresh.Conversation.SummaryThrough)
			}
			changed, calls, err = s.compact(core.NewNullContext(), run, fresh, p, 16000)
			if err != nil || !changed || calls != 1 {
				t.Fatal(changed, calls, err)
			}
			if !strings.Contains(p.sources[1], "condition-2") || !strings.Contains(p.sources[1], "condition-3") || strings.Contains(p.sources[1], "condition-4") {
				t.Fatal("second batch skipped or split a turn")
			}
			if frozen && len(fresh.Frozen) != 1 {
				t.Fatal("second batch lost frozen remainder")
			}
		})
	}
}

type contextFailureProvider struct {
	calls  int
	err    error
	before func()
}

func (p *contextFailureProvider) RunToolTurn(core.Context, int64, data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
	p.calls++
	if p.before != nil {
		p.before()
	}
	return data.TurnResponse{}, p.err
}

func TestAssistantCompactionFailurePreservesHistoryAndReason(t *testing.T) {
	for _, failure := range []error{errors.New("upstream unavailable"), context.Canceled, context.DeadlineExceeded} {
		t.Run(failure.Error(), func(t *testing.T) {
			conv := assistantConversationFixture(t, settings.Sqlite3DbType)
			old := conversationSend(t, conv, "old", "", 0, "account 42")
			conversationComplete(t, conv, old, "history")
			current := conversationSend(t, conv, "current", old.ConversationId, old.ContextVersion, "latest")
			run := contextBudgetRun(t, conv, current)
			p := &contextFailureProvider{err: failure}
			s := contextBudgetService(conv, p)
			snap, err := s.snapshot(core.NewNullContext(), run)
			if err != nil {
				t.Fatal(err)
			}
			changed, calls, err := s.compact(core.NewNullContext(), run, snap, p, 16000)
			want := failure
			if !errors.Is(failure, context.Canceled) && !errors.Is(failure, context.DeadlineExceeded) {
				want = models.AssistantReasonProviderUnavailable
			}
			if changed || calls != 1 || err != want {
				t.Fatal(changed, calls, err)
			}
			fresh, err := s.snapshot(core.NewNullContext(), run)
			if err != nil {
				t.Fatal(err)
			}
			if fresh.Conversation.Summary != "" || fresh.Conversation.SummaryThrough != 0 || len(fresh.Turns) != 1 {
				t.Fatal("failed summary mutated history")
			}
		})
	}
}

func TestAssistantPrepareCountsFailedCompactionAndHonorsAcceptedByteLimit(t *testing.T) {
	t.Chdir("../..")
	conv := assistantConversationFixture(t, settings.Sqlite3DbType)
	old := conversationSend(t, conv, "old", "", 0, "old question")
	conversationComplete(t, conv, old, strings.Repeat("中", 9000))
	current := conversationSend(t, conv, "current", old.ConversationId, old.ContextVersion, "latest")
	run := contextBudgetRun(t, conv, current)
	p := &contextFailureProvider{err: errors.New("upstream")}
	s := contextBudgetService(conv, p)
	prepared, err := s.Prepare(core.NewNullContext(), run)
	if err != nil {
		t.Fatal(err)
	}
	if p.calls != 1 || prepared.Input.PriorModelCalls != 1 {
		t.Fatal("failed call not charged", p.calls, prepared.Input.PriorModelCalls)
	}
	var limits models.AssistantLimits
	if err = json.Unmarshal([]byte(run.LimitsJSON), &limits); err != nil {
		t.Fatal(err)
	}
	limits.RequestBytes = 100
	raw, _ := json.Marshal(limits)
	run.LimitsJSON = string(raw)
	if _, err = s.Prepare(core.NewNullContext(), run); err != models.AssistantReasonContextLimit {
		t.Fatal("accepted byte cap bypassed", err)
	}
	if p.calls != 1 {
		t.Fatal("byte cap checked after model call")
	}
}

func TestAssistantRealMetadataToolResultsFitTokenBudgetOverHTTP(t *testing.T) {
	t.Chdir("../..")
	query, engine := assistantServiceFixture(t)
	// The isolated fixture has 3 visible-to-tools accounts and 3 categories.
	for i := 0; i < 37; i++ {
		if _, err := engine.Insert(&models.Account{AccountId: int64(1000 + i), Uid: 7, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Category: models.ACCOUNT_CATEGORY_CASH, Name: fmt.Sprintf("日常账户%02d", i), Currency: "CNY"}); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 121; i++ {
		if _, err := engine.Insert(&models.TransactionCategory{CategoryId: int64(2000 + i), Uid: 7, Type: models.CATEGORY_TYPE_EXPENSE, Name: fmt.Sprintf("生活支出类别%03d", i)}); err != nil {
			t.Fatal(err)
		}
	}
	tools := assistantContextTools{AssistantQueryService: query}
	calls := 0
	largestWire := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		calls++
		largestWire = max(largestWire, len(raw))
		if !data.AssistantRequestFits(raw, 16000, 128*1024) {
			t.Error("oversized real wire request")
		}
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			io.WriteString(w, `{"choices":[{"index":0,"finish_reason":"tool_calls","message":{"role":"assistant","tool_calls":[{"id":"accounts","type":"function","function":{"name":"list_accounts","arguments":"{\"page_size\":200}"}},{"id":"categories","type":"function","function":{"name":"list_categories","arguments":"{\"page_size\":200}"}}]}}]}`)
			return
		}
		var body struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Error(err)
		}
		counts := []int{}
		for _, m := range body.Messages {
			if m.Role == "tool" {
				var result data.ToolResult
				if err := json.Unmarshal([]byte(m.Content), &result); err != nil {
					t.Error(err)
				}
				if !result.OK || result.ReturnedCount == 0 || (result.HasMore && (result.NextCursor == "" || result.Coverage != data.QueryCoveragePage)) {
					t.Error("metadata page lost or pagination hidden", result)
				}
				wantMatched := "40"
				if len(counts) == 1 {
					wantMatched = "124"
				}
				if result.MatchedCount != wantMatched {
					t.Error("complete metadata count lost", result.MatchedCount)
				}
				counts = append(counts, result.ReturnedCount)
			}
		}
		if len(counts) != 2 || counts[0] <= 0 || counts[0] > 40 || counts[1] <= 0 || counts[1] > 124 {
			t.Error("invalid bounded metadata pages", counts)
		}
		io.WriteString(w, `{"choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"Account and category pages are available; more categories can be requested."}}]}`)
	}))
	defer server.Close()
	p := openai.NewOpenAICompatibleLargeLanguageModelProvider(&settings.LLMConfig{OpenAICompatibleBaseURL: server.URL, OpenAICompatibleModelID: "fixture", LargeLanguageModelAPIRequestTimeout: 2000, LargeLanguageModelAPIProxy: "none"}, false).(provider.LargeLanguageModelToolProvider)
	messages, err := llm.BuildAssistantMessages("chat", "列出所有账户和分类", nil, time.Now(), "zh_CN", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	messages[0].Content += "\n" + assistantCompactToolRules
	result, err := llm.NewAssistantRunner(p, nil, tools).Run(core.NewNullContext(), llm.AssistantRunInput{UID: 7, Mode: "chat", Messages: messages, MaxInputTokens: 16000, MaxOutputTokens: 4096, MaxRounds: 4}, nil)
	if err != nil || result.State != llm.AssistantStateCompleted || calls != 2 {
		t.Fatal(result, err, calls)
	}
	if largestWire <= 16000 {
		t.Fatal("fixture did not reproduce old byte-budget failure", largestWire)
	}
}

func TestAssistantPrepareReservesAnswerRoundAndStopsCanceledCompaction(t *testing.T) {
	t.Chdir("../..")
	conv := assistantConversationFixture(t, settings.Sqlite3DbType)
	old := conversationSend(t, conv, "old", "", 0, "old")
	conversationComplete(t, conv, old, strings.Repeat("中", 9000))
	current := conversationSend(t, conv, "current", old.ConversationId, old.ContextVersion, "latest")
	run := contextBudgetRun(t, conv, current)
	p := &contextSummaryProvider{}
	s := contextBudgetService(conv, p)
	policy, err := s.acceptedPolicy(run)
	if err != nil {
		t.Fatal(err)
	}
	policy.MaxRounds = 1
	raw, _ := json.Marshal(policy)
	run.RuntimeJSON = string(raw)
	prepared, err := s.Prepare(core.NewNullContext(), run)
	if err != nil || p.calls != 0 || prepared.Input.PriorModelCalls != 0 {
		t.Fatal("spent final answer round on summary", err, p.calls)
	}
	policy.MaxRounds = 2
	raw, _ = json.Marshal(policy)
	run.RuntimeJSON = string(raw)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	snap, err := s.snapshot(core.NewNullContext(), run)
	if err != nil {
		t.Fatal(err)
	}
	changed, calls, err := s.compact(core.NewAssistantContext(ctx, "canceled", "", ""), run, snap, p, 16000)
	if changed || calls != 0 || err == nil || p.calls != 0 {
		t.Fatal("canceled compaction called model", changed, calls, err, p.calls)
	}
}
