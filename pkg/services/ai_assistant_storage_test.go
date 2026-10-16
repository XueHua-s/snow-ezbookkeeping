package services

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"xorm.io/xorm"
	xlog "xorm.io/xorm/log"
)

func assistantStorageFixture(t *testing.T, kind string) (*AssistantStorageService, *xorm.Engine) {
	t.Helper()
	config := &settings.DatabaseConfig{DatabaseType: kind, DatabasePath: filepath.Join(t.TempDir(), "storage.db"), MaxOpenConnection: 8, MaxIdleConnection: 8}
	if kind != settings.Sqlite3DbType {
		config = assistantExternalServiceConfig(t, kind)
		config.MaxOpenConnection = 8
		config.MaxIdleConnection = 8
	}
	saved := *datastore.Container
	err := datastore.InitializeDataStore(&settings.Config{DatabaseConfig: config})
	container := *datastore.Container
	*datastore.Container = saved
	if err != nil {
		t.Fatal(err)
	}
	session := container.UserDataStore.Choose(7).NewAssistantSession(core.NewNullContext())
	engine := session.Engine()
	session.Close()
	t.Cleanup(func() { engine.Close() })
	tables := []any{&models.AssistantConversation{}, &models.AssistantMessage{}, &models.AssistantMemory{}, &models.AssistantUserState{}, &models.AssistantCommandReceipt{}, &models.AIAssistantEmbedding{}}
	for i := 0; i < 2; i++ {
		if err := container.UserDataStore.SyncStructs(tables...); err != nil {
			t.Fatal(err)
		}
	}
	s := NewAssistantStorageService(&container)
	s.now = func() time.Time { return time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC) }
	return s, engine
}

func storageCommand(s *AssistantStorageService, id string, epoch int64) models.AssistantCommand {
	return models.AssistantCommand{ClientRequestId: id, RequestIssuedAt: s.now().Format(time.RFC3339Nano), ExpectedDeletionEpoch: epoch}
}

func TestAssistantStorageDatabaseContracts(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			t.Run("receipt_epoch_rollback_and_retention", func(t *testing.T) {
				s, engine := assistantStorageFixture(t, kind)
				ctx := core.NewNullContext()
				var calls int
				create := func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
					calls++
					_, err := tx.Session.Insert(&models.AssistantConversation{Id: 1, Uid: 7, Title: "private 原文", SourceVersion: 1})
					return models.AssistantCommandResult{ConversationId: "1"}, err
				}
				cmd := storageCommand(s, "create", 0)
				first, err := s.ExecuteCommand(ctx, 7, "send", cmd, "private 原文", create)
				if err != nil {
					t.Fatal(err)
				}
				again, err := s.ExecuteCommand(ctx, 7, "send", cmd, "private 原文", create)
				if err != nil || again != first || calls != 1 {
					t.Fatal("not idempotent", err, calls)
				}
				_, err = s.ExecuteCommand(ctx, 7, "delete", cmd, "private 原文", create)
				if !errors.Is(err, models.AssistantReasonConversationConflict) {
					t.Fatal("cross action reuse", err)
				}
				_, err = s.ExecuteCommand(ctx, 7, "send", cmd, "changed", create)
				if !errors.Is(err, models.AssistantReasonConversationConflict) {
					t.Fatal("changed body", err)
				}
				unknown, err := s.GetCommand(ctx, 8, "create")
				if err != nil || unknown.Status != "unknown" {
					t.Fatal("foreign receipt", unknown, err)
				}
				fail := storageCommand(s, "rollback", 0)
				_, err = s.ExecuteCommand(ctx, 7, "send", fail, fail, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
					_, err := tx.Session.Insert(&models.AssistantMessage{Id: 5, Uid: 7, ConversationId: 1, UserMessageId: 5, Content: "must rollback"})
					if err != nil {
						return models.AssistantCommandResult{}, err
					}
					return models.AssistantCommandResult{}, errors.New("private database error")
				})
				if !errors.Is(err, models.AssistantReasonStorageUnavailable) {
					t.Fatal(err)
				}
				count, _ := engine.Where("id = ?", 5).Count(&models.AssistantMessage{})
				if count != 0 {
					t.Fatal("partial message persisted")
				}
				unknown, _ = s.GetCommand(ctx, 7, "rollback")
				if unknown.Status != "unknown" {
					t.Fatal("failed receipt persisted")
				}
				clear := storageCommand(s, "clear", 0)
				clearCalls := 0
				clearing := func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
					clearCalls++
					tx.State.DeletionEpoch++
					_, err := tx.Session.Where("uid = ?", 7).Cols("deleted").Update(&models.AssistantConversation{Deleted: true})
					return models.AssistantCommandResult{}, err
				}
				if _, err = s.ExecuteCommand(ctx, 7, "clear", clear, clear, clearing); err != nil {
					t.Fatal(err)
				}
				late := storageCommand(s, "never-accepted", 0)
				if _, err = s.ExecuteCommand(ctx, 7, "send", late, late, create); !errors.Is(err, models.AssistantReasonDataEpochConflict) {
					t.Fatal("late command revived", err)
				}
				newer := storageCommand(s, "new", 1)
				if _, err = s.ExecuteCommand(ctx, 7, "branch", newer, newer, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
					_, err := tx.Session.Insert(&models.AssistantConversation{Id: 2, Uid: 7})
					return models.AssistantCommandResult{ConversationId: "2"}, err
				}); err != nil {
					t.Fatal(err)
				}
				if _, err = s.ExecuteCommand(ctx, 7, "clear", clear, clear, clearing); err != nil || clearCalls != 1 {
					t.Fatal("clear replay", err)
				}
				found, _ := engine.Where("id = ? AND deleted = ?", 2, false).Exist(&models.AssistantConversation{})
				if !found {
					t.Fatal("clear deleted newer data")
				}
				deleted, _ := s.GetCommand(ctx, 7, "create")
				if deleted.Status != "deleted" {
					t.Fatal("deleted receipt not redacted")
				}
				// Live-object receipts outlive 48 hours. Unknown old commands do not.
				s.now = func() time.Time { return time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC) }
				if err := s.PruneReceipts(ctx, 7, 100); err != nil {
					t.Fatal(err)
				}
				live, _ := s.GetCommand(ctx, 7, "new")
				if live.Status != "accepted" {
					t.Fatal("live receipt pruned")
				}
				_, err = s.ExecuteCommand(ctx, 7, "clear", clear, clear, clearing)
				if !errors.Is(err, models.AssistantReasonConversationConflict) {
					t.Fatal("expired unknown replay", err)
				}
			})
			t.Run("concurrent_quota_and_once_settlement", func(t *testing.T) {
				s, _ := assistantStorageFixture(t, kind)
				ctx := core.NewNullContext()
				limits := settings.DefaultAIAssistantExperienceConfig().Limits
				otherInstance := NewAssistantStorageService(s.ServiceUsingDB.container)
				otherInstance.now = s.now
				limits.Conversations = 2
				limits.UserBytes = 2 * limits.ConversationHeaderBytes
				var accepted atomic.Int64
				var wg sync.WaitGroup
				start := make(chan struct{})
				for i := 0; i < 12; i++ {
					wg.Add(1)
					go func(i int) {
						defer wg.Done()
						<-start
						instance := s
						if i%2 == 1 {
							instance = otherInstance
						}
						cmd := storageCommand(instance, fmt.Sprint(i), 0)
						_, err := instance.ExecuteCommand(ctx, 7, "branch", cmd, cmd, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
							conv := &models.AssistantConversation{Id: int64(i + 1), Uid: 7}
							if err := tx.ReserveConversation(conv, &limits); err != nil {
								return models.AssistantCommandResult{}, err
							}
							_, err := tx.Session.Insert(conv)
							return models.AssistantCommandResult{ConversationId: fmt.Sprint(i + 1)}, err
						})
						if err == nil {
							accepted.Add(1)
						} else if !errors.Is(err, models.AssistantReasonQuotaExceeded) {
							t.Errorf("concurrent mutation: %v", err)
						}
					}(i)
				}
				close(start)
				wg.Wait()
				if accepted.Load() != 2 {
					t.Fatalf("accepted %d", accepted.Load())
				}
				state, err := s.GetUserState(ctx, 7)
				if err != nil || state.ConversationCount != 2 || state.ReservedBytes != int64(limits.UserBytes) {
					t.Fatal("quota state", state, err)
				}
				// Another user is independent. Reserve a run and settle its bytes once.
				limits = settings.DefaultAIAssistantExperienceConfig().Limits
				if err := s.WithUserTransaction(ctx, 8, func(tx *AssistantStorageTx) error {
					conv := &models.AssistantConversation{Id: 100, Uid: 8}
					run := &models.AssistantMessage{Id: 101, Uid: 8, ConversationId: 100, UserMessageId: 102, AttemptNumber: 1, RunState: models.AssistantRunQueued}
					if err := tx.ReserveConversation(conv, &limits); err != nil {
						return err
					}
					if err := tx.ReserveRun(conv, run, &limits); err != nil {
						return err
					}
					if _, err := tx.Session.Insert(conv); err != nil {
						return err
					}
					_, err := tx.Session.Insert(run)
					return err
				}); err != nil {
					t.Fatal(err)
				}
				for i := 0; i < 2; i++ {
					if err := s.WithUserTransaction(ctx, 8, func(tx *AssistantStorageTx) error {
						conv := &models.AssistantConversation{}
						run := &models.AssistantMessage{}
						if _, err := tx.Session.ID(100).Get(conv); err != nil {
							return err
						}
						if _, err := tx.Session.ID(101).Get(run); err != nil {
							return err
						}
						if err := tx.SettleRun(conv, run, 123); err != nil {
							return err
						}
						if _, err := tx.Session.ID(100).AllCols().Update(conv); err != nil {
							return err
						}
						_, err := tx.Session.ID(101).AllCols().Update(run)
						return err
					}); err != nil {
						t.Fatal(err)
					}
				}
				state, err = s.GetUserState(ctx, 8)
				if err != nil || state.UsedBytes != 123 || state.QueuedCount != 0 || state.RunningCount != 0 || state.ReservedBytes != int64(limits.ConversationHeaderBytes) {
					t.Fatal("double settlement", state, err)
				}
			})
			t.Run("bounded_fields_logging_and_legacy", func(t *testing.T) {
				s, engine := assistantStorageFixture(t, kind)
				var logs bytes.Buffer
				engine.SetLogger(xlog.NewSimpleLogger(&logs))
				engine.ShowSQL(true)
				secret := "SENSITIVE_CANARY_财务"
				ctx := core.NewNullContext()
				if err := s.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
					_, err := tx.Session.Insert(&models.AssistantMessage{Id: 1, Uid: 7, UserMessageId: 1, Content: strings.Repeat("文", 10922), ProcessSummary: secret, QueriesJSON: strings.Repeat("x", 32000), ReferencesJSON: strings.Repeat("y", 32000)})
					return err
				}); err != nil {
					t.Fatal(err)
				}
				// Duplicate insert errors must not include bound parameters in SQL logs.
				_ = s.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
					_, err := tx.Session.Insert(&models.AssistantMessage{Id: 1, Uid: 7, UserMessageId: 1, Content: secret})
					return err
				})
				if strings.Contains(logs.String(), secret) {
					t.Fatal("sensitive SQL logs")
				}
				engine.ShowSQL(false)
				embedding := &models.AIAssistantEmbedding{Uid: 7, TransactionId: 9, EmbeddingModel: "legacy", ContentHash: "retained", VectorData: "[1,2]"}
				if _, err := engine.Insert(embedding); err != nil {
					t.Fatal(err)
				}
				if err := engine.Sync2(&models.AssistantConversation{}, &models.AssistantMessage{}, &models.AssistantMemory{}, &models.AssistantUserState{}, &models.AssistantCommandReceipt{}); err != nil {
					t.Fatal(err)
				}
				var restored models.AIAssistantEmbedding
				found, err := engine.Where("uid = ?", 7).Get(&restored)
				if err != nil || !found || restored.VectorData != "[1,2]" {
					t.Fatal("legacy changed", err)
				}
			})
		})
	}
}
