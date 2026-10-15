package beancount

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"strings"
)

func (r *beancountDataReader) readAndSetOption(ctx core.Context, recordId string, items []string) {
	if r.getNotEmptyItemsCount(items) != 3 {
		log.Warnf(ctx, "[beancount_data_reader.readAndSetOption] cannot parse account type name option %s \"%s\", because items count in line not correct", recordId, strings.Join(items, " "))
		return
	}

	optionName := r.getNotEmptyItemByIndex(items, 1)
	optionValue := r.getNotEmptyItemByIndex(items, 2)

	switch optionName {
	case beancountOptionAssetsAccountTypeName:
		r.setAccountTypeNameMap(beancountAssetsAccountType, optionValue)
		break
	case beancountOptionLiabilitiesAccountTypeName:
		r.setAccountTypeNameMap(beancountLiabilitiesAccountType, optionValue)
		break
	case beancountOptionEquityAccountTypeName:
		r.setAccountTypeNameMap(beancountEquityAccountType, optionValue)
		break
	case beancountOptionIncomeAccountTypeName:
		r.setAccountTypeNameMap(beancountIncomeAccountType, optionValue)
		break
	case beancountOptionExpenseAccountTypeName:
		r.setAccountTypeNameMap(beancountExpensesAccountType, optionValue)
		break
	default:
		log.Warnf(ctx, "[beancount_data_reader.readAndSetOption] skip option %s \"%s\"", recordId, strings.Join(items, " "))
		break
	}
}

func (r *beancountDataReader) readAndSetTags(ctx core.Context, recordId string, items []string, currentTags []string, pushTag bool) []string {
	if r.getNotEmptyItemsCount(items) != 2 {
		log.Warnf(ctx, "[beancount_data_reader.readAndSetTags] cannot parse push/pop tag %s \"%s\", because items count in line not correct", recordId, strings.Join(items, " "))
		return currentTags
	}

	tag := r.getNotEmptyItemByIndex(items, 1)

	if len(tag) < 2 || tag[0] != beancountTagPrefix {
		log.Warnf(ctx, "[beancount_data_reader.readAndSetTags] cannot parse push/pop tag %s \"%s\", because tag is invalid", recordId, strings.Join(items, " "))
		return currentTags
	}

	tag = tag[1:]

	if pushTag {
		for i := 0; i < len(currentTags); i++ {
			if currentTags[i] == tag {
				return currentTags
			}
		}

		return append(currentTags, tag)
	} else { // pop tag
		for i := 0; i < len(currentTags); i++ {
			if currentTags[i] == tag {
				return append(currentTags[:i], currentTags[i+1:]...)
			}
		}

		return currentTags
	}
}

func (r *beancountDataReader) setAccountTypeNameMap(accountType beancountAccountType, accountTypeName string) {
	delete(r.accountTypeNameMap, r.accountTypeNameReversedMap[accountType])
	r.accountTypeNameMap[accountTypeName] = accountType
	r.accountTypeNameReversedMap[accountType] = accountTypeName
}

func (r *beancountDataReader) readAccountLine(ctx core.Context, recordId string, items []string, date string, directive beancountDirective, data *beancountData) (*beancountAccount, error) {
	if r.getNotEmptyItemsCount(items) < 3 {
		log.Warnf(ctx, "[beancount_data_reader.parseAccount] cannot parse account %s \"%s\", because items count in line not correct", recordId, strings.Join(items, " "))
		return nil, nil
	}

	var err error
	accountName := r.getNotEmptyItemByIndex(items, 2)
	account, exists := data.Accounts[accountName]

	if !exists {
		account, err = r.createAccount(ctx, recordId, data, accountName)

		if err != nil {
			return nil, err
		}
	}

	if directive == beancountDirectiveOpen {
		account.OpenDate = date
		return account, nil
	} else if directive == beancountDirectiveClose {
		account.CloseDate = date
		return account, nil
	} else {
		log.Warnf(ctx, "[beancount_data_reader.parseAccount] cannot parse account %s \"%s\", because directive is invalid", recordId, strings.Join(items, " "))
		return nil, nil
	}
}

func (r *beancountDataReader) createAccount(ctx core.Context, recordId string, data *beancountData, accountName string) (*beancountAccount, error) {
	account := &beancountAccount{
		Name:        accountName,
		AccountType: beancountUnknownAccountType,
	}

	accountNameItems := strings.Split(accountName, beancountAccountNameItemsSeparator)

	if len(accountNameItems) > 1 {
		accountType, exists := r.accountTypeNameMap[accountNameItems[0]]

		if exists {
			account.AccountType = accountType
		} else {
			log.Warnf(ctx, "[beancount_data_reader.createAccount] cannot parse account \"%s\", because account type \"%s\" is invalid", accountName, accountNameItems[0])
			return nil, errs.NewImportError(errs.ErrInvalidBeancountFile, recordId, accountName)
		}
	}

	data.Accounts[accountName] = account
	return account, nil
}
