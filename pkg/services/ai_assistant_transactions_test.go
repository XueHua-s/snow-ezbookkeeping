package services

import (
	"encoding/json"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"xorm.io/xorm"
)

func assistantTransactionRows(t *testing.T, r data.ToolResult) []assistantTransactionProjection {
	t.Helper()
	if !r.OK {
		t.Fatalf("transaction query failed: %+v", r.Error)
	}
	var rows []assistantTransactionProjection
	if err := json.Unmarshal(r.Data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != r.ReturnedCount || int64(len(rows)) != r.Metrics.ResultRecords || r.Coverage != data.QueryCoveragePage {
		t.Fatalf("incorrect page envelope: %+v", r)
	}
	if r.HasMore != (r.NextCursor != "") {
		t.Fatal("cursor/has_more disagree")
	}
	raw, _ := json.Marshal(r)
	if int64(len(raw)) != r.Metrics.SerializedBytes {
		t.Fatal("incorrect byte accounting")
	}
	return rows
}
func assistantTransactionIDs(rows []assistantTransactionProjection) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids
}
func assistantTransactionNext(s *AssistantQueryService, cursor string) data.ToolResult {
	raw, _ := json.Marshal(map[string]string{"cursor": cursor})
	return assistantCall(s, data.ToolQueryTransactions, string(raw))
}
func assistantTransactionDetail(t *testing.T, s *AssistantQueryService, id string) assistantTransactionProjection {
	t.Helper()
	raw, _ := json.Marshal(map[string]string{"transaction_id": id})
	r := assistantCall(s, data.ToolGetTransaction, string(raw))
	if !r.OK || r.MatchedCount != "1" || r.ReturnedCount != 1 || r.HasMore || len(r.References) != 1 {
		t.Fatalf("detail failed: %+v", r)
	}
	var row assistantTransactionProjection
	if err := json.Unmarshal(r.Data, &row); err != nil {
		t.Fatal(err)
	}
	return row
}
func TestAssistantTransactionsProjectionAndCanonicalDetail(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	// Private database attributes must never escape even when populated.
	if _, err := engine.ID(102).Cols("created_ip").Update(&models.Transaction{CreatedIp: "private-ip"}); err != nil {
		t.Fatal(err)
	}
	r := assistantCall(s, data.ToolQueryTransactions, `{}`)
	rows := assistantTransactionRows(t, r)
	if r.MatchedCount != "4" || !reflect.DeepEqual(assistantTransactionIDs(rows), []string{"104", "102", "101", "100"}) {
		t.Fatal(string(r.Data), r.MatchedCount)
	}
	if rows[0].CategoryID != nil || rows[0].CategoryName != "Uncategorized" || rows[0].Time != "2100-01-01T00:00:00Z" {
		t.Fatal(rows[0])
	}
	transfer := rows[1]
	if transfer.ID != "102" || transfer.SourceAccountID != "11" || transfer.DestinationAccountID != "12" || transfer.SourceAmount != "40.00" || transfer.DestinationAmount != "5.00" || transfer.Currency != "CNY" || transfer.DestinationCurrency != "USD" || transfer.Comment != "canonical" || transfer.AttachmentCount != "1" || !reflect.DeepEqual(transfer.MatchedSides, []string{"source", "destination"}) || !reflect.DeepEqual(transfer.Tags, []assistantTransactionTag{{ID: "42", Name: "Two"}}) {
		t.Fatal(transfer)
	}
	if rows[2].SourceAmount != "-20.00" || !rows[3].HideAmount || len(rows[3].Tags) != 2 {
		t.Fatal(rows)
	}
	incoming := assistantTransactionDetail(t, s, "103")
	outgoing := assistantTransactionDetail(t, s, "102")
	if !reflect.DeepEqual(incoming, outgoing) || !reflect.DeepEqual(incoming, transfer) {
		t.Fatalf("incoming differs: %+v %+v", incoming, outgoing)
	}
	for _, id := range []string{"100", "101", "104"} {
		assistantTransactionDetail(t, s, id)
	}
	for _, key := range []string{"created_ip", "private-ip", "uid", "related_id", "transaction_time", "picture_path", "storage", "similarityScore"} {
		if strings.Contains(string(r.Data), key) {
			t.Fatalf("private field escaped: %s", key)
		}
	}
}
func TestAssistantTransactionsFullHistoryAndStablePages(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	expected := []string{"100", "101", "102"}
	start := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
	for offset := 0; offset < 2500; offset += 20 {
		batch := make([]*models.Transaction, 0, 20)
		for i := offset; i < offset+20; i++ {
			id := int64(1000 + i)
			expected = append(expected, strconv.FormatInt(id, 10))
			batch = append(batch, &models.Transaction{TransactionId: id, Uid: 7, Type: models.TRANSACTION_DB_TYPE_EXPENSE, AccountId: 11, CategoryId: 31, TransactionTime: start.AddDate(0, 0, i).Unix() * 1000, Amount: int64(i)})
		}
		if _, err := engine.Insert(batch); err != nil {
			t.Fatal(err)
		}
	}
	// Place the incoming half in a distant storage batch. Pagination still
	// emits the canonical source once and resolves its related side in batches.
	if _, err := engine.ID(103).Cols("transaction_id").Update(&models.Transaction{TransactionId: 5000}); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.ID(102).Cols("related_id").Update(&models.Transaction{RelatedId: 5000}); err != nil {
		t.Fatal(err)
	}
	expected = append(expected, "104")
	for _, direction := range []string{"asc", "desc"} {
		want := append([]string{}, expected...)
		if direction == "desc" {
			for i, j := 0, len(want)-1; i < j; i, j = i+1, j-1 {
				want[i], want[j] = want[j], want[i]
			}
		}
		raw := `{"page_size":200,"order_by":{"field":"time","direction":"` + direction + `"}}`
		result := assistantCall(s, data.ToolQueryTransactions, raw)
		seen := []string{}
		pages := 0
		for {
			pages++
			if pages > 20 {
				t.Fatal("pagination did not finish")
			}
			rows := assistantTransactionRows(t, result)
			if result.MatchedCount != "2504" {
				t.Fatal(result.MatchedCount)
			}
			seen = append(seen, assistantTransactionIDs(rows)...)
			if len(result.References) > 8 {
				t.Fatal("too many references")
			}
			if !result.HasMore {
				break
			}
			cursor, err := s.decodeCursor(result.NextCursor, data.ToolQueryTransactions)
			if err != nil {
				t.Fatal(err)
			}
			if cursor.HighWater.ID != "104" || cursor.HighWater.Time != "4102444800000" || cursor.Last.ID != rows[len(rows)-1].ID {
				t.Fatalf("incorrect cursor: %+v", cursor)
			}
			result = assistantTransactionNext(s, result.NextCursor)
		}
		if !reflect.DeepEqual(seen, want) {
			t.Fatalf("%s pages omitted or repeated rows: got %d want %d", direction, len(seen), len(want))
		}
	}
}
func TestAssistantTransactionsMatchingAndAmountOrder(t *testing.T) {
	s, _ := assistantServiceFixture(t)
	cases := []struct {
		raw string
		ids []string
	}{
		{`{"filter":{"account_ids":["10","11"]}}`, []string{"104", "102", "101", "100"}},
		{`{"filter":{"category_ids":["30","31"]}}`, []string{"101", "100"}},
		{`{"filter":{"tag_ids":["41","42"],"tag_match":"any"}}`, []string{"102", "101", "100"}},
		{`{"filter":{"tag_ids":["41","42"],"tag_match":"all"}}`, []string{"100"}},
		{`{"filter":{"no_tags":true}}`, []string{"104"}},
		{`{"filter":{"account_ids":["12"],"transfer_side":"source"}}`, []string{}},
		{`{"filter":{"account_ids":["11"],"transfer_side":"destination","types":["transfer"]}}`, []string{}},
		{`{"filter":{"account_ids":["12"],"currency":"CNY"}}`, []string{}},
		{`{"filter":{"account_ids":["12"],"currency":"USD","amount_min":"30"}}`, []string{}},
		{`{"filter":{"account_ids":["12"],"currency":"USD","amount_min":"5","amount_max":"5"}}`, []string{"102"}},
		{`{"filter":{"currency":"CNY","amount_max":"-1"}}`, []string{"101"}},
		{`{"filter":{"currency":"CNY","transfer_side":"source"},"order_by":{"field":"amount","direction":"asc"}}`, []string{"101", "102", "100", "104"}},
		{`{"filter":{"currency":"CNY","transfer_side":"source"},"order_by":{"field":"amount","direction":"desc"}}`, []string{"104", "100", "102", "101"}},
		{`{"filter":{"currency":"USD","transfer_side":"destination"},"order_by":{"field":"amount","direction":"asc"}}`, []string{"102"}},
		{`{"filter":{"currency":"CNY","types":["expense"]},"order_by":{"field":"amount","direction":"asc"}}`, []string{"101", "100"}},
	}
	for _, tc := range cases {
		r := assistantCall(s, data.ToolQueryTransactions, tc.raw)
		rows := assistantTransactionRows(t, r)
		if !reflect.DeepEqual(assistantTransactionIDs(rows), tc.ids) || r.MatchedCount != strconv.Itoa(len(tc.ids)) {
			t.Fatalf("%s: %s", tc.raw, r.Data)
		}
		if len(rows) == 1 && rows[0].ID == "102" && strings.Contains(tc.raw, `"USD"`) && !reflect.DeepEqual(rows[0].MatchedSides, []string{"destination"}) {
			t.Fatal(rows[0])
		}
	}
	for _, raw := range []string{
		`{"filter":{"tag_match":"all"}}`, `{"filter":{"tag_ids":["41"],"no_tags":true}}`,
		`{"order_by":{"field":"amount","direction":"asc"}}`,
		`{"filter":{"currency":"CNY"},"order_by":{"field":"amount","direction":"desc"}}`,
	} {
		r := assistantCall(s, data.ToolQueryTransactions, raw)
		if r.OK || r.Error.Code != data.ToolErrorInvalidArgument || r.Metrics.RowsRead != 0 {
			t.Fatal(raw, r)
		}
	}
}
func TestAssistantTransactionsStableTiesAndAmountCursor(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	for _, id := range []int64{200, 201, 202} {
		if _, err := engine.Insert(&models.Transaction{TransactionId: id, Uid: 7, Type: models.TRANSACTION_DB_TYPE_EXPENSE, AccountId: 11, CategoryId: 31, TransactionTime: 1001000 + id - 199, Amount: -2000}); err != nil {
			t.Fatal(err)
		}
	}
	for _, field := range []string{"amount", "time"} {
		for _, direction := range []string{"asc", "desc"} {
			r := assistantCall(s, data.ToolQueryTransactions, `{"filter":{"types":["expense"],"currency":"CNY"},"page_size":1,"order_by":{"field":"`+field+`","direction":"`+direction+`"}}`)
			ids := []string{}
			for {
				rows := assistantTransactionRows(t, r)
				if r.MatchedCount != "5" || len(rows) != 1 {
					t.Fatal(r)
				}
				ids = append(ids, rows[0].ID)
				if !r.HasMore {
					break
				}
				cursor, err := s.decodeCursor(r.NextCursor, data.ToolQueryTransactions)
				if err != nil {
					t.Fatal(err)
				}
				if field == "amount" && cursor.Last.Amount == "" {
					t.Fatal("amount missing from continuation")
				}
				r = assistantTransactionNext(s, r.NextCursor)
			}
			want := []string{"101", "200", "201", "202", "100"}
			if field == "time" {
				want = []string{"100", "101", "200", "201", "202"}
			}
			if direction == "desc" {
				for i, j := 0, len(want)-1; i < j; i, j = i+1, j-1 {
					want[i], want[j] = want[j], want[i]
				}
			}
			if !reflect.DeepEqual(ids, want) {
				t.Fatal(field, direction, ids)
			}
		}
	}
}
func TestAssistantTransactionsByteBudgetKeepsActualLastRow(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	for _, id := range []int64{100, 101, 102, 104} {
		if _, err := engine.ID(id).Cols("comment").Update(&models.Transaction{Comment: strings.Repeat("账", 255)}); err != nil {
			t.Fatal(err)
		}
	}
	// Budget deliberately fits one complete record and cursor, but not all four.
	s.maxResultBytes = 3500
	r := assistantCall(s, data.ToolQueryTransactions, `{"page_size":4}`)
	seen := []string{}
	for {
		rows := assistantTransactionRows(t, r)
		if len(rows) == 0 || len(rows) >= 4 || r.Metrics.SerializedBytes > 3500 {
			t.Fatalf("not shortened: %+v", r)
		}
		for _, row := range rows {
			if row.Comment != strings.Repeat("账", 255) {
				t.Fatal("truncated text")
			}
		}
		seen = append(seen, assistantTransactionIDs(rows)...)
		if !r.HasMore {
			break
		}
		cursor, err := s.decodeCursor(r.NextCursor, data.ToolQueryTransactions)
		if err != nil || cursor.Last.ID != rows[len(rows)-1].ID {
			t.Fatal("cursor skipped omitted rows", err)
		}
		r = assistantTransactionNext(s, r.NextCursor)
	}
	if !reflect.DeepEqual(seen, []string{"104", "102", "101", "100"}) {
		t.Fatal(seen)
	}
	if _, err := engine.ID(104).Cols("comment").Update(&models.Transaction{Comment: strings.Repeat("x", 10000)}); err != nil {
		t.Fatal(err)
	}
	for _, call := range []struct {
		tool data.ToolName
		raw  string
	}{{data.ToolQueryTransactions, `{}`}, {data.ToolGetTransaction, `{"transaction_id":"104"}`}} {
		r := assistantCall(s, call.tool, call.raw)
		if r.OK || r.Error.Code != data.ToolErrorResultTooLarge || len(r.Data) != 0 || len(r.References) != 0 {
			t.Fatal(r)
		}
	}
}
func TestAssistantTransactionsHighWaterIsMatchingAndReauthorized(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	r := assistantCall(s, data.ToolQueryTransactions, `{"filter":{"category_ids":["30"]},"page_size":1}`)
	assistantTransactionRows(t, r)
	c, err := s.decodeCursor(r.NextCursor, data.ToolQueryTransactions)
	if err != nil || c.HighWater.ID != "101" {
		t.Fatal(c, err)
	}
	if _, err := engine.Insert(&models.Transaction{TransactionId: 200, Uid: 7, Type: models.TRANSACTION_DB_TYPE_EXPENSE, AccountId: 11, CategoryId: 31, TransactionTime: 1005000, Amount: 1}); err != nil {
		t.Fatal(err)
	}
	next := assistantTransactionNext(s, r.NextCursor)
	rows := assistantTransactionRows(t, next)
	if next.MatchedCount != "2" || !reflect.DeepEqual(assistantTransactionIDs(rows), []string{"100"}) {
		t.Fatal(next)
	}
	for _, raw := range []string{
		`{"cursor":` + strconv.Quote(r.NextCursor+"x") + `}`,
		`{"cursor":` + strconv.Quote(r.NextCursor) + `,"filter":{"category_ids":["31"]}}`,
	} {
		result := assistantCall(s, data.ToolQueryTransactions, raw)
		if result.OK || result.Error.Code != data.ToolErrorInvalidCursor {
			t.Fatal(result)
		}
	}
	if _, err := engine.Insert(&models.TransactionCategory{CategoryId: 35, Uid: 7, ParentCategoryId: 30, Name: "new child"}); err != nil {
		t.Fatal(err)
	}
	next = assistantTransactionNext(s, r.NextCursor)
	if next.OK || next.Error.Code != data.ToolErrorInvalidCursor {
		t.Fatal(next)
	}
}
func TestAssistantTransactionsInaccessibleAndBrokenRelations(t *testing.T) {
	s, _ := assistantServiceFixture(t)
	for _, id := range []string{"105", "106", "999"} {
		r := assistantCall(s, data.ToolGetTransaction, `{"transaction_id":"`+id+`"}`)
		if r.OK || r.Error.Code != data.ToolErrorNotFoundOrNotAccessible {
			t.Fatal(id, r)
		}
	}
	for _, raw := range []string{`{"filter":{"account_ids":["11","21"]}}`, `{"filter":{"category_ids":["33"]}}`, `{"filter":{"tag_ids":["44"]}}`} {
		r := assistantCall(s, data.ToolQueryTransactions, raw)
		if r.OK || r.Error.Code != data.ToolErrorNotFoundOrNotAccessible {
			t.Fatal(raw, r)
		}
	}
	cases := []struct {
		name, id string
		mutate   func(*xorm.Engine) error
	}{
		{"deleted account", "100", func(e *xorm.Engine) error {
			_, err := e.ID(100).Cols("account_id").Update(&models.Transaction{AccountId: 13})
			return err
		}},
		{"foreign account", "100", func(e *xorm.Engine) error {
			_, err := e.ID(100).Cols("account_id").Update(&models.Transaction{AccountId: 21})
			return err
		}},
		{"deleted category", "100", func(e *xorm.Engine) error {
			_, err := e.ID(100).Cols("category_id").Update(&models.Transaction{CategoryId: 33})
			return err
		}},
		{"foreign category", "100", func(e *xorm.Engine) error {
			_, err := e.ID(100).Cols("category_id").Update(&models.Transaction{CategoryId: 34})
			return err
		}},
		{"foreign tag", "100", func(e *xorm.Engine) error {
			_, err := e.Insert(&models.TransactionTagIndex{TagIndexId: 70, Uid: 7, TransactionId: 100, TagId: 44})
			return err
		}},
		{"missing transfer", "102", func(e *xorm.Engine) error { _, err := e.ID(103).Delete(&models.Transaction{}); return err }},
		{"foreign transfer", "102", func(e *xorm.Engine) error {
			_, err := e.ID(103).Cols("uid").Update(&models.Transaction{Uid: 8})
			return err
		}},
		{"deleted transfer", "102", func(e *xorm.Engine) error {
			_, err := e.ID(103).Cols("deleted").Update(&models.Transaction{Deleted: true})
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, e := assistantServiceFixture(t)
			if err := tc.mutate(e); err != nil {
				t.Fatal(err)
			}
			for _, call := range []struct {
				tool data.ToolName
				raw  string
			}{{data.ToolQueryTransactions, `{}`}, {data.ToolGetTransaction, `{"transaction_id":"` + tc.id + `"}`}} {
				r := assistantCall(s, call.tool, call.raw)
				if r.OK || r.Error.Code != data.ToolErrorDataInconsistent || len(r.Data) > 0 || len(r.References) > 0 {
					t.Fatal(r)
				}
			}
		})
	}
}
func TestAssistantTransactionsEmptyAndDeletedTag(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	if _, err := engine.Insert(&models.TransactionTagIndex{TagIndexId: 70, Uid: 7, TransactionId: 104, TagId: 43}); err != nil {
		t.Fatal(err)
	}
	rows := assistantTransactionRows(t, assistantCall(s, data.ToolQueryTransactions, `{"filter":{"no_tags":true}}`))
	if !reflect.DeepEqual(assistantTransactionIDs(rows), []string{"104"}) {
		t.Fatal(rows)
	}
	r := assistantCall(s, data.ToolQueryTransactions, `{"filter":{"keyword":"no matching text"}}`)
	rows = assistantTransactionRows(t, r)
	if len(rows) != 0 || r.MatchedCount != "0" || r.HasMore || string(r.Data) != "[]" {
		t.Fatal(r)
	}
}
func TestAssistantTransactionsExactOrderKeys(t *testing.T) {
	// Exercise exact integer comparison near both signed limits without relying
	// on subtraction or floating point, and document stable secondary keys.
	order := assistantOrder{Field: "amount", Direction: "asc"}
	keys := []assistantPageKey{{ID: "3", Time: "1", Amount: "9223372036854775807"}, {ID: "2", Time: "1", Amount: "-9223372036854775808"}, {ID: "1", Time: "1", Amount: "-9223372036854775808"}}
	sort.Slice(keys, func(i, j int) bool { return assistantCompareTransactionKeys(keys[i], keys[j], order) < 0 })
	if keys[0].ID != "1" || keys[1].ID != "2" || keys[2].ID != "3" {
		t.Fatal(keys)
	}
}

func TestAssistantTransactionsExternal(t *testing.T) {
	for _, kind := range []string{settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			s, _ := assistantServiceFixtureConfig(t, assistantExternalServiceConfig(t, kind))
			for _, raw := range []string{`{"filter":{"keyword":"%"}}`, `{"filter":{"keyword":"_"}}`, `{"filter":{"keyword":"\\"}}`, `{"filter":{"keyword":"ABC"}}`, `{"filter":{"tag_ids":["41","42"],"tag_match":"all"}}`} {
				r := assistantCall(s, data.ToolQueryTransactions, raw)
				rows := assistantTransactionRows(t, r)
				if !reflect.DeepEqual(assistantTransactionIDs(rows), []string{"100"}) || r.MatchedCount != "1" {
					t.Fatal(raw, r)
				}
			}
			for _, field := range []string{"time", "amount"} {
				r := assistantCall(s, data.ToolQueryTransactions, `{"filter":{"currency":"CNY","transfer_side":"source"},"page_size":1,"order_by":{"field":"`+field+`","direction":"asc"}}`)
				ids := []string{}
				for {
					rows := assistantTransactionRows(t, r)
					if r.MatchedCount != "4" {
						t.Fatal(r)
					}
					ids = append(ids, assistantTransactionIDs(rows)...)
					if !r.HasMore {
						break
					}
					r = assistantTransactionNext(s, r.NextCursor)
				}
				expected := []string{"100", "101", "102", "104"}
				if field == "amount" {
					expected = []string{"101", "102", "100", "104"}
				}
				if !reflect.DeepEqual(ids, expected) {
					t.Fatal(field, ids)
				}
			}
			row := assistantTransactionDetail(t, s, "103")
			if row.ID != "102" || row.Comment != "canonical" || row.AttachmentCount != "1" || row.SourceAmount != "40.00" || row.DestinationAmount != "5.00" || len(row.Tags) != 1 {
				t.Fatal(row)
			}
			for _, raw := range []string{`{"filter":{"account_ids":["11","21"]}}`, `{"filter":{"category_ids":["33"]}}`} {
				r := assistantCall(s, data.ToolQueryTransactions, raw)
				if r.OK || r.Error.Code != data.ToolErrorNotFoundOrNotAccessible {
					t.Fatal(r)
				}
			}
		})
	}
}
