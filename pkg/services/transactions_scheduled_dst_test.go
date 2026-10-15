package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreateTransactionByScheduledTransactionTemplate_TimezoneWithZeroReferenceOffset(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                       models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType:     models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY,
		ScheduledFrequency:         "1",
		ScheduledTimezoneName:      "Europe/London",
		ScheduledTimezoneUtcOffset: 60,
		CreatedUnixTime:            time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)

	expectedValue := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:  startTimeInUTC.Unix() * 1000,
		CreatedIp:        "127.0.0.1",
		ScheduledCreated: true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)

	// In summer, July 1 midnight is in the previous UTC day.
	startTimeInUTC = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = time.Date(2026, 6, 30, 23, 0, 0, 0, time.UTC).Unix() * 1000
	expectedValue.TimezoneUtcOffset = 60
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)

	// Dublin's negative DST uses the same minimum yearly offset.
	template.ScheduledTimezoneName = "Europe/Dublin"
	location, err = time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = 0
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_TimezoneWithNegativeOffset(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                       models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType:     models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:         "1",
		ScheduledTimezoneName:      "America/New_York",
		ScheduledTimezoneUtcOffset: -240,
		CreatedUnixTime:            time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2026, 12, 1, 5, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC).Unix()

	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   startTimeInUTC.Unix() * 1000,
		TimezoneUtcOffset: -300,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2026, 7, 1, 5, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = time.Date(2026, 7, 1, 4, 0, 0, 0, time.UTC).Unix() * 1000
	expectedValue.TimezoneUtcOffset = -240
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_SouthernTimezoneWithDST(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                       models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType:     models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:         "1",
		ScheduledTimezoneName:      "Australia/Sydney",
		ScheduledTimezoneUtcOffset: 660,
		CreatedUnixTime:            time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2026, 6, 30, 14, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC).Unix()

	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   startTimeInUTC.Unix() * 1000,
		TimezoneUtcOffset: 600,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2026, 12, 31, 14, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = time.Date(2026, 12, 31, 13, 0, 0, 0, time.UTC).Unix() * 1000
	expectedValue.TimezoneUtcOffset = 660
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	// Lord Howe advances by thirty minutes rather than one hour.
	template.ScheduledTimezoneName = "Australia/Lord_Howe"
	location, err = time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC = time.Date(2026, 6, 30, 13, 30, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = 630
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2026, 12, 31, 13, 30, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = time.Date(2026, 12, 31, 13, 0, 0, 0, time.UTC).Unix() * 1000
	expectedValue.TimezoneUtcOffset = 660
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_TimezoneWithNegativeDST(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                       models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType:     models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:         "1",
		ScheduledTimezoneName:      "Africa/Casablanca",
		ScheduledTimezoneUtcOffset: 60,
		CreatedUnixTime:            time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2019, 5, 10, 0, 0, 0, 0, time.UTC)

	// During Ramadan, Casablanca uses UTC+0 instead of UTC+1.
	expectedValue := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:  startTimeInUTC.Unix() * 1000,
		CreatedIp:        "127.0.0.1",
		ScheduledCreated: true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2019, 6, 10, 0, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = time.Date(2019, 6, 9, 23, 0, 0, 0, time.UTC).Unix() * 1000
	expectedValue.TimezoneUtcOffset = 60
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_DSTTransition(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:     "1",
		ScheduledTimezoneName:  "Europe/Berlin",
		CreatedUnixTime:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2026, 3, 28, 23, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC).Unix()

	// March 29 midnight is before the spring transition.
	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   startTimeInUTC.Unix() * 1000,
		TimezoneUtcOffset: 60,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2026, 3, 29, 23, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 3, 29, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = time.Date(2026, 3, 29, 22, 0, 0, 0, time.UTC).Unix() * 1000
	expectedValue.TimezoneUtcOffset = 120
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	// October 25 midnight is before the autumn transition.
	startTimeInUTC = time.Date(2026, 10, 24, 23, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = time.Date(2026, 10, 24, 22, 0, 0, 0, time.UTC).Unix() * 1000
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2026, 10, 25, 23, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = 60
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_CreatedUnixTime(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY,
		ScheduledFrequency:     "1",
		ScheduledTimezoneName:  "Europe/Moscow",
		CreatedUnixTime:        time.Date(2012, 6, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2014, 11, 30, 20, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2014, 11, 30, 0, 0, 0, 0, time.UTC).Unix()

	// A template created in 2012 retains its UTC+4 reference after the 2014 change.
	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   time.Date(2014, 11, 30, 21, 0, 0, 0, time.UTC).Unix() * 1000,
		TimezoneUtcOffset: 180,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	template.CreatedUnixTime = time.Date(2014, 6, 1, 0, 0, 0, 0, time.UTC).Unix()
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)

	startTimeInUTC = time.Date(2014, 11, 30, 21, 0, 0, 0, time.UTC)
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	// The creation year is the UTC year, even when Moscow is already in 2014.
	template.CreatedUnixTime = time.Date(2013, 12, 31, 23, 59, 59, 0, time.UTC).Unix()
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)

	startTimeInUTC = time.Date(2014, 11, 30, 20, 0, 0, 0, time.UTC)
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_TimezoneYearBoundary(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:     "1",
		ScheduledTimezoneName:  "Africa/Sao_Tome",
		CreatedUnixTime:        time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2018, 6, 1, 0, 0, 0, 0, time.UTC)

	// UTC+0 existed for only the first hour of 2018, but fixes that year's schedule.
	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   time.Date(2018, 5, 31, 23, 0, 0, 0, time.UTC).Unix() * 1000,
		TimezoneUtcOffset: 60,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)

	// After the zone returns to UTC+0, the same UTC schedule still has the right date.
	startTimeInUTC = time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = 0
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_EveryNDaysTimezoneWithDST(t *testing.T) {
	startUnixTime := time.Date(2026, 3, 27, 23, 0, 0, 0, time.UTC).Unix()
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_EVERY_N_DAYS,
		ScheduledFrequency:     "2",
		ScheduledStartTime:     &startUnixTime,
		ScheduledTimezoneName:  "Europe/Berlin",
		CreatedUnixTime:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2026, 3, 29, 23, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2026, 3, 29, 0, 0, 0, 0, time.UTC).Unix()

	// March 28 to March 30 is two calendar days, but only 47 elapsed hours.
	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   time.Date(2026, 3, 29, 22, 0, 0, 0, time.UTC).Unix() * 1000,
		TimezoneUtcOffset: 120,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2026, 3, 30, 23, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC).Unix()
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	// October 24 to October 26 is two calendar days, but 49 elapsed hours.
	startUnixTime = time.Date(2026, 10, 23, 22, 0, 0, 0, time.UTC).Unix()
	startTimeInUTC = time.Date(2026, 10, 25, 23, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	expectedValue.TimezoneUtcOffset = 60
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2026, 10, 26, 23, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC).Unix()
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)
}

func TestCreateTransactionByScheduledTransactionTemplate_EveryNDaysDateBoundary(t *testing.T) {
	startUnixTime := time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC).Unix()
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_EVERY_N_DAYS,
		ScheduledFrequency:     "2",
		ScheduledStartTime:     &startUnixTime,
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), time.UTC)
	startTimeInUTC := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)

	// February 29 counts as a calendar day in a leap year.
	expectedValue := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:  startTimeInUTC.Unix() * 1000,
		CreatedIp:        "127.0.0.1",
		ScheduledCreated: true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	startUnixTime = time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix()
	startTimeInUTC = time.Date(2027, 1, 2, 0, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_StartAndEndTimeTimezoneWithDST(t *testing.T) {
	startUnixTime := time.Date(2026, 6, 30, 22, 0, 0, 0, time.UTC).Unix()
	endUnixTime := startUnixTime
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:     "1",
		ScheduledStartTime:     &startUnixTime,
		ScheduledEndTime:       &endUnixTime,
		ScheduledTimezoneName:  "Europe/Berlin",
		CreatedUnixTime:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2026, 6, 30, 23, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC).Unix()

	// Bounds apply to the final transaction time, not the later scheduler time.
	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   time.Date(2026, 6, 30, 22, 0, 0, 0, time.UTC).Unix() * 1000,
		TimezoneUtcOffset: 120,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)

	startUnixTime++
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	startUnixTime--
	endUnixTime--
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)
}
