package models

import (
	"math/big"
)

// TransactionGeoLocationResponse represents a view-object of transaction geographic location info
type TransactionGeoLocationResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// TransactionInfoResponse represents a view-object of transaction
type TransactionInfoResponse struct {
	Id                   int64                                    `json:"id,string"`
	TimeSequenceId       int64                                    `json:"timeSequenceId,string"`
	Type                 TransactionType                          `json:"type"`
	CategoryId           int64                                    `json:"categoryId,string"`
	Category             *TransactionCategoryInfoResponse         `json:"category,omitempty"`
	Time                 int64                                    `json:"time"`
	UtcOffset            int16                                    `json:"utcOffset"`
	SourceAccountId      int64                                    `json:"sourceAccountId,string"`
	SourceAccount        *AccountInfoResponse                     `json:"sourceAccount,omitempty"`
	DestinationAccountId int64                                    `json:"destinationAccountId,string,omitempty"`
	DestinationAccount   *AccountInfoResponse                     `json:"destinationAccount,omitempty"`
	SourceAmount         int64                                    `json:"sourceAmount"`
	DestinationAmount    *int64                                   `json:"destinationAmount,omitempty"`
	HideAmount           bool                                     `json:"hideAmount"`
	TagIds               []string                                 `json:"tagIds"`
	Tags                 []*TransactionTagInfoResponse            `json:"tags,omitempty"`
	Pictures             TransactionPictureInfoBasicResponseSlice `json:"pictures,omitempty"`
	Comment              string                                   `json:"comment"`
	GeoLocation          *TransactionGeoLocationResponse          `json:"geoLocation,omitempty"`
	Editable             bool                                     `json:"editable"`
}

// TransactionCountResponse represents transaction count response
type TransactionCountResponse struct {
	TotalCount int64 `json:"totalCount"`
}

// TransactionInfoPageWrapperResponse represents a response of transaction which contains items and next id
type TransactionInfoPageWrapperResponse struct {
	Items              TransactionInfoResponseSlice `json:"items"`
	NextTimeSequenceId *int64                       `json:"nextTimeSequenceId,string"`
	TotalCount         *int64                       `json:"totalCount,omitempty"`
}

// TransactionInfoPageWrapperResponse2 represents a response of transaction which contains items and count
type TransactionInfoPageWrapperResponse2 struct {
	Items      TransactionInfoResponseSlice `json:"items"`
	TotalCount int64                        `json:"totalCount"`
}

// TransactionReconciliationStatementResponseItem represents a transaction reconciliation statement response
type TransactionReconciliationStatementResponseItem struct {
	*TransactionInfoResponse
	AccountOpeningBalance string `json:"accountOpeningBalance"`
	AccountClosingBalance string `json:"accountClosingBalance"`
}

// TransactionReconciliationStatementResponse represents the response of all transaction reconciliation statement response
type TransactionReconciliationStatementResponse struct {
	Transactions   []*TransactionReconciliationStatementResponseItem `json:"transactions"`
	TotalInflows   string                                            `json:"totalInflows"`
	TotalOutflows  string                                            `json:"totalOutflows"`
	OpeningBalance string                                            `json:"openingBalance"`
	ClosingBalance string                                            `json:"closingBalance"`
}

// TransactionUnreconciledCountRequest represents the request for unreconciled transaction counts
type TransactionUnreconciledCountRequest struct {
	AccountIds string `form:"account_ids"`
}

// TransactionUnreconciledCountItem represents the unreconciled transaction count of an account
type TransactionUnreconciledCountItem struct {
	AccountId int64 `json:"accountId,string"`
	Count     int64 `json:"count"`
}

// TransactionUnreconciledCountResponse represents the response of unreconciled transaction counts
type TransactionUnreconciledCountResponse struct {
	Items []*TransactionUnreconciledCountItem `json:"items"`
}

// TransactionStatisticResponse represents transaction statistic response
type TransactionStatisticResponse struct {
	StartTime int64                               `json:"startTime"`
	EndTime   int64                               `json:"endTime"`
	Items     []*TransactionStatisticResponseItem `json:"items"`
}

// TransactionStatisticResponseItem represents total amount item for a response
type TransactionStatisticResponseItem struct {
	CategoryId         int64                         `json:"categoryId,string"`
	AccountId          int64                         `json:"accountId,string"`
	RelatedAccountId   int64                         `json:"relatedAccountId,string,omitempty"`
	RelatedAccountType TransactionRelatedAccountType `json:"relatedAccountType,omitempty"`
	TotalAmount        string                        `json:"amount"`
}

// TransactionStatisticTrendsResponseItem represents the data within each statistic interval
type TransactionStatisticTrendsResponseItem struct {
	Year  int32                               `json:"year"`
	Month int32                               `json:"month"`
	Items []*TransactionStatisticResponseItem `json:"items"`
}

// TransactionStatisticAssetTrendsResponseItem represents the data within each statistic interval
type TransactionStatisticAssetTrendsResponseItem struct {
	Year  int32                                              `json:"year"`
	Month int32                                              `json:"month"`
	Day   int32                                              `json:"day"`
	Items []*TransactionStatisticAssetTrendsResponseDataItem `json:"items"`
}

// TransactionStatisticAssetTrendsResponseDataItem represents an asset trends data item
type TransactionStatisticAssetTrendsResponseDataItem struct {
	AccountId             int64  `json:"accountId,string"`
	AccountOpeningBalance string `json:"accountOpeningBalance"`
	AccountClosingBalance string `json:"accountClosingBalance"`
}

// TransactionAmountsResponseItem represents an item of transaction amounts
type TransactionAmountsResponseItem struct {
	StartTime int64                                       `json:"startTime"`
	EndTime   int64                                       `json:"endTime"`
	Amounts   []*TransactionAmountsResponseItemAmountInfo `json:"amounts"`
}

// TransactionDailyAmountsResponseItem represents an item of transaction daily amounts
type TransactionDailyAmountsResponseItem struct {
	Date    string                                      `json:"date"`
	Amounts []*TransactionAmountsResponseItemAmountInfo `json:"amounts"`
}

// TransactionMonthAmountsResponseItem represents an item of transaction month amounts
type TransactionMonthAmountsResponseItem struct {
	Year    int32                                       `json:"year"`
	Month   int32                                       `json:"month"`
	Amounts []*TransactionAmountsResponseItemAmountInfo `json:"amounts"`
}

// TransactionAmountsResponseItemAmountInfo represents amount info for a response item
type TransactionAmountsResponseItemAmountInfo struct {
	Currency      string `json:"currency"`
	IncomeAmount  string `json:"incomeAmount"`
	ExpenseAmount string `json:"expenseAmount"`
}

// TransactionTotalAmount represents an aggregated transaction amount
type TransactionTotalAmount struct {
	Type             TransactionDbType
	CategoryId       int64
	AccountId        int64
	RelatedAccountId int64
	Amount           *big.Int
}

// TransactionAmountsAndCurrency represents income and expense amounts with currency
type TransactionAmountsAndCurrency struct {
	Currency      string
	IncomeAmount  *big.Int
	ExpenseAmount *big.Int
}
