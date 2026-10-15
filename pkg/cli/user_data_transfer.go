package cli

import (
	"github.com/mayswind/ezbookkeeping/pkg/converters"
	"github.com/mayswind/ezbookkeeping/pkg/converters/converter"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"strings"
	"time"
)

// ExportTransaction returns csv file content according user all transactions
func (l *UserDataCli) ExportTransaction(c *core.CliContext, username string, fileType string) ([]byte, error) {
	if username == "" {
		log.CliErrorf(c, "[user_data.ExportTransaction] user name is empty")
		return nil, errs.ErrUsernameIsEmpty
	}

	uid, err := l.getUserIdByUsername(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.ExportTransaction] error occurs when getting user id by user name")
		return nil, err
	}

	accountMap, categoryMap, tagMap, _, tagIndexesMap, err := l.getUserEssentialData(c, uid, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.ExportTransaction] failed to get essential data for user \"%s\", because %s", username, err.Error())
		return nil, err
	}

	allTransactions, err := l.transactions.GetAllTransactions(c, uid, pageCountForDataExport, true)

	if err != nil {
		log.CliErrorf(c, "[user_data.ExportTransaction] failed to all transactions for user \"%s\", because %s", username, err.Error())
		return nil, err
	}

	dataExporter := converters.GetTransactionDataExporter(fileType)

	if dataExporter == nil {
		return nil, errs.ErrNotImplemented
	}

	result, err := dataExporter.ToExportedContent(c, uid, allTransactions, accountMap, categoryMap, tagMap, tagIndexesMap)

	if err != nil {
		log.CliErrorf(c, "[user_data.ExportTransaction] failed to get csv format exported data for \"%s\", because %s", username, err.Error())
		return nil, err
	}

	return result, nil
}

func (l *UserDataCli) ImportTransaction(c *core.CliContext, username string, fileType string, data []byte) error {
	if username == "" {
		log.CliErrorf(c, "[user_data.ImportTransaction] user name is empty")
		return errs.ErrUsernameIsEmpty
	}

	dataImporter, err := converters.GetTransactionDataImporter(fileType)

	if err != nil {
		return err
	}

	user, err := l.GetUserByUsername(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.ImportTransaction] failed to get user by user name \"%s\", because %s", username, err.Error())
		return err
	}

	accountMap, expenseCategoryMap, incomeCategoryMap, transferCategoryMap, tagMap, err := l.getUserEssentialDataForImport(c, user.Uid, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.ImportTransaction] failed to get essential data for user \"%s\", because %s", username, err.Error())
		return err
	}

	parsedTransactions, newAccounts, newSubExpenseCategories, newSubIncomeCategories, newSubTransferCategories, newTags, err := dataImporter.ParseImportedData(c, user, data, time.Local, converter.DefaultImporterOptions, accountMap, expenseCategoryMap, incomeCategoryMap, transferCategoryMap, tagMap)

	if err != nil {
		log.CliErrorf(c, "[user_data.ImportTransaction] failed to parse imported data for \"%s\", because %s", username, err.Error())
		return err
	}

	if len(parsedTransactions) < 1 {
		log.CliErrorf(c, "[user_data.ImportTransaction] there are no transactions in import file")
		return errs.ErrOperationFailed
	}

	if len(newAccounts) > 0 {
		accountNames := l.accounts.GetAccountNames(newAccounts)
		log.CliErrorf(c, "[user_data.ImportTransaction] there are %d accounts (%s) need to be created, please create them manually", len(newAccounts), strings.Join(accountNames, ","))
		return errs.ErrOperationFailed
	}

	if len(newSubExpenseCategories) > 0 {
		categoryNames := l.categories.GetCategoryNames(newSubExpenseCategories)
		log.CliErrorf(c, "[user_data.ImportTransaction] there are %d expense categories (%s) need to be created, please create them manually", len(newSubExpenseCategories), strings.Join(categoryNames, ","))
		return errs.ErrOperationFailed
	}

	if len(newSubIncomeCategories) > 0 {
		categoryNames := l.categories.GetCategoryNames(newSubIncomeCategories)
		log.CliErrorf(c, "[user_data.ImportTransaction] there are %d income categories (%s) need to be created, please create them manually", len(newSubIncomeCategories), strings.Join(categoryNames, ","))
		return errs.ErrOperationFailed
	}

	if len(newSubTransferCategories) > 0 {
		categoryNames := l.categories.GetCategoryNames(newSubTransferCategories)
		log.CliErrorf(c, "[user_data.ImportTransaction] there are %d transfer categories (%s) need to be created, please create them manually", len(newSubTransferCategories), strings.Join(categoryNames, ","))
		return errs.ErrOperationFailed
	}

	if len(newTags) > 0 {
		tagNames := l.tags.GetTagNames(newTags)
		log.CliErrorf(c, "[user_data.ImportTransaction] there are %d transaction tags (%s) need to be created, please create them manually", len(newTags), strings.Join(tagNames, ","))
		return errs.ErrOperationFailed
	}

	newTransactions := parsedTransactions.ToTransactionsList()
	newTransactionTagIdsMap, err := parsedTransactions.ToTransactionTagIdsMap()

	if err != nil {
		log.CliErrorf(c, "[user_data.ImportTransaction] failed to get transaction tag ids map, because %s", err.Error())
		return errs.ErrOperationFailed
	}

	err = l.transactions.BatchCreateTransactions(c, user.Uid, newTransactions, newTransactionTagIdsMap, nil)

	if err != nil {
		log.CliErrorf(c, "[user_data.ImportTransaction] failed to create transaction, because %s", err.Error())
		return err
	}

	return nil
}
