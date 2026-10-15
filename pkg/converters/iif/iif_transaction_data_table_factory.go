package iif

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"strings"
)

func createNewIIfTransactionDataTable(ctx core.Context, accountDatasets []*iifAccountDataset, transactionDatasets []*iifTransactionDataset) (*iifTransactionDataTable, error) {
	if len(transactionDatasets) < 1 {
		return nil, errs.ErrNotFoundTransactionDataInFile
	}

	incomeAccountNames, expenseAccountNames := getIncomeAndExpenseAccountNameMap(accountDatasets)

	for i := 0; i < len(transactionDatasets); i++ {
		transactionDataset := transactionDatasets[i]

		for _, requiredColumnName := range []string{
			iifTransactionDateColumnName,
			iifTransactionAccountNameColumnName,
			iifTransactionAmountColumnName,
		} {
			if _, exists := transactionDataset.TransactionDataColumnIndexes[requiredColumnName]; !exists {
				return nil, errs.ErrMissingRequiredFieldInHeaderRow
			}
		}
	}

	return &iifTransactionDataTable{
		incomeAccountNames:  incomeAccountNames,
		expenseAccountNames: expenseAccountNames,
		transactionDatasets: transactionDatasets,
	}, nil
}

func getIncomeAndExpenseAccountNameMap(accountDatasets []*iifAccountDataset) (incomeAccountNames map[string]bool, expenseAccountNames map[string]bool) {
	incomeAccountNames = make(map[string]bool)
	expenseAccountNames = make(map[string]bool)

	for i := 0; i < len(accountDatasets); i++ {
		accountDataset := accountDatasets[i]
		accountNameColumnIndex, accountNameColumnExists := accountDataset.accountDataColumnIndexes[iifAccountNameColumnName]
		accountTypeColumnIndex, accountTypeColumnExists := accountDataset.accountDataColumnIndexes[iifAccountTypeColumnName]

		if !accountNameColumnExists || accountNameColumnIndex < 0 ||
			!accountTypeColumnExists || accountTypeColumnIndex < 0 {
			continue
		}

		for j := 0; j < len(accountDataset.accounts); j++ {
			items := accountDataset.accounts[j].dataItems

			if accountNameColumnIndex >= len(items) ||
				accountTypeColumnIndex >= len(items) {
				continue
			}

			accountName := items[accountNameColumnIndex]
			accountType := items[accountTypeColumnIndex]

			if accountType == iifAccountTypeIncome {
				incomeAccountNames[accountName] = true
			} else if accountType == iifAccountTypeExpense {
				expenseAccountNames[accountName] = true
			}
		}
	}

	return incomeAccountNames, expenseAccountNames
}

func parseAmount(amount string) (int64, error) {
	return utils.ParseAmount(strings.ReplaceAll(amount, ",", ""))
}
