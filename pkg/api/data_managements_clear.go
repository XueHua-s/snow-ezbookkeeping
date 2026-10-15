package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// ClearAllDataHandler deletes all user data
func (a *DataManagementsApi) ClearAllDataHandler(c *core.WebContext) (any, *errs.Error) {
	var clearDataReq models.ClearDataRequest
	err := c.ShouldBindJSON(&clearDataReq)

	if err != nil {
		log.Warnf(c, "[data_managements.ClearAllDataHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	claims := c.GetTokenClaims()

	if claims == nil {
		log.Warnf(c, "[data_managements.ClearAllDataHandler] current token is null")
		return nil, errs.ErrInvalidToken
	} else if claims.Type != core.USER_TOKEN_TYPE_NORMAL {
		log.Warnf(c, "[data_managements.ClearAllDataHandler] token type \"%d\" is not allowed to clear all user data", claims.Type)
		return nil, errs.ErrInvalidToken
	}

	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Warnf(c, "[data_managements.ClearAllDataHandler] failed to get user for user \"uid:%d\", because %s", uid, err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if !a.users.IsPasswordEqualsUserPassword(clearDataReq.Password, user) {
		return nil, errs.ErrUserPasswordWrong
	}

	if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_CLEAR_ALL_DATA) {
		return nil, errs.ErrNotPermittedToPerformThisAction
	}

	err = a.templates.DeleteAllTemplates(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllDataHandler] failed to delete all transaction templates, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	err = a.transactions.DeleteAllTransactions(c, uid, true)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllDataHandler] failed to delete all transactions, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	err = a.categories.DeleteAllCategories(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllDataHandler] failed to delete all transaction categories, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	err = a.tags.DeleteAllTags(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllDataHandler] failed to delete all transaction tags, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	err = a.tagGroups.DeleteAllTagGroups(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllDataHandler] failed to delete all transaction tag groups, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	err = a.userCustomIcons.DeleteAllCustomIcons(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllDataHandler] failed to delete all user custom icons, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	err = a.userCustomExchangeRates.DeleteAllCustomExchangeRates(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllDataHandler] failed to delete all user custom exchange rates, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	err = a.insightsExploreres.DeleteAllExplorations(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllDataHandler] failed to delete all explorations, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[data_managements.ClearAllDataHandler] user \"uid:%d\" has cleared all data", uid)
	return true, nil
}

// ClearAllTransactionsHandler deletes all transactions
func (a *DataManagementsApi) ClearAllTransactionsHandler(c *core.WebContext) (any, *errs.Error) {
	var clearDataReq models.ClearDataRequest
	err := c.ShouldBindJSON(&clearDataReq)

	if err != nil {
		log.Warnf(c, "[data_managements.ClearAllTransactionsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	claims := c.GetTokenClaims()

	if claims == nil {
		log.Warnf(c, "[data_managements.ClearAllTransactionsHandler] current token is null")
		return nil, errs.ErrInvalidToken
	} else if claims.Type != core.USER_TOKEN_TYPE_NORMAL {
		log.Warnf(c, "[data_managements.ClearAllTransactionsHandler] token type \"%d\" is not allowed to clear all transactions", claims.Type)
		return nil, errs.ErrInvalidToken
	}

	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Warnf(c, "[data_managements.ClearAllTransactionsHandler] failed to get user for user \"uid:%d\", because %s", uid, err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if !a.users.IsPasswordEqualsUserPassword(clearDataReq.Password, user) {
		return nil, errs.ErrUserPasswordWrong
	}

	if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_CLEAR_ALL_DATA) {
		return nil, errs.ErrNotPermittedToPerformThisAction
	}

	err = a.transactions.DeleteAllTransactions(c, uid, false)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllTransactionsHandler] failed to delete all transactions, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[data_managements.ClearAllTransactionsHandler] user \"uid:%d\" has cleared all transactions", uid)
	return true, nil
}

// ClearAllTransactionsByAccountHandler deletes all transactions of specified account
func (a *DataManagementsApi) ClearAllTransactionsByAccountHandler(c *core.WebContext) (any, *errs.Error) {
	var clearDataReq models.ClearAccountTransactionsRequest
	err := c.ShouldBindJSON(&clearDataReq)

	if err != nil {
		log.Warnf(c, "[data_managements.ClearAllTransactionsByAccountHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	claims := c.GetTokenClaims()

	if claims == nil {
		log.Warnf(c, "[data_managements.ClearAllTransactionsByAccountHandler] current token is null")
		return nil, errs.ErrInvalidToken
	} else if claims.Type != core.USER_TOKEN_TYPE_NORMAL {
		log.Warnf(c, "[data_managements.ClearAllTransactionsByAccountHandler] token type \"%d\" is not allowed to clear all transactions of account", claims.Type)
		return nil, errs.ErrInvalidToken
	}

	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Warnf(c, "[data_managements.ClearAllTransactionsByAccountHandler] failed to get user for user \"uid:%d\", because %s", uid, err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if !a.users.IsPasswordEqualsUserPassword(clearDataReq.Password, user) {
		return nil, errs.ErrUserPasswordWrong
	}

	if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_CLEAR_ALL_DATA) {
		return nil, errs.ErrNotPermittedToPerformThisAction
	}

	account, err := a.accounts.GetAccountByAccountId(c, uid, clearDataReq.AccountId)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllTransactionsByAccountHandler] failed to get account \"id:%d\" for user \"uid:%d\", because %s", uid, clearDataReq.AccountId, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if account.Hidden {
		return nil, errs.ErrCannotDeleteTransactionInHiddenAccount
	}

	if account.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
		return nil, errs.ErrCannotDeleteTransactionInParentAccount
	}

	err = a.transactions.DeleteAllTransactionsOfAccount(c, uid, account.AccountId, pageCountForClearTransactions)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllTransactionsByAccountHandler] failed to delete all transactions in account \"id:%d\", because %s", account.AccountId, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[data_managements.ClearAllTransactionsByAccountHandler] user \"uid:%d\" has cleared all transactions in account \"id:%d\"", uid, account.AccountId)
	return true, nil
}
