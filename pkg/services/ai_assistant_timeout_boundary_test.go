package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

func TestAssistantWorkerDistinguishesLocalAndRunDeadlines(t *testing.T) {
	for _, tc := range []struct {
		name    string
		failure error
		reason  models.AssistantReasonCode
	}{{"provider", context.DeadlineExceeded, models.AssistantReasonProviderUnavailable}, {"query", &data.ToolError{Code: data.ToolErrorQueryTimeout}, models.AssistantReasonExecutionLimit}, {"query-continuation", errors.Join(models.AssistantReasonExecutionLimit, &data.ToolError{Code: data.ToolErrorQueryTimeout}), models.AssistantReasonExecutionLimit}, {"ordinary-limit", models.AssistantReasonExecutionLimit, models.AssistantReasonExecutionLimit}, {"run", nil, models.AssistantReasonRunTimeout}} {
		t.Run(tc.name, func(t *testing.T) {
			worker, _ := assistantRunFixture(t, settings.Sqlite3DbType)
			worker.options.Prepare = func(ctx core.Context, _ *models.AssistantMessage) (*AssistantPreparedRun, error) {
				if tc.name == "run" {
					<-ctx.Done()
					return nil, ctx.Err()
				}
				return nil, tc.failure
			}
			run, err := worker.claim(core.NewNullContext(), 7, 3)
			if err != nil {
				t.Fatal(err)
			}
			if tc.name == "run" {
				run.Deadline = time.Now().Add(15 * time.Millisecond).UnixMilli()
			}
			worker.execute(context.Background(), run)
			snapshot, err := worker.GetRunSnapshot(core.NewNullContext(), 7, "3")
			if err != nil || snapshot.ReasonCode == nil || *snapshot.ReasonCode != tc.reason {
				t.Fatal("wrong deadline layer", snapshot.ReasonCode, err)
			}
			if strings.HasPrefix(tc.name, "query") && snapshot.State != models.AssistantRunFailed {
				t.Fatal("preparation query timeout should retain failed state", snapshot.State)
			}
			knownWire := map[string]bool{"execution_limit": true, "provider_unavailable": true, "run_timeout": true}
			if !knownWire[string(*snapshot.ReasonCode)] {
				t.Fatal("new reason breaks cached clients", snapshot.ReasonCode)
			}
			if tc.name == "ordinary-limit" && snapshot.State != models.AssistantRunFailed {
				t.Fatal("unrelated execution limit state changed", snapshot.State)
			}
			if tc.name != "run" && *snapshot.ReasonCode == models.AssistantReasonRunTimeout {
				t.Fatal("local timeout blamed on run")
			}
		})
	}
}

func TestAssistantContinuationDeadlineLayersAndExactScope(t *testing.T) {
	conv := assistantConversationFixture(t, settings.Sqlite3DbType)
	old := conversationSend(t, conv, "old", "", 0, "query")
	conversationComplete(t, conv, old, "answer")
	current := conversationSend(t, conv, "next", old.ConversationId, old.ContextVersion, "next page")
	run := contextBudgetRun(t, conv, current)
	run.SourceMessageId, _ = assistantParseID(old.RunId)
	query, _ := assistantServiceFixture(t)
	first := assistantCall(query, data.ToolQueryTransactions, `{"filter":{"account_ids":["11"],"currency":"CNY","types":["expense"]},"page_size":1}`)
	if !first.OK || !first.HasMore {
		t.Fatal(first)
	}
	candidate := data.ToolContinuation{Tool: data.ToolQueryTransactions, Cursor: first.NextCursor}
	raw, _ := json.Marshal(candidate)
	cursor, err := query.decodeCursor(first.NextCursor, candidate.Tool)
	if err != nil {
		t.Fatal(err)
	}
	filter, _ := json.Marshal(cursor.Args)
	err = conv.storage.WithUserTransaction(core.NewNullContext(), 7, func(tx *AssistantStorageTx) error {
		source, err := assistantConversationMessage(tx.Session, 7, run.SourceMessageId)
		if err != nil {
			return err
		}
		source.Continuation, source.ContinuationFilterJSON = string(raw), string(filter)
		_, err = tx.Session.Where("uid = ? AND id = ?", 7, run.SourceMessageId).AllCols().Update(source)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	s := contextBudgetService(conv, &contextSummaryProvider{})
	snap, err := s.snapshot(core.NewNullContext(), run)
	if err != nil {
		t.Fatal(err)
	}
	restored, scope, restart, err := s.prepareContinuation(core.NewNullContext(), run, snap, query)
	if err != nil || restored == nil || *restored != candidate || !reflect.DeepEqual(scope, first.Scope) || restart != "" {
		t.Fatal("continuation changed", restored, scope, err)
	}
	for _, reset := range []bool{false, true} {
		t.Run(fmt.Sprint("restart=", reset), func(t *testing.T) {
			run.RestartQuery = reset
			canceled, cancel := context.WithCancel(context.Background())
			cancel()
			_, _, _, err := s.prepareContinuation(core.NewAssistantContext(canceled, "cancel", "", ""), run, snap, query)
			if !errors.Is(err, context.Canceled) {
				t.Fatal("parent cancellation became cursor error", err)
			}
			query.timeout = time.Nanosecond
			_, _, _, err = s.prepareContinuation(core.NewNullContext(), run, snap, query)
			if !errors.Is(err, models.AssistantReasonExecutionLimit) {
				t.Fatal("local query timeout became cursor error", err)
			}
		})
	}
}

func TestAssistantFailedSummaryCannotBypassChangedAuthorization(t *testing.T) {
	for _, change := range []string{"version", "cancel", "deletion"} {
		t.Run(change, func(t *testing.T) {
			t.Chdir("../..")
			conv := assistantConversationFixture(t, settings.Sqlite3DbType)
			old := conversationSend(t, conv, "old", "", 0, "original user constraints")
			conversationComplete(t, conv, old, strings.Repeat("中", 9000))
			current := conversationSend(t, conv, "current", old.ConversationId, old.ContextVersion, "continue")
			run := contextBudgetRun(t, conv, current)
			p := &contextFailureProvider{err: context.DeadlineExceeded, before: func() {
				err := conv.storage.WithUserTransaction(core.NewNullContext(), 7, func(tx *AssistantStorageTx) error {
					if change == "cancel" {
						_, err := tx.Session.Where("uid = ? AND id = ?", 7, run.Id).Cols("cancel_requested").Update(&models.AssistantMessage{CancelRequested: true})
						return err
					}
					if change == "deletion" {
						tx.State.DeletionEpoch++
						return nil
					}
					c, err := assistantConversationLoad(tx.Session, 7, run.ConversationId)
					if err != nil {
						return err
					}
					c.ContextVersion++
					return assistantConversationSave(tx, c)
				})
				if err != nil {
					t.Fatal(err)
				}
			}}
			_, err := contextBudgetService(conv, p).Prepare(core.NewNullContext(), run)
			if err != models.AssistantReasonContextChanged {
				t.Fatal("stale original history recovered", change, err)
			}
		})
	}
}

func TestAssistantPrepareHardBudgetRecoveryHonorsRoundLimit(t *testing.T) {
	for _, rounds := range []uint32{2, 12} {
		t.Run(fmt.Sprint(rounds), func(t *testing.T) {
			t.Chdir("../..")
			conv := assistantConversationFixture(t, settings.Sqlite3DbType)
			var last models.AssistantCommandResult
			for i := 0; i < 3; i++ {
				last = conversationSend(t, conv, fmt.Sprint(i), last.ConversationId, last.ContextVersion, fmt.Sprint("constraint", i))
				conversationComplete(t, conv, last, strings.Repeat("中", 5000))
			}
			current := conversationSend(t, conv, "current", last.ConversationId, last.ContextVersion, "latest")
			run := contextBudgetRun(t, conv, current)
			p := &contextFailureProvider{err: context.DeadlineExceeded}
			s := contextBudgetService(conv, p)
			policy, err := s.acceptedPolicy(run)
			if err != nil {
				t.Fatal(err)
			}
			policy.MaxRounds = rounds
			raw, _ := json.Marshal(policy)
			run.RuntimeJSON = string(raw)
			_, err = s.Prepare(core.NewNullContext(), run)
			want := 1
			if rounds == 12 {
				want = 2
			}
			if err != models.AssistantReasonProviderUnavailable || p.calls != want {
				t.Fatal("recovery exceeded original budget", err, p.calls)
			}
			snap, e := s.snapshot(core.NewNullContext(), run)
			if e != nil || snap.Conversation.Summary != "" || len(snap.Turns) != 3 {
				t.Fatal("failed recovery lost history", e)
			}
		})
	}
}
