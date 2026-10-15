package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"time"
)

// CreateScheduledTransactions saves all scheduled transactions that should be created now
func (s *TransactionService) CreateScheduledTransactions(c core.Context, currentUnixTime int64, interval time.Duration) error {
	var allTemplates []*models.TransactionTemplate
	intervalMinute := int(interval / time.Minute)
	currentTime := time.Unix(currentUnixTime, 0)
	currentMinute := (currentTime.Minute() / intervalMinute) * intervalMinute

	startTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), currentTime.Hour(), currentMinute, 0, 0, time.Local)
	startTimeInUTC := startTime.In(time.UTC)

	minutesElapsedOfDayInUtc := startTimeInUTC.Hour()*60 + startTimeInUTC.Minute()
	secondsElapsedOfDayInUtc := minutesElapsedOfDayInUtc * 60
	todayFirstTimeInUTC := startTimeInUTC.Add(time.Duration(-secondsElapsedOfDayInUtc) * time.Second)
	todayFirstUnixTimeInUTC := todayFirstTimeInUTC.Unix()

	minScheduledAt := minutesElapsedOfDayInUtc
	maxScheduledAt := minScheduledAt + intervalMinute

	for i := 0; i < s.UserDataDBCount(); i++ {
		var templates []*models.TransactionTemplate
		err := s.UserDataDBByIndex(i).NewSession(c).Where("deleted=?"+
			" AND template_type=?"+
			" AND (scheduled_frequency_type=? OR scheduled_frequency_type=? OR scheduled_frequency_type=? OR scheduled_frequency_type=? OR scheduled_frequency_type=?)"+
			" AND (scheduled_start_time IS NULL OR scheduled_start_time<=?)"+
			" AND (scheduled_end_time IS NULL OR scheduled_end_time>=?)"+
			" AND scheduled_at>=?"+
			" AND scheduled_at<?",
			false,
			models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE,
			models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_WEEKLY, models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY, models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY, models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_YEARLY, models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_EVERY_N_DAYS,
			startTime.Unix(),
			startTime.Unix(),
			minScheduledAt,
			maxScheduledAt).Find(&templates)

		if err != nil {
			return err
		}

		allTemplates = append(allTemplates, templates...)
	}

	if len(allTemplates) < 1 {
		return nil
	}

	log.Infof(c, "[transactions.CreateScheduledTransactions] should process %d scheduled transaction templates now (scheduled at from %d to %d)", len(allTemplates), minScheduledAt, maxScheduledAt)

	successCount := 0
	skipCount := 0
	failedCount := 0

	for i := 0; i < len(allTemplates); i++ {
		template := allTemplates[i]
		transaction, tagIds := s.createTransactionByScheduledTransactionTemplate(c, template, startTimeInUTC, todayFirstUnixTimeInUTC)

		if transaction == nil {
			skipCount++
			continue
		}

		err := s.CreateTransaction(c, transaction, tagIds, nil)

		if err == nil {
			successCount++
			log.Infof(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" has created a new trasaction \"id:%d\"", template.TemplateId, transaction.TransactionId)
		} else {
			failedCount++
			log.Errorf(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" failed to create new trasaction, because %s", template.TemplateId, err.Error())
		}
	}

	log.Infof(c, "[transactions.CreateScheduledTransactions] %d transactions has been created successfully, %d templates does not need to create transactions and %d transactions failed to create", successCount, skipCount, failedCount)

	return nil
}
