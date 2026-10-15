package cli

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func (l *UserDataCli) getUserIdByUsername(c *core.CliContext, username string) (int64, error) {
	user, err := l.GetUserByUsername(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.getUserIdByUsername] failed to get user by user name \"%s\", because %s", username, err.Error())
		return 0, err
	}

	return user.Uid, nil
}

func (l *UserDataCli) getUserEssentialData(c *core.CliContext, uid int64, username string) (accountMap map[int64]*models.Account, categoryMap map[int64]*models.TransactionCategory, tagMap map[int64]*models.TransactionTag, tagIndexes []*models.TransactionTagIndex, tagIndexesMap map[int64][]int64, err error) {
	if uid <= 0 {
		log.CliErrorf(c, "[user_data.getUserEssentialData] user uid \"%d\" is invalid", uid)
		return nil, nil, nil, nil, nil, errs.ErrUserIdInvalid
	}

	accounts, err := l.accounts.GetAllAccountsByUid(c, uid)

	if err != nil {
		log.CliErrorf(c, "[user_data.getUserEssentialData] failed to get accounts for user \"%s\", because %s", username, err.Error())
		return nil, nil, nil, nil, nil, err
	}

	accountMap = l.accounts.GetAccountMapByList(accounts)

	categories, err := l.categories.GetAllCategoriesByUid(c, uid, 0, -1)

	if err != nil {
		log.CliErrorf(c, "[user_data.getUserEssentialData] failed to get categories for user \"%s\", because %s", username, err.Error())
		return nil, nil, nil, nil, nil, err
	}

	categoryMap = l.categories.GetCategoryMapByList(categories)

	tags, err := l.tags.GetAllTagsByUid(c, uid)

	if err != nil {
		log.CliErrorf(c, "[user_data.getUserEssentialData] failed to get tags for user \"%s\", because %s", username, err.Error())
		return nil, nil, nil, nil, nil, err
	}

	tagMap = l.tags.GetTagMapByList(tags)

	tagIndexes, err = l.tags.GetAllTagIdsOfAllTransactions(c, uid)

	if err != nil {
		log.CliErrorf(c, "[user_data.getUserEssentialData] failed to get tag index for user \"%s\", because %s", username, err.Error())
		return nil, nil, nil, nil, nil, err
	}

	tagIndexesMap = l.tags.GetGroupedTransactionTagIds(tagIndexes)

	return accountMap, categoryMap, tagMap, tagIndexes, tagIndexesMap, nil
}

func (l *UserDataCli) getUserEssentialDataForImport(c *core.CliContext, uid int64, username string) (accountMap map[string]*models.Account, expenseCategoryMap map[string]map[string]*models.TransactionCategory, incomeCategoryMap map[string]map[string]*models.TransactionCategory, transferCategoryMap map[string]map[string]*models.TransactionCategory, tagMap map[string]*models.TransactionTag, err error) {
	if uid <= 0 {
		log.CliErrorf(c, "[user_data.getUserEssentialDataForImport] user uid \"%d\" is invalid", uid)
		return nil, nil, nil, nil, nil, errs.ErrUserIdInvalid
	}

	accounts, err := l.accounts.GetAllAccountsByUid(c, uid)

	if err != nil {
		log.CliErrorf(c, "[user_data.getUserEssentialDataForImport] failed to get accounts for user \"%s\", because %s", username, err.Error())
		return nil, nil, nil, nil, nil, err
	}

	accountMap = l.accounts.GetVisibleAccountNameMapByList(accounts)

	categories, err := l.categories.GetAllCategoriesByUid(c, uid, 0, -1)

	if err != nil {
		log.CliErrorf(c, "[user_data.getUserEssentialDataForImport] failed to get categories for user \"%s\", because %s", username, err.Error())
		return nil, nil, nil, nil, nil, err
	}

	expenseCategoryMap, incomeCategoryMap, transferCategoryMap = l.categories.GetVisibleSubCategoryNameMapByList(categories)

	tags, err := l.tags.GetAllTagsByUid(c, uid)

	if err != nil {
		log.CliErrorf(c, "[user_data.getUserEssentialDataForImport] failed to get tags for user \"%s\", because %s", username, err.Error())
		return nil, nil, nil, nil, nil, err
	}

	tagMap = l.tags.GetVisibleTagNameMapByList(tags)

	return accountMap, expenseCategoryMap, incomeCategoryMap, transferCategoryMap, tagMap, nil
}
