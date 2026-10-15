package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreateTransactionByScheduledTransactionTemplate_DisabledFrequency(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED,
		ScheduledFrequency:     "1",
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), time.UTC)
	startTimeInUTC := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)

	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)
}

func TestCreateTransactionByScheduledTransactionTemplate_InvalidFrequencyType(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TransactionScheduleFrequencyType(255),
		ScheduledFrequency:     "1",
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), time.UTC)
	startTimeInUTC := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)

	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)
}

func TestCreateTransactionByScheduledTransactionTemplate_EmptyFrequency(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), time.UTC)
	startTimeInUTC := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)

	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	template.ScheduledFrequencyType = models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_WEEKLY
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	template.ScheduledFrequencyType = models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	template.ScheduledFrequencyType = models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_YEARLY
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	template.ScheduledFrequencyType = models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_EVERY_N_DAYS
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)
}

func TestCreateTransactionByScheduledTransactionTemplate_InvalidFrequency(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:     "invalid",
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), time.UTC)
	startTimeInUTC := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)

	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	template.ScheduledFrequency = "1,invalid"
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	template.ScheduledFrequency = "9223372036854775808"
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	template.ScheduledFrequency = "1,"
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)
}

func TestCreateTransactionByScheduledTransactionTemplate_Daily(t *testing.T) {
	template := &models.TransactionTemplate{
		TemplateId:             1001,
		Uid:                    123,
		TemplateType:           models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE,
		Name:                   "Scheduled Expense",
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		CategoryId:             234,
		AccountId:              345,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:     "1",
		TagIds:                 "1001,1002,1003",
		Amount:                 456,
		RelatedAccountId:       567,
		RelatedAccountAmount:   678,
		HideAmount:             true,
		Comment:                "Scheduled transaction comment",
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), time.UTC)
	startTimeInUTC := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)

	expectedValue := &models.Transaction{
		Uid:              123,
		Type:             models.TRANSACTION_DB_TYPE_EXPENSE,
		CategoryId:       234,
		TransactionTime:  startTimeInUTC.Unix() * 1000,
		AccountId:        345,
		Amount:           456,
		HideAmount:       true,
		Comment:          "Scheduled transaction comment",
		CreatedIp:        "127.0.0.1",
		ScheduledCreated: true,
	}
	expectedTagIds := []int64{1001, 1002, 1003}
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, expectedTagIds, actualTagIds)

	// The following day creates another transaction with the same template data.
	startTimeInUTC = time.Date(2026, 12, 2, 0, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, expectedTagIds, actualTagIds)
}

func TestCreateTransactionByScheduledTransactionTemplate_Income(t *testing.T) {
	template := &models.TransactionTemplate{
		Uid:                    123,
		Type:                   models.TRANSACTION_TYPE_INCOME,
		CategoryId:             234,
		AccountId:              345,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:     "1",
		Amount:                 456,
		RelatedAccountId:       567,
		RelatedAccountAmount:   678,
		Comment:                "Scheduled income",
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), time.UTC)
	startTimeInUTC := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)

	expectedValue := &models.Transaction{
		Uid:              123,
		Type:             models.TRANSACTION_DB_TYPE_INCOME,
		CategoryId:       234,
		TransactionTime:  startTimeInUTC.Unix() * 1000,
		AccountId:        345,
		Amount:           456,
		Comment:          "Scheduled income",
		CreatedIp:        "127.0.0.1",
		ScheduledCreated: true,
	}
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, []int64{}, actualTagIds)
}

func TestCreateTransactionByScheduledTransactionTemplate_Transfer(t *testing.T) {
	template := &models.TransactionTemplate{
		Uid:                    123,
		Type:                   models.TRANSACTION_TYPE_TRANSFER,
		CategoryId:             234,
		AccountId:              345,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:     "1",
		TagIds:                 "1001,1002",
		Amount:                 456,
		RelatedAccountId:       567,
		RelatedAccountAmount:   678,
		HideAmount:             true,
		Comment:                "Scheduled transfer",
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), time.UTC)
	startTimeInUTC := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)

	expectedValue := &models.Transaction{
		Uid:                  123,
		Type:                 models.TRANSACTION_DB_TYPE_TRANSFER_OUT,
		CategoryId:           234,
		TransactionTime:      startTimeInUTC.Unix() * 1000,
		AccountId:            345,
		Amount:               456,
		RelatedAccountId:     567,
		RelatedAccountAmount: 678,
		HideAmount:           true,
		Comment:              "Scheduled transfer",
		CreatedIp:            "127.0.0.1",
		ScheduledCreated:     true,
	}
	expectedTagIds := []int64{1001, 1002}
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, expectedTagIds, actualTagIds)
}

func TestCreateTransactionByScheduledTransactionTemplate_InvalidTransactionType(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_MODIFY_BALANCE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DAILY,
		ScheduledFrequency:     "1",
		TagIds:                 "1001",
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), time.UTC)
	startTimeInUTC := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)

	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	template.Type = models.TransactionType(255)
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)
}

func TestCreateTransactionByScheduledTransactionTemplate_Weekly(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                       models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType:     models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_WEEKLY,
		ScheduledFrequency:         "0,2,4",
		ScheduledTimezoneUtcOffset: 120,
	}
	location := time.FixedZone("Template Timezone", int(template.ScheduledTimezoneUtcOffset)*60)
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), location)

	startTimeInUTC := time.Date(2026, 11, 30, 22, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC).Unix()

	// The local date is Tuesday, while the scheduler's UTC date is Monday.
	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   startTimeInUTC.Unix() * 1000,
		TimezoneUtcOffset: 120,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, []int64{}, actualTagIds)

	template.ScheduledFrequency = "1,3,5"
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	// Sunday is represented by zero.
	template.ScheduledFrequency = "0"
	startTimeInUTC = time.Date(2026, 11, 28, 22, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 11, 28, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, []int64{}, actualTagIds)
}

func TestCreateTransactionByScheduledTransactionTemplate_Monthly(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                       models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType:     models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY,
		ScheduledFrequency:         "1,15",
		ScheduledTimezoneName:      "Europe/Berlin",
		ScheduledTimezoneUtcOffset: 120,
		CreatedUnixTime:            time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	location, err := time.LoadLocation(template.ScheduledTimezoneName)
	assert.Equal(t, nil, err)

	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(template.CreatedUnixTime, location)
	startTimeInUTC := time.Date(2026, 11, 30, 23, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC).Unix()

	// A template created in summer uses the winter offset for December 1.
	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   startTimeInUTC.Unix() * 1000,
		TimezoneUtcOffset: 60,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, []int64{}, actualTagIds)

	template.ScheduledFrequency = "30"
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	// The UTC scheduler stays at 23:00, but the summer transaction is at 22:00.
	template.ScheduledFrequency = "1,15"
	startTimeInUTC = time.Date(2026, 6, 30, 23, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue.TransactionTime = time.Date(2026, 6, 30, 22, 0, 0, 0, time.UTC).Unix() * 1000
	expectedValue.TimezoneUtcOffset = 120
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, []int64{}, actualTagIds)
}

func TestCreateTransactionByScheduledTransactionTemplate_MonthlyLastDays(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY,
		ScheduledFrequency:     "1,-1,-2",
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), time.UTC)
	startTimeInUTC := time.Date(2026, 1, 30, 0, 0, 0, 0, time.UTC)

	expectedValue := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:  startTimeInUTC.Unix() * 1000,
		CreatedIp:        "127.0.0.1",
		ScheduledCreated: true,
	}
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, []int64{}, actualTagIds)

	startTimeInUTC = time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)

	// February's last day follows the actual length of the month.
	template.ScheduledFrequency = "-1"
	startTimeInUTC = time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	// A positive day that does not exist in February must not create a transaction.
	template.ScheduledFrequency = "31"
	startTimeInUTC = time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)
}

func TestCreateTransactionByScheduledTransactionTemplate_Yearly(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                       models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType:     models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_YEARLY,
		ScheduledFrequency:         "101,1201",
		ScheduledTimezoneUtcOffset: 120,
	}
	location := time.FixedZone("Template Timezone", int(template.ScheduledTimezoneUtcOffset)*60)
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), location)

	startTimeInUTC := time.Date(2026, 11, 30, 22, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC).Unix()

	expectedValue := &models.Transaction{
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:   startTimeInUTC.Unix() * 1000,
		TimezoneUtcOffset: 120,
		CreatedIp:         "127.0.0.1",
		ScheduledCreated:  true,
	}
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, []int64{}, actualTagIds)

	template.ScheduledFrequency = "1130"
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	startTimeInUTC = time.Date(2026, 12, 31, 22, 0, 0, 0, time.UTC)
	todayFirstUnixTimeInUTC = time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix()
	template.ScheduledFrequency = "101,1201"
	expectedValue.TransactionTime = startTimeInUTC.Unix() * 1000
	actualValue, _ = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, todayFirstUnixTimeInUTC)
	assert.Equal(t, expectedValue, actualValue)
}

func TestCreateTransactionByScheduledTransactionTemplate_YearlyLeapDay(t *testing.T) {
	template := &models.TransactionTemplate{
		Type:                   models.TRANSACTION_TYPE_EXPENSE,
		ScheduledFrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_YEARLY,
		ScheduledFrequency:     "229",
	}
	template.ScheduledAt = utils.GetScheduledTransactionScheduledAt(time.Now().Unix(), time.UTC)
	startTimeInUTC := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)

	expectedValue := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime:  startTimeInUTC.Unix() * 1000,
		CreatedIp:        "127.0.0.1",
		ScheduledCreated: true,
	}
	actualValue, _ := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Equal(t, expectedValue, actualValue)

	startTimeInUTC = time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	actualValue, actualTagIds := Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)

	startTimeInUTC = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	actualValue, actualTagIds = Transactions.createTransactionByScheduledTransactionTemplate(core.NewNullContext(), template, startTimeInUTC, startTimeInUTC.Unix())
	assert.Nil(t, actualValue)
	assert.Nil(t, actualTagIds)
}
