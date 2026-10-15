package models

import (
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"strings"
	"time"
)

// ParseTransactionTagFilter parses transaction tag filter from string
func ParseTransactionTagFilter(tagFilterStr string) ([]*TransactionTagFilter, error) {
	if tagFilterStr == "" || tagFilterStr == TransactionNoTagFilterValue {
		return []*TransactionTagFilter{}, nil
	}

	filters := strings.Split(tagFilterStr, ";")
	transactionTagFilters := make([]*TransactionTagFilter, 0, len(filters))

	for _, filter := range filters {
		tagFilterItem := strings.Split(filter, ":")

		if len(tagFilterItem) != 2 {
			return nil, errs.ErrFormatInvalid
		}

		tagFilterType, err := utils.StringToInt(tagFilterItem[0])

		if err != nil || (tagFilterType < int(TRANSACTION_TAG_FILTER_HAS_ANY) || tagFilterType > int(TRANSACTION_TAG_FILTER_NOT_HAS_ALL)) {
			return nil, errs.ErrFormatInvalid
		}

		textualTagIds := strings.Split(tagFilterItem[1], ",")
		tagIds := make([]int64, 0, len(textualTagIds))

		for _, tagIdStr := range textualTagIds {
			tagId, err := utils.StringToInt64(tagIdStr)

			if err != nil {
				return nil, errs.ErrTransactionTagIdInvalid
			}

			tagIds = append(tagIds, tagId)
		}

		transactionTagFilter := &TransactionTagFilter{
			TagIds: tagIds,
			Type:   TransactionTagFilterType(tagFilterType),
		}

		transactionTagFilters = append(transactionTagFilters, transactionTagFilter)
	}

	return transactionTagFilters, nil
}

// IsEditable returns whether this transaction can be edited
func (t *Transaction) IsEditable(currentUser *User, clientTimezone *time.Location, account *Account, relatedAccount *Account) bool {
	if account == nil || account.Hidden {
		return false
	}

	if t.Type == TRANSACTION_DB_TYPE_TRANSFER_OUT {
		if relatedAccount == nil || relatedAccount.Hidden {
			return false
		}
	}

	if currentUser == nil || !currentUser.CanEditTransactionByTransactionTime(t.TransactionTime, clientTimezone, account, relatedAccount) {
		return false
	}

	return true
}

// ToTransactionInfoResponse returns a view-object according to database model
func (t *Transaction) ToTransactionInfoResponse(tagIds []int64, editable bool) *TransactionInfoResponse {
	transactionType, err := t.Type.ToTransactionType()

	if err != nil {
		return nil
	}

	sourceAccountId := t.AccountId
	sourceAmount := t.Amount

	destinationAccountId := int64(0)
	var destinationAmount *int64

	if t.Type == TRANSACTION_DB_TYPE_TRANSFER_OUT {
		destinationAccountId = t.RelatedAccountId
		destinationAmount = &t.RelatedAccountAmount
	} else if t.Type == TRANSACTION_DB_TYPE_TRANSFER_IN {
		sourceAccountId = t.RelatedAccountId
		sourceAmount = t.RelatedAccountAmount

		destinationAccountId = t.AccountId
		destinationAmount = &t.Amount
	}

	geoLocation := &TransactionGeoLocationResponse{}

	if t.GeoLongitude != 0 || t.GeoLatitude != 0 {
		geoLocation.Longitude = t.GeoLongitude
		geoLocation.Latitude = t.GeoLatitude
	} else {
		geoLocation = nil
	}

	return &TransactionInfoResponse{
		LastModifiedUnixTime: t.UpdatedUnixTime,
		Id:                   t.TransactionId,
		TimeSequenceId:       t.TransactionTime,
		Type:                 transactionType,
		CategoryId:           t.CategoryId,
		Time:                 utils.GetUnixTimeFromTransactionTime(t.TransactionTime),
		UtcOffset:            t.TimezoneUtcOffset,
		SourceAccountId:      sourceAccountId,
		DestinationAccountId: destinationAccountId,
		SourceAmount:         sourceAmount,
		DestinationAmount:    destinationAmount,
		HideAmount:           t.HideAmount,
		TagIds:               utils.Int64ArrayToStringArray(tagIds),
		Comment:              t.Comment,
		GeoLocation:          geoLocation,
		Editable:             editable,
	}
}

// GetTransactionAmountsRequestItems returns request items by query parameters
func (t *TransactionAmountsRequest) GetTransactionAmountsRequestItems() ([]*TransactionAmountsRequestItem, error) {
	items := strings.Split(t.Query, "|")
	requestItems := make([]*TransactionAmountsRequestItem, 0, len(items))

	for i := 0; i < len(items); i++ {
		itemValues := strings.Split(items[i], "_")

		if len(itemValues) != 3 {
			return nil, errs.ErrQueryItemsInvalid
		}

		startTime, err := utils.StringToInt64(itemValues[1])

		if err != nil {
			return nil, err
		}

		endTime, err := utils.StringToInt64(itemValues[2])

		if err != nil {
			return nil, err
		}

		requestItem := &TransactionAmountsRequestItem{
			Name:      itemValues[0],
			StartTime: startTime,
			EndTime:   endTime,
		}

		requestItems = append(requestItems, requestItem)
	}

	return requestItems, nil
}

// GetNumericYearMonthRange returns numeric start year, start month, end year and end month
func (t *YearMonthRangeRequest) GetNumericYearMonthRange() (int32, int32, int32, int32, error) {
	var startYear, startMonth, endYear, endMonth int32
	var err error

	if t.StartYearMonth != "" {
		startYear, startMonth, err = utils.ParseNumericYearMonth(t.StartYearMonth)

		if err != nil {
			return 0, 0, 0, 0, err
		}
	}

	if t.EndYearMonth != "" {
		endYear, endMonth, err = utils.ParseNumericYearMonth(t.EndYearMonth)

		if err != nil {
			return 0, 0, 0, 0, err
		}
	}

	return startYear, startMonth, endYear, endMonth, nil
}
