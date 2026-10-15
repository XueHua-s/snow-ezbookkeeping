package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
)

const assistantCursorPurpose = "ezbookkeeping/assistant-query-cursor/v1"
const assistantMaxCursorBytes = 64 * 1024

// Keys are private implementation data, never an editable tool parameter.
// Numeric keys are strings to preserve exact database ordering through JSON.
type assistantPageKey struct {
	ID     string `json:"id,omitempty"`
	Time   string `json:"time,omitempty"`
	Amount string `json:"amount,omitempty"`
	Group  string `json:"group,omitempty"`
}
type assistantCursor struct {
	Version         int                `json:"version"`
	Tool            data.ToolName      `json:"tool"`
	UID             string             `json:"uid"`
	Args            assistantQueryArgs `json:"args"`
	Timezone        string             `json:"timezone"`
	HierarchyDigest string             `json:"hierarchy_digest"`
	HighWater       assistantPageKey   `json:"high_water"`
	Last            assistantPageKey   `json:"last"`
	IssuedAt        int64              `json:"issued_at"`
	ExpiresAt       int64              `json:"expires_at"`
}

func assistantCursorError() error {
	return &data.ToolError{Code: data.ToolErrorInvalidCursor, Message: "Cursor is invalid or expired; restart the query"}
}

// encoding/json sorts string map keys, struct field order is fixed, and our
// domain uses integers/strings only. Re-encoding is checked before verification.
func assistantCanonicalJSON(v any) ([]byte, error) { return json.Marshal(v) }
func assistantCursorKey(secret []byte, purpose string) []byte {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(purpose))
	return m.Sum(nil)
}
func assistantSignCursor(c assistantCursor, key []byte) (string, error) {
	raw, err := assistantCanonicalJSON(c)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	mac.Write(raw)
	token := base64.RawURLEncoding.EncodeToString(raw) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if len(token) > assistantMaxCursorBytes {
		return "", &data.ToolError{Code: data.ToolErrorResultTooLarge, Message: "Cursor exceeds result limit"}
	}
	return token, nil
}
func (s *AssistantQueryService) decodeCursor(token string, tool data.ToolName) (*assistantCursor, error) {
	if len(token) == 0 || len(token) > assistantMaxCursorBytes {
		return nil, assistantCursorError()
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, assistantCursorError()
	}
	raw, e := base64.RawURLEncoding.DecodeString(parts[0])
	if e != nil {
		return nil, assistantCursorError()
	}
	signature, e := base64.RawURLEncoding.DecodeString(parts[1])
	if e != nil {
		return nil, assistantCursorError()
	}
	mac := hmac.New(sha256.New, s.cursorKey)
	mac.Write(raw)
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return nil, assistantCursorError()
	}
	var c assistantCursor
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if e = dec.Decode(&c); e != nil {
		return nil, assistantCursorError()
	}
	canonical, e := assistantCanonicalJSON(c)
	if e != nil || !bytes.Equal(raw, canonical) {
		return nil, assistantCursorError()
	}
	now := s.now().Unix()
	if c.Version != 1 || c.Tool != tool || c.UID != s.uidString || c.Timezone != s.location.String() || c.IssuedAt > now || c.ExpiresAt <= now || c.ExpiresAt-c.IssuedAt != int64((24*time.Hour)/time.Second) {
		return nil, assistantCursorError()
	}
	return &c, nil
}
func assistantArgsInput(name data.ToolName, a assistantQueryArgs) map[string]any {
	raw, _ := json.Marshal(a)
	m, _ := assistantStrictJSON(raw)
	domain := assistantToolDomain(name)
	for k := range m {
		if domain.Fields[k] == nil {
			delete(m, k)
		}
	}
	if f, ok := m["filter"].(map[string]any); ok {
		if len(a.Filter.TagIDs) == 0 {
			delete(f, "tag_match")
		}
	}
	return m
}
func (s *AssistantQueryService) parseArguments(name data.ToolName, raw []byte) (assistantQueryArgs, *assistantCursor, error) {
	var zero assistantQueryArgs
	if len(raw) > 96*1024 {
		return zero, nil, assistantInvalid("arguments")
	}
	m, e := assistantStrictJSON(raw)
	if e != nil {
		return zero, nil, e
	}
	domain := assistantToolDomain(name)
	if domain == nil {
		return zero, nil, assistantInvalid("tool")
	}
	if e = domain.validate(m, "arguments"); e != nil {
		return zero, nil, e
	}
	token, _ := m["cursor"].(string)
	delete(m, "cursor")
	without, _ := json.Marshal(m)
	if len(without) > 32*1024 {
		return zero, nil, assistantInvalid("arguments")
	}
	if token == "" {
		a, e := normalizeAssistantArgs(name, m)
		return a, nil, e
	}
	cursor, e := s.decodeCursor(token, name)
	if e != nil {
		return zero, nil, e
	}
	bound := assistantArgsInput(name, cursor.Args)
	// Each explicitly supplied top-level parameter is normalized using exactly
	// the first-page rules, then compared as a whole with the signed condition.
	for k, v := range m {
		// Strict providers supply every optional property, using null for an
		// omitted argument. Null therefore does not reset a signed condition.
		if v != nil {
			bound[k] = v
		}
	}
	if e = domain.validate(bound, "arguments"); e != nil {
		return zero, nil, assistantCursorError()
	}
	a, e := normalizeAssistantArgs(name, bound)
	if e != nil {
		return zero, nil, assistantCursorError()
	}
	b1, _ := assistantCanonicalJSON(a)
	b2, _ := assistantCanonicalJSON(cursor.Args)
	if !bytes.Equal(b1, b2) {
		return zero, nil, assistantCursorError()
	}
	return a, cursor, nil
}
func (e *assistantQueryExecution) NextCursor(high, last assistantPageKey) (string, error) {
	now := e.Service.now().Unix()
	return assistantSignCursor(assistantCursor{Version: 1, Tool: e.Tool, UID: e.Service.uidString, Args: e.Args, Timezone: e.Service.location.String(), HierarchyDigest: e.HierarchyDigest, HighWater: high, Last: last, IssuedAt: now, ExpiresAt: now + int64((24*time.Hour)/time.Second)}, e.Service.cursorKey)
}
func assistantHierarchyDigest(accounts, categories []string) string {
	b, _ := assistantCanonicalJSON(struct {
		Accounts   []string `json:"accounts"`
		Categories []string `json:"categories"`
	}{accounts, categories})
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
