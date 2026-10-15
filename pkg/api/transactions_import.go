package api

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/duplicatechecker"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"strings"
)

// TransactionImportHandler imports transactions by request parameters for current user
func (a *TransactionsApi) TransactionImportHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionImportReq models.TransactionImportRequest
	err := c.ShouldBindJSON(&transactionImportReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionImportHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionImportHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()

	if a.CurrentConfig().EnableDuplicateSubmissionsCheck && transactionImportReq.ClientSessionId != "" {
		found, remark := a.GetSubmissionRemark(duplicatechecker.DUPLICATE_CHECKER_TYPE_IMPORT_TRANSACTIONS, uid, transactionImportReq.ClientSessionId)

		if found {
			items := strings.Split(remark, ":")

			if len(items) >= 2 {
				if items[0] == "finished" {
					log.Infof(c, "[transactions.TransactionImportHandler] another \"%s\" transactions has been imported for user \"uid:%d\"", items[1], uid)
					count, err := utils.StringToInt(items[1])

					if err == nil {
						return count, nil
					}
				} else if items[0] == "processing" {
					return nil, errs.ErrRepeatedRequest
				}
			} else {
				log.Warnf(c, "[transactions.TransactionImportHandler] another transaction import task may be executing, but remark \"%s\" is invalid", remark)
			}
		}
	}

	newTransactionTagIdsMap := make(map[int][]int64, len(transactionImportReq.Transactions))

	for i := 0; i < len(transactionImportReq.Transactions); i++ {
		transactionCreateReq := transactionImportReq.Transactions[i]
		tagIds, err := utils.StringArrayToInt64Array(transactionCreateReq.TagIds)

		if err != nil {
			log.Warnf(c, "[transactions.TransactionImportHandler] parse tag ids failed of transaction \"index:%d\", because %s", i, err.Error())
			return nil, errs.ErrTransactionTagIdInvalid
		}

		if len(tagIds) > models.MaximumTagsCountOfTransaction {
			return nil, errs.ErrTransactionHasTooManyTags
		}

		if transactionCreateReq.Type < models.TRANSACTION_TYPE_MODIFY_BALANCE || transactionCreateReq.Type > models.TRANSACTION_TYPE_TRANSFER {
			log.Warnf(c, "[transactions.TransactionImportHandler] transaction type of transaction \"index:%d\" is invalid", i)
			return nil, errs.ErrTransactionTypeInvalid
		}

		if transactionCreateReq.Type == models.TRANSACTION_TYPE_MODIFY_BALANCE && transactionCreateReq.CategoryId != 0 {
			log.Warnf(c, "[transactions.TransactionImportHandler] balance modification transaction \"index:%d\" cannot set category id", i)
			return nil, errs.ErrBalanceModificationTransactionCannotSetCategory
		}

		if transactionCreateReq.Type != models.TRANSACTION_TYPE_TRANSFER && transactionCreateReq.DestinationAccountId != 0 {
			log.Warnf(c, "[transactions.TransactionImportHandler] non-transfer transaction \"index:%d\" destination account cannot be set", i)
			return nil, errs.ErrTransactionDestinationAccountCannotBeSet
		} else if transactionCreateReq.Type == models.TRANSACTION_TYPE_TRANSFER && transactionCreateReq.SourceAccountId == transactionCreateReq.DestinationAccountId {
			log.Warnf(c, "[transactions.TransactionImportHandler] transfer transaction \"index:%d\" source account must not be destination account", i)
			return nil, errs.ErrTransactionSourceAndDestinationIdCannotBeEqual
		}

		if transactionCreateReq.Type != models.TRANSACTION_TYPE_TRANSFER && transactionCreateReq.DestinationAmount != 0 {
			log.Warnf(c, "[transactions.TransactionImportHandler] non-transfer transaction \"index:%d\" destination amount cannot be set", i)
			return nil, errs.ErrTransactionDestinationAmountCannotBeSet
		}

		newTransactionTagIdsMap[i] = tagIds
	}

	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionImportHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_IMPORT_TRANSACTION) {
		return nil, errs.ErrNotPermittedToPerformThisAction
	}

	newTransactions := make([]*models.Transaction, len(transactionImportReq.Transactions))

	for i := 0; i < len(transactionImportReq.Transactions); i++ {
		transactionCreateReq := transactionImportReq.Transactions[i]
		transaction := a.createNewTransactionModel(uid, transactionCreateReq, c.ClientIP())
		newTransactions[i] = transaction
	}

	allUsedAccounts, err := a.getTransactionUsedAccounts(c, uid, newTransactions)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionImportHandler] failed to get transaction used accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	for i := 0; i < len(newTransactions); i++ {
		transaction := newTransactions[i]
		transactionEditable := user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone, allUsedAccounts[transaction.AccountId], allUsedAccounts[transaction.RelatedAccountId])

		if !transactionEditable {
			log.Warnf(c, "[transactions.TransactionImportHandler] transaction \"index:%d\" is not editable for user \"uid:%d\"", i, uid)
			return nil, errs.ErrCannotCreateTransactionWithThisTransactionTime
		}
	}

	err = a.transactions.BatchCreateTransactions(c, user.Uid, newTransactions, newTransactionTagIdsMap, func(currentProcess float64) {
		a.SetSubmissionRemarkIfEnable(duplicatechecker.DUPLICATE_CHECKER_TYPE_IMPORT_TRANSACTIONS, uid, transactionImportReq.ClientSessionId, fmt.Sprintf("processing:%.2f", currentProcess))
	})
	count := len(newTransactions)

	if err != nil {
		a.RemoveSubmissionRemarkIfEnable(duplicatechecker.DUPLICATE_CHECKER_TYPE_IMPORT_TRANSACTIONS, uid, transactionImportReq.ClientSessionId)
		log.Errorf(c, "[transactions.TransactionImportHandler] failed to import %d transactions for user \"uid:%d\", because %s", count, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionImportHandler] user \"uid:%d\" has imported %d transactions successfully", uid, count)

	a.SetSubmissionRemarkIfEnable(duplicatechecker.DUPLICATE_CHECKER_TYPE_IMPORT_TRANSACTIONS, uid, transactionImportReq.ClientSessionId, fmt.Sprintf("finished:%d", count))

	return count, nil
}

// TransactionImportProcessHandler returns the process of specified transaction import task by request parameters for current user
func (a *TransactionsApi) TransactionImportProcessHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionImportProcessReq models.TransactionImportProcessRequest
	err := c.ShouldBindQuery(&transactionImportProcessReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionImportProcessHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	if !a.CurrentConfig().EnableDuplicateSubmissionsCheck {
		return nil, nil
	}

	found, remark := a.GetSubmissionRemark(duplicatechecker.DUPLICATE_CHECKER_TYPE_IMPORT_TRANSACTIONS, uid, transactionImportProcessReq.ClientSessionId)

	if !found {
		return nil, nil
	}

	items := strings.Split(remark, ":")

	if len(items) < 2 {
		return nil, nil
	}

	if items[0] == "finished" {
		return 100, nil
	} else if items[0] != "processing" {
		return nil, nil
	}

	process, err := utils.StringToFloat64(items[1])

	if err != nil {
		log.Warnf(c, "[transactions.TransactionImportProcessHandler] parse process failed, because %s", err.Error())
		return nil, nil
	}

	if process < 0 {
		return nil, nil
	} else if process >= 100 {
		process = 100
	}

	return process, nil
}
