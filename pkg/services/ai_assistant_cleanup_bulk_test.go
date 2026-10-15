package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"testing"
	"time"
)

func TestAssistantClearThousandConversations(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			storage, engine := assistantStorageFixture(t, kind)
			ctx := core.NewNullContext()
			if err := storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
				tx.State.ConversationCount = 1000
				tx.State.UsedBytes = 1000
				tx.State.ReservedBytes = 400
				tx.State.RunningCount = 2
				tx.State.QueuedCount = 2
				for id := int64(1); id <= 1000; id++ {
					c := &models.AssistantConversation{Id: id, Uid: 7, Title: "private", Summary: "private summary", UsedBytes: 1, MetadataVersion: 1, ContextVersion: 1, SourceVersion: 1}
					if id <= 4 {
						c.ActiveRunId = id
						c.ReservedBytes = 100
						run := &models.AssistantMessage{Id: id, Uid: 7, ConversationId: id, UserMessageId: 10000 + id, Role: "assistant", RunState: models.AssistantRunQueued, QuotaState: "queued", ReservedBytes: 100}
						if id > 2 {
							run.RunState = models.AssistantRunRunning
							run.QuotaState = "running"
						}
						if _, err := tx.Session.Insert(run); err != nil {
							return err
						}
					}
					if _, err := tx.Session.Insert(c); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			started := time.Now()
			if err := storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error { return tx.ClearAssistantData() }); err != nil {
				t.Fatal(err)
			}
			t.Logf("1000-conversation clear transaction: %s", time.Since(started))
			var conversations []models.AssistantConversation
			if err := engine.Where("uid = ?", 7).Find(&conversations); err != nil {
				t.Fatal(err)
			}
			byID := make(map[int64]*models.AssistantConversation, len(conversations))
			for i := range conversations {
				byID[conversations[i].Id] = &conversations[i]
			}
			for _, c := range conversations {
				if !c.Deleted || c.ActiveRunId != 0 || c.ReservedBytes != 0 || c.UsedBytes != 0 || c.MetadataVersion != 2 {
					t.Fatal("incomplete bulk tombstone", c)
				}
			}
			if err := storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
				var runs []models.AssistantMessage
				if err := tx.Session.Where("uid = ?", 7).Find(&runs); err != nil {
					return err
				}
				for i := range runs {
					run := &runs[i]
					if run.RunState != models.AssistantRunCanceled || !run.ReservationSettled || run.ReservedBytes != 0 {
						t.Fatal("unsettled canceled task", run)
					}
					c := byID[run.ConversationId]
					if err := tx.SettleRun(c, run, 100); err != nil {
						return err
					}
					if err := tx.ValidateDependencies(c, &models.AssistantInputDependencies{}); err != models.AssistantReasonContextChanged {
						t.Fatal("old worker was not fenced", err)
					}
				}
				if tx.State.UsedBytes != 0 || tx.State.ReservedBytes != 0 || tx.State.RunningCount != 0 || tx.State.QueuedCount != 0 {
					t.Fatal("late settlement changed cleared quota", tx.State)
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAssistantCleanupSkipsEmptyStates(t *testing.T) {
	storage, engine := assistantStorageFixture(t, settings.Sqlite3DbType)
	ctx := core.NewNullContext()
	for id := int64(1); id <= 1000; id++ {
		if _, err := engine.Insert(&models.AssistantUserState{Uid: id, Version: 5}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := engine.Insert(&models.AssistantConversation{Id: 2000, Uid: 1000, Deleted: true, Title: "erase"}); err != nil {
		t.Fatal(err)
	}
	cleaner := NewAssistantCleanupService(storage)
	if err := cleaner.Sweep(ctx); err != nil {
		t.Fatal(err)
	}
	n, err := engine.Count(&models.AssistantConversation{})
	if err != nil || n != 0 {
		t.Fatal("empty users delayed actual cleanup", n, err)
	}
	untouched, err := engine.Where("uid <> ? AND version = ?", 1000, 5).Count(&models.AssistantUserState{})
	if err != nil || untouched != 999 {
		t.Fatal("empty states were rewritten", untouched, err)
	}
}
