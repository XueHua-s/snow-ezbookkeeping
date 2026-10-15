package services

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"strings"
	"time"
	"xorm.io/xorm"
)

// DeleteAccount deletes an existed account from database
func (s *AccountService) DeleteAccount(c core.Context, uid int64, accountId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.Account{
		Balance:         0,
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		var accountAndSubAccounts []*models.Account
		err := sess.Where("uid=? AND deleted=? AND ((account_id=? AND parent_account_id=?) OR parent_account_id=?)", uid, false, accountId, models.LevelOneAccountParentId, accountId).Find(&accountAndSubAccounts)

		if err != nil {
			return err
		} else if len(accountAndSubAccounts) < 1 {
			return errs.ErrAccountNotFound
		}

		var accountAndSubAccountIdsConditions strings.Builder
		accountAndSubAccountIds := make([]int64, len(accountAndSubAccounts))

		for i := 0; i < len(accountAndSubAccounts); i++ {
			if accountAndSubAccountIdsConditions.Len() > 0 {
				accountAndSubAccountIdsConditions.WriteString(",")
			}

			accountAndSubAccountIdsConditions.WriteString("?")
			accountAndSubAccountIds[i] = accountAndSubAccounts[i].AccountId
		}

		var relatedTransactionsByAccount []*models.Transaction
		err = sess.Cols("transaction_id", "uid", "deleted", "account_id", "type").Where("uid=? AND deleted=?", uid, false).In("account_id", accountAndSubAccountIds).Limit(len(accountAndSubAccounts) + 1).Find(&relatedTransactionsByAccount)

		if err != nil {
			return err
		} else if len(relatedTransactionsByAccount) > len(accountAndSubAccountIds) {
			return errs.ErrAccountInUseCannotBeDeleted
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

		transactionTemplateQueryCondition := fmt.Sprintf("uid=? AND deleted=? AND (template_type=? OR (template_type=? AND scheduled_frequency_type<>? AND (scheduled_end_time IS NULL OR scheduled_end_time>=?))) AND (account_id IN (%s) OR related_account_id IN (%s))", accountAndSubAccountIdsConditions.String(), accountAndSubAccountIdsConditions.String())
		transactionTemplateQueryConditionParams := make([]any, 0, len(accountAndSubAccountIds)*2+6)
		transactionTemplateQueryConditionParams = append(transactionTemplateQueryConditionParams, uid)
		transactionTemplateQueryConditionParams = append(transactionTemplateQueryConditionParams, false)
		transactionTemplateQueryConditionParams = append(transactionTemplateQueryConditionParams, models.TRANSACTION_TEMPLATE_TYPE_NORMAL)
		transactionTemplateQueryConditionParams = append(transactionTemplateQueryConditionParams, models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE)
		transactionTemplateQueryConditionParams = append(transactionTemplateQueryConditionParams, models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED)
		transactionTemplateQueryConditionParams = append(transactionTemplateQueryConditionParams, now)

		for i := 0; i < len(accountAndSubAccountIds); i++ {
			transactionTemplateQueryConditionParams = append(transactionTemplateQueryConditionParams, accountAndSubAccountIds[i])
		}

		for i := 0; i < len(accountAndSubAccountIds); i++ {
			transactionTemplateQueryConditionParams = append(transactionTemplateQueryConditionParams, accountAndSubAccountIds[i])
		}

		exists, err := sess.Cols("uid", "deleted", "account_id", "related_account_id", "template_type", "scheduled_frequency_type", "scheduled_end_time").Where(transactionTemplateQueryCondition, transactionTemplateQueryConditionParams...).Limit(1).Exist(&models.TransactionTemplate{})

		if err != nil {
			return err
		} else if exists {
			return errs.ErrAccountInUseCannotBeDeleted
		}

		deletedRows, err := sess.Cols("balance", "deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).In("account_id", accountAndSubAccountIds).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrAccountNotFound
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

			deletedTransactionRows, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).In("transaction_id", transactionIds).Update(updateTransaction)

			if err != nil {
				return err
			} else if deletedTransactionRows < int64(len(transactionIds)) {
				log.Errorf(c, "[accounts.DeleteAccount] it should delete %d transactions, but have deleted %d actually", len(transactionIds), deletedTransactionRows)
				return errs.ErrDatabaseOperationFailed
			}
		}

		return err
	})
}

// DeleteSubAccount deletes an existed sub-account from database
func (s *AccountService) DeleteSubAccount(c core.Context, uid int64, accountId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.Account{
		Balance:         0,
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		account := &models.Account{}
		has, err := sess.Cols("account_id", "uid", "deleted", "parent_account_id").Where("uid=? AND deleted=? AND account_id=? AND parent_account_id<>?", uid, false, accountId, models.LevelOneAccountParentId).Limit(1).Get(account)

		if err != nil {
			return err
		} else if !has {
			return errs.ErrSubAccountNotFound
		}

		subAccountsCount, err := sess.Where("uid=? AND deleted=? AND parent_account_id=?", uid, false, account.ParentAccountId).Count(&models.Account{})

		if subAccountsCount <= 1 {
			return errs.ErrAccountHaveNoSubAccount
		}

		var relatedTransactionsByAccount []*models.Transaction
		err = sess.Cols("transaction_id", "uid", "deleted", "account_id", "type").Where("uid=? AND deleted=? AND account_id=?", uid, false, accountId).Limit(2).Find(&relatedTransactionsByAccount)

		if err != nil {
			return err
		} else if len(relatedTransactionsByAccount) > 1 {
			return errs.ErrSubAccountInUseCannotBeDeleted
		} else if len(relatedTransactionsByAccount) > 0 {
			for i := 0; i < len(relatedTransactionsByAccount); i++ {
				transaction := relatedTransactionsByAccount[i]

				if transaction.Type != models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
					return errs.ErrSubAccountInUseCannotBeDeleted
				}
			}
		}

		exists, err := sess.Cols("uid", "deleted", "account_id", "related_account_id", "template_type", "scheduled_frequency_type", "scheduled_end_time").Where("uid=? AND deleted=? AND (template_type=? OR (template_type=? AND scheduled_frequency_type<>? AND (scheduled_end_time IS NULL OR scheduled_end_time>=?))) AND (account_id=? OR related_account_id=?)", uid, false, models.TRANSACTION_TEMPLATE_TYPE_NORMAL, models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE, models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED, now, accountId, accountId).Limit(1).Exist(&models.TransactionTemplate{})

		if err != nil {
			return err
		} else if exists {
			return errs.ErrSubAccountInUseCannotBeDeleted
		}

		deletedRows, err := sess.Cols("balance", "deleted", "deleted_unix_time").Where("uid=? AND deleted=? AND account_id=?", uid, false, accountId).Update(updateModel)

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

			deletedTransactionRows, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).In("transaction_id", transactionIds).Update(updateTransaction)

			if err != nil {
				return err
			} else if deletedTransactionRows < int64(len(transactionIds)) {
				log.Errorf(c, "[accounts.DeleteSubAccount] it should delete %d transactions, but have deleted %d actually", len(transactionIds), deletedTransactionRows)
				return errs.ErrDatabaseOperationFailed
			}
		}

		return err
	})
}
