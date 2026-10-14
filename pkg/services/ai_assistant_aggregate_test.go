package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"xorm.io/xorm"
)

func aggregateRows(t testing.TB, raw json.RawMessage) []map[string]any {
	t.Helper()
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err, string(raw))
	}
	return rows
}
func aggregateOK(t testing.TB, s *AssistantQueryService, raw string) data.ToolResult {
	t.Helper()
	result := assistantCall(s, data.ToolAggregateTransactions, raw)
	if !result.OK {
		t.Fatalf("aggregate failed: %+v", result.Error)
	}
	if result.Coverage != data.QueryCoveragePage || result.TotalsCoverage != data.QueryCoverageAllMatching {
		t.Fatal("false coverage", result)
	}
	return result
}
func aggregateFind(t testing.TB, rows []map[string]any, flow string) map[string]any {
	t.Helper()
	for _, r := range rows {
		if r["flow"] == flow {
			return r
		}
	}
	t.Fatal("missing flow", flow, rows)
	return nil
}
func aggregateInsert(t testing.TB, engine *xorm.Engine, transactions []models.Transaction) {
	t.Helper()
	for start := 0; start < len(transactions); start += 20 {
		end := start + 20
		if end > len(transactions) {
			end = len(transactions)
		}
		if _, err := engine.Insert(transactions[start:end]); err != nil {
			t.Fatal(err)
		}
	}
}
func aggregateTruth(t *testing.T, s *AssistantQueryService, engine *xorm.Engine) {
	t.Helper()
	aggregateInsert(t, engine, []models.Transaction{
		{TransactionId: 107, Uid: 7, Type: models.TRANSACTION_DB_TYPE_INCOME, AccountId: 11, CategoryId: 31, TransactionTime: 1005000, Amount: 100000},
		{TransactionId: 108, Uid: 7, Type: models.TRANSACTION_DB_TYPE_EXPENSE, AccountId: 11, CategoryId: 31, TransactionTime: 1006000, Amount: 3000},
	})
	r := aggregateOK(t, s, `{"metrics":["count","sum","min","max"],"group_by":["account"]}`)
	if r.MatchedCount != "6" || r.CountsAdditive == nil || *r.CountsAdditive {
		t.Fatal("logical count or additivity", r)
	}
	rows := aggregateRows(t, r.Totals)
	for flow, want := range map[string]string{"expense": "110.00", "income": "1000.00", "transfer_out": "40.00", "transfer_in": "5.00", "balance_modification": "200.00"} {
		row := aggregateFind(t, rows, flow)
		if row["sum"] != want {
			t.Fatal(flow, row)
		}
	}
	expense := aggregateFind(t, rows, "expense")
	if expense["count"] != "3" || expense["min"] != "-20.00" || expense["max"] != "100.00" {
		t.Fatal(expense)
	}
	incomeMinor := big.NewInt(100000)
	if assistantFormatAmount(incomeMinor.Sub(incomeMinor, big.NewInt(11000))) != "890.00" {
		t.Fatal("independent net")
	}
	count := 0
	for _, row := range aggregateRows(t, r.Data) {
		n, _ := strconv.Atoi(row["count"].(string))
		count += n
	}
	if count != 7 {
		t.Fatal("side counts", count)
	}
	dest := aggregateOK(t, s, `{"filter":{"account_ids":["12"]}}`)
	if dest.MatchedCount != "1" || len(aggregateRows(t, dest.Totals)) != 1 || aggregateFind(t, aggregateRows(t, dest.Totals), "transfer_in")["sum"] != "5.00" {
		t.Fatal(dest)
	}
}
func TestAssistantAggregateSixTransactionTruth(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	aggregateTruth(t, s, engine)
}
func TestAssistantAggregateExternalTruth(t *testing.T) {
	for _, kind := range []string{settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			s, engine := assistantServiceFixtureConfig(t, assistantExternalServiceConfig(t, kind))
			aggregateTruth(t, s, engine)
		})
	}
}
func TestAssistantAggregateSharedFiltersAndIsolation(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	for _, tt := range []struct{ raw, count string }{
		{`{}`, "4"}, {`{"filter":{"tag_ids":["41","42"],"tag_match":"any"}}`, "3"}, {`{"filter":{"tag_ids":["41","42"],"tag_match":"all"}}`, "1"}, {`{"filter":{"no_tags":true}}`, "1"}, {`{"filter":{"account_ids":["10","11"]}}`, "4"}, {`{"filter":{"currency":"USD","amount_min":"30","account_ids":["12"]}}`, "0"},
	} {
		r := aggregateOK(t, s, tt.raw)
		if r.MatchedCount != tt.count {
			t.Fatal(tt, r)
		}
	}
	for _, raw := range []string{`{"filter":{"account_ids":["11","21"]}}`, `{"filter":{"category_ids":["31","33"]}}`, `{"filter":{"tag_ids":["41","44"]}}`, `{"filter":{"tag_ids":["43"]}}`} {
		r := assistantCall(s, data.ToolAggregateTransactions, raw)
		if r.OK || r.Error.Code != data.ToolErrorNotFoundOrNotAccessible {
			t.Fatal(raw, r)
		}
	}
	for _, raw := range []string{`{"periods":[]}`, `{"filter":{"no_tags":true,"tag_ids":["41"]}}`, `{"filter":{"tag_match":"all"}}`, `{"metrics":["count"],"order_by":{"field":"sum","direction":"asc"}}`, `{"periods":[{"name":"x","start_time":"2024-01-01T00:00:00Z","end_time":"2024-02-01T00:00:00Z"}],"comparison":{"baseline":"x","current":"missing"}}`, `{"filter":{"start_time":"2024-01-01T00:00:00Z"},"periods":[{"name":"x","start_time":"2024-01-01T00:00:00Z","end_time":"2024-02-01T00:00:00Z"}]}`} {
		r := assistantCall(s, data.ToolAggregateTransactions, raw)
		if r.OK || r.Error.Code != data.ToolErrorInvalidArgument || r.Metrics.RowsRead != 0 {
			t.Fatal(raw, r)
		}
	}
	// Corrupt same-user transaction referencing another user's required account.
	aggregateInsert(t, engine, []models.Transaction{{TransactionId: 120, Uid: 7, AccountId: 21, Type: models.TRANSACTION_DB_TYPE_EXPENSE, Amount: 1}})
	r := assistantCall(s, data.ToolAggregateTransactions, `{}`)
	if r.OK || r.Error.Code != data.ToolErrorDataInconsistent {
		t.Fatal(r)
	}
}
func TestAssistantAggregateFull2500AndOverflow(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	rows := make([]models.Transaction, 2500)
	for i := range rows {
		rows[i] = models.Transaction{TransactionId: int64(1000 + i), Uid: 7, AccountId: 11, CategoryId: 31, Type: models.TRANSACTION_DB_TYPE_EXPENSE, Amount: math.MaxInt64, TransactionTime: int64(i) * 86400000}
	}
	aggregateInsert(t, engine, rows)
	r := aggregateOK(t, s, `{"filter":{"types":["expense"]},"metrics":["count","sum","min","max"]}`)
	want := new(big.Int).Mul(big.NewInt(math.MaxInt64), big.NewInt(2500))
	want.Add(want, big.NewInt(8000))
	total := aggregateFind(t, aggregateRows(t, r.Totals), "expense")
	if r.MatchedCount != "2502" || total["count"] != "2502" || total["sum"] != assistantFormatAmount(want) || total["max"] != "92233720368547758.07" || total["min"] != "-20.00" {
		t.Fatal(total, r.MatchedCount)
	}
	if r.Metrics.RowsRead < 2502 {
		t.Fatal("did not consume all rows", r.Metrics)
	}
}
func TestAssistantAggregatePeriodsAndOrdinalComparison(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	ms := func(v string) int64 { x, _ := time.Parse(time.RFC3339, v); return x.UnixMilli() }
	aggregateInsert(t, engine, []models.Transaction{
		{TransactionId: 110, Uid: 7, AccountId: 11, CategoryId: 31, Type: models.TRANSACTION_DB_TYPE_EXPENSE, Amount: 300, TransactionTime: ms("2024-02-28T12:00:00Z")},
		{TransactionId: 111, Uid: 7, AccountId: 11, CategoryId: 31, Type: models.TRANSACTION_DB_TYPE_EXPENSE, Amount: 400, TransactionTime: ms("2024-02-29T12:00:00Z")},
		{TransactionId: 112, Uid: 7, AccountId: 11, CategoryId: 31, Type: models.TRANSACTION_DB_TYPE_EXPENSE, Amount: 999, TransactionTime: ms("2024-03-01T00:00:00Z")},
	})
	raw := `{"group_by":["day","category"],"metrics":["sum","count","min","max"],"periods":[{"name":"baseline","start_time":"2024-02-27T00:00:00Z","end_time":"2024-03-01T00:00:00Z"},{"name":"current","start_time":"2024-02-28T00:00:00Z","end_time":"2024-03-01T00:00:00Z"}],"comparison":{"baseline":"baseline","current":"current"}}`
	r := aggregateOK(t, s, raw)
	if r.MatchedCount != "2" || *r.CountsAdditive || r.Periods[0].MatchedCount != "2" || r.Periods[1].MatchedCount != "2" {
		t.Fatal(r)
	}
	rows := aggregateRows(t, r.Data)
	if len(rows) != 5 {
		t.Fatal(rows)
	}
	zero := rows[0]
	if zero["sum"] != "0.00" || zero["count"] != "0" || zero["min"] != nil || zero["max"] != nil {
		t.Fatal(zero)
	}
	comps := aggregateRows(t, r.Comparison)
	if len(comps) != 5 {
		t.Fatal(comps)
	}
	sum := comps[0]["metrics"].(map[string]any)["sum"].(map[string]any)
	if sum["baseline"] != "0.00" || sum["current"] != "3.00" || sum["percentage"] != nil || sum["reason"] != "non_positive_baseline" {
		t.Fatal(sum)
	}
	sum = comps[1]["metrics"].(map[string]any)["sum"].(map[string]any)
	if sum["percentage"] != "33.33" || sum["delta"] != "1.00" {
		t.Fatal(sum)
	}
	if comps[2]["paired"] != false || comps[2]["metrics"].(map[string]any)["sum"].(map[string]any)["reason"] != "unmatched_time_bucket" {
		t.Fatal(comps[2])
	}
}
func TestAssistantAggregateMissingCategoryComparison(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	ms := func(year int) int64 { return time.Date(year, 1, 2, 0, 0, 0, 0, time.UTC).UnixMilli() }
	aggregateInsert(t, engine, []models.Transaction{{TransactionId: 110, Uid: 7, AccountId: 11, CategoryId: 31, Type: models.TRANSACTION_DB_TYPE_EXPENSE, Amount: 300, TransactionTime: ms(2024)}, {TransactionId: 111, Uid: 7, AccountId: 11, CategoryId: 32, Type: models.TRANSACTION_DB_TYPE_EXPENSE, Amount: 400, TransactionTime: ms(2025)}})
	r := aggregateOK(t, s, `{"group_by":["category"],"periods":[{"name":"base","start_time":"2024-01-01T00:00:00Z","end_time":"2024-02-01T00:00:00Z"},{"name":"now","start_time":"2025-01-01T00:00:00Z","end_time":"2025-02-01T00:00:00Z"}],"comparison":{"baseline":"base","current":"now"}}`)
	rows := aggregateRows(t, r.Data)
	if len(rows) != 4 {
		t.Fatal(rows)
	}
	zeros := 0
	for _, row := range rows {
		if row["count"] == "0" {
			zeros++
		}
	}
	if zeros != 2 {
		t.Fatal(rows)
	}
}
func TestAssistantAggregateCalendarAndEmptyWidePages(t *testing.T) {
	s, _ := assistantServiceFixture(t)
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	s.location = loc
	r := aggregateOK(t, s, `{"filter":{"currency":"CNY","types":["expense"],"start_time":"2024-03-09T00:00:00-05:00","end_time":"2024-03-12T00:00:00-04:00"},"group_by":["day"],"metrics":["count","sum","min","max"]}`)
	rows := aggregateRows(t, r.Data)
	if len(rows) != 3 {
		t.Fatal(rows)
	}
	bucket := rows[1]["time_bucket"].(map[string]any)
	if bucket["start_time"] != "2024-03-10T00:00:00-05:00" || bucket["end_time"] != "2024-03-11T00:00:00-04:00" {
		t.Fatal(bucket)
	}
	r = aggregateOK(t, s, `{"filter":{"currency":"CNY","types":["expense"],"start_time":"2024-01-31T13:00:00-05:00","end_time":"2024-03-01T00:00:00-05:00"},"group_by":["month"]}`)
	rows = aggregateRows(t, r.Data)
	if len(rows) != 2 || rows[0]["time_bucket"].(map[string]any)["start_time"] != "2024-01-31T13:00:00-05:00" || rows[1]["time_bucket"].(map[string]any)["end_time"] != "2024-03-01T00:00:00-05:00" {
		t.Fatal(rows)
	}
	r = aggregateOK(t, s, `{"filter":{"types":["income"],"currency":"CNY","start_time":"0001-01-01T00:00:00Z","end_time":"9999-01-01T00:00:00Z"},"group_by":["day"],"page_size":2}`)
	if r.ReturnedCount != 2 || !r.HasMore || r.MatchedCount != "0" {
		t.Fatal(r)
	}
	for i := 0; i < 3; i++ {
		previous := string(r.Data)
		r = aggregateOK(t, s, string(assistantJSON(map[string]any{"cursor": r.NextCursor})))
		if !r.HasMore || string(r.Data) == previous {
			t.Fatal("virtual pagination stuck", r)
		}
	}
	empty := aggregateOK(t, s, `{"filter":{"types":["income"]}}`)
	if string(empty.Totals) != "[]" || string(empty.Data) != "[]" {
		t.Fatal(empty)
	}
}
func TestAssistantAggregateRatios(t *testing.T) {
	for _, tt := range []struct {
		delta, base  int64
		want, reason string
	}{{1, 6, "16.67", ""}, {-1, 6, "-16.67", ""}, {1, 32, "3.13", ""}, {-1, 32, "-3.13", ""}, {1, 0, "", "non_positive_baseline"}, {1, -3, "", "non_positive_baseline"}, {0, 4, "0.00", ""}} {
		value, reason := assistantAggregatePercentage(big.NewInt(tt.delta), big.NewInt(tt.base))
		if reason != tt.reason || value != nil && *value != tt.want || value == nil && tt.want != "" {
			t.Fatal(tt, value, reason)
		}
	}
}
func TestAssistantAggregateStableMetricPagesAndFreshTotals(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	for i := 0; i < 12; i++ {
		account := models.Account{AccountId: int64(200 + i), Uid: 7, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Name: "same name", Currency: "CNY"}
		if i >= 6 {
			account.Currency = "USD"
		}
		if _, err := engine.Insert(&account); err != nil {
			t.Fatal(err)
		}
		aggregateInsert(t, engine, []models.Transaction{{TransactionId: int64(300 + i), Uid: 7, AccountId: account.AccountId, Type: models.TRANSACTION_DB_TYPE_INCOME, Amount: int64(i%3-1) * 100, TransactionTime: int64(2000000 + i)}})
	}
	for _, metric := range []string{"group_key", "sum", "count", "min", "max"} {
		for _, direction := range []string{"asc", "desc"} {
			t.Run(metric+direction, func(t *testing.T) {
				args := map[string]any{"filter": map[string]any{"types": []string{"income"}}, "group_by": []string{"account"}, "metrics": []string{"sum", "count", "min", "max"}, "page_size": 200, "order_by": map[string]any{"field": metric, "direction": direction}}
				full := aggregateOK(t, s, string(assistantJSON(args)))
				want := aggregateRows(t, full.Data)
				args["page_size"] = 2
				r := aggregateOK(t, s, string(assistantJSON(args)))
				got := []map[string]any{}
				seen := map[string]bool{}
				for {
					if string(r.Totals) != string(full.Totals) || r.MatchedCount != "12" {
						t.Fatal("page changed totals", r)
					}
					for _, row := range aggregateRows(t, r.Data) {
						id := row["account_id"].(string)
						if seen[id] {
							t.Fatal("duplicate", id)
						}
						seen[id] = true
						got = append(got, row)
					}
					if !r.HasMore {
						break
					}
					r = aggregateOK(t, s, string(assistantJSON(map[string]any{"cursor": r.NextCursor})))
				}
				if string(assistantJSON(got)) != string(assistantJSON(want)) || len(got) != 12 {
					t.Fatal("deep page ordering", got, want)
				}
				// Currency order never intermixes amounts, including descending metric order.
				for i, row := range got {
					if (i < 6) != (row["currency"] == "CNY") {
						t.Fatal("currency rankings interleaved", got)
					}
				}
			})
		}
	}
	first := aggregateOK(t, s, `{"group_by":["account"],"page_size":1}`)
	aggregateInsert(t, engine, []models.Transaction{{TransactionId: 500, Uid: 7, AccountId: 11, Type: models.TRANSACTION_DB_TYPE_INCOME, Amount: 123, TransactionTime: 3000000}})
	s.now = func() time.Time { return time.Date(2026, 10, 14, 1, 0, 0, 0, time.UTC) }
	next := aggregateOK(t, s, string(assistantJSON(map[string]any{"cursor": first.NextCursor})))
	if next.MatchedCount != "17" || next.QueriedAt == first.QueriedAt || string(next.Totals) == string(first.Totals) {
		t.Fatal("cursor falsely froze totals", next)
	}
}
func TestAssistantAggregateByteBudgetAndCancellation(t *testing.T) {
	s, _ := assistantServiceFixture(t)
	s.maxResultBytes = 3000
	first := aggregateOK(t, s, `{"group_by":["month"],"filter":{"start_time":"1970-01-01T00:00:00Z","end_time":"2101-01-01T00:00:00Z"},"page_size":200}`)
	if first.ReturnedCount >= 200 || !first.HasMore {
		t.Fatal("page not shortened", first)
	}
	next := aggregateOK(t, s, string(assistantJSON(map[string]any{"cursor": first.NextCursor})))
	if string(first.Data) == string(next.Data) || string(first.Totals) != string(next.Totals) {
		t.Fatal("incorrect byte continuation")
	}
	s.maxResultBytes = 1024
	result := assistantCall(s, data.ToolAggregateTransactions, `{}`)
	if result.OK || result.Error.Code != data.ToolErrorResultTooLarge {
		t.Fatal("large totals silently truncated", result)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result = s.Execute(core.NewAssistantContext(ctx, "canceled", "en", ""), data.ToolCall{ID: "x", Name: data.ToolAggregateTransactions, Arguments: []byte(`{}`)})
	if result.OK || result.Error.Code != data.ToolErrorCanceled || len(result.Totals) > 0 {
		t.Fatal("canceled totals", result)
	}
}
func TestAssistantAggregateWideMetricZeroPaging(t *testing.T) {
	s, _ := assistantServiceFixture(t)
	for _, field := range []string{"count", "sum", "min", "max"} {
		for _, direction := range []string{"asc", "desc"} {
			raw := fmt.Sprintf(`{"filter":{"currency":"CNY","types":["expense"],"start_time":"0001-01-01T00:00:00Z","end_time":"9999-01-01T00:00:00Z"},"group_by":["day"],"metrics":["count","sum","min","max"],"order_by":{"field":%q,"direction":%q},"page_size":2}`, field, direction)
			r := aggregateOK(t, s, raw)
			for i := 0; i < 3; i++ {
				old := string(r.Data)
				r = aggregateOK(t, s, string(assistantJSON(map[string]any{"cursor": r.NextCursor})))
				if !r.HasMore || old == string(r.Data) {
					t.Fatal("metric zero pages stuck", field, direction)
				}
			}
		}
	}
}

func TestAssistantAggregateZeroCombinationsAreNotCartesian(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	if _, err := engine.Insert(&models.Account{AccountId: 14, Uid: 7, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Currency: "CNY", Name: "other"}); err != nil {
		t.Fatal(err)
	}
	aggregateInsert(t, engine, []models.Transaction{
		{TransactionId: 110, Uid: 7, AccountId: 11, CategoryId: 31, Type: models.TRANSACTION_DB_TYPE_EXPENSE, Amount: 100, TransactionTime: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC).UnixMilli()},
		{TransactionId: 111, Uid: 7, AccountId: 14, CategoryId: 32, Type: models.TRANSACTION_DB_TYPE_EXPENSE, Amount: 200, TransactionTime: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC).UnixMilli()},
	})
	r := aggregateOK(t, s, `{"group_by":["account","category"],"periods":[{"name":"base","start_time":"2024-01-01T00:00:00Z","end_time":"2024-02-01T00:00:00Z"},{"name":"now","start_time":"2025-01-01T00:00:00Z","end_time":"2025-02-01T00:00:00Z"}],"comparison":{"baseline":"base","current":"now"}}`)
	rows := aggregateRows(t, r.Data)
	if len(rows) != 4 {
		t.Fatal("constructed Cartesian product", rows)
	}
	for _, row := range rows {
		if row["account_id"] == "11" && row["category_id"] != "31" || row["account_id"] == "14" && row["category_id"] != "32" {
			t.Fatal("invented dimension pair", row)
		}
	}
}

func TestAssistantAggregateVirtualDeepSeekAndFiniteEnd(t *testing.T) {
	s, _ := assistantServiceFixture(t)
	first := aggregateOK(t, s, `{"filter":{"types":["income"],"currency":"CNY","start_time":"0001-01-01T00:00:00Z","end_time":"9999-01-01T00:00:00Z"},"group_by":["day"],"page_size":2}`)
	args, cursor, err := s.parseArguments(data.ToolAggregateTransactions, assistantJSON(map[string]any{"cursor": first.NextCursor}))
	if err != nil {
		t.Fatal(err)
	}
	start, end := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)
	a := &assistantAggregation{e: &assistantQueryExecution{Args: args}, dimension: "day", ranges: []assistantAggregateRange{{Start: &start, End: &end}}}
	// A server-signed position near the final page proves a direct seek through
	// 3.65 million virtual days, without building or walking the skipped buckets.
	g := &assistantAggregateGroup{Dimension: assistantAggregateDimension{Currency: "CNY", Flow: "income"}, Ordinal: a.bucketCount(0) - 3}
	cursor.Last.Group = string(assistantJSON(a.position(g)))
	token, err := assistantSignCursor(*cursor, s.cursorKey)
	if err != nil {
		t.Fatal(err)
	}
	last := aggregateOK(t, s, string(assistantJSON(map[string]any{"cursor": token})))
	if last.HasMore || last.ReturnedCount != 2 {
		t.Fatal(last)
	}
	rows := aggregateRows(t, last.Data)
	bucket := rows[1]["time_bucket"].(map[string]any)
	if bucket["end_time"] != "9999-01-01T00:00:00Z" {
		t.Fatal(bucket)
	}
}

func TestAssistantAggregateMidnightDSTAndSkippedDate(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}
	s.location = loc
	stamp, err := time.Parse(time.RFC3339, "2018-11-04T23:30:00-02:00")
	if err != nil {
		t.Fatal(err)
	}
	aggregateInsert(t, engine, []models.Transaction{{TransactionId: 800, Uid: 7, AccountId: 11, Type: models.TRANSACTION_DB_TYPE_INCOME, Amount: 100, TransactionTime: stamp.UnixMilli()}})
	for _, raw := range []string{
		`{"filter":{"types":["income"],"start_time":"2018-11-03T00:00:00-03:00","end_time":"2018-11-06T00:00:00-02:00"},"group_by":["day"]}`,
		`{"filter":{"types":["income"]},"group_by":["day"]}`,
	} {
		r := aggregateOK(t, s, raw)
		found := false
		for _, row := range aggregateRows(t, r.Data) {
			if row["count"] != "1" {
				continue
			}
			found = true
			bucket := row["time_bucket"].(map[string]any)
			if bucket["start_time"] != "2018-11-04T01:00:00-02:00" || bucket["end_time"] != "2018-11-05T00:00:00-02:00" {
				t.Fatal("midnight DST boundary", bucket)
			}
			start, _ := time.Parse(time.RFC3339, bucket["start_time"].(string))
			end, _ := time.Parse(time.RFC3339, bucket["end_time"].(string))
			if stamp.Before(start) || !stamp.Before(end) {
				t.Fatal("transaction outside bucket", bucket)
			}
		}
		if !found {
			t.Fatal("transaction missing", r)
		}
	}
	loc, err = time.LoadLocation("Pacific/Apia")
	if err != nil {
		t.Fatal(err)
	}
	s.location = loc
	for _, direction := range []string{"asc", "desc"} {
		r := aggregateOK(t, s, fmt.Sprintf(`{"filter":{"currency":"CNY","types":["income"],"start_time":"2011-12-29T00:00:00-10:00","end_time":"2012-01-01T00:00:00+14:00"},"group_by":["day"],"order_by":{"field":"group_key","direction":%q},"page_size":1}`, direction))
		all := aggregateRows(t, r.Data)
		for r.HasMore {
			r = aggregateOK(t, s, string(assistantJSON(map[string]any{"cursor": r.NextCursor})))
			all = append(all, aggregateRows(t, r.Data)...)
		}
		if len(all) != 2 {
			t.Fatal("nonexistent civil date emitted", all)
		}
		for _, row := range all {
			bucket := row["time_bucket"].(map[string]any)
			start, _ := time.Parse(time.RFC3339, bucket["start_time"].(string))
			end, _ := time.Parse(time.RFC3339, bucket["end_time"].(string))
			if !start.Before(end) || start.Day() == 30 {
				t.Fatal("invalid skipped-date bucket", bucket)
			}
		}
	}
}

func TestAssistantAggregateHistoricalOffsetPreservesInstant(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		t.Fatal(err)
	}
	instant := assistantCalendarBoundary(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC), loc)
	encoded := assistantAggregateTime(instant)
	decoded, err := time.Parse(time.RFC3339, encoded)
	if err != nil || !decoded.Equal(instant) {
		t.Fatal("historical second offset lost precision", instant, encoded, err)
	}
}
