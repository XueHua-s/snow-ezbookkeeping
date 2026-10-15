package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/duplicatechecker"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"slices"
)

// AccountCreateHandler saves a new account by request parameters for current user
func (a *AccountsApi) AccountCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var accountCreateReq models.AccountCreateRequest
	err := c.ShouldBindJSON(&accountCreateReq)

	if err != nil {
		log.Warnf(c, "[accounts.AccountCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[accounts.AccountCreateHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	mainAccountBalance := int64(0)

	if accountCreateReq.Balance != "" {
		mainAccountBalance, err = utils.StringToInt64(accountCreateReq.Balance)

		if err != nil {
			return nil, errs.ErrIncompleteOrIncorrectSubmission
		}
	}

	subAccountBalances := make([]int64, len(accountCreateReq.SubAccounts))

	if accountCreateReq.Category < models.ACCOUNT_CATEGORY_CASH || accountCreateReq.Category > models.ACCOUNT_CATEGORY_CERTIFICATE_OF_DEPOSIT {
		log.Warnf(c, "[accounts.AccountCreateHandler] account category invalid, category is %d", accountCreateReq.Category)
		return nil, errs.ErrAccountCategoryInvalid
	}

	if accountCreateReq.Category != models.ACCOUNT_CATEGORY_CREDIT_CARD && accountCreateReq.CreditCardStatementDate != 0 {
		log.Warnf(c, "[accounts.AccountCreateHandler] cannot set statement date with category \"%d\"", accountCreateReq.Category)
		return nil, errs.ErrCannotSetStatementDateForNonCreditCard
	}

	if accountCreateReq.Category != models.ACCOUNT_CATEGORY_CREDIT_CARD && accountCreateReq.CreditCardLimit != "" {
		log.Warnf(c, "[accounts.AccountCreateHandler] cannot set credit limit with category \"%d\"", accountCreateReq.Category)
		return nil, errs.ErrCannotSetCreditLimitForNonCreditCardAccount
	}

	creditLimitForCreditCard := int64(0)

	if accountCreateReq.CreditCardLimit != "" {
		creditLimitForCreditCard, err = utils.StringToInt64(accountCreateReq.CreditCardLimit)

		if err != nil || creditLimitForCreditCard <= 0 {
			return nil, errs.ErrIncompleteOrIncorrectSubmission
		}
	}

	if accountCreateReq.Type == models.ACCOUNT_TYPE_SINGLE_ACCOUNT {
		if len(accountCreateReq.SubAccounts) > 0 {
			log.Warnf(c, "[accounts.AccountCreateHandler] account cannot have any sub-accounts")
			return nil, errs.ErrAccountCannotHaveSubAccounts
		}

		if accountCreateReq.Currency == core.AccountCurrencyNotSetValue {
			log.Warnf(c, "[accounts.AccountCreateHandler] account must set currency")
			return nil, errs.ErrAccountCurrencyInvalid
		}

		if mainAccountBalance != 0 && accountCreateReq.BalanceTime <= 0 {
			log.Warnf(c, "[accounts.AccountCreateHandler] account balance time is not set")
			return nil, errs.ErrAccountBalanceTimeNotSet
		}
	} else if accountCreateReq.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
		if len(accountCreateReq.SubAccounts) < 1 {
			log.Warnf(c, "[accounts.AccountCreateHandler] account does not have any sub-accounts")
			return nil, errs.ErrAccountHaveNoSubAccount
		}

		if accountCreateReq.Category == models.ACCOUNT_CATEGORY_CREDIT_CARD {
			if accountCreateReq.Currency == core.AccountCurrencyNotSetValue && creditLimitForCreditCard > 0 {
				log.Warnf(c, "[accounts.AccountCreateHandler] parent account must set currency when set credit limit")
				return nil, errs.ErrMustSetParentAccountCurrencyWhenSetCreditLimit
			}
		} else {
			if accountCreateReq.Currency != core.AccountCurrencyNotSetValue {
				log.Warnf(c, "[accounts.AccountCreateHandler] parent account cannot set currency")
				return nil, errs.ErrParentAccountCannotSetCurrency
			}
		}

		if mainAccountBalance != 0 {
			log.Warnf(c, "[accounts.AccountCreateHandler] parent account cannot set balance")
			return nil, errs.ErrParentAccountCannotSetBalance
		}

		for i := 0; i < len(accountCreateReq.SubAccounts); i++ {
			subAccount := accountCreateReq.SubAccounts[i]
			subAccountBalance := int64(0)

			if subAccount.Balance != "" {
				subAccountBalance, err = utils.StringToInt64(subAccount.Balance)

				if err != nil {
					return nil, errs.ErrIncompleteOrIncorrectSubmission
				}
			}

			subAccountBalances[i] = subAccountBalance

			if subAccount.Category != accountCreateReq.Category {
				log.Warnf(c, "[accounts.AccountCreateHandler] category of sub-account#%d not equals to parent", i)
				return nil, errs.ErrSubAccountCategoryNotEqualsToParent
			}

			if subAccount.Type != models.ACCOUNT_TYPE_SINGLE_ACCOUNT {
				log.Warnf(c, "[accounts.AccountCreateHandler] sub-account#%d type invalid", i)
				return nil, errs.ErrSubAccountTypeInvalid
			}

			if subAccount.Currency == core.AccountCurrencyNotSetValue {
				log.Warnf(c, "[accounts.AccountCreateHandler] sub-account#%d must set currency", i)
				return nil, errs.ErrAccountCurrencyInvalid
			}

			if subAccountBalance != 0 && subAccount.BalanceTime <= 0 {
				log.Warnf(c, "[accounts.AccountCreateHandler] sub-account#%d balance time is not set", i)
				return nil, errs.ErrAccountBalanceTimeNotSet
			}

			if subAccount.CreditCardStatementDate != 0 {
				log.Warnf(c, "[accounts.AccountCreateHandler] sub-account#%d cannot set statement date", i)
				return nil, errs.ErrCannotSetStatementDateForSubAccount
			}

			if subAccount.CreditCardLimit != "" {
				log.Warnf(c, "[accounts.AccountCreateHandler] sub-account#%d cannot set credit limit", i)
				return nil, errs.ErrCannotSetCreditLimitForSubAccount
			}
		}
	} else {
		log.Warnf(c, "[accounts.AccountCreateHandler] account type invalid, type is %d", accountCreateReq.Type)
		return nil, errs.ErrAccountTypeInvalid
	}

	uid := c.GetCurrentUid()
	maxOrderId, err := a.accounts.GetMaxDisplayOrder(c, uid, accountCreateReq.Category)

	if err != nil {
		log.Errorf(c, "[accounts.AccountCreateHandler] failed to get max display order for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	mainAccount := a.createNewAccountModel(uid, &accountCreateReq, mainAccountBalance, &creditLimitForCreditCard, false, maxOrderId+1)
	childrenAccounts, childrenAccountBalanceTimes := a.createSubAccountModels(uid, &accountCreateReq, subAccountBalances)
	iconTypeValid := a.isAccountsIconTypeValid(c, uid, slices.Concat([]*models.Account{mainAccount}, childrenAccounts))

	if !iconTypeValid {
		log.Warnf(c, "[accounts.AccountCreateHandler] icon type invalid for user \"uid:%d\"", uid)
		return nil, errs.ErrAccountIconInvalid
	}

	if a.CurrentConfig().EnableDuplicateSubmissionsCheck && accountCreateReq.ClientSessionId != "" {
		found, remark := a.GetSubmissionRemark(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_ACCOUNT, uid, accountCreateReq.ClientSessionId)

		if found {
			log.Infof(c, "[accounts.AccountCreateHandler] another account \"id:%s\" has been created for user \"uid:%d\"", remark, uid)
			accountId, err := utils.StringToInt64(remark)

			if err == nil {
				accountAndSubAccounts, err := a.accounts.GetAccountAndSubAccountsByAccountId(c, uid, accountId)

				if err != nil {
					log.Errorf(c, "[accounts.AccountCreateHandler] failed to get existed account \"id:%d\" for user \"uid:%d\", because %s", accountId, uid, err.Error())
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

	err = a.accounts.CreateAccounts(c, mainAccount, accountCreateReq.BalanceTime, childrenAccounts, childrenAccountBalanceTimes, clientTimezone)

	if err != nil {
		log.Errorf(c, "[accounts.AccountCreateHandler] failed to create account \"id:%d\" for user \"uid:%d\", because %s", mainAccount.AccountId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[accounts.AccountCreateHandler] user \"uid:%d\" has created a new account \"id:%d\" successfully", uid, mainAccount.AccountId)

	a.SetSubmissionRemarkIfEnable(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_ACCOUNT, uid, accountCreateReq.ClientSessionId, utils.Int64ToString(mainAccount.AccountId))
	accountInfoResp := mainAccount.ToAccountInfoResponse()

	if len(childrenAccounts) > 0 {
		accountInfoResp.SubAccounts = make([]*models.AccountInfoResponse, len(childrenAccounts))

		for i := 0; i < len(childrenAccounts); i++ {
			accountInfoResp.SubAccounts[i] = childrenAccounts[i].ToAccountInfoResponse()
		}
	}

	return accountInfoResp, nil
}
