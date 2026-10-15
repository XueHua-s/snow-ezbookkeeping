package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// CategoryModifyHandler saves an existed transaction category by request parameters for current user
func (a *TransactionCategoriesApi) CategoryModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var categoryModifyReq models.TransactionCategoryModifyRequest
	err := c.ShouldBindJSON(&categoryModifyReq)

	if err != nil {
		log.Warnf(c, "[transaction_categories.CategoryModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	category, err := a.categories.GetCategoryByCategoryId(c, uid, categoryModifyReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_categories.CategoryModifyHandler] failed to get category \"id:%d\" for user \"uid:%d\", because %s", categoryModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	newCategory := &models.TransactionCategory{
		CategoryId:       category.CategoryId,
		Uid:              uid,
		ParentCategoryId: categoryModifyReq.ParentId,
		Name:             categoryModifyReq.Name,
		DisplayOrder:     category.DisplayOrder,
		Icon:             categoryModifyReq.Icon,
		IconType:         categoryModifyReq.IconType,
		Color:            categoryModifyReq.Color,
		Comment:          categoryModifyReq.Comment,
		Hidden:           categoryModifyReq.Hidden,
	}

	if newCategory.ParentCategoryId == category.ParentCategoryId &&
		newCategory.Name == category.Name &&
		newCategory.Icon == category.Icon &&
		newCategory.IconType == category.IconType &&
		newCategory.Color == category.Color &&
		newCategory.Comment == category.Comment &&
		newCategory.Hidden == category.Hidden {
		return nil, errs.ErrNothingWillBeUpdated
	}

	if category.ParentCategoryId == models.LevelOneTransactionCategoryParentId && newCategory.ParentCategoryId != models.LevelOneTransactionCategoryParentId {
		return nil, errs.Or(err, errs.ErrNotAllowChangePrimaryTransactionCategoryToSecondary)
	}

	if category.ParentCategoryId != models.LevelOneTransactionCategoryParentId && newCategory.ParentCategoryId == models.LevelOneTransactionCategoryParentId {
		return nil, errs.Or(err, errs.ErrNotAllowChangeSecondaryTransactionCategoryToPrimary)
	}

	if newCategory.ParentCategoryId != category.ParentCategoryId {
		fromPrimaryCategory, err := a.categories.GetCategoryByCategoryId(c, uid, category.ParentCategoryId)

		if err != nil {
			log.Errorf(c, "[transaction_categories.CategoryModifyHandler] failed to get old primary category \"id:%d\" of category \"id:%d\" for user \"uid:%d\", because %s", category.ParentCategoryId, categoryModifyReq.Id, uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		toPrimaryCategory, err := a.categories.GetCategoryByCategoryId(c, uid, newCategory.ParentCategoryId)

		if err != nil {
			log.Errorf(c, "[transaction_categories.CategoryModifyHandler] failed to get new primary category \"id:%d\" of category \"id:%d\" for user \"uid:%d\", because %s", newCategory.ParentCategoryId, categoryModifyReq.Id, uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		if fromPrimaryCategory.Type != toPrimaryCategory.Type {
			return nil, errs.Or(err, errs.ErrNotAllowChangePrimaryTransactionType)
		}

		if toPrimaryCategory.ParentCategoryId != models.LevelOneTransactionCategoryParentId {
			return nil, errs.Or(err, errs.ErrNotAllowUseSecondaryTransactionAsPrimaryCategory)
		}

		maxOrderId, err := a.categories.GetMaxSubCategoryDisplayOrder(c, uid, category.Type, newCategory.ParentCategoryId)

		if err != nil {
			log.Errorf(c, "[transaction_categories.CategoryModifyHandler] failed to get max display order for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		newCategory.DisplayOrder = maxOrderId + 1
	}

	iconTypeValid := a.isTransactionCategoryIconTypeValid(c, uid, []*models.TransactionCategory{newCategory})

	if !iconTypeValid {
		log.Warnf(c, "[transaction_categories.CategoryModifyHandler] icon type invalid for user \"uid:%d\"", uid)
		return nil, errs.ErrTransactionCategoryIconInvalid
	}

	err = a.categories.ModifyCategory(c, newCategory)

	if err != nil {
		log.Errorf(c, "[transaction_categories.CategoryModifyHandler] failed to update category \"id:%d\" for user \"uid:%d\", because %s", categoryModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_categories.CategoryModifyHandler] user \"uid:%d\" has updated category \"id:%d\" successfully", uid, categoryModifyReq.Id)

	newCategory.Type = category.Type
	categoryResp := newCategory.ToTransactionCategoryInfoResponse()

	return categoryResp, nil
}

// CategoryHideHandler hides an existed transaction category by request parameters for current user
func (a *TransactionCategoriesApi) CategoryHideHandler(c *core.WebContext) (any, *errs.Error) {
	var categoryHideReq models.TransactionCategoryHideRequest
	err := c.ShouldBindJSON(&categoryHideReq)

	if err != nil {
		log.Warnf(c, "[transaction_categories.CategoryHideHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.categories.HideCategory(c, uid, []int64{categoryHideReq.Id}, categoryHideReq.Hidden)

	if err != nil {
		log.Errorf(c, "[transaction_categories.CategoryHideHandler] failed to hide category \"id:%d\" for user \"uid:%d\", because %s", categoryHideReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_categories.CategoryHideHandler] user \"uid:%d\" has hidden category \"id:%d\"", uid, categoryHideReq.Id)
	return true, nil
}

// CategoryMoveHandler moves display order of existed transaction categories by request parameters for current user
func (a *TransactionCategoriesApi) CategoryMoveHandler(c *core.WebContext) (any, *errs.Error) {
	var categoryMoveReq models.TransactionCategoryMoveRequest
	err := c.ShouldBindJSON(&categoryMoveReq)

	if err != nil {
		log.Warnf(c, "[transaction_categories.CategoryMoveHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	categories := make([]*models.TransactionCategory, len(categoryMoveReq.NewDisplayOrders))

	for i := 0; i < len(categoryMoveReq.NewDisplayOrders); i++ {
		newDisplayOrder := categoryMoveReq.NewDisplayOrders[i]
		category := &models.TransactionCategory{
			Uid:          uid,
			CategoryId:   newDisplayOrder.Id,
			DisplayOrder: newDisplayOrder.DisplayOrder,
		}

		categories[i] = category
	}

	err = a.categories.ModifyCategoryDisplayOrders(c, uid, categories)

	if err != nil {
		log.Errorf(c, "[transaction_categories.CategoryMoveHandler] failed to move categories for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_categories.CategoryMoveHandler] user \"uid:%d\" has moved categories", uid)
	return true, nil
}

// CategoryDeleteHandler deletes an existed transaction category by request parameters for current user
func (a *TransactionCategoriesApi) CategoryDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var categoryDeleteReq models.TransactionCategoryDeleteRequest
	err := c.ShouldBindJSON(&categoryDeleteReq)

	if err != nil {
		log.Warnf(c, "[transaction_categories.CategoryDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.categories.DeleteCategory(c, uid, categoryDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_categories.CategoryDeleteHandler] failed to delete category \"id:%d\" for user \"uid:%d\", because %s", categoryDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_categories.CategoryDeleteHandler] user \"uid:%d\" has deleted category \"id:%d\"", uid, categoryDeleteReq.Id)
	return true, nil
}
