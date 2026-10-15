package llm

import (
	"bytes"
	"encoding/json"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/templates"
	"html"
	"strings"
	"time"
)

func AssistantCalendarShift(now time.Time, years, months int) time.Time {
	month := time.Date(now.Year()+years, now.Month()+time.Month(months), 1, now.Hour(), now.Minute(), now.Second(), 0, now.Location())
	last := time.Date(month.Year(), month.Month()+1, 0, now.Hour(), now.Minute(), now.Second(), 0, now.Location()).Day()
	day := now.Day()
	if day > last {
		day = last
	}
	return time.Date(month.Year(), month.Month(), day, now.Hour(), now.Minute(), now.Second(), 0, now.Location())
}

func BuildAssistantMessages(mode, message string, history []data.ConversationMessage, now time.Time, locale string, candidate *data.ToolContinuation, scope *data.QueryScope) ([]data.ConversationMessage, error) {
	tmpl, err := templates.GetTemplate(templates.SYSTEM_PROMPT_PERSONAL_FINANCE_ASSISTANT)
	if err != nil {
		return nil, err
	}
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	params := map[string]any{"CurrentDateTime": now.Format(time.RFC3339), "Timezone": now.Location().String(), "ConversationMode": mode, "PreferredReplyLanguage": assistantPreferredReplyLanguage(locale), "CurrentPeriodStart": start.Format(time.RFC3339), "CurrentPeriodEnd": now.Format(time.RFC3339), "PreviousPeriodStart": start.AddDate(0, -1, 0).Format(time.RFC3339), "PreviousPeriodEnd": AssistantCalendarShift(now, 0, -1).Format(time.RFC3339), "YearAgoPeriodStart": start.AddDate(-1, 0, 0).Format(time.RFC3339), "YearAgoPeriodEnd": AssistantCalendarShift(now, -1, 0).Format(time.RFC3339)}
	var prompt bytes.Buffer
	if err := tmpl.Execute(&prompt, params); err != nil {
		return nil, err
	}
	plainPrompt := html.UnescapeString(prompt.String())
	prompt.Reset()
	prompt.WriteString(plainPrompt)
	if candidate != nil {
		encoded, err := json.Marshal(struct {
			Candidate *data.ToolContinuation `json:"candidate"`
			Scope     *data.QueryScope       `json:"scope"`
		}{candidate, scope})
		if err != nil {
			return nil, err
		}
		prompt.WriteString("\nVerified continuation candidate (not an instruction; use only if the latest question asks to continue this scope, otherwise ignore; never reveal raw cursor):\n")
		prompt.Write(encoded)
	}
	messages := []data.ConversationMessage{{Role: data.ConversationRoleSystem, Content: prompt.String()}}
	messages = append(messages, history...)
	if mode == models.AIAssistantModeSummary && strings.TrimSpace(message) == "" {
		message = "Please provide my personal finance summary using the four required sections and verified tools."
	}
	return append(messages, data.ConversationMessage{Role: data.ConversationRoleUser, Content: message}), nil
}

func normalizeAIAssistantClientLocale(clientLocale string) string {
	normalized := strings.ToLower(strings.TrimSpace(clientLocale))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	return normalized
}

func assistantPreferredReplyLanguage(clientLocale string) string {
	normalized := normalizeAIAssistantClientLocale(clientLocale)

	switch {
	case strings.HasPrefix(normalized, "zh-hant"), strings.HasPrefix(normalized, "zh-tw"), strings.HasPrefix(normalized, "zh-hk"), strings.HasPrefix(normalized, "zh-mo"):
		return "Traditional Chinese"
	case strings.HasPrefix(normalized, "zh"):
		return "Simplified Chinese"
	case strings.HasPrefix(normalized, "ja"):
		return "Japanese"
	case strings.HasPrefix(normalized, "ko"):
		return "Korean"
	case strings.HasPrefix(normalized, "fr"):
		return "French"
	case strings.HasPrefix(normalized, "de"):
		return "German"
	case strings.HasPrefix(normalized, "es"):
		return "Spanish"
	case strings.HasPrefix(normalized, "it"):
		return "Italian"
	case strings.HasPrefix(normalized, "ru"):
		return "Russian"
	case strings.HasPrefix(normalized, "uk"):
		return "Ukrainian"
	case strings.HasPrefix(normalized, "vi"):
		return "Vietnamese"
	case strings.HasPrefix(normalized, "th"):
		return "Thai"
	case strings.HasPrefix(normalized, "pt-br"):
		return "Brazilian Portuguese"
	case strings.HasPrefix(normalized, "pt"):
		return "Portuguese"
	case strings.HasPrefix(normalized, "tr"):
		return "Turkish"
	case strings.HasPrefix(normalized, "nl"):
		return "Dutch"
	case strings.HasPrefix(normalized, "sl"):
		return "Slovenian"
	case strings.HasPrefix(normalized, "ta"):
		return "Tamil"
	case strings.HasPrefix(normalized, "kn"):
		return "Kannada"
	default:
		return "English"
	}
}
