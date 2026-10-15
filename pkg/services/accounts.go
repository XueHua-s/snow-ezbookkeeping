package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
	"strings"
)

// AccountService represents account service
type AccountService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a account service singleton instance
var (
	Accounts = &AccountService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetTotalAccountCountByUid returns total account count of user
func (s *AccountService) GetTotalAccountCountByUid(c core.Context, uid int64) (int64, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	count, err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).Count(&models.Account{})

	return count, err
}

// GetAllAccountsByUid returns all account models of user
func (s *AccountService) GetAllAccountsByUid(c core.Context, uid int64) ([]*models.Account, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var accounts []*models.Account
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).OrderBy("parent_account_id asc, display_order asc").Find(&accounts)

	return accounts, err
}

// GetAccountByAccountId returns account model according to account id
func (s *AccountService) GetAccountByAccountId(c core.Context, uid int64, accountId int64) (*models.Account, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if accountId <= 0 {
		return nil, errs.ErrAccountIdInvalid
	}

	account := &models.Account{}
	has, err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=? AND account_id=?", uid, false, accountId).Get(account)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrAccountNotFound
	}

	return account, err
}

// GetAccountAndSubAccountsByAccountId returns account model and sub-account models according to account id
func (s *AccountService) GetAccountAndSubAccountsByAccountId(c core.Context, uid int64, accountId int64) ([]*models.Account, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if accountId <= 0 {
		return nil, errs.ErrAccountIdInvalid
	}

	var accounts []*models.Account
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=? AND (account_id=? OR parent_account_id=?)", uid, false, accountId, accountId).OrderBy("parent_account_id asc, display_order asc").Find(&accounts)

	return accounts, err
}

// GetSubAccountsByAccountId returns sub-account models according to account id
func (s *AccountService) GetSubAccountsByAccountId(c core.Context, uid int64, accountId int64) ([]*models.Account, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if accountId <= 0 {
		return nil, errs.ErrAccountIdInvalid
	}

	var accounts []*models.Account
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=? AND parent_account_id=?", uid, false, accountId).OrderBy("display_order asc").Find(&accounts)

	return accounts, err
}

// GetSubAccountsByAccountIds returns sub-account models according to account ids
func (s *AccountService) GetSubAccountsByAccountIds(c core.Context, uid int64, accountIds []int64) ([]*models.Account, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if len(accountIds) <= 0 {
		return nil, errs.ErrAccountIdInvalid
	}

	condition := "uid=? AND deleted=?"
	conditionParams := make([]any, 0, len(accountIds)+2)
	conditionParams = append(conditionParams, uid)
	conditionParams = append(conditionParams, false)

	var accountIdConditions strings.Builder

	for i := 0; i < len(accountIds); i++ {
		if accountIds[i] <= 0 {
			return nil, errs.ErrAccountIdInvalid
		}

		if accountIdConditions.Len() > 0 {
			accountIdConditions.WriteString(",")
		}

		accountIdConditions.WriteString("?")
		conditionParams = append(conditionParams, accountIds[i])
	}

	if accountIdConditions.Len() > 1 {
		condition = condition + " AND parent_account_id IN (" + accountIdConditions.String() + ")"
	} else {
		condition = condition + " AND parent_account_id = " + accountIdConditions.String()
	}

	var accounts []*models.Account
	err := s.UserDataDB(uid).NewSession(c).Where(condition, conditionParams...).OrderBy("display_order asc").Find(&accounts)

	return accounts, err
}

// GetAccountsByAccountIds returns account models according to account ids
func (s *AccountService) GetAccountsByAccountIds(c core.Context, uid int64, accountIds []int64) (map[int64]*models.Account, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if accountIds == nil {
		return nil, errs.ErrAccountIdInvalid
	}

	var accounts []*models.Account
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).In("account_id", accountIds).Find(&accounts)

	if err != nil {
		return nil, err
	}

	accountMap := s.GetAccountMapByList(accounts)
	return accountMap, err
}

// GetMaxDisplayOrder returns the max display order according to account category
func (s *AccountService) GetMaxDisplayOrder(c core.Context, uid int64, category models.AccountCategory) (int32, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	account := &models.Account{}
	has, err := s.UserDataDB(uid).NewSession(c).Cols("uid", "deleted", "parent_account_id", "display_order").Where("uid=? AND deleted=? AND parent_account_id=? AND category=?", uid, false, models.LevelOneAccountParentId, category).OrderBy("display_order desc").Limit(1).Get(account)

	if err != nil {
		return 0, err
	}

	if has {
		return account.DisplayOrder, nil
	} else {
		return 0, nil
	}
}

// GetMaxSubAccountDisplayOrder returns the max display order of sub-account according to account category and parent account id
func (s *AccountService) GetMaxSubAccountDisplayOrder(c core.Context, uid int64, category models.AccountCategory, parentAccountId int64) (int32, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	if parentAccountId <= 0 {
		return 0, errs.ErrAccountIdInvalid
	}

	account := &models.Account{}
	has, err := s.UserDataDB(uid).NewSession(c).Cols("uid", "deleted", "parent_account_id", "display_order").Where("uid=? AND deleted=? AND parent_account_id=? AND category=?", uid, false, parentAccountId, category).OrderBy("display_order desc").Limit(1).Get(account)

	if err != nil {
		return 0, err
	}

	if has {
		return account.DisplayOrder, nil
	} else {
		return 0, nil
	}
}
