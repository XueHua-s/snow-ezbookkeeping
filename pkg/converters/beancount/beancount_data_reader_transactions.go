package beancount

import (
	"errors"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"strings"
)

func (r *beancountDataReader) readTransactionLine(ctx core.Context, recordId string, items []string, date string, directive beancountDirective, tags []string) *beancountTransactionEntry {
	transactionEntry := &beancountTransactionEntry{
		Date:      date,
		Directive: directive,
		Tags:      make([]string, 0),
		Links:     make([]string, 0),
		Metadata:  make(map[string]string),
	}

	transactionEntry.Tags = append(transactionEntry.Tags, tags...)

	allTags := make(map[string]bool, len(transactionEntry.Tags))

	for _, tag := range transactionEntry.Tags {
		allTags[tag] = true
	}

	// YYYY-MM-DD [txn|Flag] [[Payee] Narration] [#tag] [ˆlink]
	payeeNarrationFirstIndex := 2
	payeeNarrationLastIndex := len(items) - 1

	// parse remain items
	for i := payeeNarrationFirstIndex; i < len(items); i++ {
		item := items[i]

		if len(item) == 0 {
			continue
		}

		if item[0] == beancountCommentPrefix { // ; comment
			if i-1 < payeeNarrationLastIndex {
				payeeNarrationLastIndex = i - 1
			}

			break
		}

		if item[0] == beancountTagPrefix { // [#tag]
			tagName := item[1:]

			if _, exists := allTags[tagName]; !exists {
				transactionEntry.Tags = append(transactionEntry.Tags, tagName)
				allTags[tagName] = true
			}

			if i-1 < payeeNarrationLastIndex {
				payeeNarrationLastIndex = i - 1
			}
		} else if item[0] == beancountLinkPrefix { // [ˆlink]
			transactionEntry.Links = append(transactionEntry.Links, item[1:])

			if i-1 < payeeNarrationLastIndex {
				payeeNarrationLastIndex = i - 1
			}
		}
	}

	if payeeNarrationLastIndex-payeeNarrationFirstIndex >= 1 {
		transactionEntry.Payee = items[payeeNarrationFirstIndex]
		transactionEntry.Narration = items[payeeNarrationFirstIndex+1]
	} else if payeeNarrationLastIndex-payeeNarrationFirstIndex >= 0 {
		transactionEntry.Narration = items[payeeNarrationFirstIndex]
	}

	return transactionEntry
}

func (r *beancountDataReader) readTransactionPostingLine(ctx core.Context, recordId string, items []string, data *beancountData, hasFlag bool) (*beancountPosting, error) {
	// [Flag] Account Amount [{Cost}] [@ Price]
	accountNameExpectedIndex := 0

	if hasFlag {
		accountNameExpectedIndex = 1
	}

	if r.getNotEmptyItemsCount(items) <= accountNameExpectedIndex {
		log.Warnf(ctx, "[beancount_data_reader.readTransactionPostingLine] cannot parse transaction posting %s \"%s\", because items count in line not correct", recordId, strings.Join(items, " "))
		return nil, nil
	}

	accountName, accountNameActualIndex := r.getNotEmptyItemAndIndexByIndex(items, accountNameExpectedIndex)

	if accountName == "" || accountNameActualIndex < 0 {
		log.Warnf(ctx, "[beancount_data_reader.readTransactionPostingLine] cannot parse transaction posting %s \"%s\", because missing account name", recordId, strings.Join(items, " "))
		return nil, errs.NewImportError(errs.ErrMissingAccountData, recordId, accountName)
	}

	transactionPositing := &beancountPosting{
		Account:  accountName,
		Metadata: make(map[string]string),
	}

	amountActualLastIndex := -1
	transactionPositing.OriginalAmount, amountActualLastIndex = r.getOriginalAmountAndLastIndexFromIndex(items, accountNameActualIndex+1)

	if transactionPositing.OriginalAmount == "" || amountActualLastIndex < 0 {
		log.Warnf(ctx, "[beancount_data_reader.readTransactionPostingLine] cannot parse transaction posting %s \"%s\", because missing amount", recordId, strings.Join(items, " "))
		return nil, errs.NewImportError(errs.ErrAmountInvalid, recordId, transactionPositing.OriginalAmount)
	}

	finalAmount, err := evaluateBeancountAmountExpression(ctx, transactionPositing.OriginalAmount)

	if errors.Is(err, errs.ErrNumericOverflow) {
		log.Warnf(ctx, "[beancount_data_reader.readTransactionPostingLine] cannot evaluate amount expression in %s \"%s\", because %s", recordId, strings.Join(items, " "), err.Error())
		return nil, errs.NewImportError(errs.ErrNumericOverflow, recordId, transactionPositing.OriginalAmount)
	} else if err != nil {
		log.Warnf(ctx, "[beancount_data_reader.readTransactionPostingLine] cannot evaluate amount expression in %s \"%s\", because %s", recordId, strings.Join(items, " "), err.Error())
		return nil, errs.NewImportError(errs.ErrAmountInvalid, recordId, transactionPositing.OriginalAmount)
	} else {
		transactionPositing.Amount = finalAmount
	}

	commodityActualIndex := -1
	transactionPositing.Commodity, commodityActualIndex = r.getNotEmptyItemAndIndexFromIndex(items, amountActualLastIndex+1)

	if transactionPositing.Commodity == "" || commodityActualIndex < 0 {
		log.Warnf(ctx, "[beancount_data_reader.readTransactionPostingLine] cannot parse transaction posting %s \"%s\", because missing commodity", recordId, strings.Join(items, " "))
		return nil, errs.NewImportError(errs.ErrInvalidBeancountFile, recordId, transactionPositing.Commodity)
	}

	if strings.ToUpper(transactionPositing.Commodity) != transactionPositing.Commodity { // The syntax for a currency is a word all in capital letters
		log.Warnf(ctx, "[beancount_data_reader.readTransactionPostingLine] cannot parse transaction posting %s \"%s\", because commodity name is not capital letters", recordId, strings.Join(items, " "))
		return nil, errs.NewImportError(errs.ErrInvalidBeancountFile, recordId, transactionPositing.Commodity)
	}

	// parse remain items
	if commodityActualIndex > 0 {
		for i := commodityActualIndex + 1; i < len(items); i++ {
			item := items[i]

			if len(item) == 0 {
				continue
			}

			if item[0] == beancountCommentPrefix { // ; comment
				break
			}

			if len(item) == 2 && item[0] == beancountPricePrefix && item[1] == beancountPricePrefix { // [@@ TotalCost]
				totalCost, totalCostActualIndex := r.getNotEmptyItemAndIndexFromIndex(items, i+1)

				if totalCostActualIndex > 0 {
					transactionPositing.TotalCost = totalCost
					i = totalCostActualIndex

					totalCostCommodity, totalCostCommodityActualIndex := r.getNotEmptyItemAndIndexFromIndex(items, totalCostActualIndex+1)

					if totalCostCommodityActualIndex > 0 {
						transactionPositing.TotalCostCommodity = totalCostCommodity
						i = totalCostCommodityActualIndex
					}
				}
			} else if len(item) == 1 && item[0] == beancountPricePrefix { // [@ Price]
				price, priceActualIndex := r.getNotEmptyItemAndIndexFromIndex(items, i+1)

				if priceActualIndex > 0 {
					transactionPositing.Price = price
					i = priceActualIndex

					priceCommodity, priceCommodityActualIndex := r.getNotEmptyItemAndIndexFromIndex(items, priceActualIndex+1)

					if priceCommodityActualIndex > 0 {
						transactionPositing.PriceCommodity = priceCommodity
						i = priceCommodityActualIndex
					}
				}
			}
		}
	}

	if transactionPositing.Account != "" {
		_, exists := data.Accounts[transactionPositing.Account]

		if !exists {
			_, err := r.createAccount(ctx, recordId, data, transactionPositing.Account)

			if err != nil {
				return nil, err
			}
		}
	}

	return transactionPositing, nil
}

func (r *beancountDataReader) readTransactionMetadataLine(ctx core.Context, recordId string, items []string) []string {
	key := r.getNotEmptyItemByIndex(items, 0)
	value := r.getNotEmptyItemByIndex(items, 1)

	if key == "" || value == "" {
		log.Warnf(ctx, "[beancount_data_reader.readTransactionMetadataLine] cannot parse metadata %s \"%s\", because key or value is empty", recordId, strings.Join(items, " "))
		return nil
	}

	if len(key) == 0 || key[len(key)-1] != beancountMetadataKeySuffix {
		log.Warnf(ctx, "[beancount_data_reader.readTransactionMetadataLine] cannot parse metadata %s \"%s\", because key is invalid correct", recordId, strings.Join(items, " "))
		return nil
	}

	key = key[:len(key)-1]

	return []string{key, value}
}
