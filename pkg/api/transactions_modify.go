package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"strings"
)

// TransactionModifyHandler saves an existed transaction by request parameters for current user
func (a *TransactionsApi) TransactionModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionModifyReq models.TransactionModifyRequest
	err := c.ShouldBindJSON(&transactionModifyReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionModifyHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	tagIds, err := utils.StringArrayToInt64Array(transactionModifyReq.TagIds)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionModifyHandler] parse tag ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionTagIdInvalid
	}

	if len(tagIds) > models.MaximumTagsCountOfTransaction {
		return nil, errs.ErrTransactionHasTooManyTags
	}

	pictureIds, err := utils.StringArrayToInt64Array(transactionModifyReq.PictureIds)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionModifyHandler] parse picture ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionPictureIdInvalid
	}

	if len(pictureIds) > models.MaximumPicturesCountOfTransaction {
		return nil, errs.ErrTransactionHasTooManyPictures
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionModifyHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	transaction, err := a.transactions.GetTransactionByTransactionId(c, uid, transactionModifyReq.Id)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionModifyHandler] failed to get transaction \"id:%d\" for user \"uid:%d\", because %s", transactionModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		log.Warnf(c, "[transactions.TransactionModifyHandler] cannot modify transaction \"id:%d\" for user \"uid:%d\", because transaction type is transfer in", transactionModifyReq.Id, uid)
		return nil, errs.ErrTransactionTypeInvalid
	}

	newTransactionType, err := transactionModifyReq.Type.ToTransactionDbType()

	if err != nil {
		return nil, errs.ErrTransactionTypeInvalid
	}

	if (transaction.Type == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE && newTransactionType != models.TRANSACTION_DB_TYPE_MODIFY_BALANCE) ||
		(transaction.Type != models.TRANSACTION_DB_TYPE_MODIFY_BALANCE && newTransactionType == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE) {
		log.Warnf(c, "[transactions.TransactionModifyHandler] cannot modify transaction type from \"%d\" to \"%d\"", transaction.Type, newTransactionType)
		return nil, errs.ErrTransactionTypeInvalid
	}

	changeToTransfer := newTransactionType == models.TRANSACTION_DB_TYPE_TRANSFER_OUT && transaction.Type != models.TRANSACTION_DB_TYPE_TRANSFER_OUT

	if transaction.Type == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE && transactionModifyReq.CategoryId != 0 {
		log.Warnf(c, "[transactions.TransactionModifyHandler] balance modification transaction cannot set category id")
		return nil, errs.ErrBalanceModificationTransactionCannotSetCategory
	} else if transaction.Type != models.TRANSACTION_DB_TYPE_MODIFY_BALANCE && transactionModifyReq.CategoryId == 0 {
		log.Warnf(c, "[transactions.TransactionModifyHandler] non-balance modification transaction must set category id")
		return nil, errs.ErrIncompleteOrIncorrectSubmission
	}

	allTransactionTagIds, err := a.transactionTags.GetAllTagIdsOfTransactions(c, uid, []int64{transaction.TransactionId})

	if err != nil {
		log.Errorf(c, "[transactions.TransactionModifyHandler] failed to get transactions tag ids for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactionTagIds := allTransactionTagIds[transaction.TransactionId]

	if transactionTagIds == nil {
		transactionTagIds = make([]int64, 0, 0)
	}

	transactionPictureInfos, err := a.transactionPictures.GetPictureInfosByTransactionId(c, uid, transaction.TransactionId)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionModifyHandler] failed to get transaction picture infos for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactionPictureIds := a.transactionPictures.GetTransactionPictureIds(transactionPictureInfos)

	newTransaction := &models.Transaction{
		TransactionId:     transaction.TransactionId,
		Uid:               uid,
		Type:              newTransactionType,
		CategoryId:        transactionModifyReq.CategoryId,
		TransactionTime:   utils.GetMinTransactionTimeFromUnixTime(transactionModifyReq.Time),
		TimezoneUtcOffset: transactionModifyReq.UtcOffset,
		AccountId:         transactionModifyReq.SourceAccountId,
		Amount:            transactionModifyReq.SourceAmount,
		HideAmount:        transactionModifyReq.HideAmount,
		Comment:           transactionModifyReq.Comment,
	}

	if newTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
		newTransaction.RelatedAccountId = transactionModifyReq.DestinationAccountId
		newTransaction.RelatedAccountAmount = transactionModifyReq.DestinationAmount
	}

	if transactionModifyReq.GeoLocation != nil {
		newTransaction.GeoLongitude = transactionModifyReq.GeoLocation.Longitude
		newTransaction.GeoLatitude = transactionModifyReq.GeoLocation.Latitude
	}

	if newTransaction.Type == transaction.Type &&
		newTransaction.CategoryId == transaction.CategoryId &&
		utils.GetUnixTimeFromTransactionTime(newTransaction.TransactionTime) == utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime) &&
		newTransaction.TimezoneUtcOffset == transaction.TimezoneUtcOffset &&
		newTransaction.AccountId == transaction.AccountId &&
		newTransaction.Amount == transaction.Amount &&
		(newTransaction.Type != models.TRANSACTION_DB_TYPE_TRANSFER_OUT || newTransaction.RelatedAccountId == transaction.RelatedAccountId) &&
		(newTransaction.Type != models.TRANSACTION_DB_TYPE_TRANSFER_OUT || newTransaction.RelatedAccountAmount == transaction.RelatedAccountAmount) &&
		newTransaction.HideAmount == transaction.HideAmount &&
		newTransaction.Comment == transaction.Comment &&
		newTransaction.GeoLongitude == transaction.GeoLongitude &&
		newTransaction.GeoLatitude == transaction.GeoLatitude &&
		utils.Int64SliceEquals(tagIds, transactionTagIds) &&
		utils.Int64SliceEquals(pictureIds, transactionPictureIds) {
		return nil, errs.ErrNothingWillBeUpdated
	}

	allUsedAccounts, err := a.getTransactionUsedAccounts(c, uid, []*models.Transaction{transaction, newTransaction})

	if err != nil {
		log.Errorf(c, "[transactions.TransactionModifyHandler] failed to get transaction used accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactionEditable := user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone, allUsedAccounts[transaction.AccountId], allUsedAccounts[transaction.RelatedAccountId])
	newTransactionEditable := user.CanEditTransactionByTransactionTime(newTransaction.TransactionTime, clientTimezone, allUsedAccounts[newTransaction.AccountId], allUsedAccounts[newTransaction.RelatedAccountId])

	if !transactionEditable || !newTransactionEditable {
		return nil, errs.ErrCannotModifyTransactionWithThisTransactionTime
	}

	var addTransactionTagIds []int64
	var removeTransactionTagIds []int64

	if !utils.Int64SliceEquals(tagIds, transactionTagIds) {
		removeTransactionTagIds = transactionTagIds
		addTransactionTagIds = tagIds
	}

	addTransactionPictureIds := utils.Int64SliceMinus(pictureIds, transactionPictureIds)
	removeTransactionPictureIds := utils.Int64SliceMinus(transactionPictureIds, pictureIds)
	var newPictureInfos []*models.TransactionPictureInfo

	if !utils.Int64SliceEquals(pictureIds, transactionPictureIds) {
		oldAndNewPictureIds := transactionPictureIds
		oldAndNewPictureInfoMap := a.transactionPictures.GetPictureInfoMapByList(transactionPictureInfos)

		if len(addTransactionPictureIds) > 0 {
			addPictureInfos, err := a.transactionPictures.GetNewPictureInfosByPictureIds(c, uid, addTransactionPictureIds)

			if err != nil {
				log.Errorf(c, "[transactions.TransactionModifyHandler] failed to get transactions pictures for user \"uid:%d\", because %s", uid, err.Error())
				return nil, errs.Or(err, errs.ErrOperationFailed)
			}

			oldAndNewPictureIds = append(oldAndNewPictureIds, a.transactionPictures.GetTransactionPictureIds(addPictureInfos)...)
			notExistsPictureIds := utils.Int64SliceMinus(pictureIds, oldAndNewPictureIds)

			if len(notExistsPictureIds) > 0 {
				log.Errorf(c, "[transactions.TransactionModifyHandler] some pictures \"ids:%s\" does not exists for user \"uid:%d\"", strings.Join(utils.Int64ArrayToStringArray(notExistsPictureIds), ","), uid)
				return nil, errs.ErrTransactionPictureNotFound
			}

			for i := 0; i < len(addPictureInfos); i++ {
				oldAndNewPictureInfoMap[addPictureInfos[i].PictureId] = addPictureInfos[i]
			}
		}

		for i := 0; i < len(pictureIds); i++ {
			pictureId := pictureIds[i]
			pictureInfo, exists := oldAndNewPictureInfoMap[pictureId]

			if exists {
				newPictureInfos = append(newPictureInfos, pictureInfo)
			}
		}
	}

	err = a.transactions.ModifyTransaction(c, newTransaction, changeToTransfer, len(transactionTagIds), addTransactionTagIds, removeTransactionTagIds, addTransactionPictureIds, removeTransactionPictureIds)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionModifyHandler] failed to update transaction \"id:%d\" for user \"uid:%d\", because %s", transactionModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionModifyHandler] user \"uid:%d\" has updated transaction \"id:%d\" successfully", uid, transactionModifyReq.Id)

	newTransactionResp := newTransaction.ToTransactionInfoResponse(tagIds, transactionEditable)
	newTransactionResp.Pictures = a.GetTransactionPictureInfoResponseList(newPictureInfos)

	return newTransactionResp, nil
}
