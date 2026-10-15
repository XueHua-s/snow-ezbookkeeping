package services

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func assistantReadBalanceTotals(t *testing.T, result data.ToolResult) map[string]assistantBalanceTotal {
	t.Helper()
	if !result.OK || result.TotalsCoverage != data.QueryCoverageAllMatching {
		t.Fatalf("incomplete balance totals: %+v", result)
	}
	var totals []assistantBalanceTotal
	if err := json.Unmarshal(result.Totals, &totals); err != nil {
		t.Fatal(err)
	}
	got := make(map[string]assistantBalanceTotal)
	previous := ""
	for _, total := range totals {
		if previous >= total.Currency {
			t.Fatal("currency order or duplicate", totals)
		}
		previous = total.Currency
		got[total.Currency] = total
	}
	return got
}

func TestAssistantBalancesFullTotalsLeafHierarchyAndOverpayment(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	assistantInsertEntity(t, engine, &models.Account{AccountId: 14, Uid: 7, ParentAccountId: 10, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Category: models.ACCOUNT_CATEGORY_CREDIT_CARD, Name: "Card", Currency: "CNY", Balance: -30000, Hidden: true})
	assistantInsertEntity(t, engine, &models.Account{AccountId: 15, Uid: 7, Type: models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS, Category: models.ACCOUNT_CATEGORY_CASH, Name: "Empty parent", Currency: "CNY", Balance: 99999999})
	if _, err := engine.ID(10).Cols("balance").Update(&models.Account{Balance: 99999999}); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.ID(13).Cols("balance").Update(&models.Account{Balance: 88888888}); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.ID(21).Cols("balance").Update(&models.Account{Balance: 77777777}); err != nil {
		t.Fatal(err)
	}
	rows, pages := assistantMetadataWalk(t, s, data.ToolGetAccountBalances, `{"page_size":1}`)
	if len(rows) != 3 || len(pages) != 3 {
		t.Fatal("parent/deleted/foreign leaf included", rows)
	}
	expected := assistantBalanceTotal{Currency: "CNY", AssetBalance: "2000.00", LiabilityBalance: "300.00", NetAssets: "1700.00"}
	for _, page := range pages {
		totals := assistantReadBalanceTotals(t, page)
		if totals["CNY"] != expected || totals["USD"] != (assistantBalanceTotal{Currency: "USD", AssetBalance: "0.00", LiabilityBalance: "0.00", NetAssets: "0.00"}) || page.MatchedCount != "3" {
			t.Fatal("incorrect full totals", totals, page.MatchedCount)
		}
		if string(page.Totals) != string(pages[0].Totals) {
			t.Fatal("pagination changed totals")
		}
	}
	if rows[0]["account_id"] != "11" || rows[0]["signed_balance"] != "2000.00" || rows[0]["hidden"] != true || rows[0]["outstanding_balance"] != nil {
		t.Fatal("asset signed balance", rows[0])
	}
	if rows[2]["type"] != "liability" || rows[2]["signed_balance"] != "-300.00" || rows[2]["outstanding_balance"] != "300.00" || rows[2]["hidden"] != true {
		t.Fatal("liability semantics", rows[2])
	}
	selected := assistantCall(s, data.ToolGetAccountBalances, `{"account_ids":["10","11","14","10"],"page_size":1}`)
	if selected.MatchedCount != "2" || assistantReadBalanceTotals(t, selected)["CNY"] != expected || !reflect.DeepEqual(selected.Scope.AccountIDs, []string{"10", "11", "14"}) {
		t.Fatal("parent selection duplicates or wrong scope", selected)
	}
	empty := assistantCall(s, data.ToolGetAccountBalances, `{"account_ids":["15"]}`)
	if len(assistantMetadataRows(t, empty)) != 0 || string(empty.Totals) != "[]" {
		t.Fatal("empty parent became a leaf", empty)
	}
	if _, err := engine.ID(14).Cols("balance").Update(&models.Account{Balance: 5000}); err != nil {
		t.Fatal(err)
	}
	overpaid := assistantCall(s, data.ToolGetAccountBalances, `{"account_ids":["14"]}`)
	paidRows := assistantMetadataRows(t, overpaid)
	if paidRows[0]["signed_balance"] != "50.00" || paidRows[0]["outstanding_balance"] != "-50.00" {
		t.Fatal("overpayment sign lost", paidRows)
	}
	if assistantReadBalanceTotals(t, overpaid)["CNY"] != (assistantBalanceTotal{Currency: "CNY", AssetBalance: "0.00", LiabilityBalance: "-50.00", NetAssets: "50.00"}) {
		t.Fatal(string(overpaid.Totals))
	}
	// A cursor is not a balance snapshot: full totals are recomputed in its fresh read view.
	next := assistantCall(s, data.ToolGetAccountBalances, string(assistantJSON(map[string]string{"cursor": pages[0].NextCursor})))
	if assistantReadBalanceTotals(t, next)["CNY"].NetAssets != "2050.00" {
		t.Fatal("stale totals on continuation")
	}
}

func TestAssistantBalancesMixedInaccessibleIDsFail(t *testing.T) {
	s, _ := assistantServiceFixture(t)
	for _, args := range []string{`{"account_ids":["11","21"]}`, `{"account_ids":["10","13"]}`, `{"account_ids":["11","999999"]}`} {
		result := assistantCall(s, data.ToolGetAccountBalances, args)
		if result.OK || result.Error == nil || result.Error.Code != data.ToolErrorNotFoundOrNotAccessible || len(result.Data) != 0 || len(result.Totals) != 0 {
			t.Fatalf("partial result for inaccessible selection: %+v", result)
		}
	}
}

func TestAssistantBalancesExactBeyondInt64AndMinimumLiability(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	for _, account := range []*models.Account{
		{AccountId: 70, Uid: 7, Name: "Max asset 1", Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Category: models.ACCOUNT_CATEGORY_CASH, Currency: "JPY", Balance: math.MaxInt64},
		{AccountId: 71, Uid: 7, Name: "Max asset 2", Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Category: models.ACCOUNT_CATEGORY_CASH, Currency: "JPY", Balance: math.MaxInt64},
		{AccountId: 72, Uid: 7, Name: "Min liability", Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Category: models.ACCOUNT_CATEGORY_DEBT, Currency: "JPY", Balance: math.MinInt64},
	} {
		assistantInsertEntity(t, engine, account)
	}
	result := assistantCall(s, data.ToolGetAccountBalances, `{"account_ids":["70","71","72"]}`)
	rows := assistantMetadataRows(t, result)
	if rows[0]["signed_balance"] != "92233720368547758.07" || rows[2]["signed_balance"] != "-92233720368547758.08" || rows[2]["outstanding_balance"] != "92233720368547758.08" {
		t.Fatal("precision or negation overflow", rows)
	}
	expected := assistantBalanceTotal{Currency: "JPY", AssetBalance: "184467440737095516.14", LiabilityBalance: "92233720368547758.08", NetAssets: "92233720368547758.06"}
	if assistantReadBalanceTotals(t, result)["JPY"] != expected {
		t.Fatal("totals overflow", string(result.Totals))
	}
}

func TestAssistantBalancesCurrentTotalsIndependentOfPageHighWater(t *testing.T) {
	s, engine := assistantServiceFixture(t)
	first := assistantCall(s, data.ToolGetAccountBalances, `{"page_size":1}`)
	assistantMetadataRows(t, first)
	assistantInsertEntity(t, engine, &models.Account{AccountId: 90, Uid: 7, Name: "New asset", Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, Category: models.ACCOUNT_CATEGORY_CASH, Currency: "CNY", Balance: 100})
	next := assistantCall(s, data.ToolGetAccountBalances, string(assistantJSON(map[string]string{"cursor": first.NextCursor})))
	rows := assistantMetadataRows(t, next)
	if len(rows) != 1 || rows[0]["account_id"] != "12" || next.HasMore {
		t.Fatal("high-water leaked a later account", rows)
	}
	if assistantReadBalanceTotals(t, next)["CNY"].NetAssets != "2001.00" {
		t.Fatal("current full totals constrained by detail high-water")
	}
}
