package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreateTransactionByScheduledTransactionTemplate_EveryNDaysInvalidFrequency(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_EVERY_N_DAYS,
		ScheduledFrequency:     "1",
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), time.UTC)
	startTimeInUTC := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)

	// Every N days requires a start date.
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	startUnixTime := startTimeInUTC.Unix()
	template.ScheduledStartTime = &startUnixTime
	template.ScheduledFrequency = "1,2"
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	template.ScheduledFrequency = "0"
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	template.ScheduledFrequency = "-1"
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)
}

func TestCreateTransactionByScheduledTransactionTemplate_EveryNDays(t *testing.T) {
	startUnixTime := time.Date(2026, 11, 28, 0, 0, 0, 0, time.UTC).Unix()
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_EVERY_N_DAYS,
		ScheduledFrequency:     "3",
		ScheduledStartTime:     &startUnixTime,
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), time.UTC)
	startTimeInUTC := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)

	expectedValue := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:  startTimeInUTC.Unix() * 1000,
		CreatedIp:        "127.0.0.1",
		ScheduledCreated: true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2026, 12, 2, 0, 0, 0, 0, time.UTC)
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	// A negative day difference must be skipped, even if it is divisible by N.
	startTimeInUTC = time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	// The start date itself is included.
	startTimeInUTC = time.Date(2026, 11, 28, 0, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_StartAndEndTime(t *testing.T) {
	startTimeInUTC := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	startUnixTime := startTimeInUTC.Unix()
	endUnixTime := startTimeInUTC.Unix()
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:     "1",
		ScheduledStartTime:     &startUnixTime,
		ScheduledEndTime:       &endUnixTime,
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), time.UTC)

	// Both boundaries are inclusive.
	expectedValue := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:  startTimeInUTC.Unix() * 1000,
		CreatedIp:        "127.0.0.1",
		ScheduledCreated: true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)

	startUnixTime++
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	startUnixTime--
	endUnixTime--
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	template.ScheduledStartTime = nil
	endUnixTime++
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)

	template.ScheduledStartTime = &startUnixTime
	template.ScheduledEndTime = nil
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_FixedTimezone(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:     "1",
	}
	location := time.FixedZone("Template Timezone", int(template.ScheduledTimezoneUtcOffset)*60)
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), location)

	todayFirstUnixTimeInUTC := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC).Unix()
	startTimeInUTC := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)

	// Legacy templates without a timezone name or creation time keep their offset.
	expectedValue := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:  startTimeInUTC.Unix() * 1000,
		CreatedIp:        "127.0.0.1",
		ScheduledCreated: true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	template.ScheduledTimezoneUtcOffset = 60
	location = time.FixedZone("Template Timezone", int(template.ScheduledTimezoneUtcOffset)*60)
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), location)

	startTimeInUTC = time.Date(2026, 11, 30, 23, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = 60
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	template.ScheduledTimezoneUtcOffset = 120
	location = time.FixedZone("Template Timezone", int(template.ScheduledTimezoneUtcOffset)*60)
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), location)

	startTimeInUTC = time.Date(2026, 11, 30, 22, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = 120
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	template.ScheduledTimezoneUtcOffset = 180
	location = time.FixedZone("Template Timezone", int(template.ScheduledTimezoneUtcOffset)*60)
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), location)

	startTimeInUTC = time.Date(2026, 11, 30, 21, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = 180
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	template.ScheduledTimezoneUtcOffset = 330
	location = time.FixedZone("Template Timezone", int(template.ScheduledTimezoneUtcOffset)*60)
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), location)

	startTimeInUTC = time.Date(2026, 11, 30, 18, 30, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = 330
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	template.ScheduledTimezoneUtcOffset = 345
	location = time.FixedZone("Template Timezone", int(template.ScheduledTimezoneUtcOffset)*60)
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), location)

	startTimeInUTC = time.Date(2026, 11, 30, 18, 15, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = 345
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	template.ScheduledTimezoneUtcOffset = -300
	location = time.FixedZone("Template Timezone", int(template.ScheduledTimezoneUtcOffset)*60)
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), location)

	startTimeInUTC = time.Date(2026, 11, 30, 5, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = -300
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	template.ScheduledTimezoneUtcOffset = -210
	location = time.FixedZone("Template Timezone", int(template.ScheduledTimezoneUtcOffset)*60)
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), location)

	startTimeInUTC = time.Date(2026, 11, 30, 3, 30, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = -210
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	template.ScheduledTimezoneUtcOffset = -720
	location = time.FixedZone("Template Timezone", int(template.ScheduledTimezoneUtcOffset)*60)
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), location)

	startTimeInUTC = time.Date(2026, 11, 30, 12, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = -720
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	template.ScheduledTimezoneUtcOffset = 840
	location = time.FixedZone("Template Timezone", int(template.ScheduledTimezoneUtcOffset)*60)
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), location)

	startTimeInUTC = time.Date(2026, 11, 30, 10, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = 840
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_InvalidTimezone(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                       models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType:     models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY,
		ScheduledFrequency:         "1",
		ScheduledTimezoneName:      "Invalid/Timezone",
		ScheduledTimezoneUtcOffset: 120,
		CreatedUnixTime:            time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location := time.FixedZone("Template Timezone", int(template.ScheduledTimezoneUtcOffset)*60)
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)

	startTimeInUTC := time.Date(2026, 11, 30, 22, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC).Unix()

	// A timezone that cannot be loaded falls back to the stored fixed offset.
	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   startTimeInUTC.Unix() * 1000,
		TimezoneUtcOffset: 120,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_NamedFixedTimezone(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                       models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType:     models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:         "1",
		ScheduledTimezoneName:      "UTC",
		ScheduledTimezoneUtcOffset: 480,
		CreatedUnixTime:            time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := startTimeInUTC.Unix()

	// The named timezone takes precedence over the stored offset.
	expectedValue := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:  startTimeInUTC.Unix() * 1000,
		CreatedIp:        "127.0.0.1",
		ScheduledCreated: true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	// Shanghai advances to January 1 across the UTC year boundary.
	template.ScheduledTimezoneName = "Asia/Shanghai"
	location, err = time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC = time.Date(2026, 12, 31, 16, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = 480
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	template.ScheduledTimezoneName = "Asia/Kathmandu"
	location, err = time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC = time.Date(2026, 12, 31, 18, 15, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = 345
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_TimezoneWithDST(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                       models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType:     models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:         "1",
		ScheduledTimezoneName:      "Europe/Berlin",
		ScheduledTimezoneUtcOffset: 120,
		CreatedUnixTime:            time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2026, 11, 30, 23, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC).Unix()

	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   startTimeInUTC.Unix() * 1000,
		TimezoneUtcOffset: 60,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2026, 6, 30, 23, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = time.Date(2026, 6, 30, 22, 0, 0, 0, time.UTC).Unix() * 1000
	expectedValue.TimezoneUtcOffset = 120
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
}
