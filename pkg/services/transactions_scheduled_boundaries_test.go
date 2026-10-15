package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreateTransactionByScheduledTransactionTemplate_DateBoundary(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY,
		ScheduledFrequency:     "29",
		ScheduledTimezoneName:  "Asia/Shanghai",
		CreatedUnixTime:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2024, 2, 28, 16, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC).Unix()

	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   startTimeInUTC.Unix() * 1000,
		TimezoneUtcOffset: 480,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	template.ScheduledFrequency = "1"
	startTimeInUTC = time.Date(2024, 2, 29, 16, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2026, 12, 31, 16, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_FutureLeapYear(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:     "1",
		ScheduledTimezoneName:  "Europe/Berlin",
		CreatedUnixTime:        time.Date(2040, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2040, 6, 30, 23, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2040, 6, 30, 0, 0, 0, 0, time.UTC).Unix()

	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   time.Date(2040, 6, 30, 22, 0, 0, 0, time.UTC).Unix() * 1000,
		TimezoneUtcOffset: 120,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	template.ScheduledTimezoneName = "Australia/Sydney"
	location, err = time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC = time.Date(2040, 12, 31, 14, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2040, 12, 31, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = time.Date(2040, 12, 31, 13, 0, 0, 0, time.UTC).Unix() * 1000
	expectedValue.TimezoneUtcOffset = 660
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_ServerTimezone(t *testing.T) {
	originalLocation := time.Local
	t.Cleanup(func() {
		time.Local = originalLocation
	})
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY,
		ScheduledFrequency:     "1",
		ScheduledTimezoneName:  "America/New_York",
		CreatedUnixTime:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2026, 7, 1, 5, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC).Unix()

	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   time.Date(2026, 7, 1, 4, 0, 0, 0, time.UTC).Unix() * 1000,
		TimezoneUtcOffset: -240,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	time.Local = time.UTC
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	time.Local = time.FixedZone("Server UTC-12", -12*60*60)
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	time.Local = time.FixedZone("Server UTC-5", -5*60*60)
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	time.Local = time.FixedZone("Server UTC+1", 60*60)
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	time.Local = time.FixedZone("Server UTC+8", 8*60*60)
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	time.Local = time.FixedZone("Server UTC+14", 14*60*60)
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_TimezoneWithMidnightDST(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:     "1",
		ScheduledTimezoneName:  "America/Santiago",
		CreatedUnixTime:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2026, 9, 5, 4, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC).Unix()

	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   startTimeInUTC.Unix() * 1000,
		TimezoneUtcOffset: -240,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	// September 6 starts at 01:00 UTC-3; midnight must not become September 5.
	startTimeInUTC = time.Date(2026, 9, 6, 4, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = -180
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2026, 9, 7, 4, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = time.Date(2026, 9, 7, 3, 0, 0, 0, time.UTC).Unix() * 1000
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	// Havana's March 8 also starts at 01:00, with UTC-4.
	template.ScheduledTimezoneName = "America/Havana"
	location, err = time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC = time.Date(2026, 3, 8, 5, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = -240
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	// Cairo's April 24 starts at 01:00 UTC+3, within the requested local date.
	template.ScheduledTimezoneName = "Africa/Cairo"
	location, err = time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC = time.Date(2026, 4, 23, 22, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = 180
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_MidnightDSTFrequency(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY,
		ScheduledFrequency:     "6",
		ScheduledTimezoneName:  "America/Santiago",
		CreatedUnixTime:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2026, 9, 6, 4, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC).Unix()

	// A missing midnight must not cause the intended day to be skipped.
	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   startTimeInUTC.Unix() * 1000,
		TimezoneUtcOffset: -180,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, []int64{}, actualTagIds)

	template.ScheduledFrequencyType = models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_WEEKLY
	template.ScheduledFrequency = "0"
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, []int64{}, actualTagIds)

	template.ScheduledFrequencyType = models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_YEARLY
	template.ScheduledFrequency = "906"
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, []int64{}, actualTagIds)

	startUnixTime := time.Date(2026, 9, 4, 4, 0, 0, 0, time.UTC).Unix()
	template.ScheduledFrequencyType = models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_EVERY_N_DAYS
	template.ScheduledFrequency = "2"
	template.ScheduledStartTime = &startUnixTime
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, []int64{}, actualTagIds)
}

func TestCreateTransactionByScheduledTransactionTemplate_MidnightDSTStartAndEndDate(t *testing.T) {
	location, err := time.LoadLocation("America/Santiago")
	assert.Equal(t, nil, err)
	startTime, err := utils.ParseFromLongDateFirstTimeInTimezone("2026-09-06", location)
	assert.Equal(t, nil, err)
	endTime, err := utils.ParseFromLongDateLastTimeInTimezone("2026-09-06", location)
	assert.Equal(t, nil, err)
	startUnixTime := startTime.Unix()
	endUnixTime := endTime.Unix()
	template := &models.TransactionTemplate{
		TemplateType:           models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE,
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_EVERY_N_DAYS,
		ScheduledFrequency:     "1",
		ScheduledStartTime:     &startUnixTime,
		ScheduledEndTime:       &endUnixTime,
		ScheduledTimezoneName:  "America/Santiago",
		CreatedUnixTime:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2026, 9, 6, 4, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC).Unix()

	// A missing midnight is still the first scheduled calendar date.
	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   startTimeInUTC.Unix() * 1000,
		TimezoneUtcOffset: -180,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, []int64{}, actualTagIds)

	response := template.ToTransactionTemplateInfoResponse(0)
	assert.Equal(t, "2026-09-06", *response.ScheduledStartDate)
	assert.Equal(t, "2026-09-06", *response.ScheduledEndDate)

	startTimeInUTC = time.Date(2026, 9, 5, 4, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC).Unix()
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	startTimeInUTC = time.Date(2026, 9, 7, 4, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC).Unix()
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	// Cairo's start at 01:00 must not move the returned end date to April 25.
	location, err = time.LoadLocation("Africa/Cairo")
	assert.Equal(t, nil, err)
	startTime, err = utils.ParseFromLongDateFirstTimeInTimezone("2026-04-24", location)
	assert.Equal(t, nil, err)
	endTime, err = utils.ParseFromLongDateLastTimeInTimezone("2026-04-24", location)
	assert.Equal(t, nil, err)
	startUnixTime = startTime.Unix()
	endUnixTime = endTime.Unix()
	template.ScheduledTimezoneName = "Africa/Cairo"
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC = time.Date(2026, 4, 23, 22, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = 180
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
	response = template.ToTransactionTemplateInfoResponse(0)
	assert.Equal(t, "2026-04-24", *response.ScheduledStartDate)
	assert.Equal(t, "2026-04-24", *response.ScheduledEndDate)
}

func TestCreateTransactionByScheduledTransactionTemplate_AmbiguousMidnight(t *testing.T) {
	location, err := time.LoadLocation("America/Havana")
	assert.Equal(t, nil, err)
	startTime, err := utils.ParseFromLongDateFirstTimeInTimezone("2026-11-01", location)
	assert.Equal(t, nil, err)
	startUnixTime := startTime.Unix()
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY,
		ScheduledFrequency:     "1",
		ScheduledStartTime:     &startUnixTime,
		ScheduledTimezoneName:  "America/Havana",
		CreatedUnixTime:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2026, 11, 1, 5, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC).Unix()

	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   time.Date(2026, 11, 1, 4, 0, 0, 0, time.UTC).Unix() * 1000,
		TimezoneUtcOffset: -240,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, []int64{}, actualTagIds)
}

func TestCreateTransactionByScheduledTransactionTemplate_SkippedDate(t *testing.T) {
	location, err := time.LoadLocation("Pacific/Apia")
	assert.Equal(t, nil, err)
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:     "1",
		ScheduledTimezoneName:  "Pacific/Apia",
		CreatedUnixTime:        time.Date(2011, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2011, 12, 30, 10, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2011, 12, 30, 0, 0, 0, 0, time.UTC).Unix()

	// A skipped date must not create an extra transaction for December 31.
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	startTimeInUTC = time.Date(2011, 12, 31, 10, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2011, 12, 31, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   time.Date(2011, 12, 30, 10, 0, 0, 0, time.UTC).Unix() * 1000,
		TimezoneUtcOffset: 840,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, []int64{}, actualTagIds)
}
