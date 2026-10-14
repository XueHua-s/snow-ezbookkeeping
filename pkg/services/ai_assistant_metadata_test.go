package services

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"xorm.io/xorm"
)

func assistantMetadataRows(t *testing.T, result data.ToolResult) []map[string]any {
	t.Helper()
	if !result.OK {
		t.Fatalf("query failed: %+v", result.Error)
	}
	var rows []map[string]any
	if err := json.Unmarshal(result.Data, &rows); err != nil {
		t.Fatal(err)
	}
	if rows == nil || len(rows) != result.ReturnedCount {
		t.Fatalf("invalid page count/data: %+v", result)
	}
	if result.Coverage != data.QueryCoveragePage || !result.Scope.IncludesHidden || !result.Scope.ExcludesDeleted || result.QueryID == "" || result.QueriedAt == "" {
		t.Fatalf("missing evidence: %+v", result)
	}
	return rows
}

func assistantMetadataWalk(t *testing.T, s *AssistantQueryService, tool data.ToolName, raw string) ([]map[string]any, []data.ToolResult) {
	t.Helper()
	all := []map[string]any{}
	pages := []data.ToolResult{}
	previous := int64(0)
	for page := 0; page < 1000; page++ {
		result := assistantCall(s, tool, raw)
		rows := assistantMetadataRows(t, result)
		pages = append(pages, result)
		for _, row := range rows {
			key := "id"
			if tool == data.ToolGetAccountBalances {
				key = "account_id"
			}
			id, err := strconv.ParseInt(row[key].(string), 10, 64)
			if err != nil || id <= previous {
				t.Fatalf("unordered or duplicate ID: %v after %d", row[key], previous)
			}
			previous = id
		}
		all = append(all, rows...)
		if !result.HasMore {
			if result.NextCursor != "" {
				t.Fatal("terminal cursor")
			}
			return all, pages
		}
		if len(rows) == 0 || result.NextCursor == "" {
			t.Fatal("non-progressing page")
		}
		cursor, err := s.decodeCursor(result.NextCursor, tool)
		if err != nil || cursor.Last.ID != assistantCount(previous) {
			t.Fatal("cursor skips actual last record", cursor, err)
		}
		raw = string(assistantJSON(map[string]string{"cursor": result.NextCursor}))
	}
	t.Fatal("too many pages")
	return nil, nil
}

func assistantInsertEntity(t *testing.T, engine *xorm.Engine, bean any) {
	t.Helper()
	if _, err := engine.Insert(bean); err != nil {
		t.Fatal(err)
	}
}

func TestAssistantMetadataFullPagesHierarchyAndIsolation(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	for i := int64(0); i < 205; i++ {
		assistantInsertEntity(t, engine, &models.Account{AccountId: 1000 + i, Uid: 7, Name: "Same", ParentAccountId: 10, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Category: models.ACCOUNT_CATEGORY_CASH, Currency: "CNY", Hidden: i%2 == 0})
		assistantInsertEntity(t, engine, &models.TransactionCategory{CategoryId: 2000 + i, Uid: 7, Name: "Same", ParentCategoryId: 30, Type: models.CATEGORY_TYPE_EXPENSE, Hidden: i%2 == 0})
		assistantInsertEntity(t, engine, &models.TransactionTag{TagId: 3000 + i, Uid: 7, Name: "Same", TagGroupId: 40, Hidden: i%2 == 0})
	}
	assistantInsertEntity(t, engine, &models.Account{AccountId: 9007199254740993, Uid: 7, Name: "Precise ID", Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Category: models.ACCOUNT_CATEGORY_CASH, Currency: "CNY"})
	for _, tc := range []struct {
		tool                                              data.ToolName
		count                                             int
		hiddenID, parentKey, parent, deletedID, foreignID string
	}{
		{data.ToolListAccounts, 209, "11", "parent_id", "10", "13", "21"},
		{data.ToolListCategories, 208, "31", "parent_id", "30", "33", "34"},
		{data.ToolListTags, 207, "41", "tag_group_id", "40", "43", "44"},
	} {
		t.Run(string(tc.tool), func(t *testing.T) {
			rows, pages := assistantMetadataWalk(t, s, tc.tool, `{"page_size":50}`)
			if len(rows) != tc.count || len(pages) != 5 {
				t.Fatalf("incomplete pages: %d rows %d pages", len(rows), len(pages))
			}
			found := false
			for _, row := range rows {
				if row["id"] == tc.deletedID || row["id"] == tc.foreignID {
					t.Fatal("inaccessible metadata leaked", row)
				}
				if row["id"] == tc.hiddenID {
					found = true
					if row["hidden"] != true || row[tc.parentKey] != tc.parent {
						t.Fatal("hidden/hierarchy lost", row)
					}
					if tc.tool == data.ToolListTags && row["tag_group_name"] != "Group" {
						t.Fatal(row)
					}
				}
			}
			if !found {
				t.Fatal("hidden entity missing")
			}
			for _, page := range pages {
				if page.MatchedCount != strconv.Itoa(tc.count) {
					t.Fatal("wrong complete match count", page.MatchedCount)
				}
			}
		})
	}
}

func TestAssistantMetadataLiteralSearchAndTypeScope(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	for _, bean := range []any{
		&models.TransactionCategory{CategoryId: 70, Uid: 7, Name: `AbC%_\`, Type: models.CATEGORY_TYPE_INCOME},
		&models.TransactionCategory{CategoryId: 71, Uid: 7, Name: `AbC%_\`, Type: models.CATEGORY_TYPE_EXPENSE},
		&models.TransactionCategory{CategoryId: 72, Uid: 7, Name: `AbCxx`, Type: models.CATEGORY_TYPE_INCOME},
		&models.TransactionTag{TagId: 80, Uid: 7, Name: `AbC%_\`},
		&models.TransactionTag{TagId: 81, Uid: 7, Name: `AbCxx`},
	} {
		assistantInsertEntity(t, engine, bean)
	}
	for _, tc := range []struct {
		tool    data.ToolName
		keyword string
		id      string
	}{
		{data.ToolListAccounts, `a%_\`, "11"}, {data.ToolListCategories, `abc%_\`, "70"}, {data.ToolListTags, `abc%_\`, "80"},
	} {
		args := map[string]any{"keyword": tc.keyword}
		if tc.tool == data.ToolListCategories {
			args["type"] = "income"
		}
		result := assistantCall(s, tc.tool, string(assistantJSON(args)))
		rows := assistantMetadataRows(t, result)
		if len(rows) != 1 || rows[0]["id"] != tc.id || result.Scope.Keyword != tc.keyword {
			t.Fatal("literal filter or scope", rows, result.Scope)
		}
		if tc.tool == data.ToolListCategories && !reflect.DeepEqual(result.Scope.Types, []string{"income"}) {
			t.Fatal(result.Scope)
		}
	}
}

func TestAssistantMetadataEmptyLedger(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	// Only the temporary fixture database is cleared.
	for _, bean := range []any{&models.Account{}, &models.TransactionCategory{}, &models.TransactionTag{}, &models.TransactionTagGroup{}} {
		if _, err := engine.Where("uid = ?", 7).Delete(bean); err != nil {
			t.Fatal(err)
		}
	}
	for _, tool := range []data.ToolName{data.ToolListAccounts, data.ToolListCategories, data.ToolListTags, data.ToolGetAccountBalances} {
		result := assistantCall(s, tool, `{}`)
		rows := assistantMetadataRows(t, result)
		if len(rows) != 0 || result.HasMore || result.MatchedCount != "0" || result.NextCursor != "" {
			t.Fatal(result)
		}
		if tool == data.ToolGetAccountBalances && string(result.Totals) != "[]" {
			t.Fatal("invented empty totals", string(result.Totals))
		}
	}
}

func TestAssistantMetadataByteLimitedPagesDoNotSkip(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	for i := int64(0); i < 12; i++ {
		name := strings.Repeat("名", 64)
		assistantInsertEntity(t, engine, &models.Account{AccountId: 1000 + i, Uid: 7, Name: name, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Category: models.ACCOUNT_CATEGORY_CASH, Currency: "CNY"})
		assistantInsertEntity(t, engine, &models.TransactionCategory{CategoryId: 2000 + i, Uid: 7, Name: name, Type: models.CATEGORY_TYPE_EXPENSE})
		assistantInsertEntity(t, engine, &models.TransactionTag{TagId: 3000 + i, Uid: 7, Name: name})
	}
	for _, tool := range []data.ToolName{data.ToolListAccounts, data.ToolListCategories, data.ToolListTags, data.ToolGetAccountBalances} {
		t.Run(string(tool), func(t *testing.T) {
			s.maxResultBytes = 64 * 1024
			expected, _ := assistantMetadataWalk(t, s, tool, `{"page_size":10}`)
			full := assistantCall(s, tool, `{"page_size":10}`)
			encoded, _ := json.Marshal(full)
			s.maxResultBytes = len(encoded) - 300
			actual, pages := assistantMetadataWalk(t, s, tool, `{"page_size":10}`)
			if !reflect.DeepEqual(expected, actual) || pages[0].ReturnedCount >= 10 {
				t.Fatalf("byte shortening lost records: %d/%d first %d", len(actual), len(expected), pages[0].ReturnedCount)
			}
			for _, page := range pages {
				encoded, _ := json.Marshal(page)
				if len(encoded) > s.maxResultBytes {
					t.Fatal("oversized envelope")
				}
			}
		})
	}
}

func TestAssistantMetadataOversizedEntityFails(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	for _, tool := range []data.ToolName{data.ToolListAccounts, data.ToolListCategories, data.ToolListTags, data.ToolGetAccountBalances} {
		name := strings.Repeat("x", 4096)
		switch tool {
		case data.ToolListAccounts:
			assistantInsertEntity(t, engine, &models.Account{AccountId: 1, Uid: 7, Name: name, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Category: models.ACCOUNT_CATEGORY_CASH, Currency: "CNY"})
		case data.ToolListCategories:
			assistantInsertEntity(t, engine, &models.TransactionCategory{CategoryId: 1, Uid: 7, Name: name, Type: models.CATEGORY_TYPE_EXPENSE})
		case data.ToolListTags:
			assistantInsertEntity(t, engine, &models.TransactionTag{TagId: 1, Uid: 7, Name: name})
		}
		s.maxResultBytes = 2048
		result := assistantCall(s, tool, `{}`)
		if result.OK || result.Error == nil || result.Error.Code != data.ToolErrorResultTooLarge {
			t.Fatalf("%s silently truncated: %+v", tool, result)
		}
	}
}
