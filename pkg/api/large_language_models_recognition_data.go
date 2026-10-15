package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"time"
)

func (a *LargeLanguageModelsApi) getUserEssentialData(c *core.WebContext, uid int64) ([]string, map[string]*models.Account, []string, []string, []string, map[string]*models.TransactionCategory, map[string]*models.TransactionCategory, map[string]*models.TransactionCategory, []string, map[string]*models.TransactionTag, error) {
	accounts, err := a.accounts.GetAllAccountsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[large_language_models.getUserEssentialData] failed to get all accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	accountMap := a.accounts.GetVisibleAccountNameMapByList(accounts)
	accountNames := make([]string, 0, len(accounts))

	for i := 0; i < len(accounts); i++ {
		if accounts[i].Hidden || accounts[i].Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
			continue
		}

		accountNames = append(accountNames, accounts[i].Name)
	}

	categories, err := a.transactionCategories.GetAllCategoriesByUid(c, uid, 0, -1)

	if err != nil {
		log.Errorf(c, "[large_language_models.getUserEssentialData] failed to get categories for user \"uid:%d\", because %s", uid, err.Error())
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	incomeCategoryMap := make(map[string]*models.TransactionCategory)
	incomeCategoryNames := make([]string, 0)

	expenseCategoryMap := make(map[string]*models.TransactionCategory)
	expenseCategoryNames := make([]string, 0)

	transferCategoryMap := make(map[string]*models.TransactionCategory)
	transferCategoryNames := make([]string, 0)

	for i := 0; i < len(categories); i++ {
		category := categories[i]

		if category.Hidden || category.ParentCategoryId == models.LevelOneTransactionCategoryParentId {
			continue
		}

		if category.Type == models.CATEGORY_TYPE_INCOME {
			incomeCategoryMap[category.Name] = category
			incomeCategoryNames = append(incomeCategoryNames, category.Name)
		} else if category.Type == models.CATEGORY_TYPE_EXPENSE {
			expenseCategoryMap[category.Name] = category
			expenseCategoryNames = append(expenseCategoryNames, category.Name)
		} else if category.Type == models.CATEGORY_TYPE_TRANSFER {
			transferCategoryMap[category.Name] = category
			transferCategoryNames = append(transferCategoryNames, category.Name)
		}
	}

	tags, err := a.transactionTags.GetAllTagsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[large_language_models.getUserEssentialData] failed to get tags for user \"uid:%d\", because %s", uid, err.Error())
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	tagMap := a.transactionTags.GetVisibleTagNameMapByList(tags)
	tagNames := make([]string, 0, len(tags))

	for i := 0; i < len(tags); i++ {
		if tags[i].Hidden {
			continue
		}

		tagNames = append(tagNames, tags[i].Name)
	}

	return accountNames, accountMap, incomeCategoryNames, expenseCategoryNames, transferCategoryNames, incomeCategoryMap, expenseCategoryMap, transferCategoryMap, tagNames, tagMap, nil
}

func (a *LargeLanguageModelsApi) parseRecognizedTransactionResponse(c *core.WebContext, uid int64, clientTimezone *time.Location, recognizedResult *models.RecognizedTransactionResult, accountMap map[string]*models.Account, expenseCategoryMap map[string]*models.TransactionCategory, incomeCategoryMap map[string]*models.TransactionCategory, transferCategoryMap map[string]*models.TransactionCategory, tagMap map[string]*models.TransactionTag) (*models.RecognizedTransactionResponse, *errs.Error) {
	recognizedTransactionResponse := &models.RecognizedTransactionResponse{
		Type: models.TRANSACTION_TYPE_EXPENSE,
	}

	if recognizedResult == nil {
		log.Errorf(c, "[large_language_models.parseRecognizedTransactionResponse] recoginzed result is null")
		return nil, errs.ErrNoTransactionInformation
	}

	if recognizedResult.Type == "income" {
		recognizedTransactionResponse.Type = models.TRANSACTION_TYPE_INCOME

		if len(recognizedResult.CategoryName) > 0 {
			category, exists := incomeCategoryMap[recognizedResult.CategoryName]

			if exists {
				recognizedTransactionResponse.CategoryId = category.CategoryId
			}
		}
	} else if recognizedResult.Type == "expense" {
		recognizedTransactionResponse.Type = models.TRANSACTION_TYPE_EXPENSE

		if len(recognizedResult.CategoryName) > 0 {
			category, exists := expenseCategoryMap[recognizedResult.CategoryName]

			if exists {
				recognizedTransactionResponse.CategoryId = category.CategoryId
			}
		}
	} else if recognizedResult.Type == "transfer" {
		recognizedTransactionResponse.Type = models.TRANSACTION_TYPE_TRANSFER

		if len(recognizedResult.CategoryName) > 0 {
			category, exists := transferCategoryMap[recognizedResult.CategoryName]

			if exists {
				recognizedTransactionResponse.CategoryId = category.CategoryId
			}
		}
	} else if len(recognizedResult.Type) == 0 {
		return nil, errs.ErrNoTransactionInformation
	} else {
		log.Errorf(c, "[large_language_models.parseRecognizedTransactionResponse] recoginzed transaction type \"%s\" is invalid", recognizedResult.Type)
		return nil, errs.ErrOperationFailed
	}

	if len(recognizedResult.Time) > 0 {
		longDateTime := a.getLongDateTime(recognizedResult.Time)
		timestamp, err := utils.ParseFromLongDateTimeInTimeZone(longDateTime, clientTimezone)

		if err != nil {
			log.Warnf(c, "[large_language_models.parseRecognizedTransactionResponse] recoginzed time \"%s\" is invalid", recognizedResult.Time)
		} else {
			recognizedTransactionResponse.Time = timestamp.Unix()
		}
	}

	if len(recognizedResult.Amount) > 0 {
		amount, err := utils.ParseAmount(recognizedResult.Amount)

		if err != nil {
			log.Errorf(c, "[large_language_models.parseRecognizedTransactionResponse] recoginzed amount \"%s\" is invalid", recognizedResult.Amount)
			return nil, errs.ErrOperationFailed
		}

		recognizedTransactionResponse.SourceAmount = amount

		if recognizedTransactionResponse.Type == models.TRANSACTION_TYPE_TRANSFER && len(recognizedResult.DestinationAmount) > 0 {
			destinationAmount, err := utils.ParseAmount(recognizedResult.DestinationAmount)

			if err != nil {
				log.Errorf(c, "[large_language_models.parseRecognizedTransactionResponse] recoginzed destination amount \"%s\" is invalid", recognizedResult.DestinationAmount)
				return nil, errs.ErrOperationFailed
			}

			recognizedTransactionResponse.DestinationAmount = destinationAmount
		}
	}

	if len(recognizedResult.AccountName) > 0 {
		account, exists := accountMap[recognizedResult.AccountName]

		if exists {
			recognizedTransactionResponse.SourceAccountId = account.AccountId
		}
	}

	if len(recognizedResult.DestinationAccountName) > 0 {
		account, exists := accountMap[recognizedResult.DestinationAccountName]

		if exists {
			recognizedTransactionResponse.DestinationAccountId = account.AccountId
		}
	}

	if len(recognizedResult.TagNames) > 0 {
		tagIds := make([]string, 0, len(recognizedResult.TagNames))

		for i := 0; i < len(recognizedResult.TagNames); i++ {
			tagName := recognizedResult.TagNames[i]
			tag, exists := tagMap[tagName]

			if exists {
				tagIds = append(tagIds, utils.Int64ToString(tag.TagId))
			}
		}

		recognizedTransactionResponse.TagIds = tagIds
	}

	if len(recognizedResult.Description) > 0 {
		recognizedTransactionResponse.Comment = recognizedResult.Description
	}

	return recognizedTransactionResponse, nil
}

func (a *LargeLanguageModelsApi) getLongDateTime(dateTime string) string {
	if utils.IsValidLongDateTimeFormat(dateTime) {
		return dateTime
	}

	if utils.IsValidLongDateTimeWithoutSecondFormat(dateTime) {
		return dateTime + ":00"
	}

	if utils.IsValidLongDateFormat(dateTime) {
		return dateTime + " 00:00:00"
	}

	return dateTime
}
