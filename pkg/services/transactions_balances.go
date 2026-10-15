package services

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"math/big"
	"strings"
	"time"
)

// GetAllTransactionsInOneAccountWithAccountBalanceByMaxTime returns account statement within time range
func (s *TransactionService) GetAllTransactionsInOneAccountWithAccountBalanceByMaxTime(c core.Context, uid int64, pageCount int32, maxTransactionTime int64, minTransactionTime int64, accountId int64, accountCategory models.AccountCategory) ([]*models.TransactionWithAccountBalance, *big.Int, *big.Int, *big.Int, *big.Int, error) {
	if maxTransactionTime <= 0 {
		maxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(time.Now().Unix())
	}

	var allTransactions []*models.Transaction

	for maxTransactionTime > 0 {
		transactions, err := s.GetTransactionsByMaxTime(c, uid, maxTransactionTime, 0, 0, nil, []int64{accountId}, nil, false, "", "", core.MATCH_MODE_DEFAULT, false, 1, pageCount, false, true)

		if err != nil {
			return nil, nil, nil, nil, nil, err
		}

		allTransactions = append(allTransactions, transactions...)

		if len(transactions) < int(pageCount) {
			maxTransactionTime = 0
			break
		}

		maxTransactionTime = transactions[len(transactions)-1].TransactionTime - 1
	}

	allTransactionsAndAccountBalance := make([]*models.TransactionWithAccountBalance, 0, len(allTransactions))

	if len(allTransactions) < 1 {
		return allTransactionsAndAccountBalance, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), nil
	}

	totalInflows := big.NewInt(0)
	totalOutflows := big.NewInt(0)
	openingBalance := big.NewInt(0)
	accumulatedBalance := big.NewInt(0)
	lastAccumulatedBalance := big.NewInt(0)

	for i := len(allTransactions) - 1; i >= 0; i-- {
		transaction := allTransactions[i]

		if transaction.Type == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
			accumulatedBalance.Add(accumulatedBalance, big.NewInt(transaction.RelatedAccountAmount))
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_INCOME {
			accumulatedBalance.Add(accumulatedBalance, big.NewInt(transaction.Amount))
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_EXPENSE {
			accumulatedBalance.Sub(accumulatedBalance, big.NewInt(transaction.Amount))
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			accumulatedBalance.Sub(accumulatedBalance, big.NewInt(transaction.Amount))
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			accumulatedBalance.Add(accumulatedBalance, big.NewInt(transaction.Amount))
		} else {
			log.Errorf(c, "[transactions.GetAllTransactionsInOneAccountWithAccountBalanceByMaxTime] trasaction type (%d) is invalid (id:%d)", transaction.TransactionId, transaction.Type)
			return nil, nil, nil, nil, nil, errs.ErrTransactionTypeInvalid
		}

		if transaction.TransactionTime < minTransactionTime {
			openingBalance.Set(accumulatedBalance)
			lastAccumulatedBalance.Set(accumulatedBalance)
			continue
		}

		if transaction.Type == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
			if accountCategory.IsAsset() {
				totalInflows.Add(totalInflows, big.NewInt(transaction.RelatedAccountAmount))
			} else if accountCategory.IsLiability() {
				totalOutflows.Sub(totalOutflows, big.NewInt(transaction.RelatedAccountAmount))
			}
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_INCOME {
			totalInflows.Add(totalInflows, big.NewInt(transaction.Amount))
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_EXPENSE {
			totalOutflows.Add(totalOutflows, big.NewInt(transaction.Amount))
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			totalOutflows.Add(totalOutflows, big.NewInt(transaction.Amount))
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			totalInflows.Add(totalInflows, big.NewInt(transaction.Amount))
		}

		transactionsAndAccountBalance := &models.TransactionWithAccountBalance{
			Transaction:           transaction,
			AccountOpeningBalance: new(big.Int).Set(lastAccumulatedBalance),
			AccountClosingBalance: new(big.Int).Set(accumulatedBalance),
		}

		lastAccumulatedBalance.Set(accumulatedBalance)
		allTransactionsAndAccountBalance = append(allTransactionsAndAccountBalance, transactionsAndAccountBalance)
	}

	return allTransactionsAndAccountBalance, totalInflows, totalOutflows, openingBalance, new(big.Int).Set(accumulatedBalance), nil
}

// GetAllAccountsDailyOpeningAndClosingBalance returns daily opening and closing balance of all accounts within time range
func (s *TransactionService) GetAllAccountsDailyOpeningAndClosingBalance(c core.Context, uid int64, maxTransactionTime int64, minTransactionTime int64, clientTimezone *time.Location) (map[int32][]*models.TransactionWithAccountBalance, error) {
	if maxTransactionTime <= 0 {
		maxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(time.Now().Unix())
	}

	var allTransactions []*models.Transaction

	for maxTransactionTime > 0 {
		transactions, err := s.GetTransactionsByMaxTime(c, uid, maxTransactionTime, 0, 0, nil, nil, nil, false, "", "", core.MATCH_MODE_DEFAULT, false, 1, pageCountForLoadTransactionAmounts, false, false)

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

	accountDailyLastBalances := make(map[string]*models.TransactionWithAccountBalance)
	accountDailyBalances := make(map[int32][]*models.TransactionWithAccountBalance)

	if len(allTransactions) < 1 {
		return accountDailyBalances, nil
	}

	accumulatedBalances := make(map[int64]*big.Int)
	accumulatedBalancesBeforeStartTime := make(map[int64]*big.Int)

	for i := len(allTransactions) - 1; i >= 0; i-- {
		transaction := allTransactions[i]
		accumulatedBalance := accumulatedBalances[transaction.AccountId]

		if accumulatedBalance == nil {
			accumulatedBalance = big.NewInt(0)
		}

		lastAccumulatedBalance := new(big.Int).Set(accumulatedBalance)

		if transaction.Type == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
			accumulatedBalance.Add(accumulatedBalance, big.NewInt(transaction.RelatedAccountAmount))
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_INCOME {
			accumulatedBalance.Add(accumulatedBalance, big.NewInt(transaction.Amount))
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_EXPENSE {
			accumulatedBalance.Sub(accumulatedBalance, big.NewInt(transaction.Amount))
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			accumulatedBalance.Sub(accumulatedBalance, big.NewInt(transaction.Amount))
		} else if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			accumulatedBalance.Add(accumulatedBalance, big.NewInt(transaction.Amount))
		} else {
			log.Errorf(c, "[transactions.GetAllTransactionsWithAccountBalanceByMaxTime] trasaction type (%d) is invalid (id:%d)", transaction.TransactionId, transaction.Type)
			return nil, errs.ErrTransactionTypeInvalid
		}

		accumulatedBalances[transaction.AccountId] = accumulatedBalance

		if transaction.TransactionTime < minTransactionTime {
			accumulatedBalancesBeforeStartTime[transaction.AccountId] = new(big.Int).Set(accumulatedBalance)
			continue
		}

		yearMonthDay := utils.FormatUnixTimeToNumericYearMonthDay(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime), clientTimezone)
		groupKey := fmt.Sprintf("%d_%d", yearMonthDay, transaction.AccountId)
		dailyAccountBalance, exists := accountDailyLastBalances[groupKey]

		if exists {
			dailyAccountBalance.AccountClosingBalance.Set(accumulatedBalance)
		} else {
			dailyAccountBalance = &models.TransactionWithAccountBalance{
				Transaction: &models.Transaction{
					AccountId: transaction.AccountId,
				},
				AccountOpeningBalance: lastAccumulatedBalance,
				AccountClosingBalance: new(big.Int).Set(accumulatedBalance),
			}
			accountDailyLastBalances[groupKey] = dailyAccountBalance
		}
	}

	firstTransactionTime := allTransactions[len(allTransactions)-1].TransactionTime

	if minTransactionTime > firstTransactionTime {
		firstTransactionTime = minTransactionTime
	}

	firstYearMonthDay := utils.FormatUnixTimeToNumericYearMonthDay(utils.GetUnixTimeFromTransactionTime(firstTransactionTime), clientTimezone)

	// fill in the opening balance for accounts that do not have transactions on the first day
	for accountId, accumulatedBalance := range accumulatedBalancesBeforeStartTime {
		if accumulatedBalance.Sign() == 0 {
			continue
		}

		groupKey := fmt.Sprintf("%d_%d", firstYearMonthDay, accountId)

		if _, exists := accountDailyLastBalances[groupKey]; exists {
			continue
		}

		accountDailyLastBalances[groupKey] = &models.TransactionWithAccountBalance{
			Transaction: &models.Transaction{
				AccountId: accountId,
			},
			AccountOpeningBalance: new(big.Int).Set(accumulatedBalance),
			AccountClosingBalance: new(big.Int).Set(accumulatedBalance),
		}
	}

	for groupKey, transactionWithAccountBalance := range accountDailyLastBalances {
		groupKeyParts := strings.Split(groupKey, "_")
		yearMonthDay, _ := utils.StringToInt32(groupKeyParts[0])
		dailyAccountBalances, exists := accountDailyBalances[yearMonthDay]

		if !exists {
			dailyAccountBalances = make([]*models.TransactionWithAccountBalance, 0)
		}

		dailyAccountBalances = append(dailyAccountBalances, transactionWithAccountBalance)
		accountDailyBalances[yearMonthDay] = dailyAccountBalances
	}

	return accountDailyBalances, nil
}
