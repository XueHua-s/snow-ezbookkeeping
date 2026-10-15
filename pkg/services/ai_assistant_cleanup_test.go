package services

import (
	"errors"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"xorm.io/xorm"
)

func TestAssistantLifecycleDatabaseContracts(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			s := assistantConversationFixture(t, kind)
			ctx := core.NewNullContext()
			first := conversationSend(t, s, "before-clear", "", 0, "private question")
			conversationComplete(t, s, first, "private answer")
			if err := s.storage.InvalidateLedgerContext(ctx, 7); err != nil {
				t.Fatal(err)
			}
			err := s.storage.WithUserRead(ctx, 7, func(session *xorm.Session) error {
				id, _ := assistantParseID(first.ConversationId)
				c, err := assistantConversationLoad(session, 7, id)
				if err != nil {
					return err
				}
				if c.ContextInvalidatedThroughTurn != 1 || c.ContextVersion <= first.ContextVersion {
					t.Fatal("ledger fence absent", c)
				}
				n, err := session.Where("uid = ? AND conversation_id = ?", 7, id).Count(&models.AssistantMessage{})
				if n != 2 {
					t.Fatal("ledger clear destroyed readable history", n)
				}
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
			command := storageCommand(s.storage, "clear", 0)
			if _, err = s.Clear(ctx, 7, models.AssistantClearRequest{AssistantCommand: command}); err != nil {
				t.Fatal(err)
			}
			cleaner := NewAssistantCleanupService(s.storage)
			if err = cleaner.Sweep(ctx); err != nil {
				t.Fatal(err)
			}
			err = s.storage.WithUserRead(ctx, 7, func(session *xorm.Session) error {
				for _, model := range []any{&models.AssistantMessage{}, &models.AssistantConversation{}} {
					n, err := session.Where("uid = ?", 7).Count(model)
					if err != nil {
						return err
					}
					if n != 0 {
						t.Fatal("tombstone body survived", n)
					}
				}
				n, err := session.Where("uid = ?", 7).Count(&models.AssistantCommandReceipt{})
				if n == 0 {
					t.Fatal("clear lost retry receipts")
				}
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
			_, err = s.Send(ctx, 7, models.AssistantSendRequest{AssistantCommand: storageCommand(s.storage, "late", 0), Mode: "chat", Content: "late", Locale: "en", Timezone: "UTC"})
			if !errors.Is(err, models.AssistantReasonDataEpochConflict) {
				t.Fatal("late epoch accepted", err)
			}
			fresh, err := s.Send(ctx, 7, models.AssistantSendRequest{AssistantCommand: storageCommand(s.storage, "fresh", 1), Mode: "chat", Content: "new data", Locale: "en", Timezone: "UTC"})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = s.Clear(ctx, 7, models.AssistantClearRequest{AssistantCommand: command}); err != nil {
				t.Fatal(err)
			}
			if err = s.storage.WithUserRead(ctx, 7, func(session *xorm.Session) error {
				id, _ := assistantParseID(fresh.ConversationId)
				_, err := assistantConversationLoad(session, 7, id)
				return err
			}); err != nil {
				t.Fatal("replayed clear deleted new history", err)
			}
			oldNow := s.storage.now()
			if err = s.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
				tx.State.CandidateMemoryCount = 1
				tx.State.UsedBytes += 5
				_, err := tx.Session.Insert(&models.AssistantMemory{Id: 9999, Uid: 7, Status: "candidate", Text: "stale", UsedBytes: 5, CreatedAt: oldNow.Add(-8 * 24 * time.Hour).UnixMilli()})
				return err
			}); err != nil {
				t.Fatal(err)
			}
			if err = cleaner.CleanupUser(ctx, 7); err != nil {
				t.Fatal(err)
			}
			state, err := s.storage.GetUserState(ctx, 7)
			if err != nil || state.CandidateMemoryCount != 0 {
				t.Fatal("candidate retention", state, err)
			}
		})
	}
}

func TestAssistantAccountFencePreservesDisabledHistory(t *testing.T) {
	s := assistantConversationFixture(t, settings.Sqlite3DbType)
	ctx := core.NewNullContext()
	first := conversationSend(t, s, "account-question", "", 0, "keep on disable")
	conversationComplete(t, s, first, "keep answer")
	users := &UserService{assistantStorage: s.storage}
	if err := users.fenceAssistantAccount(ctx, 7, true, false); err != nil {
		t.Fatal(err)
	}
	state, err := s.storage.GetUserState(ctx, 7)
	if err != nil || !state.AccountUnavailable || state.ConversationCount != 1 {
		t.Fatal(state, err)
	}
	_, err = s.Send(ctx, 7, models.AssistantSendRequest{AssistantCommand: storageCommand(s.storage, "denied", state.DeletionEpoch), Mode: "chat", Content: "denied", Locale: "en", Timezone: "UTC"})
	if !errors.Is(err, models.AssistantReasonNotAccessible) {
		t.Fatal("disabled account accepted command", err)
	}
	if err = users.fenceAssistantAccount(ctx, 7, false, false); err != nil {
		t.Fatal(err)
	}
	if err = s.storage.WithUserRead(ctx, 7, func(session *xorm.Session) error {
		id, _ := assistantParseID(first.ConversationId)
		_, err := assistantConversationLoad(session, 7, id)
		return err
	}); err != nil {
		t.Fatal("disable destroyed history", err)
	}
	if err = users.fenceAssistantAccount(ctx, 7, true, true); err != nil {
		t.Fatal(err)
	}
	state, err = s.storage.GetUserState(ctx, 7)
	if err != nil || !state.AccountUnavailable || state.ConversationCount != 0 || state.DeletionEpoch != 2 {
		t.Fatal("delete did not permanently fence", state, err)
	}
}

func TestAssistantCleanupBoundsAndPreservesOtherUsers(t *testing.T) {
	storage, engine := assistantStorageFixture(t, settings.Sqlite3DbType)
	ctx := core.NewNullContext()
	err := storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		if _, err := tx.Session.Insert(&models.AssistantConversation{Id: 5000, Uid: 7, Deleted: true, Title: "deleted title"}); err != nil {
			return err
		}
		for i := int64(1); i <= 1001; i++ {
			if _, err := tx.Session.Insert(&models.AssistantMessage{Id: i, Uid: 7, ConversationId: 5000, UserMessageId: i, Content: "body to erase"}); err != nil {
				return err
			}
		}
		_, err := tx.Session.Insert(&models.AssistantMessage{Id: 9000, Uid: 8, ConversationId: 6000, UserMessageId: 9000, Content: "other owner"})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	legacy := &models.AIAssistantEmbedding{Uid: 7, TransactionId: 600, EmbeddingModel: "rollback-only", ContentHash: "legacy-hash", VectorData: "[1,2,3]"}
	if _, err = engine.Insert(legacy); err != nil {
		t.Fatal(err)
	}
	cleaner := NewAssistantCleanupService(storage)
	if err = cleaner.CleanupUser(ctx, 7); err != nil {
		t.Fatal(err)
	}
	remaining, err := engine.Where("uid = ?", 7).Count(&models.AssistantMessage{})
	if err != nil || remaining != 1 {
		t.Fatal("cleanup exceeded per-user batch", remaining, err)
	}
	tombstones, err := engine.Where("uid = ?", 7).Count(&models.AssistantConversation{})
	if err != nil || tombstones != 1 {
		t.Fatal("removed parent before final body", tombstones, err)
	}
	if err = cleaner.CleanupUser(ctx, 7); err != nil {
		t.Fatal(err)
	}
	remaining, err = engine.Where("uid = ?", 7).Count(&models.AssistantMessage{})
	if err != nil || remaining != 0 {
		t.Fatal(remaining, err)
	}
	other, err := engine.Where("uid = ?", 8).Count(&models.AssistantMessage{})
	if err != nil || other != 1 {
		t.Fatal("cross-owner cleanup", other, err)
	}
	var kept models.AIAssistantEmbedding
	found, err := engine.Where("uid = ? AND transaction_id = ?", 7, 600).Get(&kept)
	if err != nil || !found || kept.VectorData != legacy.VectorData || kept.ContentHash != legacy.ContentHash {
		t.Fatal("legacy embedding changed", kept, err)
	}
}
