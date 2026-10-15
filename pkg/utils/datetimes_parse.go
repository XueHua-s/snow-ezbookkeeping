package utils

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"time"
)

// GetMinUnixTimeWithSameLocalDateTime returns the minimum UnixTime for date with the same local date
func GetMinUnixTimeWithSameLocalDateTime(unixTime int64, currentUtcOffset int16) int64 {
	return unixTime + int64(currentUtcOffset)*60 - easternmostTimezoneUtcOffset*60
}

// GetMaxUnixTimeWithSameLocalDateTime returns the maximum UnixTime for date with the same local date
func GetMaxUnixTimeWithSameLocalDateTime(unixTime int64, currentUtcOffset int16) int64 {
	return unixTime + int64(currentUtcOffset)*60 - westernmostTimezoneUtcOffset*60
}

// GetDateFirstTimeInTimezone returns the first valid instant at or after a local date's beginning.
func GetDateFirstTimeInTimezone(date time.Time, timezone *time.Location) time.Time {
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	currentTime := dateOnly.Add(-24 * time.Hour).In(timezone)

	for {
		_, offset := currentTime.Zone()
		firstTime := dateOnly.Add(-time.Duration(offset) * time.Second)

		if firstTime.Before(currentTime) {
			firstTime = currentTime
		}

		_, zoneEndTime := currentTime.ZoneBounds()

		if zoneEndTime.IsZero() || !zoneEndTime.After(currentTime) || firstTime.Before(zoneEndTime) {
			return firstTime.In(timezone)
		}

		currentTime = zoneEndTime.In(timezone)
	}
}

// ParseFromLongDateFirstTimeInTimezone parses a local date using timezone rules
func ParseFromLongDateFirstTimeInTimezone(t string, timezone *time.Location) (time.Time, error) {
	date, err := time.Parse(longDateFormat, t)

	if err != nil {
		return date, err
	}

	firstTime := GetDateFirstTimeInTimezone(date, timezone)

	if firstTime.Year() != date.Year() || firstTime.Month() != date.Month() || firstTime.Day() != date.Day() {
		return time.Time{}, errs.ErrParameterInvalid
	}

	return firstTime, nil
}

// ParseFromLongDateLastTimeInTimezone returns the last second of a local date
func ParseFromLongDateLastTimeInTimezone(t string, timezone *time.Location) (time.Time, error) {
	firstTime, err := ParseFromLongDateFirstTimeInTimezone(t, timezone)

	if err != nil {
		return firstTime, err
	}

	nextDate := time.Date(firstTime.Year(), firstTime.Month(), firstTime.Day()+1, 0, 0, 0, 0, time.UTC)
	return GetDateFirstTimeInTimezone(nextDate, timezone).Add(-1 * time.Nanosecond), nil
}

// ParseFromLongDateFirstTime parses a formatted string in long date format
func ParseFromLongDateFirstTime(t string, utcOffset int16) (time.Time, error) {
	timezone := time.FixedZone("Timezone", int(utcOffset)*60)
	return time.ParseInLocation(longDateFormat, t, timezone)
}

// ParseFromLongDateLastTime parses a formatted string in long date format
func ParseFromLongDateLastTime(t string, utcOffset int16) (time.Time, error) {
	timezone := time.FixedZone("Timezone", int(utcOffset)*60)
	lastTime, err := time.ParseInLocation(longDateFormat, t, timezone)

	if err != nil {
		return lastTime, err
	}

	return lastTime.Add(24 * time.Hour).Add(-1 * time.Nanosecond), nil
}

// ParseFromLongDateTimeToMinUnixTime parses a formatted string in long date time format to minimal unix time (the westernmost timezone)
func ParseFromLongDateTimeToMinUnixTime(t string) (time.Time, error) {
	timezone := time.FixedZone("Timezone", easternmostTimezoneUtcOffset*60)
	return time.ParseInLocation(longDateTimeFormat, t, timezone)
}

// ParseFromLongDateTimeToMaxUnixTime parses a formatted string in long date time format to maximal unix time (the easternmost timezone)
func ParseFromLongDateTimeToMaxUnixTime(t string) (time.Time, error) {
	timezone := time.FixedZone("Timezone", westernmostTimezoneUtcOffset*60)
	return time.ParseInLocation(longDateTimeFormat, t, timezone)
}

// ParseFromLongDateTimeInFixedUtcOffset parses a formatted string in long date time format
func ParseFromLongDateTimeInFixedUtcOffset(t string, utcOffset int16) (time.Time, error) {
	timezone := time.FixedZone("Timezone", int(utcOffset)*60)
	return time.ParseInLocation(longDateTimeFormat, t, timezone)
}

// ParseFromLongDateTimeInTimeZone parses a formatted string in long date time format
func ParseFromLongDateTimeInTimeZone(t string, timezone *time.Location) (time.Time, error) {
	return time.ParseInLocation(longDateTimeFormat, t, timezone)
}

// ParseFromLongDateTimeWithTimezone parses a formatted string in long date time format
func ParseFromLongDateTimeWithTimezone(t string) (time.Time, error) {
	return time.Parse(longDateTimeWithTimezoneFormat, t)
}

// ParseFromLongDateTimeWithTimezone2 parses a formatted string in long date time format
func ParseFromLongDateTimeWithTimezone2(t string) (time.Time, error) {
	return time.Parse(longDateTimeWithTimezoneFormat2, t)
}

// ParseFromLongDateTimeWithTimezoneRFC3339Format parses a formatted string in long date time RFC 3378 format
func ParseFromLongDateTimeWithTimezoneRFC3339Format(t string) (time.Time, error) {
	return time.Parse(longDateTimeWithTimezoneRFC3339Format, t)
}

// ParseFromLongDateTimeWithoutSecondInFixedUtcOffset parses a formatted string in long date time format (no second) with fixed UTC offset
func ParseFromLongDateTimeWithoutSecondInFixedUtcOffset(t string, utcOffset int16) (time.Time, error) {
	timezone := time.FixedZone("Timezone", int(utcOffset)*60)
	return time.ParseInLocation(longDateTimeWithoutSecondFormat, t, timezone)
}

// ParseFromShortDateTimeInFixedUtcOffset parses a formatted string in short date time format with fixed UTC offset
func ParseFromShortDateTimeInFixedUtcOffset(t string, utcOffset int16) (time.Time, error) {
	timezone := time.FixedZone("Timezone", int(utcOffset)*60)
	return time.ParseInLocation(shortDateTimeFormat, t, timezone)
}

func ParseFromElapsedSeconds(elapsedSeconds int) (string, error) {
	if elapsedSeconds < 0 || elapsedSeconds >= 86400 {
		return "", errs.ErrFormatInvalid
	}

	second := elapsedSeconds % 60
	elapsedSeconds = elapsedSeconds - second
	elapsedSeconds = elapsedSeconds / 60

	minute := elapsedSeconds % 60
	elapsedSeconds = elapsedSeconds - minute
	elapsedSeconds = elapsedSeconds / 60

	hour := elapsedSeconds

	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second), nil
}

// IsUnixTimeEqualsYearAndMonth returns whether year and month of the unix time are equals to the specified year and month
func IsUnixTimeEqualsYearAndMonth(unixTime int64, timezone *time.Location, year int32, month int32) bool {
	date := parseFromUnixTime(unixTime).In(timezone)
	return date.Year() == int(year) && int(date.Month()) == int(month)
}

// GetMaxDayOfMonth returns the maximum day of the month for the specified year and month
func GetMaxDayOfMonth(year int, month time.Month) int {
	t := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)
	return t.Day()
}

// GetDaysBetweenDates returns the signed calendar day difference, using each time's local date
func GetDaysBetweenDates(startDate, endDate time.Time) int {
	startDateOnly := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endDateOnly := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)
	return int((endDateOnly.Unix() - startDateOnly.Unix()) / (24 * 60 * 60))
}
