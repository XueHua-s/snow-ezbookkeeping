package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/duplicatechecker"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"sort"
)

const maximumTagsCountOfTemplate = 10

// TransactionTemplatesApi represents transaction template api
type TransactionTemplatesApi struct {
	ApiUsingConfig
	ApiUsingDuplicateChecker
	templates *services.TransactionTemplateService
}

// Initialize a transaction template api singleton instance
var (
	TransactionTemplates = &TransactionTemplatesApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		ApiUsingDuplicateChecker: ApiUsingDuplicateChecker{
			ApiUsingConfig: ApiUsingConfig{
				container: settings.Container,
			},
			container: duplicatechecker.Container,
		},
		templates: services.TransactionTemplates,
	}
)

// TemplateListHandler returns transaction template list of current user
func (a *TransactionTemplatesApi) TemplateListHandler(c *core.WebContext) (any, *errs.Error) {
	var templateListReq models.TransactionTemplateListRequest
	err := c.ShouldBindQuery(&templateListReq)

	if err != nil {
		log.Warnf(c, "[transaction_templates.TemplateListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	if templateListReq.TemplateType < models.TRANSACTION_TEMPLATE_TYPE_NORMAL || templateListReq.TemplateType > models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
		log.Warnf(c, "[transaction_templates.TemplateListHandler] template type invalid, type is %d", templateListReq.TemplateType)
		return nil, errs.ErrTransactionTemplateTypeInvalid
	}

	if templateListReq.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE && !a.CurrentConfig().EnableScheduledTransaction {
		return nil, errs.ErrScheduledTransactionNotEnabled
	}

	uid := c.GetCurrentUid()
	templates, err := a.templates.GetAllTemplatesByUid(c, uid, templateListReq.TemplateType)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateListHandler] failed to get templates for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	templateResps := make(models.TransactionTemplateInfoResponseSlice, len(templates))
	serverUtcOffset := utils.GetServerTimezoneOffsetMinutes()

	for i := 0; i < len(templates); i++ {
		templateResps[i] = templates[i].ToTransactionTemplateInfoResponse(serverUtcOffset)
	}

	sort.Sort(templateResps)

	return templateResps, nil
}

// TemplateGetHandler returns one specific transaction template of current user
func (a *TransactionTemplatesApi) TemplateGetHandler(c *core.WebContext) (any, *errs.Error) {
	var templateGetReq models.TransactionTemplateGetRequest
	err := c.ShouldBindQuery(&templateGetReq)

	if err != nil {
		log.Warnf(c, "[transaction_templates.TemplateGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	template, err := a.templates.GetTemplateByTemplateId(c, uid, templateGetReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateGetHandler] failed to get template \"id:%d\" for user \"uid:%d\", because %s", templateGetReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE && !a.CurrentConfig().EnableScheduledTransaction {
		return nil, errs.ErrScheduledTransactionNotEnabled
	}

	serverUtcOffset := utils.GetServerTimezoneOffsetMinutes()
	templateResp := template.ToTransactionTemplateInfoResponse(serverUtcOffset)

	return templateResp, nil
}

// TemplateCreateHandler saves a new transaction template by request parameters for current user
func (a *TransactionTemplatesApi) TemplateCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var templateCreateReq models.TransactionTemplateCreateRequest
	err := c.ShouldBindJSON(&templateCreateReq)

	if err != nil {
		log.Warnf(c, "[transaction_templates.TemplateCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	if templateCreateReq.TemplateType < models.TRANSACTION_TEMPLATE_TYPE_NORMAL || templateCreateReq.TemplateType > models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
		log.Warnf(c, "[transaction_templates.TemplateCreateHandler] template type invalid, type is %d", templateCreateReq.TemplateType)
		return nil, errs.ErrTransactionTemplateTypeInvalid
	}

	if templateCreateReq.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE && !a.CurrentConfig().EnableScheduledTransaction {
		return nil, errs.ErrScheduledTransactionNotEnabled
	}

	if templateCreateReq.Type <= models.TRANSACTION_TYPE_MODIFY_BALANCE || templateCreateReq.Type > models.TRANSACTION_TYPE_TRANSFER {
		log.Warnf(c, "[transaction_templates.TemplateCreateHandler] transaction type invalid, type is %d", templateCreateReq.Type)
		return nil, errs.ErrTransactionTypeInvalid
	}

	if templateCreateReq.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
		if templateCreateReq.ScheduledFrequencyType == nil ||
			templateCreateReq.ScheduledFrequency == nil {
			return nil, errs.ErrScheduledTransactionFrequencyInvalid
		}

		if *templateCreateReq.ScheduledFrequencyType == models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED && *templateCreateReq.ScheduledFrequency != "" {
			return nil, errs.ErrScheduledTransactionFrequencyInvalid
		} else if *templateCreateReq.ScheduledFrequencyType != models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED && *templateCreateReq.ScheduledFrequency == "" {
			return nil, errs.ErrScheduledTransactionFrequencyInvalid
		}

		if *templateCreateReq.ScheduledFrequencyType == models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_EVERY_N_DAYS {
			frequencyValue, err := utils.StringToInt(*templateCreateReq.ScheduledFrequency)

			if err != nil || frequencyValue <= 0 {
				return nil, errs.ErrScheduledTransactionFrequencyInvalid
			}
		}

		if *templateCreateReq.ScheduledFrequencyType == models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_EVERY_N_DAYS && templateCreateReq.ScheduledStartDate == nil {
			return nil, errs.ErrScheduledTransactionStartDateRequired
		}

		if templateCreateReq.ScheduledTimezoneName == nil ||
			len(*templateCreateReq.ScheduledTimezoneName) < 1 ||
			templateCreateReq.ScheduledTimezoneUtcOffset == nil {
			return nil, errs.ErrTransactionTimeZoneInvalid
		}
	}

	if len(templateCreateReq.TagIds) > maximumTagsCountOfTemplate {
		return nil, errs.ErrTransactionTemplateHasTooManyTags
	}

	uid := c.GetCurrentUid()

	maxOrderId, err := a.templates.GetMaxDisplayOrder(c, uid, templateCreateReq.TemplateType)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateCreateHandler] failed to get max display order for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	serverUtcOffset := utils.GetServerTimezoneOffsetMinutes()
	template, err := a.createNewTemplateModel(c, uid, &templateCreateReq, maxOrderId+1)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateCreateHandler] failed to create new template for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if a.CurrentConfig().EnableDuplicateSubmissionsCheck && templateCreateReq.ClientSessionId != "" {
		found, remark := a.GetSubmissionRemark(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_TEMPLATE, uid, templateCreateReq.ClientSessionId)

		if found {
			log.Infof(c, "[transaction_templates.TemplateCreateHandler] another template \"id:%s\" has been created for user \"uid:%d\"", remark, uid)
			templateId, err := utils.StringToInt64(remark)

			if err == nil {
				template, err = a.templates.GetTemplateByTemplateId(c, uid, templateId)

				if err != nil {
					log.Errorf(c, "[transaction_templates.TemplateCreateHandler] failed to get existed template \"id:%d\" for user \"uid:%d\", because %s", templateId, uid, err.Error())
					return nil, errs.Or(err, errs.ErrOperationFailed)
				}

				templateResp := template.ToTransactionTemplateInfoResponse(serverUtcOffset)

				return templateResp, nil
			}
		}
	}

	err = a.templates.CreateTemplate(c, template)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateCreateHandler] failed to create template \"id:%d\" for user \"uid:%d\", because %s", template.TemplateId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_templates.TemplateCreateHandler] user \"uid:%d\" has created a new template \"id:%d\" successfully", uid, template.TemplateId)

	a.SetSubmissionRemarkIfEnable(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_TEMPLATE, uid, templateCreateReq.ClientSessionId, utils.Int64ToString(template.TemplateId))
	templateResp := template.ToTransactionTemplateInfoResponse(serverUtcOffset)

	return templateResp, nil
}
