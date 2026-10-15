package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"xorm.io/xorm"
)

// Search uses two stable keyset phases (titles, then messages). Joining the
// owner-filtered conversation and question avoids an SQL roundtrip per stored
// conversation, while selecting only bounded matching bodies.
func (s *AssistantConversationService) Search(ctx core.Context, uid int64, r models.AssistantSearchRequest) (models.AssistantSearchResult, error) {
	result := models.AssistantSearchResult{Items: []models.AssistantSearchMatch{}}
	parent := context.Context(context.Background())
	requestID, locale := "", ""
	if ctx != nil {
		parent = ctx
		requestID = ctx.GetContextId()
		locale = ctx.GetClientLocale()
	}
	bounded, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()
	ctx = core.NewAssistantContext(bounded, requestID, locale, "")
	if !utf8.ValidString(r.Query) || strings.TrimSpace(r.Query) == "" || utf8.RuneCountInString(r.Query) > 100 {
		return result, models.AssistantReasonInputTooLong
	}
	hash := sha256.Sum256([]byte(r.Query))
	scope := fmt.Sprintf("search-v2:%s:%t:%t", hex.EncodeToString(hash[:]), r.Archived, r.IncludeVersions)
	now := s.now()
	cursor := assistantConversationCursor{UID: uid, Scope: scope, Cutoff: now.UnixMilli(), Expires: now.Add(15 * time.Minute).UnixMilli()}
	var err error
	if r.Cursor != "" {
		cursor, err = decodeAssistantConversationCursor(r.Cursor, uid, scope, s.options.Secret, now)
		if err != nil {
			return result, err
		}
	}
	limits := s.limits()
	limit := assistantConversationPageLimit(r.Limit, limits.PageSize, limits.PageSizeMax)
	bytes := 64
	appendMatch := func(match models.AssistantSearchMatch) bool {
		raw, _ := json.Marshal(match)
		if len(result.Items) >= limit || bytes+len(raw) > limits.PageBytes {
			return false
		}
		result.Items = append(result.Items, match)
		bytes += len(raw) + 1
		return true
	}
	next := func() error {
		result.NextCursor, err = encodeAssistantConversationCursor(cursor, s.options.Secret)
		return err
	}
	err = s.storage.WithUserRead(ctx, uid, func(session *xorm.Session) error {
		if ctx != nil && ctx.Err() != nil {
			return ctx.Err()
		}
		after := int64(0)
		if cursor.ID != "" {
			var err error
			after, err = assistantParseID(cursor.ID)
			if err != nil {
				return err
			}
		}
		if cursor.Sequence == 0 {
			query := session.Where("uid = ? AND deleted = ? AND created_at <= ? AND id > ? AND title LIKE ? ESCAPE '!'", uid, false, cursor.Cutoff, after, assistantConversationSearchPattern(r.Query))
			if !r.Archived {
				query = query.And("archived = ?", false)
			}
			var titles []models.AssistantConversation
			if err := query.Asc("id").Limit(limit + 1).Find(&titles); err != nil {
				return err
			}
			for i := range titles {
				c := &titles[i]
				if !appendMatch(models.AssistantSearchMatch{ConversationId: assistantConversationID(c.Id), Title: c.Title, Snippet: c.Title}) {
					return next()
				}
				cursor.ID = assistantConversationID(c.Id)
			}
			cursor.Sequence = 1
			cursor.ID = ""
			after = 0
		}
		conversationTable := session.Engine().TableName(&models.AssistantConversation{})
		messageTable := session.Engine().TableName(&models.AssistantMessage{})
		quotedMessage := session.Engine().Quote(messageTable)
		query := session.Table(&models.AssistantMessage{}).Alias("m").Join("INNER", []string{conversationTable, "c"}, "c.id = m.conversation_id AND c.uid = m.uid").Join("INNER", []string{messageTable, "q"}, "q.id = m.user_message_id AND q.uid = m.uid").Where("m.uid = ? AND c.deleted = ? AND m.created_at <= ? AND m.id > ? AND m.content LIKE ? ESCAPE '!'", uid, false, cursor.Cutoff, after, assistantConversationSearchPattern(r.Query))
		if !r.Archived {
			query = query.And("c.archived = ?", false)
		}
		if !r.IncludeVersions {
			query = query.And("(m.role = ? OR m.id = q.selected_assistant_message_id OR (q.selected_assistant_message_id = 0 AND NOT EXISTS (SELECT 1 FROM "+quotedMessage+" newer WHERE newer.uid = m.uid AND newer.user_message_id = m.user_message_id AND newer.attempt_number > m.attempt_number)))", "user")
		}
		var messages []models.AssistantMessage
		if err := query.Select("m.*").Asc("m.id").Limit(limit - len(result.Items) + 1).Find(&messages); err != nil {
			return err
		}
		// Only result conversations are fetched, bounded by the response page size.
		titles := map[int64]string{}
		for i := range messages {
			m := &messages[i]
			title, ok := titles[m.ConversationId]
			if !ok {
				c, err := assistantConversationLoad(session, uid, m.ConversationId)
				if err != nil {
					return err
				}
				title = c.Title
				titles[m.ConversationId] = title
			}
			match := models.AssistantSearchMatch{ConversationId: assistantConversationID(m.ConversationId), MessageId: assistantConversationID(m.Id), Title: title, Snippet: assistantSearchSnippet(m.Content, r.Query)}
			if m.Role == "assistant" {
				match.MessageId = assistantConversationID(m.UserMessageId)
				match.VersionId = assistantConversationID(m.Id)
			}
			if !appendMatch(match) {
				return next()
			}
			cursor.ID = assistantConversationID(m.Id)
		}
		return nil
	})
	return result, err
}
func assistantSearchSnippet(content, query string) string {
	text := []rune(content)
	needle := []rune(strings.ToLower(query))
	lower := []rune(strings.ToLower(content))
	start := 0
	for i := 0; i+len(needle) <= len(lower); i++ {
		if string(lower[i:i+len(needle)]) == string(needle) {
			start = i - 40
			if start < 0 {
				start = 0
			}
			break
		}
	}
	end := start + 200
	if end > len(text) {
		end = len(text)
	}
	snippet := string(text[start:end])
	if start > 0 {
		snippet = "…" + snippet
	}
	if end < len(text) {
		snippet += "…"
	}
	return snippet
}
