package services

import (
	"context"
	"strings"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func TestAssistantCategoryUsageFullLedgerHierarchyAndIsolation(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	for _, category := range []*models.TransactionCategory{
		{CategoryId: 50, Uid: 7, Name: "餐饮", Type: models.CATEGORY_TYPE_EXPENSE, Hidden: true},
		{CategoryId: 51, Uid: 7, Name: "食品", Type: models.CATEGORY_TYPE_EXPENSE, Hidden: true, ParentCategoryId: 60},
		{CategoryId: 52, Uid: 7, Name: "食品", Type: models.CATEGORY_TYPE_EXPENSE, ParentCategoryId: 60},
		{CategoryId: 55, Uid: 7, Name: "删除", Type: models.CATEGORY_TYPE_EXPENSE, Deleted: true},
		{CategoryId: 60, Uid: 7, Name: "生活", Type: models.CATEGORY_TYPE_EXPENSE},
		{CategoryId: 61, Uid: 7, Name: "生活子类", Type: models.CATEGORY_TYPE_EXPENSE, ParentCategoryId: 60},
		{CategoryId: 62, Uid: 7, Name: "生活孙类", Type: models.CATEGORY_TYPE_EXPENSE, ParentCategoryId: 61},
	} {
		assistantInsertEntity(t, engine, category)
	}
	for i, tx := range []*models.Transaction{
		{Uid: 7, AccountId: 11, CategoryId: 51, Amount: 100},
		{Uid: 7, AccountId: 11, CategoryId: 51, Amount: -100},
		{Uid: 7, AccountId: 12, CategoryId: 52, Amount: 100},
		{Uid: 7, AccountId: 11, CategoryId: 61, Amount: 0},
		{Uid: 7, AccountId: 11, CategoryId: 62, Amount: 20},
		{Uid: 7, AccountId: 11, CategoryId: 51, Amount: 999, Deleted: true},
		{Uid: 8, AccountId: 21, CategoryId: 51, Amount: 999},
		{Uid: 7, AccountId: 11, CategoryId: 55, Amount: 999, Deleted: true},
	} {
		tx.TransactionId = int64(200 + i)
		tx.TransactionTime = int64(2000000 + i)
		tx.Type = models.TRANSACTION_DB_TYPE_EXPENSE
		assistantInsertEntity(t, engine, tx)
	}
	rows, pages := assistantMetadataWalk(t, s, data.ToolListCategories, `{"page_size":1}`)
	if len(pages) != 9 {
		t.Fatal("incomplete metadata pages", len(pages))
	}
	expected := map[string][2]string{"30": {"0", "2"}, "31": {"2", "2"}, "32": {"1", "1"}, "50": {"0", "0"}, "51": {"2", "2"}, "52": {"1", "1"}, "60": {"0", "5"}, "61": {"1", "2"}, "62": {"1", "1"}}
	for _, row := range rows {
		counts, ok := expected[row["id"].(string)]
		if !ok || row["direct_transaction_count"] != counts[0] || row["subtree_transaction_count"] != counts[1] {
			t.Fatal("wrong logical usage", row)
		}
	}
	zero := assistantCall(s, data.ToolAggregateTransactions, `{"filter":{"category_ids":["51"],"currency":"CNY"},"metrics":["sum","count"]}`)
	if !zero.OK || zero.MatchedCount != "2" || !strings.Contains(string(zero.Totals), "0.00") {
		t.Fatal("refund cancellation was confused with unused", zero)
	}
	// Literal keyword never becomes a semantic topic or a comment predicate.
	literal := assistantMetadataRows(t, assistantCall(s, data.ToolListCategories, `{"keyword":"餐饮","type":"expense"}`))
	if len(literal) != 1 || literal[0]["id"] != "50" || literal[0]["subtree_transaction_count"] != "0" {
		t.Fatal(literal)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	result := s.Execute(core.NewAssistantContext(canceled, "", "", ""), data.ToolCall{ID: "canceled", Name: data.ToolListCategories, Arguments: []byte(`{}`)})
	if result.OK {
		t.Fatal("canceled full-ledger usage scan succeeded")
	}
}

func TestAssistantCategoryUsageOneScanAndCorruptionIsNotZero(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	before := assistantCall(s, data.ToolListCategories, `{}`)
	for i := int64(0); i < 100; i++ {
		assistantInsertEntity(t, engine, &models.TransactionCategory{CategoryId: 2000 + i, Uid: 7, Name: "unused", Type: models.CATEGORY_TYPE_EXPENSE})
	}
	after := assistantCall(s, data.ToolListCategories, `{}`)
	if !before.OK || !after.OK || after.Metrics.RowsRead-before.Metrics.RowsRead != 100 {
		t.Fatal("category usage rescanned per entity", before.Metrics, after.Metrics)
	}
	// A broken transfer is a query failure, never evidence that its category is unused.
	if _, err := engine.Where("uid = ? AND transaction_id = ?", 7, 103).Delete(&models.Transaction{}); err != nil {
		t.Fatal(err)
	}
	broken := assistantCall(s, data.ToolListCategories, `{}`)
	if broken.OK || broken.Error == nil || broken.Error.Code != data.ToolErrorDataInconsistent {
		t.Fatal("corruption hidden as zero", broken)
	}
}

func TestAssistantCategoryResolutionPolicySurvivesSchemaCompaction(t *testing.T) {
	s, _ := assistantServiceFixture(t)
	definitions := assistantContextTools{AssistantQueryService: s}.Definitions()
	for _, definition := range definitions {
		if definition.Name == data.ToolListCategories {
			for _, term := range []string{"literal", "not a topic", "direct_transaction_count", "subtree_transaction_count", "refunds", "transfers count once"} {
				if !strings.Contains(definition.Description, term) {
					t.Fatal("policy stripped", term)
				}
			}
			return
		}
	}
	t.Fatal("category tool missing")
}
