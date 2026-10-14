package services

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

func assistantAggregateBenchmarkFixture(b testing.TB, config *settings.DatabaseConfig) *AssistantQueryService {
	b.Helper()
	saved := *datastore.Container
	err := datastore.InitializeDataStore(&settings.Config{DatabaseConfig: config})
	if err != nil {
		*datastore.Container = saved
		b.Fatal(err)
	}
	store := datastore.Container.UserDataStore
	*datastore.Container = saved
	session := store.Choose(7).NewSession(core.NewNullContext())
	engine := session.Engine()
	session.Close()
	b.Cleanup(func() { engine.Close() })
	if err = store.SyncStructs(&models.Account{}, &models.TransactionCategory{}, &models.TransactionTag{}, &models.TransactionTagGroup{}, &models.Transaction{}, &models.TransactionTagIndex{}, &models.TransactionPictureInfo{}); err != nil {
		b.Fatal(err)
	}
	if _, err = engine.Insert(&models.Account{AccountId: 11, Uid: 7, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Currency: "CNY", Hidden: true}); err != nil {
		b.Fatal(err)
	}
	for i := 0; i < 12; i++ {
		if _, err = engine.Insert(&models.TransactionCategory{CategoryId: int64(30 + i), Uid: 7, Name: "benchmark", Type: models.CATEGORY_TYPE_EXPENSE}); err != nil {
			b.Fatal(err)
		}
	}
	// Keep fixture creation bounded too: exactly 100,000 valid transactions over
	// more than eleven years, all signed amounts independently summing to -500.00.
	start := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	batch := make([]models.Transaction, 100)
	for offset := 0; offset < 100000; offset += len(batch) {
		for j := range batch {
			i := offset + j
			batch[j] = models.Transaction{TransactionId: int64(i + 1), Uid: 7, AccountId: 11, CategoryId: int64(30 + i%12), Type: models.TRANSACTION_DB_TYPE_EXPENSE, Amount: int64(i%100 - 50), TransactionTime: start + int64(i)*3600000}
		}
		aggregateInsert(b, engine, batch)
	}
	s, err := NewAssistantQueryService(store, 7, AssistantQueryOptions{Secret: []byte("benchmark-only"), QueryTimeout: time.Minute})
	if err != nil {
		b.Fatal(err)
	}
	return s
}
func BenchmarkAssistantAggregate100K(b *testing.B) {
	s := assistantAggregateBenchmarkFixture(b, &settings.DatabaseConfig{DatabaseType: settings.Sqlite3DbType, DatabasePath: filepath.Join(b.TempDir(), "aggregate-benchmark.db"), MaxOpenConnection: 2, MaxIdleConnection: 2})
	firstArgs := `{"group_by":["month","category"],"order_by":{"field":"sum","direction":"desc"},"page_size":50}`
	deep := aggregateOK(b, s, firstArgs)
	for i := 0; i < 10; i++ {
		if !deep.HasMore {
			b.Fatal("fixture too small for deep page")
		}
		deep = aggregateOK(b, s, string(assistantJSON(map[string]any{"cursor": deep.NextCursor})))
	}
	deepArgs := string(assistantJSON(map[string]any{"cursor": deep.NextCursor}))
	// A signed deep continuation is prepared outside timing from the known
	// fixture key (50,000 rows from the end), avoiding 1,000 warm-up scans.
	detail := assistantCall(s, data.ToolQueryTransactions, `{}`)
	if !detail.OK || !detail.HasMore {
		b.Fatal("detail fixture failed", detail.Error)
	}
	cursor, err := s.decodeCursor(detail.NextCursor, data.ToolQueryTransactions)
	if err != nil {
		b.Fatal(err)
	}
	cursor.Last = assistantPageKey{ID: "50001", Time: strconv.FormatInt(time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()+50000*3600000, 10)}
	signed, err := assistantSignCursor(*cursor, s.cursorKey)
	if err != nil {
		b.Fatal(err)
	}
	for _, test := range []struct {
		name, args string
		tool       data.ToolName
		firstID    string
	}{
		{"all_history", `{}`, data.ToolAggregateTransactions, ""},
		{"month_category_first", firstArgs, data.ToolAggregateTransactions, ""},
		{"month_category_deep_550", deepArgs, data.ToolAggregateTransactions, ""},
		{"detail_first", `{}`, data.ToolQueryTransactions, "100000"},
		{"detail_deep_50000", string(assistantJSON(map[string]string{"cursor": signed})), data.ToolQueryTransactions, "50000"},
	} {
		b.Run(test.name, func(b *testing.B) {
			runtime.GC()
			var before runtime.MemStats
			runtime.ReadMemStats(&before)
			stop := make(chan struct{})
			var wait sync.WaitGroup
			wait.Add(1)
			peak := before.HeapAlloc
			go func() {
				defer wait.Done()
				ticker := time.NewTicker(5 * time.Millisecond)
				defer ticker.Stop()
				for {
					select {
					case <-stop:
						return
					case <-ticker.C:
						var m runtime.MemStats
						runtime.ReadMemStats(&m)
						if m.HeapAlloc > peak {
							peak = m.HeapAlloc
						}
					}
				}
			}()
			var rows int64
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := assistantCall(s, test.tool, test.args)
				if !result.OK || result.MatchedCount != "100000" {
					b.Fatalf("incomplete aggregate %+v count=%s", result.Error, result.MatchedCount)
				}
				if test.tool == data.ToolAggregateTransactions {
					totals := aggregateRows(b, result.Totals)
					if len(totals) != 1 || totals[0]["sum"] != "-500.00" {
						b.Fatal("wrong independent total", totals)
					}
				} else {
					var rows []assistantTransactionProjection
					if err := json.Unmarshal(result.Data, &rows); err != nil {
						b.Fatal(err)
					}
					if len(rows) != 50 || rows[0].ID != test.firstID || !result.HasMore {
						b.Fatal("wrong detail page", len(rows))
					}
				}
				rows += result.Metrics.RowsRead
			}
			b.StopTimer()
			close(stop)
			wait.Wait()
			b.ReportMetric(float64(rows)/float64(b.N), "rows/op")
			b.ReportMetric(float64(peak-before.HeapAlloc), "peak-heap-growth-B")
		})
	}
}
