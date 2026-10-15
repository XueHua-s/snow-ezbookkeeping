package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"strings"
	"time"
)

func (s *TransactionService) createTransactionByScheduledTransactionTemplate(c core.Context, template *models.TransactionTemplate, startTimeInUTC time.Time, todayFirstUnixTimeInUTC int64) (*models.Transaction, []int64) {
	if template.ScheduledFrequencyType == models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED {
		log.Infof(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" disabled scheduled transaction frequency", template.TemplateId)
		return nil, nil
	}

	if (template.ScheduledFrequencyType != models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_WEEKLY &&
		template.ScheduledFrequencyType != models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY &&
		template.ScheduledFrequencyType != models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY &&
		template.ScheduledFrequencyType != models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_YEARLY &&
		template.ScheduledFrequencyType != models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_EVERY_N_DAYS) ||
		template.ScheduledFrequency == "" {
		log.Warnf(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" has invalid scheduled transaction frequency", template.TemplateId)
		return nil, nil
	}

	frequencyValues, err := utils.StringArrayToInt64Array(strings.Split(template.ScheduledFrequency, ","))

	if err != nil {
		log.Warnf(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" has invalid scheduled transaction frequency, because %s", template.TemplateId, err.Error())
		return nil, nil
	}

	templateTimezone := time.FixedZone("Template Timezone", int(template.ScheduledTimezoneUtcOffset)*60)

	if len(template.ScheduledTimezoneName) > 0 {
		location, err := time.LoadLocation(template.ScheduledTimezoneName)

		if err != nil {
			log.Warnf(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" cannot load timezone \"%s\", because %s", template.TemplateId, template.ScheduledTimezoneName, err.Error())
		} else {
			templateTimezone = location
		}
	}

	transactionTime := utils.GetScheduledTransactionFinalTime(todayFirstUnixTimeInUTC, template.ScheduledAt, template.CreatedUnixTime, templateTimezone)

	if transactionTime == nil {
		log.Warnf(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" does not need to create transaction, the scheduled local date does not exist", template.TemplateId)
		return nil, nil
	}

	transactionTimezoneUtcOffset := utils.GetTimezoneOffsetMinutes(transactionTime.Unix(), templateTimezone)

	if template.ScheduledFrequencyType == models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY {
		maxDayInMonth := utils.GetMaxDayOfMonth(transactionTime.Year(), transactionTime.Month())

		for i := 0; i < len(frequencyValues); i++ {
			if frequencyValues[i] < 0 {
				frequencyValues[i] = int64(maxDayInMonth) + frequencyValues[i] + 1
			}
		}
	}

	frequencyValueSet := utils.ToSet(frequencyValues)

	if template.ScheduledFrequencyType == models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_WEEKLY && !frequencyValueSet[int64(transactionTime.Weekday())] {
		log.Infof(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" does not need to create transaction, today is %s", template.TemplateId, startTimeInUTC.Weekday())
		return nil, nil
	} else if template.ScheduledFrequencyType == models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY && !frequencyValueSet[int64(transactionTime.Day())] {
		log.Infof(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" does not need to create transaction, today is %d of month", template.TemplateId, startTimeInUTC.Day())
		return nil, nil
	} else if template.ScheduledFrequencyType == models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_YEARLY && !frequencyValueSet[int64(transactionTime.Month())*100+int64(transactionTime.Day())] {
		log.Infof(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" does not need to create transaction, today is %d-%d of year", template.TemplateId, startTimeInUTC.Month(), startTimeInUTC.Day())
		return nil, nil
	} else if template.ScheduledFrequencyType == models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_EVERY_N_DAYS {
		if template.ScheduledStartTime == nil || len(frequencyValues) != 1 || frequencyValues[0] <= 0 {
			log.Warnf(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" has invalid scheduled transaction frequency for every N days", template.TemplateId)
			return nil, nil
		}

		n := frequencyValues[0]
		startDate := time.Unix(*template.ScheduledStartTime, 0).In(templateTimezone)
		daysDiff := utils.GetDaysBetweenDates(startDate, *transactionTime)

		if daysDiff < 0 || int64(daysDiff)%n != 0 {
			log.Infof(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" does not need to create transaction, days diff is %d with interval %d", template.TemplateId, daysDiff, n)
			return nil, nil
		}
	}

	if template.ScheduledStartTime != nil && *template.ScheduledStartTime > transactionTime.Unix() {
		log.Infof(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" does not need to create transaction, now is earlier than the start time %d", template.TemplateId, *template.ScheduledStartTime)
		return nil, nil
	}

	if template.ScheduledEndTime != nil && *template.ScheduledEndTime < transactionTime.Unix() {
		log.Infof(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" does not need to create transaction, now is later than the end time %d", template.TemplateId, *template.ScheduledEndTime)
		return nil, nil
	}

	var transactionDbType models.TransactionDbType

	if template.Type == models.TRANSACTION_TYPE_EXPENSE {
		transactionDbType = models.TRANSACTION_DB_TYPE_EXPENSE
	} else if template.Type == models.TRANSACTION_TYPE_INCOME {
		transactionDbType = models.TRANSACTION_DB_TYPE_INCOME
	} else if template.Type == models.TRANSACTION_TYPE_TRANSFER {
		transactionDbType = models.TRANSACTION_DB_TYPE_TRANSFER_OUT
	} else {
		log.Warnf(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" has invalid transaction type", template.TemplateId)
		return nil, nil
	}

	transaction := &models.Transaction{
		Uid:               template.Uid,
		Type:              transactionDbType,
		CategoryId:        template.CategoryId,
		TransactionTime:   utils.GetMinTransactionTimeFromUnixTime(transactionTime.Unix()),
		TimezoneUtcOffset: transactionTimezoneUtcOffset,
		AccountId:         template.AccountId,
		Amount:            template.Amount,
		HideAmount:        template.HideAmount,
		Comment:           template.Comment,
		CreatedIp:         c.ClientIP(),
		ScheduledCreated:  true,
	}

	if template.Type == models.TRANSACTION_TYPE_TRANSFER {
		transaction.RelatedAccountId = template.RelatedAccountId
		transaction.RelatedAccountAmount = template.RelatedAccountAmount
	}

	tagIds := template.GetTagIds()
	return transaction, tagIds
}
