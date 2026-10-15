package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/duplicatechecker"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"math"
)

const pageCountForAccountStatement = 1000

const pageCountForMovingAccountTransactions = 1000

// TransactionsApi represents transaction api
type TransactionsApi struct {
	ApiUsingConfig
	ApiUsingDuplicateChecker
	transactions          *services.TransactionService
	transactionCategories *services.TransactionCategoryService
	transactionTags       *services.TransactionTagService
	transactionPictures   *services.TransactionPictureService
	accounts              *services.AccountService
	users                 *services.UserService
}

// Initialize a transaction api singleton instance
var (
	Transactions = &TransactionsApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		ApiUsingDuplicateChecker: ApiUsingDuplicateChecker{
			ApiUsingConfig: ApiUsingConfig{
				container: settings.Container,
			},
			container: duplicatechecker.Container,
		},
		transactions:          services.Transactions,
		transactionCategories: services.TransactionCategories,
		transactionTags:       services.TransactionTags,
		transactionPictures:   services.TransactionPictures,
		accounts:              services.Accounts,
		users:                 services.Users,
	}
)

// TransactionCountHandler returns transaction total count of current user
func (a *TransactionsApi) TransactionCountHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionCountReq models.TransactionCountRequest
	err := c.ShouldBindQuery(&transactionCountReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionCountHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	allAccountIds, err := a.accounts.GetAccountOrSubAccountIds(c, transactionCountReq.AccountIds, uid)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionCountHandler] get account error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allCategoryIds, err := a.transactionCategories.GetCategoryOrSubCategoryIds(c, transactionCountReq.CategoryIds, uid)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionCountHandler] get transaction category error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	noTags := transactionCountReq.TagFilter == models.TransactionNoTagFilterValue
	var tagFilters []*models.TransactionTagFilter

	if !noTags {
		tagFilters, err = models.ParseTransactionTagFilter(transactionCountReq.TagFilter)

		if err != nil {
			log.Warnf(c, "[transactions.TransactionCountHandler] parse transaction filters error, because %s", err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	totalCount, err := a.transactions.GetTransactionCount(c, uid, transactionCountReq.MaxTime, transactionCountReq.MinTime, transactionCountReq.Type, allCategoryIds, allAccountIds, tagFilters, noTags, transactionCountReq.AmountFilter, transactionCountReq.Keyword, transactionCountReq.MatchMode, transactionCountReq.MustHavePictures)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionCountHandler] failed to get transaction count for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	countResp := &models.TransactionCountResponse{
		TotalCount: totalCount,
	}

	return countResp, nil
}

// TransactionListHandler returns transaction list of current user
func (a *TransactionsApi) TransactionListHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionListReq models.TransactionListByMaxTimeRequest
	err := c.ShouldBindQuery(&transactionListReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionListHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionListHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	allAccountIds, err := a.accounts.GetAccountOrSubAccountIds(c, transactionListReq.AccountIds, uid)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionListHandler] get account error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allCategoryIds, err := a.transactionCategories.GetCategoryOrSubCategoryIds(c, transactionListReq.CategoryIds, uid)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionListHandler] get transaction category error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	noTags := transactionListReq.TagFilter == models.TransactionNoTagFilterValue
	var tagFilters []*models.TransactionTagFilter

	if !noTags {
		tagFilters, err = models.ParseTransactionTagFilter(transactionListReq.TagFilter)

		if err != nil {
			log.Warnf(c, "[transactions.TransactionListHandler] parse transaction tag filters error, because %s", err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	var totalCount int64

	if transactionListReq.WithCount {
		totalCount, err = a.transactions.GetTransactionCount(c, uid, transactionListReq.MaxTime, transactionListReq.MinTime, transactionListReq.Type, allCategoryIds, allAccountIds, tagFilters, noTags, transactionListReq.AmountFilter, transactionListReq.Keyword, transactionListReq.MatchMode, transactionListReq.MustHavePictures)

		if err != nil {
			log.Errorf(c, "[transactions.TransactionListHandler] failed to get transaction count for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	transactions, err := a.transactions.GetTransactionsByMaxTime(c, uid, transactionListReq.MaxTime, transactionListReq.MinTime, transactionListReq.Type, allCategoryIds, allAccountIds, tagFilters, noTags, transactionListReq.AmountFilter, transactionListReq.Keyword, transactionListReq.MatchMode, transactionListReq.MustHavePictures, transactionListReq.Page, transactionListReq.Count, true, true)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionListHandler] failed to get transactions earlier than \"%d\" for user \"uid:%d\", because %s", transactionListReq.MaxTime, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	hasMore := false
	var nextTimeSequenceId *int64

	if len(transactions) > int(transactionListReq.Count) {
		hasMore = true
		nextTimeSequenceId = &transactions[transactionListReq.Count].TransactionTime
		transactions = transactions[:transactionListReq.Count]
	}

	accountMap, categoryMap, tagMap, allTransactionTagIds, pictureInfoMap, err := a.getTransactionEssentialDataByTransactionIds(c, user, transactions, transactionListReq.WithPictures, transactionListReq.TrimCategory, transactionListReq.TrimTag)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionListHandler] failed to get essential data for assembling transaction result for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactions = a.filterTransactions(c, uid, transactions, accountMap)
	transactionResult, err := a.getTransactionResponseListResult(c, user, transactions, accountMap, categoryMap, tagMap, allTransactionTagIds, pictureInfoMap, clientTimezone, transactionListReq.WithPictures, transactionListReq.TrimAccount, transactionListReq.TrimCategory, transactionListReq.TrimTag)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionListHandler] failed to assemble transaction result for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactionResps := &models.TransactionInfoPageWrapperResponse{
		Items: transactionResult,
	}

	if hasMore {
		transactionResps.NextTimeSequenceId = nextTimeSequenceId
	}

	if transactionListReq.WithCount {
		transactionResps.TotalCount = &totalCount
	}

	return transactionResps, nil
}

// TransactionMonthListHandler returns all transaction list of current user by month
func (a *TransactionsApi) TransactionMonthListHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionListReq models.TransactionListInMonthByPageRequest
	err := c.ShouldBindQuery(&transactionListReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionMonthListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionMonthListHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionMonthListHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	allAccountIds, err := a.accounts.GetAccountOrSubAccountIds(c, transactionListReq.AccountIds, uid)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionMonthListHandler] get account error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allCategoryIds, err := a.transactionCategories.GetCategoryOrSubCategoryIds(c, transactionListReq.CategoryIds, uid)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionMonthListHandler] get transaction category error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	noTags := transactionListReq.TagFilter == models.TransactionNoTagFilterValue
	var tagFilters []*models.TransactionTagFilter

	if !noTags {
		tagFilters, err = models.ParseTransactionTagFilter(transactionListReq.TagFilter)

		if err != nil {
			log.Warnf(c, "[transactions.TransactionMonthListHandler] parse transaction tag filters error, because %s", err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	transactions, err := a.transactions.GetTransactionsInMonthByPage(c, uid, transactionListReq.Year, transactionListReq.Month, transactionListReq.Type, allCategoryIds, allAccountIds, tagFilters, noTags, transactionListReq.AmountFilter, transactionListReq.Keyword, transactionListReq.MatchMode, transactionListReq.MustHavePictures)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionMonthListHandler] failed to get transactions in month \"%d-%d\" for user \"uid:%d\", because %s", transactionListReq.Year, transactionListReq.Month, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	accountMap, categoryMap, tagMap, allTransactionTagIds, pictureInfoMap, err := a.getTransactionEssentialDataByTransactionIds(c, user, transactions, transactionListReq.WithPictures, transactionListReq.TrimCategory, transactionListReq.TrimTag)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionMonthListHandler] failed to get essential data for assembling transaction result for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactions = a.filterTransactions(c, uid, transactions, accountMap)
	transactionResult, err := a.getTransactionResponseListResult(c, user, transactions, accountMap, categoryMap, tagMap, allTransactionTagIds, pictureInfoMap, clientTimezone, transactionListReq.WithPictures, transactionListReq.TrimAccount, transactionListReq.TrimCategory, transactionListReq.TrimTag)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionMonthListHandler] failed to assemble transaction result for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactionResps := &models.TransactionInfoPageWrapperResponse2{
		Items:      transactionResult,
		TotalCount: int64(transactionResult.Len()),
	}

	return transactionResps, nil
}

// TransactionListAllHandler returns all transaction list of current user
func (a *TransactionsApi) TransactionListAllHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionAllListReq models.TransactionAllListRequest
	err := c.ShouldBindQuery(&transactionAllListReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionListAllHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionListAllHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionListAllHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	allAccountIds, err := a.accounts.GetAccountOrSubAccountIds(c, transactionAllListReq.AccountIds, uid)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionListAllHandler] get account error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allCategoryIds, err := a.transactionCategories.GetCategoryOrSubCategoryIds(c, transactionAllListReq.CategoryIds, uid)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionListAllHandler] get transaction category error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	noTags := transactionAllListReq.TagFilter == models.TransactionNoTagFilterValue
	var tagFilters []*models.TransactionTagFilter

	if !noTags {
		tagFilters, err = models.ParseTransactionTagFilter(transactionAllListReq.TagFilter)

		if err != nil {
			log.Warnf(c, "[transactions.TransactionListAllHandler] parse transaction tag filters error, because %s", err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	maxTransactionTime := int64(math.MaxInt64)
	minTransactionTime := int64(0)

	if transactionAllListReq.EndTime > 0 {
		maxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(transactionAllListReq.EndTime)
	}

	if transactionAllListReq.StartTime > 0 {
		minTransactionTime = utils.GetMinTransactionTimeFromUnixTime(transactionAllListReq.StartTime)
	}

	allTransactions, err := a.transactions.GetAllSpecifiedTransactions(c, uid, maxTransactionTime, minTransactionTime, transactionAllListReq.Type, allCategoryIds, allAccountIds, tagFilters, noTags, transactionAllListReq.AmountFilter, transactionAllListReq.Keyword, transactionAllListReq.MatchMode, transactionAllListReq.MustHavePictures, pageCountForDataExport, true)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionListAllHandler] failed to get all transactions for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	var accountMap map[int64]*models.Account
	var categoryMap map[int64]*models.TransactionCategory
	var tagMap map[int64]*models.TransactionTag
	var allTransactionTagIds map[int64][]int64
	var pictureInfoMap map[int64][]*models.TransactionPictureInfo

	if minTransactionTime == 0 && maxTransactionTime == math.MaxInt64 && len(allCategoryIds) < 1 && len(allAccountIds) < 1 && len(tagFilters) < 1 && transactionAllListReq.AmountFilter == "" && transactionAllListReq.Keyword == "" {
		accountMap, categoryMap, tagMap, allTransactionTagIds, pictureInfoMap, err = a.getTransactionAllEssentialData(c, user, transactionAllListReq.WithPictures, transactionAllListReq.TrimCategory, transactionAllListReq.TrimTag)
	} else {
		accountMap, categoryMap, tagMap, allTransactionTagIds, pictureInfoMap, err = a.getTransactionEssentialDataByTransactionIds(c, user, allTransactions, transactionAllListReq.WithPictures, transactionAllListReq.TrimCategory, transactionAllListReq.TrimTag)
	}

	if err != nil {
		log.Errorf(c, "[transactions.TransactionListAllHandler] failed to get essential data for assembling transaction result for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allTransactions = a.filterTransactions(c, uid, allTransactions, accountMap)
	transactionResult, err := a.getTransactionResponseListResult(c, user, allTransactions, accountMap, categoryMap, tagMap, allTransactionTagIds, pictureInfoMap, clientTimezone, transactionAllListReq.WithPictures, transactionAllListReq.TrimAccount, transactionAllListReq.TrimCategory, transactionAllListReq.TrimTag)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionListAllHandler] failed to assemble transaction result for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return transactionResult, nil
}
