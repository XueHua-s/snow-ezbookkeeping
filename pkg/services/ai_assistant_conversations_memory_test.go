package services

import (
	"encoding/json"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

func TestAssistantConversationMemoryOffRevokesOtherRunsWithoutRevokingBranch(t *testing.T) {
	s := assistantConversationFixture(t, settings.Sqlite3DbType)
	ctx := core.NewNullContext()
	source := conversationSend(t, s, "source", "", 0, "回答简洁")
	conversationComplete(t, s, source, "已了解偏好")
	branch, err := s.Branch(ctx, 7, models.AssistantBranchRequest{AssistantCommand: storageCommand(s.storage, "branch", 0), SourceConversationId: source.ConversationId, SourceMessageId: source.UserMessageId, ExpectedContextVersion: source.ContextVersion, Kind: "continue"})
	if err != nil {
		t.Fatal(err)
	}
	own := conversationSend(t, s, "own-active", source.ConversationId, source.ContextVersion, "使用记忆")
	other := conversationSend(t, s, "other", "", 0, "另一会话使用记忆")
	err = s.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		cid, _ := assistantParseID(source.ConversationId)
		c, err := assistantConversationLoad(tx.Session, 7, cid)
		if err != nil {
			return err
		}
		sources, _ := json.Marshal([]models.AssistantSourceDependency{{ConversationId: source.ConversationId, MessageId: source.UserMessageId, Version: c.SourceVersion}})
		tx.State.MemoryEnabled = true
		tx.State.ActiveMemoryCount++
		if _, err = tx.Session.Insert(&models.AssistantMemory{Id: 999, Uid: 7, Key: "style", Text: "简洁", Status: "active", SourcesJSON: string(sources), Version: 1, SchemaVersion: 1}); err != nil {
			return err
		}
		dep, _ := json.Marshal(models.AssistantInputDependencies{MemoryMode: "inherit", Memories: []models.AssistantMemoryDependency{{Id: "999", Version: 1}}})
		for _, result := range []models.AssistantCommandResult{own, other} {
			id, _ := assistantParseID(result.RunId)
			m, err := assistantConversationMessage(tx.Session, 7, id)
			if err != nil {
				return err
			}
			m.InputDependenciesJSON = string(dep)
			if _, err = tx.Session.Where("uid = ? AND id = ?", 7, id).AllCols().Update(m); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	off := "off"
	_, err = s.Update(ctx, 7, models.AssistantUpdateConversationRequest{AssistantCommand: storageCommand(s.storage, "off", 0), Id: source.ConversationId, ExpectedMetadataVersion: source.MetadataVersion, ExpectedContextVersion: &own.ContextVersion, MemoryMode: &off})
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range []models.AssistantCommandResult{own, other} {
		snapshot, err := s.GetMessage(ctx, 7, result.RunId)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.State != models.AssistantRunCanceled {
			t.Fatal("used revoked memory remains active", snapshot.State)
		}
	}
	state, err := s.storage.GetUserState(ctx, 7)
	if err != nil {
		t.Fatal(err)
	}
	if state.QueuedCount != 0 || state.ActiveMemoryCount != 0 {
		t.Fatal("revocation settlement/count wrong", state)
	}
	detail, err := s.Get(ctx, 7, models.AssistantGetConversationRequest{Id: branch.ConversationId})
	if err != nil {
		t.Fatal(err)
	}
	if detail.Conversation.SourceInvalidated {
		t.Fatal("memory mode revoked valid branch history")
	}
	ownState, err := s.State(ctx, 7, source.ConversationId)
	if err != nil {
		t.Fatal(err)
	}
	if ownState.ActiveRunId != "" {
		t.Fatal("stale metadata update resurrected settled active run")
	}
}
