package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestParseNumericYearMonth(t *testing.T) {
	expectedYear := int32(2024)
	expectedMonth := int32(3)
	actualYear, actualMonth, err := ParseNumericYearMonth("2024-03")
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedYear, actualYear)
	assert.Equal(t, expectedMonth, actualMonth)
}

func TestFormatUnixTimeToLongDate(t *testing.T) {
	unixTime := int64(1617228083)
	utcTimezone := time.FixedZone("Test Timezone", 0)      // UTC
	utc8Timezone := time.FixedZone("Test Timezone", 28800) // UTC+8

	expectedValue := "2021-03-31"
	actualValue := FormatUnixTimeToLongDate(unixTime, utcTimezone)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2021-04-01"
	actualValue = FormatUnixTimeToLongDate(unixTime, utc8Timezone)
	assert.Equal(t, expectedValue, actualValue)
}

func TestFormatUnixTimeToLongDateTimeWithTimezone(t *testing.T) {
	unixTime := int64(1617228083)
	utcTimezone := time.FixedZone("Test Timezone", 0)      // UTC
	utc8Timezone := time.FixedZone("Test Timezone", 28800) // UTC+8

	expectedValue := "2021-03-31 22:01:23Z"
	actualValue := FormatUnixTimeToLongDateTimeWithTimezone(unixTime, utcTimezone)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2021-04-01 06:01:23+08:00"
	actualValue = FormatUnixTimeToLongDateTimeWithTimezone(unixTime, utc8Timezone)
	assert.Equal(t, expectedValue, actualValue)
}

func TestFormatUnixTimeToLongDateTimeWithTimezoneRFC3339Format(t *testing.T) {
	unixTime := int64(1617228083)
	utcTimezone := time.FixedZone("Test Timezone", 0)      // UTC
	utc8Timezone := time.FixedZone("Test Timezone", 28800) // UTC+8

	expectedValue := "2021-03-31T22:01:23Z"
	actualValue := FormatUnixTimeToLongDateTimeWithTimezoneRFC3339Format(unixTime, utcTimezone)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2021-04-01T06:01:23+08:00"
	actualValue = FormatUnixTimeToLongDateTimeWithTimezoneRFC3339Format(unixTime, utc8Timezone)
	assert.Equal(t, expectedValue, actualValue)
}

func TestFormatUnixTimeToLongDateTime(t *testing.T) {
	unixTime := int64(1617228083)
	utcTimezone := time.FixedZone("Test Timezone", 0)      // UTC
	utc8Timezone := time.FixedZone("Test Timezone", 28800) // UTC+8

	expectedValue := "2021-03-31 22:01:23"
	actualValue := FormatUnixTimeToLongDateTime(unixTime, utcTimezone)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2021-04-01 06:01:23"
	actualValue = FormatUnixTimeToLongDateTime(unixTime, utc8Timezone)
	assert.Equal(t, expectedValue, actualValue)
}

func TestFormatYearMonthDayToLongDateTime(t *testing.T) {
	expectedValue := "2025-06-01 00:00:00"
	actualValue, err := FormatYearMonthDayToLongDateTime("25", "06", "01")
	assert.Nil(t, err)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2025-06-01 00:00:00"
	actualValue, err = FormatYearMonthDayToLongDateTime("25", "6", "1")
	assert.Nil(t, err)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "1990-06-01 00:00:00"
	actualValue, err = FormatYearMonthDayToLongDateTime("90", "06", "01")
	assert.Nil(t, err)
	assert.Equal(t, expectedValue, actualValue)
}

func TestFormatUnixTimeToLongDateTimeWithoutSecond(t *testing.T) {
	unixTime := int64(1617228083)
	utcTimezone := time.FixedZone("Test Timezone", 0)      // UTC
	utc8Timezone := time.FixedZone("Test Timezone", 28800) // UTC+8

	expectedValue := "2021-03-31 22:01"
	actualValue := FormatUnixTimeToLongDateTimeWithoutSecond(unixTime, utcTimezone)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2021-04-01 06:01"
	actualValue = FormatUnixTimeToLongDateTimeWithoutSecond(unixTime, utc8Timezone)
	assert.Equal(t, expectedValue, actualValue)
}

func TestFormatUnixTimeToYearMonth(t *testing.T) {
	unixTime := int64(1617228083)
	utcTimezone := time.FixedZone("Test Timezone", 0)      // UTC
	utc8Timezone := time.FixedZone("Test Timezone", 28800) // UTC+8

	expectedValue := "2021-03"
	actualValue := FormatUnixTimeToYearMonth(unixTime, utcTimezone)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2021-04"
	actualValue = FormatUnixTimeToYearMonth(unixTime, utc8Timezone)
	assert.Equal(t, expectedValue, actualValue)
}

func TestFormatUnixTimeToNumericYearMonth(t *testing.T) {
	unixTime := int64(1617228083)
	utcTimezone := time.FixedZone("Test Timezone", 0)      // UTC
	utc8Timezone := time.FixedZone("Test Timezone", 28800) // UTC+8

	expectedValue := int32(202103)
	actualValue := FormatUnixTimeToNumericYearMonth(unixTime, utcTimezone)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = int32(202104)
	actualValue = FormatUnixTimeToNumericYearMonth(unixTime, utc8Timezone)
	assert.Equal(t, expectedValue, actualValue)
}

func TestFormatUnixTimeToNumericYearMonthDay(t *testing.T) {
	unixTime := int64(1617228083)
	utcTimezone := time.FixedZone("Test Timezone", 0)      // UTC
	utc8Timezone := time.FixedZone("Test Timezone", 28800) // UTC+8

	expectedValue := int32(20210331)
	actualValue := FormatUnixTimeToNumericYearMonthDay(unixTime, utcTimezone)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = int32(20210401)
	actualValue = FormatUnixTimeToNumericYearMonthDay(unixTime, utc8Timezone)
	assert.Equal(t, expectedValue, actualValue)
}

func TestFormatUnixTimeToNumericLocalDateTime(t *testing.T) {
	unixTime := int64(1617228083)
	utcTimezone := time.FixedZone("Test Timezone", 0)      // UTC
	utc8Timezone := time.FixedZone("Test Timezone", 28800) // UTC+8

	expectedValue := int64(20210331220123)
	actualValue := FormatUnixTimeToNumericLocalDateTime(unixTime, utcTimezone)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = int64(20210401060123)
	actualValue = FormatUnixTimeToNumericLocalDateTime(unixTime, utc8Timezone)
	assert.Equal(t, expectedValue, actualValue)
}

func TestFormatNumericYearMonthDayToLongDate(t *testing.T) {
	expectedValue := "2026-12-31"
	actualValue := FormatNumericYearMonthDayToLongDate(int32(20261231))
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-01-01"
	actualValue = FormatNumericYearMonthDayToLongDate(int32(20260101))
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetMinUnixTimeWithSameLocalDateTime(t *testing.T) {
	expectedValue := int64(1690797600)
	actualValue := GetMinUnixTimeWithSameLocalDateTime(1690819200, 480)
	assert.Equal(t, expectedValue, actualValue)

	actualValue = GetMinUnixTimeWithSameLocalDateTime(1690873200, -420)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetMaxUnixTimeWithSameLocalDateTime(t *testing.T) {
	expectedValue := int64(1690891200)
	actualValue := GetMaxUnixTimeWithSameLocalDateTime(1690819200, 480)
	assert.Equal(t, expectedValue, actualValue)

	actualValue = GetMaxUnixTimeWithSameLocalDateTime(1690873200, -420)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetDateFirstTimeInTimezone(t *testing.T) {
	date := time.Date(2026, 10, 4, 12, 34, 56, 789, time.UTC)

	expectedValue := "2026-10-04T00:00:00Z"
	actualTime := GetDateFirstTimeInTimezone(date, time.UTC)
	actualValue := actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, time.UTC, actualTime.Location())

	nepalLocation := time.FixedZone("Nepal", 345*60)
	expectedValue = "2026-10-04T00:00:00+05:45"
	actualTime = GetDateFirstTimeInTimezone(date, nepalLocation)
	actualValue = actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, nepalLocation, actualTime.Location())

	expectedValue = "2026-10-04T00:00:00-12:00"
	actualValue = GetDateFirstTimeInTimezone(date, time.FixedZone("UTC-12", -720*60)).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-10-04T00:00:00+14:00"
	actualValue = GetDateFirstTimeInTimezone(date, time.FixedZone("UTC+14", 840*60)).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetDateFirstTimeInTimezone_InputDateInDifferentTimezone(t *testing.T) {
	// Use the input's calendar date, even when its UTC date is different.
	date := time.Date(2026, 1, 1, 0, 30, 0, 0, time.FixedZone("UTC+14", 840*60))
	expectedValue := "2026-01-01T00:00:00Z"
	actualValue := GetDateFirstTimeInTimezone(date, time.UTC).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	date = time.Date(2026, 12, 31, 23, 30, 0, 0, time.FixedZone("UTC-12", -720*60))
	expectedValue = "2026-12-31T00:00:00Z"
	actualValue = GetDateFirstTimeInTimezone(date, time.UTC).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetDateFirstTimeInTimezone_TimezoneWithDST(t *testing.T) {
	berlinLocation, err := time.LoadLocation("Europe/Berlin")
	assert.Equal(t, nil, err)
	lordHoweLocation, err := time.LoadLocation("Australia/Lord_Howe")
	assert.Equal(t, nil, err)

	expectedValue := "2026-03-29T00:00:00+01:00"
	actualValue := GetDateFirstTimeInTimezone(time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC), berlinLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-03-30T00:00:00+02:00"
	actualValue = GetDateFirstTimeInTimezone(time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC), berlinLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-10-25T00:00:00+02:00"
	actualValue = GetDateFirstTimeInTimezone(time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC), berlinLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-10-26T00:00:00+01:00"
	actualValue = GetDateFirstTimeInTimezone(time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC), berlinLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-10-04T00:00:00+10:30"
	actualValue = GetDateFirstTimeInTimezone(time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC), lordHoweLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-10-05T00:00:00+11:00"
	actualValue = GetDateFirstTimeInTimezone(time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC), lordHoweLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetDateFirstTimeInTimezone_TimezoneWithMidnightDST(t *testing.T) {
	santiagoLocation, err := time.LoadLocation("America/Santiago")
	assert.Equal(t, nil, err)
	havanaLocation, err := time.LoadLocation("America/Havana")
	assert.Equal(t, nil, err)
	cairoLocation, err := time.LoadLocation("Africa/Cairo")
	assert.Equal(t, nil, err)

	expectedValue := "2026-09-05T00:00:00-04:00"
	actualValue := GetDateFirstTimeInTimezone(time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC), santiagoLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-09-06T01:00:00-03:00"
	actualValue = GetDateFirstTimeInTimezone(time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC), santiagoLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-09-07T00:00:00-03:00"
	actualValue = GetDateFirstTimeInTimezone(time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC), santiagoLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-03-08T01:00:00-04:00"
	actualValue = GetDateFirstTimeInTimezone(time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC), havanaLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-04-24T01:00:00+03:00"
	actualValue = GetDateFirstTimeInTimezone(time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC), cairoLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetDateFirstTimeInTimezone_AmbiguousMidnight(t *testing.T) {
	havanaLocation, err := time.LoadLocation("America/Havana")
	assert.Equal(t, nil, err)

	// Select the first occurrence of midnight when it occurs twice.
	expectedValue := "2026-11-01T00:00:00-04:00"
	actualValue := GetDateFirstTimeInTimezone(time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC), havanaLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-11-02T00:00:00-05:00"
	actualValue = GetDateFirstTimeInTimezone(time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC), havanaLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetDateFirstTimeInTimezone_SkippedDate(t *testing.T) {
	apiaLocation, err := time.LoadLocation("Pacific/Apia")
	assert.Equal(t, nil, err)

	expectedValue := "2011-12-29T00:00:00-10:00"
	actualValue := GetDateFirstTimeInTimezone(time.Date(2011, 12, 29, 12, 0, 0, 0, time.UTC), apiaLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	// A skipped date advances to the beginning of the next existing date.
	expectedValue = "2011-12-31T00:00:00+14:00"
	actualValue = GetDateFirstTimeInTimezone(time.Date(2011, 12, 30, 12, 0, 0, 0, time.UTC), apiaLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	actualValue = GetDateFirstTimeInTimezone(time.Date(2011, 12, 31, 12, 0, 0, 0, time.UTC), apiaLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetDateFirstTimeInTimezone_FutureLeapYear(t *testing.T) {
	berlinLocation, err := time.LoadLocation("Europe/Berlin")
	assert.Equal(t, nil, err)
	sydneyLocation, err := time.LoadLocation("Australia/Sydney")
	assert.Equal(t, nil, err)

	expectedValue := "2040-02-29T00:00:00+01:00"
	actualValue := GetDateFirstTimeInTimezone(time.Date(2040, 2, 29, 12, 0, 0, 0, time.UTC), berlinLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2040-12-31T00:00:00+01:00"
	actualValue = GetDateFirstTimeInTimezone(time.Date(2040, 12, 31, 12, 0, 0, 0, time.UTC), berlinLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2040-12-31T00:00:00+11:00"
	actualValue = GetDateFirstTimeInTimezone(time.Date(2040, 12, 31, 12, 0, 0, 0, time.UTC), sydneyLocation).Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)
}
