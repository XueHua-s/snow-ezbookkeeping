package api

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"time"
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

	if a.assistantStorage == nil {
		return nil, errs.ErrOperationFailed
	}
	command, created, err := a.beginAssistantForAllData(c, clearDataReq.AssistantCommand)
	if err != nil {
		return nil, assistantExperienceError(err)
	}
	if !created {
		completed, err := a.assistantStorage.GetCommandOperationCompleted(c, uid, command.ClientRequestId)
		if err != nil {
			return nil, assistantExperienceError(err)
		}
		if completed {
			return true, nil
		}
		return nil, errs.ErrUserDataClearIncomplete
	}

	err = a.templates.DeleteAllTemplates(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllDataHandler] failed to delete all transaction templates, because %s", err.Error())
		return nil, errs.ErrUserDataClearIncomplete
	}

	err = a.transactions.DeleteAllTransactions(c, uid, true)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllDataHandler] failed to delete all transactions, because %s", err.Error())
		return nil, errs.ErrUserDataClearIncomplete
	}

	err = a.categories.DeleteAllCategories(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllDataHandler] failed to delete all transaction categories, because %s", err.Error())
		return nil, errs.ErrUserDataClearIncomplete
	}

	err = a.tags.DeleteAllTags(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllDataHandler] failed to delete all transaction tags, because %s", err.Error())
		return nil, errs.ErrUserDataClearIncomplete
	}

	err = a.tagGroups.DeleteAllTagGroups(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllDataHandler] failed to delete all transaction tag groups, because %s", err.Error())
		return nil, errs.ErrUserDataClearIncomplete
	}

	err = a.userCustomIcons.DeleteAllCustomIcons(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllDataHandler] failed to delete all user custom icons, because %s", err.Error())
		return nil, errs.ErrUserDataClearIncomplete
	}

	err = a.userCustomExchangeRates.DeleteAllCustomExchangeRates(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllDataHandler] failed to delete all user custom exchange rates, because %s", err.Error())
		return nil, errs.ErrUserDataClearIncomplete
	}

	err = a.insightsExploreres.DeleteAllExplorations(c, uid)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllDataHandler] failed to delete all explorations, because %s", err.Error())
		return nil, errs.ErrUserDataClearIncomplete
	}

	if err = a.assistantStorage.CompleteCommandOperation(c, uid, command.ClientRequestId); err != nil {
		return nil, errs.ErrUserDataClearIncomplete
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

	if a.assistantStorage == nil {
		return nil, errs.ErrOperationFailed
	}
	if err = a.assistantStorage.InvalidateLedgerContext(c, uid); err != nil {
		return nil, errs.ErrOperationFailed
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

	if a.assistantStorage == nil {
		return nil, errs.ErrOperationFailed
	}
	if err = a.assistantStorage.InvalidateLedgerContext(c, uid); err != nil {
		return nil, errs.ErrOperationFailed
	}

	err = a.transactions.DeleteAllTransactionsOfAccount(c, uid, account.AccountId, pageCountForClearTransactions)

	if err != nil {
		log.Errorf(c, "[data_managements.ClearAllTransactionsByAccountHandler] failed to delete all transactions in account \"id:%d\", because %s", account.AccountId, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[data_managements.ClearAllTransactionsByAccountHandler] user \"uid:%d\" has cleared all transactions in account \"id:%d\"", uid, account.AccountId)
	return true, nil
}

// The legacy request has no command ID. A token-bound first-clear command is
// stable across retries and can only clear epoch zero; renewed confirmation in
// a capable client is required for later clears. Passwords never enter receipts.
func (a *DataManagementsApi) beginAssistantForAllData(c *core.WebContext, command *models.AssistantCommand) (*models.AssistantCommand, bool, error) {
	if command == nil {
		claims := c.GetTokenClaims()
		digest := sha256.Sum256([]byte("clear-all:" + claims.UserTokenId))
		command = &models.AssistantCommand{ClientRequestId: "legacy-clear-" + hex.EncodeToString(digest[:]), RequestIssuedAt: time.Unix(claims.IssuedAt, 0).UTC().Format(time.RFC3339Nano), ExpectedDeletionEpoch: 0}
	}
	_, created, err := a.assistantStorage.ExecuteCommandWithAcceptance(c, c.GetCurrentUid(), "user_data.clear_assistant", *command, *command, func(tx *services.AssistantStorageTx) (models.AssistantCommandResult, error) {
		return models.AssistantCommandResult{}, tx.ClearAssistantData()
	})
	return command, created, err
}
