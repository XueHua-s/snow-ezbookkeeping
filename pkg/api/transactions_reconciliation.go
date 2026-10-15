package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"math/big"
	"strings"
)

// TransactionReconciliationStatementHandler returns transaction reconciliation statement list of current user
func (a *TransactionsApi) TransactionReconciliationStatementHandler(c *core.WebContext) (any, *errs.Error) {
	var reconciliationStatementRequest models.TransactionReconciliationStatementRequest
	err := c.ShouldBindQuery(&reconciliationStatementRequest)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionReconciliationStatementHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionReconciliationStatementHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionReconciliationStatementHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	account, err := a.accounts.GetAccountByAccountId(c, uid, reconciliationStatementRequest.AccountId)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionReconciliationStatementHandler] failed to get account \"id:%d\" for user \"uid:%d\", because %s", reconciliationStatementRequest.AccountId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if account.Type != models.ACCOUNT_TYPE_SINGLE_ACCOUNT {
		log.Errorf(c, "[transactions.TransactionReconciliationStatementHandler] account \"id:%d\" for user \"uid:%d\" is not a single account", reconciliationStatementRequest.AccountId, uid)
		return nil, errs.ErrAccountTypeInvalid
	}

	maxTransactionTime := int64(0)

	if reconciliationStatementRequest.EndTime > 0 {
		maxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(reconciliationStatementRequest.EndTime)
	}

	minTransactionTime := int64(0)

	if reconciliationStatementRequest.StartTime > 0 {
		minTransactionTime = utils.GetMinTransactionTimeFromUnixTime(reconciliationStatementRequest.StartTime)
	}

	transactionsWithAccountBalance, totalInflows, totalOutflows, openingBalance, closingBalance, err := a.transactions.GetAllTransactionsInOneAccountWithAccountBalanceByMaxTime(c, uid, pageCountForAccountStatement, maxTransactionTime, minTransactionTime, reconciliationStatementRequest.AccountId, account.Category)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionReconciliationStatementHandler] failed to get transactions from \"%d\" to \"%d\" for user \"uid:%d\", because %s", reconciliationStatementRequest.StartTime, reconciliationStatementRequest.EndTime, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactions := make([]*models.Transaction, len(transactionsWithAccountBalance))
	transactionAccountBalanceMap := make(map[int64]*models.TransactionWithAccountBalance, len(transactionsWithAccountBalance))

	for i := 0; i < len(transactionsWithAccountBalance); i++ {
		transactionWithBalance := transactionsWithAccountBalance[i]
		transactions[i] = transactionWithBalance.Transaction
		transactionAccountBalanceMap[transactionWithBalance.TransactionId] = transactionWithBalance
		transactionAccountBalanceMap[transactionWithBalance.RelatedId] = transactionWithBalance
	}

	allAccountIds := make([]int64, 0, len(transactions)*2)

	for i := 0; i < len(transactions); i++ {
		allAccountIds = append(allAccountIds, transactions[i].AccountId)

		if transactions[i].Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN || transactions[i].Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			allAccountIds = append(allAccountIds, transactions[i].RelatedAccountId)
		}
	}

	allAccounts, err := a.accounts.GetAccountsByAccountIds(c, uid, utils.ToUniqueInt64Slice(allAccountIds))

	if err != nil {
		log.Errorf(c, "[transactions.TransactionReconciliationStatementHandler] failed to get essential data for assembling transaction result for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactionResult, err := a.getTransactionResponseListResult(c, user, transactions, allAccounts, nil, nil, nil, nil, clientTimezone, false, true, true, true)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionReconciliationStatementHandler] failed to assemble transaction result for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	responseItems := make([]*models.TransactionReconciliationStatementResponseItem, len(transactionResult))

	for i := 0; i < len(transactionResult); i++ {
		transactionResult := transactionResult[i]
		accountOpeningBalance := big.NewInt(0)
		accountClosingBalance := big.NewInt(0)

		if transactionWithBalance, exists := transactionAccountBalanceMap[transactionResult.Id]; exists {
			accountOpeningBalance = transactionWithBalance.AccountOpeningBalance
			accountClosingBalance = transactionWithBalance.AccountClosingBalance
		} else {
			log.Warnf(c, "[transactions.TransactionReconciliationStatementHandler] missing account balance for transaction \"id:%d\" of user \"uid:%d\"", transactionResult.Id, uid)
		}

		responseItems[i] = &models.TransactionReconciliationStatementResponseItem{
			TransactionInfoResponse: transactionResult,
			AccountOpeningBalance:   accountOpeningBalance.String(),
			AccountClosingBalance:   accountClosingBalance.String(),
		}
	}

	reconciliationStatementResp := &models.TransactionReconciliationStatementResponse{
		Transactions:   responseItems,
		TotalInflows:   totalInflows.String(),
		TotalOutflows:  totalOutflows.String(),
		OpeningBalance: openingBalance.String(),
		ClosingBalance: closingBalance.String(),
	}

	return reconciliationStatementResp, nil
}

// TransactionUnreconciledTransactionCountsHandler returns unreconciled transaction counts of current user's accounts
func (a *TransactionsApi) TransactionUnreconciledTransactionCountsHandler(c *core.WebContext) (any, *errs.Error) {
	var unreconciledCountReq models.TransactionUnreconciledCountRequest
	err := c.ShouldBindQuery(&unreconciledCountReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionUnreconciledTransactionCountsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	selectedAccountIds := make(map[int64]bool)

	if unreconciledCountReq.AccountIds != "" {
		accountIds, err := utils.StringArrayToInt64Array(strings.Split(unreconciledCountReq.AccountIds, ","))

		if err != nil {
			return nil, errs.ErrAccountIdInvalid
		}

		for _, accountId := range accountIds {
			selectedAccountIds[accountId] = true
		}
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionUnreconciledTransactionCountsHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if !user.UseLastReconciledTime {
		return nil, errs.ErrLastReconciledTimeIsNotEnabled
	}

	accounts, err := a.accounts.GetAllAccountsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionUnreconciledTransactionCountsHandler] failed to get all accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	accountLastReconciledTimes := make(map[int64]int64, len(accounts))

	for _, account := range accounts {
		if account.Hidden || account.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
			continue
		}

		if len(selectedAccountIds) > 0 && !selectedAccountIds[account.AccountId] {
			continue
		}

		if account.Extend != nil && account.Extend.LastReconciledTime != nil && *account.Extend.LastReconciledTime > 0 {
			accountLastReconciledTimes[account.AccountId] = *account.Extend.LastReconciledTime
		} else {
			accountLastReconciledTimes[account.AccountId] = 0
		}
	}

	counts, err := a.transactions.GetUnreconciledTransactionCounts(c, uid, accountLastReconciledTimes)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionUnreconciledTransactionCountsHandler] failed to get unreconciled transaction counts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return &models.TransactionUnreconciledCountResponse{
		Items: counts,
	}, nil
}
