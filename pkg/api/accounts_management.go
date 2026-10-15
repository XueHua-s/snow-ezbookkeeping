package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// AccountUpdateLastReconciledTimeHandler updates last reconciled time of an existed account by request parameters for current user
func (a *AccountsApi) AccountUpdateLastReconciledTimeHandler(c *core.WebContext) (any, *errs.Error) {
	var accountUpdateReq models.AccountUpdateLastReconciledTimeRequest
	err := c.ShouldBindJSON(&accountUpdateReq)

	if err != nil {
		log.Warnf(c, "[accounts.AccountUpdateLastReconciledTimeHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	if accountUpdateReq.Id <= 0 {
		return nil, errs.ErrAccountIdInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[accounts.AccountUpdateLastReconciledTimeHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if !user.UseLastReconciledTime {
		return nil, errs.ErrLastReconciledTimeIsNotEnabled
	}

	account, err := a.accounts.GetAccountByAccountId(c, uid, accountUpdateReq.Id)

	if err != nil {
		log.Errorf(c, "[accounts.AccountUpdateLastReconciledTimeHandler] failed to get account \"id:%d\" for user \"uid:%d\", because %s", accountUpdateReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if account.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
		return nil, errs.ErrParentAccountCannotSetLastReconciledTime
	}

	if account.Extend == nil {
		account.Extend = &models.AccountExtend{}
	}

	if account.Extend.LastReconciledTime != nil && accountUpdateReq.LastReconciledTime < *account.Extend.LastReconciledTime {
		return nil, errs.ErrCannotSetLastReconciledTimeBeforeCurrent
	} else if account.Extend.LastReconciledTime != nil && accountUpdateReq.LastReconciledTime == *account.Extend.LastReconciledTime {
		return nil, errs.ErrNothingWillBeUpdated
	}

	account.Extend.LastReconciledTime = &accountUpdateReq.LastReconciledTime

	err = a.accounts.UpdateAccountExtend(c, uid, account)

	if err != nil {
		log.Errorf(c, "[accounts.AccountUpdateLastReconciledTimeHandler] failed to update last reconciled time for account \"id:%d\" of user \"uid:%d\", because %s", account.AccountId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[accounts.AccountUpdateLastReconciledTimeHandler] user \"uid:%d\" has updated last reconciled time \"%d\" for account \"id:%d\"", uid, account.Extend.LastReconciledTime, account.AccountId)
	return true, nil
}

// AccountHideHandler hides an existed account by request parameters for current user
func (a *AccountsApi) AccountHideHandler(c *core.WebContext) (any, *errs.Error) {
	var accountHideReq models.AccountHideRequest
	err := c.ShouldBindJSON(&accountHideReq)

	if err != nil {
		log.Warnf(c, "[accounts.AccountHideHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.accounts.HideAccount(c, uid, []int64{accountHideReq.Id}, accountHideReq.Hidden)

	if err != nil {
		log.Errorf(c, "[accounts.AccountHideHandler] failed to hide account \"id:%d\" for user \"uid:%d\", because %s", accountHideReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[accounts.AccountHideHandler] user \"uid:%d\" has hidden account \"id:%d\"", uid, accountHideReq.Id)
	return true, nil
}

// AccountMoveHandler moves display order of existed accounts by request parameters for current user
func (a *AccountsApi) AccountMoveHandler(c *core.WebContext) (any, *errs.Error) {
	var accountMoveReq models.AccountMoveRequest
	err := c.ShouldBindJSON(&accountMoveReq)

	if err != nil {
		log.Warnf(c, "[accounts.AccountMoveHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	accounts := make([]*models.Account, len(accountMoveReq.NewDisplayOrders))

	for i := 0; i < len(accountMoveReq.NewDisplayOrders); i++ {
		newDisplayOrder := accountMoveReq.NewDisplayOrders[i]
		account := &models.Account{
			Uid:          uid,
			AccountId:    newDisplayOrder.Id,
			DisplayOrder: newDisplayOrder.DisplayOrder,
		}

		accounts[i] = account
	}

	err = a.accounts.ModifyAccountDisplayOrders(c, uid, accounts)

	if err != nil {
		log.Errorf(c, "[accounts.AccountMoveHandler] failed to move accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[accounts.AccountMoveHandler] user \"uid:%d\" has moved accounts", uid)
	return true, nil
}

// AccountDeleteHandler deletes an existed account by request parameters for current user
func (a *AccountsApi) AccountDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var accountDeleteReq models.AccountDeleteRequest
	err := c.ShouldBindJSON(&accountDeleteReq)

	if err != nil {
		log.Warnf(c, "[accounts.AccountDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.accounts.DeleteAccount(c, uid, accountDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[accounts.AccountDeleteHandler] failed to delete account \"id:%d\" for user \"uid:%d\", because %s", accountDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[accounts.AccountDeleteHandler] user \"uid:%d\" has deleted account \"id:%d\"", uid, accountDeleteReq.Id)
	return true, nil
}

// SubAccountDeleteHandler deletes an existed sub-account by request parameters for current user
func (a *AccountsApi) SubAccountDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var accountDeleteReq models.AccountDeleteRequest
	err := c.ShouldBindJSON(&accountDeleteReq)

	if err != nil {
		log.Warnf(c, "[accounts.SubAccountDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.accounts.DeleteSubAccount(c, uid, accountDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[accounts.SubAccountDeleteHandler] failed to delete sub-account \"id:%d\" for user \"uid:%d\", because %s", accountDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[accounts.SubAccountDeleteHandler] user \"uid:%d\" has deleted sub-account \"id:%d\"", uid, accountDeleteReq.Id)
	return true, nil
}
