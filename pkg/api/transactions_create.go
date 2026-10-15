package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/duplicatechecker"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"strings"
)

// TransactionCreateHandler saves a new transaction by request parameters for current user
func (a *TransactionsApi) TransactionCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionCreateReq models.TransactionCreateRequest
	err := c.ShouldBindJSON(&transactionCreateReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionCreateHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	tagIds, err := utils.StringArrayToInt64Array(transactionCreateReq.TagIds)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionCreateHandler] parse tag ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionTagIdInvalid
	}

	if len(tagIds) > models.MaximumTagsCountOfTransaction {
		return nil, errs.ErrTransactionHasTooManyTags
	}

	pictureIds, err := utils.StringArrayToInt64Array(transactionCreateReq.PictureIds)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionCreateHandler] parse picture ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionPictureIdInvalid
	}

	if len(pictureIds) > models.MaximumPicturesCountOfTransaction {
		return nil, errs.ErrTransactionHasTooManyPictures
	}

	if transactionCreateReq.Type < models.TRANSACTION_TYPE_MODIFY_BALANCE || transactionCreateReq.Type > models.TRANSACTION_TYPE_TRANSFER {
		log.Warnf(c, "[transactions.TransactionCreateHandler] transaction type is invalid")
		return nil, errs.ErrTransactionTypeInvalid
	}

	if transactionCreateReq.Type == models.TRANSACTION_TYPE_MODIFY_BALANCE && transactionCreateReq.CategoryId != 0 {
		log.Warnf(c, "[transactions.TransactionCreateHandler] balance modification transaction cannot set category id")
		return nil, errs.ErrBalanceModificationTransactionCannotSetCategory
	}

	if transactionCreateReq.Type != models.TRANSACTION_TYPE_TRANSFER && transactionCreateReq.DestinationAccountId != 0 {
		log.Warnf(c, "[transactions.TransactionCreateHandler] non-transfer transaction destination account cannot be set")
		return nil, errs.ErrTransactionDestinationAccountCannotBeSet
	} else if transactionCreateReq.Type == models.TRANSACTION_TYPE_TRANSFER && transactionCreateReq.SourceAccountId == transactionCreateReq.DestinationAccountId {
		log.Warnf(c, "[transactions.TransactionCreateHandler] transfer transaction source account must not be destination account")
		return nil, errs.ErrTransactionSourceAndDestinationIdCannotBeEqual
	}

	if transactionCreateReq.Type != models.TRANSACTION_TYPE_TRANSFER && transactionCreateReq.DestinationAmount != 0 {
		log.Warnf(c, "[transactions.TransactionCreateHandler] non-transfer transaction destination amount cannot be set")
		return nil, errs.ErrTransactionDestinationAmountCannotBeSet
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionCreateHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	transaction := a.createNewTransactionModel(uid, &transactionCreateReq, c.ClientIP())

	allUsedAccounts, err := a.getTransactionUsedAccounts(c, uid, []*models.Transaction{transaction})

	if err != nil {
		log.Errorf(c, "[transactions.TransactionCreateHandler] failed to get transaction used accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactionEditable := user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone, allUsedAccounts[transaction.AccountId], allUsedAccounts[transaction.RelatedAccountId])

	if !transactionEditable {
		return nil, errs.ErrCannotCreateTransactionWithThisTransactionTime
	}

	var pictureInfos []*models.TransactionPictureInfo

	if len(pictureIds) > 0 {
		pictureInfos, err = a.transactionPictures.GetNewPictureInfosByPictureIds(c, uid, pictureIds)

		if err != nil {
			log.Errorf(c, "[transactions.TransactionCreateHandler] failed to get transactions pictures for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		notExistsPictureIds := utils.Int64SliceMinus(pictureIds, a.transactionPictures.GetTransactionPictureIds(pictureInfos))

		if len(notExistsPictureIds) > 0 {
			log.Errorf(c, "[transactions.TransactionCreateHandler] some pictures \"ids:%s\" does not exists for user \"uid:%d\"", strings.Join(utils.Int64ArrayToStringArray(notExistsPictureIds), ","), uid)
			return nil, errs.ErrTransactionPictureNotFound
		}
	}

	if a.CurrentConfig().EnableDuplicateSubmissionsCheck && transactionCreateReq.ClientSessionId != "" {
		found, remark := a.GetSubmissionRemark(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_TRANSACTION, uid, transactionCreateReq.ClientSessionId)

		if found {
			log.Infof(c, "[transactions.TransactionCreateHandler] another transaction \"id:%s\" has been created for user \"uid:%d\"", remark, uid)
			transactionId, err := utils.StringToInt64(remark)

			if err == nil {
				transaction, err = a.transactions.GetTransactionByTransactionId(c, uid, transactionId)

				if err != nil {
					log.Errorf(c, "[transactions.TransactionCreateHandler] failed to get existed transaction \"id:%d\" for user \"uid:%d\", because %s", transactionId, uid, err.Error())
					return nil, errs.Or(err, errs.ErrOperationFailed)
				}

				transactionResp := transaction.ToTransactionInfoResponse(tagIds, transactionEditable)
				transactionResp.Pictures = a.GetTransactionPictureInfoResponseList(pictureInfos)

				return transactionResp, nil
			}
		}
	}

	err = a.transactions.CreateTransaction(c, transaction, tagIds, pictureIds)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionCreateHandler] failed to create transaction \"id:%d\" for user \"uid:%d\", because %s", transaction.TransactionId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionCreateHandler] user \"uid:%d\" has created a new transaction \"id:%d\" successfully", uid, transaction.TransactionId)

	a.SetSubmissionRemarkIfEnable(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_TRANSACTION, uid, transactionCreateReq.ClientSessionId, utils.Int64ToString(transaction.TransactionId))
	transactionResp := transaction.ToTransactionInfoResponse(tagIds, transactionEditable)
	transactionResp.Pictures = a.GetTransactionPictureInfoResponseList(pictureInfos)

	return transactionResp, nil
}
