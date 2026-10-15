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

// CreateAccounts saves a new account model to database
func (s *AccountService) CreateAccounts(c core.Context, mainAccount *models.Account, mainAccountBalanceTime int64, childrenAccounts []*models.Account, childrenAccountBalanceTimes []int64, clientTimezone *time.Location) error {
	if mainAccount.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	needAccountUuidCount := uint16(len(childrenAccounts) + 1)
	accountUuids := s.GenerateUuids(uuid.UUID_TYPE_ACCOUNT, needAccountUuidCount)

	if len(accountUuids) < int(needAccountUuidCount) {
		return errs.ErrSystemIsBusy
	}

	now := time.Now().Unix()

	allAccounts := make([]*models.Account, len(childrenAccounts)+1)
	var allInitTransactions []*models.Transaction

	mainAccount.AccountId = accountUuids[0]
	allAccounts[0] = mainAccount

	if mainAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
		for i := 0; i < len(childrenAccounts); i++ {
			childAccount := childrenAccounts[i]
			childAccount.AccountId = accountUuids[i+1]
			childAccount.ParentAccountId = mainAccount.AccountId
			childAccount.Uid = mainAccount.Uid
			childAccount.Type = models.ACCOUNT_TYPE_SINGLE_ACCOUNT

			allAccounts[i+1] = childrenAccounts[i]
		}
	}

	defaultTransactionTime := utils.GetMinTransactionTimeFromUnixTime(now)

	for i := 0; i < len(allAccounts); i++ {
		allAccounts[i].Deleted = false
		allAccounts[i].CreatedUnixTime = now
		allAccounts[i].UpdatedUnixTime = now

		if allAccounts[i].Balance != 0 {
			transactionId := s.GenerateUuid(uuid.UUID_TYPE_TRANSACTION)

			if transactionId < 1 {
				return errs.ErrSystemIsBusy
			}

			transactionTime := defaultTransactionTime
			transactionUtcOffset := utils.GetTimezoneOffsetMinutes(now, clientTimezone)

			if i == 0 && mainAccountBalanceTime > 0 {
				transactionTime = utils.GetMinTransactionTimeFromUnixTime(mainAccountBalanceTime)
				transactionUtcOffset = utils.GetTimezoneOffsetMinutes(mainAccountBalanceTime, clientTimezone)
			} else if i > 0 && len(childrenAccountBalanceTimes) > i-1 && childrenAccountBalanceTimes[i-1] > 0 {
				transactionTime = utils.GetMinTransactionTimeFromUnixTime(childrenAccountBalanceTimes[i-1])
				transactionUtcOffset = utils.GetTimezoneOffsetMinutes(childrenAccountBalanceTimes[i-1], clientTimezone)
			} else {
				defaultTransactionTime++
			}

			newTransaction := &models.Transaction{
				TransactionId:        transactionId,
				Uid:                  allAccounts[i].Uid,
				Deleted:              false,
				Type:                 models.TRANSACTION_DB_TYPE_MODIFY_BALANCE,
				TransactionTime:      transactionTime,
				TimezoneUtcOffset:    transactionUtcOffset,
				AccountId:            allAccounts[i].AccountId,
				Amount:               allAccounts[i].Balance,
				RelatedAccountId:     allAccounts[i].AccountId,
				RelatedAccountAmount: allAccounts[i].Balance,
				CreatedUnixTime:      now,
				UpdatedUnixTime:      now,
			}

			allInitTransactions = append(allInitTransactions, newTransaction)
		}
	}

	userDataDb := s.UserDataDB(mainAccount.Uid)

	return userDataDb.DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(allAccounts); i++ {
			account := allAccounts[i]
			_, err := sess.Insert(account)

			if err != nil {
				return err
			}
		}

		for i := 0; i < len(allInitTransactions); i++ {
			transaction := allInitTransactions[i]

			insertTransactionSavePointName := "insert_transaction"
			err := userDataDb.SetSavePoint(sess, insertTransactionSavePointName)

			if err != nil {
				log.Errorf(c, "[accounts.CreateAccounts] failed to set save point \"%s\", because %s", insertTransactionSavePointName, err.Error())
				return err
			}

			createdRows, err := sess.Insert(transaction)

			if err != nil || createdRows < 1 { // maybe another transaction has same time
				if err != nil {
					log.Warnf(c, "[accounts.CreateAccounts] cannot create transaction, because %s, regenerate transaction time value", err.Error())
				} else {
					log.Warnf(c, "[accounts.CreateAccounts] cannot create transaction, regenerate transaction time value")
				}

				err = userDataDb.RollbackToSavePoint(sess, insertTransactionSavePointName)

				if err != nil {
					log.Errorf(c, "[accounts.CreateAccounts] failed to rollback to save point \"%s\", because %s", insertTransactionSavePointName, err.Error())
					return err
				}

				sameSecondLatestTransaction := &models.Transaction{}
				minTransactionTime := utils.GetMinTransactionTimeFromUnixTime(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime))
				maxTransactionTime := utils.GetMaxTransactionTimeFromUnixTime(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime))

				has, err := sess.Where("uid=? AND transaction_time>=? AND transaction_time<=?", transaction.Uid, minTransactionTime, maxTransactionTime).OrderBy("transaction_time desc").Limit(1).Get(sameSecondLatestTransaction)

				if err != nil {
					return err
				} else if !has {
					log.Errorf(c, "[accounts.CreateAccounts] it should have transactions in %d - %d, but result is empty", minTransactionTime, maxTransactionTime)
					return errs.ErrDatabaseOperationFailed
				} else if sameSecondLatestTransaction.TransactionTime == maxTransactionTime-1 {
					return errs.ErrTooMuchTransactionInOneSecond
				}

				transaction.TransactionTime = sameSecondLatestTransaction.TransactionTime + 1
				createdRows, err := sess.Insert(transaction)

				if err != nil {
					log.Errorf(c, "[accounts.CreateAccounts] failed to add transaction again, because %s", err.Error())
					return err
				} else if createdRows < 1 {
					log.Errorf(c, "[accounts.CreateAccounts] failed to add transaction again")
					return errs.ErrDatabaseOperationFailed
				}
			}
		}

		return nil
	})
}
