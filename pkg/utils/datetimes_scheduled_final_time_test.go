package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetScheduledTransactionFinalTime_FixedTimezone(t *testing.T) {
	referenceUnixTime := time.Now().Unix()

	todayFirstUnixTimeInUTC := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC).Unix()

	timezone := time.FixedZone("Test Timezone", 0)
	expectedValue := time.Date(2026, 11, 30, 0, 0, 0, 0, timezone)
	actualValue := GetScheduledTransactionFinalTime(todayFirstUnixTimeInUTC, 0, referenceUnixTime, timezone)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	// legacy templates retain their fixed offset and advance to the next local date
	timezone = time.FixedZone("Test Timezone", 120*60)
	expectedValue = time.Date(2026, 12, 1, 0, 0, 0, 0, timezone)
	actualValue = GetScheduledTransactionFinalTime(todayFirstUnixTimeInUTC, 1320, referenceUnixTime, timezone)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	timezone = time.FixedZone("Test Timezone", 345*60)
	expectedValue = time.Date(2026, 12, 1, 0, 0, 0, 0, timezone)
	actualValue = GetScheduledTransactionFinalTime(todayFirstUnixTimeInUTC, 1095, referenceUnixTime, timezone)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	// negative offsets retain the same local date
	timezone = time.FixedZone("Test Timezone", -300*60)
	expectedValue = time.Date(2026, 11, 30, 0, 0, 0, 0, timezone)
	actualValue = GetScheduledTransactionFinalTime(todayFirstUnixTimeInUTC, 300, referenceUnixTime, timezone)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	timezone = time.FixedZone("Test Timezone", -210*60)
	expectedValue = time.Date(2026, 11, 30, 0, 0, 0, 0, timezone)
	actualValue = GetScheduledTransactionFinalTime(todayFirstUnixTimeInUTC, 210, referenceUnixTime, timezone)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	timezone = time.FixedZone("Test Timezone", -720*60)
	expectedValue = time.Date(2026, 11, 30, 0, 0, 0, 0, timezone)
	actualValue = GetScheduledTransactionFinalTime(todayFirstUnixTimeInUTC, 720, referenceUnixTime, timezone)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	timezone = time.FixedZone("Test Timezone", 840*60)
	expectedValue = time.Date(2026, 12, 1, 0, 0, 0, 0, timezone)
	actualValue = GetScheduledTransactionFinalTime(todayFirstUnixTimeInUTC, 600, referenceUnixTime, timezone)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}
}

func TestGetScheduledTransactionFinalTime_DateBoundary(t *testing.T) {
	referenceUnixTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

	shanghaiLocation, err := time.LoadLocation("Asia/Shanghai")
	assert.Equal(t, nil, err)

	// across a year boundary
	expectedValue := time.Date(2027, 1, 1, 0, 0, 0, 0, shanghaiLocation)
	actualValue := GetScheduledTransactionFinalTime(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix(), 960, referenceUnixTime, shanghaiLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	// across a leap day
	expectedValue = time.Date(2024, 2, 29, 0, 0, 0, 0, shanghaiLocation)
	actualValue = GetScheduledTransactionFinalTime(time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC).Unix(), 960, referenceUnixTime, shanghaiLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	expectedValue = time.Date(2024, 3, 1, 0, 0, 0, 0, shanghaiLocation)
	actualValue = GetScheduledTransactionFinalTime(time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC).Unix(), 960, referenceUnixTime, shanghaiLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}
}

func TestGetScheduledTransactionFinalTime_TimezoneWithDST(t *testing.T) {
	referenceUnixTime := time.Now().Unix()

	berlinLocation, err := time.LoadLocation("Europe/Berlin")
	assert.Equal(t, nil, err)
	londonLocation, err := time.LoadLocation("Europe/London")
	assert.Equal(t, nil, err)
	newYorkLocation, err := time.LoadLocation("America/New_York")
	assert.Equal(t, nil, err)

	// during standard time (UTC+1)
	expectedValue := time.Date(2026, 12, 1, 0, 0, 0, 0, berlinLocation)
	actualValue := GetScheduledTransactionFinalTime(time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC).Unix(), 1380, referenceUnixTime, berlinLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	// during daylight saving time (UTC+2), the scheduled UTC minute stays unchanged
	expectedValue = time.Date(2026, 7, 1, 0, 0, 0, 0, berlinLocation)
	actualValue = GetScheduledTransactionFinalTime(time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC).Unix(), 1380, referenceUnixTime, berlinLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	// the reference offset is zero, but the transaction date uses daylight saving time
	expectedValue = time.Date(2026, 7, 1, 0, 0, 0, 0, londonLocation)
	actualValue = GetScheduledTransactionFinalTime(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC).Unix(), 0, referenceUnixTime, londonLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	// during daylight saving time with a negative offset (UTC-4)
	expectedValue = time.Date(2026, 7, 1, 0, 0, 0, 0, newYorkLocation)
	actualValue = GetScheduledTransactionFinalTime(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC).Unix(), 300, referenceUnixTime, newYorkLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}
}

func TestGetScheduledTransactionFinalTime_DSTTransition(t *testing.T) {
	referenceUnixTime := time.Now().Unix()

	berlinLocation, err := time.LoadLocation("Europe/Berlin")
	assert.Equal(t, nil, err)

	// midnight before daylight saving time starts (UTC+1)
	expectedValue := time.Date(2026, 3, 29, 0, 0, 0, 0, berlinLocation)
	actualValue := GetScheduledTransactionFinalTime(time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC).Unix(), 1380, referenceUnixTime, berlinLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	// midnight after daylight saving time starts (UTC+2)
	expectedValue = time.Date(2026, 3, 30, 0, 0, 0, 0, berlinLocation)
	actualValue = GetScheduledTransactionFinalTime(time.Date(2026, 3, 29, 0, 0, 0, 0, time.UTC).Unix(), 1380, referenceUnixTime, berlinLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	// midnight before daylight saving time ends (UTC+2)
	expectedValue = time.Date(2026, 10, 25, 0, 0, 0, 0, berlinLocation)
	actualValue = GetScheduledTransactionFinalTime(time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC).Unix(), 1380, referenceUnixTime, berlinLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	// midnight after daylight saving time ends (UTC+1)
	expectedValue = time.Date(2026, 10, 26, 0, 0, 0, 0, berlinLocation)
	actualValue = GetScheduledTransactionFinalTime(time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC).Unix(), 1380, referenceUnixTime, berlinLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}
}

func TestGetScheduledTransactionFinalTime_SouthernTimezoneWithDST(t *testing.T) {
	referenceUnixTime := time.Now().Unix()

	sydneyLocation, err := time.LoadLocation("Australia/Sydney")
	assert.Equal(t, nil, err)
	lordHoweLocation, err := time.LoadLocation("Australia/Lord_Howe")
	assert.Equal(t, nil, err)

	// during standard time (UTC+10), using the minimum offset in the creation year
	expectedValue := time.Date(2026, 7, 1, 0, 0, 0, 0, sydneyLocation)
	actualValue := GetScheduledTransactionFinalTime(time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC).Unix(), 840, referenceUnixTime, sydneyLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	// during daylight saving time (UTC+11)
	expectedValue = time.Date(2027, 1, 1, 0, 0, 0, 0, sydneyLocation)
	actualValue = GetScheduledTransactionFinalTime(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix(), 840, referenceUnixTime, sydneyLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	// during standard time with a half-hour daylight saving time difference (UTC+10:30)
	expectedValue = time.Date(2026, 7, 1, 0, 0, 0, 0, lordHoweLocation)
	actualValue = GetScheduledTransactionFinalTime(time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC).Unix(), 810, referenceUnixTime, lordHoweLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	// during daylight saving time (UTC+11)
	expectedValue = time.Date(2027, 1, 1, 0, 0, 0, 0, lordHoweLocation)
	actualValue = GetScheduledTransactionFinalTime(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix(), 810, referenceUnixTime, lordHoweLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}
}

func TestGetScheduledTransactionFinalTime_CreatedUnixTime(t *testing.T) {
	moscowLocation, err := time.LoadLocation("Europe/Moscow")
	assert.Equal(t, nil, err)
	todayFirstUnixTimeInUTC := time.Date(2014, 11, 30, 0, 0, 0, 0, time.UTC).Unix()

	// a template created in 2012 keeps its UTC+4 reference offset when executed in 2014
	createdUnixTime := time.Date(2012, 6, 1, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue := time.Date(2014, 12, 1, 0, 0, 0, 0, moscowLocation)
	actualValue := GetScheduledTransactionFinalTime(todayFirstUnixTimeInUTC, 1200, createdUnixTime, moscowLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}

	// a template created in 2014 uses UTC+3 for the same transaction date
	createdUnixTime = time.Date(2014, 6, 1, 0, 0, 0, 0, time.UTC).Unix()
	actualValue = GetScheduledTransactionFinalTime(todayFirstUnixTimeInUTC, 1260, createdUnixTime, moscowLocation)
	assert.NotNil(t, actualValue)
	if actualValue != nil {
		assert.Equal(t, expectedValue, *actualValue)
	}
}

func TestGetScheduledTransactionFinalTime_TimezoneWithMidnightDST(t *testing.T) {
	santiagoLocation, err := time.LoadLocation("America/Santiago")
	assert.Equal(t, nil, err)
	havanaLocation, err := time.LoadLocation("America/Havana")
	assert.Equal(t, nil, err)
	cairoLocation, err := time.LoadLocation("Africa/Cairo")
	assert.Equal(t, nil, err)
	createdUnixTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

	scheduledAt := GetScheduledTransactionScheduledAt(createdUnixTime, santiagoLocation)
	expectedValue := "2026-09-06T01:00:00-03:00"
	actualTime := GetScheduledTransactionFinalTime(time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC).Unix(), scheduledAt, createdUnixTime, santiagoLocation)
	assert.NotNil(t, actualTime)
	if actualTime != nil {
		actualValue := actualTime.Format(time.RFC3339)
		assert.Equal(t, expectedValue, actualValue)
	}

	scheduledAt = GetScheduledTransactionScheduledAt(createdUnixTime, havanaLocation)
	expectedValue = "2026-03-08T01:00:00-04:00"
	actualTime = GetScheduledTransactionFinalTime(time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC).Unix(), scheduledAt, createdUnixTime, havanaLocation)
	assert.NotNil(t, actualTime)
	if actualTime != nil {
		actualValue := actualTime.Format(time.RFC3339)
		assert.Equal(t, expectedValue, actualValue)
	}

	// Repeated midnight uses its first occurrence.
	expectedValue = "2026-11-01T00:00:00-04:00"
	actualTime = GetScheduledTransactionFinalTime(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC).Unix(), scheduledAt, createdUnixTime, havanaLocation)
	assert.NotNil(t, actualTime)
	if actualTime != nil {
		actualValue := actualTime.Format(time.RFC3339)
		assert.Equal(t, expectedValue, actualValue)
	}

	scheduledAt = GetScheduledTransactionScheduledAt(createdUnixTime, cairoLocation)
	expectedValue = "2026-04-24T01:00:00+03:00"
	actualTime = GetScheduledTransactionFinalTime(time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC).Unix(), scheduledAt, createdUnixTime, cairoLocation)
	assert.NotNil(t, actualTime)
	if actualTime != nil {
		actualValue := actualTime.Format(time.RFC3339)
		assert.Equal(t, expectedValue, actualValue)
	}
}

func TestGetScheduledTransactionFinalTime_SkippedDate(t *testing.T) {
	apiaLocation, err := time.LoadLocation("Pacific/Apia")
	assert.Equal(t, nil, err)
	createdUnixTime := time.Date(2011, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	scheduledAt := GetScheduledTransactionScheduledAt(createdUnixTime, apiaLocation)

	actualTime := GetScheduledTransactionFinalTime(time.Date(2011, 12, 30, 0, 0, 0, 0, time.UTC).Unix(), scheduledAt, createdUnixTime, apiaLocation)
	assert.Nil(t, actualTime)

	expectedValue := "2011-12-31T00:00:00+14:00"
	actualTime = GetScheduledTransactionFinalTime(time.Date(2011, 12, 31, 0, 0, 0, 0, time.UTC).Unix(), scheduledAt, createdUnixTime, apiaLocation)
	assert.NotNil(t, actualTime)
	if actualTime != nil {
		actualValue := actualTime.Format(time.RFC3339)
		assert.Equal(t, expectedValue, actualValue)
	}
}

func TestGetScheduledTransactionFinalTime_NonMidnight(t *testing.T) {
	timezone := time.FixedZone("Template Timezone", 120*60)
	createdUnixTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	scheduledAt := GetScheduledTransactionScheduledAt(createdUnixTime, timezone) + 15

	// Existing non-midnight schedules retain their wall-clock time.
	expectedValue := "2026-12-01T00:15:00+02:00"
	actualTime := GetScheduledTransactionFinalTime(time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC).Unix(), scheduledAt, createdUnixTime, timezone)
	assert.NotNil(t, actualTime)
	if actualTime != nil {
		actualValue := actualTime.Format(time.RFC3339)
		assert.Equal(t, expectedValue, actualValue)
	}
}
