package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestIsUnixTimeEqualsYearAndMonth(t *testing.T) {
	actualValue := IsUnixTimeEqualsYearAndMonth(1691947440, time.UTC, 2023, 8)
	assert.Equal(t, true, actualValue)

	actualValue = IsUnixTimeEqualsYearAndMonth(1690847999, time.UTC, 2023, 8)
	assert.Equal(t, false, actualValue)
}

func TestGetMaxDayOfMonth(t *testing.T) {
	expectedValue := 31
	actualValue := GetMaxDayOfMonth(2023, 1)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = 28
	actualValue = GetMaxDayOfMonth(2023, 2)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = 29
	actualValue = GetMaxDayOfMonth(2024, 2)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = 30
	actualValue = GetMaxDayOfMonth(2023, 4)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = 31
	actualValue = GetMaxDayOfMonth(2023, 12)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = 28
	actualValue = GetMaxDayOfMonth(2100, 2)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetDaysBetweenDates(t *testing.T) {
	shanghaiLocation, err := time.LoadLocation("Asia/Shanghai")
	assert.Equal(t, nil, err)

	startDate := time.Date(2026, 3, 28, 0, 0, 0, 0, shanghaiLocation)
	endDate := time.Date(2026, 3, 30, 0, 0, 0, 0, shanghaiLocation)
	expectedValue := 2
	actualValue := GetDaysBetweenDates(startDate, endDate)
	assert.Equal(t, expectedValue, actualValue)

	// across a leap day
	startDate = time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)
	endDate = time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	expectedValue = 2
	actualValue = GetDaysBetweenDates(startDate, endDate)
	assert.Equal(t, expectedValue, actualValue)

	// across a year boundary
	startDate = time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	endDate = time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedValue = 1
	actualValue = GetDaysBetweenDates(startDate, endDate)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetDaysBetweenDates_FixedTimezone(t *testing.T) {
	timezone := time.FixedZone("Test Timezone", 120*60)
	startDate := time.Date(2026, 3, 28, 0, 0, 0, 0, timezone)
	endDate := time.Date(2026, 3, 30, 0, 0, 0, 0, timezone)
	expectedValue := 2
	actualValue := GetDaysBetweenDates(startDate, endDate)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = -2
	actualValue = GetDaysBetweenDates(endDate, startDate)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetDaysBetweenDates_TimezoneWithDST(t *testing.T) {
	berlinLocation, err := time.LoadLocation("Europe/Berlin")
	assert.Equal(t, nil, err)
	sydneyLocation, err := time.LoadLocation("Australia/Sydney")
	assert.Equal(t, nil, err)
	lordHoweLocation, err := time.LoadLocation("Australia/Lord_Howe")
	assert.Equal(t, nil, err)

	// daylight saving time starts (47 elapsed hours)
	startDate := time.Date(2026, 3, 28, 0, 0, 0, 0, berlinLocation)
	endDate := time.Date(2026, 3, 30, 0, 0, 0, 0, berlinLocation)
	expectedValue := 2
	actualValue := GetDaysBetweenDates(startDate, endDate)
	assert.Equal(t, expectedValue, actualValue)

	// daylight saving time ends (49 elapsed hours)
	startDate = time.Date(2026, 10, 24, 0, 0, 0, 0, berlinLocation)
	endDate = time.Date(2026, 10, 26, 0, 0, 0, 0, berlinLocation)
	expectedValue = 2
	actualValue = GetDaysBetweenDates(startDate, endDate)
	assert.Equal(t, expectedValue, actualValue)

	// across standard time and daylight saving time
	startDate = time.Date(2026, 1, 1, 0, 0, 0, 0, berlinLocation)
	endDate = time.Date(2026, 7, 1, 0, 0, 0, 0, berlinLocation)
	expectedValue = 181
	actualValue = GetDaysBetweenDates(startDate, endDate)
	assert.Equal(t, expectedValue, actualValue)

	// daylight saving time starts in the southern hemisphere
	startDate = time.Date(2026, 10, 3, 0, 0, 0, 0, sydneyLocation)
	endDate = time.Date(2026, 10, 5, 0, 0, 0, 0, sydneyLocation)
	expectedValue = 2
	actualValue = GetDaysBetweenDates(startDate, endDate)
	assert.Equal(t, expectedValue, actualValue)

	// daylight saving time ends in the southern hemisphere
	startDate = time.Date(2026, 4, 4, 0, 0, 0, 0, sydneyLocation)
	endDate = time.Date(2026, 4, 6, 0, 0, 0, 0, sydneyLocation)
	expectedValue = 2
	actualValue = GetDaysBetweenDates(startDate, endDate)
	assert.Equal(t, expectedValue, actualValue)

	// daylight saving time starts with a half-hour change
	startDate = time.Date(2026, 10, 3, 0, 0, 0, 0, lordHoweLocation)
	endDate = time.Date(2026, 10, 5, 0, 0, 0, 0, lordHoweLocation)
	expectedValue = 2
	actualValue = GetDaysBetweenDates(startDate, endDate)
	assert.Equal(t, expectedValue, actualValue)

	// the end date is before the start date across daylight saving time
	startDate = time.Date(2026, 3, 30, 0, 0, 0, 0, berlinLocation)
	endDate = time.Date(2026, 3, 29, 0, 0, 0, 0, berlinLocation)
	expectedValue = -1
	actualValue = GetDaysBetweenDates(startDate, endDate)
	assert.Equal(t, expectedValue, actualValue)

	// the start date and end date are the same
	startDate = time.Date(2026, 3, 29, 0, 0, 0, 0, berlinLocation)
	endDate = time.Date(2026, 3, 29, 0, 0, 0, 0, berlinLocation)
	expectedValue = 0
	actualValue = GetDaysBetweenDates(startDate, endDate)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetTimezoneOffsetMinutes_FixedTimezone(t *testing.T) {
	timezone := time.FixedZone("Test Timezone", 120*60)
	expectedValue := int16(120)
	actualValue := GetTimezoneOffsetMinutes(time.Now().Unix(), timezone)
	assert.Equal(t, expectedValue, actualValue)

	timezone = time.FixedZone("Test Timezone", 345*60)
	expectedValue = int16(345)
	actualValue = GetTimezoneOffsetMinutes(time.Now().Unix(), timezone)
	assert.Equal(t, expectedValue, actualValue)

	timezone = time.FixedZone("Test Timezone", -720*60)
	expectedValue = int16(-720)
	actualValue = GetTimezoneOffsetMinutes(time.Now().Unix(), timezone)
	assert.Equal(t, expectedValue, actualValue)

	timezone = time.FixedZone("Test Timezone", 0)
	expectedValue = int16(0)
	actualValue = GetTimezoneOffsetMinutes(time.Now().Unix(), timezone)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetTimezoneOffsetMinutes_TimezoneWithDST(t *testing.T) {
	londonLocation, err := time.LoadLocation("Europe/London")
	assert.Equal(t, nil, err)

	// during standard time (UTC+0)
	expectedValue := int16(0)
	actualValue := GetTimezoneOffsetMinutes(1577858483, londonLocation) // 2020-01-01 06:01:23 +00:00
	assert.Equal(t, expectedValue, actualValue)

	// during daylight saving time (UTC+1)
	expectedValue = int16(60)
	actualValue = GetTimezoneOffsetMinutes(1619845283, londonLocation) // 2021-05-01 06:01:23 +01:00
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetMinimumTimezoneOffsetMinutes_TimezoneWithNegativeDST(t *testing.T) {
	dublinLocation, err := time.LoadLocation("Europe/Dublin")
	assert.Equal(t, nil, err)
	casablancaLocation, err := time.LoadLocation("Africa/Casablanca")
	assert.Equal(t, nil, err)

	referenceUnixTime := time.Now().Unix()
	expectedValue := int16(0)
	actualValue := GetMinimumTimezoneOffsetMinutes(referenceUnixTime, dublinLocation)
	assert.Equal(t, expectedValue, actualValue)

	// the lower offset occurs during Ramadan, outside both January and July
	referenceUnixTime = time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue = int16(0)
	actualValue = GetMinimumTimezoneOffsetMinutes(referenceUnixTime, casablancaLocation)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetMinimumTimezoneOffsetMinutes_YearBoundary(t *testing.T) {
	saoTomeLocation, err := time.LoadLocation("Africa/Sao_Tome")
	assert.Equal(t, nil, err)
	moscowLocation, err := time.LoadLocation("Europe/Moscow")
	assert.Equal(t, nil, err)

	// UTC+0 applies for only the first hour of 2018, before the change to UTC+1
	referenceUnixTime := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue := int16(0)
	actualValue := GetMinimumTimezoneOffsetMinutes(referenceUnixTime, saoTomeLocation)
	assert.Equal(t, expectedValue, actualValue)

	// the lower offset introduced in 2014 must not affect the minimum for 2013
	referenceUnixTime = time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue = int16(240)
	actualValue = GetMinimumTimezoneOffsetMinutes(referenceUnixTime, moscowLocation)
	assert.Equal(t, expectedValue, actualValue)

	referenceUnixTime = time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue = int16(180)
	actualValue = GetMinimumTimezoneOffsetMinutes(referenceUnixTime, moscowLocation)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetMinimumTimezoneOffsetMinutes_FutureLeapYear(t *testing.T) {
	berlinLocation, err := time.LoadLocation("Europe/Berlin")
	assert.Equal(t, nil, err)
	sydneyLocation, err := time.LoadLocation("Australia/Sydney")
	assert.Equal(t, nil, err)
	referenceUnixTime := time.Date(2040, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

	// future rules must advance past the last day of a leap year
	expectedValue := int16(60)
	actualValue := GetMinimumTimezoneOffsetMinutes(referenceUnixTime, berlinLocation)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = int16(600)
	actualValue = GetMinimumTimezoneOffsetMinutes(referenceUnixTime, sydneyLocation)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetScheduledTransactionScheduledAt_FixedTimezone(t *testing.T) {
	referenceUnixTime := time.Now().Unix()

	timezone := time.FixedZone("Test Timezone", 0)
	expectedValue := int16(0)
	actualValue := GetScheduledTransactionScheduledAt(referenceUnixTime, timezone)
	assert.Equal(t, expectedValue, actualValue)

	timezone = time.FixedZone("Test Timezone", 120*60)
	expectedValue = int16(1320)
	actualValue = GetScheduledTransactionScheduledAt(referenceUnixTime, timezone)
	assert.Equal(t, expectedValue, actualValue)

	// legacy templates with no creation time retain the fixed-offset schedule
	actualValue = GetScheduledTransactionScheduledAt(0, timezone)
	assert.Equal(t, expectedValue, actualValue)

	timezone = time.FixedZone("Test Timezone", 345*60)
	expectedValue = int16(1095)
	actualValue = GetScheduledTransactionScheduledAt(referenceUnixTime, timezone)
	assert.Equal(t, expectedValue, actualValue)

	timezone = time.FixedZone("Test Timezone", -300*60)
	expectedValue = int16(300)
	actualValue = GetScheduledTransactionScheduledAt(referenceUnixTime, timezone)
	assert.Equal(t, expectedValue, actualValue)

	timezone = time.FixedZone("Test Timezone", -210*60)
	expectedValue = int16(210)
	actualValue = GetScheduledTransactionScheduledAt(referenceUnixTime, timezone)
	assert.Equal(t, expectedValue, actualValue)

	timezone = time.FixedZone("Test Timezone", -720*60)
	expectedValue = int16(720)
	actualValue = GetScheduledTransactionScheduledAt(referenceUnixTime, timezone)
	assert.Equal(t, expectedValue, actualValue)

	timezone = time.FixedZone("Test Timezone", 840*60)
	expectedValue = int16(600)
	actualValue = GetScheduledTransactionScheduledAt(referenceUnixTime, timezone)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetScheduledTransactionScheduledAt_TimezoneWithDST(t *testing.T) {
	referenceUnixTime := time.Now().Unix()

	berlinLocation, err := time.LoadLocation("Europe/Berlin")
	assert.Equal(t, nil, err)
	londonLocation, err := time.LoadLocation("Europe/London")
	assert.Equal(t, nil, err)
	newYorkLocation, err := time.LoadLocation("America/New_York")
	assert.Equal(t, nil, err)
	sydneyLocation, err := time.LoadLocation("Australia/Sydney")
	assert.Equal(t, nil, err)
	lordHoweLocation, err := time.LoadLocation("Australia/Lord_Howe")
	assert.Equal(t, nil, err)

	// the minimum yearly offset uses standard time in the northern hemisphere
	expectedValue := int16(1380)
	actualValue := GetScheduledTransactionScheduledAt(referenceUnixTime, berlinLocation)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = int16(0)
	actualValue = GetScheduledTransactionScheduledAt(referenceUnixTime, londonLocation)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = int16(300)
	actualValue = GetScheduledTransactionScheduledAt(referenceUnixTime, newYorkLocation)
	assert.Equal(t, expectedValue, actualValue)

	// the minimum yearly offset also uses standard time in the southern hemisphere
	expectedValue = int16(840)
	actualValue = GetScheduledTransactionScheduledAt(referenceUnixTime, sydneyLocation)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = int16(810)
	actualValue = GetScheduledTransactionScheduledAt(referenceUnixTime, lordHoweLocation)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetScheduledTransactionScheduledAt_DifferentYears(t *testing.T) {
	moscowLocation, err := time.LoadLocation("Europe/Moscow")
	assert.Equal(t, nil, err)

	// standard time was UTC+3 in 2010
	createdUnixTime := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue := int16(1260)
	actualValue := GetScheduledTransactionScheduledAt(createdUnixTime, moscowLocation)
	assert.Equal(t, expectedValue, actualValue)

	// UTC+4 was used throughout 2012
	createdUnixTime = time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue = int16(1200)
	actualValue = GetScheduledTransactionScheduledAt(createdUnixTime, moscowLocation)
	assert.Equal(t, expectedValue, actualValue)

	// the minimum changed to UTC+3 in October 2014
	createdUnixTime = time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue = int16(1260)
	actualValue = GetScheduledTransactionScheduledAt(createdUnixTime, moscowLocation)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetScheduledTransactionScheduledAt_CreationYearBoundary(t *testing.T) {
	moscowLocation, err := time.LoadLocation("Europe/Moscow")
	assert.Equal(t, nil, err)

	// use the UTC creation year even when the local date is already in the next year
	createdUnixTime := time.Date(2013, 12, 31, 23, 59, 59, 0, time.UTC).Unix()
	expectedValue := int16(1200)
	actualValue := GetScheduledTransactionScheduledAt(createdUnixTime, moscowLocation)
	assert.Equal(t, expectedValue, actualValue)

	createdUnixTime = time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	expectedValue = int16(1260)
	actualValue = GetScheduledTransactionScheduledAt(createdUnixTime, moscowLocation)
	assert.Equal(t, expectedValue, actualValue)
}
