package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/duplicatechecker"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"slices"
	"sort"
)

// AccountModifyHandler saves an existed account by request parameters for current user
func (a *AccountsApi) AccountModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var accountModifyReq models.AccountModifyRequest
	err := c.ShouldBindJSON(&accountModifyReq)

	if err != nil {
		log.Warnf(c, "[accounts.AccountModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	if accountModifyReq.Id <= 0 {
		return nil, errs.ErrAccountIdInvalid
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[accounts.AccountModifyHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	if accountModifyReq.Category < models.ACCOUNT_CATEGORY_CASH || accountModifyReq.Category > models.ACCOUNT_CATEGORY_CERTIFICATE_OF_DEPOSIT {
		log.Warnf(c, "[accounts.AccountModifyHandler] account category invalid, category is %d", accountModifyReq.Category)
		return nil, errs.ErrAccountCategoryInvalid
	}

	if accountModifyReq.Category != models.ACCOUNT_CATEGORY_CREDIT_CARD && accountModifyReq.CreditCardStatementDate != 0 {
		log.Warnf(c, "[accounts.AccountModifyHandler] cannot set statement date with category \"%d\"", accountModifyReq.Category)
		return nil, errs.ErrCannotSetStatementDateForNonCreditCard
	}

	if accountModifyReq.Category != models.ACCOUNT_CATEGORY_CREDIT_CARD && accountModifyReq.CreditCardLimit != "" {
		log.Warnf(c, "[accounts.AccountModifyHandler] cannot set credit limit with category \"%d\"", accountModifyReq.Category)
		return nil, errs.ErrCannotSetCreditLimitForNonCreditCardAccount
	}

	creditLimitForCreditCard := int64(0)

	if accountModifyReq.CreditCardLimit != "" {
		creditLimitForCreditCard, err = utils.StringToInt64(accountModifyReq.CreditCardLimit)

		if err != nil || creditLimitForCreditCard <= 0 {
			return nil, errs.ErrIncompleteOrIncorrectSubmission
		}
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[accounts.AccountModifyHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	accountAndSubAccounts, err := a.accounts.GetAccountAndSubAccountsByAccountId(c, uid, accountModifyReq.Id)

	if err != nil {
		log.Errorf(c, "[accounts.AccountModifyHandler] failed to get account \"id:%d\" for user \"uid:%d\", because %s", accountModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	accountMap := a.accounts.GetAccountMapByList(accountAndSubAccounts)
	mainAccount, exists := accountMap[accountModifyReq.Id]
	subAccountBalances := make([]int64, len(accountModifyReq.SubAccounts))

	if !exists {
		return nil, errs.ErrAccountNotFound
	}

	if accountModifyReq.Balance != nil {
		return nil, errs.ErrNotSupportedChangeBalance
	}

	if accountModifyReq.BalanceTime != nil {
		return nil, errs.ErrNotSupportedChangeBalanceTime
	}

	updateMainAccountCurrency := false

	if mainAccount.Type == models.ACCOUNT_TYPE_SINGLE_ACCOUNT {
		if len(accountModifyReq.SubAccounts) > 0 {
			log.Warnf(c, "[accounts.AccountModifyHandler] account cannot have any sub-accounts")
			return nil, errs.ErrAccountCannotHaveSubAccounts
		}

		if accountModifyReq.Currency != nil && mainAccount.Currency != *accountModifyReq.Currency {
			return nil, errs.ErrNotSupportedChangeCurrency
		}
	} else if mainAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
		if len(accountModifyReq.SubAccounts) < 1 {
			log.Warnf(c, "[accounts.AccountModifyHandler] account does not have any sub-accounts")
			return nil, errs.ErrAccountHaveNoSubAccount
		}

		if accountModifyReq.Category == models.ACCOUNT_CATEGORY_CREDIT_CARD {
			if ((accountModifyReq.Currency == nil && mainAccount.Currency == core.AccountCurrencyNotSetValue) || (accountModifyReq.Currency != nil && *accountModifyReq.Currency == core.AccountCurrencyNotSetValue)) && creditLimitForCreditCard > 0 {
				log.Warnf(c, "[accounts.AccountModifyHandler] parent account must set currency when set credit limit")
				return nil, errs.ErrMustSetParentAccountCurrencyWhenSetCreditLimit
			} else if accountModifyReq.Currency != nil && mainAccount.Currency != *accountModifyReq.Currency {
				log.Infof(c, "[accounts.AccountModifyHandler] will change parent account curreny from %s to %s", mainAccount.Currency, *accountModifyReq.Currency)
				updateMainAccountCurrency = true
			}
		} else {
			if accountModifyReq.Currency != nil && mainAccount.Currency != *accountModifyReq.Currency {
				return nil, errs.ErrNotSupportedChangeCurrency
			}
		}

		for i := 0; i < len(accountModifyReq.SubAccounts); i++ {
			subAccountReq := accountModifyReq.SubAccounts[i]
			subAccountBalances[i] = 0

			if subAccountReq.Category != accountModifyReq.Category {
				log.Warnf(c, "[accounts.AccountModifyHandler] category of sub-account#%d not equals to parent", i)
				return nil, errs.ErrSubAccountCategoryNotEqualsToParent
			}

			if subAccountReq.Id == 0 { // create new sub-account
				if subAccountReq.Currency == nil {
					log.Warnf(c, "[accounts.AccountModifyHandler] sub-account#%d not set currency", i)
					return nil, errs.ErrAccountCurrencyInvalid
				} else if subAccountReq.Currency != nil && *subAccountReq.Currency == core.AccountCurrencyNotSetValue {
					log.Warnf(c, "[accounts.AccountModifyHandler] sub-account#%d must set currency", i)
					return nil, errs.ErrAccountCurrencyInvalid
				}

				subAccountBalance := int64(0)

				if subAccountReq.Balance != nil && *subAccountReq.Balance != "" {
					subAccountBalance, err = utils.StringToInt64(*subAccountReq.Balance)

					if err != nil {
						return nil, errs.ErrIncompleteOrIncorrectSubmission
					}
				}

				subAccountBalances[i] = subAccountBalance

				if subAccountBalance == 0 {
					defaultBalanceTime := int64(0)
					subAccountReq.BalanceTime = &defaultBalanceTime
				}

				if subAccountBalance != 0 && (subAccountReq.BalanceTime == nil || *subAccountReq.BalanceTime <= 0) {
					log.Warnf(c, "[accounts.AccountModifyHandler] sub-account#%d balance time is not set", i)
					return nil, errs.ErrAccountBalanceTimeNotSet
				}
			} else { // modify existed sub-account
				subAccount, exists := accountMap[subAccountReq.Id]

				if !exists {
					return nil, errs.ErrAccountNotFound
				}

				if subAccountReq.Currency != nil && subAccount.Currency != *subAccountReq.Currency {
					return nil, errs.ErrNotSupportedChangeCurrency
				}

				if subAccountReq.Balance != nil {
					return nil, errs.ErrNotSupportedChangeBalance
				}

				if subAccountReq.BalanceTime != nil {
					return nil, errs.ErrNotSupportedChangeBalanceTime
				}
			}

			if subAccountReq.CreditCardStatementDate != 0 {
				log.Warnf(c, "[accounts.AccountModifyHandler] sub-account#%d cannot set statement date", i)
				return nil, errs.ErrCannotSetStatementDateForSubAccount
			}

			if subAccountReq.CreditCardLimit != "" {
				log.Warnf(c, "[accounts.AccountModifyHandler] sub-account#%d cannot set credit limit", i)
				return nil, errs.ErrCannotSetCreditLimitForSubAccount
			}
		}
	}

	anythingUpdate := false
	var toUpdateAccounts []*models.Account
	var toAddAccounts []*models.Account
	var toAddAccountBalanceTimes []int64
	var toDeleteAccountIds []int64

	toUpdateAccount, err := a.getToUpdateAccount(user, &accountModifyReq, &creditLimitForCreditCard, mainAccount, false)

	if err != nil {
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if toUpdateAccount != nil {
		if toUpdateAccount.Category != mainAccount.Category {
			maxOrderId, err := a.accounts.GetMaxDisplayOrder(c, uid, toUpdateAccount.Category)

			if err != nil {
				log.Errorf(c, "[accounts.AccountModifyHandler] failed to get max display order for user \"uid:%d\", because %s", uid, err.Error())
				return nil, errs.Or(err, errs.ErrOperationFailed)
			}

			toUpdateAccount.DisplayOrder = maxOrderId + 1
		}

		anythingUpdate = true
		toUpdateAccounts = append(toUpdateAccounts, toUpdateAccount)
	}

	toDeleteAccountIds = a.getToDeleteSubAccountIds(&accountModifyReq, mainAccount, accountAndSubAccounts)

	if len(toDeleteAccountIds) > 0 {
		anythingUpdate = true
	}

	maxOrderId := int32(0)

	for i := 0; i < len(accountAndSubAccounts); i++ {
		account := accountAndSubAccounts[i]

		if account.AccountId != mainAccount.AccountId && account.DisplayOrder > maxOrderId {
			maxOrderId = account.DisplayOrder
		}
	}

	for i := 0; i < len(accountModifyReq.SubAccounts); i++ {
		subAccountReq := accountModifyReq.SubAccounts[i]

		if _, exists := accountMap[subAccountReq.Id]; !exists {
			anythingUpdate = true
			maxOrderId = maxOrderId + 1
			newSubAccount := a.createNewSubAccountModelForModify(uid, mainAccount.Type, subAccountReq, subAccountBalances[i], maxOrderId)
			toAddAccounts = append(toAddAccounts, newSubAccount)

			if subAccountReq.BalanceTime != nil {
				toAddAccountBalanceTimes = append(toAddAccountBalanceTimes, *subAccountReq.BalanceTime)
			} else {
				toAddAccountBalanceTimes = append(toAddAccountBalanceTimes, 0)
			}
		} else {
			toUpdateSubAccount, err := a.getToUpdateAccount(user, subAccountReq, nil, accountMap[subAccountReq.Id], true)

			if err != nil {
				return nil, errs.Or(err, errs.ErrOperationFailed)
			}

			if toUpdateSubAccount != nil {
				anythingUpdate = true
				toUpdateAccounts = append(toUpdateAccounts, toUpdateSubAccount)
			}
		}
	}

	if !anythingUpdate {
		return nil, errs.ErrNothingWillBeUpdated
	}

	iconTypeValid := a.isAccountsIconTypeValid(c, uid, slices.Concat(toAddAccounts, toUpdateAccounts))

	if !iconTypeValid {
		log.Warnf(c, "[accounts.AccountModifyHandler] icon type invalid for user \"uid:%d\"", uid)
		return nil, errs.ErrAccountIconInvalid
	}

	if len(toAddAccounts) > 0 && a.CurrentConfig().EnableDuplicateSubmissionsCheck && accountModifyReq.ClientSessionId != "" {
		found, remark := a.GetSubmissionRemark(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_SUBACCOUNT, uid, accountModifyReq.ClientSessionId)

		if found {
			log.Infof(c, "[accounts.AccountModifyHandler] another account \"id:%s\" modification has been created for user \"uid:%d\"", remark, uid)
			accountId, err := utils.StringToInt64(remark)

			if err == nil {
				accountAndSubAccounts, err := a.accounts.GetAccountAndSubAccountsByAccountId(c, uid, accountId)

				if err != nil {
					log.Errorf(c, "[accounts.AccountModifyHandler] failed to get existed account \"id:%d\" for user \"uid:%d\", because %s", accountId, uid, err.Error())
					return nil, errs.Or(err, errs.ErrOperationFailed)
				}

				accountMap := a.accounts.GetAccountMapByList(accountAndSubAccounts)
				mainAccount, exists := accountMap[accountId]

				if !exists {
					return nil, errs.ErrOperationFailed
				}

				accountInfoResp := mainAccount.ToAccountInfoResponse()

				for i := 0; i < len(accountAndSubAccounts); i++ {
					if accountAndSubAccounts[i].ParentAccountId == mainAccount.AccountId {
						subAccountResp := accountAndSubAccounts[i].ToAccountInfoResponse()
						accountInfoResp.SubAccounts = append(accountInfoResp.SubAccounts, subAccountResp)
					}
				}

				return accountInfoResp, nil
			}
		}
	}

	err = a.accounts.ModifyAccounts(c, mainAccount, toUpdateAccounts, toAddAccounts, toAddAccountBalanceTimes, toDeleteAccountIds, updateMainAccountCurrency, clientTimezone)

	if err != nil {
		log.Errorf(c, "[accounts.AccountModifyHandler] failed to update account \"id:%d\" for user \"uid:%d\", because %s", accountModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[accounts.AccountModifyHandler] user \"uid:%d\" has updated account \"id:%d\" successfully", uid, accountModifyReq.Id)

	if len(toAddAccounts) > 0 {
		a.SetSubmissionRemarkIfEnable(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_SUBACCOUNT, uid, accountModifyReq.ClientSessionId, utils.Int64ToString(mainAccount.AccountId))
	}

	accountRespMap := make(map[int64]*models.AccountInfoResponse)

	for i := 0; i < len(toUpdateAccounts); i++ {
		account := toUpdateAccounts[i]
		oldAccount := accountMap[account.AccountId]

		account.Type = oldAccount.Type
		account.ParentAccountId = oldAccount.ParentAccountId
		account.Balance = oldAccount.Balance

		accountResp := account.ToAccountInfoResponse()
		accountRespMap[accountResp.Id] = accountResp
	}

	for i := 0; i < len(toAddAccounts); i++ {
		account := toAddAccounts[i]
		accountResp := account.ToAccountInfoResponse()
		accountRespMap[accountResp.Id] = accountResp
	}

	deletedAccountIds := make(map[int64]bool)

	for i := 0; i < len(toDeleteAccountIds); i++ {
		deletedAccountIds[toDeleteAccountIds[i]] = true
	}

	for i := 0; i < len(accountAndSubAccounts); i++ {
		oldAccount := accountAndSubAccounts[i]
		_, exists := accountRespMap[oldAccount.AccountId]

		if !exists && !deletedAccountIds[oldAccount.AccountId] {
			oldAccountResp := oldAccount.ToAccountInfoResponse()
			accountRespMap[oldAccountResp.Id] = oldAccountResp
		}
	}

	accountResp := accountRespMap[accountModifyReq.Id]

	for i := 0; i < len(accountAndSubAccounts); i++ {
		account := accountAndSubAccounts[i]

		if account.ParentAccountId == accountResp.Id && !deletedAccountIds[account.AccountId] {
			subAccountResp := accountRespMap[account.AccountId]
			accountResp.SubAccounts = append(accountResp.SubAccounts, subAccountResp)
		}
	}

	for i := 0; i < len(toAddAccounts); i++ {
		account := toAddAccounts[i]

		if account.ParentAccountId == accountResp.Id {
			subAccountResp := accountRespMap[account.AccountId]
			accountResp.SubAccounts = append(accountResp.SubAccounts, subAccountResp)
		}
	}

	sort.Sort(accountResp.SubAccounts)

	return accountResp, nil
}
