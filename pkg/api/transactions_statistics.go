package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"sort"
)

// TransactionStatisticsHandler returns transaction statistics of current user
func (a *TransactionsApi) TransactionStatisticsHandler(c *core.WebContext) (any, *errs.Error) {
	var statisticReq models.TransactionStatisticRequest
	err := c.ShouldBindQuery(&statisticReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionStatisticsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionStatisticsHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	noTags := statisticReq.TagFilter == models.TransactionNoTagFilterValue
	var tagFilters []*models.TransactionTagFilter

	if !noTags {
		tagFilters, err = models.ParseTransactionTagFilter(statisticReq.TagFilter)

		if err != nil {
			log.Warnf(c, "[transactions.TransactionStatisticsHandler] parse transaction tag filters error, because %s", err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	uid := c.GetCurrentUid()
	totalAmounts, err := a.transactions.GetAccountsAndCategoriesTotalInflowAndOutflow(c, uid, statisticReq.StartTime, statisticReq.EndTime, tagFilters, noTags, statisticReq.Keyword, statisticReq.MatchMode, clientTimezone, statisticReq.UseTransactionTimezone)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionStatisticsHandler] failed to get accounts and categories total income and expense for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	statisticResp := &models.TransactionStatisticResponse{
		StartTime: statisticReq.StartTime,
		EndTime:   statisticReq.EndTime,
	}

	statisticResp.Items = make([]*models.TransactionStatisticResponseItem, len(totalAmounts))

	for i := 0; i < len(totalAmounts); i++ {
		totalAmountItem := totalAmounts[i]
		statisticResp.Items[i] = &models.TransactionStatisticResponseItem{
			CategoryId:  totalAmountItem.CategoryId,
			AccountId:   totalAmountItem.AccountId,
			TotalAmount: totalAmountItem.Amount.String(),
		}

		if totalAmountItem.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || totalAmountItem.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			statisticResp.Items[i].RelatedAccountId = totalAmountItem.RelatedAccountId
			statisticResp.Items[i].RelatedAccountType, _ = totalAmountItem.Type.ToTransactionRelatedAccountType()
		}
	}

	return statisticResp, nil
}

// TransactionStatisticsTrendsHandler returns transaction statistics trends of current user
func (a *TransactionsApi) TransactionStatisticsTrendsHandler(c *core.WebContext) (any, *errs.Error) {
	var statisticTrendsReq models.TransactionStatisticTrendsRequest
	err := c.ShouldBindQuery(&statisticTrendsReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionStatisticsTrendsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionStatisticsTrendsHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	startYear, startMonth, endYear, endMonth, err := statisticTrendsReq.GetNumericYearMonthRange()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionStatisticsTrendsHandler] cannot parse year month, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	noTags := statisticTrendsReq.TagFilter == models.TransactionNoTagFilterValue
	var tagFilters []*models.TransactionTagFilter

	if !noTags {
		tagFilters, err = models.ParseTransactionTagFilter(statisticTrendsReq.TagFilter)

		if err != nil {
			log.Warnf(c, "[transactions.TransactionStatisticsTrendsHandler] parse transaction tag filters error, because %s", err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	uid := c.GetCurrentUid()
	allMonthlyTotalAmounts, err := a.transactions.GetAccountsAndCategoriesMonthlyInflowAndOutflow(c, uid, startYear, startMonth, endYear, endMonth, tagFilters, noTags, statisticTrendsReq.Keyword, statisticTrendsReq.MatchMode, clientTimezone, statisticTrendsReq.UseTransactionTimezone)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionStatisticsTrendsHandler] failed to get accounts and categories total income and expense for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	statisticTrendsResp := make(models.TransactionStatisticTrendsResponseItemSlice, 0, len(allMonthlyTotalAmounts))

	for yearMonth, monthlyTotalAmounts := range allMonthlyTotalAmounts {
		monthlyStatisticResp := &models.TransactionStatisticTrendsResponseItem{
			Year:  yearMonth / 100,
			Month: yearMonth % 100,
			Items: make([]*models.TransactionStatisticResponseItem, len(monthlyTotalAmounts)),
		}

		for i := 0; i < len(monthlyTotalAmounts); i++ {
			totalAmountItem := monthlyTotalAmounts[i]
			monthlyStatisticResp.Items[i] = &models.TransactionStatisticResponseItem{
				CategoryId:  totalAmountItem.CategoryId,
				AccountId:   totalAmountItem.AccountId,
				TotalAmount: totalAmountItem.Amount.String(),
			}

			if totalAmountItem.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || totalAmountItem.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
				monthlyStatisticResp.Items[i].RelatedAccountId = totalAmountItem.RelatedAccountId
				monthlyStatisticResp.Items[i].RelatedAccountType, _ = totalAmountItem.Type.ToTransactionRelatedAccountType()
			}
		}

		statisticTrendsResp = append(statisticTrendsResp, monthlyStatisticResp)
	}

	sort.Sort(statisticTrendsResp)

	return statisticTrendsResp, nil
}

// TransactionStatisticsAssetTrendsHandler returns transaction statistics asset trends of current user
func (a *TransactionsApi) TransactionStatisticsAssetTrendsHandler(c *core.WebContext) (any, *errs.Error) {
	var statisticAssetTrendsReq models.TransactionStatisticAssetTrendsRequest
	err := c.ShouldBindQuery(&statisticAssetTrendsReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionStatisticsAssetTrendsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionStatisticsAssetTrendsHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()

	maxTransactionTime := int64(0)

	if statisticAssetTrendsReq.EndTime > 0 {
		maxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(statisticAssetTrendsReq.EndTime)
	}

	minTransactionTime := int64(0)

	if statisticAssetTrendsReq.StartTime > 0 {
		minTransactionTime = utils.GetMinTransactionTimeFromUnixTime(statisticAssetTrendsReq.StartTime)
	}

	accountDailyBalances, err := a.transactions.GetAllAccountsDailyOpeningAndClosingBalance(c, uid, maxTransactionTime, minTransactionTime, clientTimezone)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionStatisticsAssetTrendsHandler] failed to get transactions from \"%d\" to \"%d\" for user \"uid:%d\", because %s", statisticAssetTrendsReq.StartTime, statisticAssetTrendsReq.EndTime, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	statisticAssetTrendsResp := make(models.TransactionStatisticAssetTrendsResponseItemSlice, 0)

	for yearMonthDay, dailyAccountBalances := range accountDailyBalances {
		dailyStatisticResp := &models.TransactionStatisticAssetTrendsResponseItem{
			Year:  yearMonthDay / 10000,
			Month: (yearMonthDay % 10000) / 100,
			Day:   yearMonthDay % 100,
			Items: make([]*models.TransactionStatisticAssetTrendsResponseDataItem, len(dailyAccountBalances)),
		}

		for i := 0; i < len(dailyAccountBalances); i++ {
			accountBalance := dailyAccountBalances[i]
			dailyStatisticResp.Items[i] = &models.TransactionStatisticAssetTrendsResponseDataItem{
				AccountId:             accountBalance.AccountId,
				AccountOpeningBalance: accountBalance.AccountOpeningBalance.String(),
				AccountClosingBalance: accountBalance.AccountClosingBalance.String(),
			}
		}

		statisticAssetTrendsResp = append(statisticAssetTrendsResp, dailyStatisticResp)
	}

	sort.Sort(statisticAssetTrendsResp)

	return statisticAssetTrendsResp, nil
}
