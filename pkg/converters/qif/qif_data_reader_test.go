package qif

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQifDataReaderParse(t *testing.T) {
	reader := &qifDataReader{
		allLines: []string{
			"!Type:Bank",
			"D2024/10/9",
			"T-123.45",
			"^",
			"D2024/10/12",
			"T+234.56",
			"^",
			"!Type:Cash",
			"D2024/9/1",
			"T100.00",
			"POpening Balance",
			"L[Wallet]",
			"^",
			"!Type:Memorized",
			"KC",
			"T-123.45",
			"12024/10/13",
			"23",
			"31",
			"42",
			"512.34",
			"6100.45",
			"7234.56",
			"^",
			"!Type:Invst",
			"D2024/10/14",
			"NBuy",
			"YTest",
			"I12.34",
			"Q10",
			"T-123.4",
			"^",
			"!Account",
			"NTest Account",
			"^",
			"NWallet",
			"^",
			"!Type:Cat",
			"NTest Category",
			"I",
			"^",
			"!Type:Class",
			"NTest Class",
			"DFoo Bar",
			"^",
		},
	}
	context := core.NewNullContext()

	actualData, err := reader.read(context)
	assert.Nil(t, err)

	assert.Equal(t, 2, len(actualData.BankAccountTransactions))
	assert.Equal(t, "2024/10/9", actualData.BankAccountTransactions[0].Date)
	assert.Equal(t, "-123.45", actualData.BankAccountTransactions[0].Amount)
	assert.Equal(t, "2024/10/12", actualData.BankAccountTransactions[1].Date)
	assert.Equal(t, "+234.56", actualData.BankAccountTransactions[1].Amount)

	assert.Equal(t, 1, len(actualData.CashAccountTransactions))
	assert.Equal(t, "2024/9/1", actualData.CashAccountTransactions[0].Date)
	assert.Equal(t, "100.00", actualData.CashAccountTransactions[0].Amount)
	assert.Equal(t, "Opening Balance", actualData.CashAccountTransactions[0].Payee)
	assert.Equal(t, "[Wallet]", actualData.CashAccountTransactions[0].Category)

	assert.Equal(t, 1, len(actualData.MemorizedTransactions))
	assert.Equal(t, qifCheckTransactionType, actualData.MemorizedTransactions[0].TransactionType)
	assert.Equal(t, "-123.45", actualData.MemorizedTransactions[0].Amount)
	assert.Equal(t, "2024/10/13", actualData.MemorizedTransactions[0].Amortization.FirstPaymentDate)
	assert.Equal(t, "3", actualData.MemorizedTransactions[0].Amortization.TotalYearsForLoan)
	assert.Equal(t, "1", actualData.MemorizedTransactions[0].Amortization.NumberOfPayments)
	assert.Equal(t, "2", actualData.MemorizedTransactions[0].Amortization.NumberOfPeriodsPerYear)
	assert.Equal(t, "12.34", actualData.MemorizedTransactions[0].Amortization.InterestRate)
	assert.Equal(t, "100.45", actualData.MemorizedTransactions[0].Amortization.CurrentLoanBalance)
	assert.Equal(t, "234.56", actualData.MemorizedTransactions[0].Amortization.OriginalLoanAmount)

	assert.Equal(t, 1, len(actualData.InvestmentAccountTransactions))
	assert.Equal(t, "2024/10/14", actualData.InvestmentAccountTransactions[0].Date)
	assert.Equal(t, "Buy", actualData.InvestmentAccountTransactions[0].Action)
	assert.Equal(t, "Test", actualData.InvestmentAccountTransactions[0].Security)
	assert.Equal(t, "12.34", actualData.InvestmentAccountTransactions[0].Price)
	assert.Equal(t, "10", actualData.InvestmentAccountTransactions[0].Quantity)
	assert.Equal(t, "-123.4", actualData.InvestmentAccountTransactions[0].Amount)

	assert.Equal(t, 2, len(actualData.Accounts))
	assert.Equal(t, "Test Account", actualData.Accounts[0].Name)
	assert.Equal(t, "Wallet", actualData.Accounts[1].Name)

	assert.Equal(t, 1, len(actualData.Categories))
	assert.Equal(t, "Test Category", actualData.Categories[0].Name)
	assert.Equal(t, qifIncomeTransaction, actualData.Categories[0].CategoryType)

	assert.Equal(t, 1, len(actualData.Classes))
	assert.Equal(t, "Test Class", actualData.Classes[0].Name)
	assert.Equal(t, "Foo Bar", actualData.Classes[0].Description)
}

func TestQifDataReaderParse_AccountEntryBeforeTransaction(t *testing.T) {
	reader := &qifDataReader{
		allLines: []string{
			"!Account",
			"NTest Account",
			"^",
			"!Type:Bank",
			"D2024/10/9",
			"T-123.45",
			"^",
			"D2024/10/12",
			"T+234.56",
			"^",
			"!Account",
			"NWallet",
			"^",
			"!Type:Cash",
			"D2024/9/1",
			"T100.00",
			"POpening Balance",
			"L[Wallet]",
			"^",
		},
	}
	context := core.NewNullContext()

	actualData, err := reader.read(context)
	assert.Nil(t, err)

	assert.Equal(t, 2, len(actualData.BankAccountTransactions))
	assert.Equal(t, "2024/10/9", actualData.BankAccountTransactions[0].Date)
	assert.Equal(t, "-123.45", actualData.BankAccountTransactions[0].Amount)
	assert.Equal(t, "2024/10/12", actualData.BankAccountTransactions[1].Date)
	assert.Equal(t, "+234.56", actualData.BankAccountTransactions[1].Amount)

	assert.Equal(t, 1, len(actualData.CashAccountTransactions))
	assert.Equal(t, "2024/9/1", actualData.CashAccountTransactions[0].Date)
	assert.Equal(t, "100.00", actualData.CashAccountTransactions[0].Amount)
	assert.Equal(t, "Opening Balance", actualData.CashAccountTransactions[0].Payee)
	assert.Equal(t, "[Wallet]", actualData.CashAccountTransactions[0].Category)

	assert.Equal(t, 2, len(actualData.Accounts))
	assert.Equal(t, "Test Account", actualData.Accounts[0].Name)
	assert.Equal(t, "Wallet", actualData.Accounts[1].Name)
}

func TestQifDataReaderParse_EmptyContent(t *testing.T) {
	reader := &qifDataReader{
		allLines: []string{},
	}
	context := core.NewNullContext()

	_, err := reader.read(context)
	assert.EqualError(t, err, errs.ErrNotFoundTransactionDataInFile.Message)
}

func TestQifDataReaderParse_EmptyEntry(t *testing.T) {
	reader := &qifDataReader{
		allLines: []string{
			"!Type:Bank",
			"^",
			"!Type:Cash",
			"^",
			"!Type:Memorized",
			"^",
			"!Type:Invst",
			"^",
			"!Account",
			"^",
			"!Type:Cat",
			"^",
			"!Type:Class",
			"^",
		},
	}
	context := core.NewNullContext()

	actualData, err := reader.read(context)
	assert.Nil(t, err)

	assert.Equal(t, 0, len(actualData.BankAccountTransactions))
	assert.Equal(t, 0, len(actualData.CashAccountTransactions))
	assert.Equal(t, 0, len(actualData.MemorizedTransactions))
	assert.Equal(t, 0, len(actualData.InvestmentAccountTransactions))
	assert.Equal(t, 0, len(actualData.Accounts))
	assert.Equal(t, 0, len(actualData.Categories))
	assert.Equal(t, 0, len(actualData.Classes))
}

func TestQifDataReaderParse_UnsupportedEntryHeader(t *testing.T) {
	reader := &qifDataReader{
		allLines: []string{
			"!Type:Bank",
			"D2024/10/9",
			"T-123.45",
			"^",
			"!Type:Unknown",
			"D2024/10/11",
			"T100.00",
			"^",
		},
	}
	context := core.NewNullContext()

	actualData, err := reader.read(context)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(actualData.BankAccountTransactions))
	assert.Equal(t, "2024/10/9", actualData.BankAccountTransactions[0].Date)
	assert.Equal(t, "-123.45", actualData.BankAccountTransactions[0].Amount)
}

func TestQifDataReaderParse_UnsupportedLine(t *testing.T) {
	reader := &qifDataReader{
		allLines: []string{
			"!Type:Bank",
			"D2024/10/9",
			"T-123.45",
			"^",
			"!Option:Unknown",
			"D2024/10/11",
			"T100.00",
			"^",
		},
	}
	context := core.NewNullContext()

	actualData, err := reader.read(context)
	assert.Nil(t, err)

	assert.Equal(t, 2, len(actualData.BankAccountTransactions))
	assert.Equal(t, "2024/10/9", actualData.BankAccountTransactions[0].Date)
	assert.Equal(t, "-123.45", actualData.BankAccountTransactions[0].Amount)
	assert.Equal(t, "2024/10/11", actualData.BankAccountTransactions[1].Date)
	assert.Equal(t, "100.00", actualData.BankAccountTransactions[1].Amount)
}

func TestQifDataReaderParse_NewEntryHeaderAfterUnclosedEntry(t *testing.T) {
	reader := &qifDataReader{
		allLines: []string{
			"!Type:Bank",
			"D2024/10/9",
			"T-123.45",
			"!Type:Cash",
			"D2024/9/1",
			"T100.00",
			"POpening Balance",
			"L[Wallet]",
			"^",
		},
	}
	context := core.NewNullContext()

	_, err := reader.read(context)
	assert.EqualError(t, err, errs.ErrInvalidQIFFile.Message)
}
