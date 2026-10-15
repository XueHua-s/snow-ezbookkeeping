package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

func TestAssistantConversationBranchQueryRestartAuthorization(t *testing.T) {
	s := assistantConversationFixture(t, settings.Sqlite3DbType)
	ctx := core.NewNullContext()
	first := conversationSend(t, s, "first", "", 0, "账单查询")
	conversationComplete(t, s, first, "历史结果")
	err := s.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		id, _ := assistantParseID(first.RunId)
		run, err := assistantConversationMessage(tx.Session, 7, id)
		if err != nil {
			return err
		}
		run.Continuation = "expired cursor"
		run.ContinuationExpiresAt = tx.Now.UnixMilli() - 1
		run.ContinuationFilterJSON = `{"startTime":"2026-01-01"}`
		_, err = tx.Session.Where("uid = ? AND id = ?", 7, id).AllCols().Update(run)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	second := conversationSend(t, s, "second", first.ConversationId, first.ContextVersion, "后续不应传入")
	conversationComplete(t, s, second, "未来回答")
	branch, err := s.Branch(ctx, 7, models.AssistantBranchRequest{AssistantCommand: storageCommand(s.storage, "branch", 0), SourceConversationId: first.ConversationId, SourceMessageId: first.RunId, ExpectedContextVersion: second.ContextVersion, Kind: "continue"})
	if err != nil {
		t.Fatal(err)
	}
	request := models.AssistantSendRequest{AssistantCommand: storageCommand(s.storage, "next", 0), ConversationId: branch.ConversationId, ExpectedContextVersion: branch.ContextVersion, Mode: "next_page", Content: "下一页", SourceMessageId: first.RunId}
	if _, err = s.Send(ctx, 7, request); !errors.Is(err, models.AssistantReasonCursorExpired) {
		t.Fatal("expired cursor accepted", err)
	}
	request.RestartQuery = true
	request.ClientRequestId = "restart"
	result, err := s.Send(ctx, 7, request)
	if err != nil {
		t.Fatal(err)
	}
	err = s.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		id, _ := assistantParseID(result.RunId)
		run, err := assistantConversationMessage(tx.Session, 7, id)
		if err != nil {
			return err
		}
		if !run.RestartQuery || run.ContinuationFilterJSON == "" || assistantConversationID(run.SourceMessageId) != first.RunId {
			t.Fatal("restart did not persist authorized source/filter")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	otherBranch, err := s.Branch(ctx, 7, models.AssistantBranchRequest{AssistantCommand: storageCommand(s.storage, "other-branch", 0), SourceConversationId: first.ConversationId, SourceMessageId: first.RunId, ExpectedContextVersion: second.ContextVersion, Kind: "continue"})
	if err != nil {
		t.Fatal(err)
	}
	request.ConversationId = otherBranch.ConversationId
	request.ExpectedContextVersion = otherBranch.ContextVersion
	request.SourceMessageId = second.RunId
	request.ClientRequestId = "forbidden-future"
	if _, err = s.Send(ctx, 7, request); !errors.Is(err, models.AssistantReasonConversationConflict) {
		t.Fatal("branch accepted future result", err)
	}
}

func TestAssistantConversationExportViewedVersionDoesNotChangeContext(t *testing.T) {
	s := assistantConversationFixture(t, settings.Sqlite3DbType)
	ctx := core.NewNullContext()
	first := conversationSend(t, s, "first", "", 0, "question")
	conversationComplete(t, s, first, "OLDER_SELECTED_FOR_EXPORT")
	second, err := s.Retry(ctx, 7, models.AssistantRetryRequest{AssistantCommand: storageCommand(s.storage, "regen", 0), ConversationId: first.ConversationId, UserMessageId: first.UserMessageId, ExpectedContextVersion: first.ContextVersion, Intent: "regenerate"})
	if err != nil {
		t.Fatal(err)
	}
	conversationComplete(t, s, second, "CURRENT_CONTEXT_ANSWER")
	var out bytes.Buffer
	request := models.AssistantExportRequest{Id: first.ConversationId, Format: "markdown", ViewedVersionIds: []string{first.RunId}}
	if err = s.Export(ctx, 7, request, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "OLDER_SELECTED_FOR_EXPORT") || strings.Contains(out.String(), "CURRENT_CONTEXT_ANSWER") {
		t.Fatal("export ignored viewed version")
	}
	state, err := s.State(ctx, 7, first.ConversationId)
	if err != nil {
		t.Fatal(err)
	}
	if state.SelectedAssistantMessageId != second.RunId {
		t.Fatal("export changed context")
	}
	request.ViewedVersionIds = []string{first.RunId, second.RunId}
	if err = s.Export(ctx, 7, request, &out); !errors.Is(err, models.AssistantReasonConversationConflict) {
		t.Fatal("duplicate turn mapping accepted")
	}
	request.ViewedVersionIds = []string{first.RunId}
	request.Id = ""
	request.All = true
	if err = s.Export(ctx, 7, request, &out); !errors.Is(err, models.AssistantReasonConversationConflict) {
		t.Fatal("all export accepted display mapping")
	}
	request.All = false
	request.Id = first.ConversationId
	if err = s.Export(ctx, 8, request, &out); !errors.Is(err, models.AssistantReasonNotAccessible) {
		t.Fatal("foreign display version exported")
	}
}

func TestAssistantConversationBranchCannotReintroduceClearedLedgerPrefix(t *testing.T) {
	s := assistantConversationFixture(t, settings.Sqlite3DbType)
	ctx := core.NewNullContext()
	first := conversationSend(t, s, "first", "", 0, "old ledger question")
	conversationComplete(t, s, first, "old ledger answer")
	second := conversationSend(t, s, "second", first.ConversationId, first.ContextVersion, "old followup")
	conversationComplete(t, s, second, "old followup answer")
	if err := s.storage.InvalidateLedgerContext(ctx, 7); err != nil {
		t.Fatal(err)
	}
	state, err := s.State(ctx, 7, first.ConversationId)
	if err != nil {
		t.Fatal(err)
	}
	branch, err := s.Branch(ctx, 7, models.AssistantBranchRequest{AssistantCommand: storageCommand(s.storage, "branch-after-clear", 0), SourceConversationId: first.ConversationId, SourceMessageId: second.UserMessageId, ExpectedContextVersion: state.ContextVersion, Kind: "edit"})
	if err != nil {
		t.Fatal(err)
	}
	err = s.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		id, _ := assistantParseID(branch.ConversationId)
		c, err := assistantConversationLoad(tx.Session, 7, id)
		if err != nil {
			return err
		}
		var prefix []AssistantFrozenTurn
		if err = json.Unmarshal([]byte(c.FrozenPrefixJSON), &prefix); err != nil {
			return err
		}
		if len(prefix) != 0 || c.SeedContext != "" {
			t.Fatal("cleared ledger context was frozen into branch")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
