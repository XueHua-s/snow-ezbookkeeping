package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

const pageCountForLoadAllTransactionTagIndexes = 1000

// TransactionTagService represents transaction tag service
type TransactionTagService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a transaction tag service singleton instance
var (
	TransactionTags = &TransactionTagService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetTotalTagCountByUid returns total tag count of user
func (s *TransactionTagService) GetTotalTagCountByUid(c core.Context, uid int64) (int64, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	count, err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).Count(&models.TransactionTag{})

	return count, err
}

// GetAllTagsByUid returns all transaction tag models of user
func (s *TransactionTagService) GetAllTagsByUid(c core.Context, uid int64) ([]*models.TransactionTag, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var tags []*models.TransactionTag
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).Find(&tags)

	return tags, err
}

// GetTagByTagId returns a transaction tag model according to transaction tag id
func (s *TransactionTagService) GetTagByTagId(c core.Context, uid int64, tagId int64) (*models.TransactionTag, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if tagId <= 0 {
		return nil, errs.ErrTransactionTagIdInvalid
	}

	tag := &models.TransactionTag{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(tagId).Where("uid=? AND deleted=?", uid, false).Get(tag)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrTransactionTagNotFound
	}

	return tag, nil
}

// GetTagsByTagIds returns transaction tag models according to transaction tag ids
func (s *TransactionTagService) GetTagsByTagIds(c core.Context, uid int64, tagIds []int64) (map[int64]*models.TransactionTag, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if tagIds == nil {
		return nil, errs.ErrTransactionTagIdInvalid
	}

	var tags []*models.TransactionTag
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).In("tag_id", tagIds).Find(&tags)

	if err != nil {
		return nil, err
	}

	tagMap := s.GetTagMapByList(tags)
	return tagMap, err
}

// GetMaxDisplayOrder returns the max display order
func (s *TransactionTagService) GetMaxDisplayOrder(c core.Context, uid int64, tagGroupId int64) (int32, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	tag := &models.TransactionTag{}
	has, err := s.UserDataDB(uid).NewSession(c).Cols("uid", "deleted", "display_order").Where("uid=? AND deleted=? AND tag_group_id=?", uid, false, tagGroupId).OrderBy("display_order desc").Limit(1).Get(tag)

	if err != nil {
		return 0, err
	}

	if has {
		return tag.DisplayOrder, nil
	} else {
		return 0, nil
	}
}

// GetAllTagIdsOfAllTransactions returns all transaction tag ids
func (s *TransactionTagService) GetAllTagIdsOfAllTransactions(c core.Context, uid int64) ([]*models.TransactionTagIndex, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	condition := "uid=? AND deleted=?"
	conditionParams := make([]any, 0, 2)
	conditionParams = append(conditionParams, uid)
	conditionParams = append(conditionParams, false)

	var allTransactionTagIndexes []*models.TransactionTagIndex

	maxTransactionTagIndexId := int64(0)

	for maxTransactionTagIndexId >= 0 {
		var tagIndexes []*models.TransactionTagIndex

		finalCondition := condition
		finalConditionParams := make([]any, 0, 3)
		finalConditionParams = append(finalConditionParams, conditionParams...)

		if maxTransactionTagIndexId > 0 {
			finalCondition = finalCondition + " AND tag_index_id<=?"
			finalConditionParams = append(finalConditionParams, maxTransactionTagIndexId)
		}

		err := s.UserDataDB(uid).NewSession(c).Where(finalCondition, finalConditionParams...).Limit(pageCountForLoadAllTransactionTagIndexes, 0).OrderBy("tag_index_id desc").Find(&tagIndexes)

		if err != nil {
			return nil, err
		}

		allTransactionTagIndexes = append(allTransactionTagIndexes, tagIndexes...)

		if len(tagIndexes) < pageCountForLoadAllTransactionTagIndexes {
			maxTransactionTagIndexId = -1
			break
		}

		maxTransactionTagIndexId = tagIndexes[len(tagIndexes)-1].TagIndexId - 1
	}

	return allTransactionTagIndexes, nil
}

// GetAllTagIdsMapOfAllTransactions returns all transaction tag ids map grouped by transaction id
func (s *TransactionTagService) GetAllTagIdsMapOfAllTransactions(c core.Context, uid int64) (map[int64][]int64, error) {
	tagIndexes, err := s.GetAllTagIdsOfAllTransactions(c, uid)

	if err != nil {
		return nil, err
	}

	allTransactionTagIds := s.GetGroupedTransactionTagIds(tagIndexes)

	return allTransactionTagIds, err
}

// GetAllTagIdsOfTransactions returns transaction tag ids for given transactions
func (s *TransactionTagService) GetAllTagIdsOfTransactions(c core.Context, uid int64, transactionIds []int64) (map[int64][]int64, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var tagIndexes []*models.TransactionTagIndex
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).In("transaction_id", transactionIds).OrderBy("transaction_id asc, tag_index_id asc").Find(&tagIndexes)

	allTransactionTagIds := s.GetGroupedTransactionTagIds(tagIndexes)

	return allTransactionTagIds, err
}
