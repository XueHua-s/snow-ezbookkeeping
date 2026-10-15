package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"strings"
	"time"
	"xorm.io/xorm"
)

// ExistsTagName returns whether the given tag name exists
func (s *TransactionTagService) ExistsTagName(c core.Context, uid int64, name string) (bool, error) {
	if name == "" {
		return false, errs.ErrTransactionTagNameIsEmpty
	}

	return s.UserDataDB(uid).NewSession(c).Cols("name").Where("uid=? AND deleted=? AND name=?", uid, false, name).Exist(&models.TransactionTag{})
}

// ModifyTagIndexTransactionTime updates transaction time of given transaction tag indexes
func (s *TransactionTagService) ModifyTagIndexTransactionTime(c core.Context, uid int64, tagIndexes []*models.TransactionTagIndex) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	for i := 0; i < len(tagIndexes); i++ {
		tagIndexes[i].UpdatedUnixTime = time.Now().Unix()
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(tagIndexes); i++ {
			tagIndex := tagIndexes[i]
			updatedRows, err := sess.ID(tagIndex.TagIndexId).Cols("transaction_time", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(tagIndex)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				return errs.ErrTransactionTagIndexNotFound
			}
		}

		return nil
	})
}

// GetTagMapByList returns a transaction tag map by a list
func (s *TransactionTagService) GetTagMapByList(tags []*models.TransactionTag) map[int64]*models.TransactionTag {
	tagMap := make(map[int64]*models.TransactionTag)

	for i := 0; i < len(tags); i++ {
		tag := tags[i]
		tagMap[tag.TagId] = tag
	}
	return tagMap
}

// GetVisibleTagNameMapByList returns a visible transaction tag map by a list
func (s *TransactionTagService) GetVisibleTagNameMapByList(tags []*models.TransactionTag) map[string]*models.TransactionTag {
	tagMap := make(map[string]*models.TransactionTag)

	for i := 0; i < len(tags); i++ {
		tag := tags[i]

		if tag.Hidden {
			continue
		}

		tagMap[tag.Name] = tag
	}

	return tagMap
}

// GetTagNames returns a list with tag names from tag models list
func (s *TransactionTagService) GetTagNames(tags []*models.TransactionTag) []string {
	tagNames := make([]string, len(tags))

	for i := 0; i < len(tags); i++ {
		tagNames[i] = tags[i].Name
	}

	return tagNames
}

// GetGroupedTransactionTagIds returns a map of transaction tag ids grouped by transaction id
func (s *TransactionTagService) GetGroupedTransactionTagIds(tagIndexes []*models.TransactionTagIndex) map[int64][]int64 {
	allTransactionTagIds := make(map[int64][]int64)

	for i := 0; i < len(tagIndexes); i++ {
		tagIndex := tagIndexes[i]

		var transactionTagIds []int64

		if _, exists := allTransactionTagIds[tagIndex.TransactionId]; exists {
			transactionTagIds = allTransactionTagIds[tagIndex.TransactionId]
		}

		transactionTagIds = append(transactionTagIds, tagIndex.TagId)
		allTransactionTagIds[tagIndex.TransactionId] = transactionTagIds
	}

	for _, tagIds := range allTransactionTagIds {
		utils.Int64Sort(tagIds)
	}

	return allTransactionTagIds
}

// GetTagIds converts a comma-separated string of tag ids into a slice of int64
func (s *TransactionTagService) GetTagIds(tagIds string) ([]int64, error) {
	if tagIds == "" || tagIds == "0" {
		return nil, nil
	}

	requestTagIds, err := utils.StringArrayToInt64Array(strings.Split(tagIds, ","))

	if err != nil {
		return nil, errs.Or(err, errs.ErrTransactionTagIdInvalid)
	}

	return requestTagIds, nil
}

// GetTransactionTagIds returns a slice of all transaction tag ids from a map of transaction tag ids grouped by transaction id
func (s *TransactionTagService) GetTransactionTagIds(allTransactionTagIds map[int64][]int64) []int64 {
	allTagIds := make([]int64, 0, len(allTransactionTagIds))

	for _, tagIds := range allTransactionTagIds {
		allTagIds = append(allTagIds, tagIds...)
	}

	return allTagIds
}
