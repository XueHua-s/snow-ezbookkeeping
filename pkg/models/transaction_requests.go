package models

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
)

// TransactionGeoLocationRequest represents all parameters of transaction geographic location info update request
type TransactionGeoLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

// TransactionCreateRequest represents all parameters of transaction creation request
type TransactionCreateRequest struct {
	Type                 TransactionType                `json:"type" binding:"required"`
	CategoryId           int64                          `json:"categoryId,string"`
	Time                 int64                          `json:"time" binding:"required,min=1"`
	UtcOffset            int16                          `json:"utcOffset" binding:"min=-720,max=840"`
	SourceAccountId      int64                          `json:"sourceAccountId,string" binding:"required,min=1"`
	DestinationAccountId int64                          `json:"destinationAccountId,string" binding:"min=0"`
	SourceAmount         int64                          `json:"sourceAmount" binding:"validTransactionAmount"`
	DestinationAmount    int64                          `json:"destinationAmount" binding:"validTransactionAmount"`
	HideAmount           bool                           `json:"hideAmount"`
	TagIds               []string                       `json:"tagIds"`
	PictureIds           []string                       `json:"pictureIds"`
	Comment              string                         `json:"comment" binding:"max=255"`
	GeoLocation          *TransactionGeoLocationRequest `json:"geoLocation" binding:"omitempty"`
	ClientSessionId      string                         `json:"clientSessionId"`
}

// TransactionModifyRequest represents all parameters of transaction modification request
type TransactionModifyRequest struct {
	Id                   int64                          `json:"id,string" binding:"required,min=1"`
	Type                 TransactionType                `json:"type" binding:"required"`
	CategoryId           int64                          `json:"categoryId,string"`
	Time                 int64                          `json:"time" binding:"required,min=1"`
	UtcOffset            int16                          `json:"utcOffset" binding:"min=-720,max=840"`
	SourceAccountId      int64                          `json:"sourceAccountId,string" binding:"required,min=1"`
	DestinationAccountId int64                          `json:"destinationAccountId,string" binding:"min=0"`
	SourceAmount         int64                          `json:"sourceAmount" binding:"validTransactionAmount"`
	DestinationAmount    int64                          `json:"destinationAmount" binding:"validTransactionAmount"`
	HideAmount           bool                           `json:"hideAmount"`
	TagIds               []string                       `json:"tagIds"`
	PictureIds           []string                       `json:"pictureIds"`
	Comment              string                         `json:"comment" binding:"max=255"`
	GeoLocation          *TransactionGeoLocationRequest `json:"geoLocation" binding:"omitempty"`
}

// TransactionImportRequest represents all parameters of transaction import request
type TransactionImportRequest struct {
	Transactions    []*TransactionCreateRequest `json:"transactions"`
	ClientSessionId string                      `json:"clientSessionId"`
}

// TransactionImportProcessRequest represents all parameters of transaction import process request
type TransactionImportProcessRequest struct {
	ClientSessionId string `form:"client_session_id"`
}

type TransactionTagFilter struct {
	TagIds []int64
	Type   TransactionTagFilterType
}

// TransactionCountRequest represents transaction count request
type TransactionCountRequest struct {
	Type             TransactionType `form:"type" binding:"min=0,max=4"`
	CategoryIds      string          `form:"category_ids"`
	AccountIds       string          `form:"account_ids"`
	TagFilter        string          `form:"tag_filter" binding:"validTagFilter"`
	AmountFilter     string          `form:"amount_filter" binding:"validAmountFilter"`
	Keyword          string          `form:"keyword"`
	MatchMode        core.MatchMode  `form:"match_mode" binding:"min=0,max=1"`
	MustHavePictures bool            `form:"must_have_pictures"`
	MaxTime          int64           `form:"max_time" binding:"min=0"` // Transaction time sequence id
	MinTime          int64           `form:"min_time" binding:"min=0"` // Transaction time sequence id
}

// TransactionListByMaxTimeRequest represents all parameters of transaction listing by max time request
type TransactionListByMaxTimeRequest struct {
	Type             TransactionType `form:"type" binding:"min=0,max=4"`
	CategoryIds      string          `form:"category_ids"`
	AccountIds       string          `form:"account_ids"`
	TagFilter        string          `form:"tag_filter" binding:"validTagFilter"`
	AmountFilter     string          `form:"amount_filter" binding:"validAmountFilter"`
	Keyword          string          `form:"keyword"`
	MatchMode        core.MatchMode  `form:"match_mode" binding:"min=0,max=1"`
	MustHavePictures bool            `form:"must_have_pictures"`
	MaxTime          int64           `form:"max_time" binding:"min=0"` // Transaction time sequence id
	MinTime          int64           `form:"min_time" binding:"min=0"` // Transaction time sequence id
	Page             int32           `form:"page" binding:"min=0"`
	Count            int32           `form:"count" binding:"required,min=1,max=50"`
	WithCount        bool            `form:"with_count"`
	WithPictures     bool            `form:"with_pictures"`
	TrimAccount      bool            `form:"trim_account"`
	TrimCategory     bool            `form:"trim_category"`
	TrimTag          bool            `form:"trim_tag"`
}

// TransactionListInMonthByPageRequest represents all parameters of transaction listing by month request
type TransactionListInMonthByPageRequest struct {
	Year             int32           `form:"year" binding:"required,min=1"`
	Month            int32           `form:"month" binding:"required,min=1"`
	Type             TransactionType `form:"type" binding:"min=0,max=4"`
	CategoryIds      string          `form:"category_ids"`
	AccountIds       string          `form:"account_ids"`
	TagFilter        string          `form:"tag_filter" binding:"validTagFilter"`
	AmountFilter     string          `form:"amount_filter" binding:"validAmountFilter"`
	Keyword          string          `form:"keyword"`
	MatchMode        core.MatchMode  `form:"match_mode" binding:"min=0,max=1"`
	MustHavePictures bool            `form:"must_have_pictures"`
	WithPictures     bool            `form:"with_pictures"`
	TrimAccount      bool            `form:"trim_account"`
	TrimCategory     bool            `form:"trim_category"`
	TrimTag          bool            `form:"trim_tag"`
}

// TransactionAllListRequest represents all parameters of all transaction listing request
type TransactionAllListRequest struct {
	Type             TransactionType `form:"type" binding:"min=0,max=4"`
	CategoryIds      string          `form:"category_ids"`
	AccountIds       string          `form:"account_ids"`
	TagFilter        string          `form:"tag_filter" binding:"validTagFilter"`
	AmountFilter     string          `form:"amount_filter" binding:"validAmountFilter"`
	Keyword          string          `form:"keyword"`
	MatchMode        core.MatchMode  `form:"match_mode" binding:"min=0,max=1"`
	MustHavePictures bool            `form:"must_have_pictures"`
	StartTime        int64           `form:"start_time" binding:"min=0"`
	EndTime          int64           `form:"end_time" binding:"min=0"`
	WithPictures     bool            `form:"with_pictures"`
	TrimAccount      bool            `form:"trim_account"`
	TrimCategory     bool            `form:"trim_category"`
	TrimTag          bool            `form:"trim_tag"`
}

// TransactionReconciliationStatementRequest represents all parameters of transaction reconciliation statement request
type TransactionReconciliationStatementRequest struct {
	AccountId int64 `form:"account_id,string" binding:"required,min=1"`
	StartTime int64 `form:"start_time"`
	EndTime   int64 `form:"end_time"`
}

// TransactionStatisticRequest represents all parameters of transaction statistic request
type TransactionStatisticRequest struct {
	StartTime              int64          `form:"start_time" binding:"min=0"`
	EndTime                int64          `form:"end_time" binding:"min=0"`
	TagFilter              string         `form:"tag_filter" binding:"validTagFilter"`
	Keyword                string         `form:"keyword"`
	MatchMode              core.MatchMode `form:"match_mode" binding:"min=0,max=1"`
	UseTransactionTimezone bool           `form:"use_transaction_timezone"`
}

// TransactionStatisticTrendsRequest represents all parameters of transaction statistic trends request
type TransactionStatisticTrendsRequest struct {
	YearMonthRangeRequest
	TagFilter              string         `form:"tag_filter" binding:"validTagFilter"`
	Keyword                string         `form:"keyword"`
	MatchMode              core.MatchMode `form:"match_mode" binding:"min=0,max=1"`
	UseTransactionTimezone bool           `form:"use_transaction_timezone"`
}

// TransactionStatisticAssetTrendsRequest represents all parameters of transaction statistic asset trends request
type TransactionStatisticAssetTrendsRequest struct {
	StartTime int64 `form:"start_time"`
	EndTime   int64 `form:"end_time"`
}

// TransactionAmountsRequest represents all parameters of transaction amounts request
type TransactionAmountsRequest struct {
	Query                  string `form:"query"`
	ExcludeAccountIds      string `form:"exclude_account_ids"`
	ExcludeCategoryIds     string `form:"exclude_category_ids"`
	TagFilter              string `form:"tag_filter"`
	UseTransactionTimezone bool   `form:"use_transaction_timezone"`
}

// TransactionDailyAmountsRequest represents all parameters of transaction daily amounts request
type TransactionDailyAmountsRequest struct {
	StartTime              int64  `form:"start_time" binding:"required,min=1"`
	EndTime                int64  `form:"end_time" binding:"required,min=1"`
	ExcludeAccountIds      string `form:"exclude_account_ids"`
	ExcludeCategoryIds     string `form:"exclude_category_ids"`
	TagFilter              string `form:"tag_filter"`
	UseTransactionTimezone bool   `form:"use_transaction_timezone"`
}

// TransactionAmountsRequestItem represents an item of transaction amounts request
type TransactionAmountsRequestItem struct {
	Name      string
	StartTime int64
	EndTime   int64
}

// TransactionGetRequest represents all parameters of transaction getting request
type TransactionGetRequest struct {
	Id           int64 `form:"id,string" binding:"required,min=1"`
	WithPictures bool  `form:"with_pictures"`
	TrimAccount  bool  `form:"trim_account"`
	TrimCategory bool  `form:"trim_category"`
	TrimTag      bool  `form:"trim_tag"`
}

// TransactionBatchUpdateCategoryRequest represents all parameters of transaction batch update category request
type TransactionBatchUpdateCategoryRequest struct {
	TransactionIds []string `json:"transactionIds" binding:"required"`
	CategoryId     int64    `json:"categoryId,string" binding:"required"`
}

// TransactionBatchUpdateAccountRequest represents all parameters of transaction batch update account request
type TransactionBatchUpdateAccountRequest struct {
	TransactionIds       []string `json:"transactionIds" binding:"required"`
	AccountId            int64    `json:"accountId,string" binding:"required"`
	IsDestinationAccount bool     `json:"isDestinationAccount"`
}

// TransactionBatchAddTagsRequest represents all parameters of transaction batch add tags request
type TransactionBatchAddTagsRequest struct {
	TransactionIds []string `json:"transactionIds" binding:"required"`
	TagIds         []string `json:"tagIds" binding:"required"`
}

// TransactionBatchRemoveTagsRequest represents all parameters of transaction batch remove tags request
type TransactionBatchRemoveTagsRequest struct {
	TransactionIds []string `json:"transactionIds" binding:"required"`
	TagIds         []string `json:"tagIds" binding:"required"`
}

// TransactionBatchClearTagsRequest represents all parameters of transaction batch clear tags request
type TransactionBatchClearTagsRequest struct {
	TransactionIds []string `json:"transactionIds" binding:"required"`
}

// TransactionMoveBetweenAccountsRequest represents all parameters of moving all transactions between accounts request
type TransactionMoveBetweenAccountsRequest struct {
	FromAccountId int64 `json:"fromAccountId,string" binding:"required,min=1"`
	ToAccountId   int64 `json:"toAccountId,string" binding:"required,min=1"`
}

// TransactionDeleteRequest represents all parameters of transaction deleting request
type TransactionDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// TransactionBatchDeleteRequest represents all parameters of transaction batch deleting request
type TransactionBatchDeleteRequest struct {
	Ids      []string `json:"ids,string" binding:"required"`
	Password string   `json:"password" binding:"omitempty,min=6,max=128"`
}

// YearMonthRangeRequest represents all parameters of a request with year and month range
type YearMonthRangeRequest struct {
	StartYearMonth string `form:"start_year_month"`
	EndYearMonth   string `form:"end_year_month"`
}
