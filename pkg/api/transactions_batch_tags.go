package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// TransactionBatchAddTagsHandler batch add tags to transactions by request parameters for current user
func (a *TransactionsApi) TransactionBatchAddTagsHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionBatchUpdateReq models.TransactionBatchAddTagsRequest
	err := c.ShouldBindJSON(&transactionBatchUpdateReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchAddTagsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchAddTagsHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	transactionIds, err := utils.StringArrayToInt64Array(transactionBatchUpdateReq.TransactionIds)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchAddTagsHandler] parse transaction ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionIdInvalid
	}

	tagIds, err := utils.StringArrayToInt64Array(transactionBatchUpdateReq.TagIds)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchAddTagsHandler] parse tag ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionTagIdInvalid
	}

	tagIds = utils.ToUniqueInt64Slice(tagIds)

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionBatchAddTagsHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	tags, err := a.transactionTags.GetTagsByTagIds(c, uid, tagIds)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchAddTagsHandler] failed to get tags for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if len(tags) != len(tagIds) {
		log.Warnf(c, "[transactions.TransactionBatchAddTagsHandler] some tags do not exist for user \"uid:%d\"", uid)
		return nil, errs.ErrTransactionTagNotFound
	}

	transactions, err := a.transactions.GetTransactionsByTransactionIds(c, uid, transactionIds)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchAddTagsHandler] failed to get transactions for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allUsedAccounts, err := a.getTransactionUsedAccounts(c, uid, transactions)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchAddTagsHandler] failed to get transaction used accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactionTagIndexes, err := a.transactionTags.GetAllTagIdsOfTransactions(c, uid, transactionIds)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchAddTagsHandler] failed to get transactions tag indexes for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allNewTransactionTagIndexes := make(map[int64][]int64, len(transactions))

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			log.Warnf(c, "[transactions.TransactionBatchAddTagsHandler] cannot modify transaction \"id:%d\" for user \"uid:%d\", because transaction type is transfer in", transaction.TransactionId, uid)
			return nil, errs.ErrTransactionTypeInvalid
		}

		transactionEditable := user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone, allUsedAccounts[transaction.AccountId], allUsedAccounts[transaction.RelatedAccountId])

		if !transactionEditable {
			log.Warnf(c, "[transactions.TransactionBatchAddTagsHandler] transaction \"id:%d\" is not editable for user \"uid:%d\"", transaction.TransactionId, uid)
			return nil, errs.ErrCannotModifyTransactionWithThisTransactionTime
		}

		existedTagIds := transactionTagIndexes[transaction.TransactionId]
		existedTagIdsMap := make(map[int64]bool, len(existedTagIds))

		for j := 0; j < len(existedTagIds); j++ {
			existedTagIdsMap[existedTagIds[j]] = true
		}

		var newTagIds []int64

		for j := 0; j < len(tagIds); j++ {
			tagId := tagIds[j]

			if _, exists := existedTagIdsMap[tagId]; !exists {
				newTagIds = append(newTagIds, tagId)
			}
		}

		allNewTransactionTagIndexes[transaction.TransactionId] = newTagIds
	}

	err = a.transactions.BatchAddTagsToTransactions(c, uid, transactions, allNewTransactionTagIndexes)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchAddTagsHandler] failed to batch update transactions tags for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionBatchAddTagsHandler] user \"uid:%d\" has batch updated tag of %d transactions successfully", uid, len(allNewTransactionTagIndexes))
	return true, nil
}

// TransactionBatchRemoveTagsHandler batch remove tags from transactions by request parameters for current user
func (a *TransactionsApi) TransactionBatchRemoveTagsHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionBatchUpdateReq models.TransactionBatchRemoveTagsRequest
	err := c.ShouldBindJSON(&transactionBatchUpdateReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchRemoveTagsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchRemoveTagsHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	transactionIds, err := utils.StringArrayToInt64Array(transactionBatchUpdateReq.TransactionIds)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchRemoveTagsHandler] parse transaction ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionIdInvalid
	}

	tagIds, err := utils.StringArrayToInt64Array(transactionBatchUpdateReq.TagIds)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchRemoveTagsHandler] parse tag ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionTagIdInvalid
	}

	tagIds = utils.ToUniqueInt64Slice(tagIds)

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionBatchRemoveTagsHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	tags, err := a.transactionTags.GetTagsByTagIds(c, uid, tagIds)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchRemoveTagsHandler] failed to get tags for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if len(tags) != len(tagIds) {
		log.Warnf(c, "[transactions.TransactionBatchRemoveTagsHandler] some tags do not exist for user \"uid:%d\"", uid)
		return nil, errs.ErrTransactionTagNotFound
	}

	transactions, err := a.transactions.GetTransactionsByTransactionIds(c, uid, transactionIds)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchRemoveTagsHandler] failed to get transactions for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allUsedAccounts, err := a.getTransactionUsedAccounts(c, uid, transactions)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchRemoveTagsHandler] failed to get transaction used accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allTransactionIds := make([]int64, 0, len(transactions))

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			log.Warnf(c, "[transactions.TransactionBatchRemoveTagsHandler] cannot modify transaction \"id:%d\" for user \"uid:%d\", because transaction type is transfer in", transaction.TransactionId, uid)
			return nil, errs.ErrTransactionTypeInvalid
		}

		transactionEditable := user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone, allUsedAccounts[transaction.AccountId], allUsedAccounts[transaction.RelatedAccountId])

		if !transactionEditable {
			log.Warnf(c, "[transactions.TransactionBatchRemoveTagsHandler] transaction \"id:%d\" is not editable for user \"uid:%d\"", transaction.TransactionId, uid)
			return nil, errs.ErrCannotModifyTransactionWithThisTransactionTime
		}

		allTransactionIds = append(allTransactionIds, transaction.TransactionId)
	}

	err = a.transactions.BatchRemoveTagsFromTransactions(c, uid, allTransactionIds, tagIds)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchRemoveTagsHandler] failed to batch update transactions tags for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionBatchRemoveTagsHandler] user \"uid:%d\" has batch updated tag of %d transactions successfully", uid, len(allTransactionIds))
	return true, nil
}

// TransactionBatchClearTagsHandler batch clear all tags from transactions by request parameters for current user
func (a *TransactionsApi) TransactionBatchClearTagsHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionBatchUpdateReq models.TransactionBatchClearTagsRequest
	err := c.ShouldBindJSON(&transactionBatchUpdateReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchClearTagsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchClearTagsHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	transactionIds, err := utils.StringArrayToInt64Array(transactionBatchUpdateReq.TransactionIds)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionBatchClearTagsHandler] parse transaction ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionIdInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionBatchClearTagsHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	transactions, err := a.transactions.GetTransactionsByTransactionIds(c, uid, transactionIds)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchClearTagsHandler] failed to get transactions for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allUsedAccounts, err := a.getTransactionUsedAccounts(c, uid, transactions)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchClearTagsHandler] failed to get transaction used accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allTransactionIds := make([]int64, 0, len(transactions))

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			log.Warnf(c, "[transactions.TransactionBatchClearTagsHandler] cannot modify transaction \"id:%d\" for user \"uid:%d\", because transaction type is transfer in", transaction.TransactionId, uid)
			return nil, errs.ErrTransactionTypeInvalid
		}

		transactionEditable := user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone, allUsedAccounts[transaction.AccountId], allUsedAccounts[transaction.RelatedAccountId])

		if !transactionEditable {
			log.Warnf(c, "[transactions.TransactionBatchClearTagsHandler] transaction \"id:%d\" is not editable for user \"uid:%d\"", transaction.TransactionId, uid)
			return nil, errs.ErrCannotModifyTransactionWithThisTransactionTime
		}

		allTransactionIds = append(allTransactionIds, transaction.TransactionId)
	}

	err = a.transactions.BatchClearAllTagsFromTransactions(c, uid, allTransactionIds)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionBatchClearTagsHandler] failed to batch update transactions tags for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionBatchClearTagsHandler] user \"uid:%d\" has batch updated tag of %d transactions successfully", uid, len(allTransactionIds))
	return true, nil
}
