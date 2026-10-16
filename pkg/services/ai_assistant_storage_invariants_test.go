package services

import (
	"errors"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

func TestAssistantStorageRevocationAndRetention(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			s, _ := assistantStorageFixture(t, kind)
			ctx := core.NewNullContext()
			cmd := storageCommand(s, "old-live", 0)
			if _, err := s.ExecuteCommand(ctx, 7, "branch", cmd, cmd, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
				_, err := tx.Session.Insert(&models.AssistantConversation{Id: 1, Uid: 7, SourceVersion: 1})
				return models.AssistantCommandResult{ConversationId: "1"}, err
			}); err != nil {
				t.Fatal(err)
			}
			// Deletion occurs long after initial receipt's minimum age.
			deletedAt := s.now().Add(10 * 24 * time.Hour)
			s.now = func() time.Time { return deletedAt }
			if err := s.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
				if err := tx.TombstoneReceipts(1, 0, false); err != nil {
					return err
				}
				_, err := tx.Session.Where("uid = ? AND id = ?", 7, 1).Cols("deleted", "deleted_at", "source_version").Update(&models.AssistantConversation{Deleted: true, DeletedAt: tx.Now.UnixMilli(), SourceVersion: 2})
				return err
			}); err != nil {
				t.Fatal(err)
			}
			if err := s.PruneReceipts(ctx, 7, 100); err != nil {
				t.Fatal(err)
			}
			receipt, err := s.GetCommand(ctx, 7, "old-live")
			if err != nil || receipt.Status != "deleted" {
				t.Fatal("receipt lost immediately after deletion", receipt, err)
			}
			s.now = func() time.Time { return deletedAt.Add(47 * time.Hour) }
			if err := s.PruneReceipts(ctx, 7, 100); err != nil {
				t.Fatal(err)
			}
			receipt, _ = s.GetCommand(ctx, 7, "old-live")
			if receipt.Status != "deleted" {
				t.Fatal("receipt lost before minimum retention")
			}
			s.now = func() time.Time { return deletedAt.Add(49 * time.Hour) }
			if err := s.PruneReceipts(ctx, 7, 100); err != nil {
				t.Fatal(err)
			}
			receipt, _ = s.GetCommand(ctx, 7, "old-live")
			if receipt.Status != "unknown" {
				t.Fatal("expired tombstone retained")
			}

			if err := s.WithUserTransaction(ctx, 8, func(tx *AssistantStorageTx) error {
				tx.State.MemoryEnabled = true
				tx.State.MemoryUseEpoch = 9
				for _, bean := range []any{&models.AssistantConversation{Id: 8, Uid: 8, SourceVersion: 1, ContextVersion: 3}, &models.AssistantConversation{Id: 9, Uid: 8, SourceVersion: 1}, &models.AssistantMemory{Id: 8, Uid: 8, Status: "active", Version: 1}, &models.AssistantMessage{Id: 8, Uid: 8, ConversationId: 8, UserMessageId: 8}} {
					if _, err := tx.Session.Insert(bean); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			validate := func(expected bool, mutate func(*AssistantStorageTx) error) {
				t.Helper()
				err := s.WithUserTransaction(ctx, 8, func(tx *AssistantStorageTx) error {
					if mutate != nil {
						if err := mutate(tx); err != nil {
							return err
						}
					}
					conv := &models.AssistantConversation{Id: 9, Uid: 8}
					deps := &models.AssistantInputDependencies{Sources: []models.AssistantSourceDependency{{ConversationId: "8", MessageId: "8", Version: 1}}, Memories: []models.AssistantMemoryDependency{{Id: "8", Version: 1}}}
					return tx.ValidateDependencies(conv, deps)
				})
				if expected && err != nil {
					t.Fatal("valid dependencies rejected", err)
				}
				if !expected && !errors.Is(err, models.AssistantReasonContextChanged) {
					t.Fatal("revoked dependencies accepted", err)
				}
			}
			validate(true, nil)
			validate(true, func(tx *AssistantStorageTx) error {
				_, err := tx.Session.ID(8).Cols("context_version").Update(&models.AssistantConversation{ContextVersion: 99})
				return err
			})
			validate(false, func(tx *AssistantStorageTx) error {
				_, err := tx.Session.ID(8).Cols("source_version").Update(&models.AssistantConversation{SourceVersion: 2})
				return err
			})
			validate(false, func(tx *AssistantStorageTx) error {
				_, err := tx.Session.ID(8).Cols("version").Update(&models.AssistantMemory{Version: 2})
				return err
			})
			validate(false, func(tx *AssistantStorageTx) error { tx.State.MemoryEnabled = false; return nil })
			validate(false, func(tx *AssistantStorageTx) error { tx.State.DeletionEpoch++; return nil })
			validate(false, func(tx *AssistantStorageTx) error { tx.State.AccountUnavailable = true; return nil })
			if err := s.WithUserTransaction(ctx, 8, func(tx *AssistantStorageTx) error { tx.State.AccountUnavailable = true; return nil }); err != nil {
				t.Fatal(err)
			}
			blocked := storageCommand(s, "disabled-account", 0)
			if _, err := s.ExecuteCommand(ctx, 8, "send", blocked, blocked, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
				t.Fatal("disabled account mutation executed")
				return models.AssistantCommandResult{}, nil
			}); !errors.Is(err, models.AssistantReasonNotAccessible) {
				t.Fatal("disabled account accepted", err)
			}

		})
	}
}
