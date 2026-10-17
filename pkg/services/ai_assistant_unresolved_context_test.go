package services

import (
	"strings"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

func TestAssistantPreparePreservesFailedQuestionWithoutFailedAnswer(t *testing.T) {
	t.Chdir("../..")
	conv := assistantConversationFixture(t, settings.Sqlite3DbType)
	first := conversationSend(t, conv, "first", "", 0, "比较2026与2024")
	conversationComplete(t, conv, first, "previous completed analysis")
	failed := conversationSend(t, conv, "failed", first.ConversationId, first.ContextVersion, "2022呢，2023为何更多？账户42，按人民币")
	conversationComplete(t, conv, failed, "UNVERIFIED PARTIAL ANSWER MUST NOT ENTER CONTEXT")
	err := conv.storage.WithUserTransaction(core.NewNullContext(), 7, func(tx *AssistantStorageTx) error {
		qid, _ := assistantParseID(failed.UserMessageId)
		rid, _ := assistantParseID(failed.RunId)
		if _, e := tx.Session.Where("uid = ? AND id = ?", 7, qid).Cols("selected_assistant_message_id").Update(&models.AssistantMessage{SelectedAssistantMessageId: 0}); e != nil {
			return e
		}
		_, e := tx.Session.Where("uid = ? AND id = ?", 7, rid).Cols("run_state").Update(&models.AssistantMessage{RunState: models.AssistantRunIncomplete})
		return e
	})
	if err != nil {
		t.Fatal(err)
	}
	current := conversationSend(t, conv, "continue", failed.ConversationId, failed.ContextVersion, "继续查询")
	prepared, err := contextBudgetService(conv, &contextSummaryProvider{}).Prepare(core.NewNullContext(), contextBudgetRun(t, conv, current))
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, message := range prepared.Input.Messages {
		joined += message.Content + "\n"
	}
	if !strings.Contains(joined, "2022呢，2023为何更多？账户42，按人民币") || strings.Contains(joined, "UNVERIFIED PARTIAL") {
		t.Fatal("failed question lost or failed answer leaked")
	}
	if prepared.Input.Messages[len(prepared.Input.Messages)-1].Content != "继续查询" {
		t.Fatal("latest intent moved")
	}
	// Once answered in a later turn, old failed questions must participate in
	// normal bounded compaction instead of accumulating after newer summaries.
	conversationComplete(t, conv, current, strings.Repeat("历史分析需要压缩。", 600))
	next := conversationSend(t, conv, "next", current.ConversationId, current.ContextVersion, "更正：2023按USD比较")
	p := &contextSummaryProvider{}
	prepared, err = contextBudgetService(conv, p).Prepare(core.NewNullContext(), contextBudgetRun(t, conv, next))
	if err != nil {
		t.Fatal(err)
	}
	sources := strings.Join(p.sources, "\n")
	if p.calls == 0 || !strings.Contains(sources, "2022呢，2023为何更多？账户42，按人民币") || strings.Contains(sources, "UNVERIFIED PARTIAL") {
		t.Fatal("unresolved question not safely compacted")
	}
	for _, message := range prepared.Input.Messages {
		if strings.Contains(message.Content, "2022呢，2023为何更多？账户42，按人民币") {
			t.Fatal("covered failed question reinserted after newer summary")
		}
	}
	if prepared.Input.Messages[len(prepared.Input.Messages)-1].Content != "更正：2023按USD比较" {
		t.Fatal("new correction lost")
	}
}

func TestAssistantUnresolvedIntentRetainsCorrectionOrder(t *testing.T) {
	turns := []assistantContextTurn{{Question: &models.AssistantMessage{TurnSequence: 2, Content: "correction CNY"}, Answer: &models.AssistantMessage{Content: "done"}}}
	unresolved := []*models.AssistantMessage{{TurnSequence: 1, Content: "old USD"}, {TurnSequence: 3, Content: "new year 2023"}}
	messages := assistantContextMessages(assistantContextTurns(turns, unresolved))
	var users []string
	for _, message := range messages {
		if message.Role == data.ConversationRoleUser {
			users = append(users, message.Content)
		}
	}
	if len(users) != 3 || !strings.Contains(users[0], "old USD") || !strings.Contains(users[1], "correction CNY") || !strings.Contains(users[2], "new year 2023") {
		t.Fatal(users)
	}
}
