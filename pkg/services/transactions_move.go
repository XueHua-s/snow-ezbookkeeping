package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"time"
	"xorm.io/xorm"
)

// MoveAllTransactionsBetweenAccounts moves all transactions from one account to another account, and combine balance modification transactions if necessary
func (s *TransactionService) MoveAllTransactionsBetweenAccounts(c core.Context, uid int64, fromAccountId int64, toAccountId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if fromAccountId <= 0 || toAccountId <= 0 {
		return errs.ErrAccountIdInvalid
	}

	if fromAccountId == toAccountId {
		return errs.ErrCannotMoveTransactionToSameAccount
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		// get and verify from and to account
		fromAccount := &models.Account{}
		has, err := sess.ID(fromAccountId).Where("uid=? AND deleted=?", uid, false).Get(fromAccount)

		if err != nil {
			return err
		} else if !has {
			return errs.ErrAccountNotFound
		}

		toAccount := &models.Account{}
		has, err = sess.ID(toAccountId).Where("uid=? AND deleted=?", uid, false).Get(toAccount)

		if err != nil {
			return err
		} else if !has {
			return errs.ErrAccountNotFound
		}

		if fromAccount.Hidden || toAccount.Hidden {
			return errs.ErrCannotMoveTransactionFromOrToHiddenAccount
		}

		if fromAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS || toAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
			return errs.ErrCannotMoveTransactionFromOrToParentAccount
		}

		if fromAccount.Currency != toAccount.Currency {
			return errs.ErrCannotMoveTransactionBetweenAccountsWithDifferentCurrencies
		}

		// combine balance modification transaction
		var balanceModificationTransactions []*models.Transaction
		err = sess.Where("uid=? AND deleted=? AND type=? AND (account_id=? OR account_id=?)", uid, false, models.TRANSACTION_DB_TYPE_MODIFY_BALANCE, fromAccountId, toAccountId).Find(&balanceModificationTransactions)

		if err != nil {
			return err
		}

		if len(balanceModificationTransactions) > 2 {
			log.Errorf(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" has more than 2 balance modification transactions in account \"id:%d\" and account \"id:%d\", cannot combine balance modification transaction", uid, fromAccountId, toAccountId)
			return errs.ErrOperationFailed
		} else if len(balanceModificationTransactions) == 2 && balanceModificationTransactions[0].AccountId != balanceModificationTransactions[1].AccountId {
			// if two balance modification transactions exist, merge the amounts into the earlier one and delete the later transaction
			var earlierTransaction *models.Transaction
			var laterTransaction *models.Transaction

			if balanceModificationTransactions[0].TransactionTime < balanceModificationTransactions[1].TransactionTime {
				earlierTransaction = balanceModificationTransactions[0]
				laterTransaction = balanceModificationTransactions[1]
			} else {
				earlierTransaction = balanceModificationTransactions[1]
				laterTransaction = balanceModificationTransactions[0]
			}

			mergedAmount, validAmount := utils.AddInt64(earlierTransaction.Amount, laterTransaction.Amount)
			mergedRelatedAccountAmount, validRelatedAmount := utils.AddInt64(earlierTransaction.RelatedAccountAmount, laterTransaction.RelatedAccountAmount)

			if !validAmount || !validRelatedAmount ||
				(mergedAmount <= models.MinimumTransactionAmount || mergedAmount >= models.MaximumTransactionAmount) ||
				(mergedRelatedAccountAmount <= models.MinimumTransactionAmount || mergedRelatedAccountAmount >= models.MaximumTransactionAmount) {
				log.Errorf(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" cannot combine balance modification transactions \"id:%d\" and \"id:%d\", because the merged amount exceeds the supported range", uid, earlierTransaction.TransactionId, laterTransaction.TransactionId)
				return errs.ErrMergedBalanceModificationTransactionAmountOverflow
			}

			earlierTransaction.Amount = mergedAmount
			earlierTransaction.RelatedAccountAmount = mergedRelatedAccountAmount
			earlierTransaction.UpdatedUnixTime = time.Now().Unix()

			updatedRows, err := sess.ID(earlierTransaction.TransactionId).Cols("amount", "related_account_amount", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(earlierTransaction)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.MoveAllTransactionsBetweenAccounts] failed to update earlier balance modification transaction")
				return errs.ErrDatabaseOperationFailed
			}

			laterTransaction.Deleted = true
			laterTransaction.DeletedUnixTime = time.Now().Unix()

			deletedRows, err := sess.ID(laterTransaction.TransactionId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(laterTransaction)

			if err != nil {
				return err
			} else if deletedRows < 1 {
				log.Errorf(c, "[transactions.MoveAllTransactionsBetweenAccounts] failed to delete later balance modification transaction")
				return errs.ErrDatabaseOperationFailed
			}

			log.Infof(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" has combined two balance modification transactions \"id:%d\" and \"id:%d\", retained transaction is \"id:%d\"", uid, earlierTransaction.TransactionId, laterTransaction.TransactionId, earlierTransaction.TransactionId)
		} else if len(balanceModificationTransactions) == 1 {
			// when merging a new balance modification transaction, if its date is later than the account's earliest transaction, update the balance modification transaction time accordingly
			anotherAccountId := int64(0)

			if balanceModificationTransactions[0].AccountId == fromAccountId {
				anotherAccountId = toAccountId
			} else if balanceModificationTransactions[0].AccountId == toAccountId {
				anotherAccountId = fromAccountId
			} else {
				log.Errorf(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" has a balance modification transaction \"id:%d\" which account id is neither \"%d\" nor \"%d\"", uid, balanceModificationTransactions[0].TransactionId, fromAccountId, toAccountId)
				return errs.ErrOperationFailed
			}

			earliestTransaction := &models.Transaction{}
			has, err := sess.Where("uid=? AND deleted=? AND account_id=?", uid, false, anotherAccountId).OrderBy("transaction_time asc").Limit(1).Get(earliestTransaction)

			if err != nil {
				return err
			} else if has && balanceModificationTransactions[0].TransactionTime > earliestTransaction.TransactionTime {
				balanceModificationTransaction := balanceModificationTransactions[0]
				earliestTransactionUnixTime := utils.GetUnixTimeFromTransactionTime(earliestTransaction.TransactionTime)

				newBalanceModificationMinTransactionTime := utils.GetMinTransactionTimeFromUnixTime(earliestTransactionUnixTime - 1)
				newBalanceModificationMaxTransactionTime := utils.GetMaxTransactionTimeFromUnixTime(earliestTransactionUnixTime - 1)
				balanceModificationTransaction.TransactionTime = newBalanceModificationMinTransactionTime

				sameSecondLatestTransaction := &models.Transaction{}
				has, err = sess.Where("uid=? AND transaction_time>=? AND transaction_time<=?", balanceModificationTransaction.Uid, newBalanceModificationMinTransactionTime, newBalanceModificationMaxTransactionTime).OrderBy("transaction_time desc").Limit(1).Get(sameSecondLatestTransaction)

				if err != nil {
					log.Errorf(c, "[transactions.MoveAllTransactionsBetweenAccounts] failed to get trasaction time, because %s", err.Error())
					return err
				}

				if has && sameSecondLatestTransaction.TransactionTime < newBalanceModificationMaxTransactionTime-1 {
					balanceModificationTransaction.TransactionTime = sameSecondLatestTransaction.TransactionTime + 1
				} else if has && sameSecondLatestTransaction.TransactionTime == newBalanceModificationMaxTransactionTime-1 {
					return errs.ErrTooMuchTransactionInOneSecond
				}

				balanceModificationTransaction.UpdatedUnixTime = time.Now().Unix()

				if balanceModificationTransaction.TransactionTime < 0 {
					balanceModificationTransaction.TransactionTime = 0
				}

				updatedRows, err := sess.ID(balanceModificationTransaction.TransactionId).Cols("transaction_time", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(balanceModificationTransaction)

				if err != nil {
					return err
				} else if updatedRows < 1 {
					log.Errorf(c, "[transactions.MoveAllTransactionsBetweenAccounts] failed to update balance modification transaction time")
					return errs.ErrDatabaseOperationFailed
				}

				log.Infof(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" has updated balance modification transaction \"id:%d\" time to %d, because earliest transaction time in account \"id:%d\" is %d", uid, balanceModificationTransaction.TransactionId, balanceModificationTransaction.TransactionTime, toAccountId, earliestTransaction.TransactionTime)
			}
		}

		// update all transactions of from account
		updateModel := &models.Transaction{
			AccountId:       toAccountId,
			UpdatedUnixTime: time.Now().Unix(),
		}

		updatedRows, err := sess.Cols("account_id", "updated_unix_time").Where("uid=? AND deleted=? AND account_id=?", uid, false, fromAccountId).Update(updateModel)

		if err != nil {
			return err
		}

		if updatedRows > 0 {
			log.Infof(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" has moved %d transactions from account \"id:%d\" to account \"id:%d\"", uid, updatedRows, fromAccountId, toAccountId)
		}

		// update all related transactions of from account
		updateRelatedModel := &models.Transaction{
			RelatedAccountId: toAccountId,
			UpdatedUnixTime:  time.Now().Unix(),
		}

		relatedUpdatedRows, err := sess.Cols("related_account_id", "updated_unix_time").Where("uid=? AND deleted=? AND related_account_id=?", uid, false, fromAccountId).Update(updateRelatedModel)

		if err != nil {
			return err
		}

		if updatedRows > 0 {
			log.Infof(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" has moved %d related transactions from account \"id:%d\" to account \"id:%d\"", uid, relatedUpdatedRows, fromAccountId, toAccountId)
		}

		// delete all transfer transactions which related account id and account id are both
		deletedModel := &models.Transaction{
			Deleted:         true,
			DeletedUnixTime: time.Now().Unix(),
		}

		deletedRows, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=? AND (type=? OR type=?) AND account_id=? AND related_account_id=?", uid, false, models.TRANSACTION_DB_TYPE_TRANSFER_OUT, models.TRANSACTION_DB_TYPE_TRANSFER_IN, toAccountId, toAccountId).Update(deletedModel)

		if err != nil {
			return err
		}

		if deletedRows > 0 {
			log.Infof(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" has deleted %d transactions which account id and related account id are both \"%d\"", uid, deletedRows, toAccountId)
		}

		// update account balance
		if fromAccount.Balance != 0 {
			newToAccountBalance, ok := utils.AddInt64(toAccount.Balance, fromAccount.Balance)

			if !ok {
				return errs.ErrAccountBalanceOverflow
			}

			toAccount.UpdatedUnixTime = time.Now().Unix()
			updatedRows, err := s.updateAccountBalance(sess, toAccount, fromAccount.Balance)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.MoveAllTransactionsBetweenAccounts] failed to update to account balance")
				return errs.ErrDatabaseOperationFailed
			}

			log.Infof(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" has updated account \"id:%d\" balance from %d to %d", uid, toAccountId, toAccount.Balance, newToAccountBalance)

			fromAccount.Balance = 0
			fromAccount.UpdatedUnixTime = time.Now().Unix()
			updatedRows, err = sess.ID(fromAccount.AccountId).Cols("balance", "updated_unix_time").Where("uid=? AND deleted=?", fromAccount.Uid, false).Update(fromAccount)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.MoveAllTransactionsBetweenAccounts] failed to update from account balance")
				return errs.ErrDatabaseOperationFailed
			}
		}

		return nil
	})
}
