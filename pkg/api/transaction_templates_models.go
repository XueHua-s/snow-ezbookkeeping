package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"sort"
	"strings"
	"time"
)

func (a *TransactionTemplatesApi) createNewTemplateModel(c *core.WebContext, uid int64, templateCreateReq *models.TransactionTemplateCreateRequest, order int32) (*models.TransactionTemplate, error) {
	template := &models.TransactionTemplate{
		Uid:                  uid,
		TemplateType:         templateCreateReq.TemplateType,
		Name:                 templateCreateReq.Name,
		Type:                 templateCreateReq.Type,
		CategoryId:           templateCreateReq.CategoryId,
		AccountId:            templateCreateReq.SourceAccountId,
		TagIds:               strings.Join(templateCreateReq.TagIds, ","),
		Amount:               templateCreateReq.SourceAmount,
		RelatedAccountId:     templateCreateReq.DestinationAccountId,
		RelatedAccountAmount: templateCreateReq.DestinationAmount,
		HideAmount:           templateCreateReq.HideAmount,
		Comment:              templateCreateReq.Comment,
		DisplayOrder:         order,
	}

	if templateCreateReq.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
		template.ScheduledFrequencyType = *templateCreateReq.ScheduledFrequencyType
		template.ScheduledFrequency = a.getOrderedFrequencyValues(*templateCreateReq.ScheduledFrequency)
		template.ScheduledTimezoneName = *templateCreateReq.ScheduledTimezoneName
		template.ScheduledTimezoneUtcOffset = *templateCreateReq.ScheduledTimezoneUtcOffset

		location, err := time.LoadLocation(template.ScheduledTimezoneName)

		if err != nil {
			log.Warnf(c, "[transaction_templates.createNewTemplateModel] failed to load timezone location \"%s\" for user \"uid:%d\", because %s", template.ScheduledTimezoneName, uid, err.Error())
			return nil, errs.ErrTransactionTimeZoneInvalid
		}

		template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), location)

		if templateCreateReq.ScheduledStartDate != nil {
			startTime, err := utils.ParseFromLongDateFirstTimeInTimezone(*templateCreateReq.ScheduledStartDate, location)

			if err != nil {
				return nil, err
			}

			startUnixTime := startTime.Unix()
			template.ScheduledStartTime = &startUnixTime
		}

		if templateCreateReq.ScheduledEndDate != nil {
			endTime, err := utils.ParseFromLongDateLastTimeInTimezone(*templateCreateReq.ScheduledEndDate, location)

			if err != nil {
				return nil, err
			}

			endUnixTime := endTime.Unix()
			template.ScheduledEndTime = &endUnixTime
		}

		if template.ScheduledStartTime != nil && template.ScheduledEndTime != nil && *template.ScheduledStartTime > *template.ScheduledEndTime {
			return nil, errs.ErrScheduledTransactionTemplateStartDataLaterThanEndDate
		}
	}

	return template, nil
}

func (a *TransactionTemplatesApi) getOrderedFrequencyValues(frequencyValue string) string {
	if frequencyValue == "" {
		return ""
	}

	items := strings.Split(frequencyValue, ",")
	values := make([]int, 0, len(items))
	valueExistMap := make(map[int]bool)

	for i := 0; i < len(items); i++ {
		value, err := utils.StringToInt(items[i])

		if err != nil {
			continue
		}

		if _, exists := valueExistMap[value]; !exists {
			values = append(values, value)
			valueExistMap[value] = true
		}
	}

	sort.Ints(values)

	var sortedFrequencyValueBuilder strings.Builder

	for i := 0; i < len(values); i++ {
		if sortedFrequencyValueBuilder.Len() > 0 {
			sortedFrequencyValueBuilder.WriteRune(',')
		}

		sortedFrequencyValueBuilder.WriteString(utils.IntToString(values[i]))
	}

	return sortedFrequencyValueBuilder.String()
}
