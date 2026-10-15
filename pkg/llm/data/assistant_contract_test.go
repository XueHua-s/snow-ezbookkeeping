package data_test

import (
	"encoding/json"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
)

func TestToolResultPreservesExactFinancialValuesAndCoverage(t *testing.T) {
	no := false
	result := data.ToolResult{
		OK: true, QueryID: "q1", Scope: &data.QueryScope{Timezone: "Asia/Shanghai", AccountIDs: []string{"9007199254740993"}, IncludesHidden: true, ExcludesDeleted: true},
		MatchedCount: "9007199254740993", ReturnedCount: 1, Coverage: data.QueryCoveragePage,
		Data:           json.RawMessage(`[{"transaction_id":"9007199254740993","amount":"900719925474099300.00"}]`),
		Totals:         json.RawMessage(`[{"currency":"CNY","sum":"900719925474099300.00","min":null,"max":null}]`),
		TotalsCoverage: data.QueryCoverageAllMatching, GroupsHasMore: &no, CountsAdditive: &no, Warnings: []string{},
		Comparison: json.RawMessage(`[{"percentage":null,"reason":"non_positive_baseline"}]`),
		References: []data.TransactionReference{{ID: "PRIVATE_REFERENCE"}},
		Metrics:    data.ToolExecutionMetrics{RowsRead: 99, Duration: time.Second},
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["matched_count"] != "9007199254740993" || decoded["counts_additive"] != false || decoded["groups_has_more"] != false || decoded["totals_coverage"] != "all_matching" {
		t.Fatalf("lost result contract: %s", encoded)
	}
	row := decoded["data"].([]any)[0].(map[string]any)
	if row["transaction_id"] != "9007199254740993" || row["amount"] != "900719925474099300.00" {
		t.Fatalf("lost precision: %s", encoded)
	}
	scope := decoded["scope"].(map[string]any)
	if scope["start_time"] != nil || scope["end_time"] != nil {
		t.Fatalf("unbounded times must be null: %s", encoded)
	}
	for _, forbidden := range []string{"PRIVATE_REFERENCE", "RowsRead", "Metrics", "references"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("internal side channel serialized: %s", encoded)
		}
	}
	var roundTrip data.ToolResult
	if err := json.Unmarshal(encoded, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if string(roundTrip.Comparison) != string(result.Comparison) {
		t.Fatal("comparison projection lost")
	}
}

func TestToolFailureIsNotAnEmptySuccess(t *testing.T) {
	encoded, err := json.Marshal(data.ToolResult{Error: &data.ToolError{Code: data.ToolErrorNotFoundOrNotAccessible, Message: "Entity is not accessible", Retryable: false}})
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(encoded, &result); err != nil {
		t.Fatal(err)
	}
	if result["ok"] != false || result["error"].(map[string]any)["code"] != "NOT_FOUND_OR_NOT_ACCESSIBLE" {
		t.Fatalf("invalid failure envelope: %s", encoded)
	}
	if _, exists := result["matched_count"]; exists {
		t.Fatal("failed query must not invent an empty matched count")
	}
}

func TestToolTurnKeepsPairingWithoutSerializingPrivateState(t *testing.T) {
	call := data.ToolCall{ID: "call-1", Name: data.ToolQueryTransactions, Arguments: json.RawMessage(`{"filter":{"keyword":"literal","keyword":"duplicate"}}`)}
	input := data.ToolTurnInput{Messages: []data.ConversationMessage{
		{Role: data.ConversationRoleAssistant, ToolCalls: []data.ToolCall{call}},
		{Role: data.ConversationRoleTool, ToolCallID: call.ID, ToolResult: &data.ToolResult{OK: true}},
	}, State: map[string]string{"secret": "PRIVATE_PROVIDER_STATE"}}
	encoded, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "PRIVATE_PROVIDER_STATE") {
		t.Fatal("private state escaped")
	}
	var decoded data.ToolTurnInput
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Messages[0].ToolCalls[0].ID != decoded.Messages[1].ToolCallID {
		t.Fatal("lost call/result pair")
	}
	if !strings.Contains(string(decoded.Messages[0].ToolCalls[0].Arguments), `"keyword":"literal","keyword":"duplicate"`) {
		t.Fatal("raw arguments lost duplicate-key evidence before validation")
	}
	response, err := json.Marshal(data.TurnResponse{State: "PRIVATE_PROVIDER_STATE", Completed: true})
	if err != nil || strings.Contains(string(response), "PRIVATE_PROVIDER_STATE") {
		t.Fatalf("private output state escaped: %s, %v", response, err)
	}
}

func TestIndependentFinancialTruthFixture(t *testing.T) {
	// This oracle uses only fixture fields and big.Int; it deliberately imports
	// no production transaction/aggregation/conversion service.
	raw, err := os.ReadFile("../../services/testdata/ai_assistant_truth.json")
	if err != nil {
		t.Fatal(err)
	}
	type entry struct {
		ID                  string `json:"id"`
		Type                string `json:"type"`
		Currency            string `json:"currency"`
		Amount              string `json:"amount"`
		Minor               string `json:"amount_minor"`
		DestinationCurrency string `json:"destination_currency"`
		DestinationMinor    string `json:"destination_amount_minor"`
		RelatedID           string `json:"related_id"`
	}
	type total struct{ Currency, Flow, Count, Sum, Min, Max string }
	var fixture struct {
		Transactions []entry `json:"transactions"`
		TransferIn   entry   `json:"transfer_in_storage_record"`
		Expected     struct {
			Count        string                              `json:"logical_count"`
			StorageCount string                              `json:"storage_count"`
			Totals       []total                             `json:"totals"`
			NetIncome    []struct{ Currency, Amount string } `json:"net_income"`
		} `json:"expected"`
		Balances struct {
			Accounts []struct {
				Type          string
				SignedBalance string `json:"signed_balance"`
			} `json:"accounts"`
			Expected struct {
				NetAssets          string `json:"net_assets"`
				OutstandingBalance string `json:"outstanding_balance"`
			} `json:"expected"`
		} `json:"balance_fixture"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Transactions) != 6 || fixture.Expected.Count != "6" || fixture.Expected.StorageCount != "7" {
		t.Fatal("truth fixture must contain six logical / seven storage transactions")
	}
	type stats struct {
		count         int64
		sum, min, max *big.Int
	}
	byFlow := map[string]*stats{}
	add := func(currency, flow, minor string) {
		n, ok := new(big.Int).SetString(minor, 10)
		if !ok {
			t.Fatalf("invalid minor units: %q", minor)
		}
		key := currency + "/" + flow
		st := byFlow[key]
		if st == nil {
			st = &stats{sum: new(big.Int), min: new(big.Int).Set(n), max: new(big.Int).Set(n)}
			byFlow[key] = st
		}
		st.count++
		st.sum.Add(st.sum, n)
		if n.Cmp(st.min) < 0 {
			st.min.Set(n)
		}
		if n.Cmp(st.max) > 0 {
			st.max.Set(n)
		}
	}
	for _, tx := range fixture.Transactions {
		rational, ok := new(big.Rat).SetString(tx.Amount)
		if !ok {
			t.Fatal(tx.Amount)
		}
		rational.Mul(rational, big.NewRat(100, 1))
		if !rational.IsInt() || rational.Num().String() != tx.Minor {
			t.Fatal("fixture decimal/minor mismatch")
		}
		if tx.Type == "transfer" {
			if tx.RelatedID != fixture.TransferIn.ID || fixture.TransferIn.RelatedID != tx.ID {
				t.Fatal("transfer pair is not reciprocal")
			}
			add(tx.Currency, "transfer_out", tx.Minor)
			add(tx.DestinationCurrency, "transfer_in", tx.DestinationMinor)
		} else {
			add(tx.Currency, tx.Type, tx.Minor)
		}
	}
	fixed := func(n *big.Int) string { return new(big.Rat).SetFrac(n, big.NewInt(100)).FloatString(2) }
	for _, want := range fixture.Expected.Totals {
		got := byFlow[want.Currency+"/"+want.Flow]
		if got == nil || big.NewInt(got.count).String() != want.Count || fixed(got.sum) != want.Sum || fixed(got.min) != want.Min || fixed(got.max) != want.Max {
			t.Fatalf("independent total differs: %+v, got %+v", want, got)
		}
	}
	net := new(big.Int).Sub(byFlow["CNY/income"].sum, byFlow["CNY/expense"].sum)
	if fixed(net) != "890.00" || fixture.Expected.NetIncome[0].Amount != "890.00" {
		t.Fatal("net income must be 890.00")
	}
	balance := new(big.Rat)
	for _, account := range fixture.Balances.Accounts {
		if account.Type == "parent" {
			continue
		}
		n, ok := new(big.Rat).SetString(account.SignedBalance)
		if !ok {
			t.Fatal(account.SignedBalance)
		}
		balance.Add(balance, n)
		if account.Type == "liability" && new(big.Rat).Neg(n).FloatString(2) != fixture.Balances.Expected.OutstandingBalance {
			t.Fatal("outstanding balance must negate signed liability")
		}
	}
	if balance.FloatString(2) != "1700.00" || fixture.Balances.Expected.NetAssets != "1700.00" {
		t.Fatal("net assets must be 1700.00, excluding parent")
	}
}
