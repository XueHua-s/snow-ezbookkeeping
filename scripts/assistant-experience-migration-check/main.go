// This opt-in helper operates only on dedicated migration fixture databases.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
	"xorm.io/xorm"
)

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	if len(os.Args) != 3 {
		panic("usage: helper config mode")
	}
	cfg, e := settings.LoadConfiguration(os.Args[1])
	must(e)
	db := cfg.DatabaseConfig
	if db.DatabaseType == "sqlite3" {
		if !strings.Contains(db.DatabasePath, "ebk_experience_migration_") {
			panic("not a fixture path")
		}
	} else if !strings.HasPrefix(db.DatabaseName, "ebk_experience_migration_") {
		panic("not a fixture database")
	}
	must(datastore.InitializeDataStore(cfg))
	session := datastore.Container.UserDataStore.Choose(7).NewAssistantSession(core.NewNullContext())
	engine := session.Engine()
	session.Close()
	defer engine.Close()
	storage := services.NewAssistantStorageService(datastore.Container)
	switch os.Args[2] {
	case "legacy":
		_, e = engine.Insert(&models.Transaction{TransactionId: 9001, Uid: 7, Amount: 9007199254740993, TransactionTime: 1000000, Comment: "synthetic ledger sentinel"}, &models.AIAssistantEmbedding{Uid: 7, TransactionId: 9001, EmbeddingModel: "legacy-sentinel", ContentHash: "unchanged", VectorData: "[0.125,-0.25]", CreatedUnixTime: 123, UpdatedUnixTime: 456})
		must(e)
	case "seed":
		seed(storage, engine)
	case "manifest":
		manifest(engine, false)
	case "legacy-manifest":
		manifest(engine, true)
	case "recover":
		recoverRuns(storage)
	default:
		panic("unknown mode")
	}
}
func seed(storage *services.AssistantStorageService, engine *xorm.Engine) {
	must(storage.WithUserTransaction(core.NewNullContext(), 7, func(tx *services.AssistantStorageTx) error {
		c := &models.AssistantConversation{Id: 1, Uid: 7, Title: "restored synthetic conversation", ActiveRunId: 3, NextTurnSequence: 2, ContextVersion: 1, MetadataVersion: 1, SourceVersion: 1, MemoryMode: "inherit", SchemaVersion: 1}
		q := &models.AssistantMessage{Id: 2, Uid: 7, ConversationId: 1, UserMessageId: 2, Role: "user", Content: "synthetic question", TurnSequence: 1, SchemaVersion: 1}
		run := &models.AssistantMessage{Id: 3, Uid: 7, ConversationId: 1, UserMessageId: 2, Role: "assistant", AttemptNumber: 1, TurnSequence: 1, Content: "saved partial answer", RunState: models.AssistantRunQueued, SnapshotRevision: 1, SchemaVersion: 1, CreatedAt: time.Now().Add(-time.Hour).UnixMilli()}
		lim := settings.DefaultAIAssistantExperienceConfig().Limits
		if e := tx.ReserveConversation(c, &lim); e != nil {
			return e
		}
		if e := tx.ReserveRun(c, run, &lim); e != nil {
			return e
		}
		if e := tx.MarkRunRunning(run, &lim); e != nil {
			return e
		}
		run.RunState = models.AssistantRunRunning
		run.Owner = "expired-fixture-owner"
		run.LeaseEpoch = 4
		run.LeaseExpiresAt = time.Now().Add(-time.Minute).UnixMilli()
		// Completed historical answer has an expired cursor; restoring must not revive it.
		priorC := &models.AssistantConversation{Id: 10, Uid: 7, Title: "expired cursor fixture", NextTurnSequence: 2, ContextVersion: 1, MetadataVersion: 1, MemoryMode: "inherit", SchemaVersion: 1}
		priorQ := &models.AssistantMessage{Id: 11, Uid: 7, ConversationId: 10, UserMessageId: 11, Role: "user", Content: "prior question", TurnSequence: 1, SelectedAssistantMessageId: 12, SchemaVersion: 1}
		priorA := &models.AssistantMessage{Id: 12, Uid: 7, ConversationId: 10, UserMessageId: 11, Role: "assistant", Content: "prior answer", TurnSequence: 1, AttemptNumber: 1, RunState: models.AssistantRunCompleted, SnapshotRevision: 1, SchemaVersion: 1, Continuation: "synthetic-expired-cursor", ContinuationExpiresAt: time.Now().Add(-time.Hour).UnixMilli()}
		if e := tx.ReserveConversation(priorC, &lim); e != nil {
			return e
		}
		for _, row := range []any{c, q, run, priorC, priorQ, priorA, &models.AssistantMemory{Id: 4, Uid: 7, Key: "style", Text: "synthetic concise preference", Status: "active", Version: 1, SchemaVersion: 1}, &models.AssistantCommandReceipt{Uid: 7, ClientRequestId: "migration-receipt", CommandType: "send", ConversationId: 1, ResultJSON: `{"status":"accepted","conversationId":"1"}`, CanonicalRequestHash: "synthetic"}} {
			if _, e := tx.Session.Insert(row); e != nil {
				return e
			}
		}
		return nil
	}))
	fmt.Println("seed: all five assistant record types populated")
}
func manifest(engine *xorm.Engine, legacyOnly bool) {
	output := map[string]any{}
	rows := map[string]any{"ledger": &[]models.Transaction{}, "embedding": &[]models.AIAssistantEmbedding{}, "conversation": &[]models.AssistantConversation{}, "message": &[]models.AssistantMessage{}, "memory": &[]models.AssistantMemory{}, "state": &[]models.AssistantUserState{}, "receipt": &[]models.AssistantCommandReceipt{}}
	for name, target := range rows {
		if legacyOnly && name != "ledger" && name != "embedding" {
			continue
		}
		must(engine.Find(target))
		raw, e := json.Marshal(target)
		must(e)
		hash := sha256.Sum256(raw)
		output[name] = map[string]any{"sha256": hex.EncodeToString(hash[:]), "rows": json.RawMessage(raw)}
	}
	raw, e := json.MarshalIndent(output, "", "  ")
	must(e)
	fmt.Println(string(raw))
}
func recoverRuns(storage *services.AssistantStorageService) {
	calls := 0
	worker := services.NewAssistantRunService(storage, services.AssistantRunOptions{GenerationAvailable: func(core.Context, int64) bool { calls++; return true }})
	worker.Start(context.Background())
	defer worker.Stop(context.Background())
	end := time.Now().Add(10 * time.Second)
	for {
		snapshot, e := worker.GetRunSnapshot(core.NewNullContext(), 7, "3")
		must(e)
		if snapshot.State == models.AssistantRunInterrupted {
			if snapshot.Content != "saved partial answer" {
				panic("partial content lost")
			}
			break
		}
		if time.Now().After(end) {
			panic("expired run not interrupted")
		}
		time.Sleep(50 * time.Millisecond)
	}
	if calls != 0 {
		panic("restored expired run attempted generation")
	}
	conversations := services.NewAssistantConversationService(storage, services.AssistantConversationOptions{GenerationAvailable: func(core.Context, int64) bool { return true }, GenerateID: func(_ uuid.UuidType) int64 { return 100 }})
	_, e := conversations.Send(core.NewNullContext(), 7, models.AssistantSendRequest{AssistantCommand: models.AssistantCommand{ClientRequestId: "expired-cursor-check", RequestIssuedAt: time.Now().UTC().Format(time.RFC3339Nano)}, ConversationId: "10", ExpectedContextVersion: 1, Mode: "next_page", Content: "next page", SourceMessageId: "12"})
	if e != models.AssistantReasonCursorExpired {
		panic(fmt.Sprintf("expired cursor accepted: %v", e))
	}
	fmt.Println("recovery: interrupted; partial body preserved; zero provider eligibility calls")
}
