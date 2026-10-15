package utils

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"strings"
	"time"
)

// GetTimezoneOffsetMinutes returns offset minutes according specified timezone
func GetTimezoneOffsetMinutes(unixTime int64, timezone *time.Location) int16 {
	_, tzOffset := parseFromUnixTime(unixTime).In(timezone).Zone()
	tzMinuteOffset := int16(tzOffset / 60)

	return tzMinuteOffset
}

// GetMinimumTimezoneOffsetMinutes returns the minimum offset in the UTC year containing the Unix time
func GetMinimumTimezoneOffsetMinutes(unixTime int64, location *time.Location) int16 {
	year := time.Unix(unixTime, 0).UTC().Year()
	currentTime := time.Date(year, 1, 1, 0, 0, 0, 0, location)
	yearEndTime := time.Date(year+1, 1, 1, 0, 0, 0, 0, location)
	minimumOffset := GetTimezoneOffsetMinutes(currentTime.Unix(), location)

	for currentTime.Before(yearEndTime) {
		offset := GetTimezoneOffsetMinutes(currentTime.Unix(), location)

		if offset < minimumOffset {
			minimumOffset = offset
		}

		_, zoneEndTime := currentTime.ZoneBounds()

		if zoneEndTime.IsZero() || !zoneEndTime.Before(yearEndTime) {
			break
		}

		if !zoneEndTime.After(currentTime) {
			// POSIX timezone rules can return a non-advancing bound at the end of a leap year.
			currentTime = currentTime.Add(24 * time.Hour)
		} else {
			currentTime = zoneEndTime
		}
	}

	return minimumOffset
}

// GetScheduledTransactionScheduledAt returns the scheduled time using the minimum offset in the creation year
func GetScheduledTransactionScheduledAt(createdUnixTime int64, location *time.Location) int16 {
	referenceOffset := GetMinimumTimezoneOffsetMinutes(createdUnixTime, location)
	minutesElapsedOfDayInUtc := (24*60 - int(referenceOffset)) % (24 * 60)
	return int16(minutesElapsedOfDayInUtc)
}

// GetScheduledTransactionFinalTime returns the final time of scheduled transaction in given location
func GetScheduledTransactionFinalTime(todayFirstUnixTimeInUTC int64, scheduledAt int16, createdUnixTime int64, location *time.Location) *time.Time {
	referenceOffset := GetMinimumTimezoneOffsetMinutes(createdUnixTime, location)
	referenceTimezone := time.FixedZone("Template Timezone", int(referenceOffset)*60)

	expectedUnixTime := todayFirstUnixTimeInUTC + int64(scheduledAt)*60
	expectedTime := time.Unix(expectedUnixTime, 0).In(referenceTimezone)

	if expectedTime.Hour() == 0 && expectedTime.Minute() == 0 && expectedTime.Second() == 0 {
		firstTime := GetDateFirstTimeInTimezone(expectedTime, location)

		if firstTime.Year() != expectedTime.Year() || firstTime.Month() != expectedTime.Month() || firstTime.Day() != expectedTime.Day() {
			return nil
		}

		return &firstTime
	}

	finalTime := time.Date(expectedTime.Year(), expectedTime.Month(), expectedTime.Day(), expectedTime.Hour(), expectedTime.Minute(), expectedTime.Second(), 0, location)
	return &finalTime
}

// GetServerTimezoneOffsetMinutes returns offset minutes of current server timezone
func GetServerTimezoneOffsetMinutes() int16 {
	_, tzOffset := time.Now().Zone()
	tzMinuteOffset := int16(tzOffset / 60)

	return tzMinuteOffset
}

// FormatTimezoneOffset returns "+/-HH:MM" format of timezone
func FormatTimezoneOffset(unixTime int64, timezone *time.Location) string {
	tzMinutesOffset := GetTimezoneOffsetMinutes(unixTime, timezone)

	sign := "+"

	if tzMinutesOffset < 0 {
		sign = "-"
		tzMinutesOffset = -tzMinutesOffset
	}

	hourAbsOffset := tzMinutesOffset / 60
	minuteAbsOffset := tzMinutesOffset % 60

	return fmt.Sprintf("%s%02d:%02d", sign, hourAbsOffset, minuteAbsOffset)
}

// FormatTimezoneOffsetFromHoursOffset returns "+/-HH:MM" format of timezone from hours offset
func FormatTimezoneOffsetFromHoursOffset(hoursOffset string) (string, error) {
	hoursOffsetValue, err := StringToFloat64(hoursOffset)

	if err != nil {
		return "", errs.ErrFormatInvalid
	}

	tzMinutesOffset := int16(hoursOffsetValue * 60)

	sign := "+"

	if tzMinutesOffset < 0 {
		sign = "-"
		tzMinutesOffset = -tzMinutesOffset
	}

	hourAbsOffset := tzMinutesOffset / 60
	minuteAbsOffset := tzMinutesOffset % 60

	return fmt.Sprintf("%s%02d:%02d", sign, hourAbsOffset, minuteAbsOffset), nil
}

// ParseFromTimezoneOffset parses a formatted string in timezone offset format
func ParseFromTimezoneOffset(tzOffset string) (*time.Location, error) {
	if len(tzOffset) != 6 { // +/-HH:MM
		return nil, errs.ErrFormatInvalid
	}

	sign := tzOffset[0]

	if sign != '-' && sign != '+' {
		return nil, errs.ErrFormatInvalid
	}

	offsets := strings.Split(tzOffset[1:], ":")

	if len(offsets) != 2 {
		return nil, errs.ErrFormatInvalid
	}

	hourAbsOffset, err := StringToInt(offsets[0])

	if err != nil {
		return nil, err
	}

	minuteAbsOffset, err := StringToInt(offsets[1])

	if err != nil {
		return nil, err
	}

	totalMinuteOffset := hourAbsOffset*60 + minuteAbsOffset

	if sign == '-' {
		totalMinuteOffset = -totalMinuteOffset
	}

	totalOffset := totalMinuteOffset * 60
	return time.FixedZone("Timezone", totalOffset), nil
}
