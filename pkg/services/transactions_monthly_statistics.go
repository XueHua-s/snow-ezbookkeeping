package services

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"math/big"
	"strings"
	"time"
	"xorm.io/builder"
)

// GetUnreconciledTransactionCounts returns transaction counts after each account's last reconciled time
func (s *TransactionService) GetUnreconciledTransactionCounts(c core.Context, uid int64, accountLastReconciledTimes map[int64]int64) ([]*models.TransactionUnreconciledCountItem, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	accountConditions := make([]builder.Cond, 0, len(accountLastReconciledTimes))

	for accountId, lastReconciledTime := range accountLastReconciledTimes {
		accountConditions = append(accountConditions, builder.And(
			builder.Eq{"account_id": accountId},
			builder.Gt{"transaction_time": utils.GetMaxTransactionTimeFromUnixTime(lastReconciledTime)},
		))
	}

	counts := make([]*models.TransactionUnreconciledCountItem, 0)

	if len(accountConditions) < 1 {
		return counts, nil
	}

	err := s.UserDataDB(uid).NewSession(c).Table(&models.Transaction{}).Select("account_id, SUM(1) AS count").Where(builder.And(builder.Eq{"uid": uid, "deleted": false}, builder.Or(accountConditions...))).GroupBy("account_id").Having("SUM(1) > 0").OrderBy("account_id asc").Find(&counts)

	return counts, err
}

// GetAccountsAndCategoriesMonthlyInflowAndOutflow returns the every accounts monthly inflows and outflows amount by specific date range
func (s *TransactionService) GetAccountsAndCategoriesMonthlyInflowAndOutflow(c core.Context, uid int64, startYear int32, startMonth int32, endYear int32, endMonth int32, tagFilters []*models.TransactionTagFilter, noTags bool, keyword string, matchMode core.MatchMode, clientTimezone *time.Location, useTransactionTimezone bool) (map[int32][]*models.TransactionTotalAmount, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var startTransactionTime, endTransactionTime int64
	var err error

	if startYear > 0 && startMonth > 0 {
		startTransactionTime, _, err = utils.GetTransactionTimeRangeByYearMonth(startYear, startMonth)

		if err != nil {
			return nil, errs.ErrSystemError
		}
	}

	if endYear > 0 && endMonth > 0 {
		_, endTransactionTime, err = utils.GetTransactionTimeRangeByYearMonth(endYear, endMonth)

		if err != nil {
			return nil, errs.ErrSystemError
		}
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

	startYearMonth := startYear*100 + startMonth
	endYearMonth := endYear*100 + endMonth
	transactionsMonthlyAmountsMap := make(map[string]*models.TransactionTotalAmount)
	transactionsMonthlyAmounts := make(map[int32][]*models.TransactionTotalAmount)

	for i := 0; i < len(allTransactions); i++ {
		transaction := allTransactions[i]
		timeZone := clientTimezone

		if useTransactionTimezone {
			timeZone = time.FixedZone("Transaction Timezone", int(transaction.TimezoneUtcOffset)*60)
		}

		yearMonth := utils.FormatUnixTimeToNumericYearMonth(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime), timeZone)

		if (startYearMonth > 0 && yearMonth < startYearMonth) || (endYearMonth > 0 && yearMonth > endYearMonth) {
			continue
		}

		groupKey := fmt.Sprintf("%d_%d_%d", yearMonth, transaction.CategoryId, transaction.AccountId)

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			groupKey = fmt.Sprintf("%d_%d_%d_%d_%d", yearMonth, transaction.CategoryId, transaction.AccountId, transaction.RelatedAccountId, transaction.Type)
		}

		transactionAmounts, exists := transactionsMonthlyAmountsMap[groupKey]

		if !exists {
			transactionAmounts = &models.TransactionTotalAmount{
				Type:             transaction.Type,
				CategoryId:       transaction.CategoryId,
				AccountId:        transaction.AccountId,
				RelatedAccountId: transaction.RelatedAccountId,
				Amount:           big.NewInt(0),
			}

			transactionsMonthlyAmountsMap[groupKey] = transactionAmounts
		}

		transactionAmounts.Amount.Add(transactionAmounts.Amount, big.NewInt(transaction.Amount))
	}

	for groupKey, transaction := range transactionsMonthlyAmountsMap {
		groupKeyParts := strings.Split(groupKey, "_")
		yearMonth, _ := utils.StringToInt32(groupKeyParts[0])
		monthlyAmounts, exists := transactionsMonthlyAmounts[yearMonth]

		if !exists {
			monthlyAmounts = make([]*models.TransactionTotalAmount, 0)
		}

		monthlyAmounts = append(monthlyAmounts, transaction)
		transactionsMonthlyAmounts[yearMonth] = monthlyAmounts
	}

	return transactionsMonthlyAmounts, nil
}

// GetTransactionMapByList returns a transaction map by a list
func (s *TransactionService) GetTransactionMapByList(transactions []*models.Transaction) map[int64]*models.Transaction {
	transactionMap := make(map[int64]*models.Transaction)

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]
		transactionMap[transaction.TransactionId] = transaction
	}

	return transactionMap
}

// GetTransactionIds returns transaction ids list
func (s *TransactionService) GetTransactionIds(transactions []*models.Transaction) []int64 {
	transactionIds := make([]int64, len(transactions))

	for i := 0; i < len(transactions); i++ {
		transactionIds[i] = transactions[i].TransactionId
	}

	return transactionIds
}
