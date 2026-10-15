package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
	"time"
	"xorm.io/xorm"
)

// ModifyAccounts saves an existed account model to database
func (s *AccountService) ModifyAccounts(c core.Context, mainAccount *models.Account, updateAccounts []*models.Account, addSubAccounts []*models.Account, addSubAccountBalanceTimes []int64, removeSubAccountIds []int64, updateMainAccountCurrency bool, clientTimezone *time.Location) error {
	if mainAccount.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	needAccountUuidCount := uint16(len(addSubAccounts))
	newAccountUuids := s.GenerateUuids(uuid.UUID_TYPE_ACCOUNT, needAccountUuidCount)

	if len(newAccountUuids) < int(needAccountUuidCount) {
		return errs.ErrSystemIsBusy
	}

	now := time.Now().Unix()

	var addInitTransactions []*models.Transaction

	for i := 0; i < len(updateAccounts); i++ {
		updateAccounts[i].UpdatedUnixTime = now
	}

	if mainAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
		defaultTransactionTime := utils.GetMinTransactionTimeFromUnixTime(now)

		for i := 0; i < len(addSubAccounts); i++ {
			childAccount := addSubAccounts[i]
			childAccount.AccountId = newAccountUuids[i]
			childAccount.ParentAccountId = mainAccount.AccountId
			childAccount.Uid = mainAccount.Uid
			childAccount.Type = models.ACCOUNT_TYPE_SINGLE_ACCOUNT
			childAccount.Deleted = false
			childAccount.CreatedUnixTime = now
			childAccount.UpdatedUnixTime = now

			if childAccount.Balance != 0 {
				transactionId := s.GenerateUuid(uuid.UUID_TYPE_TRANSACTION)

				if transactionId < 1 {
					return errs.ErrSystemIsBusy
				}

				transactionTime := defaultTransactionTime
				transactionUtcOffset := utils.GetTimezoneOffsetMinutes(now, clientTimezone)

				if len(addSubAccountBalanceTimes) > i && addSubAccountBalanceTimes[i] > 0 {
					transactionTime = utils.GetMinTransactionTimeFromUnixTime(addSubAccountBalanceTimes[i])
					transactionUtcOffset = utils.GetTimezoneOffsetMinutes(addSubAccountBalanceTimes[i], clientTimezone)
				} else {
					defaultTransactionTime++
				}

				newTransaction := &models.Transaction{
					TransactionId:        transactionId,
					Uid:                  childAccount.Uid,
					Deleted:              false,
					Type:                 models.TRANSACTION_DB_TYPE_MODIFY_BALANCE,
					TransactionTime:      transactionTime,
					TimezoneUtcOffset:    transactionUtcOffset,
					AccountId:            childAccount.AccountId,
					Amount:               childAccount.Balance,
					RelatedAccountId:     childAccount.AccountId,
					RelatedAccountAmount: childAccount.Balance,
					CreatedUnixTime:      now,
					UpdatedUnixTime:      now,
				}

				addInitTransactions = append(addInitTransactions, newTransaction)
			}
		}
	}

	userDataDb := s.UserDataDB(mainAccount.Uid)

	return userDataDb.DoTransaction(c, func(sess *xorm.Session) error {
		// update accounts
		for i := 0; i < len(updateAccounts); i++ {
			account := updateAccounts[i]
			updateColumns := []string{"name", "display_order", "category", "icon", "icon_type", "color", "comment", "extend", "hidden", "updated_unix_time"}

			if updateMainAccountCurrency && account.AccountId == mainAccount.AccountId {
				updateColumns = append(updateColumns, "currency")
			}

			updatedRows, err := sess.ID(account.AccountId).Cols(updateColumns...).Where("uid=? AND deleted=?", account.Uid, false).Update(account)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				return errs.ErrAccountNotFound
			}
		}

		// add new sub accounts
		for i := 0; i < len(addSubAccounts); i++ {
			account := addSubAccounts[i]
			_, err := sess.Insert(account)

			if err != nil {
				return err
			}
		}

		// add init transaction for new sub accounts
		for i := 0; i < len(addInitTransactions); i++ {
			transaction := addInitTransactions[i]

			insertTransactionSavePointName := "insert_transaction"
			err := userDataDb.SetSavePoint(sess, insertTransactionSavePointName)

			if err != nil {
				log.Errorf(c, "[accounts.ModifyAccounts] failed to set save point \"%s\", because %s", insertTransactionSavePointName, err.Error())
				return err
			}

			createdRows, err := sess.Insert(transaction)

			if err != nil || createdRows < 1 { // maybe another transaction has same time
				if err != nil {
					log.Warnf(c, "[accounts.ModifyAccounts] cannot create trasaction, because %s, regenerate transaction time value", err.Error())
				} else {
					log.Warnf(c, "[accounts.ModifyAccounts] cannot create trasaction, regenerate transaction time value")
				}

				err = userDataDb.RollbackToSavePoint(sess, insertTransactionSavePointName)

				if err != nil {
					log.Errorf(c, "[accounts.ModifyAccounts] failed to rollback to save point \"%s\", because %s", insertTransactionSavePointName, err.Error())
					return err
				}

				sameSecondLatestTransaction := &models.Transaction{}
				minTransactionTime := utils.GetMinTransactionTimeFromUnixTime(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime))
				maxTransactionTime := utils.GetMaxTransactionTimeFromUnixTime(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime))

				has, err := sess.Where("uid=? AND transaction_time>=? AND transaction_time<=?", transaction.Uid, minTransactionTime, maxTransactionTime).OrderBy("transaction_time desc").Limit(1).Get(sameSecondLatestTransaction)

				if err != nil {
					return err
				} else if !has {
					log.Errorf(c, "[accounts.ModifyAccounts] it should have transactions in %d - %d, but result is empty", minTransactionTime, maxTransactionTime)
					return errs.ErrDatabaseOperationFailed
				} else if sameSecondLatestTransaction.TransactionTime == maxTransactionTime-1 {
					return errs.ErrTooMuchTransactionInOneSecond
				}

				transaction.TransactionTime = sameSecondLatestTransaction.TransactionTime + 1
				createdRows, err := sess.Insert(transaction)

				if err != nil {
					log.Errorf(c, "[accounts.ModifyAccounts] failed to add transaction again, because %s", err.Error())
					return err
				} else if createdRows < 1 {
					log.Errorf(c, "[accounts.ModifyAccounts] failed to add transaction again")
					return errs.ErrDatabaseOperationFailed
				}
			}
		}

		// remove sub accounts
		if len(removeSubAccountIds) > 0 {
			subAccountsCount, err := sess.Where("uid=? AND deleted=? AND parent_account_id=?", mainAccount.Uid, false, mainAccount.AccountId).Count(&models.Account{})

			if subAccountsCount <= int64(len(removeSubAccountIds)) {
				return errs.ErrAccountHaveNoSubAccount
			}

			var relatedTransactionsByAccount []*models.Transaction
			err = sess.Cols("transaction_id", "uid", "deleted", "account_id", "type").Where("uid=? AND deleted=?", mainAccount.Uid, false).In("account_id", removeSubAccountIds).Limit(len(removeSubAccountIds) + 1).Find(&relatedTransactionsByAccount)

			if err != nil {
				return err
			} else if len(relatedTransactionsByAccount) > len(removeSubAccountIds) {
				return errs.ErrSubAccountInUseCannotBeDeleted
			} else if len(relatedTransactionsByAccount) > 0 {
				accountTransactionExists := make(map[int64]bool)

				for i := 0; i < len(relatedTransactionsByAccount); i++ {
					transaction := relatedTransactionsByAccount[i]

					if transaction.Type != models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
						return errs.ErrAccountInUseCannotBeDeleted
					} else if _, exists := accountTransactionExists[transaction.AccountId]; exists {
						return errs.ErrAccountInUseCannotBeDeleted
					}

					accountTransactionExists[transaction.AccountId] = true
				}
			}

			deleteAccountUpdateModel := &models.Account{
				Balance:         0,
				Deleted:         true,
				DeletedUnixTime: now,
			}

			deletedRows, err := sess.Cols("balance", "deleted", "deleted_unix_time").Where("uid=? AND deleted=?", mainAccount.Uid, false).In("account_id", removeSubAccountIds).Update(deleteAccountUpdateModel)

			if err != nil {
				return err
			} else if deletedRows < 1 {
				return errs.ErrSubAccountNotFound
			}

			if len(relatedTransactionsByAccount) > 0 {
				updateTransaction := &models.Transaction{
					Deleted:         true,
					DeletedUnixTime: now,
				}

				transactionIds := make([]int64, len(relatedTransactionsByAccount))

				for i := 0; i < len(relatedTransactionsByAccount); i++ {
					transactionIds[i] = relatedTransactionsByAccount[i].TransactionId
				}

				deletedTransactionRows, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", mainAccount.Uid, false).In("transaction_id", transactionIds).Update(updateTransaction)

				if err != nil {
					return err
				} else if deletedTransactionRows < int64(len(transactionIds)) {
					log.Errorf(c, "[accounts.ModifyAccounts] it should delete %d transactions, but have deleted %d actually", len(transactionIds), deletedTransactionRows)
					return errs.ErrDatabaseOperationFailed
				}
			}
		}

		return nil
	})
}

// UpdateAccountExtend updates extend field of given account
func (s *AccountService) UpdateAccountExtend(c core.Context, uid int64, account *models.Account) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	account.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(account.AccountId).Cols("extend", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(account)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrAccountNotFound
		}

		return nil
	})
}

// HideAccount updates hidden field of given accounts
func (s *AccountService) HideAccount(c core.Context, uid int64, ids []int64, hidden bool) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.Account{
		Hidden:          hidden,
		UpdatedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.Cols("hidden", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).In("account_id", ids).Update(updateModel)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrAccountNotFound
		}

		return nil
	})
}

// ModifyAccountDisplayOrders updates display order of given accounts
func (s *AccountService) ModifyAccountDisplayOrders(c core.Context, uid int64, accounts []*models.Account) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	for i := 0; i < len(accounts); i++ {
		accounts[i].UpdatedUnixTime = time.Now().Unix()
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(accounts); i++ {
			account := accounts[i]
			updatedRows, err := sess.ID(account.AccountId).Cols("display_order", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(account)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				return errs.ErrAccountNotFound
			}
		}

		return nil
	})
}
