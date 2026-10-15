package services

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"strings"
	"time"
	"xorm.io/builder"
	"xorm.io/xorm"
)

func (s *TransactionService) getAllTransactionsInSpecifiedDateRange(c core.Context, uid int64, startUnixTime int64, endUnixTime int64, excludeAccountIds []int64, excludeCategoryIds []int64, tagFilters []*models.TransactionTagFilter, clientTimezone *time.Location) ([]*models.Transaction, error) {
	startUnixTime = utils.GetMinUnixTimeWithSameLocalDateTime(startUnixTime, utils.GetTimezoneOffsetMinutes(startUnixTime, clientTimezone))
	endUnixTime = utils.GetMaxUnixTimeWithSameLocalDateTime(endUnixTime, utils.GetTimezoneOffsetMinutes(endUnixTime, clientTimezone))

	startTransactionTime := utils.GetMinTransactionTimeFromUnixTime(startUnixTime)
	endTransactionTime := utils.GetMaxTransactionTimeFromUnixTime(endUnixTime)

	condition := "uid=? AND deleted=? AND (type=? OR type=?)"
	conditionParams := make([]any, 0, 4+len(excludeAccountIds)+len(excludeCategoryIds))
	conditionParams = append(conditionParams, uid)
	conditionParams = append(conditionParams, false)
	conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_INCOME)
	conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_EXPENSE)

	if len(excludeAccountIds) > 0 {
		var accountIdsCondition strings.Builder
		accountIdConditionParams := make([]any, 0, len(excludeAccountIds))

		for i := 0; i < len(excludeAccountIds); i++ {
			if i > 0 {
				accountIdsCondition.WriteString(",")
			}

			accountIdsCondition.WriteString("?")
			accountIdConditionParams = append(accountIdConditionParams, excludeAccountIds[i])
		}

		condition = condition + " AND account_id NOT IN (" + accountIdsCondition.String() + ")"
		conditionParams = append(conditionParams, accountIdConditionParams...)
	}

	if len(excludeCategoryIds) > 0 {
		var categoryIdsCondition strings.Builder
		categoryIdConditionParams := make([]any, 0, len(excludeCategoryIds))

		for i := 0; i < len(excludeCategoryIds); i++ {
			if i > 0 {
				categoryIdsCondition.WriteString(",")
			}

			categoryIdsCondition.WriteString("?")
			categoryIdConditionParams = append(categoryIdConditionParams, excludeCategoryIds[i])
		}

		condition = condition + " AND category_id NOT IN (" + categoryIdsCondition.String() + ")"
		conditionParams = append(conditionParams, categoryIdConditionParams...)
	}

	condition = condition + " AND transaction_time>=? AND transaction_time<=?"

	minTransactionTime := startTransactionTime
	maxTransactionTime := endTransactionTime
	var allTransactions []*models.Transaction

	for maxTransactionTime > 0 {
		var transactions []*models.Transaction

		finalConditionParams := make([]any, 0, 6)
		finalConditionParams = append(finalConditionParams, conditionParams...)
		finalConditionParams = append(finalConditionParams, minTransactionTime)
		finalConditionParams = append(finalConditionParams, maxTransactionTime)

		sess := s.UserDataDB(uid).NewSession(c).Select("type, account_id, transaction_time, timezone_utc_offset, amount").Where(condition, finalConditionParams...)
		sess = s.appendFilterTagIdsConditionToQuery(sess, uid, maxTransactionTime, minTransactionTime, tagFilters, false)

		err := sess.Limit(pageCountForLoadTransactionAmounts, 0).OrderBy("transaction_time desc").Find(&transactions)

		if err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, transactions...)

		if len(transactions) < pageCountForLoadTransactionAmounts {
			maxTransactionTime = 0
			break
		}

		maxTransactionTime = transactions[len(transactions)-1].TransactionTime - 1
	}

	return allTransactions, nil
}

func (s *TransactionService) buildTransactionQueryCondition(uid int64, maxTransactionTime int64, minTransactionTime int64, transactionDbType models.TransactionDbType, categoryIds []int64, accountIds []int64, tagFilters []*models.TransactionTagFilter, amountFilter string, keyword string, matchMode core.MatchMode, noDuplicated bool) (string, []any) {
	condition := "uid=? AND deleted=?"
	conditionParams := make([]any, 0, 16)
	conditionParams = append(conditionParams, uid)
	conditionParams = append(conditionParams, false)

	if maxTransactionTime > 0 {
		condition = condition + " AND transaction_time<=?"
		conditionParams = append(conditionParams, maxTransactionTime)
	}

	if minTransactionTime > 0 {
		condition = condition + " AND transaction_time>=?"
		conditionParams = append(conditionParams, minTransactionTime)
	}

	var accountIdsCondition strings.Builder
	accountIdConditionParams := make([]any, 0, len(accountIds))

	for i := 0; i < len(accountIds); i++ {
		if i > 0 {
			accountIdsCondition.WriteString(",")
		}

		accountIdsCondition.WriteString("?")
		accountIdConditionParams = append(accountIdConditionParams, accountIds[i])
	}

	if models.TRANSACTION_DB_TYPE_MODIFY_BALANCE <= transactionDbType && transactionDbType <= models.TRANSACTION_DB_TYPE_EXPENSE {
		condition = condition + " AND type=?"
		conditionParams = append(conditionParams, transactionDbType)
	} else if transactionDbType == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transactionDbType == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		if len(accountIds) == 0 {
			condition = condition + " AND type=?"
			conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
		} else if len(accountIds) == 1 {
			condition = condition + " AND (type=? OR type=?)"
			conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
			conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_IN)
		} else { // len(accountsIds) > 1
			condition = condition + " AND (type=? OR (type=? AND related_account_id NOT IN (" + accountIdsCondition.String() + ")))"
			conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
			conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_IN)
			conditionParams = append(conditionParams, accountIdConditionParams...)
		}
	} else {
		if noDuplicated {
			if len(accountIds) == 0 {
				condition = condition + " AND (type=? OR type=? OR type=? OR type=?)"
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_MODIFY_BALANCE)
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_INCOME)
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_EXPENSE)
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
			} else if len(accountIds) == 1 {
				// Do Nothing
			} else { // len(accountsIds) > 1
				condition = condition + " AND (type=? OR type=? OR type=? OR type=? OR (type=? AND related_account_id NOT IN (" + accountIdsCondition.String() + ")))"
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_MODIFY_BALANCE)
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_INCOME)
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_EXPENSE)
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_IN)
				conditionParams = append(conditionParams, accountIdConditionParams...)
			}
		}
	}

	if len(categoryIds) > 0 {
		var conditions strings.Builder

		for i := 0; i < len(categoryIds); i++ {
			if i > 0 {
				conditions.WriteString(",")
			}

			conditions.WriteString("?")
			conditionParams = append(conditionParams, categoryIds[i])
		}

		if conditions.Len() > 1 {
			condition = condition + " AND category_id IN (" + conditions.String() + ")"
		} else {
			condition = condition + " AND category_id = " + conditions.String()
		}
	}

	if len(accountIds) > 0 {
		if accountIdsCondition.Len() > 1 {
			condition = condition + " AND account_id IN (" + accountIdsCondition.String() + ")"
		} else {
			condition = condition + " AND account_id = " + accountIdsCondition.String()
		}

		conditionParams = append(conditionParams, accountIdConditionParams...)
	}

	if amountFilter != "" {
		amountFilterItems := strings.Split(amountFilter, ":")

		if len(amountFilterItems) == 2 && amountFilterItems[0] == "gt" {
			value, err := utils.StringToInt64(amountFilterItems[1])

			if err == nil {
				condition = condition + " AND amount > ?"
				conditionParams = append(conditionParams, value)
			}
		} else if len(amountFilterItems) == 2 && amountFilterItems[0] == "lt" {
			value, err := utils.StringToInt64(amountFilterItems[1])

			if err == nil {
				condition = condition + " AND amount < ?"
				conditionParams = append(conditionParams, value)
			}
		} else if len(amountFilterItems) == 2 && amountFilterItems[0] == "eq" {
			value, err := utils.StringToInt64(amountFilterItems[1])

			if err == nil {
				condition = condition + " AND amount = ?"
				conditionParams = append(conditionParams, value)
			}
		} else if len(amountFilterItems) == 2 && amountFilterItems[0] == "ne" {
			value, err := utils.StringToInt64(amountFilterItems[1])

			if err == nil {
				condition = condition + " AND amount <> ?"
				conditionParams = append(conditionParams, value)
			}
		} else if len(amountFilterItems) == 3 && amountFilterItems[0] == "bt" {
			value1, err := utils.StringToInt64(amountFilterItems[1])
			value2, err := utils.StringToInt64(amountFilterItems[2])

			if err == nil {
				condition = condition + " AND amount >= ? AND amount <= ?"
				conditionParams = append(conditionParams, value1)
				conditionParams = append(conditionParams, value2)
			}
		} else if len(amountFilterItems) == 3 && amountFilterItems[0] == "nb" {
			value1, err := utils.StringToInt64(amountFilterItems[1])
			value2, err := utils.StringToInt64(amountFilterItems[2])

			if err == nil {
				condition = condition + " AND (amount < ? OR amount > ?)"
				conditionParams = append(conditionParams, value1)
				conditionParams = append(conditionParams, value2)
			}
		}
	}

	if keyword != "" {
		if matchMode == core.MATCH_MODE_IGNORE_CASE {
			condition = condition + " AND LOWER(comment) LIKE LOWER(?)"
		} else {
			condition = condition + " AND comment LIKE ?"
		}
		conditionParams = append(conditionParams, "%%"+keyword+"%%")
	}

	return condition, conditionParams
}

func (s *TransactionService) appendFilterTagIdsConditionToQuery(sess *xorm.Session, uid int64, maxTransactionTime int64, minTransactionTime int64, tagFilters []*models.TransactionTagFilter, noTags bool) *xorm.Session {
	if noTags {
		subQueryCondition := builder.And(builder.Eq{"uid": uid}, builder.Eq{"deleted": false})

		if maxTransactionTime > 0 {
			subQueryCondition = subQueryCondition.And(builder.Lte{"transaction_time": maxTransactionTime})
		}

		if minTransactionTime > 0 {
			subQueryCondition = subQueryCondition.And(builder.Gte{"transaction_time": minTransactionTime})
		}

		subQuery := builder.Select("transaction_id").From("transaction_tag_index").Where(subQueryCondition)
		sess.NotIn("transaction_id", subQuery).NotIn("related_id", subQuery)
		return sess
	}

	if len(tagFilters) < 1 {
		return sess
	}

	for i := 0; i < len(tagFilters); i++ {
		tagFilter := tagFilters[i]
		subQueryCondition := builder.And(builder.Eq{"uid": uid}, builder.Eq{"deleted": false})

		if maxTransactionTime > 0 {
			subQueryCondition = subQueryCondition.And(builder.Lte{"transaction_time": maxTransactionTime})
		}

		if minTransactionTime > 0 {
			subQueryCondition = subQueryCondition.And(builder.Gte{"transaction_time": minTransactionTime})
		}

		subQueryCondition = subQueryCondition.And(builder.In("tag_id", tagFilter.TagIds))
		subQuery := builder.Select("transaction_id").From("transaction_tag_index").Where(subQueryCondition)

		if tagFilter.Type == models.TRANSACTION_TAG_FILTER_HAS_ALL || tagFilter.Type == models.TRANSACTION_TAG_FILTER_NOT_HAS_ALL {
			subQuery = subQuery.GroupBy("transaction_id").Having(fmt.Sprintf("COUNT(DISTINCT tag_id) >= %d", len(tagFilter.TagIds)))
		}

		if tagFilter.Type == models.TRANSACTION_TAG_FILTER_HAS_ANY || tagFilter.Type == models.TRANSACTION_TAG_FILTER_HAS_ALL {
			sess.And(builder.Or(builder.In("transaction_id", subQuery), builder.In("related_id", subQuery)))
		} else if tagFilter.Type == models.TRANSACTION_TAG_FILTER_NOT_HAS_ANY || tagFilter.Type == models.TRANSACTION_TAG_FILTER_NOT_HAS_ALL {
			sess.NotIn("transaction_id", subQuery).NotIn("related_id", subQuery)
		}
	}

	return sess
}

func (s *TransactionService) appendFilterPicturesConditionToQuery(sess *xorm.Session, uid int64, mustHavePictures bool) *xorm.Session {
	if !mustHavePictures {
		return sess
	}

	subQuery := builder.Select("transaction_id").From("transaction_picture_info").Where(builder.And(builder.Eq{"uid": uid}, builder.Eq{"deleted": false}, builder.Neq{"transaction_id": models.TransactionPictureNewPictureTransactionId}))
	sess.And(builder.Or(builder.In("transaction_id", subQuery), builder.In("related_id", subQuery)))

	return sess
}
