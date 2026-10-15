package utils

import (
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFormatTimezoneOffset_FixedTimezone(t *testing.T) {
	timezone := time.FixedZone("Test Timezone", 120*60)
	expectedValue := "+02:00"
	actualValue := FormatTimezoneOffset(time.Now().Unix(), timezone)
	assert.Equal(t, expectedValue, actualValue)

	timezone = time.FixedZone("Test Timezone", 345*60)
	expectedValue = "+05:45"
	actualValue = FormatTimezoneOffset(time.Now().Unix(), timezone)
	assert.Equal(t, expectedValue, actualValue)

	timezone = time.FixedZone("Test Timezone", -720*60)
	expectedValue = "-12:00"
	actualValue = FormatTimezoneOffset(time.Now().Unix(), timezone)
	assert.Equal(t, expectedValue, actualValue)

	timezone = time.FixedZone("Test Timezone", -150*60)
	expectedValue = "-02:30"
	actualValue = FormatTimezoneOffset(time.Now().Unix(), timezone)
	assert.Equal(t, expectedValue, actualValue)

	timezone = time.FixedZone("Test Timezone", 0)
	expectedValue = "+00:00"
	actualValue = FormatTimezoneOffset(time.Now().Unix(), timezone)
	assert.Equal(t, expectedValue, actualValue)

	timezone = time.FixedZone("Test Timezone", -30*60)
	expectedValue = "-00:30"
	actualValue = FormatTimezoneOffset(time.Now().Unix(), timezone)
	assert.Equal(t, expectedValue, actualValue)
}

func TestFormatTimezoneOffset_TimezoneWithDST(t *testing.T) {
	londonLocation, err := time.LoadLocation("Europe/London")
	assert.Equal(t, nil, err)

	// during standard time (UTC+0)
	expectedValue := "+00:00"
	actualValue := FormatTimezoneOffset(1577858483, londonLocation) // 2020-01-01 06:01:23 +00:00
	assert.Equal(t, expectedValue, actualValue)

	// during daylight saving time (UTC+1)
	expectedValue = "+01:00"
	actualValue = FormatTimezoneOffset(1619845283, londonLocation) // 2021-05-01 06:01:23 +01:00
	assert.Equal(t, expectedValue, actualValue)
}

func TestFormatTimezoneOffsetFromHoursOffset(t *testing.T) {
	expectedValue := "+02:00"
	actualValue, err := FormatTimezoneOffsetFromHoursOffset("2")
	assert.Nil(t, err)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "+05:45"
	actualValue, err = FormatTimezoneOffsetFromHoursOffset("+5.75")
	assert.Nil(t, err)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "-12:00"
	actualValue, err = FormatTimezoneOffsetFromHoursOffset("-12.00")
	assert.Nil(t, err)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "-02:30"
	actualValue, err = FormatTimezoneOffsetFromHoursOffset("-2.5")
	assert.Nil(t, err)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "+00:00"
	actualValue, err = FormatTimezoneOffsetFromHoursOffset("0")
	assert.Nil(t, err)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "-00:30"
	actualValue, err = FormatTimezoneOffsetFromHoursOffset("-0.5")
	assert.Nil(t, err)
	assert.Equal(t, expectedValue, actualValue)
}

func TestFormatTimezoneOffsetFromHoursOffset_InvalidHoursOffset(t *testing.T) {
	_, err := FormatTimezoneOffsetFromHoursOffset("")
	assert.EqualError(t, err, errs.ErrFormatInvalid.Message)

	_, err = FormatTimezoneOffsetFromHoursOffset("+")
	assert.EqualError(t, err, errs.ErrFormatInvalid.Message)

	_, err = FormatTimezoneOffsetFromHoursOffset("-")
	assert.EqualError(t, err, errs.ErrFormatInvalid.Message)

	_, err = FormatTimezoneOffsetFromHoursOffset("a")
	assert.EqualError(t, err, errs.ErrFormatInvalid.Message)
}

func TestParseFromTimezoneOffset(t *testing.T) {
	expectedValue := time.FixedZone("Timezone", 120*60)
	actualValue, err := ParseFromTimezoneOffset("+02:00")
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = time.FixedZone("Timezone", 345*60)
	actualValue, err = ParseFromTimezoneOffset("+05:45")
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = time.FixedZone("Timezone", -720*60)
	actualValue, err = ParseFromTimezoneOffset("-12:00")
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = time.FixedZone("Timezone", -150*60)
	actualValue, err = ParseFromTimezoneOffset("-02:30")
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = time.FixedZone("Timezone", 0)
	actualValue, err = ParseFromTimezoneOffset("+00:00")
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedValue, actualValue)

	actualValue, err = ParseFromTimezoneOffset("00:00")
	assert.NotEqual(t, nil, err)

	actualValue, err = ParseFromTimezoneOffset("0")
	assert.NotEqual(t, nil, err)

	actualValue, err = ParseFromTimezoneOffset("1000")
	assert.NotEqual(t, nil, err)
}

func TestGetMinTransactionTimeFromUnixTime(t *testing.T) {
	expectedValue := int64(1617228083000)
	actualValue := GetMinTransactionTimeFromUnixTime(1617228083)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetMaxTransactionTimeFromUnixTime(t *testing.T) {
	expectedValue := int64(1617228083999)
	actualValue := GetMaxTransactionTimeFromUnixTime(1617228083)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetUnixTimeFromTransactionTime(t *testing.T) {
	expectedValue := int64(1617228083)
	actualValue := GetUnixTimeFromTransactionTime(1617228083999)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetTransactionTimeRangeByYearMonth(t *testing.T) {
	expectedMinValue := int64(1704016800000)
	expectedMaxValue := int64(1706788799999)
	actualMinValue, actualMaxValue, err := GetTransactionTimeRangeByYearMonth(2024, 1)
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedMinValue, actualMinValue)
	assert.Equal(t, expectedMaxValue, actualMaxValue)
}

func TestGetStartOfDay(t *testing.T) {
	expectedValue := int64(1617148800) // 2021-03-31 00:00:00 UTC
	actualValue := GetStartOfDay(time.Unix(1617228083, 0).In(time.UTC))
	assert.Equal(t, expectedValue, actualValue.Unix())

	expectedValue = int64(1617206400) // 2021-04-01 00:00:00 UTC+8
	actualValue = GetStartOfDay(time.Unix(1617228083, 0).In(time.FixedZone("Test Timezone", 28800)))
	assert.Equal(t, expectedValue, actualValue.Unix())
}

func TestParseFromUnixTime(t *testing.T) {
	expectedValue := int64(1617228083)
	actualTime := parseFromUnixTime(expectedValue)
	actualValue := actualTime.Unix()
	assert.Equal(t, expectedValue, actualValue)
}
