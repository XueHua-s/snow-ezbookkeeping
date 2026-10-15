package utils

import (
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestParseFromLongDateFirstTimeInTimezone_TimezoneWithMidnightDST(t *testing.T) {
	santiagoLocation, err := time.LoadLocation("America/Santiago")
	assert.Equal(t, nil, err)
	havanaLocation, err := time.LoadLocation("America/Havana")
	assert.Equal(t, nil, err)
	cairoLocation, err := time.LoadLocation("Africa/Cairo")
	assert.Equal(t, nil, err)

	// Missing midnight must advance to the first valid time on the requested date.
	expectedValue := "2026-09-06T01:00:00-03:00"
	actualTime, err := ParseFromLongDateFirstTimeInTimezone("2026-09-06", santiagoLocation)
	assert.Equal(t, nil, err)
	actualValue := actualTime.Format(time.RFC3339)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-03-08T01:00:00-04:00"
	actualTime, err = ParseFromLongDateFirstTimeInTimezone("2026-03-08", havanaLocation)
	assert.Equal(t, nil, err)
	actualValue = actualTime.Format(time.RFC3339)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-04-24T01:00:00+03:00"
	actualTime, err = ParseFromLongDateFirstTimeInTimezone("2026-04-24", cairoLocation)
	assert.Equal(t, nil, err)
	actualValue = actualTime.Format(time.RFC3339)
	assert.Equal(t, expectedValue, actualValue)

	// The following date starts at midnight again.
	expectedValue = "2026-09-07T00:00:00-03:00"
	actualTime, err = ParseFromLongDateFirstTimeInTimezone("2026-09-07", santiagoLocation)
	assert.Equal(t, nil, err)
	actualValue = actualTime.Format(time.RFC3339)
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateFirstTimeInTimezone_AmbiguousMidnight(t *testing.T) {
	havanaLocation, err := time.LoadLocation("America/Havana")
	assert.Equal(t, nil, err)

	// Midnight occurs twice; the date begins at the first occurrence, still UTC-4.
	expectedValue := "2026-11-01T00:00:00-04:00"
	actualTime, err := ParseFromLongDateFirstTimeInTimezone("2026-11-01", havanaLocation)
	assert.Equal(t, nil, err)
	actualValue := actualTime.Format(time.RFC3339)
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateFirstTimeInTimezone_InvalidDate(t *testing.T) {
	_, err := ParseFromLongDateFirstTimeInTimezone("2026-02-30", time.UTC)
	assert.NotEqual(t, nil, err)

	apiaLocation, err := time.LoadLocation("Pacific/Apia")
	assert.Equal(t, nil, err)

	// This entire local date was skipped when Samoa crossed the date line.
	_, err = ParseFromLongDateFirstTimeInTimezone("2011-12-30", apiaLocation)
	assert.Equal(t, errs.ErrParameterInvalid, err)
}

func TestParseFromLongDateLastTimeInTimezone(t *testing.T) {
	shanghaiLocation, err := time.LoadLocation("Asia/Shanghai")
	assert.Equal(t, nil, err)

	expectedValue := "2026-03-29T23:59:59.999999999+08:00"
	actualTime, err := ParseFromLongDateLastTimeInTimezone("2026-03-29", shanghaiLocation)
	assert.Equal(t, nil, err)

	actualValue := actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-12-31T23:59:59.999999999Z"
	actualTime, err = ParseFromLongDateLastTimeInTimezone("2026-12-31", time.UTC)
	assert.Equal(t, nil, err)

	actualValue = actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateLastTimeInTimezone_TimezoneWithDST(t *testing.T) {
	berlinLocation, err := time.LoadLocation("Europe/Berlin")
	assert.Equal(t, nil, err)
	sydneyLocation, err := time.LoadLocation("Australia/Sydney")
	assert.Equal(t, nil, err)
	lordHoweLocation, err := time.LoadLocation("Australia/Lord_Howe")
	assert.Equal(t, nil, err)

	// daylight saving time starts (23-hour day)
	expectedValue := "2026-03-29T23:59:59.999999999+02:00"
	actualTime, err := ParseFromLongDateLastTimeInTimezone("2026-03-29", berlinLocation)
	assert.Equal(t, nil, err)

	actualValue := actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	// daylight saving time ends (25-hour day)
	expectedValue = "2026-10-25T23:59:59.999999999+01:00"
	actualTime, err = ParseFromLongDateLastTimeInTimezone("2026-10-25", berlinLocation)
	assert.Equal(t, nil, err)

	actualValue = actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	// daylight saving time starts in the southern hemisphere
	expectedValue = "2026-10-04T23:59:59.999999999+11:00"
	actualTime, err = ParseFromLongDateLastTimeInTimezone("2026-10-04", sydneyLocation)
	assert.Equal(t, nil, err)

	actualValue = actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	// daylight saving time ends in the southern hemisphere
	expectedValue = "2026-04-05T23:59:59.999999999+10:00"
	actualTime, err = ParseFromLongDateLastTimeInTimezone("2026-04-05", sydneyLocation)
	assert.Equal(t, nil, err)

	actualValue = actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	// daylight saving time starts with a half-hour change
	expectedValue = "2026-10-04T23:59:59.999999999+11:00"
	actualTime, err = ParseFromLongDateLastTimeInTimezone("2026-10-04", lordHoweLocation)
	assert.Equal(t, nil, err)

	actualValue = actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateLastTimeInTimezone_InvalidDate(t *testing.T) {
	_, err := ParseFromLongDateLastTimeInTimezone("2026-02-30", time.UTC)
	assert.NotEqual(t, nil, err)
}

func TestParseFromLongDateLastTimeInTimezone_TimezoneWithMidnightDST(t *testing.T) {
	santiagoLocation, err := time.LoadLocation("America/Santiago")
	assert.Equal(t, nil, err)
	havanaLocation, err := time.LoadLocation("America/Havana")
	assert.Equal(t, nil, err)
	cairoLocation, err := time.LoadLocation("Africa/Cairo")
	assert.Equal(t, nil, err)

	// The day before missing midnight must end at its own last second.
	expectedValue := "2026-09-05T23:59:59.999999999-04:00"
	actualTime, err := ParseFromLongDateLastTimeInTimezone("2026-09-05", santiagoLocation)
	assert.Equal(t, nil, err)
	actualValue := actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-09-06T23:59:59.999999999-03:00"
	actualTime, err = ParseFromLongDateLastTimeInTimezone("2026-09-06", santiagoLocation)
	assert.Equal(t, nil, err)
	actualValue = actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-09-07T23:59:59.999999999-03:00"
	actualTime, err = ParseFromLongDateLastTimeInTimezone("2026-09-07", santiagoLocation)
	assert.Equal(t, nil, err)
	actualValue = actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-03-07T23:59:59.999999999-05:00"
	actualTime, err = ParseFromLongDateLastTimeInTimezone("2026-03-07", havanaLocation)
	assert.Equal(t, nil, err)
	actualValue = actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-03-08T23:59:59.999999999-04:00"
	actualTime, err = ParseFromLongDateLastTimeInTimezone("2026-03-08", havanaLocation)
	assert.Equal(t, nil, err)
	actualValue = actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-04-23T23:59:59.999999999+02:00"
	actualTime, err = ParseFromLongDateLastTimeInTimezone("2026-04-23", cairoLocation)
	assert.Equal(t, nil, err)
	actualValue = actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	// Starting at 01:00 must not move the end time into April 25.
	expectedValue = "2026-04-24T23:59:59.999999999+03:00"
	actualTime, err = ParseFromLongDateLastTimeInTimezone("2026-04-24", cairoLocation)
	assert.Equal(t, nil, err)
	actualValue = actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateLastTimeInTimezone_AmbiguousMidnight(t *testing.T) {
	havanaLocation, err := time.LoadLocation("America/Havana")
	assert.Equal(t, nil, err)
	cairoLocation, err := time.LoadLocation("Africa/Cairo")
	assert.Equal(t, nil, err)

	// The previous date ends before the first occurrence of repeated midnight.
	expectedValue := "2026-10-31T23:59:59.999999999-04:00"
	actualTime, err := ParseFromLongDateLastTimeInTimezone("2026-10-31", havanaLocation)
	assert.Equal(t, nil, err)
	actualValue := actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2026-11-01T23:59:59.999999999-05:00"
	actualTime, err = ParseFromLongDateLastTimeInTimezone("2026-11-01", havanaLocation)
	assert.Equal(t, nil, err)
	actualValue = actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	// Cairo repeats the previous day's 23:00 hour instead of midnight.
	expectedValue = "2026-10-29T23:59:59.999999999+02:00"
	actualTime, err = ParseFromLongDateLastTimeInTimezone("2026-10-29", cairoLocation)
	assert.Equal(t, nil, err)
	actualValue = actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateLastTimeInTimezone_SkippedDate(t *testing.T) {
	apiaLocation, err := time.LoadLocation("Pacific/Apia")
	assert.Equal(t, nil, err)

	// The next valid date after December 29 is December 31.
	expectedValue := "2011-12-29T23:59:59.999999999-10:00"
	actualTime, err := ParseFromLongDateLastTimeInTimezone("2011-12-29", apiaLocation)
	assert.Equal(t, nil, err)
	actualValue := actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	_, err = ParseFromLongDateLastTimeInTimezone("2011-12-30", apiaLocation)
	assert.Equal(t, errs.ErrParameterInvalid, err)

	expectedValue = "2011-12-31T23:59:59.999999999+14:00"
	actualTime, err = ParseFromLongDateLastTimeInTimezone("2011-12-31", apiaLocation)
	assert.Equal(t, nil, err)
	actualValue = actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateLastTimeInTimezone_FutureLeapYear(t *testing.T) {
	berlinLocation, err := time.LoadLocation("Europe/Berlin")
	assert.Equal(t, nil, err)
	sydneyLocation, err := time.LoadLocation("Australia/Sydney")
	assert.Equal(t, nil, err)

	// Future timezone bounds must advance correctly past a leap year's final day.
	expectedValue := "2040-12-31T23:59:59.999999999+01:00"
	actualTime, err := ParseFromLongDateLastTimeInTimezone("2040-12-31", berlinLocation)
	assert.Equal(t, nil, err)
	actualValue := actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "2040-12-31T23:59:59.999999999+11:00"
	actualTime, err = ParseFromLongDateLastTimeInTimezone("2040-12-31", sydneyLocation)
	assert.Equal(t, nil, err)
	actualValue = actualTime.Format(time.RFC3339Nano)
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateFirstTime(t *testing.T) {
	expectedValue := int64(1690819200)
	actualTime, err := ParseFromLongDateFirstTime("2023-08-01", 480)
	assert.Equal(t, nil, err)

	actualValue := actualTime.Unix()
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateLastTime(t *testing.T) {
	expectedValue := int64(1690905599)
	actualTime, err := ParseFromLongDateLastTime("2023-08-01", 480)
	assert.Equal(t, nil, err)

	actualValue := actualTime.Unix()
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateTimeToMinUnixTime(t *testing.T) {
	expectedValue := int64(1690797600)
	actualTime, err := ParseFromLongDateTimeToMinUnixTime("2023-08-01 00:00:00")
	assert.Equal(t, nil, err)

	actualValue := actualTime.Unix()
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateTimeToMaxUnixTime(t *testing.T) {
	expectedValue := int64(1690891200)
	actualTime, err := ParseFromLongDateTimeToMaxUnixTime("2023-08-01 00:00:00")
	assert.Equal(t, nil, err)

	actualValue := actualTime.Unix()
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateTimeInFixedUtcOffset(t *testing.T) {
	expectedValue := int64(1617228083)
	actualTime, err := ParseFromLongDateTimeInFixedUtcOffset("2021-04-01 06:01:23", 480)
	assert.Equal(t, nil, err)

	actualValue := actualTime.Unix()
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateTimeInTimeZone(t *testing.T) {
	londonLocation, err := time.LoadLocation("Europe/London")
	assert.Equal(t, nil, err)

	// during standard time (UTC+0)
	expectedValue := int64(1577858483)
	actualTime, err := ParseFromLongDateTimeInTimeZone("2020-01-01 06:01:23", londonLocation)
	assert.Equal(t, nil, err)

	actualValue := actualTime.Unix()
	assert.Equal(t, expectedValue, actualValue)

	// during daylight saving time (UTC+1)
	expectedValue = int64(1619845283)
	actualTime, err = ParseFromLongDateTimeInTimeZone("2021-05-01 06:01:23", londonLocation)
	assert.Equal(t, nil, err)

	actualValue = actualTime.Unix()
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateTimeWithTimezone(t *testing.T) {
	expectedValue := int64(1617238883)
	actualTime, err := ParseFromLongDateTimeWithTimezone("2021-04-01 06:01:23+05:00")
	assert.Equal(t, nil, err)

	actualValue := actualTime.Unix()
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateTimeWithTimezone2(t *testing.T) {
	expectedValue := int64(1617238883)
	actualTime, err := ParseFromLongDateTimeWithTimezone2("2021-04-01 06:01:23 +0500")
	assert.Equal(t, nil, err)

	actualValue := actualTime.Unix()
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateTimeWithTimezoneRFC3339Format(t *testing.T) {
	expectedValue := int64(1617238883)
	actualTime, err := ParseFromLongDateTimeWithTimezoneRFC3339Format("2021-04-01T06:01:23+05:00")
	assert.Equal(t, nil, err)

	actualValue := actualTime.Unix()
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromLongDateTimeWithoutSecondInFixedUtcOffset(t *testing.T) {
	expectedValue := int64(1691947440)
	actualTime, err := ParseFromLongDateTimeWithoutSecondInFixedUtcOffset("2023-08-13 17:24", 0)
	assert.Equal(t, nil, err)

	actualValue := actualTime.Unix()
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromShortDateTimeInFixedUtcOffset(t *testing.T) {
	expectedValue := int64(1617228083)
	actualTime, err := ParseFromShortDateTimeInFixedUtcOffset("2021-4-1 6:1:23", 480)
	assert.Equal(t, nil, err)

	actualValue := actualTime.Unix()
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromElapsedSeconds(t *testing.T) {
	expectedValue := "00:00:00"
	actualValue, err := ParseFromElapsedSeconds(0)
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "00:00:09"
	actualValue, err = ParseFromElapsedSeconds(9)
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "00:01:08"
	actualValue, err = ParseFromElapsedSeconds(68)
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "01:00:07"
	actualValue, err = ParseFromElapsedSeconds(3607)
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedValue, actualValue)

	expectedValue = "23:59:59"
	actualValue, err = ParseFromElapsedSeconds(86399)
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedValue, actualValue)
}

func TestParseFromElapsedSeconds_InvalidTime(t *testing.T) {
	_, err := ParseFromElapsedSeconds(-1)
	assert.NotEqual(t, nil, err)

	_, err = ParseFromElapsedSeconds(86400)
	assert.NotEqual(t, nil, err)
}
