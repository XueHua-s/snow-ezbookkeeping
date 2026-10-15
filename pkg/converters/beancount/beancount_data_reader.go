package beancount

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"strings"
)

const beancountDefaultAssetsAccountTypeName = "Assets"

const beancountDefaultLiabilitiesAccountTypeName = "Liabilities"

const beancountDefaultEquityAccountTypeName = "Equity"

const beancountDefaultIncomeAccountTypeName = "Income"

const beancountDefaultExpenseAccountTypeName = "Expenses"

const beancountOptionAssetsAccountTypeName = "name_assets"

const beancountOptionLiabilitiesAccountTypeName = "name_liabilities"

const beancountOptionEquityAccountTypeName = "name_equity"

const beancountOptionIncomeAccountTypeName = "name_income"

const beancountOptionExpenseAccountTypeName = "name_expenses"

const beancountCommentPrefix = ';'

const beancountAccountNameItemsSeparator = ":"

const beancountMetadataKeySuffix = ':'

const beancountPricePrefix = '@'

const beancountLinkPrefix = '^'

const beancountTagPrefix = '#'

// beancountDataReader defines the structure of Beancount data reader
type beancountDataReader struct {
	accountTypeNameMap         map[string]beancountAccountType
	accountTypeNameReversedMap map[beancountAccountType]string
	allData                    []beancountLineData
}

type beancountLineData struct {
	items      []string
	lineNumber int
}

// read returns the imported Beancount data
// Reference: https://beancount.github.io/docs/beancount_language_syntax.html
func (r *beancountDataReader) read(ctx core.Context) (*beancountData, error) {
	if len(r.allData) < 1 {
		return nil, errs.ErrNotFoundTransactionDataInFile
	}

	data := &beancountData{
		Accounts:     make(map[string]*beancountAccount),
		Transactions: make([]*beancountTransactionEntry, 0),
	}

	var err error
	var currentTransactionEntry *beancountTransactionEntry
	var currentTransactionPosting *beancountPosting
	var currentTags []string

	for i := 0; i < len(r.allData); i++ {
		lineData := r.allData[i]
		items := lineData.items
		recordId := fmt.Sprintf("line:%d", lineData.lineNumber)

		if len(items) == 0 || (len(items) == 1 && len(items[0]) == 0) || (len(r.getNotEmptyItemByIndex(items, 0)) > 0 && r.getNotEmptyItemByIndex(items, 0)[0] == beancountCommentPrefix) { // skip empty or comment lines
			continue
		}

		if r.getNotEmptyItemsCount(items) < 2 {
			log.Warnf(ctx, "[beancount_data_reader.read] cannot parse %s \"%s\", because not enough items in line", recordId, strings.Join(items, " "))
			continue
		}

		firstItem := items[0]

		if firstItem == "include" { // not support include directive
			return nil, errs.NewImportError(errs.ErrBeancountFileNotSupportInclude, recordId)
		} else if firstItem == "plugin" { // skip plugin directive lines
			currentTransactionEntry, currentTransactionPosting = r.updateCurrentState(data, currentTransactionEntry, currentTransactionPosting)
			continue
		} else if firstItem == "option" {
			currentTransactionEntry, currentTransactionPosting = r.updateCurrentState(data, currentTransactionEntry, currentTransactionPosting)
			r.readAndSetOption(ctx, recordId, items)
			continue
		} else if firstItem == "pushtag" {
			currentTransactionEntry, currentTransactionPosting = r.updateCurrentState(data, currentTransactionEntry, currentTransactionPosting)
			currentTags = r.readAndSetTags(ctx, recordId, items, currentTags, true)
			continue
		} else if firstItem == "poptag" {
			currentTransactionEntry, currentTransactionPosting = r.updateCurrentState(data, currentTransactionEntry, currentTransactionPosting)
			currentTags = r.readAndSetTags(ctx, recordId, items, currentTags, false)
			continue
		}

		if len(firstItem) == 0 { // original line has space prefix, maybe transaction posting or metadata line
			actualFirstItem := r.getNotEmptyItemByIndex(items, 0)

			if len(actualFirstItem) == 0 { // skip empty lines
				continue
			}

			if ('A' <= actualFirstItem[0] && actualFirstItem[0] <= 'Z') || actualFirstItem[0] == '!' { // transaction posting
				if currentTransactionEntry != nil && currentTransactionPosting != nil {
					currentTransactionEntry.Postings = append(currentTransactionEntry.Postings, currentTransactionPosting)
					currentTransactionPosting = nil
				}

				currentTransactionPosting, err = r.readTransactionPostingLine(ctx, recordId, items, data, actualFirstItem[0] == '!')

				if err != nil {
					return nil, err
				}
			} else if 'a' <= actualFirstItem[0] && actualFirstItem[0] <= 'z' { // metadata
				metadata := r.readTransactionMetadataLine(ctx, recordId, items)

				if metadata == nil {
					continue
				}

				metadataKey := metadata[0]
				metadataValue := metadata[1]

				if currentTransactionPosting != nil {
					if _, exists := currentTransactionPosting.Metadata[metadataKey]; !exists {
						currentTransactionPosting.Metadata[metadataKey] = metadataValue
					}
				} else if currentTransactionEntry != nil {
					if _, exists := currentTransactionEntry.Metadata[metadataKey]; !exists {
						currentTransactionEntry.Metadata[metadataKey] = metadataValue
					}
				}
			} else {
				log.Warnf(ctx, "[beancount_data_reader.read] cannot parse %s \"%s\", because line prefix is invalid", recordId, strings.Join(items, " "))
				currentTransactionEntry, currentTransactionPosting = r.updateCurrentState(data, currentTransactionEntry, currentTransactionPosting)
				continue
			}
		} else if _, err := utils.ParseFromLongDateFirstTime(firstItem, 0); err == nil { // original line has date as first item
			currentTransactionEntry, currentTransactionPosting = r.updateCurrentState(data, currentTransactionEntry, currentTransactionPosting)

			directive := r.getNotEmptyItemByIndex(items, 1)

			if directive == string(beancountDirectiveOpen) ||
				directive == string(beancountDirectiveClose) {
				_, err := r.readAccountLine(ctx, recordId, items, firstItem, beancountDirective(directive), data)

				if err != nil {
					return nil, err
				}
			} else if directive == string(beancountDirectiveTransaction) ||
				directive == string(beancountDirectiveCompletedTransaction) ||
				directive == string(beancountDirectiveInCompleteTransaction) ||
				directive == string(beancountDirectivePaddingTransaction) {
				currentTransactionEntry = r.readTransactionLine(ctx, recordId, items, firstItem, beancountDirective(directive), currentTags)
			} else if directive == string(beancountDirectiveCommodity) ||
				directive == string(beancountDirectivePrice) ||
				directive == string(beancountDirectiveNote) ||
				directive == string(beancountDirectiveDocument) ||
				directive == string(beancountDirectiveEvent) ||
				directive == string(beancountDirectiveBalance) ||
				directive == string(beancountDirectivePad) ||
				directive == string(beancountDirectiveQuery) ||
				directive == string(beancountDirectiveCustom) { // skip commodity / price / note / document / event / balance / pad / query / custom lines
				continue
			} else {
				log.Warnf(ctx, "[beancount_data_reader.read] cannot parse %s \"%s\", because directive is unknown", recordId, strings.Join(items, " "))
				continue
			}
		} else { // first item not start with date or space
			currentTransactionEntry, currentTransactionPosting = r.updateCurrentState(data, currentTransactionEntry, currentTransactionPosting)
			continue
		}
	}

	if currentTransactionEntry != nil {
		if currentTransactionPosting != nil {
			currentTransactionEntry.Postings = append(currentTransactionEntry.Postings, currentTransactionPosting)
			currentTransactionPosting = nil
		}

		data.Transactions = append(data.Transactions, currentTransactionEntry)
		currentTransactionEntry = nil
	}

	return data, nil
}

func (r *beancountDataReader) updateCurrentState(data *beancountData, currentTransactionEntry *beancountTransactionEntry, currentTransactionPosting *beancountPosting) (*beancountTransactionEntry, *beancountPosting) {
	if currentTransactionEntry != nil {
		if currentTransactionPosting != nil {
			currentTransactionEntry.Postings = append(currentTransactionEntry.Postings, currentTransactionPosting)
			currentTransactionPosting = nil
		}

		data.Transactions = append(data.Transactions, currentTransactionEntry)
		currentTransactionEntry = nil
		currentTransactionPosting = nil
	}

	return currentTransactionEntry, currentTransactionPosting
}
