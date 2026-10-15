package models

import (
	"strings"
)

// TransactionInfoResponseSlice represents the slice data structure of TransactionInfoResponse
type TransactionInfoResponseSlice []*TransactionInfoResponse

// Len returns the count of items
func (s TransactionInfoResponseSlice) Len() int {
	return len(s)
}

// Swap swaps two items
func (s TransactionInfoResponseSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item is less than the second one
func (s TransactionInfoResponseSlice) Less(i, j int) bool {
	if s[i].Time != s[j].Time {
		return s[i].Time > s[j].Time
	}

	return s[i].Id > s[j].Id
}

// TransactionStatisticTrendsResponseItemSlice represents the slice data structure of TransactionStatisticTrendsResponseItem
type TransactionStatisticTrendsResponseItemSlice []*TransactionStatisticTrendsResponseItem

// Len returns the count of items
func (s TransactionStatisticTrendsResponseItemSlice) Len() int {
	return len(s)
}

// Swap swaps two items
func (s TransactionStatisticTrendsResponseItemSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item is less than the second one
func (s TransactionStatisticTrendsResponseItemSlice) Less(i, j int) bool {
	if s[i].Year != s[j].Year {
		return s[i].Year < s[j].Year
	}

	return s[i].Month < s[j].Month
}

// TransactionStatisticAssetTrendsResponseItemSlice represents the slice data structure of TransactionStatisticAssetTrendsResponseItem
type TransactionStatisticAssetTrendsResponseItemSlice []*TransactionStatisticAssetTrendsResponseItem

// Len returns the count of items
func (s TransactionStatisticAssetTrendsResponseItemSlice) Len() int {
	return len(s)
}

// Swap swaps two items
func (s TransactionStatisticAssetTrendsResponseItemSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item is less than the second one
func (s TransactionStatisticAssetTrendsResponseItemSlice) Less(i, j int) bool {
	if s[i].Year != s[j].Year {
		return s[i].Year < s[j].Year
	}

	if s[i].Month != s[j].Month {
		return s[i].Month < s[j].Month
	}

	return s[i].Day < s[j].Day
}

// TransactionAmountsResponseItemAmountInfoSlice represents the slice data structure of TransactionAmountsResponseItemAmountInfo
type TransactionAmountsResponseItemAmountInfoSlice []*TransactionAmountsResponseItemAmountInfo

// Len returns the count of items
func (s TransactionAmountsResponseItemAmountInfoSlice) Len() int {
	return len(s)
}

// Swap swaps two items
func (s TransactionAmountsResponseItemAmountInfoSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item is less than the second one
func (s TransactionAmountsResponseItemAmountInfoSlice) Less(i, j int) bool {
	return strings.Compare(s[i].Currency, s[j].Currency) < 0
}
