package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
	"time"
	"xorm.io/xorm"
)

// BatchUpdateTransactionsCategory batch updates the categories of transactions
func (s *TransactionService) BatchUpdateTransactionsCategory(c core.Context, uid int64, transactionIds []int64, newCategoryId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if len(transactionIds) < 1 {
		return errs.ErrTransactionIdInvalid
	}

	if newCategoryId <= 0 {
		return errs.ErrTransactionCategoryIdInvalid
	}

	uniqueTransactionIds := utils.ToUniqueInt64Slice(transactionIds)
	now := time.Now().Unix()

	updateModel := &models.Transaction{
		CategoryId:      newCategoryId,
		UpdatedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.Cols("category_id", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).In("transaction_id", uniqueTransactionIds).Update(updateModel)

		if err != nil {
			return err
		} else if updatedRows < int64(len(uniqueTransactionIds)) {
			return errs.ErrTransactionNotFound
		}

		return err
	})
}

// BatchAddTagsToTransactions batch adds tags to transactions
func (s *TransactionService) BatchAddTagsToTransactions(c core.Context, uid int64, transactions []*models.Transaction, addTransactionTagIds map[int64][]int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if len(addTransactionTagIds) < 1 {
		return errs.ErrTransactionIdInvalid
	}

	now := time.Now().Unix()
	transactionTagIndexes := make([]*models.TransactionTagIndex, 0, len(addTransactionTagIds))
	transactionsMap := make(map[int64]*models.Transaction, len(transactions))
	transactionTagIdsMap := make(map[int64]bool, 0)

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]
		transactionsMap[transaction.TransactionId] = transaction
	}

	for transactionId, tagIds := range addTransactionTagIds {
		if transactionId <= 0 {
			return errs.ErrTransactionIdInvalid
		}

		transaction, exists := transactionsMap[transactionId]

		if !exists || transaction == nil {
			return errs.ErrTransactionNotFound
		}

		tagIds = utils.ToUniqueInt64Slice(tagIds)

		for i := 0; i < len(tagIds); i++ {
			tagId := tagIds[i]

			if tagId <= 0 {
				return errs.ErrTransactionTagIdInvalid
			}

			transactionTagIndexes = append(transactionTagIndexes, &models.TransactionTagIndex{
				Uid:             uid,
				Deleted:         false,
				TransactionTime: transaction.TransactionTime,
				TagId:           tagId,
				TransactionId:   transactionId,
				CreatedUnixTime: now,
				UpdatedUnixTime: now,
			})

			transactionTagIdsMap[tagId] = true
		}
	}

	tagIndexUuids := s.GenerateUuids(uuid.UUID_TYPE_TAG_INDEX, uint16(len(transactionTagIndexes)))

	if len(tagIndexUuids) < len(transactionTagIndexes) {
		return errs.ErrCannotAddTagsToTooManyTransactionsOneTime
	}

	for i := 0; i < len(transactionTagIndexes); i++ {
		transactionTagIndexes[i].TagIndexId = tagIndexUuids[i]
	}

	tagIds := make([]int64, 0, len(transactionTagIdsMap))

	for tagId := range transactionTagIdsMap {
		tagIds = append(tagIds, tagId)
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		// Get and verify tags
		err := s.isTagsValid(sess, uid, transactionTagIndexes, tagIds)

		if err != nil {
			return err
		}

		for i := 0; i < len(transactionTagIndexes); i++ {
			transactionTagIndex := transactionTagIndexes[i]
			_, err := sess.Insert(transactionTagIndex)

			if err != nil {
				return err
			}
		}

		return nil
	})
}

// BatchRemoveTagsFromTransactions batch removes tags from transactions
func (s *TransactionService) BatchRemoveTagsFromTransactions(c core.Context, uid int64, transactionIds []int64, tagIds []int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if len(transactionIds) < 1 {
		return errs.ErrTransactionIdInvalid
	}

	uniqueTransactionIds := utils.ToUniqueInt64Slice(transactionIds)
	uniqueTagIds := utils.ToUniqueInt64Slice(tagIds)
	now := time.Now().Unix()

	tagIndexUpdateModel := &models.TransactionTagIndex{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		deletedRows, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).In("transaction_id", uniqueTransactionIds).In("tag_id", uniqueTagIds).Update(tagIndexUpdateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrTransactionTagNotFound
		}

		return nil
	})
}

// BatchClearAllTagsFromTransactions batch clears all tags from transactions
func (s *TransactionService) BatchClearAllTagsFromTransactions(c core.Context, uid int64, transactionIds []int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if len(transactionIds) < 1 {
		return errs.ErrTransactionIdInvalid
	}

	uniqueTransactionIds := utils.ToUniqueInt64Slice(transactionIds)
	now := time.Now().Unix()

	tagIndexUpdateModel := &models.TransactionTagIndex{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		deletedRows, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).In("transaction_id", uniqueTransactionIds).Update(tagIndexUpdateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrTransactionTagNotFound
		}

		return nil
	})
}
