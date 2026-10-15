package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"time"
)

// GetTransactionsByMaxTimeUpToCount returns transactions before given time and up to given count
func (s *TransactionService) GetTransactionsByMaxTimeUpToCount(c core.Context, uid int64, maxTransactionTime int64, minTransactionTime int64, transactionType models.TransactionType, categoryIds []int64, accountIds []int64, tagFilters []*models.TransactionTagFilter, noTags bool, amountFilter string, keyword string, matchMode core.MatchMode, mustHavePictures bool, page int32, count int32, pageCount int32, needOneMoreItem bool, noDuplicated bool) ([]*models.Transaction, error) {
	if maxTransactionTime <= 0 {
		maxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(time.Now().Unix())
	}

	if page < 0 {
		return nil, errs.ErrPageIndexInvalid
	} else if page == 0 {
		page = 1
	}

	if count < 1 {
		return nil, errs.ErrPageCountInvalid
	}

	finalExpectedCount := int(count)

	if needOneMoreItem {
		finalExpectedCount++
	}

	var allTransactions []*models.Transaction
	startOffset := int((page - 1) * count)
	firstFetchCount := int(pageCount)

	if finalExpectedCount < firstFetchCount {
		firstFetchCount = finalExpectedCount
	}

	transactions, err := s.getTransactionsByMaxTimeWithOffset(c, uid, maxTransactionTime, minTransactionTime, transactionType, categoryIds, accountIds, tagFilters, noTags, amountFilter, keyword, matchMode, mustHavePictures, startOffset, firstFetchCount, noDuplicated)

	if err != nil {
		return nil, err
	}

	allTransactions = append(allTransactions, transactions...)

	if len(transactions) < firstFetchCount {
		return allTransactions, nil
	}

	maxTransactionTime = transactions[len(transactions)-1].TransactionTime - 1

	for len(allTransactions) < finalExpectedCount && maxTransactionTime > 0 {
		remainingCount := finalExpectedCount - len(allTransactions)
		fetchCount := int(pageCount)

		if remainingCount < fetchCount {
			fetchCount = remainingCount
		}

		transactions, err := s.GetTransactionsByMaxTime(c, uid, maxTransactionTime, minTransactionTime, transactionType, categoryIds, accountIds, tagFilters, noTags, amountFilter, keyword, matchMode, mustHavePictures, 1, int32(fetchCount), false, noDuplicated)

		if err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, transactions...)

		if len(transactions) < fetchCount {
			break
		}

		maxTransactionTime = transactions[len(transactions)-1].TransactionTime - 1
	}

	return allTransactions, nil
}

// GetTransactionsByMaxTime returns transactions before given time
func (s *TransactionService) GetTransactionsByMaxTime(c core.Context, uid int64, maxTransactionTime int64, minTransactionTime int64, transactionType models.TransactionType, categoryIds []int64, accountIds []int64, tagFilters []*models.TransactionTagFilter, noTags bool, amountFilter string, keyword string, matchMode core.MatchMode, mustHavePictures bool, page int32, count int32, needOneMoreItem bool, noDuplicated bool) ([]*models.Transaction, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if page < 0 {
		return nil, errs.ErrPageIndexInvalid
	} else if page == 0 {
		page = 1
	}

	if count < 1 {
		return nil, errs.ErrPageCountInvalid
	}

	finalCount := int(count)

	if needOneMoreItem {
		finalCount++
	}

	return s.getTransactionsByMaxTimeWithOffset(c, uid, maxTransactionTime, minTransactionTime, transactionType, categoryIds, accountIds, tagFilters, noTags, amountFilter, keyword, matchMode, mustHavePictures, int(count*(page-1)), finalCount, noDuplicated)
}

// getTransactionsByMaxTimeWithOffset returns transactions before given time with explicit offset and limit
func (s *TransactionService) getTransactionsByMaxTimeWithOffset(c core.Context, uid int64, maxTransactionTime int64, minTransactionTime int64, transactionType models.TransactionType, categoryIds []int64, accountIds []int64, tagFilters []*models.TransactionTagFilter, noTags bool, amountFilter string, keyword string, matchMode core.MatchMode, mustHavePictures bool, offset int, limit int, noDuplicated bool) ([]*models.Transaction, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var err error
	var transactionDbType models.TransactionDbType = 0

	if transactionType > 0 {
		transactionDbType, err = transactionType.ToTransactionDbType()

		if err != nil {
			return nil, err
		}
	}

	var transactions []*models.Transaction

	condition, conditionParams := s.buildTransactionQueryCondition(uid, maxTransactionTime, minTransactionTime, transactionDbType, categoryIds, accountIds, tagFilters, amountFilter, keyword, matchMode, noDuplicated)
	sess := s.UserDataDB(uid).NewSession(c).Where(condition, conditionParams...)
	sess = s.appendFilterTagIdsConditionToQuery(sess, uid, maxTransactionTime, minTransactionTime, tagFilters, noTags)
	sess = s.appendFilterPicturesConditionToQuery(sess, uid, mustHavePictures)

	err = sess.Limit(limit, offset).OrderBy("transaction_time desc").Find(&transactions)

	return transactions, err
}

// GetTransactionsInMonthByPage returns all transactions in given year and month
func (s *TransactionService) GetTransactionsInMonthByPage(c core.Context, uid int64, year int32, month int32, transactionType models.TransactionType, categoryIds []int64, accountIds []int64, tagFilters []*models.TransactionTagFilter, noTags bool, amountFilter string, keyword string, matchMode core.MatchMode, mustHavePictures bool) ([]*models.Transaction, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var err error
	var transactionDbType models.TransactionDbType = 0

	if transactionType > 0 {
		transactionDbType, err = transactionType.ToTransactionDbType()

		if err != nil {
			return nil, err
		}
	}

	minTransactionTime, maxTransactionTime, err := utils.GetTransactionTimeRangeByYearMonth(year, month)

	if err != nil {
		return nil, errs.ErrSystemError
	}

	var transactions []*models.Transaction

	condition, conditionParams := s.buildTransactionQueryCondition(uid, maxTransactionTime, minTransactionTime, transactionDbType, categoryIds, accountIds, tagFilters, amountFilter, keyword, matchMode, true)
	sess := s.UserDataDB(uid).NewSession(c).Where(condition, conditionParams...)
	sess = s.appendFilterTagIdsConditionToQuery(sess, uid, maxTransactionTime, minTransactionTime, tagFilters, noTags)
	sess = s.appendFilterPicturesConditionToQuery(sess, uid, mustHavePictures)

	err = sess.OrderBy("transaction_time desc").Find(&transactions)

	transactionsInMonth := make([]*models.Transaction, 0, len(transactions))

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]
		transactionUnixTime := utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime)
		transactionTimeZone := time.FixedZone("Transaction Timezone", int(transaction.TimezoneUtcOffset)*60)

		if utils.IsUnixTimeEqualsYearAndMonth(transactionUnixTime, transactionTimeZone, year, month) {
			transactionsInMonth = append(transactionsInMonth, transaction)
		}
	}

	return transactionsInMonth, err
}

// GetTransactionByTransactionId returns a transaction model according to transaction id
func (s *TransactionService) GetTransactionByTransactionId(c core.Context, uid int64, transactionId int64) (*models.Transaction, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if transactionId <= 0 {
		return nil, errs.ErrTransactionIdInvalid
	}

	transaction := &models.Transaction{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(transactionId).Where("uid=? AND deleted=?", uid, false).Get(transaction)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrTransactionNotFound
	}

	return transaction, nil
}

// GetTransactionsByTransactionIds returns transaction models according to transaction ids
func (s *TransactionService) GetTransactionsByTransactionIds(c core.Context, uid int64, transactionIds []int64) ([]*models.Transaction, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if len(transactionIds) <= 0 {
		return nil, errs.ErrTransactionIdInvalid
	}

	var transactions []*models.Transaction
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).In("transaction_id", transactionIds).Find(&transactions)

	return transactions, err
}

// GetAllTransactionCount returns total count of transactions
func (s *TransactionService) GetAllTransactionCount(c core.Context, uid int64) (int64, error) {
	return s.GetTransactionCount(c, uid, 0, 0, 0, nil, nil, nil, false, "", "", core.MATCH_MODE_DEFAULT, false)
}

// GetTransactionCount returns count of transactions
func (s *TransactionService) GetTransactionCount(c core.Context, uid int64, maxTransactionTime int64, minTransactionTime int64, transactionType models.TransactionType, categoryIds []int64, accountIds []int64, tagFilters []*models.TransactionTagFilter, noTags bool, amountFilter string, keyword string, matchMode core.MatchMode, mustHavePictures bool) (int64, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	var err error
	var transactionDbType models.TransactionDbType = 0

	if transactionType > 0 {
		transactionDbType, err = transactionType.ToTransactionDbType()

		if err != nil {
			return 0, err
		}
	}

	condition, conditionParams := s.buildTransactionQueryCondition(uid, maxTransactionTime, minTransactionTime, transactionDbType, categoryIds, accountIds, tagFilters, amountFilter, keyword, matchMode, true)
	sess := s.UserDataDB(uid).NewSession(c).Where(condition, conditionParams...)
	sess = s.appendFilterTagIdsConditionToQuery(sess, uid, maxTransactionTime, minTransactionTime, tagFilters, noTags)
	sess = s.appendFilterPicturesConditionToQuery(sess, uid, mustHavePictures)

	return sess.Count(&models.Transaction{})
}
