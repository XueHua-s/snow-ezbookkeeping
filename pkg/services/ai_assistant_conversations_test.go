package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

func assistantConversationFixture(t *testing.T, kind string) *AssistantConversationService {
	storage, _ := assistantStorageFixture(t, kind)
	var id atomic.Int64
	id.Store(1000)
	s := NewAssistantConversationService(storage, AssistantConversationOptions{Secret: []byte("conversation-test-key"), GenerateID: func(uuid.UuidType) int64 { return id.Add(1) }, GenerationAvailable: func(core.Context, int64) bool { return true }})
	s.now = storage.now
	return s
}
func conversationSend(t *testing.T, s *AssistantConversationService, key, conversation string, version int64, content string) models.AssistantCommandResult {
	t.Helper()
	result, err := s.Send(core.NewNullContext(), 7, models.AssistantSendRequest{AssistantCommand: storageCommand(s.storage, key, 0), ConversationId: conversation, ExpectedContextVersion: version, Mode: "chat", Content: content, Locale: "zh_CN", Timezone: "Asia/Shanghai"})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func conversationComplete(t *testing.T, s *AssistantConversationService, result models.AssistantCommandResult, content string) {
	t.Helper()
	err := s.storage.WithUserTransaction(core.NewNullContext(), 7, func(tx *AssistantStorageTx) error {
		cid, _ := assistantParseID(result.ConversationId)
		rid, _ := assistantParseID(result.RunId)
		c, err := assistantConversationLoad(tx.Session, 7, cid)
		if err != nil {
			return err
		}
		run, err := assistantConversationMessage(tx.Session, 7, rid)
		if err != nil {
			return err
		}
		q, err := assistantConversationMessage(tx.Session, 7, run.UserMessageId)
		if err != nil {
			return err
		}
		run.Content = content
		run.RunState = models.AssistantRunCompleted
		run.CompletedAt = tx.Now.UnixMilli()
		run.SnapshotRevision++
		actual, err := AssistantAttemptStorageBytes(q, run)
		if err != nil {
			return err
		}
		if err = tx.SettleRun(c, run, actual); err != nil {
			return err
		}
		c.ActiveRunId = 0
		q.SelectedAssistantMessageId = run.Id
		if _, err = tx.Session.Where("uid = ? AND id = ?", 7, q.Id).AllCols().Update(q); err != nil {
			return err
		}
		if _, err = tx.Session.Where("uid = ? AND id = ?", 7, run.Id).AllCols().Update(run); err != nil {
			return err
		}
		return assistantConversationSave(tx, c)
	})
	if err != nil {
		t.Fatal(err)
	}
}
func assistantTestConversationsImmutableBranchAndUID(t *testing.T, kind string) {
	s := assistantConversationFixture(t, kind)
	ctx := core.NewNullContext()
	one := conversationSend(t, s, "one", "", 0, "原始问题")
	conversationComplete(t, s, one, "first answer")
	two := conversationSend(t, s, "two", one.ConversationId, one.ContextVersion, "第二问题")
	conversationComplete(t, s, two, "future answer excluded")
	branch, err := s.Branch(ctx, 7, models.AssistantBranchRequest{AssistantCommand: storageCommand(s.storage, "branch", 0), SourceConversationId: one.ConversationId, SourceMessageId: two.UserMessageId, ExpectedContextVersion: two.ContextVersion, Kind: "edit"})
	if err != nil {
		t.Fatal(err)
	}
	var frozen []AssistantFrozenTurn
	err = s.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		id, _ := assistantParseID(branch.ConversationId)
		c, err := assistantConversationLoad(tx.Session, 7, id)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(c.FrozenPrefixJSON), &frozen)
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(frozen) != 1 || frozen[0].AssistantMessageId != one.RunId {
		t.Fatalf("branch leaked future or lost past: %#v", frozen)
	}
	three := conversationSend(t, s, "three", one.ConversationId, two.ContextVersion, "父会话继续")
	conversationComplete(t, s, three, "parent appended")
	branchSend := conversationSend(t, s, "branch-send", branch.ConversationId, branch.ContextVersion, "edited question")
	if branchSend.ConversationId != branch.ConversationId {
		t.Fatal("branch not reused")
	}
	if _, err = s.Get(ctx, 8, models.AssistantGetConversationRequest{Id: one.ConversationId}); !errors.Is(err, models.AssistantReasonNotAccessible) {
		t.Fatalf("foreign conversation: %v", err)
	}
	if _, err = s.GetMessage(ctx, 8, one.RunId); !errors.Is(err, models.AssistantReasonNotAccessible) {
		t.Fatalf("foreign version: %v", err)
	}
	if _, err = s.Branch(ctx, 8, models.AssistantBranchRequest{AssistantCommand: storageCommand(s.storage, "steal", 0), SourceConversationId: one.ConversationId, SourceMessageId: one.UserMessageId, ExpectedContextVersion: three.ContextVersion, Kind: "edit"}); !errors.Is(err, models.AssistantReasonNotAccessible) {
		t.Fatalf("foreign branch: %v", err)
	}
	state, err := s.State(ctx, 7, one.ConversationId)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Delete(ctx, 7, models.AssistantDeleteConversationRequest{AssistantCommand: storageCommand(s.storage, "delete", 0), Id: one.ConversationId, ExpectedMetadataVersion: state.MetadataVersion}); err != nil {
		t.Fatal(err)
	}
	detail, err := s.Get(ctx, 7, models.AssistantGetConversationRequest{Id: branch.ConversationId})
	if err != nil {
		t.Fatal(err)
	}
	if !detail.Conversation.SourceInvalidated || len(detail.Messages) != 1 {
		t.Fatal("source invalidation deleted independent branch or retained seed")
	}
	if _, err = s.Get(ctx, 7, models.AssistantGetConversationRequest{Id: one.ConversationId}); !errors.Is(err, models.AssistantReasonNotAccessible) {
		t.Fatal("deleted history accessible")
	}
}
func assistantTestConversationsReceiptClearAndLiteralSearch(t *testing.T, kind string) {
	s := assistantConversationFixture(t, kind)
	ctx := core.NewNullContext()
	r := models.AssistantSendRequest{AssistantCommand: storageCommand(s.storage, "send", 0), Mode: "chat", Content: "中文 100%_\\ 文本"}
	first, err := s.Send(ctx, 7, r)
	if err != nil {
		t.Fatal(err)
	}
	again, err := s.Send(ctx, 7, r)
	if err != nil || first != again {
		t.Fatal("send dedupe", err)
	}
	conversationComplete(t, s, first, "literal %_\\ answer")
	for _, query := range []string{"中文", "%_\\"} {
		found, err := s.Search(ctx, 7, models.AssistantSearchRequest{Query: query, Limit: 1})
		if err != nil || len(found.Items) != 1 {
			t.Fatalf("literal search %q: %#v %v", query, found, err)
		}
		seen := map[string]bool{}
		for page := 0; page < 10; page++ {
			for _, match := range found.Items {
				key := match.ConversationId + ":" + match.MessageId + ":" + match.VersionId
				if seen[key] {
					t.Fatal("duplicate search result")
				}
				seen[key] = true
			}
			if found.NextCursor == "" {
				break
			}
			found, err = s.Search(ctx, 7, models.AssistantSearchRequest{Query: query, Limit: 1, Cursor: found.NextCursor})
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	var export bytes.Buffer
	if err = s.Export(ctx, 7, models.AssistantExportRequest{All: true, Format: "json", IncludeVersions: true}, &export); err != nil {
		t.Fatal(err)
	}
	if !json.Valid(export.Bytes()) || !strings.Contains(export.String(), "literal") {
		t.Fatalf("bad export: %s", export.String())
	}
	clear := models.AssistantClearRequest{AssistantCommand: storageCommand(s.storage, "clear", 0)}
	cleared, err := s.Clear(ctx, 7, clear)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Send(ctx, 7, models.AssistantSendRequest{AssistantCommand: storageCommand(s.storage, "delayed-never-accepted", 0), Mode: "chat", Content: "must not resurrect"}); !errors.Is(err, models.AssistantReasonDataEpochConflict) {
		t.Fatal("old unaccepted request revived", err)
	}
	newR := models.AssistantSendRequest{AssistantCommand: storageCommand(s.storage, "new", cleared.DeletionEpoch), Mode: "chat", Content: "new after clear"}
	newResult, err := s.Send(ctx, 7, newR)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Clear(ctx, 7, clear); err != nil {
		t.Fatal(err)
	}
	if _, err = s.State(ctx, 7, newResult.ConversationId); err != nil {
		t.Fatal("old clear receipt erased new data", err)
	}
}
func assistantTestConversationsPagingAndRetryPreservesQuestion(t *testing.T, kind string) {
	s := assistantConversationFixture(t, kind)
	ctx := core.NewNullContext()
	var last models.AssistantCommandResult
	for i := 0; i < 7; i++ {
		last = conversationSend(t, s, fmt.Sprint("send", i), last.ConversationId, last.ContextVersion, fmt.Sprint("question", i))
		conversationComplete(t, s, last, strings.Repeat("答", 1000))
	}
	retry, err := s.Retry(ctx, 7, models.AssistantRetryRequest{AssistantCommand: storageCommand(s.storage, "regen", 0), ConversationId: last.ConversationId, UserMessageId: last.UserMessageId, ExpectedContextVersion: last.ContextVersion, Intent: "regenerate"})
	if err != nil {
		t.Fatal(err)
	}
	conversationComplete(t, s, retry, "new answer")
	detail, err := s.Get(ctx, 7, models.AssistantGetConversationRequest{Id: last.ConversationId, Limit: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Messages) != 3 || !detail.HasMore || detail.Conversation.TurnCount != 7 {
		t.Fatal("logical page or retry round count incorrect")
	}
	seen := map[string]bool{}
	for {
		for _, message := range detail.Messages {
			if seen[message.Id] {
				t.Fatal("duplicate logical message")
			}
			seen[message.Id] = true
		}
		if !detail.HasMore {
			break
		}
		detail, err = s.Get(ctx, 7, models.AssistantGetConversationRequest{Id: last.ConversationId, Limit: 3, BeforeTurnSequence: detail.Messages[0].TurnSequence})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(seen) != 7 {
		t.Fatal("missing messages")
	}
	versions, err := s.Versions(ctx, 7, models.AssistantVersionsRequest{UserMessageId: last.UserMessageId, Limit: 1})
	if err != nil || len(versions.Items) != 1 || versions.NextCursor == "" {
		t.Fatal("version page", err)
	}
	more, err := s.Versions(ctx, 7, models.AssistantVersionsRequest{UserMessageId: last.UserMessageId, Limit: 1, Cursor: versions.NextCursor})
	if err != nil || len(more.Items) != 1 || more.Items[0].Id == versions.Items[0].Id {
		t.Fatal("version pagination", err)
	}
}

func TestAssistantConversationsDatabaseContracts(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			t.Run("ImmutableBranchAndUID", func(t *testing.T) { assistantTestConversationsImmutableBranchAndUID(t, kind) })
			t.Run("ReceiptClearAndLiteralSearch", func(t *testing.T) { assistantTestConversationsReceiptClearAndLiteralSearch(t, kind) })
			t.Run("PagingAndRetryPreservesQuestion", func(t *testing.T) { assistantTestConversationsPagingAndRetryPreservesQuestion(t, kind) })
		})
	}
}

func TestAssistantConversationByteBudgetAndCompactedAncestor(t *testing.T) {
	s := assistantConversationFixture(t, settings.Sqlite3DbType)
	ctx := core.NewNullContext()
	first := conversationSend(t, s, "ancestor", "", 0, "祖先条件保留")
	conversationComplete(t, s, first, strings.Repeat("a", 30000))
	branch, err := s.Branch(ctx, 7, models.AssistantBranchRequest{AssistantCommand: storageCommand(s.storage, "first-branch", 0), SourceConversationId: first.ConversationId, SourceMessageId: first.UserMessageId, ExpectedContextVersion: first.ContextVersion, Kind: "continue"})
	if err != nil {
		t.Fatal(err)
	}
	err = s.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		id, _ := assistantParseID(branch.ConversationId)
		c, err := assistantConversationLoad(tx.Session, 7, id)
		if err != nil {
			return err
		}
		c.SeedContext = `{"topic":"祖先条件保留","schemaVersion":1}`
		c.FrozenPrefixJSON = ""
		if err = assistantSettleConversationHeader(tx, c); err != nil {
			return err
		}
		return assistantConversationSave(tx, c)
	})
	if err != nil {
		t.Fatal(err)
	}
	own := conversationSend(t, s, "own", branch.ConversationId, branch.ContextVersion, "中间会话问题")
	conversationComplete(t, s, own, strings.Repeat("b", 30000))
	child, err := s.Branch(ctx, 7, models.AssistantBranchRequest{AssistantCommand: storageCommand(s.storage, "child", 0), SourceConversationId: branch.ConversationId, SourceMessageId: own.UserMessageId, ExpectedContextVersion: own.ContextVersion, Kind: "continue"})
	if err != nil {
		t.Fatal(err)
	}
	err = s.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		id, _ := assistantParseID(child.ConversationId)
		c, err := assistantConversationLoad(tx.Session, 7, id)
		if err != nil {
			return err
		}
		if !strings.Contains(c.SeedContext, "祖先条件保留") || !strings.Contains(c.FrozenPrefixJSON, own.RunId) {
			t.Fatal("compacted ancestor or own prefix lost")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4; i++ {
		own = conversationSend(t, s, fmt.Sprintf("large-%d", i), own.ConversationId, own.ContextVersion, "large")
		conversationComplete(t, s, own, strings.Repeat("c", 30000))
	}
	cfg := settings.DefaultAIAssistantExperienceConfig()
	cfg.Limits.PageBytes = 70000
	s.options.Config = func() settings.AIAssistantExperienceConfig { return cfg }
	page, err := s.Get(ctx, 7, models.AssistantGetConversationRequest{Id: own.ConversationId, Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) > cfg.Limits.PageBytes || !page.HasMore || len(page.Messages) != 2 {
		t.Fatalf("page ignored byte budget: bytes=%d count=%d", len(raw), len(page.Messages))
	}
}
