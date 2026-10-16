package services

import (
	"encoding/json"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
	"testing"
	"time"
)

func TestAssistantMemoryConfirmForgetAndRunningRevocation(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) { AssistantMemoryConfirmForgetAndRunningRevocation(t, kind) })
	}
}
func AssistantMemoryConfirmForgetAndRunningRevocation(t *testing.T, kind string) {
	conv := assistantConversationFixture(t, kind)
	ctx := core.NewNullContext()
	cfg := &settings.Config{EnableAIAssistant: true}
	s := NewAssistantMemoryService(conv.storage, func() *settings.Config { return cfg })
	var id int64 = 9000
	s.generateID = func(uuid.UuidType) int64 { id++; return id }
	list, e := s.List(ctx, 7)
	if e != nil || list.Enabled {
		t.Fatal("memory not default off", e)
	}
	set, e := s.Settings(ctx, 7, models.AssistantMemorySettingsRequest{AssistantCommand: storageCommand(conv.storage, "enable", 0), ExpectedSettingsVersion: 1, Enabled: true})
	if e != nil {
		t.Fatal(e)
	}
	source := conversationSend(t, conv, "source", "", 0, "I prefer short answers")
	conversationComplete(t, conv, source, "Acknowledged as a preference, not yet saved.")
	saved, e := s.Save(ctx, 7, models.AssistantSaveMemoryRequest{AssistantCommand: storageCommand(conv.storage, "save", 0), ExpectedMemoryUseEpoch: set.MemoryUseEpoch, Text: "I prefer short answers", Key: "response style", Type: "preference", Intent: "add", SourceMessageIds: []string{source.UserMessageId}})
	if e != nil {
		t.Fatal(e)
	}
	active := conversationSend(t, conv, "active", source.ConversationId, source.ContextVersion, "Give a detailed answer now")
	var run *models.AssistantMessage
	_ = conv.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		rid, _ := assistantParseID(active.RunId)
		run, _ = assistantConversationMessage(tx.Session, 7, rid)
		return nil
	})
	contextService := NewAssistantContextService(conv.storage, func() *settings.Config { return cfg })
	snap, e := contextService.snapshot(ctx, run)
	if e != nil {
		t.Fatal(e)
	}
	text, e := s.selectContext(ctx, snap)
	if e != nil || text == "" || len(snap.Dependencies.Memories) != 1 {
		t.Fatal(text, e, snap.Dependencies)
	}
	_ = conv.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		raw, _ := json.Marshal(snap.Dependencies)
		run.InputDependenciesJSON = string(raw)
		_, e := tx.Session.Where("uid = ? AND id = ?", 7, run.Id).AllCols().Update(run)
		return e
	})
	deleted, e := s.Delete(ctx, 7, models.AssistantDeleteMemoryRequest{AssistantCommand: storageCommand(conv.storage, "forget", 0), Id: saved.MemoryId, ExpectedVersion: 1, ExpectedMemoryUseEpoch: saved.MemoryUseEpoch})
	if e != nil {
		t.Fatal(e)
	}
	_ = conv.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		current, _ := assistantConversationMessage(tx.Session, 7, run.Id)
		if current.RunState != models.AssistantRunCanceled {
			t.Fatal(current.RunState)
		}
		return nil
	})
	// Replaying the original delete cannot remove a later explicit user memory.
	newer, e := s.Save(ctx, 7, models.AssistantSaveMemoryRequest{AssistantCommand: storageCommand(conv.storage, "remember-new", 0), ExpectedMemoryUseEpoch: deleted.MemoryUseEpoch, Text: "Please answer in detail", Key: "response style", Type: "preference", Intent: "add"})
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.Delete(ctx, 7, models.AssistantDeleteMemoryRequest{AssistantCommand: storageCommand(conv.storage, "forget", 0), Id: saved.MemoryId, ExpectedVersion: 1, ExpectedMemoryUseEpoch: saved.MemoryUseEpoch})
	if e != nil {
		t.Fatal(e)
	}
	list, e = s.List(ctx, 7)
	if e != nil || len(list.Items) != 1 || list.Items[0].Id != newer.MemoryId {
		t.Fatal(list, e)
	}
}
func TestAssistantMemoryCandidatesRequireUserQuoteAndExclusion(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) { AssistantMemoryCandidatesRequireUserQuoteAndExclusion(t, kind) })
	}
}
func AssistantMemoryCandidatesRequireUserQuoteAndExclusion(t *testing.T, kind string) {
	conv := assistantConversationFixture(t, kind)
	ctx := core.NewNullContext()
	s := NewAssistantMemoryService(conv.storage, func() *settings.Config { return &settings.Config{} })
	var id int64 = 9500
	s.generateID = func(uuid.UuidType) int64 { id++; return id }
	source := conversationSend(t, conv, "source", "", 0, "I prefer short answers")
	conversationComplete(t, conv, source, "answer")
	batch := &assistantMemoryBatch{UID: 7, Sources: map[string]string{source.UserMessageId: "I prefer short answers"}}
	err := conv.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		// A model guess cannot become a candidate even with a valid source ID.
		if e := s.insertCandidate(tx, batch, assistantMemoryCandidate{Key: "health", Type: "preference", Text: "I prefer medical help", Quote: "I prefer medical help", SourceMessageID: source.UserMessageId}); e != nil {
			return e
		}
		if tx.State.CandidateMemoryCount != 0 {
			t.Fatal("invented quote persisted")
		}
		if e := s.insertCandidate(tx, batch, assistantMemoryCandidate{Key: "style", Type: "preference", Text: "I prefer short answers", Quote: "I prefer short answers", SourceMessageID: source.UserMessageId}); e != nil {
			return e
		}
		var m models.AssistantMemory
		_, e := tx.Session.Where("uid = ?", 7).Get(&m)
		if e != nil {
			return e
		}
		if e = assistantForgetMemory(tx, &m); e != nil {
			return e
		}
		return s.insertCandidate(tx, batch, assistantMemoryCandidate{Key: "style", Type: "preference", Text: "I prefer short answers", Quote: "I prefer short answers", SourceMessageID: source.UserMessageId})
	})
	if err != nil {
		t.Fatal(err)
	}
	state, e := conv.storage.GetUserState(ctx, 7)
	if e != nil || state.CandidateMemoryCount != 0 {
		t.Fatal(state, e)
	}
}

func TestAssistantMemoryBackgroundIdleBudgetAndDisableFence(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) { AssistantMemoryBackgroundIdleBudgetAndDisableFence(t, kind) })
	}
}
func AssistantMemoryBackgroundIdleBudgetAndDisableFence(t *testing.T, kind string) {
	conv := assistantConversationFixture(t, kind)
	ctx := core.NewNullContext()
	cfg := &settings.Config{EnableAIAssistant: true}
	s := NewAssistantMemoryService(conv.storage, func() *settings.Config { return cfg })
	var id int64 = 9800
	s.generateID = func(uuid.UuidType) int64 { id++; return id }
	_, e := s.Settings(ctx, 7, models.AssistantMemorySettingsRequest{AssistantCommand: storageCommand(conv.storage, "enable-bg", 0), ExpectedSettingsVersion: 1, Enabled: true})
	if e != nil {
		t.Fatal(e)
	}
	source := conversationSend(t, conv, "source-bg", "", 0, "I prefer short answers")
	conversationComplete(t, conv, source, "answer")
	batch, e := s.claimBatch(ctx, 7)
	if e != nil || batch != nil {
		t.Fatal("did not respect idle delay", e)
	}
	before := conv.storage.now()
	conv.storage.now = func() time.Time { return before.Add(11 * time.Minute) }
	batch, e = s.claimBatch(ctx, 7)
	if e != nil || batch == nil {
		t.Fatal("idle batch missing", e)
	}
	p := &contextSummaryProvider{before: func() {
		_, err := s.Settings(ctx, 7, models.AssistantMemorySettingsRequest{AssistantCommand: storageCommand(conv.storage, "disable-bg", 0), ExpectedSettingsVersion: 2, Enabled: false})
		if err != nil {
			t.Error(err)
		}
	}}
	s.providers = func() (provider.LargeLanguageModelToolProvider, provider.LargeLanguageModelToolProvider, error) {
		return p, nil, nil
	}
	if e = s.extractBatch(ctx, batch); e != nil {
		t.Fatal(e)
	}
	state, e := conv.storage.GetUserState(ctx, 7)
	if e != nil || state.MemoryEnabled || state.CandidateMemoryCount != 0 || state.BackgroundCalls != 1 {
		t.Fatal(state, e)
	}
	// A previously claimed batch is not sent again after its epoch is revoked.
	if e = s.extractBatch(ctx, batch); e != nil {
		t.Fatal(e)
	}
	if p.calls != 1 {
		t.Fatal("revoked batch re-sent", p.calls)
	}
}
