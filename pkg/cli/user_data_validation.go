package cli

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func (l *UserDataCli) checkTransactionAccount(c *core.CliContext, transaction *models.Transaction, accountMap map[int64]*models.Account, accountHasChild map[int64]bool) error {
	account, exists := accountMap[transaction.AccountId]

	if !exists {
		log.CliErrorf(c, "[user_data.checkTransactionAccount] the account \"id:%d\" of transaction \"id:%d\" does not exist", transaction.AccountId, transaction.TransactionId)
		return errs.ErrAccountNotFound
	}

	if account.ParentAccountId == models.LevelOneAccountParentId && accountHasChild[account.AccountId] {
		log.CliErrorf(c, "[user_data.checkTransactionAccount] the account \"id:%d\" of transaction \"id:%d\" is not a sub-account", transaction.AccountId, transaction.TransactionId)
		return errs.ErrOperationFailed
	}

	if transaction.RelatedAccountId > 0 {
		relatedAccount, exists := accountMap[transaction.RelatedAccountId]

		if !exists {
			log.CliErrorf(c, "[user_data.checkTransactionAccount] the related account \"id:%d\" of transaction \"id:%d\" does not exist", transaction.RelatedAccountId, transaction.TransactionId)
			return errs.ErrAccountNotFound
		}

		if relatedAccount.ParentAccountId == models.LevelOneAccountParentId && accountHasChild[relatedAccount.AccountId] {
			log.CliErrorf(c, "[user_data.checkTransactionAccount] the related account \"id:%d\" of transaction \"id:%d\" is not a sub-account", transaction.RelatedAccountId, transaction.TransactionId)
			return errs.ErrOperationFailed
		}
	}

	return nil
}

func (l *UserDataCli) checkTransactionCategory(c *core.CliContext, transaction *models.Transaction, categoryMap map[int64]*models.TransactionCategory) error {
	if transaction.Type == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
		if transaction.CategoryId > 0 {
			log.CliErrorf(c, "[user_data.checkTransactionCategory] transaction \"id:%d\" is balance modification transaction, but has category \"id:%d\"", transaction.TransactionId, transaction.CategoryId)
			return errs.ErrBalanceModificationTransactionCannotSetCategory
		} else {
			return nil
		}
	}

	category, exists := categoryMap[transaction.CategoryId]

	if !exists {
		log.CliErrorf(c, "[user_data.checkTransactionCategory] the transaction category \"id:%d\" of transaction \"id:%d\" does not exist", transaction.CategoryId, transaction.TransactionId)
		return errs.ErrTransactionCategoryNotFound
	}

	if category.ParentCategoryId == models.LevelOneTransactionCategoryParentId {
		log.CliErrorf(c, "[user_data.checkTransactionCategory] the transaction category \"id:%d\" of transaction \"id:%d\" is not a sub category", transaction.CategoryId, transaction.TransactionId)
		return errs.ErrOperationFailed
	}

	return nil
}

func (l *UserDataCli) checkTransactionTag(c *core.CliContext, transactionId int64, allTagIndexesMap map[int64][]int64, tagMap map[int64]*models.TransactionTag) error {
	tagIndexes, exists := allTagIndexesMap[transactionId]

	if !exists {
		return nil
	}

	for i := 0; i < len(tagIndexes); i++ {
		tagIndex := tagIndexes[i]
		tag, exists := tagMap[tagIndex]

		if !exists {
			log.CliErrorf(c, "[user_data.checkTransactionTag] the transaction tag \"id:%d\" of transaction \"id:%d\" does not exist", tag.TagId, transactionId)
			return errs.ErrTransactionTagNotFound
		}
	}

	return nil
}

func (l *UserDataCli) checkTransactionRelatedTransaction(c *core.CliContext, transaction *models.Transaction, transactionMap map[int64]*models.Transaction, accountMap map[int64]*models.Account) error {
	if transaction.Type != models.TRANSACTION_DB_TYPE_TRANSFER_OUT && transaction.Type != models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		return nil
	}

	relatedTransaction, exists := transactionMap[transaction.RelatedId]

	if !exists {
		log.CliErrorf(c, "[user_data.checkTransactionRelatedTransaction] the related transaction \"id:%d\" of transaction \"id:%d\" does not exist", transaction.RelatedId, transaction.TransactionId)
		return errs.ErrTransactionNotFound
	}

	if transaction.RelatedId != relatedTransaction.TransactionId || transaction.TransactionId != relatedTransaction.RelatedId {
		log.CliErrorf(c, "[user_data.checkTransactionRelatedTransaction] related ids of transaction \"id:%d\" and transaction \"id:%d\" are not equal", transaction.RelatedId, transaction.TransactionId)
		return errs.ErrOperationFailed
	}

	if transaction.RelatedAccountId != relatedTransaction.AccountId || transaction.AccountId != relatedTransaction.RelatedAccountId {
		log.CliErrorf(c, "[user_data.checkTransactionRelatedTransaction] related account ids of transaction \"id:%d\" and transaction \"id:%d\" are not equal", transaction.RelatedId, transaction.TransactionId)
		return errs.ErrOperationFailed
	}

	if transaction.RelatedAccountAmount != relatedTransaction.Amount || transaction.Amount != relatedTransaction.RelatedAccountAmount {
		log.CliErrorf(c, "[user_data.checkTransactionRelatedTransaction] related amounts of transaction \"id:%d\" and transaction \"id:%d\" are not equal", transaction.RelatedId, transaction.TransactionId)
		return errs.ErrOperationFailed
	}

	account := accountMap[transaction.AccountId]
	relatedAccount := accountMap[transaction.RelatedAccountId]

	if account.Currency == relatedAccount.Currency && transaction.Amount != transaction.RelatedAccountAmount {
		log.CliWarnf(c, "[user_data.checkTransactionRelatedTransaction] transfer-in amount and transfer-out amount of transaction \"id:%d\" are not equal", transaction.TransactionId)
	}

	return nil
}
