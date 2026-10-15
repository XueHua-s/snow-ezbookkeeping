package cli

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// CheckTransactionAndAccount checks whether all user transactions and all user accounts are correct
func (l *UserDataCli) CheckTransactionAndAccount(c *core.CliContext, username string) (bool, error) {
	if username == "" {
		log.CliErrorf(c, "[user_data.CheckTransactionAndAccount] user name is empty")
		return false, errs.ErrUsernameIsEmpty
	}

	uid, err := l.getUserIdByUsername(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.CheckTransactionAndAccount] error occurs when getting user id by user name")
		return false, err
	}

	accountMap, categoryMap, tagMap, tagIndexes, tagIndexesMap, err := l.getUserEssentialData(c, uid, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.CheckTransactionAndAccount] failed to get essential data for user \"%s\", because %s", username, err.Error())
		return false, err
	}

	accountHasChild := make(map[int64]bool)

	for _, account := range accountMap {
		if account.ParentAccountId > models.LevelOneAccountParentId {
			accountHasChild[account.ParentAccountId] = true
		}
	}

	allTransactions, err := l.transactions.GetAllTransactions(c, uid, pageCountForGettingTransactions, false)

	if err != nil {
		log.CliErrorf(c, "[user_data.CheckTransactionAndAccount] failed to all transactions for user \"%s\", because %s", username, err.Error())
		return false, err
	}

	transactionMap := l.transactions.GetTransactionMapByList(allTransactions)
	accountBalance := make(map[int64]int64)

	for i := len(allTransactions) - 1; i >= 0; i-- {
		transaction := allTransactions[i]

		err := l.checkTransactionAccount(c, transaction, accountMap, accountHasChild)

		if err != nil {
			return false, err
		}

		err = l.checkTransactionCategory(c, transaction, categoryMap)

		if err != nil {
			return false, err
		}

		err = l.checkTransactionTag(c, transaction.TransactionId, tagIndexesMap, tagMap)

		if err != nil {
			return false, err
		}

		err = l.checkTransactionRelatedTransaction(c, transaction, transactionMap, accountMap)

		if err != nil {
			return false, err
		}

		balance, exists := accountBalance[transaction.AccountId]

		if !exists {
			balance = 0
		}

		if transaction.Type == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
			balance = balance + transaction.RelatedAccountAmount
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_INCOME {
			balance = balance + transaction.Amount
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_EXPENSE {
			balance = balance - transaction.Amount
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			balance = balance - transaction.Amount
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			balance = balance + transaction.Amount
		} else {
			log.CliErrorf(c, "[user_data.CheckTransactionAndAccount] transaction type of transaction \"id:%d\" is invalid", transaction.TransactionId)
			return false, errs.ErrOperationFailed
		}

		accountBalance[transaction.AccountId] = balance
	}

	for _, account := range accountMap {
		actualBalance, exists := accountBalance[account.AccountId]

		if !exists && account.Balance == 0 {
			continue
		}

		if !exists && account.Balance != 0 {
			log.CliErrorf(c, "[user_data.CheckTransactionAndAccount] account \"id:%d\" balance is not correct, expected balance is %d, but there is no transaction actually", account.AccountId, account.Balance)
			return false, errs.ErrOperationFailed
		}

		if account.Balance != actualBalance {
			log.CliErrorf(c, "[user_data.CheckTransactionAndAccount] account \"id:%d\" balance is not correct, expected balance is %d, but actual balance is %d", account.AccountId, account.Balance, actualBalance)
			return false, errs.ErrOperationFailed
		}
	}

	for accountId, actualBalance := range accountBalance {
		_, exists := accountMap[accountId]

		if !exists {
			log.CliErrorf(c, "[user_data.CheckTransactionAndAccount] account \"id:%d\" does not exist, but there are some transactions of this account actually, and actual balance is %d", accountId, actualBalance)
			return false, errs.ErrOperationFailed
		}
	}

	for i := 0; i < len(tagIndexes); i++ {
		tagIndex := tagIndexes[i]

		if tagIndex.TransactionTime < 1 {
			log.CliErrorf(c, "[user_data.CheckTransactionAndAccount] transaction tag index \"id:%d\" does not have transaction time", tagIndex.TagIndexId)
			return false, errs.ErrOperationFailed
		}
	}

	return true, nil
}

// FixTransactionTagIndexWithTransactionTime fixes user transaction tag index data with transaction time
func (l *UserDataCli) FixTransactionTagIndexWithTransactionTime(c *core.CliContext, username string) (bool, error) {
	if username == "" {
		log.CliErrorf(c, "[user_data.FixTransactionTagIndexWithTransactionTime] user name is empty")
		return false, errs.ErrUsernameIsEmpty
	}

	uid, err := l.getUserIdByUsername(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.FixTransactionTagIndexWithTransactionTime] error occurs when getting user id by user name")
		return false, err
	}

	tagIndexes, err := l.tags.GetAllTagIdsOfAllTransactions(c, uid)

	if err != nil {
		log.CliErrorf(c, "[user_data.FixTransactionTagIndexWithTransactionTime] failed to get tag index for user \"%s\", because %s", username, err.Error())
		return false, err
	}

	invalidTagIndexes := make([]*models.TransactionTagIndex, 0, len(tagIndexes))

	for i := 0; i < len(tagIndexes); i++ {
		tagIndex := tagIndexes[i]

		if tagIndex.TransactionTime < 1 {
			invalidTagIndexes = append(invalidTagIndexes, tagIndex)
		}
	}

	if len(invalidTagIndexes) < 1 {
		log.CliErrorf(c, "[user_data.FixTransactionTagIndexWithTransactionTime] all user transaction tag index data has been checked, there is no problem with user data")
		return false, errs.ErrOperationFailed
	}

	allTransactions, err := l.transactions.GetAllTransactions(c, uid, pageCountForGettingTransactions, false)

	if err != nil {
		log.CliErrorf(c, "[user_data.FixTransactionTagIndexWithTransactionTime] failed to all transactions for user \"%s\", because %s", username, err.Error())
		return false, err
	}

	transactionMap := l.transactions.GetTransactionMapByList(allTransactions)

	for i := 0; i < len(invalidTagIndexes); i++ {
		tagIndex := invalidTagIndexes[i]
		transaction, exists := transactionMap[tagIndex.TransactionId]

		if !exists {
			continue
		}

		tagIndex.TransactionTime = transaction.TransactionTime
	}

	err = l.tags.ModifyTagIndexTransactionTime(c, uid, invalidTagIndexes)

	if err != nil {
		log.CliErrorf(c, "[user_data.FixTransactionTagIndexWithTransactionTime] failed to update transaction tag index for user \"%s\", because %s", username, err.Error())
		return false, err
	}

	return true, nil
}
