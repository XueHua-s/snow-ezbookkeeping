package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
	"math"
	"time"
	"xorm.io/xorm"
)

// CreateTransaction saves a new transaction to database
func (s *TransactionService) CreateTransaction(c core.Context, transaction *models.Transaction, tagIds []int64, pictureIds []int64) error {
	if transaction.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	// Check whether account id is valid
	err := s.isAccountIdValid(transaction)

	if err != nil {
		return err
	}

	now := time.Now().Unix()

	needTransactionUuidCount := 1

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		needTransactionUuidCount = 2
	}

	transactionUuids := s.GenerateUuids(uuid.UUID_TYPE_TRANSACTION, uint16(needTransactionUuidCount))

	if len(transactionUuids) < needTransactionUuidCount {
		return errs.ErrSystemIsBusy
	}

	tagIds = utils.ToUniqueInt64Slice(tagIds)
	needTagIndexUuidCount := uint16(len(tagIds))
	tagIndexUuids := s.GenerateUuids(uuid.UUID_TYPE_TAG_INDEX, needTagIndexUuidCount)

	if len(tagIndexUuids) < int(needTagIndexUuidCount) {
		return errs.ErrSystemIsBusy
	}

	transaction.TransactionId = transactionUuids[0]

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		transaction.RelatedId = transactionUuids[1]
	}

	transaction.TransactionTime = utils.GetMinTransactionTimeFromUnixTime(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime))

	transaction.CreatedUnixTime = now
	transaction.UpdatedUnixTime = now

	transactionTagIndexes := make([]*models.TransactionTagIndex, len(tagIds))

	for i := 0; i < len(tagIds); i++ {
		transactionTagIndexes[i] = &models.TransactionTagIndex{
			TagIndexId:      tagIndexUuids[i],
			Uid:             transaction.Uid,
			Deleted:         false,
			TagId:           tagIds[i],
			TransactionId:   transaction.TransactionId,
			CreatedUnixTime: now,
			UpdatedUnixTime: now,
		}
	}

	pictureUpdateModel := &models.TransactionPictureInfo{
		TransactionId:   transaction.TransactionId,
		UpdatedUnixTime: now,
	}

	userDataDb := s.UserDataDB(transaction.Uid)

	return userDataDb.DoTransaction(c, func(sess *xorm.Session) error {
		return s.doCreateTransaction(c, userDataDb, sess, transaction, transactionTagIndexes, tagIds, pictureIds, pictureUpdateModel)
	})
}

// BatchCreateTransactions saves new transactions to database
func (s *TransactionService) BatchCreateTransactions(c core.Context, uid int64, transactions []*models.Transaction, allTagIds map[int][]int64, processHandler core.TaskProcessUpdateHandler) error {
	now := time.Now().Unix()
	currentProcess := float64(0)
	processUpdateStep := int(math.Max(100.0, float64(len(transactions)/100.0)))

	needTransactionUuidCount := uint16(0)
	needTagIndexUuidCount := uint16(0)

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]

		if transaction.Uid != uid {
			return errs.ErrUserIdInvalid
		}

		// Check whether account id is valid
		err := s.isAccountIdValid(transaction)

		if err != nil {
			return err
		}

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			needTransactionUuidCount += 2
		} else {
			needTransactionUuidCount++
		}

		transaction.TransactionTime = utils.GetMinTransactionTimeFromUnixTime(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime))

		transaction.CreatedUnixTime = now
		transaction.UpdatedUnixTime = now
	}

	for index, tagIds := range allTagIds {
		if index < 0 || index >= len(transactions) {
			return errs.ErrOperationFailed
		}

		uniqueTagIds := utils.ToUniqueInt64Slice(tagIds)
		needTagIndexUuidCount += uint16(len(uniqueTagIds))
	}

	if needTransactionUuidCount > uint16(65535) || needTagIndexUuidCount > uint16(65535) {
		return errs.ErrImportTooManyTransaction
	}

	transactionUuids := s.GenerateUuids(uuid.UUID_TYPE_TRANSACTION, needTransactionUuidCount)
	transactionUuidIndex := 0

	if len(transactionUuids) < int(needTransactionUuidCount) {
		return errs.ErrSystemIsBusy
	}

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]

		transaction.TransactionId = transactionUuids[transactionUuidIndex]
		transactionUuidIndex++

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			transaction.RelatedId = transactionUuids[transactionUuidIndex]
			transactionUuidIndex++
		}
	}

	tagIndexUuids := s.GenerateUuids(uuid.UUID_TYPE_TAG_INDEX, needTagIndexUuidCount)
	tagIndexUuidIndex := 0

	if len(tagIndexUuids) < int(needTagIndexUuidCount) {
		return errs.ErrSystemIsBusy
	}

	allTransactionTagIndexes := make(map[int64][]*models.TransactionTagIndex)
	allTransactionTagIds := make(map[int64][]int64)

	for index, tagIds := range allTagIds {
		transaction := transactions[index]
		uniqueTagIds := utils.ToUniqueInt64Slice(tagIds)

		transactionTagIndexes := make([]*models.TransactionTagIndex, len(uniqueTagIds))

		for i := 0; i < len(uniqueTagIds); i++ {
			transactionTagIndexes[i] = &models.TransactionTagIndex{
				TagIndexId:      tagIndexUuids[tagIndexUuidIndex],
				Uid:             transaction.Uid,
				Deleted:         false,
				TagId:           uniqueTagIds[i],
				TransactionId:   transaction.TransactionId,
				CreatedUnixTime: now,
				UpdatedUnixTime: now,
			}

			tagIndexUuidIndex++
		}

		allTransactionTagIndexes[transaction.TransactionId] = transactionTagIndexes
		allTransactionTagIds[transaction.TransactionId] = uniqueTagIds
	}

	userDataDb := s.UserDataDB(uid)

	return userDataDb.DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(transactions); i++ {
			transaction := transactions[i]
			transactionTagIndexes := allTransactionTagIndexes[transaction.TransactionId]
			transactionTagIds := allTransactionTagIds[transaction.TransactionId]
			err := s.doCreateTransaction(c, userDataDb, sess, transaction, transactionTagIndexes, transactionTagIds, nil, nil)

			currentProcess = float64(i) / float64(len(transactions)) * 100

			if processHandler != nil && i%processUpdateStep == 0 {
				processHandler(currentProcess)
			}

			if err != nil {
				transactionUnixTime := utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime)
				transactionTimeZone := time.FixedZone("Transaction Timezone", int(transaction.TimezoneUtcOffset)*60)
				log.Errorf(c, "[transactions.BatchCreateTransactions] failed to create trasaction (datetime: %s, type: %s, amount: %d)", utils.FormatUnixTimeToLongDateTime(transactionUnixTime, transactionTimeZone), transaction.Type, transaction.Amount)
				return err
			}
		}

		return nil
	})
}
