package utils

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"strings"
	"time"
)

const (
	longDateFormat                        = "2006-01-02"
	longDateTimeFormat                    = "2006-01-02 15:04:05"
	longDateTimeWithTimezoneFormat        = "2006-01-02 15:04:05Z07:00"
	longDateTimeWithTimezoneFormat2       = "2006-01-02 15:04:05 Z0700"
	longDateTimeWithTimezoneRFC3339Format = "2006-01-02T15:04:05Z07:00"
	longDateTimeWithoutSecondFormat       = "2006-01-02 15:04"
	shortDateTimeFormat                   = "2006-1-2 15:4:5"
	yearMonthDateTimeFormat               = "2006-01"
	westernmostTimezoneUtcOffset          = -720 // Etc/GMT+12 (UTC-12:00)
	easternmostTimezoneUtcOffset          = 840  // Pacific/Kiritimati (UTC+14:00)
)

// ParseNumericYearMonth returns numeric year and month from textual content
func ParseNumericYearMonth(yearMonth string) (int32, int32, error) {
	yearMonthParts := strings.Split(yearMonth, "-")

	if len(yearMonthParts) != 2 {
		return 0, 0, errs.ErrParameterInvalid
	}

	year, err := StringToInt32(yearMonthParts[0])

	if err != nil {
		return 0, 0, err
	}

	month, err := StringToInt32(yearMonthParts[1])

	if err != nil {
		return 0, 0, err
	}

	return year, month, nil
}

// FormatUnixTimeToLongDate returns a textual representation of the unix time formatted by long date time format
func FormatUnixTimeToLongDate(unixTime int64, timezone *time.Location) string {
	t := parseFromUnixTime(unixTime)

	if timezone != nil {
		t = t.In(timezone)
	}

	return t.Format(longDateFormat)
}

// FormatUnixTimeToLongDateTime returns a textual representation of the unix time formatted by long date time format
func FormatUnixTimeToLongDateTime(unixTime int64, timezone *time.Location) string {
	t := parseFromUnixTime(unixTime)

	if timezone != nil {
		t = t.In(timezone)
	}

	return t.Format(longDateTimeFormat)
}

// FormatUnixTimeToLongDateTimeWithTimezone returns a textual representation of the unix time formatted by long date time with timezone format
func FormatUnixTimeToLongDateTimeWithTimezone(unixTime int64, timezone *time.Location) string {
	t := parseFromUnixTime(unixTime)

	if timezone != nil {
		t = t.In(timezone)
	}

	return t.Format(longDateTimeWithTimezoneFormat)
}

// FormatUnixTimeToLongDateTimeWithTimezoneRFC3339Format returns a textual representation of the unix time formatted by long date time with timezone RFC 3339 format
func FormatUnixTimeToLongDateTimeWithTimezoneRFC3339Format(unixTime int64, timezone *time.Location) string {
	t := parseFromUnixTime(unixTime)

	if timezone != nil {
		t = t.In(timezone)
	}

	return t.Format(longDateTimeWithTimezoneRFC3339Format)
}

func FormatYearMonthDayToLongDateTime(year string, month string, day string) (string, error) {
	if len(year) == 2 {
		yearLast2Digits, err := StringToInt(year)

		if err != nil {
			return "", err
		}

		currentYear := time.Now().Year()
		currentYearLast2Digits := currentYear % 100

		if yearLast2Digits <= currentYearLast2Digits {
			year = IntToString(currentYear/100) + year
		} else {
			year = IntToString(currentYear/100-1) + year
		}
	}

	if len(month) < 2 {
		month = "0" + month
	}

	if len(day) < 2 {
		day = "0" + day
	}

	return fmt.Sprintf("%s-%s-%s 00:00:00", year, month, day), nil
}

// FormatUnixTimeToLongDateTimeInServerTimezone returns a textual representation of the unix time formatted by long date time format
func FormatUnixTimeToLongDateTimeInServerTimezone(unixTime int64) string {
	return parseFromUnixTime(unixTime).Format(longDateTimeFormat)
}

// FormatUnixTimeToLongDateTimeWithoutSecond returns a textual representation of the unix time formatted by long date time format (no second)
func FormatUnixTimeToLongDateTimeWithoutSecond(unixTime int64, timezone *time.Location) string {
	t := parseFromUnixTime(unixTime)

	if timezone != nil {
		t = t.In(timezone)
	}

	return t.Format(longDateTimeWithoutSecondFormat)
}

// FormatUnixTimeToYearMonth returns year and month of specified unix time
func FormatUnixTimeToYearMonth(unixTime int64, timezone *time.Location) string {
	t := parseFromUnixTime(unixTime)

	if timezone != nil {
		t = t.In(timezone)
	}

	return t.Format(yearMonthDateTimeFormat)
}

// FormatUnixTimeToNumericYearMonth returns numeric year and month of specified unix time
func FormatUnixTimeToNumericYearMonth(unixTime int64, timezone *time.Location) int32 {
	t := parseFromUnixTime(unixTime)

	if timezone != nil {
		t = t.In(timezone)
	}

	return int32(t.Year())*100 + int32(t.Month())
}

// FormatUnixTimeToNumericYearMonthDay returns numeric year, month and day of specified unix time
func FormatUnixTimeToNumericYearMonthDay(unixTime int64, timezone *time.Location) int32 {
	t := parseFromUnixTime(unixTime)

	if timezone != nil {
		t = t.In(timezone)
	}

	return int32(t.Year())*10000 + int32(t.Month())*100 + int32(t.Day())
}

// FormatUnixTimeToNumericLocalDateTime returns numeric year, month, day, hour, minute and second of specified unix time
func FormatUnixTimeToNumericLocalDateTime(unixTime int64, timezone *time.Location) int64 {
	t := parseFromUnixTime(unixTime)

	if timezone != nil {
		t = t.In(timezone)
	}

	localDateTime := int64(t.Year())
	localDateTime = localDateTime*100 + int64(t.Month())
	localDateTime = localDateTime*100 + int64(t.Day())
	localDateTime = localDateTime*100 + int64(t.Hour())
	localDateTime = localDateTime*100 + int64(t.Minute())
	localDateTime = localDateTime*100 + int64(t.Second())

	return localDateTime
}

// FormatNumericYearMonthDayToLongDate returns a textual representation of the numeric year, month and day formatted by long date format
func FormatNumericYearMonthDayToLongDate(yearMonthDay int32) string {
	year := yearMonthDay / 10000
	month := (yearMonthDay % 10000) / 100
	day := yearMonthDay % 100

	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}
