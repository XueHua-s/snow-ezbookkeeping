package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func assistantCursorFixture(t *testing.T, s *AssistantQueryService, raw string) string {
	t.Helper()
	args, _, err := s.parseArguments(data.ToolQueryTransactions, []byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	now := s.now().Unix()
	cursor := assistantCursor{Version: 1, Tool: data.ToolQueryTransactions, UID: s.uidString, Args: args, Timezone: s.location.String(), HierarchyDigest: assistantHierarchyDigest(nil, nil), HighWater: assistantPageKey{ID: "9007199254740999", Time: "4102444800000"}, Last: assistantPageKey{ID: "9007199254740993", Time: "1767225600000"}, IssuedAt: now, ExpiresAt: now + 86400}
	token, err := assistantSignCursor(cursor, s.cursorKey)
	if err != nil {
		t.Fatal(err)
	}
	return token
}
func TestAssistantCursorOnlyCanonicalEquivalenceAndMismatch(t *testing.T) {
	s := assistantParserForTest()
	token := assistantCursorFixture(t, s, `{"filter":{"account_ids":["2","1","2"],"start_time":"2026-01-01T08:00:00+08:00","currency":"cny","amount_min":"1"},"page_size":50}`)
	cases := []struct {
		name  string
		extra map[string]any
		valid bool
	}{
		{"cursor only", nil, true},
		{"explicit default page", map[string]any{"page_size": 50}, true},
		{"equivalent filter", map[string]any{"filter": map[string]any{"account_ids": []string{"01", "2"}, "start_time": "2026-01-01T00:00:00Z", "currency": "CNY", "amount_min": "01.00"}}, true},
		{"different page", map[string]any{"page_size": 51}, false},
		{"different amount", map[string]any{"filter": map[string]any{"account_ids": []string{"1", "2"}, "start_time": "2026-01-01T00:00:00Z", "currency": "CNY", "amount_min": "2.00"}}, false},
		{"null optional filter is omitted", map[string]any{"filter": nil}, true},
		{"different ordering", map[string]any{"order_by": map[string]any{"field": "time", "direction": "asc"}}, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			m := map[string]any{"cursor": token}
			for k, v := range tt.extra {
				m[k] = v
			}
			raw, _ := json.Marshal(m)
			args, cursor, err := s.parseArguments(data.ToolQueryTransactions, raw)
			if (err == nil) != tt.valid {
				t.Fatalf("valid=%v error=%v", tt.valid, err)
			}
			if err == nil && (cursor == nil || *args.Filter.StartMillis != 1767225600000 || *args.Filter.MinMinor != 100) {
				t.Fatal("derived fields not restored")
			}
		})
	}
	defaultToken := assistantCursorFixture(t, s, `{}`)
	raw, _ := json.Marshal(map[string]any{"cursor": defaultToken, "filter": map[string]any{"types": []string{"income", "expense", "transfer", "balance_modification"}}})
	if _, _, err := s.parseArguments(data.ToolQueryTransactions, raw); err != nil {
		t.Fatal("explicit all-types differs from default", err)
	}
}
func TestAssistantCursorAuthenticationExpiryAndPurpose(t *testing.T) {
	s := assistantParserForTest()
	token := assistantCursorFixture(t, s, `{}`)
	now := s.now()
	cases := []struct {
		name   string
		mutate func(*AssistantQueryService)
		token  func(string) string
	}{
		{"cross uid", func(s *AssistantQueryService) { s.uidString = "8" }, nil},
		{"new secret", func(s *AssistantQueryService) {
			s.cursorKey = assistantCursorKey([]byte("rotated"), assistantCursorPurpose)
		}, nil},
		{"other purpose", func(s *AssistantQueryService) {
			s.cursorKey = assistantCursorKey([]byte("isolated-test-secret"), "login-token")
		}, nil},
		{"timezone", func(s *AssistantQueryService) { s.location = time.FixedZone("Asia/Test", 8*3600) }, nil},
		{"expires exactly 24h", func(s *AssistantQueryService) { s.now = func() time.Time { return now.Add(24 * time.Hour) } }, nil},
		{"future issuance", func(s *AssistantQueryService) { s.now = func() time.Time { return now.Add(-time.Second) } }, nil},
		{"tampered payload", nil, func(token string) string {
			parts := strings.Split(token, ".")
			b, _ := base64.RawURLEncoding.DecodeString(parts[0])
			b = []byte(strings.Replace(string(b), `"uid":"7"`, `"uid":"8"`, 1))
			return base64.RawURLEncoding.EncodeToString(b) + "." + parts[1]
		}},
		{"tampered MAC", nil, func(token string) string {
			parts := strings.Split(token, ".")
			sig, _ := base64.RawURLEncoding.DecodeString(parts[1])
			sig[0] ^= 0xff
			return parts[0] + "." + base64.RawURLEncoding.EncodeToString(sig)
		}},
		{"oversized", nil, func(string) string { return strings.Repeat("x", 64*1024+1) }},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			copy := *s
			tok := token
			if tt.mutate != nil {
				tt.mutate(&copy)
			}
			if tt.token != nil {
				tok = tt.token(tok)
			}
			if _, err := copy.decodeCursor(tok, data.ToolQueryTransactions); err == nil {
				t.Fatal("invalid cursor accepted")
			}
		})
	}
	copy := *s
	copy.now = func() time.Time { return now.Add(24*time.Hour - time.Second) }
	if _, err := copy.decodeCursor(token, data.ToolQueryTransactions); err != nil {
		t.Fatal(err)
	}
	if _, err := s.decodeCursor(token, data.ToolAggregateTransactions); err == nil {
		t.Fatal("cross tool accepted")
	}
	// Canonical encoding prevents an alternate representation from becoming a
	// second valid encoding even when a trusted issuer accidentally signs it.
	parts := strings.Split(token, ".")
	raw, _ := base64.RawURLEncoding.DecodeString(parts[0])
	raw = append([]byte(" "), raw...)
	mac := hmac.New(sha256.New, s.cursorKey)
	mac.Write(raw)
	noncanonical := base64.RawURLEncoding.EncodeToString(raw) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if _, err := s.decodeCursor(noncanonical, data.ToolQueryTransactions); err == nil {
		t.Fatal("noncanonical JSON accepted")
	}
}
func TestAssistantCursorFixedSigningVector(t *testing.T) {
	// Independently computed with Python hashlib/hmac, frozen here. Canonical
	// JSON uses fixed struct field order and lexicographic map keys, no floats.
	key := assistantCursorKey([]byte("isolated-test-secret"), assistantCursorPurpose)
	if hex.EncodeToString(key) != "9d973e6ce9016167076e27a03ac19337b662f7f8b94342d4a4a50e9e50af074e" {
		t.Fatalf("derived=%s", hex.EncodeToString(key))
	}
	s := assistantParserForTest()
	token := assistantCursorFixture(t, s, `{}`)
	parts := strings.Split(token, ".")
	if parts[1] != "LSJiHr1coNZqKqLPbH_q4Gkxd5tVI2U7j3SnrtMHJIw" {
		t.Fatalf("MAC=%s payload=%s", parts[1], parts[0])
	}
	raw, err := assistantCanonicalJSON(map[string]any{"z": "<>&", "a": map[string]any{"b": 2, "a": 1}})
	if err != nil || string(raw) != `{"a":{"a":1,"b":2},"z":"\u003c\u003e\u0026"}` {
		t.Fatalf("canonical=%s", raw)
	}
}
func TestAssistantCursorReauthorizationAndParentExpansion(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	s.handlers[data.ToolQueryTransactions] = func(e *assistantQueryExecution) (data.ToolResult, error) {
		token, err := e.NextCursor(assistantPageKey{Time: "4102444800000", ID: "104"}, assistantPageKey{Time: "1000000", ID: "100"})
		if err != nil {
			return data.ToolResult{}, err
		}
		return data.ToolResult{OK: true, Data: assistantJSON([]string{"100"}), ReturnedCount: 1, HasMore: true, NextCursor: token}, nil
	}
	first := assistantCall(s, data.ToolQueryTransactions, `{"filter":{"account_ids":["10"],"category_ids":["30"]}}`)
	if !first.OK {
		t.Fatal(first.Error)
	}
	candidate := data.ToolContinuation{Tool: data.ToolQueryTransactions, Cursor: first.NextCursor}
	if err := s.ValidateContinuation(core.NewNullContext(), candidate); err != nil {
		t.Fatal(err)
	}
	cursor, err := s.decodeCursor(candidate.Cursor, candidate.Tool)
	if err != nil || strings.Join(cursor.Args.Filter.AccountIDs, ",") != "10" || strings.Join(cursor.Args.Filter.CategoryIDs, ",") != "30" {
		t.Fatal("original parents not retained", cursor, err)
	}
	if _, err := engine.Insert(&models.Account{AccountId: 14, Uid: 7, ParentAccountId: 10, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Name: "new child", Currency: "CNY"}); err != nil {
		t.Fatal(err)
	}
	err = s.ValidateContinuation(core.NewNullContext(), candidate)
	if err == nil || assistantSafeError(err).Code != data.ToolErrorInvalidCursor {
		t.Fatal("changed hierarchy accepted", err)
	}
	if _, err := engine.Exec("UPDATE account SET deleted=1 WHERE account_id=10"); err != nil {
		t.Fatal(err)
	}
	err = s.ValidateContinuation(core.NewNullContext(), candidate)
	if err == nil || assistantSafeError(err).Code != data.ToolErrorNotFoundOrNotAccessible {
		t.Fatal("deleted original parent not reauthorized", err)
	}
}

func TestAssistantNormalizeToolArgumentsCacheKey(t *testing.T) {
	s := assistantParserForTest()
	normalize := func(raw string) string {
		t.Helper()
		b, err := s.NormalizeToolArguments(data.ToolCall{Name: data.ToolQueryTransactions, Arguments: []byte(raw)})
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	defaults := normalize(`{}`)
	for _, raw := range []string{`{"filter":null,"order_by":null,"cursor":null,"page_size":null}`, `{"filter":{},"page_size":50,"order_by":{"field":"time","direction":"desc"}}`, `{"filter":{"types":["income","expense","transfer","balance_modification"]}}`} {
		if normalize(raw) != defaults {
			t.Fatal("equivalent defaults produce different keys")
		}
	}
	if normalize(`{"filter":{"account_ids":["2","1","2"]}}`) != normalize(`{"filter":{"account_ids":["01","2"]}}`) {
		t.Fatal("deduplicated IDs differ")
	}
	if normalize(`{"filter":{"keyword":"a"}}`) == normalize(`{"filter":{"keyword":"b"}}`) {
		t.Fatal("different filters collide")
	}
	token := assistantCursorFixture(t, s, `{}`)
	cursor, err := s.decodeCursor(token, data.ToolQueryTransactions)
	if err != nil {
		t.Fatal(err)
	}
	cursor.Last.ID = "9007199254740994"
	other, err := assistantSignCursor(*cursor, s.cursorKey)
	if err != nil {
		t.Fatal(err)
	}
	first, _ := json.Marshal(map[string]any{"cursor": token})
	second, _ := json.Marshal(map[string]any{"cursor": other})
	if normalize(string(first)) == normalize(string(second)) {
		t.Fatal("different cursor pages collide")
	}
}

func TestAssistantCursorStrictProviderNullFieldsPreserveBoundConditions(t *testing.T) {
	s := assistantParserForTest()
	token := assistantCursorFixture(t, s, `{"filter":{"account_ids":["2"],"currency":"CNY","amount_min":"9"},"page_size":17,"order_by":{"field":"time","direction":"asc"}}`)
	raw, _ := json.Marshal(map[string]any{"cursor": token, "filter": nil, "page_size": nil, "order_by": nil})
	args, cursor, err := s.parseArguments(data.ToolQueryTransactions, raw)
	if err != nil || cursor == nil || args.PageSize != 17 || args.OrderBy.Direction != "asc" || args.Filter.AmountMin == nil || *args.Filter.AmountMin != "9.00" {
		t.Fatalf("strict nullable continuation changed conditions: %+v %v", args, err)
	}
}

func TestAssistantDescribeContinuationScopeDoesNotExecute(t *testing.T) {
	s, _ := assistantServiceFixture(t)
	s.handlers[data.ToolQueryTransactions] = func(e *assistantQueryExecution) (data.ToolResult, error) {
		token, err := e.NextCursor(assistantPageKey{ID: "104", Time: "4102444800000"}, assistantPageKey{ID: "100", Time: "1000000"})
		return data.ToolResult{OK: true, Data: assistantJSON([]string{"100"}), NextCursor: token, HasMore: true}, err
	}
	first := assistantCall(s, data.ToolQueryTransactions, `{"filter":{"account_ids":["10"],"keyword":"ABC"}}`)
	if !first.OK {
		t.Fatal(first.Error)
	}
	s.handlers[data.ToolQueryTransactions] = func(*assistantQueryExecution) (data.ToolResult, error) {
		t.Fatal("continuation description executed tool")
		return data.ToolResult{}, nil
	}
	scope, err := s.DescribeContinuation(core.NewNullContext(), data.ToolContinuation{Tool: data.ToolQueryTransactions, Cursor: first.NextCursor})
	if err != nil || scope == nil || scope.Keyword != "abc" || strings.Join(scope.AccountIDs, ",") != "10,11" || !scope.IncludesHidden {
		t.Fatalf("scope=%+v err=%v", scope, err)
	}
}

func TestAssistantContinuationRequiresSignedNonemptyCursor(t *testing.T) {
	s := assistantParserForTest()
	s.handlers = map[data.ToolName]assistantToolHandler{data.ToolQueryTransactions: func(*assistantQueryExecution) (data.ToolResult, error) {
		t.Fatal("unsigned continuation executed handler")
		return data.ToolResult{}, nil
	}}
	// No database exists in this fixture: rejecting before a read view is part of
	// the contract, not merely rejecting after an unnecessary ledger read.
	candidate := data.ToolContinuation{Tool: data.ToolQueryTransactions, Cursor: ""}
	scope, err := s.DescribeContinuation(core.NewNullContext(), candidate)
	if scope != nil || err == nil || assistantSafeError(err).Code != data.ToolErrorInvalidCursor {
		t.Fatal("unsigned description accepted", scope, err)
	}
	if err = s.ValidateContinuation(core.NewNullContext(), candidate); err == nil || assistantSafeError(err).Code != data.ToolErrorInvalidCursor {
		t.Fatal("unsigned continuation accepted", err)
	}
}
