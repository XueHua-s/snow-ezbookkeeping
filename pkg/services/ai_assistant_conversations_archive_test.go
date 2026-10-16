package services

import (
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

func TestAssistantConversationIncludeArchivedListsAndSearchesBothStates(t *testing.T) {
	s := assistantConversationFixture(t, settings.Sqlite3DbType)
	ctx := core.NewNullContext()
	active := conversationSend(t, s, "active", "", 0, "matching active")
	conversationComplete(t, s, active, "answer")
	archived := conversationSend(t, s, "archived", "", 0, "matching archived")
	conversationComplete(t, s, archived, "answer")
	yes := true
	if _, err := s.Update(ctx, 7, models.AssistantUpdateConversationRequest{AssistantCommand: storageCommand(s.storage, "archive", 0), Id: archived.ConversationId, ExpectedMetadataVersion: archived.MetadataVersion, Archived: &yes}); err != nil {
		t.Fatal(err)
	}
	for _, include := range []bool{false, true} {
		result, err := s.List(ctx, 7, models.AssistantListRequest{Archived: include})
		if err != nil {
			t.Fatal(err)
		}
		want := 1
		if include {
			want = 2
		}
		if len(result.Items) != want {
			t.Fatalf("includeArchived=%v: got %d conversations, want %d", include, len(result.Items), want)
		}
		seen := map[string]bool{}
		for _, c := range result.Items {
			seen[c.Id] = true
		}
		if !seen[active.ConversationId] || seen[archived.ConversationId] != include {
			t.Fatal("archived filter excluded active conversations or included archived without opt-in")
		}
		search, err := s.Search(ctx, 7, models.AssistantSearchRequest{Query: "matching", Archived: include})
		if err != nil {
			t.Fatal(err)
		}
		found := map[string]bool{}
		for _, match := range search.Items {
			found[match.ConversationId] = true
		}
		if !found[active.ConversationId] || found[archived.ConversationId] != include {
			t.Fatal("search/list archive semantics differ")
		}
	}
}
