package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"math/big"
	"slices"
	"sort"
	"strings"
)

// TransactionAmountsHandler returns transaction amounts of current user
func (a *TransactionsApi) TransactionAmountsHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionAmountsReq models.TransactionAmountsRequest
	err := c.ShouldBindQuery(&transactionAmountsReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionAmountsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	requestItems, err := transactionAmountsReq.GetTransactionAmountsRequestItems()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionAmountsHandler] get request item failed, because %s", err.Error())
		return nil, errs.ErrQueryItemsInvalid
	}

	if len(requestItems) < 1 {
		log.Warnf(c, "[transactions.TransactionAmountsHandler] parse request failed, because there are no valid items")
		return nil, errs.ErrQueryItemsEmpty
	}

	if len(requestItems) > 20 {
		log.Warnf(c, "[transactions.TransactionAmountsHandler] parse request failed, because there are too many items")
		return nil, errs.ErrQueryItemsTooMuch
	}

	excludeAccountIds := make([]int64, 0)
	excludeCategoryIds := make([]int64, 0)

	if transactionAmountsReq.ExcludeAccountIds != "" {
		excludeAccountIds, err = utils.StringArrayToInt64Array(strings.Split(transactionAmountsReq.ExcludeAccountIds, ","))

		if err != nil {
			return nil, errs.ErrAccountIdInvalid
		}
	}

	if transactionAmountsReq.ExcludeCategoryIds != "" {
		excludeCategoryIds, err = utils.StringArrayToInt64Array(strings.Split(transactionAmountsReq.ExcludeCategoryIds, ","))

		if err != nil {
			return nil, errs.ErrTransactionCategoryIdInvalid
		}
	}

	tagFilters, err := models.ParseTransactionTagFilter(transactionAmountsReq.TagFilter)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionAmountsHandler] parse transaction tag filters error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionAmountsHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()

	accounts, err := a.accounts.GetAllAccountsByUid(c, uid)
	accountMap := a.accounts.GetAccountMapByList(accounts)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionAmountsHandler] failed to get all accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	amountsResp := orderedmap.New[string, *models.TransactionAmountsResponseItem]()

	for i := 0; i < len(requestItems); i++ {
		requestItem := requestItems[i]

		incomeAmounts, expenseAmounts, err := a.transactions.GetAccountsTotalIncomeAndExpense(c, uid, requestItem.StartTime, requestItem.EndTime, excludeAccountIds, excludeCategoryIds, tagFilters, clientTimezone, transactionAmountsReq.UseTransactionTimezone)

		if err != nil {
			log.Errorf(c, "[transactions.TransactionAmountsHandler] failed to get transaction amounts item for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		amountsMap := make(map[string]*models.TransactionAmountsAndCurrency)

		for accountId, incomeAmount := range incomeAmounts {
			account, exists := accountMap[accountId]

			if !exists {
				log.Warnf(c, "[transactions.TransactionAmountsHandler] cannot find account for account \"id:%d\" of user \"uid:%d\"", accountId, uid)
				continue
			}

			totalAmounts, exists := amountsMap[account.Currency]

			if !exists {
				totalAmounts = &models.TransactionAmountsAndCurrency{
					Currency:      account.Currency,
					IncomeAmount:  big.NewInt(0),
					ExpenseAmount: big.NewInt(0),
				}
			}

			totalAmounts.IncomeAmount.Add(totalAmounts.IncomeAmount, incomeAmount)
			amountsMap[account.Currency] = totalAmounts
		}

		for accountId, expenseAmount := range expenseAmounts {
			account, exists := accountMap[accountId]

			if !exists {
				log.Warnf(c, "[transactions.TransactionAmountsHandler] cannot find account for account \"id:%d\" of user \"uid:%d\"", accountId, uid)
				continue
			}

			totalAmounts, exists := amountsMap[account.Currency]

			if !exists {
				totalAmounts = &models.TransactionAmountsAndCurrency{
					Currency:      account.Currency,
					IncomeAmount:  big.NewInt(0),
					ExpenseAmount: big.NewInt(0),
				}
			}

			totalAmounts.ExpenseAmount.Add(totalAmounts.ExpenseAmount, expenseAmount)
			amountsMap[account.Currency] = totalAmounts
		}

		allTotalAmounts := make(models.TransactionAmountsResponseItemAmountInfoSlice, 0)

		for _, totalAmounts := range amountsMap {
			allTotalAmounts = append(allTotalAmounts, &models.TransactionAmountsResponseItemAmountInfo{
				Currency:      totalAmounts.Currency,
				IncomeAmount:  totalAmounts.IncomeAmount.String(),
				ExpenseAmount: totalAmounts.ExpenseAmount.String(),
			})
		}

		sort.Sort(allTotalAmounts)

		amountsResp.Set(requestItem.Name, &models.TransactionAmountsResponseItem{
			StartTime: requestItem.StartTime,
			EndTime:   requestItem.EndTime,
			Amounts:   allTotalAmounts,
		})
	}

	return amountsResp, nil
}

// TransactionDailyAmountsHandler returns daily transaction amounts of current user
func (a *TransactionsApi) TransactionDailyAmountsHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionAmountsReq models.TransactionDailyAmountsRequest
	err := c.ShouldBindQuery(&transactionAmountsReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionDailyAmountsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	if transactionAmountsReq.EndTime < transactionAmountsReq.StartTime {
		return nil, errs.ErrDateRangeInvalid
	}

	excludeAccountIds := make([]int64, 0)
	excludeCategoryIds := make([]int64, 0)

	if transactionAmountsReq.ExcludeAccountIds != "" {
		excludeAccountIds, err = utils.StringArrayToInt64Array(strings.Split(transactionAmountsReq.ExcludeAccountIds, ","))

		if err != nil {
			return nil, errs.ErrAccountIdInvalid
		}
	}

	if transactionAmountsReq.ExcludeCategoryIds != "" {
		excludeCategoryIds, err = utils.StringArrayToInt64Array(strings.Split(transactionAmountsReq.ExcludeCategoryIds, ","))

		if err != nil {
			return nil, errs.ErrTransactionCategoryIdInvalid
		}
	}

	tagFilters, err := models.ParseTransactionTagFilter(transactionAmountsReq.TagFilter)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionDailyAmountsHandler] parse transaction tag filters error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionDailyAmountsHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	accounts, err := a.accounts.GetAllAccountsByUid(c, uid)
	accountMap := a.accounts.GetAccountMapByList(accounts)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionDailyAmountsHandler] failed to get all accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	incomeAmounts, expenseAmounts, err := a.transactions.GetAccountsDailyIncomeAndExpense(c, uid, transactionAmountsReq.StartTime, transactionAmountsReq.EndTime, excludeAccountIds, excludeCategoryIds, tagFilters, clientTimezone, transactionAmountsReq.UseTransactionTimezone)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionDailyAmountsHandler] failed to get daily amounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allDates := make(map[int32]bool)

	for date := range incomeAmounts {
		allDates[date] = true
	}

	for date := range expenseAmounts {
		allDates[date] = true
	}

	dates := make([]int32, 0, len(allDates))

	for date := range allDates {
		dates = append(dates, date)
	}

	slices.Sort(dates)

	result := make([]*models.TransactionDailyAmountsResponseItem, 0, len(dates))

	for _, date := range dates {
		amountsByCurrency := make(map[string]*models.TransactionAmountsAndCurrency)

		for accountId, amount := range incomeAmounts[date] {
			account, exists := accountMap[accountId]

			if !exists {
				log.Warnf(c, "[transactions.TransactionDailyAmountsHandler] cannot find account for account \"id:%d\" of user \"uid:%d\"", accountId, uid)
				continue
			}

			totalAmounts, exists := amountsByCurrency[account.Currency]

			if !exists {
				totalAmounts = &models.TransactionAmountsAndCurrency{
					Currency:      account.Currency,
					IncomeAmount:  big.NewInt(0),
					ExpenseAmount: big.NewInt(0),
				}
			}

			totalAmounts.IncomeAmount.Add(totalAmounts.IncomeAmount, amount)
			amountsByCurrency[account.Currency] = totalAmounts
		}

		for accountId, amount := range expenseAmounts[date] {
			account, exists := accountMap[accountId]

			if !exists {
				log.Warnf(c, "[transactions.TransactionDailyAmountsHandler] cannot find account for account \"id:%d\" of user \"uid:%d\"", accountId, uid)
				continue
			}

			totalAmounts, exists := amountsByCurrency[account.Currency]

			if !exists {
				totalAmounts = &models.TransactionAmountsAndCurrency{
					Currency:      account.Currency,
					IncomeAmount:  big.NewInt(0),
					ExpenseAmount: big.NewInt(0),
				}
			}

			totalAmounts.ExpenseAmount.Add(totalAmounts.ExpenseAmount, amount)
			amountsByCurrency[account.Currency] = totalAmounts
		}

		dailyTotalAmounts := make(models.TransactionAmountsResponseItemAmountInfoSlice, 0, len(amountsByCurrency))

		for _, total := range amountsByCurrency {
			dailyTotalAmounts = append(dailyTotalAmounts, &models.TransactionAmountsResponseItemAmountInfo{
				Currency:      total.Currency,
				IncomeAmount:  total.IncomeAmount.String(),
				ExpenseAmount: total.ExpenseAmount.String(),
			})
		}

		sort.Sort(dailyTotalAmounts)

		result = append(result, &models.TransactionDailyAmountsResponseItem{
			Date:    utils.FormatNumericYearMonthDayToLongDate(date),
			Amounts: dailyTotalAmounts,
		})
	}

	return result, nil
}
