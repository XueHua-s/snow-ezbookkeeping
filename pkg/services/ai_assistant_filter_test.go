package services

import (
	"encoding/json"
	"math"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
)

func assistantParserForTest() *AssistantQueryService {
	return &AssistantQueryService{uid: 7, uidString: "7", location: time.UTC, now: func() time.Time { return time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC) }, cursorKey: assistantCursorKey([]byte("isolated-test-secret"), assistantCursorPurpose)}
}
func TestAssistantStrictArgumentMatrix(t *testing.T) {
	tests := []struct {
		name  string
		tool  data.ToolName
		raw   string
		valid bool
	}{
		{"defaults", data.ToolQueryTransactions, `{}`, true},
		{"null defaults", data.ToolQueryTransactions, `{"filter":null,"order_by":null,"page_size":null,"cursor":null}`, true},
		{"array empty account", data.ToolQueryTransactions, `{"filter":{"account_ids":[]}}`, false},
		{"array empty category", data.ToolQueryTransactions, `{"filter":{"category_ids":[]}}`, false},
		{"array empty tags", data.ToolQueryTransactions, `{"filter":{"tag_ids":[]}}`, false},
		{"array empty type", data.ToolQueryTransactions, `{"filter":{"types":[]}}`, false},
		{"arrays null", data.ToolQueryTransactions, `{"filter":{"account_ids":null,"category_ids":null,"tag_ids":null,"types":null}}`, true},
		{"any", data.ToolQueryTransactions, `{"filter":{"tag_ids":["2","1","2"],"tag_match":"any"}}`, true},
		{"all", data.ToolQueryTransactions, `{"filter":{"tag_ids":["2","1"],"tag_match":"all"}}`, true},
		{"no tags", data.ToolQueryTransactions, `{"filter":{"no_tags":true}}`, true},
		{"tags conflict", data.ToolQueryTransactions, `{"filter":{"no_tags":true,"tag_ids":["1"]}}`, false},
		{"isolated tag match", data.ToolQueryTransactions, `{"filter":{"tag_match":"any"}}`, false},
		{"isolated tag match null", data.ToolQueryTransactions, `{"filter":{"tag_match":null}}`, true},
		{"unknown root", data.ToolQueryTransactions, `{"uid":"8"}`, false},
		{"unknown nested", data.ToolQueryTransactions, `{"filter":{"sql":"DROP TABLE transaction"}}`, false},
		{"duplicate root", data.ToolQueryTransactions, `{"filter":null,"filter":{}}`, false},
		{"duplicate nested escaped", data.ToolQueryTransactions, `{"filter":{"keyword":"a","key\u0077ord":"b"}}`, false},
		{"duplicate nested null", data.ToolAggregateTransactions, `{"comparison":{"current":null,"current":"a","baseline":"b"}}`, false},
		{"two roots", data.ToolQueryTransactions, `{} {}`, false},
		{"nonobject", data.ToolQueryTransactions, `[]`, false},
		{"null root", data.ToolQueryTransactions, `null`, false},
		{"wrong string", data.ToolQueryTransactions, `{"filter":{"keyword":42}}`, false},
		{"invalid enum", data.ToolQueryTransactions, `{"filter":{"types":["withdrawal"]}}`, false},
		{"null array element", data.ToolQueryTransactions, `{"filter":{"account_ids":[null]}}`, false},
		{"numeric id", data.ToolQueryTransactions, `{"filter":{"account_ids":[123]}}`, false},
		{"empty id", data.ToolQueryTransactions, `{"filter":{"account_ids":[""]}}`, false},
		{"zero id", data.ToolQueryTransactions, `{"filter":{"account_ids":["0"]}}`, false},
		{"id overflow", data.ToolQueryTransactions, `{"filter":{"account_ids":["9223372036854775808"]}}`, false},
		{"negative id", data.ToolQueryTransactions, `{"filter":{"account_ids":["-1"]}}`, false},
		{"page zero", data.ToolQueryTransactions, `{"page_size":0}`, false},
		{"page negative", data.ToolQueryTransactions, `{"page_size":-1}`, false},
		{"page over", data.ToolQueryTransactions, `{"page_size":201}`, false},
		{"page floating", data.ToolQueryTransactions, `{"page_size":1.0}`, false},
		{"page scientific", data.ToolQueryTransactions, `{"page_size":1e2}`, false},
		{"page overflow", data.ToolQueryTransactions, `{"page_size":9223372036854775808}`, false},
		{"page 200", data.ToolQueryTransactions, `{"page_size":200}`, true},
		{"empty required id", data.ToolGetTransaction, `{"transaction_id":""}`, false},
		{"missing required id", data.ToolGetTransaction, `{}`, false},
		{"null required id", data.ToolGetTransaction, `{"transaction_id":null}`, false},
		{"large string id", data.ToolGetTransaction, `{"transaction_id":"9007199254740993"}`, true},
		{"fraction time", data.ToolQueryTransactions, `{"filter":{"start_time":"2026-01-01T00:00:00.001Z"}}`, false},
		{"zero fraction time", data.ToolQueryTransactions, `{"filter":{"start_time":"2026-01-01T00:00:00.000Z"}}`, false},
		{"time overflow", data.ToolQueryTransactions, `{"filter":{"start_time":"999999999999-01-01T00:00:00Z"}}`, false},
		{"time inverted", data.ToolQueryTransactions, `{"filter":{"start_time":"2026-01-01T00:00:00Z","end_time":"2026-01-01T00:00:00Z"}}`, false},
		{"time epoch", data.ToolQueryTransactions, `{"filter":{"start_time":"1970-01-01T00:00:00Z"}}`, true},
		{"amount no currency", data.ToolQueryTransactions, `{"filter":{"amount_min":"0"}}`, false},
		{"amount negative", data.ToolQueryTransactions, `{"filter":{"currency":"cny","amount_min":"-20","amount_max":"1.2"}}`, true},
		{"amount overflow", data.ToolQueryTransactions, `{"filter":{"currency":"CNY","amount_max":"92233720368547758.08"}}`, false},
		{"amount underflow", data.ToolQueryTransactions, `{"filter":{"currency":"CNY","amount_min":"-92233720368547758.09"}}`, false},
		{"amount inverted", data.ToolQueryTransactions, `{"filter":{"currency":"CNY","amount_min":"5","amount_max":"-3"}}`, false},
		{"amount precision", data.ToolQueryTransactions, `{"filter":{"currency":"CNY","amount_min":"0.001"}}`, false},
		{"amount scientific", data.ToolQueryTransactions, `{"filter":{"currency":"CNY","amount_min":"1e3"}}`, false},
		{"amount NaN", data.ToolQueryTransactions, `{"filter":{"currency":"CNY","amount_min":"NaN"}}`, false},
		{"amount comma", data.ToolQueryTransactions, `{"filter":{"currency":"CNY","amount_min":"1,000"}}`, false},
		{"amount sort multicurrency", data.ToolQueryTransactions, `{"order_by":{"field":"amount","direction":"asc"}}`, false},
		{"amount sort transfer unspecified", data.ToolQueryTransactions, `{"filter":{"currency":"CNY"},"order_by":{"field":"amount","direction":"asc"}}`, false},
		{"amount sort side", data.ToolQueryTransactions, `{"filter":{"currency":"CNY","transfer_side":"source"},"order_by":{"field":"amount","direction":"asc"}}`, true},
		{"amount sort expense", data.ToolQueryTransactions, `{"filter":{"currency":"CNY","types":["expense"]},"order_by":{"field":"amount","direction":"desc"}}`, true},
		{"empty groups", data.ToolAggregateTransactions, `{"group_by":[]}`, true},
		{"duplicate groups", data.ToolAggregateTransactions, `{"group_by":["month","month"]}`, false},
		{"multiple time groups", data.ToolAggregateTransactions, `{"group_by":["month","day"]}`, false},
		{"unsupported week", data.ToolAggregateTransactions, `{"group_by":["week"]}`, false},
		{"empty metrics", data.ToolAggregateTransactions, `{"metrics":[]}`, false},
		{"unselected sort", data.ToolAggregateTransactions, `{"metrics":["count"],"order_by":{"field":"sum","direction":"asc"}}`, false},
		{"selected sort", data.ToolAggregateTransactions, `{"metrics":["count"],"order_by":{"field":"count","direction":"desc"}}`, true},
		{"empty periods", data.ToolAggregateTransactions, `{"periods":[]}`, false},
		{"missing period bound", data.ToolAggregateTransactions, `{"periods":[{"name":"p","start_time":"2026-01-01T00:00:00Z"}]}`, false},
		{"null period bound", data.ToolAggregateTransactions, `{"periods":[{"name":"p","start_time":null,"end_time":"2026-02-01T00:00:00Z"}]}`, false},
		{"empty period name", data.ToolAggregateTransactions, `{"periods":[{"name":"","start_time":"2026-01-01T00:00:00Z","end_time":"2026-02-01T00:00:00Z"}]}`, false},
		{"periods conflicting filter", data.ToolAggregateTransactions, `{"filter":{"start_time":"2026-01-01T00:00:00Z"},"periods":[{"name":"p","start_time":"2026-01-01T00:00:00Z","end_time":"2026-02-01T00:00:00Z"}]}`, false},
		{"undefined comparison", data.ToolAggregateTransactions, `{"comparison":{"baseline":"a","current":"b"}}`, false},
		{"balance empty ids", data.ToolGetAccountBalances, `{"account_ids":[]}`, false},
		{"literal keywords", data.ToolListTags, `{"keyword":"%_\\' OR 1=1"}`, true},
	}
	s := assistantParserForTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := s.parseArguments(tt.tool, []byte(tt.raw))
			if (err == nil) != tt.valid {
				t.Fatalf("valid=%v err=%v", tt.valid, err)
			}
		})
	}
	p := `{"name":"p","start_time":"2026-01-01T00:00:00Z","end_time":"2026-02-01T00:00:00Z"}`
	if _, _, err := s.parseArguments(data.ToolAggregateTransactions, []byte(`{"periods":[`+p+`,`+p+`]}`)); err == nil {
		t.Fatal("duplicate period names accepted")
	}
}
func TestAssistantNormalizationAndExactAmounts(t *testing.T) {
	s := assistantParserForTest()
	a, _, err := s.parseArguments(data.ToolQueryTransactions, []byte(`{"filter":{"start_time":"2026-01-01T08:00:00+08:00","account_ids":["002","1","2"],"tag_ids":["2","2"],"currency":"cny","amount_min":"-000.0","amount_max":"92233720368547758.07","keyword":"A%_\\"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if *a.Filter.StartTime != "2026-01-01T00:00:00Z" || *a.Filter.StartMillis != 1767225600000 || a.PageSize != 50 || a.OrderBy.Field != "time" || a.OrderBy.Direction != "desc" || strings.Join(a.Filter.AccountIDs, ",") != "1,2" || *a.Filter.AmountMin != "0.00" || *a.Filter.MaxMinor != math.MaxInt64 || a.Filter.TransferSide != "either" || a.Filter.TagMatch != "any" || len(a.Filter.Types) != 4 {
		t.Fatalf("bad normalization: %+v %+v", a, a.Filter)
	}
	n, formatted, err := assistantAmount("-92233720368547758.08")
	if err != nil || n != math.MinInt64 || formatted != "-92233720368547758.08" {
		t.Fatalf("minimum: %d %s %v", n, formatted, err)
	}
	total := new(big.Int).Mul(big.NewInt(math.MaxInt64), big.NewInt(100000))
	if assistantFormatAmount(total) != "9223372036854775807000.00" {
		t.Fatal("exact total lost")
	}
	for _, keyword := range []string{"%", "_", `\`, "AbC"} {
		if !assistantLiteralContains("aBc%_\\", keyword) {
			t.Fatal(keyword)
		}
	}
	if assistantLiteralContains("ordinary", "%") || assistantLiteralContains("ordinary", "_") {
		t.Fatal("wildcard interpreted")
	}
}
func TestAssistantArgumentLimitsAndSchemaSingleSource(t *testing.T) {
	s := assistantParserForTest()
	raw, _ := json.Marshal(map[string]any{"keyword": strings.Repeat("x", 33*1024)})
	if _, _, err := s.parseArguments(data.ToolListAccounts, raw); err == nil {
		t.Fatal("32KiB bypass")
	}
	if _, _, err := s.parseArguments(data.ToolListAccounts, []byte(strings.Repeat(" ", 97*1024)+`{}`)); err == nil {
		t.Fatal("96KiB bypass")
	}
	if _, _, err := s.parseArguments(data.ToolListAccounts, []byte{'{', '"', 'k', 'e', 'y', 'w', 'o', 'r', 'd', '"', ':', '"', 0xff, '"', '}'}); err == nil {
		t.Fatal("invalid UTF8")
	}
	for _, name := range assistantToolNames {
		domain := assistantToolDomain(name)
		schema := domain.schema()
		properties := schema["properties"].(map[string]any)
		if len(properties) != len(domain.Fields) || schema["additionalProperties"] != false {
			t.Fatal(name)
		}
		for key := range properties {
			if domain.Fields[key] == nil {
				t.Fatal("schema drift")
			}
		}
	}
}
