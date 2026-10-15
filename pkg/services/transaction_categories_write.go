package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
	"time"
	"xorm.io/xorm"
)

// CreateCategory saves a new transaction category model to database
func (s *TransactionCategoryService) CreateCategory(c core.Context, category *models.TransactionCategory) error {
	if category.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	category.CategoryId = s.GenerateUuid(uuid.UUID_TYPE_CATEGORY)

	if category.CategoryId < 1 {
		return errs.ErrSystemIsBusy
	}

	category.Deleted = false
	category.CreatedUnixTime = time.Now().Unix()
	category.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(category.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(category)
		return err
	})
}

// CreateCategories saves a few transaction category models to database
func (s *TransactionCategoryService) CreateCategories(c core.Context, uid int64, categories map[*models.TransactionCategory][]*models.TransactionCategory) ([]*models.TransactionCategory, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var allCategories []*models.TransactionCategory
	primaryCategories := categories[nil]

	needPrimaryCategoryUuidCount := uint16(len(primaryCategories))
	primaryCategoryUuids := s.GenerateUuids(uuid.UUID_TYPE_CATEGORY, needPrimaryCategoryUuidCount)

	if len(primaryCategoryUuids) < int(needPrimaryCategoryUuidCount) {
		return nil, errs.ErrSystemIsBusy
	}

	for i := 0; i < len(primaryCategories); i++ {
		primaryCategory := primaryCategories[i]
		primaryCategory.CategoryId = primaryCategoryUuids[i]
		primaryCategory.Deleted = false
		primaryCategory.CreatedUnixTime = time.Now().Unix()
		primaryCategory.UpdatedUnixTime = time.Now().Unix()

		allCategories = append(allCategories, primaryCategory)

		secondaryCategories := categories[primaryCategory]

		needSecondaryCategoryUuidCount := uint16(len(secondaryCategories))
		secondaryCategoryUuids := s.GenerateUuids(uuid.UUID_TYPE_CATEGORY, needSecondaryCategoryUuidCount)

		if len(secondaryCategoryUuids) < int(needSecondaryCategoryUuidCount) {
			return nil, errs.ErrSystemIsBusy
		}

		for j := 0; j < len(secondaryCategories); j++ {
			secondaryCategory := secondaryCategories[j]
			secondaryCategory.CategoryId = secondaryCategoryUuids[j]
			secondaryCategory.ParentCategoryId = primaryCategory.CategoryId
			secondaryCategory.Deleted = false
			secondaryCategory.CreatedUnixTime = time.Now().Unix()
			secondaryCategory.UpdatedUnixTime = time.Now().Unix()

			allCategories = append(allCategories, secondaryCategory)
		}
	}

	err := s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(allCategories); i++ {
			category := allCategories[i]
			_, err := sess.Insert(category)

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return allCategories, nil
}

// ModifyCategory saves an existed transaction category model to database
func (s *TransactionCategoryService) ModifyCategory(c core.Context, category *models.TransactionCategory) error {
	if category.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	category.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(category.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(category.CategoryId).Cols("parent_category_id", "name", "display_order", "icon", "icon_type", "color", "comment", "hidden", "updated_unix_time").Where("uid=? AND deleted=?", category.Uid, false).Update(category)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrTransactionCategoryNotFound
		}

		return nil
	})
}

// HideCategory updates hidden field of given transaction categories
func (s *TransactionCategoryService) HideCategory(c core.Context, uid int64, ids []int64, hidden bool) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.TransactionCategory{
		Hidden:          hidden,
		UpdatedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.Cols("hidden", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).In("category_id", ids).Update(updateModel)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrTransactionCategoryNotFound
		}

		return nil
	})
}

// ModifyCategoryDisplayOrders updates display order of given transaction categories
func (s *TransactionCategoryService) ModifyCategoryDisplayOrders(c core.Context, uid int64, categories []*models.TransactionCategory) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	for i := 0; i < len(categories); i++ {
		categories[i].UpdatedUnixTime = time.Now().Unix()
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(categories); i++ {
			category := categories[i]
			updatedRows, err := sess.ID(category.CategoryId).Cols("display_order", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(category)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				return errs.ErrTransactionCategoryNotFound
			}
		}

		return nil
	})
}

// DeleteCategory deletes an existed transaction category from database
func (s *TransactionCategoryService) DeleteCategory(c core.Context, uid int64, categoryId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.TransactionCategory{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		var categoryAndSubCategories []*models.TransactionCategory
		err := sess.Where("uid=? AND deleted=? AND (category_id=? OR parent_category_id=?)", uid, false, categoryId, categoryId).Find(&categoryAndSubCategories)

		if err != nil {
			return err
		} else if len(categoryAndSubCategories) < 1 {
			return errs.ErrTransactionCategoryNotFound
		}

		categoryAndSubCategoryIds := make([]int64, len(categoryAndSubCategories))

		for i := 0; i < len(categoryAndSubCategories); i++ {
			categoryAndSubCategoryIds[i] = categoryAndSubCategories[i].CategoryId
		}

		exists, err := sess.Cols("uid", "deleted", "category_id").Where("uid=? AND deleted=?", uid, false).In("category_id", categoryAndSubCategoryIds).Limit(1).Exist(&models.Transaction{})

		if err != nil {
			return err
		} else if exists {
			return errs.ErrTransactionCategoryInUseCannotBeDeleted
		}

		exists, err = sess.Cols("uid", "deleted", "category_id", "template_type", "scheduled_frequency_type", "scheduled_end_time").Where("uid=? AND deleted=? AND (template_type=? OR (template_type=? AND scheduled_frequency_type<>? AND (scheduled_end_time IS NULL OR scheduled_end_time>=?)))", uid, false, models.TRANSACTION_TEMPLATE_TYPE_NORMAL, models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE, models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED, now).In("category_id", categoryAndSubCategoryIds).Limit(1).Exist(&models.TransactionTemplate{})

		if err != nil {
			return err
		} else if exists {
			return errs.ErrTransactionCategoryInUseCannotBeDeleted
		}

		deletedRows, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).In("category_id", categoryAndSubCategoryIds).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrTransactionCategoryNotFound
		}

		return err
	})
}

// DeleteAllCategories deletes all existed transaction categories from database
func (s *TransactionCategoryService) DeleteAllCategories(c core.Context, uid int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.TransactionCategory{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		exists, err := sess.Cols("uid", "deleted", "category_id").Where("uid=? AND deleted=? AND category_id<>?", uid, false, 0).Limit(1).Exist(&models.Transaction{})

		if err != nil {
			return err
		} else if exists {
			return errs.ErrTransactionCategoryInUseCannotBeDeleted
		}

		_, err = sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		}

		return nil
	})
}
