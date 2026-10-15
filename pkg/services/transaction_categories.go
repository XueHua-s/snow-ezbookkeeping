package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
	"strings"
)

// TransactionCategoryService represents transaction category service
type TransactionCategoryService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a transaction category service singleton instance
var (
	TransactionCategories = &TransactionCategoryService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetTotalCategoryCountByUid returns total category count of user
func (s *TransactionCategoryService) GetTotalCategoryCountByUid(c core.Context, uid int64) (int64, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	count, err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).Count(&models.TransactionCategory{})

	return count, err
}

// GetAllCategoriesByUid returns all transaction category models of user
func (s *TransactionCategoryService) GetAllCategoriesByUid(c core.Context, uid int64, categoryType models.TransactionCategoryType, parentCategoryId int64) ([]*models.TransactionCategory, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	condition := "uid=? AND deleted=?"
	conditionParams := make([]any, 0, 8)
	conditionParams = append(conditionParams, uid)
	conditionParams = append(conditionParams, false)

	if categoryType > 0 {
		condition = condition + " AND type=?"
		conditionParams = append(conditionParams, categoryType)
	}

	if parentCategoryId >= 0 {
		condition = condition + " AND parent_category_id=?"
		conditionParams = append(conditionParams, parentCategoryId)
	}

	var categories []*models.TransactionCategory
	err := s.UserDataDB(uid).NewSession(c).Where(condition, conditionParams...).OrderBy("type asc, parent_category_id asc, display_order asc").Find(&categories)

	return categories, err
}

// GetSubCategoriesByCategoryIds returns sub-category models according to category ids
func (s *TransactionCategoryService) GetSubCategoriesByCategoryIds(c core.Context, uid int64, categoryIds []int64) ([]*models.TransactionCategory, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if len(categoryIds) <= 0 {
		return nil, errs.ErrTransactionCategoryIdInvalid
	}

	condition := "uid=? AND deleted=?"
	conditionParams := make([]any, 0, len(categoryIds)+2)
	conditionParams = append(conditionParams, uid)
	conditionParams = append(conditionParams, false)

	var categoryIdConditions strings.Builder

	for i := 0; i < len(categoryIds); i++ {
		if categoryIds[i] <= 0 {
			return nil, errs.ErrTransactionCategoryIdInvalid
		}

		if categoryIdConditions.Len() > 0 {
			categoryIdConditions.WriteString(",")
		}

		categoryIdConditions.WriteString("?")
		conditionParams = append(conditionParams, categoryIds[i])
	}

	if categoryIdConditions.Len() > 1 {
		condition = condition + " AND parent_category_id IN (" + categoryIdConditions.String() + ")"
	} else {
		condition = condition + " AND parent_category_id = " + categoryIdConditions.String()
	}

	var categories []*models.TransactionCategory
	err := s.UserDataDB(uid).NewSession(c).Where(condition, conditionParams...).OrderBy("display_order asc").Find(&categories)

	return categories, err
}

// GetCategoryByCategoryId returns a transaction category model according to transaction category id
func (s *TransactionCategoryService) GetCategoryByCategoryId(c core.Context, uid int64, categoryId int64) (*models.TransactionCategory, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if categoryId <= 0 {
		return nil, errs.ErrTransactionCategoryIdInvalid
	}

	category := &models.TransactionCategory{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(categoryId).Where("uid=? AND deleted=?", uid, false).Get(category)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrTransactionCategoryNotFound
	}

	return category, nil
}

// GetCategoriesByCategoryIds returns transaction category models according to transaction category ids
func (s *TransactionCategoryService) GetCategoriesByCategoryIds(c core.Context, uid int64, categoryIds []int64) (map[int64]*models.TransactionCategory, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if categoryIds == nil {
		return nil, errs.ErrTransactionCategoryIdInvalid
	}

	var categories []*models.TransactionCategory
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).In("category_id", categoryIds).Find(&categories)

	if err != nil {
		return nil, err
	}

	categoryMap := s.GetCategoryMapByList(categories)
	return categoryMap, err
}

// GetMaxDisplayOrder returns the max display order according to transaction category type
func (s *TransactionCategoryService) GetMaxDisplayOrder(c core.Context, uid int64, categoryType models.TransactionCategoryType) (int32, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	category := &models.TransactionCategory{}
	has, err := s.UserDataDB(uid).NewSession(c).Cols("uid", "deleted", "parent_category_id", "display_order").Where("uid=? AND deleted=? AND type=? AND parent_category_id=?", uid, false, categoryType, models.LevelOneTransactionCategoryParentId).OrderBy("display_order desc").Limit(1).Get(category)

	if err != nil {
		return 0, err
	}

	if has {
		return category.DisplayOrder, nil
	} else {
		return 0, nil
	}
}

// GetMaxSubCategoryDisplayOrder returns the max display order of sub transaction category according to transaction category type and parent transaction category id
func (s *TransactionCategoryService) GetMaxSubCategoryDisplayOrder(c core.Context, uid int64, categoryType models.TransactionCategoryType, parentCategoryId int64) (int32, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	if parentCategoryId <= 0 {
		return 0, errs.ErrTransactionCategoryIdInvalid
	}

	category := &models.TransactionCategory{}
	has, err := s.UserDataDB(uid).NewSession(c).Cols("uid", "deleted", "parent_category_id", "display_order").Where("uid=? AND deleted=? AND type=? AND parent_category_id=?", uid, false, categoryType, parentCategoryId).OrderBy("display_order desc").Limit(1).Get(category)

	if err != nil {
		return 0, err
	}

	if has {
		return category.DisplayOrder, nil
	} else {
		return 0, nil
	}
}
