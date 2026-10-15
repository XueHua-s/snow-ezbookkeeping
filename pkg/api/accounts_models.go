package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func (a *AccountsApi) createNewAccountModel(uid int64, accountCreateReq *models.AccountCreateRequest, balance int64, creditLimitForCreditCard *int64, isSubAccount bool, order int32) *models.Account {
	accountExtend := &models.AccountExtend{}

	if !isSubAccount && accountCreateReq.Category == models.ACCOUNT_CATEGORY_CREDIT_CARD {
		accountExtend.CreditCardStatementDate = &accountCreateReq.CreditCardStatementDate
		accountExtend.CreditCardLimit = creditLimitForCreditCard
	}

	return &models.Account{
		Uid:          uid,
		Name:         accountCreateReq.Name,
		DisplayOrder: order,
		Category:     accountCreateReq.Category,
		Type:         accountCreateReq.Type,
		Icon:         accountCreateReq.Icon,
		IconType:     accountCreateReq.IconType,
		Color:        accountCreateReq.Color,
		Currency:     accountCreateReq.Currency,
		Balance:      balance,
		Comment:      accountCreateReq.Comment,
		Extend:       accountExtend,
	}
}

func (a *AccountsApi) createNewSubAccountModelForModify(uid int64, accountType models.AccountType, accountModifyReq *models.AccountModifyRequest, balance int64, order int32) *models.Account {
	accountExtend := &models.AccountExtend{}

	return &models.Account{
		Uid:          uid,
		Name:         accountModifyReq.Name,
		DisplayOrder: order,
		Category:     accountModifyReq.Category,
		Type:         accountType,
		Icon:         accountModifyReq.Icon,
		IconType:     accountModifyReq.IconType,
		Color:        accountModifyReq.Color,
		Currency:     *accountModifyReq.Currency,
		Balance:      balance,
		Comment:      accountModifyReq.Comment,
		Extend:       accountExtend,
	}
}

func (a *AccountsApi) createSubAccountModels(uid int64, accountCreateReq *models.AccountCreateRequest, balances []int64) ([]*models.Account, []int64) {
	if len(accountCreateReq.SubAccounts) <= 0 {
		return nil, nil
	}

	childrenAccounts := make([]*models.Account, len(accountCreateReq.SubAccounts))
	childrenAccountBalanceTimes := make([]int64, len(accountCreateReq.SubAccounts))

	for i := int32(0); i < int32(len(accountCreateReq.SubAccounts)); i++ {
		childrenAccounts[i] = a.createNewAccountModel(uid, accountCreateReq.SubAccounts[i], balances[i], nil, true, i+1)
		childrenAccountBalanceTimes[i] = accountCreateReq.SubAccounts[i].BalanceTime
	}

	return childrenAccounts, childrenAccountBalanceTimes
}

func (a *AccountsApi) getToUpdateAccount(user *models.User, accountModifyReq *models.AccountModifyRequest, creditLimitForCreditCard *int64, oldAccount *models.Account, isSubAccount bool) (*models.Account, error) {
	newAccountExtend := &models.AccountExtend{}
	newAccountExtend.LastReconciledTime = accountModifyReq.LastReconciledTime

	if !isSubAccount && accountModifyReq.Category == models.ACCOUNT_CATEGORY_CREDIT_CARD {
		newAccountExtend.CreditCardStatementDate = &accountModifyReq.CreditCardStatementDate
		newAccountExtend.CreditCardLimit = creditLimitForCreditCard
	}

	newAccount := &models.Account{
		AccountId:    oldAccount.AccountId,
		Uid:          user.Uid,
		Name:         accountModifyReq.Name,
		DisplayOrder: oldAccount.DisplayOrder,
		Category:     accountModifyReq.Category,
		Icon:         accountModifyReq.Icon,
		IconType:     accountModifyReq.IconType,
		Color:        accountModifyReq.Color,
		Currency:     oldAccount.Currency,
		Comment:      accountModifyReq.Comment,
		Extend:       newAccountExtend,
		Hidden:       accountModifyReq.Hidden,
	}

	if !isSubAccount && accountModifyReq.Currency != nil {
		newAccount.Currency = *accountModifyReq.Currency
	}

	if newAccount.Name != oldAccount.Name ||
		newAccount.Category != oldAccount.Category ||
		newAccount.Icon != oldAccount.Icon ||
		newAccount.IconType != oldAccount.IconType ||
		newAccount.Color != oldAccount.Color ||
		newAccount.Currency != oldAccount.Currency ||
		newAccount.Comment != oldAccount.Comment ||
		newAccount.Hidden != oldAccount.Hidden {
		return newAccount, nil
	}

	oldAccountExtend := oldAccount.Extend

	if (newAccountExtend.LastReconciledTime != nil && (oldAccountExtend == nil || oldAccountExtend.LastReconciledTime == nil)) ||
		(newAccountExtend.LastReconciledTime == nil && oldAccountExtend != nil && oldAccountExtend.LastReconciledTime != nil) ||
		(newAccountExtend.LastReconciledTime != nil && oldAccountExtend != nil && oldAccountExtend.LastReconciledTime != nil && *newAccountExtend.LastReconciledTime != *oldAccountExtend.LastReconciledTime) {
		if !user.UseLastReconciledTime {
			return nil, errs.ErrLastReconciledTimeIsNotEnabled
		}

		return newAccount, nil
	}

	if (newAccountExtend.CreditCardStatementDate != nil && (oldAccountExtend == nil || oldAccountExtend.CreditCardStatementDate == nil)) ||
		(newAccountExtend.CreditCardStatementDate == nil && oldAccountExtend != nil && oldAccountExtend.CreditCardStatementDate != nil) ||
		(newAccountExtend.CreditCardStatementDate != nil && oldAccountExtend != nil && oldAccountExtend.CreditCardStatementDate != nil && *newAccountExtend.CreditCardStatementDate != *oldAccountExtend.CreditCardStatementDate) {
		return newAccount, nil
	}

	if (newAccountExtend.CreditCardLimit != nil && (oldAccountExtend == nil || oldAccountExtend.CreditCardLimit == nil)) ||
		(newAccountExtend.CreditCardLimit == nil && oldAccountExtend != nil && oldAccountExtend.CreditCardLimit != nil) ||
		(newAccountExtend.CreditCardLimit != nil && oldAccountExtend != nil && oldAccountExtend.CreditCardLimit != nil && *newAccountExtend.CreditCardLimit != *oldAccountExtend.CreditCardLimit) {
		return newAccount, nil
	}

	return nil, nil
}

func (a *AccountsApi) getToDeleteSubAccountIds(accountModifyReq *models.AccountModifyRequest, mainAccount *models.Account, accountAndSubAccounts []*models.Account) []int64 {
	newSubAccountIds := make(map[int64]bool, len(accountModifyReq.SubAccounts))

	for i := 0; i < len(accountModifyReq.SubAccounts); i++ {
		newSubAccountIds[accountModifyReq.SubAccounts[i].Id] = true
	}

	toDeleteAccountIds := make([]int64, 0)

	for i := 0; i < len(accountAndSubAccounts); i++ {
		subAccount := accountAndSubAccounts[i]

		if subAccount.AccountId == mainAccount.AccountId {
			continue
		}

		if _, exists := newSubAccountIds[subAccount.AccountId]; !exists {
			toDeleteAccountIds = append(toDeleteAccountIds, subAccount.AccountId)
		}
	}

	return toDeleteAccountIds
}

func (a *AccountsApi) isAccountsIconTypeValid(c *core.WebContext, uid int64, accounts []*models.Account) bool {
	iconIds := make([]int64, 0)

	for _, account := range accounts {
		if !account.IconType.IsValid() {
			return false
		}

		if account.IconType == core.ICON_TYPE_USER_CUSTOM {
			iconIds = append(iconIds, account.Icon)
		}
	}

	if len(iconIds) < 1 {
		return true
	}

	iconExists, err := a.icons.ExistsCustomIcons(c, uid, iconIds)

	if err != nil {
		log.Errorf(c, "[accounts.isAccountsIconTypeValid] failed to check custom icons for user \"uid:%d\", because %s", uid, err.Error())
		return false
	}

	return iconExists
}
