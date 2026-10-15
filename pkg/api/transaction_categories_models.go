package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"sort"
)

func (a *TransactionCategoriesApi) createBatchCategories(c *core.WebContext, uid int64, categoryCreateBatchReq *models.TransactionCategoryCreateBatchRequest) ([]*models.TransactionCategory, error) {
	var err error
	categoryTypeMaxOrderMap := make(map[models.TransactionCategoryType]int32)
	categories := make([]*models.TransactionCategory, 0)
	categoriesMap := make(map[*models.TransactionCategory][]*models.TransactionCategory)
	categoriesMap[nil] = make([]*models.TransactionCategory, len(categoryCreateBatchReq.Categories))
	totalCount := 0

	for i := 0; i < len(categoryCreateBatchReq.Categories); i++ {
		categoryCreateReq := categoryCreateBatchReq.Categories[i]
		var maxOrderId, exists = categoryTypeMaxOrderMap[categoryCreateReq.Type]

		if !exists {
			maxOrderId, err = a.categories.GetMaxDisplayOrder(c, uid, categoryCreateReq.Type)

			if err != nil {
				log.Errorf(c, "[transaction_categories.CategoryCreateBatchHandler] failed to get max display order for user \"uid:%d\", because %s", uid, err.Error())
				return nil, errs.Or(err, errs.ErrOperationFailed)
			}
		}

		category := a.createNewCategoryModel(uid, &models.TransactionCategoryCreateRequest{
			Name:     categoryCreateReq.Name,
			Type:     categoryCreateReq.Type,
			Icon:     categoryCreateReq.Icon,
			IconType: categoryCreateReq.IconType,
			Color:    categoryCreateReq.Color,
		}, maxOrderId+1)

		categories = append(categories, category)
		categoriesMap[category] = make([]*models.TransactionCategory, len(categoryCreateReq.SubCategories))

		for j := int32(0); j < int32(len(categoryCreateReq.SubCategories)); j++ {
			subCategory := a.createNewCategoryModel(uid, categoryCreateReq.SubCategories[j], j+1)

			categories = append(categories, subCategory)
			categoriesMap[category][j] = subCategory
			totalCount++
		}

		categoriesMap[nil][i] = category
		categoryTypeMaxOrderMap[categoryCreateReq.Type] = maxOrderId + 1
		totalCount++
	}

	iconTypeValid := a.isTransactionCategoryIconTypeValid(c, uid, categories)

	if !iconTypeValid {
		log.Warnf(c, "[transaction_categories.createBatchCategories] icon type invalid for user \"uid:%d\"", uid)
		return nil, errs.ErrTransactionCategoryIconInvalid
	}

	categories, err = a.categories.CreateCategories(c, uid, categoriesMap)

	if err != nil {
		log.Errorf(c, "[transaction_categories.createBatchCategories] failed to create categories for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_categories.createBatchCategories] user \"uid:%d\" has created categories successfully", uid)

	return categories, nil
}

func (a *TransactionCategoriesApi) createNewCategoryModel(uid int64, categoryCreateReq *models.TransactionCategoryCreateRequest, order int32) *models.TransactionCategory {
	return &models.TransactionCategory{
		Uid:              uid,
		Name:             categoryCreateReq.Name,
		Type:             categoryCreateReq.Type,
		ParentCategoryId: categoryCreateReq.ParentId,
		DisplayOrder:     order,
		Icon:             categoryCreateReq.Icon,
		IconType:         categoryCreateReq.IconType,
		Color:            categoryCreateReq.Color,
		Comment:          categoryCreateReq.Comment,
	}
}

func (a *TransactionCategoriesApi) getTransactionCategoryListByTypeResponse(categories []*models.TransactionCategory, parentId int64) (map[models.TransactionCategoryType]models.TransactionCategoryInfoResponseSlice, *errs.Error) {
	categoryResps := make([]*models.TransactionCategoryInfoResponse, len(categories))
	categoryRespMap := make(map[int64]*models.TransactionCategoryInfoResponse)

	for i := 0; i < len(categories); i++ {
		categoryResps[i] = categories[i].ToTransactionCategoryInfoResponse()
		categoryRespMap[categoryResps[i].Id] = categoryResps[i]
	}

	for i := 0; i < len(categoryResps); i++ {
		categoryResp := categoryResps[i]

		if categoryResp.ParentId <= models.LevelOneTransactionCategoryParentId {
			continue
		}

		parentCategory, parentExists := categoryRespMap[categoryResp.ParentId]

		if !parentExists || parentCategory == nil {
			continue
		}

		parentCategory.SubCategories = append(parentCategory.SubCategories, categoryResp)
	}

	finalCategoryResps := make(models.TransactionCategoryInfoResponseSlice, 0)

	for i := 0; i < len(categoryResps); i++ {
		if parentId <= 0 && categoryResps[i].ParentId == models.LevelOneTransactionCategoryParentId {
			sort.Sort(categoryResps[i].SubCategories)
			finalCategoryResps = append(finalCategoryResps, categoryResps[i])
		} else if parentId > 0 && categoryResps[i].ParentId == parentId {
			finalCategoryResps = append(finalCategoryResps, categoryResps[i])
		}
	}

	sort.Sort(finalCategoryResps)

	typeCategoryMapResponse := make(map[models.TransactionCategoryType]models.TransactionCategoryInfoResponseSlice)

	for i := 0; i < len(finalCategoryResps); i++ {
		category := finalCategoryResps[i]
		categoryList, _ := typeCategoryMapResponse[category.Type]

		categoryList = append(categoryList, category)
		typeCategoryMapResponse[category.Type] = categoryList
	}

	return typeCategoryMapResponse, nil
}

func (a *TransactionCategoriesApi) isTransactionCategoryIconTypeValid(c *core.WebContext, uid int64, categories []*models.TransactionCategory) bool {
	iconIds := make([]int64, 0)

	for _, category := range categories {
		if !category.IconType.IsValid() {
			return false
		}

		if category.IconType == core.ICON_TYPE_USER_CUSTOM {
			iconIds = append(iconIds, category.Icon)
		}
	}

	if len(iconIds) < 1 {
		return true
	}

	iconExists, err := a.icons.ExistsCustomIcons(c, uid, iconIds)

	if err != nil {
		log.Errorf(c, "[category.isTransactionCategoryIconTypeValid] failed to check custom icons for user \"uid:%d\", because %s", uid, err.Error())
		return false
	}

	return iconExists
}
