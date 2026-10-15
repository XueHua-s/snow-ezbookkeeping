package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// TransactionBatchUpdateCategoriesHandler batch updates categories of transactions by request parameters for current user
func (a *TransactionsApi) TransactionBatchUpdateCategoriesHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionBatchUpdateReq models.TransactionBatchUpdateCategoryRequest
	err := c.ShouldBindJSON(&transactionBatchUpdateReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchUpdateCategoriesHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchUpdateCategoriesHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	transactionIds, err := utils.StringArrayToInt64Array(transactionBatchUpdateReq.TransactionIds)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchUpdateCategoriesHandler] parse transaction ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionIdInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionBatchUpdateCategoriesHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	category, err := a.transactionCategories.GetCategoryByCategoryId(c, uid, transactionBatchUpdateReq.CategoryId)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchUpdateCategoriesHandler] failed to get category \"id:%d\" for user \"uid:%d\", because %s", transactionBatchUpdateReq.CategoryId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if category.ParentCategoryId == models.LevelOneTransactionCategoryParentId {
		log.Warnf(c, "[transactions.TransactionBatchUpdateCategoriesHandler] transaction category \"id:%d\" is not a sub category", category.CategoryId)
		return nil, errs.ErrCannotUsePrimaryCategoryForTransaction
	}

	var expectedTransactionType models.TransactionDbType

	if category.Type == models.CATEGORY_TYPE_EXPENSE {
		expectedTransactionType = models.TRANSACTION_DB_TYPE_EXPENSE
	} else if category.Type == models.CATEGORY_TYPE_INCOME {
		expectedTransactionType = models.TRANSACTION_DB_TYPE_INCOME
	} else if category.Type == models.CATEGORY_TYPE_TRANSFER {
		expectedTransactionType = models.TRANSACTION_DB_TYPE_TRANSFER_OUT
	}

	transactions, err := a.transactions.GetTransactionsByTransactionIds(c, uid, transactionIds)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchUpdateCategoriesHandler] failed to get transactions for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allUsedAccounts, err := a.getTransactionUsedAccounts(c, uid, transactions)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchUpdateCategoriesHandler] failed to get transaction used accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allTransactionIds := make([]int64, 0, len(transactions))

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]

		if transaction.Type != expectedTransactionType {
			log.Warnf(c, "[transactions.TransactionBatchUpdateCategoriesHandler] transaction \"id:%d\" type is not expected type \"%d\" for user \"uid:%d\"", transaction.TransactionId, expectedTransactionType, uid)
			return nil, errs.ErrTransactionTypeInvalid
		}

		transactionEditable := user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone, allUsedAccounts[transaction.AccountId], allUsedAccounts[transaction.RelatedAccountId])

		if !transactionEditable {
			log.Warnf(c, "[transactions.TransactionBatchUpdateCategoriesHandler] transaction \"id:%d\" is not editable for user \"uid:%d\"", transaction.TransactionId, uid)
			return nil, errs.ErrCannotModifyTransactionWithThisTransactionTime
		}

		allTransactionIds = append(allTransactionIds, transaction.TransactionId)

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			allTransactionIds = append(allTransactionIds, transaction.RelatedId)
		}
	}

	err = a.transactions.BatchUpdateTransactionsCategory(c, uid, allTransactionIds, category.CategoryId)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchUpdateCategoriesHandler] failed to batch update transactions category for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionBatchUpdateCategoriesHandler] user \"uid:%d\" has batch updated category of %d transactions successfully", uid, len(transactionBatchUpdateReq.TransactionIds))
	return true, nil
}

// TransactionBatchUpdateAccountsHandler batch updates accounts of transactions by request parameters for current user
func (a *TransactionsApi) TransactionBatchUpdateAccountsHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionBatchUpdateReq models.TransactionBatchUpdateAccountRequest
	err := c.ShouldBindJSON(&transactionBatchUpdateReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchUpdateAccountsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchUpdateAccountsHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	transactionIds, err := utils.StringArrayToInt64Array(transactionBatchUpdateReq.TransactionIds)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchUpdateAccountsHandler] parse transaction ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionIdInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionBatchUpdateAccountsHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	allAccounts, err := a.accounts.GetAllAccountsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchUpdateAccountsHandler] failed to get all accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	accountMap := a.accounts.GetAccountMapByList(allAccounts)
	account, exists := accountMap[transactionBatchUpdateReq.AccountId]

	if !exists || account == nil {
		log.Warnf(c, "[transactions.TransactionBatchUpdateAccountsHandler] account \"id:%d\" does not exist for user \"uid:%d\"", transactionBatchUpdateReq.AccountId, uid)
		return nil, errs.ErrAccountNotFound
	}

	if account.Hidden {
		log.Warnf(c, "[transactions.TransactionBatchUpdateAccountsHandler] account \"id:%d\" is hidden for user \"uid:%d\"", account.AccountId, uid)
		return nil, errs.ErrCannotMoveTransactionFromOrToHiddenAccount
	}

	if account.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
		log.Warnf(c, "[transactions.TransactionBatchUpdateAccountsHandler] account \"id:%d\" is a parent account, cannot be used for transaction of user \"uid:%d\"", account.AccountId, uid)
		return nil, errs.ErrCannotModifyTransactionInParentAccount
	}

	transactions, err := a.transactions.GetTransactionsByTransactionIds(c, uid, transactionIds)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchUpdateAccountsHandler] failed to get transactions for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			log.Warnf(c, "[transactions.TransactionBatchUpdateAccountsHandler] cannot modify transaction \"id:%d\" for user \"uid:%d\", because transaction type is transfer in", transaction.TransactionId, uid)
			return nil, errs.ErrTransactionTypeInvalid
		}

		if transactionBatchUpdateReq.IsDestinationAccount && transaction.Type != models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			log.Warnf(c, "[transactions.TransactionBatchUpdateAccountsHandler] cannot update destination account of non-transfer transaction \"id:%d\" for user \"uid:%d\"", transaction.TransactionId, uid)
			return nil, errs.ErrTransactionDestinationAccountCannotBeSet
		}

		if !transactionBatchUpdateReq.IsDestinationAccount && account.AccountId == transaction.RelatedAccountId {
			log.Warnf(c, "[transactions.TransactionBatchUpdateAccountsHandler] cannot update account to same destination account of transaction \"id:%d\" for user \"uid:%d\"", transaction.TransactionId, uid)
			return nil, errs.ErrTransactionSourceAndDestinationIdCannotBeEqual
		} else if transactionBatchUpdateReq.IsDestinationAccount && account.AccountId == transaction.AccountId {
			log.Warnf(c, "[transactions.TransactionBatchUpdateAccountsHandler] cannot update destination account to same source account of transaction \"id:%d\" for user \"uid:%d\"", transaction.TransactionId, uid)
			return nil, errs.ErrTransactionSourceAndDestinationIdCannotBeEqual
		}

		var oldAccount *models.Account

		if !transactionBatchUpdateReq.IsDestinationAccount {
			oldAccount = accountMap[transaction.AccountId]
		} else if transactionBatchUpdateReq.IsDestinationAccount {
			oldAccount = accountMap[transaction.RelatedAccountId]
		}

		if oldAccount == nil {
			log.Warnf(c, "[transactions.TransactionBatchUpdateAccountsHandler] the original account of transaction \"id:%d\" does not exist for user \"uid:%d\"", transaction.TransactionId, uid)
			return nil, errs.ErrAccountNotFound
		}

		if oldAccount.Hidden {
			log.Warnf(c, "[transactions.TransactionBatchUpdateAccountsHandler] the original account of transaction \"id:%d\" is hidden for user \"uid:%d\"", transaction.TransactionId, uid)
			return nil, errs.ErrCannotMoveTransactionFromOrToHiddenAccount
		}

		if oldAccount.Currency != account.Currency {
			log.Warnf(c, "[transactions.TransactionBatchUpdateAccountsHandler] cannot update account of transaction \"id:%d\", because the original account currency \"%s\" is different from updated account currency \"%s\" for user \"uid:%d\"", transaction.TransactionId, oldAccount.Currency, account.Currency, uid)
			return nil, errs.ErrCannotMoveTransactionBetweenAccountsWithDifferentCurrencies
		}

		newSourceAccount := accountMap[transaction.AccountId]
		newDestinationAccount := accountMap[transaction.RelatedAccountId]

		if !transactionBatchUpdateReq.IsDestinationAccount && transaction.AccountId != account.AccountId {
			newSourceAccount = account
		} else if transactionBatchUpdateReq.IsDestinationAccount && transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT && transaction.RelatedAccountId != account.AccountId {
			newDestinationAccount = account
		}

		transactionEditable := user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone, accountMap[transaction.AccountId], accountMap[transaction.RelatedAccountId])
		newTransactionEditable := user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone, newSourceAccount, newDestinationAccount)

		if !transactionEditable || !newTransactionEditable {
			log.Warnf(c, "[transactions.TransactionBatchUpdateAccountsHandler] transaction \"id:%d\" is not editable for user \"uid:%d\"", transaction.TransactionId, uid)
			return nil, errs.ErrCannotModifyTransactionWithThisTransactionTime
		}
	}

	updatedCount := 0

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]

		if !transactionBatchUpdateReq.IsDestinationAccount && transaction.AccountId != account.AccountId {
			transaction.AccountId = account.AccountId
		} else if transactionBatchUpdateReq.IsDestinationAccount && transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT && transaction.RelatedAccountId != account.AccountId {
			transaction.RelatedAccountId = account.AccountId
		} else {
			log.Warnf(c, "[transactions.TransactionBatchUpdateAccountsHandler] skip updating transaction \"id:%d\", because the original account is same as updated account for user \"uid:%d\"", transaction.TransactionId, uid)
			continue
		}

		err = a.transactions.ModifyTransaction(c, transaction, false, 0, nil, nil, nil, nil)

		if err != nil {
			log.Errorf(c, "[transactions.TransactionBatchUpdateAccountsHandler] failed to update transaction \"id:%d\" for user \"uid:%d\", because %s", transaction.TransactionId, uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		updatedCount++
	}

	if updatedCount < 1 {
		return nil, errs.ErrNothingWillBeUpdated
	}

	log.Infof(c, "[transactions.TransactionBatchUpdateAccountsHandler] user \"uid:%d\" has batch updated account of %d transactions successfully", uid, updatedCount)
	return true, nil
}
