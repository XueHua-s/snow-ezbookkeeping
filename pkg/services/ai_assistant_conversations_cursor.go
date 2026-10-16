package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// A cursor fixes the query identity and upper time boundary, without retaining results.
type assistantConversationCursor struct {
	UID      int64  `json:"u"`
	Scope    string `json:"s"`
	Cutoff   int64  `json:"c"`
	Expires  int64  `json:"e"`
	Time     int64  `json:"t"`
	ID       string `json:"i"`
	Sequence int64  `json:"q,omitempty"`
}

func encodeAssistantConversationCursor(cursor assistantConversationCursor, secret []byte) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func decodeAssistantConversationCursor(value string, uid int64, scope string, secret []byte, now time.Time) (assistantConversationCursor, error) {
	var cursor assistantConversationCursor
	if len(value) > 4096 || len(secret) == 0 {
		return cursor, models.AssistantReasonCursorExpired
	}
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return cursor, models.AssistantReasonCursorExpired
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return cursor, models.AssistantReasonCursorExpired
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return cursor, models.AssistantReasonCursorExpired
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	if !hmac.Equal(signature, mac.Sum(nil)) || json.Unmarshal(payload, &cursor) != nil || cursor.UID != uid || cursor.Scope != scope || cursor.Expires <= now.UnixMilli() || cursor.Cutoff > now.UnixMilli() {
		return assistantConversationCursor{}, models.AssistantReasonCursorExpired
	}
	return cursor, nil
}

// LIKE uses ! as the explicit SQL escape so literal backslashes work on all three engines.
func assistantConversationSearchPattern(query string) string {
	return "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(query) + "%"
}
