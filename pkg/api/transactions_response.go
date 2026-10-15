package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"sort"
	"time"
)

func (a *TransactionsApi) filterTransactions(c *core.WebContext, uid int64, transactions []*models.Transaction, accountMap map[int64]*models.Account) []*models.Transaction {
	finalTransactions := make([]*models.Transaction, 0, len(transactions))

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]

		if _, exists := accountMap[transaction.AccountId]; !exists {
			log.Warnf(c, "[transactions.filterTransactions] account of transaction \"id:%d\" does not exist for user \"uid:%d\"", transaction.TransactionId, uid)
			continue
		}

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			if _, exists := accountMap[transaction.RelatedAccountId]; !exists {
				log.Warnf(c, "[transactions.filterTransactions] related account of transaction \"id:%d\" does not exist for user \"uid:%d\"", transaction.TransactionId, uid)
				continue
			}
		}

		finalTransactions = append(finalTransactions, transaction)
	}

	return finalTransactions
}

func (a *TransactionsApi) getTransactionTagInfoResponses(tagIds []int64, allTransactionTags map[int64]*models.TransactionTag) []*models.TransactionTagInfoResponse {
	allTags := make([]*models.TransactionTagInfoResponse, 0, len(tagIds))

	for i := 0; i < len(tagIds); i++ {
		tag := allTransactionTags[tagIds[i]]

		if tag == nil {
			continue
		}

		allTags = append(allTags, tag.ToTransactionTagInfoResponse())
	}

	return allTags
}

func (a *TransactionsApi) getTransactionAllEssentialData(c *core.WebContext, user *models.User, withPictures bool, trimCategory bool, trimTag bool) (accountMap map[int64]*models.Account, categoryMap map[int64]*models.TransactionCategory, tagMap map[int64]*models.TransactionTag, allTransactionTagIds map[int64][]int64, pictureInfoMap map[int64][]*models.TransactionPictureInfo, err error) {
	uid := user.Uid
	allAccounts, err := a.accounts.GetAllAccountsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[transactions.getTransactionAllEssentialData] failed to get all accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, nil, nil, nil, nil, err
	}

	accountMap = a.accounts.GetAccountMapByList(allAccounts)

	allTagIndexes, err := a.transactionTags.GetAllTagIdsOfAllTransactions(c, uid)

	if err != nil {
		log.Errorf(c, "[transactions.getTransactionAllEssentialData] failed to get all transactions tag ids for user \"uid:%d\", because %s", uid, err.Error())
		return nil, nil, nil, nil, nil, err
	}

	allTransactionTagIds = a.transactionTags.GetGroupedTransactionTagIds(allTagIndexes)

	if !trimCategory {
		allCategories, err := a.transactionCategories.GetAllCategoriesByUid(c, uid, 0, -1)

		if err != nil {
			log.Errorf(c, "[transactions.getTransactionAllEssentialData] failed to get all transactions categories for user \"uid:%d\", because %s", uid, err.Error())
			return nil, nil, nil, nil, nil, err
		}

		categoryMap = a.transactionCategories.GetCategoryMapByList(allCategories)
	}

	if !trimTag {
		allTags, err := a.transactionTags.GetAllTagsByUid(c, uid)

		if err != nil {
			log.Errorf(c, "[transactions.getTransactionAllEssentialData] failed to get all transactions tags for user \"uid:%d\", because %s", uid, err.Error())
			return nil, nil, nil, nil, nil, err
		}

		tagMap = a.transactionTags.GetTagMapByList(allTags)
	}

	if withPictures && a.CurrentConfig().EnableTransactionPictures {
		pictureInfoMap, err = a.transactionPictures.GetAllPictureInfosOfAllTransactions(c, uid)

		if err != nil {
			log.Errorf(c, "[transactions.getTransactionAllEssentialData] failed to get all transactions pictures for user \"uid:%d\", because %s", uid, err.Error())
			return nil, nil, nil, nil, nil, err
		}
	}

	return accountMap, categoryMap, tagMap, allTransactionTagIds, pictureInfoMap, nil
}

func (a *TransactionsApi) getTransactionEssentialDataByTransactionIds(c *core.WebContext, user *models.User, transactions []*models.Transaction, withPictures bool, trimCategory bool, trimTag bool) (accountMap map[int64]*models.Account, categoryMap map[int64]*models.TransactionCategory, tagMap map[int64]*models.TransactionTag, allTransactionTagIds map[int64][]int64, pictureInfoMap map[int64][]*models.TransactionPictureInfo, err error) {
	uid := user.Uid
	transactionIds := make([]int64, len(transactions))
	accountIds := make([]int64, 0, len(transactions)*2)
	categoryIds := make([]int64, 0, len(transactions))

	for i := 0; i < len(transactions); i++ {
		transactionId := transactions[i].TransactionId

		if transactions[i].Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			transactionId = transactions[i].RelatedId
		}

		transactionIds[i] = transactionId
		accountIds = append(accountIds, transactions[i].AccountId)

		if transactions[i].Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN || transactions[i].Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			accountIds = append(accountIds, transactions[i].RelatedAccountId)
		}

		categoryIds = append(categoryIds, transactions[i].CategoryId)
	}

	accountMap, err = a.accounts.GetAccountsByAccountIds(c, uid, utils.ToUniqueInt64Slice(accountIds))

	if err != nil {
		log.Errorf(c, "[transactions.getTransactionEssentialDataByTransactionIds] failed to get accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, nil, nil, nil, nil, err
	}

	allTransactionTagIds, err = a.transactionTags.GetAllTagIdsOfTransactions(c, uid, transactionIds)

	if err != nil {
		log.Errorf(c, "[transactions.getTransactionEssentialDataByTransactionIds] failed to get transactions tag ids for user \"uid:%d\", because %s", uid, err.Error())
		return nil, nil, nil, nil, nil, err
	}

	if !trimCategory {
		categoryMap, err = a.transactionCategories.GetCategoriesByCategoryIds(c, uid, utils.ToUniqueInt64Slice(categoryIds))

		if err != nil {
			log.Errorf(c, "[transactions.getTransactionEssentialDataByTransactionIds] failed to get transactions categories for user \"uid:%d\", because %s", uid, err.Error())
			return nil, nil, nil, nil, nil, err
		}
	}

	if !trimTag {
		tagMap, err = a.transactionTags.GetTagsByTagIds(c, uid, utils.ToUniqueInt64Slice(a.transactionTags.GetTransactionTagIds(allTransactionTagIds)))

		if err != nil {
			log.Errorf(c, "[transactions.getTransactionEssentialDataByTransactionIds] failed to get transactions tags for user \"uid:%d\", because %s", uid, err.Error())
			return nil, nil, nil, nil, nil, err
		}
	}

	if withPictures && a.CurrentConfig().EnableTransactionPictures {
		pictureInfoMap, err = a.transactionPictures.GetPictureInfosByTransactionIds(c, uid, utils.ToUniqueInt64Slice(a.transactions.GetTransactionIds(transactions)))

		if err != nil {
			log.Errorf(c, "[transactions.getTransactionEssentialDataByTransactionIds] failed to get transactions pictures for user \"uid:%d\", because %s", uid, err.Error())
			return nil, nil, nil, nil, nil, err
		}
	}

	return accountMap, categoryMap, tagMap, allTransactionTagIds, pictureInfoMap, nil
}

func (a *TransactionsApi) getTransactionUsedAccounts(c *core.WebContext, uid int64, transactions []*models.Transaction) (map[int64]*models.Account, error) {
	accountIds := make([]int64, 0, len(transactions)*2)

	for i := 0; i < len(transactions); i++ {
		accountIds = append(accountIds, transactions[i].AccountId)

		if transactions[i].Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN || transactions[i].Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			accountIds = append(accountIds, transactions[i].RelatedAccountId)
		}
	}

	accountMap, err := a.accounts.GetAccountsByAccountIds(c, uid, utils.ToUniqueInt64Slice(accountIds))

	if err != nil {
		log.Errorf(c, "[transactions.getTransactionUsedAccounts] failed to get accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	for i := 0; i < len(transactions); i++ {
		if _, exists := accountMap[transactions[i].AccountId]; !exists {
			log.Warnf(c, "[transactions.getTransactionUsedAccounts] account of transaction \"id:%d\" does not exist for user \"uid:%d\"", transactions[i].TransactionId, uid)
			return nil, errs.ErrSourceAccountNotFound
		}

		if transactions[i].Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN || transactions[i].Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			if _, exists := accountMap[transactions[i].RelatedAccountId]; !exists {
				log.Warnf(c, "[transactions.getTransactionUsedAccounts] related account of transaction \"id:%d\" does not exist for user \"uid:%d\"", transactions[i].TransactionId, uid)
				return nil, errs.ErrDestinationAccountNotFound
			}
		}
	}

	return accountMap, nil
}

func (a *TransactionsApi) getTransactionResponseListResult(c *core.WebContext, user *models.User, transactions []*models.Transaction, allAccounts map[int64]*models.Account, categoryMap map[int64]*models.TransactionCategory, tagMap map[int64]*models.TransactionTag, allTransactionTagIds map[int64][]int64, pictureInfoMap map[int64][]*models.TransactionPictureInfo, clientTimezone *time.Location, withPictures bool, trimAccount bool, trimCategory bool, trimTag bool) (models.TransactionInfoResponseSlice, error) {
	result := make(models.TransactionInfoResponseSlice, len(transactions))

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			transaction = a.transactions.GetRelatedTransferTransaction(transaction)
		}

		transactionEditable := transaction.IsEditable(user, clientTimezone, allAccounts[transaction.AccountId], allAccounts[transaction.RelatedAccountId])
		transactionTagIds := allTransactionTagIds[transaction.TransactionId]
		result[i] = transaction.ToTransactionInfoResponse(transactionTagIds, transactionEditable)

		if !trimAccount {
			if sourceAccount := allAccounts[transaction.AccountId]; sourceAccount != nil {
				result[i].SourceAccount = sourceAccount.ToAccountInfoResponse()
			}

			if destinationAccount := allAccounts[transaction.RelatedAccountId]; destinationAccount != nil {
				result[i].DestinationAccount = destinationAccount.ToAccountInfoResponse()
			}
		}

		if !trimCategory && categoryMap != nil {
			if category := categoryMap[transaction.CategoryId]; category != nil {
				result[i].Category = category.ToTransactionCategoryInfoResponse()
			}
		}

		if !trimTag && tagMap != nil && transactionTagIds != nil {
			result[i].Tags = a.getTransactionTagInfoResponses(transactionTagIds, tagMap)
		}

		if withPictures && a.CurrentConfig().EnableTransactionPictures && pictureInfoMap != nil {
			pictureInfos, exists := pictureInfoMap[transaction.TransactionId]

			if exists {
				result[i].Pictures = a.GetTransactionPictureInfoResponseList(pictureInfos)
			}
		}
	}

	sort.Sort(result)

	return result, nil
}

func (a *TransactionsApi) createNewTransactionModel(uid int64, transactionCreateReq *models.TransactionCreateRequest, clientIp string) *models.Transaction {
	var transactionDbType models.TransactionDbType

	if transactionCreateReq.Type == models.TRANSACTION_TYPE_MODIFY_BALANCE {
		transactionDbType = models.TRANSACTION_DB_TYPE_MODIFY_BALANCE
	} else if transactionCreateReq.Type == models.TRANSACTION_TYPE_EXPENSE {
		transactionDbType = models.TRANSACTION_DB_TYPE_EXPENSE
	} else if transactionCreateReq.Type == models.TRANSACTION_TYPE_INCOME {
		transactionDbType = models.TRANSACTION_DB_TYPE_INCOME
	} else if transactionCreateReq.Type == models.TRANSACTION_TYPE_TRANSFER {
		transactionDbType = models.TRANSACTION_DB_TYPE_TRANSFER_OUT
	}

	transaction := &models.Transaction{
		Uid:               uid,
		Type:              transactionDbType,
		CategoryId:        transactionCreateReq.CategoryId,
		TransactionTime:   utils.GetMinTransactionTimeFromUnixTime(transactionCreateReq.Time),
		TimezoneUtcOffset: transactionCreateReq.UtcOffset,
		AccountId:         transactionCreateReq.SourceAccountId,
		Amount:            transactionCreateReq.SourceAmount,
		HideAmount:        transactionCreateReq.HideAmount,
		Comment:           transactionCreateReq.Comment,
		CreatedIp:         clientIp,
	}

	if transactionCreateReq.Type == models.TRANSACTION_TYPE_TRANSFER {
		transaction.RelatedAccountId = transactionCreateReq.DestinationAccountId
		transaction.RelatedAccountAmount = transactionCreateReq.DestinationAmount
	}

	if transactionCreateReq.GeoLocation != nil {
		transaction.GeoLongitude = transactionCreateReq.GeoLocation.Longitude
		transaction.GeoLatitude = transactionCreateReq.GeoLocation.Latitude
	}

	return transaction
}
