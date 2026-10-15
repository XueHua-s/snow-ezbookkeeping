package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"strings"
)

// GetAccountMapByList returns an account map by a list
func (s *AccountService) GetAccountMapByList(accounts []*models.Account) map[int64]*models.Account {
	accountMap := make(map[int64]*models.Account)

	for i := 0; i < len(accounts); i++ {
		account := accounts[i]
		accountMap[account.AccountId] = account
	}
	return accountMap
}

// GetVisibleAccountNameMapByList returns visible account map by a list
func (s *AccountService) GetVisibleAccountNameMapByList(accounts []*models.Account) map[string]*models.Account {
	accountMap := make(map[string]*models.Account)

	for i := 0; i < len(accounts); i++ {
		account := accounts[i]

		if account.Hidden {
			continue
		}

		if account.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
			continue
		}

		accountMap[account.Name] = account
	}
	return accountMap
}

// GetAccountNames returns a list with account names from account models list
func (s *AccountService) GetAccountNames(accounts []*models.Account) []string {
	accountNames := make([]string, len(accounts))

	for i := 0; i < len(accounts); i++ {
		accountNames[i] = accounts[i].Name
	}

	return accountNames
}

// GetAccountOrSubAccountIds returns a list of account ids or sub-account ids according to given account ids
func (s *AccountService) GetAccountOrSubAccountIds(c core.Context, accountIds string, uid int64) ([]int64, error) {
	if accountIds == "" || accountIds == "0" {
		return nil, nil
	}

	requestAccountIds, err := utils.StringArrayToInt64Array(strings.Split(accountIds, ","))

	if err != nil {
		return nil, errs.Or(err, errs.ErrAccountIdInvalid)
	}

	var allAccountIds []int64

	if len(requestAccountIds) > 0 {
		allSubAccounts, err := s.GetSubAccountsByAccountIds(c, uid, requestAccountIds)

		if err != nil {
			return nil, err
		}

		accountIdsMap := make(map[int64]int32, len(requestAccountIds))

		for i := 0; i < len(requestAccountIds); i++ {
			accountIdsMap[requestAccountIds[i]] = 0
		}

		for i := 0; i < len(allSubAccounts); i++ {
			subAccount := allSubAccounts[i]

			if refCount, exists := accountIdsMap[subAccount.ParentAccountId]; exists {
				accountIdsMap[subAccount.ParentAccountId] = refCount + 1
			} else {
				accountIdsMap[subAccount.ParentAccountId] = 1
			}

			if _, exists := accountIdsMap[subAccount.AccountId]; exists {
				delete(accountIdsMap, subAccount.AccountId)
			}

			allAccountIds = append(allAccountIds, subAccount.AccountId)
		}

		for accountId, refCount := range accountIdsMap {
			if refCount < 1 {
				allAccountIds = append(allAccountIds, accountId)
			}
		}
	}

	return allAccountIds, nil
}

// GetAccountOrSubAccountIdsByAccountName returns a list of account ids or sub-account ids according to given account name
func (s *AccountService) GetAccountOrSubAccountIdsByAccountName(accounts []*models.Account, accountName string) []int64 {
	accountIds := make([]int64, 0)
	parentAccountIds := make([]int64, 0)
	childAccountByParentAccountId := make(map[int64][]*models.Account)

	for i := 0; i < len(accounts); i++ {
		account := accounts[i]

		if account.Name == accountName {
			if account.Type == models.ACCOUNT_TYPE_SINGLE_ACCOUNT {
				accountIds = append(accountIds, account.AccountId)
			} else if account.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
				parentAccountIds = append(parentAccountIds, account.AccountId)
			}
		} else if account.ParentAccountId > 0 {
			childAccounts, exists := childAccountByParentAccountId[account.ParentAccountId]

			if !exists {
				childAccounts = make([]*models.Account, 0)
			}

			childAccounts = append(childAccounts, account)
			childAccountByParentAccountId[account.ParentAccountId] = childAccounts
		}
	}

	for i := 0; i < len(parentAccountIds); i++ {
		parentAccountId := parentAccountIds[i]

		if childAccounts, exists := childAccountByParentAccountId[parentAccountId]; exists {
			for j := 0; j < len(childAccounts); j++ {
				childAccount := childAccounts[j]
				accountIds = append(accountIds, childAccount.AccountId)
			}
		}
	}

	return accountIds
}
