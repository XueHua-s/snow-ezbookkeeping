package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"strings"
)

// GetCategoryMapByList returns a transaction category map by a list
func (s *TransactionCategoryService) GetCategoryMapByList(categories []*models.TransactionCategory) map[int64]*models.TransactionCategory {
	categoryMap := make(map[int64]*models.TransactionCategory)

	for i := 0; i < len(categories); i++ {
		category := categories[i]
		categoryMap[category.CategoryId] = category
	}
	return categoryMap
}

// GetVisibleSubCategoryNameMapByList returns visible sub transaction category map by a list
func (s *TransactionCategoryService) GetVisibleSubCategoryNameMapByList(categories []*models.TransactionCategory) (expenseCategoryMap map[string]map[string]*models.TransactionCategory, incomeCategoryMap map[string]map[string]*models.TransactionCategory, transferCategoryMap map[string]map[string]*models.TransactionCategory) {
	categoryMap := make(map[int64]*models.TransactionCategory, len(categories))
	expenseCategoryMap = make(map[string]map[string]*models.TransactionCategory)
	incomeCategoryMap = make(map[string]map[string]*models.TransactionCategory)
	transferCategoryMap = make(map[string]map[string]*models.TransactionCategory)

	for i := 0; i < len(categories); i++ {
		category := categories[i]
		categoryMap[category.CategoryId] = category
	}

	for i := 0; i < len(categories); i++ {
		category := categories[i]

		if category.Hidden {
			continue
		}

		if category.ParentCategoryId == models.LevelOneTransactionCategoryParentId {
			continue
		}

		parentCategory, exists := categoryMap[category.ParentCategoryId]

		if !exists {
			continue
		}

		var categories map[string]*models.TransactionCategory

		if category.Type == models.CATEGORY_TYPE_INCOME {
			categories, exists = incomeCategoryMap[category.Name]

			if !exists {
				categories = make(map[string]*models.TransactionCategory)
				incomeCategoryMap[category.Name] = categories
			}
		} else if category.Type == models.CATEGORY_TYPE_EXPENSE {
			categories, exists = expenseCategoryMap[category.Name]

			if !exists {
				categories = make(map[string]*models.TransactionCategory)
				expenseCategoryMap[category.Name] = categories
			}
		} else if category.Type == models.CATEGORY_TYPE_TRANSFER {
			categories, exists = transferCategoryMap[category.Name]

			if !exists {
				categories = make(map[string]*models.TransactionCategory)
				transferCategoryMap[category.Name] = categories
			}
		} else {
			continue
		}

		categories[parentCategory.Name] = category
	}

	return expenseCategoryMap, incomeCategoryMap, transferCategoryMap
}

// GetCategoryNames returns a list with transaction category names from transaction category models list
func (s *TransactionCategoryService) GetCategoryNames(categories []*models.TransactionCategory) []string {
	categoryNames := make([]string, len(categories))

	for i := 0; i < len(categories); i++ {
		categoryNames[i] = categories[i].Name
	}

	return categoryNames
}

// GetCategoryOrSubCategoryIds returns all category ids and sub-category ids according to given category ids
func (s *TransactionCategoryService) GetCategoryOrSubCategoryIds(c core.Context, categoryIds string, uid int64) ([]int64, error) {
	if categoryIds == "" || categoryIds == "0" {
		return nil, nil
	}

	requestCategoryIds, err := utils.StringArrayToInt64Array(strings.Split(categoryIds, ","))

	if err != nil {
		return nil, errs.Or(err, errs.ErrTransactionCategoryIdInvalid)
	}

	var allCategoryIds []int64

	if len(requestCategoryIds) > 0 {
		allSubCategories, err := s.GetSubCategoriesByCategoryIds(c, uid, requestCategoryIds)

		if err != nil {
			return nil, err
		}

		categoryIdsMap := make(map[int64]int32, len(requestCategoryIds))

		for i := 0; i < len(requestCategoryIds); i++ {
			categoryIdsMap[requestCategoryIds[i]] = 0
		}

		for i := 0; i < len(allSubCategories); i++ {
			subCategory := allSubCategories[i]

			if refCount, exists := categoryIdsMap[subCategory.ParentCategoryId]; exists {
				categoryIdsMap[subCategory.ParentCategoryId] = refCount + 1
			} else {
				categoryIdsMap[subCategory.ParentCategoryId] = 1
			}

			if _, exists := categoryIdsMap[subCategory.CategoryId]; exists {
				delete(categoryIdsMap, subCategory.CategoryId)
			}

			allCategoryIds = append(allCategoryIds, subCategory.CategoryId)
		}

		for accountId, refCount := range categoryIdsMap {
			if refCount < 1 {
				allCategoryIds = append(allCategoryIds, accountId)
			}
		}
	}

	return allCategoryIds, nil
}

// GetCategoryOrSubCategoryIdsByCategoryName returns a list of transaction category ids or sub-category ids according to given category name
func (s *TransactionCategoryService) GetCategoryOrSubCategoryIdsByCategoryName(categories []*models.TransactionCategory, categoryName string) []int64 {
	categoryIds := make([]int64, 0)
	parentCategoryIds := make([]int64, 0)
	childCategoryByParentCategoryId := make(map[int64][]*models.TransactionCategory)

	for i := 0; i < len(categories); i++ {
		category := categories[i]

		if category.Name == categoryName {
			if category.ParentCategoryId != models.LevelOneTransactionCategoryParentId {
				categoryIds = append(categoryIds, category.CategoryId)
			} else if category.ParentCategoryId == models.LevelOneTransactionCategoryParentId {
				parentCategoryIds = append(parentCategoryIds, category.CategoryId)
			}
		} else if category.ParentCategoryId != models.LevelOneTransactionCategoryParentId {
			childCategories, exists := childCategoryByParentCategoryId[category.ParentCategoryId]

			if !exists {
				childCategories = make([]*models.TransactionCategory, 0)
			}

			childCategories = append(childCategories, category)
			childCategoryByParentCategoryId[category.ParentCategoryId] = childCategories
		}
	}

	for i := 0; i < len(parentCategoryIds); i++ {
		parentCategoryId := parentCategoryIds[i]

		if childCategories, exists := childCategoryByParentCategoryId[parentCategoryId]; exists {
			for j := 0; j < len(childCategories); j++ {
				childCategory := childCategories[j]
				categoryIds = append(categoryIds, childCategory.CategoryId)
			}
		}
	}

	return categoryIds
}
