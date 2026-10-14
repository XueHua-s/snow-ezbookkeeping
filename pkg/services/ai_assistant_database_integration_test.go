package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"xorm.io/xorm"
)

// Every driver runs the same real registry and independent financial oracle.
// External databases are created and dropped by the guarded opt-in helper;
// SQLite always lives under t.TempDir. No production handler is substituted.
func TestAssistantDatabaseContracts(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			fixture := func(t *testing.T) (*AssistantQueryService, *xorm.Engine) {
				config := &settings.DatabaseConfig{DatabaseType: kind, DatabasePath: filepath.Join(t.TempDir(), "contracts.db"), MaxOpenConnection: 2, MaxIdleConnection: 2}
				if kind != settings.Sqlite3DbType {
					config = assistantExternalServiceConfig(t, kind)
				}
				return assistantServiceFixtureConfig(t, config)
			}
			t.Run("seven_tools_six_truth_and_balances", func(t *testing.T) {
				s, engine := fixture(t)
				// Put the six-transaction truth in one period; the separate large
				// fixture below retains a future transaction as its high-water.
				if _, err := engine.ID(104).Cols("transaction_time").Update(&models.Transaction{TransactionTime: 1004000}); err != nil {
					t.Fatal(err)
				}
				aggregateTruth(t, s, engine)
				assistantInsertEntity(t, engine, &models.Account{AccountId: 14, Uid: 7, ParentAccountId: 10, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Category: models.ACCOUNT_CATEGORY_CREDIT_CARD, Currency: "CNY", Balance: -30000, Hidden: true})
				if len(s.Definitions()) != 7 {
					t.Fatal("not seven tools")
				}
				for _, def := range s.Definitions() {
					raw := `{}`
					if def.Name == data.ToolGetTransaction {
						raw = `{"transaction_id":"103"}`
					}
					r := assistantCall(s, def.Name, raw)
					if !r.OK || r.QueryID == "" || r.Scope == nil || r.QueriedAt == "" {
						t.Fatalf("%s: %+v", def.Name, r.Error)
					}
				}
				flows := aggregateOK(t, s, `{"filter":{"start_time":"1970-01-01T00:00:00Z","end_time":"1970-02-01T00:00:00Z"}}`)
				if flows.MatchedCount != "6" || aggregateFind(t, aggregateRows(t, flows.Totals), "expense")["sum"] != "110.00" || aggregateFind(t, aggregateRows(t, flows.Totals), "income")["sum"] != "1000.00" {
					t.Fatal("six-record period truth differs", flows)
				}
				_, pages := assistantMetadataWalk(t, s, data.ToolGetAccountBalances, `{"page_size":1}`)
				for _, page := range pages {
					if assistantReadBalanceTotals(t, page)["CNY"] != (assistantBalanceTotal{Currency: "CNY", AssetBalance: "2000.00", LiabilityBalance: "300.00", NetAssets: "1700.00"}) {
						t.Fatal("balance is not income-minus-expense", string(page.Totals))
					}
				}
				for _, tool := range []data.ToolName{data.ToolListAccounts, data.ToolListCategories, data.ToolListTags} {
					assistantMetadataWalk(t, s, tool, `{"page_size":1}`)
				}
				t.Log("six logical records: income=1000.00 expense=110.00 net=890.00; current net assets=1700.00; side counts=7 counts_additive=false")
			})
			t.Run("filters_and_rejection", func(t *testing.T) {
				s, engine := fixture(t)
				for _, tc := range []struct{ raw, count string }{
					{`{}`, "4"}, {`{"filter":{"keyword":"%"}}`, "1"}, {`{"filter":{"keyword":"_"}}`, "1"}, {`{"filter":{"keyword":"\\"}}`, "1"}, {`{"filter":{"keyword":"ABC"}}`, "1"},
					{`{"filter":{"tag_ids":["41","42"],"tag_match":"all"}}`, "1"}, {`{"filter":{"tag_ids":["41","42"],"tag_match":"any"}}`, "3"}, {`{"filter":{"no_tags":true}}`, "1"},
					{`{"filter":{"account_ids":["12"],"currency":"USD","amount_min":"30"}}`, "0"},
				} {
					for _, tool := range []data.ToolName{data.ToolQueryTransactions, data.ToolAggregateTransactions} {
						r := assistantCall(s, tool, tc.raw)
						if !r.OK || r.MatchedCount != tc.count {
							t.Fatalf("%s %s: %+v", tool, tc.raw, r)
						}
					}
				}
				for _, raw := range []string{`{"filter":{"account_ids":["11","21"]}}`, `{"filter":{"account_ids":["13"]}}`, `{"filter":{"category_ids":["33"]}}`, `{"filter":{"tag_ids":["43"]}}`, `{"filter":{"tag_ids":["41","44"]}}`} {
					for _, tool := range []data.ToolName{data.ToolQueryTransactions, data.ToolAggregateTransactions} {
						r := assistantCall(s, tool, raw)
						if r.OK || r.Error.Code != data.ToolErrorNotFoundOrNotAccessible {
							t.Fatal("deleted or foreign selection accepted", r)
						}
					}
				}
				for _, raw := range []string{`{"periods":[]}`, `{"filter":{"no_tags":true,"tag_ids":["41"]}}`, `{"filter":{"tag_match":"all"}}`, `{"periods":[{"name":"x","start_time":"2024-01-01T00:00:00Z","end_time":"2024-02-01T00:00:00Z"}],"comparison":{"baseline":"x","current":"missing"}}`, `{"filter":{"start_time":"2024-01-01T00:00:00Z"},"periods":[{"name":"x","start_time":"2024-01-01T00:00:00Z","end_time":"2024-02-01T00:00:00Z"}]}`} {
					r := assistantCall(s, data.ToolAggregateTransactions, raw)
					if r.OK || r.Error.Code != data.ToolErrorInvalidArgument || r.Metrics.RowsRead != 0 {
						t.Fatal("invalid combination reached DB", r)
					}
				}
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				r := s.Execute(core.NewAssistantContext(ctx, "canceled-matrix", "en", ""), data.ToolCall{ID: "cancel", Name: data.ToolAggregateTransactions, Arguments: []byte(`{}`)})
				if r.OK || r.Error.Code != data.ToolErrorCanceled || len(r.Totals) != 0 {
					t.Fatal("canceled query produced totals", r)
				}
				// The same pool remains usable and a corrupt required association
				// must fail explicitly instead of vanishing in an inner join.
				aggregateInsert(t, engine, []models.Transaction{{TransactionId: 200, Uid: 7, AccountId: 21, Type: models.TRANSACTION_DB_TYPE_EXPENSE, TransactionTime: 2000000}})
				r = assistantCall(s, data.ToolAggregateTransactions, `{}`)
				if r.OK || r.Error.Code != data.ToolErrorDataInconsistent {
					t.Fatal("broken association omitted", r)
				}
			})
			t.Run("2500_history_and_deep_ranking", func(t *testing.T) {
				s, engine := fixture(t)
				start := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
				want := []string{"100", "101", "102"}
				// An independent arithmetic oracle, not the production accumulator.
				for offset := 0; offset < 2500; offset += 100 {
					batch := make([]models.Transaction, 100)
					for j := range batch {
						i := offset + j
						batch[j] = models.Transaction{TransactionId: int64(1000 + i), Uid: 7, AccountId: 11, CategoryId: 31, Type: models.TRANSACTION_DB_TYPE_EXPENSE, Amount: int64(i), TransactionTime: start.AddDate(0, 0, i).UnixMilli()}
						want = append(want, strconv.Itoa(1000+i))
					}
					aggregateInsert(t, engine, batch)
				}
				want = append(want, "104")
				r := assistantCall(s, data.ToolQueryTransactions, `{"page_size":200,"order_by":{"field":"time","direction":"asc"}}`)
				got := []string{}
				for page := 0; ; page++ {
					if page > 30 || r.MatchedCount != "2504" {
						t.Fatal("incomplete or nonterminating pages", r.MatchedCount)
					}
					got = append(got, assistantTransactionIDs(assistantTransactionRows(t, r))...)
					if !r.HasMore {
						break
					}
					r = assistantTransactionNext(s, r.NextCursor)
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("history omitted/repeated: got=%d want=%d", len(got), len(want))
				}
				full := aggregateOK(t, s, `{"filter":{"start_time":"2010-01-01T00:00:00Z","end_time":"2017-01-01T00:00:00Z"}}`)
				if full.MatchedCount != "2500" || aggregateFind(t, aggregateRows(t, full.Totals), "expense")["sum"] != "31237.50" {
					t.Fatal("2500 arithmetic sum (2499*2500/2 minor units)", full)
				}
				// Give 2500 separate groups the same exact metric. Deep pages must
				// use a stable complete key even when every ranking metric ties.
				for _, direction := range []string{"asc", "desc"} {
					r := aggregateOK(t, s, `{"filter":{"types":["expense"],"start_time":"2010-01-01T00:00:00Z"},"group_by":["day"],"order_by":{"field":"count","direction":"`+direction+`"},"page_size":200}`)
					seen := make(map[string]bool)
					for page := 0; ; page++ {
						if page > 30 || r.MatchedCount != "2500" || string(r.Totals) != string(full.Totals) {
							t.Fatal("deep ranking totals changed")
						}
						for _, row := range aggregateRows(t, r.Data) {
							bucket := row["time_bucket"].(map[string]any)
							key, _ := json.Marshal(bucket["start_time"])
							if seen[string(key)] || row["count"] != "1" {
								t.Fatal("duplicate day or wrong count", row)
							}
							seen[string(key)] = true
							at, err := time.Parse(time.RFC3339, bucket["start_time"].(string))
							if err != nil {
								t.Fatal(err)
							}
							i := int64(at.Sub(start) / (24 * time.Hour))
							if i < 0 || i >= 2500 || row["sum"] != fmt.Sprintf("%d.%02d", i/100, i%100) {
								t.Fatal("independent bucket sum", row)
							}
						}
						if !r.HasMore {
							break
						}
						r = aggregateOK(t, s, string(assistantJSON(map[string]string{"cursor": r.NextCursor})))
					}
					if len(seen) != 2500 {
						t.Fatal("missing ranked day", len(seen))
					}
				}
				t.Log("2500 daily rows over 6.84 years + original/future rows: all 2504 logical details and 2500 tied day groups reached; exact sum=31237.50")
			})
		})
	}
}

// Slow, opt-in measurement. Plans describe storage batch access, while tool
// metrics describe all application-consumed rows over the entire ledger.
func TestAssistantDatabase100KPlans(t *testing.T) {
	if os.Getenv("EBK_ASSISTANT_PERF_TEST") != "1" {
		t.Skip("100K database measurements require EBK_ASSISTANT_PERF_TEST=1")
	}
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			config := &settings.DatabaseConfig{DatabaseType: kind, DatabasePath: filepath.Join(t.TempDir(), "plans.db"), MaxOpenConnection: 2, MaxIdleConnection: 2}
			if kind != settings.Sqlite3DbType {
				config = assistantExternalServiceConfig(t, kind)
			}
			s := assistantAggregateBenchmarkFixture(t, config)
			c := core.NewAssistantContext(context.Background(), "100k-plan-fixture", "en", "")
			// Refresh optimizer statistics only in this freshly-created fixture.
			session := s.database.NewSession(c)
			engine := session.Engine()
			session.Close()
			analyze := `ANALYZE "transaction"`
			if kind == settings.MySqlDbType {
				analyze = "ANALYZE TABLE `transaction`"
			}
			if _, err := engine.Exec(analyze); err != nil {
				t.Fatal(err)
			}
			err := s.database.WithAssistantReadView(c, func(view datastore.AssistantReadView) error {
				d := view.Dialect()
				versionQuery := "SELECT version()"
				if kind == settings.Sqlite3DbType {
					versionQuery = "SELECT sqlite_version()"
				}
				var version string
				if err := view.QueryRowContext(c, versionQuery).Scan(&version); err != nil {
					return err
				}
				t.Logf("database=%s version=%s fixture=100000", kind, version)
				columns := strings.Split(assistantTransactionColumns, ",")
				for i := range columns {
					columns[i] = d.QuoteIdentifier(columns[i])
				}
				prefix := "EXPLAIN ANALYZE "
				if kind == settings.Sqlite3DbType {
					prefix = "EXPLAIN QUERY PLAN "
				}
				if kind == settings.PostgresDbType {
					prefix = "EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) "
				}
				base := "SELECT " + strings.Join(columns, ",") + " FROM " + d.TableName(&models.Transaction{}) + " WHERE uid = ? AND deleted = ? AND "
				for _, sample := range []struct {
					name, suffix string
					args         []any
				}{
					{"first_batch", "transaction_id > ? ORDER BY transaction_id ASC LIMIT ?", []any{int64(7), false, int64(0), 256}},
					{"deep_batch", "transaction_id > ? ORDER BY transaction_id ASC LIMIT ?", []any{int64(7), false, int64(90000), 256}},
					{"high_water", "type <> ? ORDER BY transaction_time DESC, transaction_id DESC LIMIT 1", []any{int64(7), false, models.TRANSACTION_DB_TYPE_TRANSFER_IN}},
				} {
					rows, err := view.QueryContext(c, prefix+base+sample.suffix, sample.args...)
					if err != nil {
						return err
					}
					names, err := rows.Columns()
					if err != nil {
						rows.Close()
						return err
					}
					for rows.Next() {
						values := make([]any, len(names))
						targets := make([]any, len(names))
						for i := range values {
							targets[i] = &values[i]
						}
						if err := rows.Scan(targets...); err != nil {
							rows.Close()
							return err
						}
						for i, v := range values {
							if b, ok := v.([]byte); ok {
								values[i] = string(b)
							}
						}
						t.Logf("plan %s: %v", sample.name, values)
					}
					if err := rows.Err(); err != nil {
						rows.Close()
						return err
					}
					rows.Close()
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			for _, sample := range []struct {
				name string
				tool data.ToolName
				args string
			}{
				{"all_history", data.ToolAggregateTransactions, `{}`},
				{"month_category", data.ToolAggregateTransactions, `{"group_by":["month","category"]}`},
				{"detail_first", data.ToolQueryTransactions, `{}`},
			} {
				start := time.Now()
				r := assistantCall(s, sample.tool, sample.args)
				if !r.OK || r.MatchedCount != "100000" || r.Metrics.RowsRead < 100000 {
					t.Fatal("100K incomplete", r.Error, r.MatchedCount)
				}
				if sample.tool == data.ToolAggregateTransactions && aggregateFind(t, aggregateRows(t, r.Totals), "expense")["sum"] != "-500.00" {
					t.Fatal("100K independent sum mismatch")
				}
				t.Logf("measurement %s elapsed=%s application_rows=%d tool_duration=%s result_records=%d returned_bytes=%d", sample.name, time.Since(start), r.Metrics.RowsRead, r.Metrics.Duration, r.Metrics.ResultRecords, r.Metrics.SerializedBytes)
			}
		})
	}
}
