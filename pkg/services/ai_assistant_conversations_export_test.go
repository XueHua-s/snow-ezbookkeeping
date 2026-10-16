package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

func TestAssistantConversationExportBatchesKeepQuestionVersionGroups(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			s := assistantConversationFixture(t, kind)
			ctx := core.NewNullContext()
			first := conversationSend(t, s, "first", "", 0, "question one")
			conversationComplete(t, s, first, "answer 1")
			last := first
			for i := 2; i <= 4; i++ {
				var err error
				last, err = s.Retry(ctx, 7, models.AssistantRetryRequest{AssistantCommand: storageCommand(s.storage, fmt.Sprint("retry", i), 0), ConversationId: first.ConversationId, UserMessageId: first.UserMessageId, ExpectedContextVersion: last.ContextVersion, Intent: "regenerate"})
				if err != nil {
					t.Fatal(err)
				}
				conversationComplete(t, s, last, fmt.Sprint("answer ", i))
			}
			second := conversationSend(t, s, "second", last.ConversationId, last.ContextVersion, "question two")
			conversationComplete(t, s, second, "second answer")
			cfg := settings.DefaultAIAssistantExperienceConfig()
			cfg.ExportBatchSize = 2
			s.options.Config = func() settings.AIAssistantExperienceConfig { return cfg }
			for _, allVersions := range []bool{true, false} {
				var out bytes.Buffer
				if err := s.Export(ctx, 7, models.AssistantExportRequest{Id: first.ConversationId, Format: "json", IncludeVersions: allVersions}, &out); err != nil {
					t.Fatal(err)
				}
				var parsed struct {
					Conversations []struct {
						Messages []struct {
							Question models.AssistantMessageDTO    `json:"question"`
							Answers  []models.AssistantRunSnapshot `json:"answers"`
						} `json:"messages"`
					} `json:"conversations"`
				}
				if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
					t.Fatal(err)
				}
				if len(parsed.Conversations) != 1 || len(parsed.Conversations[0].Messages) != 2 {
					t.Fatal("question grouping was cut at batch boundary")
				}
				messages := parsed.Conversations[0].Messages
				want := 1
				if allVersions {
					want = 4
				}
				if len(messages[0].Answers) != want || len(messages[1].Answers) != 1 {
					t.Fatal("versions missing or duplicated")
				}
				for i, answer := range messages[0].Answers {
					if answer.UserMessageId != first.UserMessageId {
						t.Fatal("answer attached to wrong question")
					}
					if allVersions && answer.AttemptNumber != i+1 {
						t.Fatal("unstable version ordering")
					}
				}
				if !allVersions && messages[0].Answers[0].RunId != last.RunId {
					t.Fatal("default export did not select final successful answer")
				}
			}
		})
	}
}
