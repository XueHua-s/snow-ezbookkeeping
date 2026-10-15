package iif

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/converters/datatable"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

const iifAccountNameColumnName = "NAME"

const iifAccountTypeColumnName = "ACCNTTYPE"

const iifAccountTypeIncome = "INC"

const iifAccountTypeExpense = "EXP"

const iifTransactionTypeColumnName = "TRNSTYPE"

const iifTransactionDateColumnName = "DATE"

const iifTransactionAccountNameColumnName = "ACCNT"

const iifTransactionNameColumnName = "NAME"

const iifTransactionAmountColumnName = "AMOUNT"

const iifTransactionMemoColumnName = "MEMO"

const iifTransactionTypeBeginningBalance = "BEGINBALCHECK"

const iifTransactionCategorySeparator = ":"

var iifTransactionSupportedColumns = map[datatable.TransactionDataTableColumn]bool{
	datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME:     true,
	datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE:     true,
	datatable.TRANSACTION_DATA_TABLE_CATEGORY:             true,
	datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY:         true,
	datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME:         true,
	datatable.TRANSACTION_DATA_TABLE_AMOUNT:               true,
	datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME: true,
	datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT:       true,
	datatable.TRANSACTION_DATA_TABLE_DESCRIPTION:          true,
}

// iifTransactionDataTable defines the structure of intuit interchange format (iif) transaction data table
type iifTransactionDataTable struct {
	incomeAccountNames  map[string]bool
	expenseAccountNames map[string]bool
	transactionDatasets []*iifTransactionDataset
}

// iifTransactionDataRow defines the structure of intuit interchange format (iif) transaction data row
type iifTransactionDataRow struct {
	dataTable  *iifTransactionDataTable
	finalItems map[datatable.TransactionDataTableColumn]string
}

// iifTransactionDataRowIterator defines the structure of intuit interchange format (iif) transaction data row iterator
type iifTransactionDataRowIterator struct {
	dataTable             *iifTransactionDataTable
	currentDatasetIndex   int
	currentIndexInDataset int
	currentSplitDataIndex int
}

// HasColumn returns whether the transaction data table has specified column
func (t *iifTransactionDataTable) HasColumn(column datatable.TransactionDataTableColumn) bool {
	_, exists := iifTransactionSupportedColumns[column]
	return exists
}

// TransactionRowCount returns the total count of transaction data row
func (t *iifTransactionDataTable) TransactionRowCount() int {
	totalDataRowCount := 0

	for i := 0; i < len(t.transactionDatasets); i++ {
		datasets := t.transactionDatasets[i]

		for j := 0; j < len(datasets.Transactions); j++ {
			transaction := datasets.Transactions[j]

			if transaction.SplitData != nil {
				totalDataRowCount += len(transaction.SplitData)
			}
		}
	}

	return totalDataRowCount
}

// TransactionRowIterator returns the iterator of transaction data row
func (t *iifTransactionDataTable) TransactionRowIterator() datatable.TransactionDataRowIterator {
	return &iifTransactionDataRowIterator{
		dataTable:             t,
		currentDatasetIndex:   0,
		currentIndexInDataset: 0,
		currentSplitDataIndex: -1,
	}
}

// IsValid returns whether this row is valid data for importing
func (r *iifTransactionDataRow) IsValid() bool {
	return true
}

// GetData returns the data in the specified column type
func (r *iifTransactionDataRow) GetData(column datatable.TransactionDataTableColumn) string {
	_, exists := iifTransactionSupportedColumns[column]

	if exists {
		return r.finalItems[column]
	}

	return ""
}

// HasNext returns whether the iterator does not reach the end
func (t *iifTransactionDataRowIterator) HasNext() bool {
	allDatasets := t.dataTable.transactionDatasets

	if t.currentDatasetIndex >= len(allDatasets) {
		return false
	}

	currentDataset := allDatasets[t.currentDatasetIndex]

	if t.currentIndexInDataset+1 < len(currentDataset.Transactions) {
		return true
	} else if t.currentIndexInDataset < len(currentDataset.Transactions) &&
		t.currentSplitDataIndex+1 < len(currentDataset.Transactions[t.currentIndexInDataset].SplitData) {
		return true
	}

	for i := t.currentDatasetIndex + 1; i < len(allDatasets); i++ {
		dataset := allDatasets[i]

		if len(dataset.Transactions) < 1 {
			continue
		}

		return true
	}

	return false
}

// CurrentRowId returns current row id
func (t *iifTransactionDataRowIterator) CurrentRowId() string {
	if t.currentDatasetIndex < 0 || t.currentDatasetIndex >= len(t.dataTable.transactionDatasets) {
		return ""
	}

	dataset := t.dataTable.transactionDatasets[t.currentDatasetIndex]

	if dataset == nil || t.currentIndexInDataset < 0 || t.currentIndexInDataset >= len(dataset.Transactions) {
		return ""
	}

	transaction := dataset.Transactions[t.currentIndexInDataset]

	if transaction == nil || t.currentSplitDataIndex < 0 || t.currentSplitDataIndex >= len(transaction.SplitData) {
		return ""
	}

	return fmt.Sprintf("dataset:%d-record:%d-split:%d", t.currentDatasetIndex+1, t.currentIndexInDataset+1, t.currentSplitDataIndex+1)
}

// Next returns the next transaction data row
func (t *iifTransactionDataRowIterator) Next(ctx core.Context, user *models.User) (daraRow datatable.TransactionDataRow, err error) {
	allDatasets := t.dataTable.transactionDatasets

	for i := t.currentDatasetIndex; i < len(allDatasets); i++ {
		foundNextRow := false
		dataset := allDatasets[i]

		for j := t.currentIndexInDataset; j < len(dataset.Transactions); j++ {
			if t.currentSplitDataIndex+1 < len(dataset.Transactions[j].SplitData) {
				t.currentSplitDataIndex++
				foundNextRow = true
				break
			}

			t.currentIndexInDataset++
			t.currentSplitDataIndex = -1
		}

		if foundNextRow {
			break
		}

		t.currentDatasetIndex++
		t.currentIndexInDataset = 0
		t.currentSplitDataIndex = -1
	}

	if t.currentDatasetIndex >= len(allDatasets) {
		return nil, nil
	}

	currentDataset := allDatasets[t.currentDatasetIndex]

	if t.currentIndexInDataset >= len(currentDataset.Transactions) {
		return nil, nil
	}

	data := currentDataset.Transactions[t.currentIndexInDataset]

	if len(data.SplitData) < 1 {
		log.Errorf(ctx, "[iif_transaction_data_table.Next] cannot parsing transaction in %s, because split data is empty", t.CurrentRowId())
		return nil, errs.NewImportError(errs.ErrInvalidIIFFile, t.CurrentRowId())
	}

	if t.currentSplitDataIndex >= len(data.SplitData) {
		return nil, nil
	}

	if len(data.SplitData) > 1 {
		_, err := t.isSplitTransactionSupported(ctx, currentDataset, data)

		if err != nil {
			return nil, err
		}
	}

	rowItems, err := t.parseTransaction(ctx, user, currentDataset, data, t.currentSplitDataIndex)

	if err != nil {
		log.Errorf(ctx, "[iif_transaction_data_table.Next] cannot parsing transaction in %s, because %s", t.CurrentRowId(), err.Error())
		return nil, err
	}

	return &iifTransactionDataRow{
		dataTable:  t.dataTable,
		finalItems: rowItems,
	}, nil
}
