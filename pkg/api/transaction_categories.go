package api

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/duplicatechecker"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// TransactionCategoriesApi represents transaction category api
type TransactionCategoriesApi struct {
	ApiUsingConfig
	ApiUsingDuplicateChecker
	categories *services.TransactionCategoryService
	icons      *services.UserCustomIconService
}

// Initialize a transaction category api singleton instance
var (
	TransactionCategories = &TransactionCategoriesApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		ApiUsingDuplicateChecker: ApiUsingDuplicateChecker{
			ApiUsingConfig: ApiUsingConfig{
				container: settings.Container,
			},
			container: duplicatechecker.Container,
		},
		categories: services.TransactionCategories,
		icons:      services.UserCustomIcons,
	}
)

// CategoryListHandler returns transaction category list of current user
func (a *TransactionCategoriesApi) CategoryListHandler(c *core.WebContext) (any, *errs.Error) {
	var categoryListReq models.TransactionCategoryListRequest
	err := c.ShouldBindQuery(&categoryListReq)

	if err != nil {
		log.Warnf(c, "[transaction_categories.CategoryListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	categories, err := a.categories.GetAllCategoriesByUid(c, uid, categoryListReq.Type, categoryListReq.ParentId)

	if err != nil {
		log.Errorf(c, "[transaction_categories.CategoryListHandler] failed to get categories for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return a.getTransactionCategoryListByTypeResponse(categories, categoryListReq.ParentId)
}

// CategoryGetHandler returns one specific transaction category of current user
func (a *TransactionCategoriesApi) CategoryGetHandler(c *core.WebContext) (any, *errs.Error) {
	var categoryGetReq models.TransactionCategoryGetRequest
	err := c.ShouldBindQuery(&categoryGetReq)

	if err != nil {
		log.Warnf(c, "[transaction_categories.CategoryGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	category, err := a.categories.GetCategoryByCategoryId(c, uid, categoryGetReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_categories.CategoryGetHandler] failed to get category \"id:%d\" for user \"uid:%d\", because %s", categoryGetReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	categoryResp := category.ToTransactionCategoryInfoResponse()

	return categoryResp, nil
}

// CategoryCreateHandler saves a new transaction category by request parameters for current user
func (a *TransactionCategoriesApi) CategoryCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var categoryCreateReq models.TransactionCategoryCreateRequest
	err := c.ShouldBindJSON(&categoryCreateReq)

	if err != nil {
		log.Warnf(c, "[transaction_categories.CategoryCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	if categoryCreateReq.Type < models.CATEGORY_TYPE_INCOME || categoryCreateReq.Type > models.CATEGORY_TYPE_TRANSFER {
		log.Warnf(c, "[transaction_categories.CategoryCreateHandler] category type invalid, type is %d", categoryCreateReq.Type)
		return nil, errs.ErrTransactionCategoryTypeInvalid
	}

	uid := c.GetCurrentUid()

	if categoryCreateReq.ParentId > 0 {
		parentCategory, err := a.categories.GetCategoryByCategoryId(c, uid, categoryCreateReq.ParentId)

		if err != nil {
			log.Errorf(c, "[transaction_categories.CategoryCreateHandler] failed to get parent category \"id:%d\" for user \"uid:%d\", because %s", categoryCreateReq.ParentId, uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		if parentCategory == nil {
			log.Warnf(c, "[transaction_categories.CategoryCreateHandler] parent category \"id:%d\" does not exist for user \"uid:%d\"", categoryCreateReq.ParentId, uid)
			return nil, errs.ErrParentTransactionCategoryNotFound
		}

		if parentCategory.ParentCategoryId > 0 {
			log.Warnf(c, "[transaction_categories.CategoryCreateHandler] parent category \"id:%d\" has another parent category \"id:%d\" for user \"uid:%d\"", parentCategory.CategoryId, parentCategory.ParentCategoryId, uid)
			return nil, errs.ErrCannotAddToSecondaryTransactionCategory
		}
	}

	var maxOrderId int32

	if categoryCreateReq.ParentId <= 0 {
		maxOrderId, err = a.categories.GetMaxDisplayOrder(c, uid, categoryCreateReq.Type)
	} else {
		maxOrderId, err = a.categories.GetMaxSubCategoryDisplayOrder(c, uid, categoryCreateReq.Type, categoryCreateReq.ParentId)
	}

	if err != nil {
		log.Errorf(c, "[transaction_categories.CategoryCreateHandler] failed to get max display order for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	category := a.createNewCategoryModel(uid, &categoryCreateReq, maxOrderId+1)
	iconTypeValid := a.isTransactionCategoryIconTypeValid(c, uid, []*models.TransactionCategory{category})

	if !iconTypeValid {
		log.Warnf(c, "[transaction_categories.CategoryCreateHandler] icon type invalid for user \"uid:%d\"", uid)
		return nil, errs.ErrTransactionCategoryIconInvalid
	}

	if a.CurrentConfig().EnableDuplicateSubmissionsCheck && categoryCreateReq.ClientSessionId != "" {
		found, remark := a.GetSubmissionRemark(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_CATEGORY, uid, categoryCreateReq.ClientSessionId)

		if found {
			log.Infof(c, "[transaction_categories.CategoryCreateHandler] another category \"id:%s\" has been created for user \"uid:%d\"", remark, uid)
			categoryId, err := utils.StringToInt64(remark)

			if err == nil {
				category, err = a.categories.GetCategoryByCategoryId(c, uid, categoryId)

				if err != nil {
					log.Errorf(c, "[transaction_categories.CategoryCreateHandler] failed to get existed category \"id:%d\" for user \"uid:%d\", because %s", categoryId, uid, err.Error())
					return nil, errs.Or(err, errs.ErrOperationFailed)
				}

				categoryResp := category.ToTransactionCategoryInfoResponse()

				return categoryResp, nil
			}
		}
	}

	err = a.categories.CreateCategory(c, category)

	if err != nil {
		log.Errorf(c, "[transaction_categories.CategoryCreateHandler] failed to create category \"id:%d\" for user \"uid:%d\", because %s", category.CategoryId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_categories.CategoryCreateHandler] user \"uid:%d\" has created a new category \"id:%d\" successfully", uid, category.CategoryId)

	a.SetSubmissionRemarkIfEnable(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_CATEGORY, uid, categoryCreateReq.ClientSessionId, utils.Int64ToString(category.CategoryId))
	categoryResp := category.ToTransactionCategoryInfoResponse()

	return categoryResp, nil
}

// CategoryCreateBatchHandler saves some new transaction category by request parameters for current user
func (a *TransactionCategoriesApi) CategoryCreateBatchHandler(c *core.WebContext) (any, *errs.Error) {
	var categoryCreateBatchReq models.TransactionCategoryCreateBatchRequest
	err := c.ShouldBindBodyWith(&categoryCreateBatchReq, binding.JSON)

	if err != nil {
		log.Warnf(c, "[transaction_categories.CategoryCreateBatchHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	categories, err := a.createBatchCategories(c, uid, &categoryCreateBatchReq)

	if err != nil {
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return a.getTransactionCategoryListByTypeResponse(categories, 0)
}
