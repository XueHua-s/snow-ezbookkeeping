package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// TransactionMoveAllBetweenAccountsHandler moves all transactions from one account to another account for current user
func (a *TransactionsApi) TransactionMoveAllBetweenAccountsHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionMoveReq models.TransactionMoveBetweenAccountsRequest
	err := c.ShouldBindJSON(&transactionMoveReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionMoveAllBetweenAccountsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionMoveAllBetweenAccountsHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionMoveAllBetweenAccountsHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if transactionMoveReq.FromAccountId == transactionMoveReq.ToAccountId {
		return nil, errs.ErrCannotMoveTransactionToSameAccount
	}

	accountMap, err := a.accounts.GetAccountsByAccountIds(c, uid, []int64{transactionMoveReq.FromAccountId, transactionMoveReq.ToAccountId})

	if err != nil {
		log.Errorf(c, "[transactions.TransactionMoveAllBetweenAccountsHandler] failed to get accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	fromAccount, exists := accountMap[transactionMoveReq.FromAccountId]

	if !exists {
		return nil, errs.ErrSourceAccountNotFound
	}

	toAccount, exists := accountMap[transactionMoveReq.ToAccountId]

	if !exists {
		return nil, errs.ErrDestinationAccountNotFound
	}

	if fromAccount.Hidden || toAccount.Hidden {
		return nil, errs.ErrCannotMoveTransactionFromOrToHiddenAccount
	}

	if fromAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS || toAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
		return nil, errs.ErrCannotMoveTransactionFromOrToParentAccount
	}

	if fromAccount.Currency != toAccount.Currency {
		return nil, errs.ErrCannotMoveTransactionBetweenAccountsWithDifferentCurrencies
	}

	transactions, err := a.transactions.GetAllSpecifiedTransactions(c, uid, 0, 0, 0, nil, []int64{fromAccount.AccountId}, nil, false, "", "", core.MATCH_MODE_DEFAULT, false, pageCountForMovingAccountTransactions, true)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionMoveAllBetweenAccountsHandler] failed to get all transactions of account \"id:%d\" for user \"uid:%d\", because %s", fromAccount.AccountId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allUsedAccounts, err := a.getTransactionUsedAccounts(c, uid, transactions)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionMoveAllBetweenAccountsHandler] failed to get transaction used accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]
		transactionEditable := user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone, allUsedAccounts[transaction.AccountId], allUsedAccounts[transaction.RelatedAccountId])
		newTransactionEditable := transactionEditable

		if transaction.AccountId == fromAccount.AccountId {
			newTransactionEditable = user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone, toAccount, allUsedAccounts[transaction.RelatedAccountId])
		} else if transaction.RelatedAccountId == fromAccount.AccountId {
			newTransactionEditable = user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone, allUsedAccounts[transaction.AccountId], toAccount)
		}

		if !transactionEditable || !newTransactionEditable {
			log.Warnf(c, "[transactions.TransactionMoveAllBetweenAccountsHandler] transaction \"id:%d\" is not editable for user \"uid:%d\"", transaction.TransactionId, uid)
			return nil, errs.ErrCannotModifyTransactionWithThisTransactionTime
		}
	}

	err = a.transactions.MoveAllTransactionsBetweenAccounts(c, uid, fromAccount.AccountId, toAccount.AccountId)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionMoveAllBetweenAccountsHandler] failed to move all transactions from account \"id:%d\" to account \"id:%d\" for user \"uid:%d\", because %s", transactionMoveReq.FromAccountId, transactionMoveReq.ToAccountId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionMoveAllBetweenAccountsHandler] user \"uid:%d\" has moved all transactions from account \"id:%d\" to account \"id:%d\" successfully", uid, transactionMoveReq.FromAccountId, transactionMoveReq.ToAccountId)
	return true, nil
}

// TransactionDeleteHandler deletes an existed transaction by request parameters for current user
func (a *TransactionsApi) TransactionDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionDeleteReq models.TransactionDeleteRequest
	err := c.ShouldBindJSON(&transactionDeleteReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionDeleteHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionDeleteHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	transaction, err := a.transactions.GetTransactionByTransactionId(c, uid, transactionDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionDeleteHandler] failed to get transaction \"id:%d\" for user \"uid:%d\", because %s", transactionDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		log.Warnf(c, "[transactions.TransactionDeleteHandler] cannot delete transaction \"id:%d\" for user \"uid:%d\", because transaction type is transfer in", transactionDeleteReq.Id, uid)
		return nil, errs.ErrTransactionTypeInvalid
	}

	allUsedAccounts, err := a.getTransactionUsedAccounts(c, uid, []*models.Transaction{transaction})

	if err != nil {
		log.Errorf(c, "[transactions.TransactionDeleteHandler] failed to get transaction used accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactionEditable := user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone, allUsedAccounts[transaction.AccountId], allUsedAccounts[transaction.RelatedAccountId])

	if !transactionEditable {
		return nil, errs.ErrCannotDeleteTransactionWithThisTransactionTime
	}

	err = a.transactions.DeleteTransaction(c, uid, transactionDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionDeleteHandler] failed to delete transaction \"id:%d\" for user \"uid:%d\", because %s", transactionDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionDeleteHandler] user \"uid:%d\" has deleted transaction \"id:%d\"", uid, transactionDeleteReq.Id)
	return true, nil
}

// TransactionBatchDeleteHandler deletes existed transactions by request parameters for current user
func (a *TransactionsApi) TransactionBatchDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionBatchDeleteReq models.TransactionBatchDeleteRequest
	err := c.ShouldBindJSON(&transactionBatchDeleteReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchDeleteHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	transactionIds, err := utils.StringArrayToInt64Array(transactionBatchDeleteReq.Ids)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchDeleteHandler] parse transaction ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionIdInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionBatchDeleteHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if !a.users.IsPasswordEqualsUserPassword(transactionBatchDeleteReq.Password, user) {
		return nil, errs.ErrUserPasswordWrong
	}

	transactions, err := a.transactions.GetTransactionsByTransactionIds(c, uid, transactionIds)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchDeleteHandler] failed to get transactions for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allUsedAccounts, err := a.getTransactionUsedAccounts(c, uid, transactions)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchDeleteHandler] failed to get transaction used accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]
		transactionEditable := user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone, allUsedAccounts[transaction.AccountId], allUsedAccounts[transaction.RelatedAccountId])

		if !transactionEditable {
			log.Warnf(c, "[transactions.TransactionBatchDeleteHandler] transaction \"id:%d\" is not editable for user \"uid:%d\"", transaction.TransactionId, uid)
			return nil, errs.ErrCannotDeleteTransactionWithThisTransactionTime
		}
	}

	deletedCount := 0

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]
		err = a.transactions.DeleteTransaction(c, uid, transaction.TransactionId)

		if err != nil {
			log.Errorf(c, "[transactions.TransactionBatchDeleteHandler] failed to delete transaction \"id:%d\" for user \"uid:%d\", because %s", transaction.TransactionId, uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		deletedCount++
	}

	log.Infof(c, "[transactions.TransactionBatchDeleteHandler] user \"uid:%d\" has deleted %d transactions", uid, deletedCount)
	return true, nil
}
