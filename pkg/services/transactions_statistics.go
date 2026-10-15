package services

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"math/big"
	"time"
)

// GetRelatedTransferTransaction returns the related transaction for transfer transaction
func (s *TransactionService) GetRelatedTransferTransaction(originalTransaction *models.Transaction) *models.Transaction {
	var relatedType models.TransactionDbType
	var relatedTransactionTime int64

	if originalTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
		relatedType = models.TRANSACTION_DB_TYPE_TRANSFER_IN
		relatedTransactionTime = originalTransaction.TransactionTime + 1
	} else if originalTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		relatedType = models.TRANSACTION_DB_TYPE_TRANSFER_OUT
		relatedTransactionTime = originalTransaction.TransactionTime - 1
	} else {
		return nil
	}

	relatedTransaction := &models.Transaction{
		TransactionId:        originalTransaction.RelatedId,
		Uid:                  originalTransaction.Uid,
		Deleted:              originalTransaction.Deleted,
		Type:                 relatedType,
		CategoryId:           originalTransaction.CategoryId,
		TransactionTime:      relatedTransactionTime,
		TimezoneUtcOffset:    originalTransaction.TimezoneUtcOffset,
		AccountId:            originalTransaction.RelatedAccountId,
		Amount:               originalTransaction.RelatedAccountAmount,
		RelatedId:            originalTransaction.TransactionId,
		RelatedAccountId:     originalTransaction.AccountId,
		RelatedAccountAmount: originalTransaction.Amount,
		Comment:              originalTransaction.Comment,
		GeoLongitude:         originalTransaction.GeoLongitude,
		GeoLatitude:          originalTransaction.GeoLatitude,
		CreatedIp:            originalTransaction.CreatedIp,
		CreatedUnixTime:      originalTransaction.CreatedUnixTime,
		UpdatedUnixTime:      originalTransaction.UpdatedUnixTime,
		DeletedUnixTime:      originalTransaction.DeletedUnixTime,
	}

	return relatedTransaction
}

// GetAccountsTotalIncomeAndExpense returns the every accounts total income and expense amount by specific date range
func (s *TransactionService) GetAccountsTotalIncomeAndExpense(c core.Context, uid int64, startUnixTime int64, endUnixTime int64, excludeAccountIds []int64, excludeCategoryIds []int64, tagFilters []*models.TransactionTagFilter, clientTimezone *time.Location, useTransactionTimezone bool) (map[int64]*big.Int, map[int64]*big.Int, error) {
	if uid <= 0 {
		return nil, nil, errs.ErrUserIdInvalid
	}

	startLocalDateTime := utils.FormatUnixTimeToNumericLocalDateTime(startUnixTime, clientTimezone)
	endLocalDateTime := utils.FormatUnixTimeToNumericLocalDateTime(endUnixTime, clientTimezone)

	allTransactions, err := s.getAllTransactionsInSpecifiedDateRange(c, uid, startUnixTime, endUnixTime, excludeAccountIds, excludeCategoryIds, tagFilters, clientTimezone)

	if err != nil {
		return nil, nil, err
	}

	incomeAmounts := make(map[int64]*big.Int)
	expenseAmounts := make(map[int64]*big.Int)

	for i := 0; i < len(allTransactions); i++ {
		transaction := allTransactions[i]
		timeZone := clientTimezone

		if useTransactionTimezone {
			timeZone = time.FixedZone("Transaction Timezone", int(transaction.TimezoneUtcOffset)*60)
		}

		localDateTime := utils.FormatUnixTimeToNumericLocalDateTime(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime), timeZone)

		if localDateTime < startLocalDateTime || localDateTime > endLocalDateTime {
			continue
		}

		var amountsMap map[int64]*big.Int

		if transaction.Type == models.TRANSACTION_DB_TYPE_INCOME {
			amountsMap = incomeAmounts
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_EXPENSE {
			amountsMap = expenseAmounts
		}

		totalAmounts, exists := amountsMap[transaction.AccountId]

		if !exists {
			totalAmounts = big.NewInt(0)
		}

		totalAmounts.Add(totalAmounts, big.NewInt(transaction.Amount))
		amountsMap[transaction.AccountId] = totalAmounts
	}

	return incomeAmounts, expenseAmounts, nil
}

// GetAccountsDailyIncomeAndExpense returns daily income and expense amounts grouped by account
func (s *TransactionService) GetAccountsDailyIncomeAndExpense(c core.Context, uid int64, startUnixTime int64, endUnixTime int64, excludeAccountIds []int64, excludeCategoryIds []int64, tagFilters []*models.TransactionTagFilter, clientTimezone *time.Location, useTransactionTimezone bool) (map[int32]map[int64]*big.Int, map[int32]map[int64]*big.Int, error) {
	if uid <= 0 {
		return nil, nil, errs.ErrUserIdInvalid
	}

	startLocalDateTime := utils.FormatUnixTimeToNumericLocalDateTime(startUnixTime, clientTimezone)
	endLocalDateTime := utils.FormatUnixTimeToNumericLocalDateTime(endUnixTime, clientTimezone)

	allTransactions, err := s.getAllTransactionsInSpecifiedDateRange(c, uid, startUnixTime, endUnixTime, excludeAccountIds, excludeCategoryIds, tagFilters, clientTimezone)

	if err != nil {
		return nil, nil, err
	}

	incomeAmounts := make(map[int32]map[int64]*big.Int)
	expenseAmounts := make(map[int32]map[int64]*big.Int)

	for i := 0; i < len(allTransactions); i++ {
		transaction := allTransactions[i]
		timeZone := clientTimezone

		if useTransactionTimezone {
			timeZone = time.FixedZone("Transaction Timezone", int(transaction.TimezoneUtcOffset)*60)
		}

		unixTime := utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime)
		localDateTime := utils.FormatUnixTimeToNumericLocalDateTime(unixTime, timeZone)

		if localDateTime < startLocalDateTime || localDateTime > endLocalDateTime {
			continue
		}

		var allAmounts map[int32]map[int64]*big.Int

		if transaction.Type == models.TRANSACTION_DB_TYPE_INCOME {
			allAmounts = incomeAmounts
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_EXPENSE {
			allAmounts = expenseAmounts
		}

		yearMonthDay := utils.FormatUnixTimeToNumericYearMonthDay(unixTime, timeZone)
		dailyAmounts, exists := allAmounts[yearMonthDay]

		if !exists {
			dailyAmounts = make(map[int64]*big.Int)
		}

		amount, exists := dailyAmounts[transaction.AccountId]

		if !exists {
			amount = big.NewInt(0)
		}

		amount.Add(amount, big.NewInt(transaction.Amount))
		dailyAmounts[transaction.AccountId] = amount
		allAmounts[yearMonthDay] = dailyAmounts
	}

	return incomeAmounts, expenseAmounts, nil
}

// GetAccountsAndCategoriesTotalInflowAndOutflow returns the every accounts and categories total inflows and outflows amount by specific date range
func (s *TransactionService) GetAccountsAndCategoriesTotalInflowAndOutflow(c core.Context, uid int64, startUnixTime int64, endUnixTime int64, tagFilters []*models.TransactionTagFilter, noTags bool, keyword string, matchMode core.MatchMode, clientTimezone *time.Location, useTransactionTimezone bool) ([]*models.TransactionTotalAmount, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var startLocalDateTime, endLocalDateTime, startTransactionTime, endTransactionTime int64

	if startUnixTime > 0 {
		startLocalDateTime = utils.FormatUnixTimeToNumericLocalDateTime(startUnixTime, clientTimezone)
		startUnixTime = utils.GetMinUnixTimeWithSameLocalDateTime(startUnixTime, utils.GetTimezoneOffsetMinutes(startUnixTime, clientTimezone))
		startTransactionTime = utils.GetMinTransactionTimeFromUnixTime(startUnixTime)
	}

	if endUnixTime > 0 {
		endLocalDateTime = utils.FormatUnixTimeToNumericLocalDateTime(endUnixTime, clientTimezone)
		endUnixTime = utils.GetMaxUnixTimeWithSameLocalDateTime(endUnixTime, utils.GetTimezoneOffsetMinutes(endUnixTime, clientTimezone))
		endTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(endUnixTime)
	}

	condition := "uid=? AND deleted=? AND (type=? OR type=? OR type=? OR type=?)"
	conditionParams := make([]any, 0, 6)
	conditionParams = append(conditionParams, uid)
	conditionParams = append(conditionParams, false)
	conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_INCOME)
	conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_EXPENSE)
	conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
	conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_IN)

	minTransactionTime := startTransactionTime
	maxTransactionTime := endTransactionTime
	var allTransactions []*models.Transaction

	for maxTransactionTime >= 0 {
		var transactions []*models.Transaction

		finalCondition := condition
		finalConditionParams := make([]any, 0, 6)
		finalConditionParams = append(finalConditionParams, conditionParams...)

		if minTransactionTime > 0 {
			finalCondition = finalCondition + " AND transaction_time>=?"
			finalConditionParams = append(finalConditionParams, minTransactionTime)
		}

		if maxTransactionTime > 0 {
			finalCondition = finalCondition + " AND transaction_time<=?"
			finalConditionParams = append(finalConditionParams, maxTransactionTime)
		}

		if keyword != "" {
			if matchMode == core.MATCH_MODE_IGNORE_CASE {
				finalCondition = finalCondition + " AND LOWER(comment) LIKE LOWER(?)"
			} else {
				finalCondition = finalCondition + " AND comment LIKE ?"
			}
			finalConditionParams = append(finalConditionParams, "%%"+keyword+"%%")
		}

		sess := s.UserDataDB(uid).NewSession(c).Select("type, category_id, account_id, related_account_id, transaction_time, timezone_utc_offset, amount").Where(finalCondition, finalConditionParams...)
		sess = s.appendFilterTagIdsConditionToQuery(sess, uid, maxTransactionTime, minTransactionTime, tagFilters, noTags)

		err := sess.Limit(pageCountForLoadTransactionAmounts, 0).OrderBy("transaction_time desc").Find(&transactions)

		if err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, transactions...)

		if len(transactions) < pageCountForLoadTransactionAmounts {
			maxTransactionTime = -1
			break
		}

		maxTransactionTime = transactions[len(transactions)-1].TransactionTime - 1
	}

	transactionTotalAmountsMap := make(map[string]*models.TransactionTotalAmount)

	for i := 0; i < len(allTransactions); i++ {
		transaction := allTransactions[i]
		timeZone := clientTimezone

		if useTransactionTimezone {
			timeZone = time.FixedZone("Transaction Timezone", int(transaction.TimezoneUtcOffset)*60)
		}

		localDateTime := utils.FormatUnixTimeToNumericLocalDateTime(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime), timeZone)

		if (startLocalDateTime > 0 && localDateTime < startLocalDateTime) || (endLocalDateTime > 0 && localDateTime > endLocalDateTime) {
			continue
		}

		groupKey := fmt.Sprintf("%d_%d", transaction.CategoryId, transaction.AccountId)

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			groupKey = fmt.Sprintf("%d_%d_%d_%d", transaction.CategoryId, transaction.AccountId, transaction.RelatedAccountId, transaction.Type)
		}

		totalAmounts, exists := transactionTotalAmountsMap[groupKey]

		if !exists {
			totalAmounts = &models.TransactionTotalAmount{
				Type:             transaction.Type,
				CategoryId:       transaction.CategoryId,
				AccountId:        transaction.AccountId,
				RelatedAccountId: transaction.RelatedAccountId,
				Amount:           big.NewInt(0),
			}

			transactionTotalAmountsMap[groupKey] = totalAmounts
		}

		totalAmounts.Amount.Add(totalAmounts.Amount, big.NewInt(transaction.Amount))
	}

	transactionTotalAmounts := make([]*models.TransactionTotalAmount, 0, len(transactionTotalAmountsMap))

	for _, totalAmounts := range transactionTotalAmountsMap {
		transactionTotalAmounts = append(transactionTotalAmounts, totalAmounts)
	}

	return transactionTotalAmounts, nil
}
