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

// CreateTag saves a new transaction tag model to database
func (s *TransactionTagService) CreateTag(c core.Context, tag *models.TransactionTag) error {
	if tag.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	exists, err := s.ExistsTagName(c, tag.Uid, tag.Name)

	if err != nil {
		return err
	} else if exists {
		return errs.ErrTransactionTagNameAlreadyExists
	}

	tag.TagId = s.GenerateUuid(uuid.UUID_TYPE_TAG)

	if tag.TagId < 1 {
		return errs.ErrSystemIsBusy
	}

	tag.Deleted = false
	tag.CreatedUnixTime = time.Now().Unix()
	tag.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(tag.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(tag)
		return err
	})
}

// CreateTags saves a few transaction tag models to database
func (s *TransactionTagService) CreateTags(c core.Context, uid int64, tags []*models.TransactionTag, skipExists bool) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	allTagNames := make([]string, len(tags))

	for i := 0; i < len(tags); i++ {
		allTagNames[i] = tags[i].Name
	}

	var existTags []*models.TransactionTag
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).In("name", allTagNames).Find(&existTags)

	if err != nil {
		return err
	} else if !skipExists && len(existTags) > 0 {
		return errs.ErrTransactionTagNameAlreadyExists
	}

	existsNameTagMap := make(map[string]*models.TransactionTag, len(existTags))

	for i := 0; i < len(existTags); i++ {
		tag := existTags[i]
		existsNameTagMap[tag.Name] = tag
	}

	newTags := make([]*models.TransactionTag, 0, len(tags)-len(existTags))

	for i := 0; i < len(tags); i++ {
		tag := tags[i]
		existsTag, exists := existsNameTagMap[tag.Name]

		if exists {
			tag.FillFromOtherTag(existsTag)
			continue
		}

		newTags = append(newTags, tag)
	}

	tagUuids := s.GenerateUuids(uuid.UUID_TYPE_TAG_INDEX, uint16(len(newTags)))

	if len(tagUuids) < len(newTags) {
		return errs.ErrSystemIsBusy
	}

	for i := 0; i < len(newTags); i++ {
		tag := newTags[i]
		tag.TagId = tagUuids[i]
		tag.Deleted = false
		tag.CreatedUnixTime = time.Now().Unix()
		tag.UpdatedUnixTime = time.Now().Unix()
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(newTags); i++ {
			tag := newTags[i]
			_, err := sess.Insert(tag)

			if err != nil {
				return err
			}
		}

		return nil
	})
}

// ModifyTag saves an existed transaction tag model to database
func (s *TransactionTagService) ModifyTag(c core.Context, tag *models.TransactionTag, tagNameChanged bool) error {
	if tag.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if tagNameChanged {
		exists, err := s.ExistsTagName(c, tag.Uid, tag.Name)

		if err != nil {
			return err
		} else if exists {
			return errs.ErrTransactionTagNameAlreadyExists
		}
	}

	tag.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(tag.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(tag.TagId).Cols("name", "tag_group_id", "display_order", "updated_unix_time").Where("uid=? AND deleted=?", tag.Uid, false).Update(tag)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrTransactionTagNotFound
		}

		return err
	})
}

// HideTag updates hidden field of given transaction tags
func (s *TransactionTagService) HideTag(c core.Context, uid int64, ids []int64, hidden bool) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.TransactionTag{
		Hidden:          hidden,
		UpdatedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.Cols("hidden", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).In("tag_id", ids).Update(updateModel)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrTransactionTagNotFound
		}

		return err
	})
}

// ModifyTagDisplayOrders updates display order of given transaction tags
func (s *TransactionTagService) ModifyTagDisplayOrders(c core.Context, uid int64, tags []*models.TransactionTag) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	for i := 0; i < len(tags); i++ {
		tags[i].UpdatedUnixTime = time.Now().Unix()
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(tags); i++ {
			tag := tags[i]
			updatedRows, err := sess.ID(tag.TagId).Cols("display_order", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(tag)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				return errs.ErrTransactionTagNotFound
			}
		}

		return nil
	})
}

// DeleteTag deletes an existed transaction tag from database
func (s *TransactionTagService) DeleteTag(c core.Context, uid int64, tagId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.TransactionTag{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		exists, err := sess.Cols("uid", "tag_id").Where("uid=? AND deleted=? AND tag_id=?", uid, false, tagId).Limit(1).Exist(&models.TransactionTagIndex{})

		if err != nil {
			return err
		} else if exists {
			return errs.ErrTransactionTagInUseCannotBeDeleted
		}

		var relatedTransactionTemplatesByTag []*models.TransactionTemplate
		err = sess.Cols("uid", "deleted", "tag_ids", "template_type", "scheduled_frequency_type", "scheduled_end_time").Where("uid=? AND deleted=? AND (template_type=? OR (template_type=? AND scheduled_frequency_type<>? AND (scheduled_end_time IS NULL OR scheduled_end_time>=?))) AND tag_ids LIKE ?", uid, false, models.TRANSACTION_TEMPLATE_TYPE_NORMAL, models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE, models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED, now, "%%"+utils.Int64ToString(tagId)+"%%").Find(&relatedTransactionTemplatesByTag)

		if err != nil {
			return err
		}

		for i := 0; i < len(relatedTransactionTemplatesByTag); i++ {
			template := relatedTransactionTemplatesByTag[i]
			tagIds, err := s.GetTagIds(template.TagIds)

			if err != nil {
				return err
			}

			for j := 0; j < len(tagIds); j++ {
				if tagIds[j] == tagId {
					return errs.ErrTransactionTagInUseCannotBeDeleted
				}
			}
		}

		deletedRows, err := sess.ID(tagId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrTransactionTagNotFound
		}

		return err
	})
}

// DeleteAllTags deletes all existed transaction tags from database
func (s *TransactionTagService) DeleteAllTags(c core.Context, uid int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.TransactionTag{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		exists, err := sess.Cols("uid", "deleted").Where("uid=? AND deleted=?", uid, false).Limit(1).Exist(&models.TransactionTagIndex{})

		if err != nil {
			return err
		} else if exists {
			return errs.ErrTransactionTagInUseCannotBeDeleted
		}

		_, err = sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		}

		return nil
	})
}
