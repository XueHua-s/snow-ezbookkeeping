package utils

import (
	"fmt"
	"time"
)

// GetMinTransactionTimeFromUnixTime returns the minimum transaction time from unix time
func GetMinTransactionTimeFromUnixTime(unixTime int64) int64 {
	return unixTime * 1000
}

// GetMaxTransactionTimeFromUnixTime returns the maximum transaction time from unix time
func GetMaxTransactionTimeFromUnixTime(unixTime int64) int64 {
	return unixTime*1000 + 999
}

// GetUnixTimeFromTransactionTime returns unix time from the transaction time
func GetUnixTimeFromTransactionTime(transactionTime int64) int64 {
	return transactionTime / 1000
}

// GetTransactionTimeRangeByYearMonth returns the transaction time range by specified year and month
func GetTransactionTimeRangeByYearMonth(year int32, month int32) (int64, int64, error) {
	startMinUnixTime, err := ParseFromLongDateTimeToMinUnixTime(fmt.Sprintf("%d-%02d-01 00:00:00", year, month))
	startMaxUnixTime, err := ParseFromLongDateTimeToMaxUnixTime(fmt.Sprintf("%d-%02d-01 00:00:00", year, month))

	if err != nil {
		return 0, 0, err
	}

	endMaxUnixTime := startMaxUnixTime.AddDate(0, 1, 0)

	minTransactionTime := GetMinTransactionTimeFromUnixTime(startMinUnixTime.Unix())
	maxTransactionTime := GetMinTransactionTimeFromUnixTime(endMaxUnixTime.Unix()) - 1

	return minTransactionTime, maxTransactionTime, nil
}

// GetStartOfDay returns the start time of the day for the specified time
func GetStartOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// parseFromUnixTime parses a unix time and returns a golang time struct
func parseFromUnixTime(unixTime int64) time.Time {
	return time.Unix(unixTime, 0)
}
