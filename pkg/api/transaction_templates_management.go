package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// TemplateHideHandler hides a transaction template by request parameters for current user
func (a *TransactionTemplatesApi) TemplateHideHandler(c *core.WebContext) (any, *errs.Error) {
	var templateHideReq models.TransactionTemplateHideRequest
	err := c.ShouldBindJSON(&templateHideReq)

	if err != nil {
		log.Warnf(c, "[transaction_templates.TemplateHideHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	template, err := a.templates.GetTemplateByTemplateId(c, uid, templateHideReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateHideHandler] failed to get template \"id:%d\" for user \"uid:%d\", because %s", templateHideReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE && !a.CurrentConfig().EnableScheduledTransaction {
		return nil, errs.ErrScheduledTransactionNotEnabled
	}

	err = a.templates.HideTemplate(c, uid, []int64{templateHideReq.Id}, templateHideReq.Hidden)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateHideHandler] failed to hide template \"id:%d\" for user \"uid:%d\", because %s", templateHideReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_templates.TemplateHideHandler] user \"uid:%d\" has hidden template \"id:%d\"", uid, templateHideReq.Id)
	return true, nil
}

// TemplateMoveHandler moves display order of existed transaction templates by request parameters for current user
func (a *TransactionTemplatesApi) TemplateMoveHandler(c *core.WebContext) (any, *errs.Error) {
	var templateMoveReq models.TransactionTemplateMoveRequest
	err := c.ShouldBindJSON(&templateMoveReq)

	if err != nil {
		log.Warnf(c, "[transaction_templates.TemplateMoveHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	if len(templateMoveReq.NewDisplayOrders) > 0 {
		template, err := a.templates.GetTemplateByTemplateId(c, uid, templateMoveReq.NewDisplayOrders[0].Id)

		if err != nil {
			log.Errorf(c, "[transaction_templates.TemplateMoveHandler] failed to get template \"id:%d\" for user \"uid:%d\", because %s", templateMoveReq.NewDisplayOrders[0].Id, uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE && !a.CurrentConfig().EnableScheduledTransaction {
			return nil, errs.ErrScheduledTransactionNotEnabled
		}
	}

	templates := make([]*models.TransactionTemplate, len(templateMoveReq.NewDisplayOrders))

	for i := 0; i < len(templateMoveReq.NewDisplayOrders); i++ {
		newDisplayOrder := templateMoveReq.NewDisplayOrders[i]
		template := &models.TransactionTemplate{
			Uid:          uid,
			TemplateId:   newDisplayOrder.Id,
			DisplayOrder: newDisplayOrder.DisplayOrder,
		}

		templates[i] = template
	}

	err = a.templates.ModifyTemplateDisplayOrders(c, uid, templates)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateMoveHandler] failed to move templates for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_templates.TemplateMoveHandler] user \"uid:%d\" has moved templates", uid)
	return true, nil
}

// TemplateDeleteHandler deletes an existed transaction template by request parameters for current user
func (a *TransactionTemplatesApi) TemplateDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var templateDeleteReq models.TransactionTemplateDeleteRequest
	err := c.ShouldBindJSON(&templateDeleteReq)

	if err != nil {
		log.Warnf(c, "[transaction_templates.TemplateDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	template, err := a.templates.GetTemplateByTemplateId(c, uid, templateDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateDeleteHandler] failed to get template \"id:%d\" for user \"uid:%d\", because %s", templateDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE && !a.CurrentConfig().EnableScheduledTransaction {
		return nil, errs.ErrScheduledTransactionNotEnabled
	}

	err = a.templates.DeleteTemplate(c, uid, templateDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateDeleteHandler] failed to delete template \"id:%d\" for user \"uid:%d\", because %s", templateDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_templates.TemplateDeleteHandler] user \"uid:%d\" has deleted template \"id:%d\"", uid, templateDeleteReq.Id)
	return true, nil
}
