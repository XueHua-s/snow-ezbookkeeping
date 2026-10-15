package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"strings"
	"time"
)

// TemplateModifyHandler saves an existed transaction template by request parameters for current user
func (a *TransactionTemplatesApi) TemplateModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var templateModifyReq models.TransactionTemplateModifyRequest
	err := c.ShouldBindJSON(&templateModifyReq)

	if err != nil {
		log.Warnf(c, "[transaction_templates.TemplateModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	if templateModifyReq.Type <= models.TRANSACTION_TYPE_MODIFY_BALANCE || templateModifyReq.Type > models.TRANSACTION_TYPE_TRANSFER {
		log.Warnf(c, "[transaction_templates.TemplateModifyHandler] transaction type invalid, type is %d", templateModifyReq.Type)
		return nil, errs.ErrTransactionTypeInvalid
	}

	uid := c.GetCurrentUid()
	template, err := a.templates.GetTemplateByTemplateId(c, uid, templateModifyReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateModifyHandler] failed to get template \"id:%d\" for user \"uid:%d\", because %s", templateModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE && !a.CurrentConfig().EnableScheduledTransaction {
		return nil, errs.ErrScheduledTransactionNotEnabled
	}

	if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
		if templateModifyReq.ScheduledFrequencyType == nil ||
			templateModifyReq.ScheduledFrequency == nil {
			return nil, errs.ErrScheduledTransactionFrequencyInvalid
		}

		if *templateModifyReq.ScheduledFrequencyType == models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED && *templateModifyReq.ScheduledFrequency != "" {
			return nil, errs.ErrScheduledTransactionFrequencyInvalid
		} else if *templateModifyReq.ScheduledFrequencyType != models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED && *templateModifyReq.ScheduledFrequency == "" {
			return nil, errs.ErrScheduledTransactionFrequencyInvalid
		}

		if *templateModifyReq.ScheduledFrequencyType == models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_EVERY_N_DAYS {
			frequencyValue, err := utils.StringToInt(*templateModifyReq.ScheduledFrequency)

			if err != nil || frequencyValue <= 0 {
				return nil, errs.ErrScheduledTransactionFrequencyInvalid
			}
		}

		if *templateModifyReq.ScheduledFrequencyType == models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_EVERY_N_DAYS && templateModifyReq.ScheduledStartDate == nil {
			return nil, errs.ErrScheduledTransactionStartDateRequired
		}

		if templateModifyReq.ScheduledTimezoneName == nil ||
			len(*templateModifyReq.ScheduledTimezoneName) < 1 ||
			templateModifyReq.ScheduledTimezoneUtcOffset == nil {
			return nil, errs.ErrTransactionTimeZoneInvalid
		}
	}

	if len(templateModifyReq.TagIds) > maximumTagsCountOfTemplate {
		return nil, errs.ErrTransactionTemplateHasTooManyTags
	}

	newTemplate := &models.TransactionTemplate{
		TemplateId:           template.TemplateId,
		Uid:                  uid,
		Name:                 templateModifyReq.Name,
		Type:                 templateModifyReq.Type,
		CategoryId:           templateModifyReq.CategoryId,
		AccountId:            templateModifyReq.SourceAccountId,
		TagIds:               strings.Join(templateModifyReq.TagIds, ","),
		Amount:               templateModifyReq.SourceAmount,
		RelatedAccountId:     templateModifyReq.DestinationAccountId,
		RelatedAccountAmount: templateModifyReq.DestinationAmount,
		HideAmount:           templateModifyReq.HideAmount,
		Comment:              templateModifyReq.Comment,
	}

	if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
		newTemplate.ScheduledFrequencyType = *templateModifyReq.ScheduledFrequencyType
		newTemplate.ScheduledFrequency = a.getOrderedFrequencyValues(*templateModifyReq.ScheduledFrequency)
		newTemplate.ScheduledTimezoneName = *templateModifyReq.ScheduledTimezoneName
		newTemplate.ScheduledTimezoneUtcOffset = *templateModifyReq.ScheduledTimezoneUtcOffset

		location, err := time.LoadLocation(newTemplate.ScheduledTimezoneName)

		if err != nil {
			log.Warnf(c, "[transaction_templates.TemplateModifyHandler] failed to load timezone location \"%s\" for user \"uid:%d\", because %s", newTemplate.ScheduledTimezoneName, uid, err.Error())
			return nil, errs.ErrTransactionTimeZoneInvalid
		}

		newTemplate.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)

		if templateModifyReq.ScheduledStartDate != nil {
			startTime, err := utils.ParseFromLongDateFirstTimeInTimezone(*templateModifyReq.ScheduledStartDate, location)

			if err != nil {
				log.Errorf(c, "[transaction_templates.TemplateModifyHandler] failed to parse scheduled start date for user \"uid:%d\", because %s", uid, err.Error())
				return nil, errs.Or(err, errs.ErrOperationFailed)
			}

			startUnixTime := startTime.Unix()
			newTemplate.ScheduledStartTime = &startUnixTime
		}

		if templateModifyReq.ScheduledEndDate != nil {
			endTime, err := utils.ParseFromLongDateLastTimeInTimezone(*templateModifyReq.ScheduledEndDate, location)

			if err != nil {
				log.Errorf(c, "[transaction_templates.TemplateModifyHandler] failed to parse scheduled end date for user \"uid:%d\", because %s", uid, err.Error())
				return nil, errs.Or(err, errs.ErrOperationFailed)
			}

			endUnixTime := endTime.Unix()
			newTemplate.ScheduledEndTime = &endUnixTime
		}

		if newTemplate.ScheduledStartTime != nil && newTemplate.ScheduledEndTime != nil && *newTemplate.ScheduledStartTime > *newTemplate.ScheduledEndTime {
			return nil, errs.ErrScheduledTransactionTemplateStartDataLaterThanEndDate
		}
	}

	if newTemplate.Name == template.Name &&
		newTemplate.Type == template.Type &&
		newTemplate.CategoryId == template.CategoryId &&
		newTemplate.AccountId == template.AccountId &&
		newTemplate.TagIds == template.TagIds &&
		newTemplate.Amount == template.Amount &&
		newTemplate.RelatedAccountId == template.RelatedAccountId &&
		newTemplate.RelatedAccountAmount == template.RelatedAccountAmount &&
		newTemplate.HideAmount == template.HideAmount &&
		newTemplate.Comment == template.Comment {
		if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_NORMAL {
			return nil, errs.ErrNothingWillBeUpdated
		} else if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
			if newTemplate.ScheduledFrequencyType == template.ScheduledFrequencyType &&
				newTemplate.ScheduledFrequency == template.ScheduledFrequency &&
				newTemplate.ScheduledStartTime == template.ScheduledStartTime &&
				newTemplate.ScheduledEndTime == template.ScheduledEndTime &&
				newTemplate.ScheduledAt == template.ScheduledAt &&
				newTemplate.ScheduledTimezoneName == template.ScheduledTimezoneName &&
				newTemplate.ScheduledTimezoneUtcOffset == template.ScheduledTimezoneUtcOffset {
				return nil, errs.ErrNothingWillBeUpdated
			}
		}
	}

	err = a.templates.ModifyTemplate(c, newTemplate)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateModifyHandler] failed to update template \"id:%d\" for user \"uid:%d\", because %s", templateModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_templates.TemplateModifyHandler] user \"uid:%d\" has updated template \"id:%d\" successfully", uid, templateModifyReq.Id)

	serverUtcOffset := utils.GetServerTimezoneOffsetMinutes()
	newTemplate.TemplateType = template.TemplateType
	newTemplate.DisplayOrder = template.DisplayOrder
	newTemplate.Hidden = template.Hidden
	templateResp := newTemplate.ToTransactionTemplateInfoResponse(serverUtcOffset)

	return templateResp, nil
}
