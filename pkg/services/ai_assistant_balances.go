package services

import (
	"math/big"
	"sort"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

type assistantAccountBalance struct {
	AccountID          string  `json:"account_id"`
	Name               string  `json:"name"`
	ParentID           string  `json:"parent_id"`
	Type               string  `json:"type"`
	Currency           string  `json:"currency"`
	Hidden             bool    `json:"hidden"`
	SignedBalance      string  `json:"signed_balance"`
	OutstandingBalance *string `json:"outstanding_balance"`
}
type assistantBalanceTotal struct {
	Currency         string `json:"currency"`
	AssetBalance     string `json:"asset_balance"`
	LiabilityBalance string `json:"liability_balance"`
	NetAssets        string `json:"net_assets"`
}
type assistantBalanceAccumulator struct{ assets, liabilities, net big.Int }

func assistantAccountBalances(e *assistantQueryExecution) (data.ToolResult, error) {
	selected := make(map[int64]bool, len(e.AccountIDs))
	for _, value := range e.AccountIDs {
		id, _ := assistantID(value)
		selected[id] = true
	}
	parents := make(map[int64]bool)
	for _, account := range e.Metadata.Accounts {
		if account.ParentAccountId != 0 {
			parents[account.ParentAccountId] = true
		}
	}
	rows := make([]assistantEntityRow, 0, len(e.Metadata.Accounts))
	sums := make(map[string]*assistantBalanceAccumulator)
	for id, account := range e.Metadata.Accounts {
		if err := e.Context.Err(); err != nil {
			return data.ToolResult{}, err
		}
		// An empty multi-account container is still a parent, never an asset leaf.
		if account.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS || parents[id] || len(selected) > 0 && !selected[id] {
			continue
		}
		if account.Type != models.ACCOUNT_TYPE_SINGLE_ACCOUNT || (!account.Category.IsAsset() && !account.Category.IsLiability()) {
			return data.ToolResult{}, assistantInconsistent()
		}
		signed := big.NewInt(account.Balance)
		row := assistantAccountBalance{AccountID: assistantCount(id), Name: account.Name, ParentID: assistantCount(account.ParentAccountId), Type: "asset", Currency: account.Currency, Hidden: account.Hidden, SignedBalance: assistantFormatAmount(signed)}
		sum := sums[account.Currency]
		if sum == nil {
			sum = &assistantBalanceAccumulator{}
			sums[account.Currency] = sum
		}
		sum.net.Add(&sum.net, signed)
		if account.Category.IsLiability() {
			row.Type = "liability"
			outstanding := new(big.Int).Neg(signed)
			formatted := assistantFormatAmount(outstanding)
			row.OutstandingBalance = &formatted
			sum.liabilities.Add(&sum.liabilities, outstanding)
		} else {
			sum.assets.Add(&sum.assets, signed)
		}
		rows = append(rows, assistantEntityRow{id, row})
	}
	currencies := make([]string, 0, len(sums))
	for currency := range sums {
		currencies = append(currencies, currency)
	}
	sort.Strings(currencies)
	totals := make([]assistantBalanceTotal, 0, len(currencies))
	for _, currency := range currencies {
		sum := sums[currency]
		totals = append(totals, assistantBalanceTotal{currency, assistantFormatAmount(&sum.assets), assistantFormatAmount(&sum.liabilities), assistantFormatAmount(&sum.net)})
	}
	return assistantEntityPage(e, rows, data.ToolResult{Totals: assistantJSON(totals), TotalsCoverage: data.QueryCoverageAllMatching})
}
