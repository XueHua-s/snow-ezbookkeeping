package iif

import (
	"github.com/mayswind/ezbookkeeping/pkg/converters/datatable"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"strings"
)

func (t *iifTransactionDataRowIterator) parseTransaction(ctx core.Context, user *models.User, dataset *iifTransactionDataset, transactionData *iifTransactionData, splitDataIndex int) (map[datatable.TransactionDataTableColumn]string, error) {
	var err error

	data := make(map[datatable.TransactionDataTableColumn]string, len(iifTransactionSupportedColumns))
	data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME], err = t.parseTransactionTime(dataset, transactionData)

	if err != nil {
		return nil, err
	}

	transactionType, _ := dataset.getSplitDataItemValue(transactionData.SplitData[splitDataIndex], iifTransactionTypeColumnName)
	mainAccountName, _ := dataset.getTransactionDataItemValue(transactionData, iifTransactionAccountNameColumnName)
	splitAccountName, _ := dataset.getSplitDataItemValue(transactionData.SplitData[splitDataIndex], iifTransactionAccountNameColumnName)
	mainAmount, _ := dataset.getTransactionDataItemValue(transactionData, iifTransactionAmountColumnName)
	splitAmount, _ := dataset.getSplitDataItemValue(transactionData.SplitData[splitDataIndex], iifTransactionAmountColumnName)
	mainAmountNum, err := parseAmount(mainAmount)

	if err != nil {
		return nil, errs.NewImportError(errs.ErrAmountInvalid, t.CurrentRowId(), mainAmount)
	}

	splitAmountNum, err := parseAmount(splitAmount)

	if err != nil {
		return nil, errs.NewImportError(errs.ErrAmountInvalid, t.CurrentRowId(), splitAmount)
	}

	if transactionType == iifTransactionTypeBeginningBalance { // balance modification
		data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE] = iifTransactionTypeNameMapping[models.TRANSACTION_TYPE_MODIFY_BALANCE]
		data[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME] = mainAccountName
		data[datatable.TRANSACTION_DATA_TABLE_AMOUNT] = utils.FormatAmount(mainAmountNum)
	} else if (t.dataTable.incomeAccountNames[mainAccountName] && !t.dataTable.incomeAccountNames[splitAccountName] && !t.dataTable.expenseAccountNames[splitAccountName]) ||
		(t.dataTable.incomeAccountNames[splitAccountName] && !t.dataTable.incomeAccountNames[mainAccountName] && !t.dataTable.expenseAccountNames[mainAccountName]) { // income
		data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE] = iifTransactionTypeNameMapping[models.TRANSACTION_TYPE_INCOME]
		categoryName := ""
		accountName := ""
		amountNum := int64(0)

		if t.dataTable.incomeAccountNames[mainAccountName] && !t.dataTable.incomeAccountNames[splitAccountName] {
			categoryName = mainAccountName
			accountName = splitAccountName

			if len(transactionData.SplitData) > 1 {
				amountNum = splitAmountNum
			} else {
				amountNum = -mainAmountNum
			}
		} else if t.dataTable.incomeAccountNames[splitAccountName] && !t.dataTable.incomeAccountNames[mainAccountName] {
			categoryName = splitAccountName
			accountName = mainAccountName

			if len(transactionData.SplitData) > 1 {
				amountNum = -splitAmountNum
			} else {
				amountNum = mainAmountNum
			}
		} else {
			log.Errorf(ctx, "[iif_transaction_data_table.parseTransaction] cannot parse transaction, because main account \"%s\" and split account \"%s\" are all income account", mainAccountName, splitAccountName)
			return nil, errs.NewImportError(errs.ErrInvalidIIFFile, t.CurrentRowId())
		}

		categoryNames := strings.Split(categoryName, iifTransactionCategorySeparator)

		if len(categoryNames) > 1 {
			data[datatable.TRANSACTION_DATA_TABLE_CATEGORY] = categoryNames[0]
			data[datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY] = categoryNames[len(categoryNames)-1]
		} else {
			data[datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY] = categoryName
		}

		data[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME] = accountName
		data[datatable.TRANSACTION_DATA_TABLE_AMOUNT] = utils.FormatAmount(amountNum)
	} else if (t.dataTable.expenseAccountNames[mainAccountName] && !t.dataTable.expenseAccountNames[splitAccountName] && !t.dataTable.incomeAccountNames[splitAccountName]) ||
		(t.dataTable.expenseAccountNames[splitAccountName] && !t.dataTable.expenseAccountNames[mainAccountName] && !t.dataTable.incomeAccountNames[mainAccountName]) { // expense
		data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE] = iifTransactionTypeNameMapping[models.TRANSACTION_TYPE_EXPENSE]
		categoryName := ""
		accountName := ""
		amountNum := int64(0)

		if t.dataTable.expenseAccountNames[mainAccountName] && !t.dataTable.expenseAccountNames[splitAccountName] {
			categoryName = mainAccountName
			accountName = splitAccountName

			if len(transactionData.SplitData) > 1 {
				amountNum = -splitAmountNum
			} else {
				amountNum = mainAmountNum
			}
		} else if t.dataTable.expenseAccountNames[splitAccountName] && !t.dataTable.expenseAccountNames[mainAccountName] {
			categoryName = splitAccountName
			accountName = mainAccountName

			if len(transactionData.SplitData) > 1 {
				amountNum = splitAmountNum
			} else {
				amountNum = -mainAmountNum
			}
		} else {
			log.Errorf(ctx, "[iif_transaction_data_table.parseTransaction] cannot parse transaction, because main account \"%s\" and split account \"%s\" are all expense account", mainAccountName, splitAccountName)
			return nil, errs.NewImportError(errs.ErrInvalidIIFFile, t.CurrentRowId())
		}

		categoryNames := strings.Split(categoryName, iifTransactionCategorySeparator)

		if len(categoryNames) > 1 {
			data[datatable.TRANSACTION_DATA_TABLE_CATEGORY] = categoryNames[0]
			data[datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY] = categoryNames[len(categoryNames)-1]
		} else {
			data[datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY] = categoryName
		}

		data[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME] = accountName
		data[datatable.TRANSACTION_DATA_TABLE_AMOUNT] = utils.FormatAmount(amountNum)
	} else {
		data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE] = iifTransactionTypeNameMapping[models.TRANSACTION_TYPE_TRANSFER]
		data[datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY] = ""
		amountNum := int64(0)
		relatedAmountNum := int64(0)
		mainAccountTransferToSplitAccount := false

		if len(transactionData.SplitData) > 1 {
			amountNum = splitAmountNum
			relatedAmountNum = splitAmountNum
			mainAccountTransferToSplitAccount = amountNum >= 0
		} else {
			if mainAmountNum >= 0 {
				amountNum = splitAmountNum
				relatedAmountNum = mainAmountNum
				mainAccountTransferToSplitAccount = false
			} else if splitAmountNum >= 0 {
				amountNum = mainAmountNum
				relatedAmountNum = splitAmountNum
				mainAccountTransferToSplitAccount = true
			}
		}

		if mainAccountTransferToSplitAccount {
			data[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME] = mainAccountName
			data[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME] = splitAccountName
		} else {
			data[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME] = splitAccountName
			data[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME] = mainAccountName
		}

		if amountNum >= 0 {
			data[datatable.TRANSACTION_DATA_TABLE_AMOUNT] = utils.FormatAmount(amountNum)
		} else {
			data[datatable.TRANSACTION_DATA_TABLE_AMOUNT] = utils.FormatAmount(-amountNum)
		}

		if relatedAmountNum >= 0 {
			data[datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT] = utils.FormatAmount(relatedAmountNum)
		} else {
			data[datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT] = utils.FormatAmount(-relatedAmountNum)
		}
	}

	if splitMemo, _ := dataset.getSplitDataItemValue(transactionData.SplitData[splitDataIndex], iifTransactionMemoColumnName); splitMemo != "" {
		data[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION] = splitMemo
	} else if memo, _ := dataset.getTransactionDataItemValue(transactionData, iifTransactionMemoColumnName); memo != "" {
		data[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION] = memo
	} else if splitName, _ := dataset.getSplitDataItemValue(transactionData.SplitData[splitDataIndex], iifTransactionNameColumnName); splitName != "" {
		data[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION] = splitName
	} else if name, _ := dataset.getTransactionDataItemValue(transactionData, iifTransactionNameColumnName); name != "" {
		data[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION] = name
	} else {
		data[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION] = ""
	}

	return data, nil
}

func (t *iifTransactionDataRowIterator) isSplitTransactionSupported(ctx core.Context, dataset *iifTransactionDataset, transactionData *iifTransactionData) (bool, error) {
	supportSplitTransactions := true
	transactionType, _ := dataset.getTransactionDataItemValue(transactionData, iifTransactionTypeColumnName)

	if transactionType == iifTransactionTypeBeginningBalance { // balance modification
		supportSplitTransactions = false
		log.Errorf(ctx, "[iif_transaction_data_table.isSplitTransactionSupported] cannot parse split balance modification in %s", t.CurrentRowId())
	} else {
		transactionAmountStr, _ := dataset.getTransactionDataItemValue(transactionData, iifTransactionAmountColumnName)
		transactionAmount, err := parseAmount(transactionAmountStr)

		if err != nil {
			log.Errorf(ctx, "[iif_transaction_data_table.isSplitTransactionSupported] cannot parsing transaction in %s, because transaction amount \"%s\" is invalid", t.CurrentRowId(), transactionAmountStr)
			return false, errs.NewImportError(errs.ErrAmountInvalid, t.CurrentRowId(), transactionAmountStr)
		}

		splitTotalAmount := int64(0)

		for i := 0; i < len(transactionData.SplitData); i++ {
			splitAmountStr, _ := dataset.getSplitDataItemValue(transactionData.SplitData[i], iifTransactionAmountColumnName)
			splitAmount, err := parseAmount(splitAmountStr)

			if err != nil {
				log.Errorf(ctx, "[iif_transaction_data_table.isSplitTransactionSupported] cannot parsing transaction in %s, because split amount \"%s\" is invalid", t.CurrentRowId(), splitAmountStr)
				return false, errs.NewImportError(errs.ErrAmountInvalid, t.CurrentRowId(), splitAmountStr)
			}

			splitTotalAmount += splitAmount
		}

		if splitTotalAmount != -transactionAmount {
			supportSplitTransactions = false
			log.Errorf(ctx, "[iif_transaction_data_table.isSplitTransactionSupported] cannot parse split %s, because the sum amount of each split data \"%d\" not equal to the transaction amount \"%d\"", t.CurrentRowId(), splitTotalAmount, -transactionAmount)
		}
	}

	if len(transactionData.SplitData) > 1 && !supportSplitTransactions {
		return false, errs.NewImportError(errs.ErrNotSupportedSplitTransactions, t.CurrentRowId())
	}

	return true, nil
}

func (t *iifTransactionDataRowIterator) parseTransactionTime(dataset *iifTransactionDataset, transactionData *iifTransactionData) (string, error) {
	date, _ := dataset.getTransactionDataItemValue(transactionData, iifTransactionDateColumnName)
	dateParts := strings.Split(date, "/")

	if len(dateParts) != 3 {
		return "", errs.NewImportError(errs.ErrTransactionTimeInvalid, t.CurrentRowId(), date)
	}

	month := dateParts[0]
	day := dateParts[1]
	year := dateParts[2]

	if utils.IsValidYearMonthDayLongOrShortDateFormat(date) && !utils.IsValidMonthDayYearLongOrShortDateFormat(date) {
		year = dateParts[0]
		month = dateParts[1]
		day = dateParts[2]
	}

	return utils.FormatYearMonthDayToLongDateTime(year, month, day)
}
