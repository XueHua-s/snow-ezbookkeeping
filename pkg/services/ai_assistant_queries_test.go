package services

import (
	"context"
	"encoding/json"
	"path/filepath"
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

// Each fixture uses an isolated temporary database and restores the global
// container immediately. Tests never parallelize configuration initialization.
func assistantServiceFixture(t *testing.T) (*AssistantQueryService, *xorm.Engine) {
	t.Helper()
	return assistantServiceFixtureConfig(t, &settings.DatabaseConfig{DatabaseType: settings.Sqlite3DbType, DatabasePath: filepath.Join(t.TempDir(), "assistant.db"), MaxOpenConnection: 2, MaxIdleConnection: 2})
}

func assistantServiceFixtureConfig(t *testing.T, config *settings.DatabaseConfig) (*AssistantQueryService, *xorm.Engine) {
	t.Helper()
	saved := *datastore.Container
	err := datastore.InitializeDataStore(&settings.Config{DatabaseConfig: config})
	if err != nil {
		*datastore.Container = saved
		t.Fatal(err)
	}
	store := datastore.Container.UserDataStore
	*datastore.Container = saved
	session := store.Choose(7).NewSession(core.NewNullContext())
	engine := session.Engine()
	session.Close()
	t.Cleanup(func() { engine.Close() })
	if err = store.SyncStructs(&models.Account{}, &models.TransactionCategory{}, &models.TransactionTag{}, &models.TransactionTagGroup{}, &models.Transaction{}, &models.TransactionTagIndex{}, &models.TransactionPictureInfo{}); err != nil {
		t.Fatal(err)
	}
	s, err := NewAssistantQueryService(store, 7, AssistantQueryOptions{Secret: []byte("isolated-test-secret"), Now: func() time.Time { return time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC) }})
	if err != nil {
		t.Fatal(err)
	}
	for _, bean := range []any{
		&models.Account{AccountId: 10, Uid: 7, Type: models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS, Name: "Parent"},
		&models.Account{AccountId: 11, Uid: 7, ParentAccountId: 10, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Category: models.ACCOUNT_CATEGORY_CASH, Name: "A%_\\", Currency: "CNY", Balance: 200000, Hidden: true},
		&models.Account{AccountId: 12, Uid: 7, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Category: models.ACCOUNT_CATEGORY_CASH, Name: "USD", Currency: "USD"},
		&models.Account{AccountId: 13, Uid: 7, Deleted: true, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Name: "deleted", Currency: "CNY"},
		&models.Account{AccountId: 21, Uid: 8, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Name: "other user", Currency: "CNY"},
		&models.TransactionCategory{CategoryId: 30, Uid: 7, Name: "Parent category"},
		&models.TransactionCategory{CategoryId: 31, Uid: 7, ParentCategoryId: 30, Type: models.CATEGORY_TYPE_EXPENSE, Name: "Expense", Hidden: true},
		&models.TransactionCategory{CategoryId: 32, Uid: 7, Type: models.CATEGORY_TYPE_TRANSFER, Name: "Transfer"},
		&models.TransactionCategory{CategoryId: 33, Uid: 7, Deleted: true, Name: "deleted category"},
		&models.TransactionCategory{CategoryId: 34, Uid: 8, Name: "foreign category"},
		&models.TransactionTagGroup{TagGroupId: 40, Uid: 7, Name: "Group"},
		&models.TransactionTag{TagId: 41, Uid: 7, TagGroupId: 40, Name: "One", Hidden: true},
		&models.TransactionTag{TagId: 42, Uid: 7, Name: "Two"},
		&models.TransactionTag{TagId: 43, Uid: 7, Name: "Deleted", Deleted: true},
		&models.TransactionTag{TagId: 44, Uid: 8, Name: "Foreign"},
		&models.Transaction{TransactionId: 100, Uid: 7, Type: models.TRANSACTION_DB_TYPE_EXPENSE, AccountId: 11, CategoryId: 31, TransactionTime: 1000000, Amount: 10000, Comment: "aBc%_\\", HideAmount: true},
		&models.Transaction{TransactionId: 101, Uid: 7, Type: models.TRANSACTION_DB_TYPE_EXPENSE, AccountId: 11, CategoryId: 31, TransactionTime: 1001000, Amount: -2000},
		&models.Transaction{TransactionId: 102, Uid: 7, Type: models.TRANSACTION_DB_TYPE_TRANSFER_OUT, AccountId: 11, CategoryId: 32, TransactionTime: 1002000, Amount: 4000, RelatedId: 103, RelatedAccountId: 12, RelatedAccountAmount: 500, Comment: "canonical"},
		&models.Transaction{TransactionId: 103, Uid: 7, Type: models.TRANSACTION_DB_TYPE_TRANSFER_IN, AccountId: 12, CategoryId: 32, TransactionTime: 1002001, Amount: 500, RelatedId: 102, RelatedAccountId: 11, RelatedAccountAmount: 4000, Comment: "not canonical"},
		&models.Transaction{TransactionId: 104, Uid: 7, Type: models.TRANSACTION_DB_TYPE_MODIFY_BALANCE, AccountId: 11, TransactionTime: 4102444800000, Amount: 20000},
		&models.Transaction{TransactionId: 105, Uid: 7, Deleted: true, Type: models.TRANSACTION_DB_TYPE_EXPENSE, AccountId: 11, CategoryId: 31, TransactionTime: 1003000, Amount: 999999},
		&models.Transaction{TransactionId: 106, Uid: 8, Type: models.TRANSACTION_DB_TYPE_EXPENSE, AccountId: 21, CategoryId: 34, TransactionTime: 1004000, Amount: 888888},
		&models.TransactionTagIndex{TagIndexId: 50, Uid: 7, TransactionId: 100, TagId: 41},
		&models.TransactionTagIndex{TagIndexId: 51, Uid: 7, TransactionId: 100, TagId: 42},
		&models.TransactionTagIndex{TagIndexId: 52, Uid: 7, TransactionId: 101, TagId: 41},
		&models.TransactionTagIndex{TagIndexId: 53, Uid: 7, TransactionId: 102, TagId: 42},
		&models.TransactionTagIndex{TagIndexId: 54, Uid: 7, TransactionId: 104, TagId: 41, Deleted: true},
		&models.TransactionTagIndex{TagIndexId: 55, Uid: 8, TransactionId: 104, TagId: 44},
		&models.TransactionPictureInfo{PictureId: 60, Uid: 7, TransactionId: 102},
		&models.TransactionPictureInfo{PictureId: 61, Uid: 7, TransactionId: 102, Deleted: true},
		&models.TransactionPictureInfo{PictureId: 62, Uid: 8, TransactionId: 102},
	} {
		if _, err = engine.Insert(bean); err != nil {
			t.Fatal(err)
		}
	}
	return s, engine
}
func assistantScanTestHandler(e *assistantQueryExecution) (data.ToolResult, error) {
	ids := []string{}
	err := e.ScanTransactions(nil, func(tx *assistantLogicalTransaction) error {
		ids = append(ids, strconv.FormatInt(tx.Record.TransactionId, 10))
		return nil
	})
	if err != nil {
		return data.ToolResult{}, err
	}
	return data.ToolResult{OK: true, ReturnedCount: len(ids), MatchedCount: strconv.Itoa(len(ids)), Data: assistantJSON(ids), Coverage: data.QueryCoverageAllMatching}, nil
}
func assistantCall(s *AssistantQueryService, tool data.ToolName, raw string) data.ToolResult {
	return s.Execute(core.NewAssistantContext(context.Background(), "fixture-request", "en", "127.0.0.1"), data.ToolCall{ID: "call", Name: tool, Arguments: []byte(raw)})
}
func TestAssistantStaticRegistryAndUnsupportedBoundary(t *testing.T) {
	s, _ := assistantServiceFixture(t)
	definitions := s.Definitions()
	if len(definitions) != 7 || len(s.handlers) != 7 {
		t.Fatal("registry incomplete")
	}
	for _, definition := range definitions {
		// Keep this foundation test independent of later concrete tool bodies.
		// Step 09 separately asserts all seven production factory handlers.
		s.handlers[definition.Name] = assistantToolNotImplemented
		raw := `{}`
		if definition.Name == data.ToolGetTransaction {
			raw = `{"transaction_id":"100"}`
		}
		result := assistantCall(s, definition.Name, raw)
		if !result.OK && (result.Error == nil || result.Error.Code != data.ToolErrorProviderUnsupported) {
			t.Fatalf("factory %s: %+v", definition.Name, result)
		}
		var schema map[string]any
		if err := json.Unmarshal(definition.Parameters, &schema); err != nil || schema["additionalProperties"] != false {
			t.Fatal(definition.Name)
		}
	}
	for _, name := range []data.ToolName{"delete_transaction", "execute_sql", "get_password", "query_transactions;DROP TABLE transaction"} {
		r := assistantCall(s, name, `{}`)
		if r.OK || r.Error.Code != data.ToolErrorInvalidArgument || r.Metrics.RowsRead != 0 {
			t.Fatal(name, r)
		}
	}
}
func TestAssistantReadOwnershipHiddenDeletedAndTagCombinations(t *testing.T) {
	s, _ := assistantServiceFixture(t)
	s.handlers[data.ToolQueryTransactions] = assistantScanTestHandler
	tests := []struct {
		name, raw, ids string
		code           data.ToolErrorCode
	}{
		{"all hidden future effective", `{}`, `["100","101","102","104"]`, ""},
		{"parent child dedup", `{"filter":{"account_ids":["10","11","10"]}}`, `["100","101","102","104"]`, ""},
		{"parent category", `{"filter":{"category_ids":["30","31"]}}`, `["100","101"]`, ""},
		{"mixed accounts", `{"filter":{"account_ids":["11","21"]}}`, "", data.ToolErrorNotFoundOrNotAccessible},
		{"deleted account", `{"filter":{"account_ids":["13"]}}`, "", data.ToolErrorNotFoundOrNotAccessible},
		{"mixed category", `{"filter":{"category_ids":["31","34"]}}`, "", data.ToolErrorNotFoundOrNotAccessible},
		{"deleted category", `{"filter":{"category_ids":["33"]}}`, "", data.ToolErrorNotFoundOrNotAccessible},
		{"mixed tags", `{"filter":{"tag_ids":["41","44"]}}`, "", data.ToolErrorNotFoundOrNotAccessible},
		{"deleted tag", `{"filter":{"tag_ids":["43"]}}`, "", data.ToolErrorNotFoundOrNotAccessible},
		{"any tags", `{"filter":{"tag_ids":["41","42"],"tag_match":"any"}}`, `["100","101","102"]`, ""},
		{"all tags", `{"filter":{"tag_ids":["41","42"],"tag_match":"all"}}`, `["100"]`, ""},
		{"no tags", `{"filter":{"no_tags":true}}`, `["104"]`, ""},
		{"literal percent", `{"filter":{"keyword":"%"}}`, `["100"]`, ""},
		{"literal underscore", `{"filter":{"keyword":"_"}}`, `["100"]`, ""},
		{"literal backslash", `{"filter":{"keyword":"\\"}}`, `["100"]`, ""},
		{"case insensitive", `{"filter":{"keyword":"ABC"}}`, `["100"]`, ""},
		{"half open seconds", `{"filter":{"start_time":"1970-01-01T00:16:40Z","end_time":"1970-01-01T00:16:42Z"}}`, `["100","101"]`, ""},
		{"refund", `{"filter":{"currency":"CNY","amount_max":"-1"}}`, `["101"]`, ""},
		{"destination source amount cannot mix", `{"filter":{"account_ids":["12"],"currency":"USD","amount_min":"30"}}`, `[]`, ""},
		{"destination match", `{"filter":{"account_ids":["12"],"currency":"USD","amount_min":"5","amount_max":"5"}}`, `["102"]`, ""},
		{"source only destination excluded", `{"filter":{"account_ids":["12"],"transfer_side":"source"}}`, `[]`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := assistantCall(s, data.ToolQueryTransactions, tt.raw)
			if tt.code != "" {
				if r.OK || r.Error == nil || r.Error.Code != tt.code {
					t.Fatalf("%+v", r)
				}
				return
			}
			if !r.OK || string(r.Data) != tt.ids {
				t.Fatalf("data=%s err=%+v", r.Data, r.Error)
			}
			if !r.Scope.IncludesHidden || !r.Scope.ExcludesDeleted {
				t.Fatal("false scope")
			}
		})
	}
}
func TestAssistantLogicalTransferDetailsAndConsistency(t *testing.T) {
	tests := []struct {
		name, sql string
		bad       bool
	}{
		{"reciprocal", "", false},
		{"missing source", "DELETE FROM `transaction` WHERE transaction_id=102", true},
		{"cross uid source", "UPDATE `transaction` SET uid=8 WHERE transaction_id=102", true},
		{"deleted source", "UPDATE `transaction` SET deleted=1 WHERE transaction_id=102", true},
		{"missing destination", "DELETE FROM `transaction` WHERE transaction_id=103", true},
		{"wrong reciprocal", "UPDATE `transaction` SET related_id=100 WHERE transaction_id=103", true},
		{"wrong amount", "UPDATE `transaction` SET related_account_amount=1 WHERE transaction_id=103", true},
		{"deleted required account", "UPDATE account SET deleted=1 WHERE account_id=11", true},
		{"foreign required account", "UPDATE account SET uid=8 WHERE account_id=11", true},
		{"deleted required category", "UPDATE transaction_category SET deleted=1 WHERE category_id=31", true},
		{"foreign active tag relation", "UPDATE transaction_tag_index SET tag_id=44 WHERE tag_index_id=50", true},
		{"missing active tag relation", "UPDATE transaction_tag_index SET tag_id=999 WHERE tag_index_id=50", true},
		{"deleted tag excluded", "UPDATE transaction_tag_index SET tag_id=43 WHERE tag_index_id=50", false},
		{"missing tag group", "UPDATE transaction_tag SET tag_group_id=999 WHERE tag_id=41", true},
		{"foreign parent account", "UPDATE account SET parent_account_id=21 WHERE account_id=11", true},
		{"foreign parent category", "UPDATE transaction_category SET parent_category_id=34 WHERE category_id=31", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, engine := assistantServiceFixture(t)
			if tt.sql != "" {
				if _, err := engine.Exec(tt.sql); err != nil {
					t.Fatal(err)
				}
			}
			s.handlers[data.ToolQueryTransactions] = assistantScanTestHandler
			r := assistantCall(s, data.ToolQueryTransactions, `{}`)
			if tt.bad {
				if r.OK || r.Error.Code != data.ToolErrorDataInconsistent {
					t.Fatalf("%+v", r)
				}
				return
			}
			s.handlers[data.ToolGetTransaction] = func(e *assistantQueryExecution) (data.ToolResult, error) {
				tx, err := e.ReadTransaction(e.Args.TransactionID)
				if err != nil {
					return data.ToolResult{}, err
				}
				if tx.Record.TransactionId != 102 || tx.Record.Comment != "canonical" || tx.AttachmentCount != 1 || len(tx.Tags) != 1 || tx.Tags[0].TagId != 42 || len(tx.MatchedSides) != 2 {
					t.Fatalf("wrong canonical: %+v", tx)
				}
				return data.ToolResult{OK: true, ReturnedCount: 1, Data: assistantJSON([]data.TransactionReference{tx.Reference()})}, nil
			}
			r = assistantCall(s, data.ToolGetTransaction, `{"transaction_id":"103"}`)
			if !r.OK || strings.Contains(string(r.Data), "not canonical") {
				t.Fatal(r)
			}
			for _, id := range []string{"105", "106", "999"} {
				r = assistantCall(s, data.ToolGetTransaction, `{"transaction_id":"`+id+`"}`)
				if r.OK || r.Error.Code != data.ToolErrorNotFoundOrNotAccessible {
					t.Fatal(id, r)
				}
			}
		})
	}
}
func TestAssistantQueryObservationAndBoundedBatch(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	// Put the transfer destination in a later batch; canonical source rows must
	// still be visited once and the distant incoming row must be validated.
	if _, err := engine.Exec("UPDATE `transaction` SET transaction_id=9000 WHERE transaction_id=103"); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Exec("UPDATE `transaction` SET related_id=9000 WHERE transaction_id=102"); err != nil {
		t.Fatal(err)
	}
	var observation AssistantQueryObservation
	s.observe = func(o AssistantQueryObservation) { observation = o }
	s.handlers[data.ToolQueryTransactions] = func(e *assistantQueryExecution) (data.ToolResult, error) {
		count := int64(0)
		high, err := e.HighWater()
		if err != nil {
			return data.ToolResult{}, err
		}
		if high.Time != "4102444800000" {
			t.Fatal("high water was implicitly now", high)
		}
		err = e.ScanTransactions(&high, func(tx *assistantLogicalTransaction) error { count++; return nil })
		return data.ToolResult{OK: true, ReturnedCount: 1, MatchedCount: assistantCount(count), Data: assistantJSON([]string{assistantCount(count)})}, err
	}
	batch := make([]models.Transaction, 2500)
	for i := range batch {
		batch[i] = models.Transaction{TransactionId: int64(1000 + i), Uid: 7, Type: models.TRANSACTION_DB_TYPE_EXPENSE, CategoryId: 31, AccountId: 11, Amount: 1, TransactionTime: int64(5000000 + i)}
	}
	for start := 0; start < len(batch); start += 20 {
		if _, err := engine.Insert(batch[start:min(start+20, len(batch))]); err != nil {
			t.Fatal(err)
		}
	}
	r := assistantCall(s, data.ToolQueryTransactions, `{}`)
	if !r.OK || r.MatchedCount != "2504" || r.Metrics.RowsRead < 2504 || r.Metrics.ResultRecords != 1 || observation.Metrics != r.Metrics || observation.RequestID != "fixture-request" {
		t.Fatalf("%+v %+v", r, observation)
	}
	encoded, _ := json.Marshal(r)
	if int64(len(encoded)) != r.Metrics.SerializedBytes || r.Metrics.Duration <= 0 {
		t.Fatal("wrong measurements")
	}
	s.handlers[data.ToolQueryTransactions] = func(e *assistantQueryExecution) (data.ToolResult, error) {
		return data.ToolResult{OK: true, Data: assistantJSON([]string{}), ReturnedCount: 0}, nil
	}
	r = assistantCall(s, data.ToolQueryTransactions, `{}`)
	if !r.OK || r.Metrics.RowsRead == 0 || r.Metrics.ResultRecords != 0 || r.Metrics.SerializedBytes == 0 {
		t.Fatal("empty results masquerade as zero DB work", r)
	}
	s.handlers[data.ToolQueryTransactions] = func(e *assistantQueryExecution) (data.ToolResult, error) {
		return data.ToolResult{OK: true, Data: assistantJSON([]string{strings.Repeat("x", 70000)}), ReturnedCount: 1}, nil
	}
	r = assistantCall(s, data.ToolQueryTransactions, `{}`)
	if r.OK || r.Error.Code != data.ToolErrorResultTooLarge || r.Metrics.ResultRecords != 0 || r.Metrics.RowsRead == 0 {
		t.Fatal(r)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r = s.Execute(core.NewAssistantContext(ctx, "cancel", "en", ""), data.ToolCall{Name: data.ToolQueryTransactions, Arguments: []byte(`{}`)})
	if r.OK || r.Error.Code != data.ToolErrorCanceled {
		t.Fatal(r)
	}
}

func TestAssistantPageFitsShrinksWithoutSkippingAndMeasuresActualRecords(t *testing.T) {
	s, _ := assistantServiceFixture(t)
	s.maxResultBytes = 4096
	s.handlers[data.ToolQueryTransactions] = func(e *assistantQueryExecution) (data.ToolResult, error) {
		records := []string{}
		result := data.ToolResult{OK: true, Coverage: data.QueryCoveragePage, MatchedCount: "5"}
		for i := 1; i <= 5; i++ {
			candidate := append(append([]string{}, records...), strings.Repeat("x", 900))
			cursor, err := e.NextCursor(assistantPageKey{Time: "5000", ID: "5"}, assistantPageKey{Time: strconv.Itoa(i * 1000), ID: strconv.Itoa(i)})
			if err != nil {
				return data.ToolResult{}, err
			}
			prospective := data.ToolResult{OK: true, Coverage: data.QueryCoveragePage, Data: assistantJSON(candidate), ReturnedCount: len(candidate), MatchedCount: "5", HasMore: true, NextCursor: cursor}
			if !e.PageFits(prospective) {
				break
			}
			records = candidate
			result = prospective
		}
		return result, nil
	}
	r := assistantCall(s, data.ToolQueryTransactions, `{}`)
	if !r.OK || r.ReturnedCount < 1 || r.ReturnedCount >= 5 || !r.HasMore {
		t.Fatalf("page failed to shrink: %+v", r)
	}
	cursor, err := s.decodeCursor(r.NextCursor, data.ToolQueryTransactions)
	if err != nil || cursor.Last.ID != strconv.Itoa(r.ReturnedCount) {
		t.Fatal("cursor skipped an unreturned item", cursor, err)
	}
	if r.Metrics.ResultRecords != int64(r.ReturnedCount) || r.Metrics.SerializedBytes > 4096 {
		t.Fatal("inaccurate shortened-page metrics", r.Metrics)
	}
	s.handlers[data.ToolQueryTransactions] = func(e *assistantQueryExecution) (data.ToolResult, error) {
		return data.ToolResult{OK: true, ReturnedCount: 99, Data: assistantJSON([]string{"one"})}, nil
	}
	r = assistantCall(s, data.ToolQueryTransactions, `{}`)
	if !r.OK || r.ReturnedCount != 1 || r.Metrics.ResultRecords != 1 {
		t.Fatal("trusted inaccurate handler count", r)
	}
}

func TestAssistantMetadataScopeUsesTopLevelFilters(t *testing.T) {
	s, _ := assistantServiceFixture(t)
	for _, tool := range []data.ToolName{data.ToolListAccounts, data.ToolListTags, data.ToolListCategories} {
		s.handlers[tool] = func(e *assistantQueryExecution) (data.ToolResult, error) {
			return data.ToolResult{OK: true, Data: assistantJSON([]string{})}, nil
		}
		raw := `{"keyword":"ABC%"}`
		if tool == data.ToolListCategories {
			raw = `{"keyword":"ABC%","type":"expense"}`
		}
		r := assistantCall(s, tool, raw)
		if !r.OK || r.Scope.Keyword != "abc%" {
			t.Fatal("metadata scope omitted keyword", r)
		}
		if tool == data.ToolListCategories && (len(r.Scope.Types) != 1 || r.Scope.Types[0] != "expense") {
			t.Fatal("category scope omitted type", r.Scope)
		}
	}
}
