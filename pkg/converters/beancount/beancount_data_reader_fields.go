package beancount

import (
	"bytes"
	"encoding/csv"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
	"strings"
)

func (r *beancountDataReader) getNotEmptyItemByIndex(items []string, index int) string {
	item, _ := r.getNotEmptyItemAndIndexByIndex(items, index)
	return item
}

func (r *beancountDataReader) getNotEmptyItemAndIndexByIndex(items []string, index int) (string, int) {
	count := -1

	for i := 0; i < len(items); i++ {
		item := items[i]

		if len(item) == 0 {
			continue
		}

		count++

		if count == index {
			return items[i], i
		}
	}

	return "", -1
}

func (r *beancountDataReader) getNotEmptyItemAndIndexFromIndex(items []string, startIndex int) (string, int) {
	for i := startIndex; i < len(items); i++ {
		item := items[i]

		if len(item) == 0 {
			continue
		}

		return item, i
	}

	return "", -1
}

func (r *beancountDataReader) getNotEmptyItemsCount(items []string) int {
	count := 0

	for i := 0; i < len(items); i++ {
		if len(items[i]) > 0 {
			count++
		}
	}

	return count
}

func (r *beancountDataReader) getOriginalAmountAndLastIndexFromIndex(items []string, startIndex int) (string, int) {
	amountBuilder := strings.Builder{}
	lastIndex := -1

	for i := startIndex; i < len(items); i++ {
		item := items[i]

		if len(item) == 0 {
			continue
		}

		valid := true

		// The Amount in “Postings” can also be an arithmetic expression using ( ) * / - +
		for j := 0; j < len(item); j++ {
			if !(item[j] >= '0' && item[j] <= '9') && item[j] != '.' && item[j] != '(' && item[j] != ')' &&
				item[j] != '*' && item[j] != '/' && item[j] != '-' && item[j] != '+' {
				valid = false
				break
			}
		}

		if !valid {
			break
		}

		if amountBuilder.Len() > 0 {
			amountBuilder.WriteRune(' ')
		}

		amountBuilder.WriteString(item)
		lastIndex = i
	}

	return amountBuilder.String(), lastIndex
}

func createNewBeancountDataReader(ctx core.Context, data []byte) (*beancountDataReader, error) {
	fallback := unicode.UTF8.NewDecoder()
	reader := transform.NewReader(bytes.NewReader(data), unicode.BOMOverride(fallback))
	csvReader := csv.NewReader(reader)
	csvReader.Comma = ' '
	csvReader.FieldsPerRecord = -1

	allData := make([]beancountLineData, 0)

	for {
		items, err := csvReader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Errorf(ctx, "[beancount_data_reader.createNewBeancountDataReader] cannot parse data, because %s", err.Error())
			return nil, errs.ErrInvalidBeancountFile
		}

		lineNumber, _ := csvReader.FieldPos(0)
		allData = append(allData, beancountLineData{
			lineNumber: lineNumber,
			items:      items,
		})
	}

	return &beancountDataReader{
		accountTypeNameMap: map[string]beancountAccountType{
			beancountDefaultAssetsAccountTypeName:      beancountAssetsAccountType,
			beancountDefaultLiabilitiesAccountTypeName: beancountLiabilitiesAccountType,
			beancountDefaultEquityAccountTypeName:      beancountEquityAccountType,
			beancountDefaultIncomeAccountTypeName:      beancountIncomeAccountType,
			beancountDefaultExpenseAccountTypeName:     beancountExpensesAccountType,
		},
		accountTypeNameReversedMap: map[beancountAccountType]string{
			beancountAssetsAccountType:      beancountDefaultAssetsAccountTypeName,
			beancountLiabilitiesAccountType: beancountDefaultLiabilitiesAccountTypeName,
			beancountEquityAccountType:      beancountDefaultEquityAccountTypeName,
			beancountIncomeAccountType:      beancountDefaultIncomeAccountTypeName,
			beancountExpensesAccountType:    beancountDefaultExpenseAccountTypeName,
		},
		allData: allData,
	}, nil
}
