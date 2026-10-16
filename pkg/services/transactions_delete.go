package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"time"
	"xorm.io/xorm"
)

// DeleteTransaction deletes an existed transaction from database
func (s *TransactionService) DeleteTransaction(c core.Context, uid int64, transactionId int64) error {
	return s.deleteTransaction(c, uid, transactionId, nil)
}

func (s *TransactionService) deleteTransaction(c core.Context, uid int64, transactionId int64, clear *assistantLedgerClear) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.Transaction{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	tagIndexUpdateModel := &models.TransactionTagIndex{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	pictureUpdateModel := &models.TransactionPictureInfo{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.withLedgerClearTransaction(c, uid, clear, func(sess *xorm.Session) error {
		// Get and verify current transaction
		oldTransaction := &models.Transaction{}
		has, err := sess.ID(transactionId).Where("uid=? AND deleted=?", uid, false).Get(oldTransaction)

		if err != nil {
			return err
		} else if !has {
			return errs.ErrTransactionNotFound
		}

		// Get and verify source and destination account
		sourceAccount, destinationAccount, err := s.getAccountModels(sess, oldTransaction)

		if err != nil {
			return err
		}

		if sourceAccount.Hidden || (destinationAccount != nil && destinationAccount.Hidden) {
			return errs.ErrCannotDeleteTransactionInHiddenAccount
		}

		if sourceAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS || (destinationAccount != nil && destinationAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS) {
			return errs.ErrCannotDeleteTransactionInParentAccount
		}

		// Update transaction row to deleted
		deletedRows, err := sess.ID(oldTransaction.TransactionId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrTransactionNotFound
		}

		if oldTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || oldTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			deletedRows, err = sess.ID(oldTransaction.RelatedId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

			if err != nil {
				return err
			} else if deletedRows < 1 {
				return errs.ErrTransactionNotFound
			}
		}

		// Update transaction tag index
		_, err = sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=? AND transaction_id=?", uid, false, oldTransaction.TransactionId).Update(tagIndexUpdateModel)

		if err != nil {
			return err
		}

		// Update transaction picture
		_, err = sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=? AND transaction_id=?", uid, false, oldTransaction.TransactionId).Update(pictureUpdateModel)

		if err != nil {
			return err
		}

		// Update account table
		if oldTransaction.Type == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
			if oldTransaction.RelatedAccountAmount != 0 {
				sourceAccount.UpdatedUnixTime = time.Now().Unix()
				updatedRows, err := s.updateAccountBalance(sess, sourceAccount, -oldTransaction.RelatedAccountAmount)

				if err != nil {
					return err
				} else if updatedRows < 1 {
					log.Errorf(c, "[transactions.DeleteTransaction] failed to update account balance")
					return errs.ErrDatabaseOperationFailed
				}
			}
		} else if oldTransaction.Type == models.TRANSACTION_DB_TYPE_INCOME {
			if oldTransaction.Amount != 0 {
				sourceAccount.UpdatedUnixTime = time.Now().Unix()
				updatedRows, err := s.updateAccountBalance(sess, sourceAccount, -oldTransaction.Amount)

				if err != nil {
					return err
				} else if updatedRows < 1 {
					log.Errorf(c, "[transactions.DeleteTransaction] failed to update account balance")
					return errs.ErrDatabaseOperationFailed
				}
			}
		} else if oldTransaction.Type == models.TRANSACTION_DB_TYPE_EXPENSE {
			if oldTransaction.Amount != 0 {
				sourceAccount.UpdatedUnixTime = time.Now().Unix()
				updatedRows, err := s.updateAccountBalance(sess, sourceAccount, oldTransaction.Amount)

				if err != nil {
					return err
				} else if updatedRows < 1 {
					log.Errorf(c, "[transactions.DeleteTransaction] failed to update account balance")
					return errs.ErrDatabaseOperationFailed
				}
			}
		} else if oldTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			if oldTransaction.Amount != 0 {
				sourceAccount.UpdatedUnixTime = time.Now().Unix()
				updatedSourceRows, err := s.updateAccountBalance(sess, sourceAccount, oldTransaction.Amount)

				if err != nil {
					return err
				} else if updatedSourceRows < 1 {
					log.Errorf(c, "[transactions.DeleteTransaction] failed to update account balance")
					return errs.ErrDatabaseOperationFailed
				}
			}

			if oldTransaction.RelatedAccountAmount != 0 {
				destinationAccount.UpdatedUnixTime = time.Now().Unix()
				updatedDestinationRows, err := s.updateAccountBalance(sess, destinationAccount, -oldTransaction.RelatedAccountAmount)

				if err != nil {
					return err
				} else if updatedDestinationRows < 1 {
					log.Errorf(c, "[transactions.DeleteTransaction] failed to update related account balance")
					return errs.ErrDatabaseOperationFailed
				}
			}
		} else if oldTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			return errs.ErrTransactionTypeInvalid
		}

		return err
	})
}

// DeleteAllTransactions deletes all existed transactions from database
func (s *TransactionService) DeleteAllTransactions(c core.Context, uid int64, deleteAccount bool, assistant *AssistantStorageService) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.Transaction{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	tagIndexUpdateModel := &models.TransactionTagIndex{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	pictureUpdateModel := &models.TransactionPictureInfo{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	accountUpdateModel := &models.Account{
		Balance:         0,
		Deleted:         deleteAccount,
		DeletedUnixTime: now,
	}

	return s.withLedgerClearTransaction(c, uid, newAssistantLedgerClear(assistant, uid), func(sess *xorm.Session) error {
		// Update all transactions to deleted
		_, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		}

		// Update all transaction tag index to deleted
		_, err = sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(tagIndexUpdateModel)

		if err != nil {
			return err
		}

		// Update all transaction pictures to deleted
		_, err = sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(pictureUpdateModel)

		if err != nil {
			return err
		}

		// Update all accounts to deleted or set amount to zero
		_, err = sess.Cols("balance", "deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(accountUpdateModel)

		if err != nil {
			return err
		}

		return nil
	})
}

// DeleteAllTransactionsOfAccount deletes all existed transactions of specific account from database
func (s *TransactionService) DeleteAllTransactionsOfAccount(c core.Context, uid int64, accountId int64, pageCount int32, assistant *AssistantStorageService) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if accountId <= 0 {
		return errs.ErrAccountIdInvalid
	}

	transactions, err := s.GetAllSpecifiedTransactions(c, uid, 0, 0, 0, nil, []int64{accountId}, nil, false, "", "", core.MATCH_MODE_DEFAULT, false, pageCount, true)

	if err != nil {
		return err
	}

	if len(transactions) < 1 {
		return nil
	}

	clear := newAssistantLedgerClear(assistant, uid)
	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			err = s.deleteTransaction(c, uid, transaction.RelatedId, clear)
		} else {
			err = s.deleteTransaction(c, uid, transaction.TransactionId, clear)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
