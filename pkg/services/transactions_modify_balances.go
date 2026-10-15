package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"time"
	"xorm.io/xorm"
)

// updateModifiedTransactionBalances applies the old and new transaction's balance
// effects inside the caller's database transaction. Account objects may alias when
// the transaction keeps its accounts; preserve update ordering for those cases.
func (s *TransactionService) updateModifiedTransactionBalances(c core.Context, sess *xorm.Session, transaction, oldTransaction *models.Transaction, sourceAccount, destinationAccount, oldSourceAccount, oldDestinationAccount *models.Account) error {
	// Update account table
	if transaction.Type == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
		if transaction.AccountId != oldTransaction.AccountId {
			return errs.ErrBalanceModificationTransactionCannotChangeAccountId
		}

		if transaction.Amount != oldTransaction.Amount && transaction.RelatedAccountAmount != oldTransaction.RelatedAccountAmount {
			sourceAccount.UpdatedUnixTime = time.Now().Unix()
			updatedRows, err := s.updateAccountBalance(sess, sourceAccount, transaction.RelatedAccountAmount-oldTransaction.RelatedAccountAmount)

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
				return errs.ErrDatabaseOperationFailed
			}
		}
	} else if transaction.Type == models.TRANSACTION_DB_TYPE_INCOME {
		var oldAccountOldAmount = oldTransaction.Amount
		var oldAccountNewAmount int64 = 0
		var newAccountNewAmount int64 = 0

		if oldTransaction.Type == models.TRANSACTION_DB_TYPE_EXPENSE || oldTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			oldAccountOldAmount = -oldTransaction.Amount
		}

		if transaction.AccountId == oldTransaction.AccountId {
			oldAccountNewAmount = transaction.Amount
		} else if transaction.AccountId != oldTransaction.AccountId {
			newAccountNewAmount = transaction.Amount
		}

		if oldAccountNewAmount != oldAccountOldAmount {
			oldSourceAccount.UpdatedUnixTime = time.Now().Unix()
			updatedRows, err := s.updateAccountBalance(sess, oldSourceAccount, oldAccountNewAmount-oldAccountOldAmount)

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
				return errs.ErrDatabaseOperationFailed
			}
		}

		if newAccountNewAmount != 0 {
			sourceAccount.UpdatedUnixTime = time.Now().Unix()
			updatedRows, err := s.updateAccountBalance(sess, sourceAccount, newAccountNewAmount)

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
				return errs.ErrDatabaseOperationFailed
			}
		}

		if oldTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			oldDestinationAccount.UpdatedUnixTime = time.Now().Unix()
			updatedRows, err := s.updateAccountBalance(sess, oldDestinationAccount, -oldTransaction.RelatedAccountAmount)

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update related account balance, because %s", err.Error())
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update related account balance")
				return errs.ErrDatabaseOperationFailed
			}
		}
	} else if transaction.Type == models.TRANSACTION_DB_TYPE_EXPENSE {
		var oldAccountOldAmount = oldTransaction.Amount
		var oldAccountNewAmount int64 = 0
		var newAccountNewAmount int64 = 0

		if oldTransaction.Type == models.TRANSACTION_DB_TYPE_INCOME {
			oldAccountOldAmount = -oldTransaction.Amount
		}

		if transaction.AccountId == oldTransaction.AccountId {
			oldAccountNewAmount = transaction.Amount
		} else if transaction.AccountId != oldTransaction.AccountId {
			newAccountNewAmount = transaction.Amount
		}

		if oldAccountNewAmount != oldAccountOldAmount {
			oldSourceAccount.UpdatedUnixTime = time.Now().Unix()
			updatedRows, err := s.updateAccountBalance(sess, oldSourceAccount, oldAccountOldAmount-oldAccountNewAmount)

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
				return errs.ErrDatabaseOperationFailed
			}
		}

		if newAccountNewAmount != 0 {
			sourceAccount.UpdatedUnixTime = time.Now().Unix()
			updatedRows, err := s.updateAccountBalance(sess, sourceAccount, -newAccountNewAmount)

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
				return errs.ErrDatabaseOperationFailed
			}
		}

		if oldTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			oldDestinationAccount.UpdatedUnixTime = time.Now().Unix()
			updatedRows, err := s.updateAccountBalance(sess, oldDestinationAccount, -oldTransaction.RelatedAccountAmount)

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update related account balance, because %s", err.Error())
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update related account balance")
				return errs.ErrDatabaseOperationFailed
			}
		}
	} else if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
		var oldSourceAccountOldAmount = oldTransaction.Amount
		var oldSourceAccountNewAmount int64 = 0
		var newSourceAccountNewAmount int64 = 0

		if oldTransaction.Type == models.TRANSACTION_DB_TYPE_INCOME {
			oldSourceAccountOldAmount = -oldTransaction.Amount
		}

		if transaction.AccountId == oldTransaction.AccountId {
			oldSourceAccountNewAmount = transaction.Amount
		} else if transaction.AccountId != oldTransaction.AccountId {
			newSourceAccountNewAmount = transaction.Amount
		}

		if oldSourceAccountNewAmount != oldSourceAccountOldAmount {
			oldSourceAccount.UpdatedUnixTime = time.Now().Unix()
			updatedRows, err := s.updateAccountBalance(sess, oldSourceAccount, oldSourceAccountOldAmount-oldSourceAccountNewAmount)

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
				return errs.ErrDatabaseOperationFailed
			}
		}

		if newSourceAccountNewAmount != 0 {
			sourceAccount.UpdatedUnixTime = time.Now().Unix()
			updatedRows, err := s.updateAccountBalance(sess, sourceAccount, -newSourceAccountNewAmount)

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
				return errs.ErrDatabaseOperationFailed
			}
		}

		var oldDestinationAccountNewAmount int64 = 0
		var newDestinationAccountNewAmount int64 = 0

		if transaction.RelatedAccountId == oldTransaction.RelatedAccountId {
			oldDestinationAccountNewAmount = transaction.RelatedAccountAmount
		} else if transaction.RelatedAccountId != oldTransaction.RelatedAccountId {
			newDestinationAccountNewAmount = transaction.RelatedAccountAmount
		}

		if oldTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			if oldDestinationAccountNewAmount != oldTransaction.RelatedAccountAmount {
				oldDestinationAccount.UpdatedUnixTime = time.Now().Unix()
				updatedRows, err := s.updateAccountBalance(sess, oldDestinationAccount, oldDestinationAccountNewAmount-oldTransaction.RelatedAccountAmount)

				if err != nil {
					log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
					return err
				} else if updatedRows < 1 {
					log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
					return errs.ErrDatabaseOperationFailed
				}
			}
		}

		if newDestinationAccountNewAmount != 0 {
			destinationAccount.UpdatedUnixTime = time.Now().Unix()
			updatedRows, err := s.updateAccountBalance(sess, destinationAccount, newDestinationAccountNewAmount)

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
				return errs.ErrDatabaseOperationFailed
			}
		}
	} else if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		return errs.ErrTransactionTypeInvalid
	}
	return nil
}
