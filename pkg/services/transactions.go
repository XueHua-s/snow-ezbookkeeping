package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
	"time"
)

const pageCountForLoadTransactionAmounts = 1000

// TransactionService represents transaction service
type TransactionService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a transaction service singleton instance
var (
	Transactions = &TransactionService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetTotalTransactionCountByUid returns total transaction count of user
func (s *TransactionService) GetTotalTransactionCountByUid(c core.Context, uid int64) (int64, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	count, err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).Count(&models.Transaction{})

	return count, err
}

// GetAllTransactions returns all transactions
func (s *TransactionService) GetAllTransactions(c core.Context, uid int64, pageCount int32, noDuplicated bool) ([]*models.Transaction, error) {
	maxTransactionTime := utils.GetMaxTransactionTimeFromUnixTime(time.Now().Unix())
	var allTransactions []*models.Transaction

	for maxTransactionTime > 0 {
		transactions, err := s.GetTransactionsByMaxTime(c, uid, maxTransactionTime, 0, 0, nil, nil, nil, false, "", "", core.MATCH_MODE_DEFAULT, false, 1, pageCount, false, noDuplicated)

		if err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, transactions...)

		if len(transactions) < int(pageCount) {
			maxTransactionTime = 0
			break
		}

		maxTransactionTime = transactions[len(transactions)-1].TransactionTime - 1
	}

	return allTransactions, nil
}

// GetAllSpecifiedTransactions returns all transactions that match given conditions
func (s *TransactionService) GetAllSpecifiedTransactions(c core.Context, uid int64, maxTransactionTime int64, minTransactionTime int64, transactionType models.TransactionType, categoryIds []int64, accountIds []int64, tagFilters []*models.TransactionTagFilter, noTags bool, amountFilter string, keyword string, matchMode core.MatchMode, mustHavePictures bool, pageCount int32, noDuplicated bool) ([]*models.Transaction, error) {
	if maxTransactionTime <= 0 {
		maxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(time.Now().Unix())
	}

	var allTransactions []*models.Transaction

	for maxTransactionTime > 0 {
		transactions, err := s.GetTransactionsByMaxTime(c, uid, maxTransactionTime, minTransactionTime, transactionType, categoryIds, accountIds, tagFilters, noTags, amountFilter, keyword, matchMode, mustHavePictures, 1, pageCount, false, noDuplicated)

		if err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, transactions...)

		if len(transactions) < int(pageCount) {
			maxTransactionTime = 0
			break
		}

		maxTransactionTime = transactions[len(transactions)-1].TransactionTime - 1
	}

	return allTransactions, nil
}
