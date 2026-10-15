package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// TransactionGetHandler returns one specific transaction of current user
func (a *TransactionsApi) TransactionGetHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionGetReq models.TransactionGetRequest
	err := c.ShouldBindQuery(&transactionGetReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionGetHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionGetHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	transaction, err := a.transactions.GetTransactionByTransactionId(c, uid, transactionGetReq.Id)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionGetHandler] failed to get transaction \"id:%d\" for user \"uid:%d\", because %s", transactionGetReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		transaction = a.transactions.GetRelatedTransferTransaction(transaction)
	}

	accountIds := make([]int64, 0, 2)
	accountIds = append(accountIds, transaction.AccountId)

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
		accountIds = append(accountIds, transaction.RelatedAccountId)
		accountIds = utils.ToUniqueInt64Slice(accountIds)
	}

	accountMap, err := a.accounts.GetAccountsByAccountIds(c, uid, accountIds)

	if _, exists := accountMap[transaction.AccountId]; !exists {
		log.Warnf(c, "[transactions.TransactionGetHandler] account of transaction \"id:%d\" does not exist for user \"uid:%d\"", transaction.TransactionId, uid)
		return nil, errs.ErrTransactionNotFound
	}

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
		if _, exists := accountMap[transaction.RelatedAccountId]; !exists {
			log.Warnf(c, "[transactions.TransactionGetHandler] related account of transaction \"id:%d\" does not exist for user \"uid:%d\"", transaction.TransactionId, uid)
			return nil, errs.ErrTransactionNotFound
		}
	}

	allTransactionTagIds, err := a.transactionTags.GetAllTagIdsOfTransactions(c, uid, []int64{transaction.TransactionId})

	if err != nil {
		log.Errorf(c, "[transactions.TransactionGetHandler] failed to get transactions tag ids for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	var category *models.TransactionCategory
	var tagMap map[int64]*models.TransactionTag
	var pictureInfos []*models.TransactionPictureInfo

	if !transactionGetReq.TrimCategory && transaction.Type != models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
		category, err = a.transactionCategories.GetCategoryByCategoryId(c, uid, transaction.CategoryId)

		if err != nil {
			log.Errorf(c, "[transactions.TransactionGetHandler] failed to get transactions category for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	if !transactionGetReq.TrimTag {
		tagMap, err = a.transactionTags.GetTagsByTagIds(c, uid, utils.ToUniqueInt64Slice(a.transactionTags.GetTransactionTagIds(allTransactionTagIds)))

		if err != nil {
			log.Errorf(c, "[transactions.TransactionGetHandler] failed to get transactions tags for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	if transactionGetReq.WithPictures && a.CurrentConfig().EnableTransactionPictures {
		pictureInfos, err = a.transactionPictures.GetPictureInfosByTransactionId(c, uid, transaction.TransactionId)

		if err != nil {
			log.Errorf(c, "[transactions.TransactionGetHandler] failed to get transactions pictures for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	transactionEditable := transaction.IsEditable(user, clientTimezone, accountMap[transaction.AccountId], accountMap[transaction.RelatedAccountId])
	transactionTagIds := allTransactionTagIds[transaction.TransactionId]
	transactionResp := transaction.ToTransactionInfoResponse(transactionTagIds, transactionEditable)

	if !transactionGetReq.TrimAccount {
		if sourceAccount := accountMap[transaction.AccountId]; sourceAccount != nil {
			transactionResp.SourceAccount = sourceAccount.ToAccountInfoResponse()
		}

		if destinationAccount := accountMap[transaction.RelatedAccountId]; destinationAccount != nil {
			transactionResp.DestinationAccount = destinationAccount.ToAccountInfoResponse()
		}
	}

	if !transactionGetReq.TrimCategory {
		if category != nil {
			transactionResp.Category = category.ToTransactionCategoryInfoResponse()
		}
	}

	if !transactionGetReq.TrimTag {
		transactionResp.Tags = a.getTransactionTagInfoResponses(transactionTagIds, tagMap)
	}

	if transactionGetReq.WithPictures && a.CurrentConfig().EnableTransactionPictures {
		transactionResp.Pictures = a.GetTransactionPictureInfoResponseList(pictureInfos)
	}

	return transactionResp, nil
}
