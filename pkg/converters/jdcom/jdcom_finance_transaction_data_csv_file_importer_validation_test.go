package jdcom

import (
	"github.com/mayswind/ezbookkeeping/pkg/converters/converter"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestJDComFinanceCsvFileImporterParseImportedData_UnknownStatusTransaction(t *testing.T) {
	importer := JDComFinanceTransactionDataCsvFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	data := "导出信息：\n" +
		"京东账号名：xxxxxx\n" +
		"日期区间：2025-01-01 至 2025-09-01\n" +
		"\n" +
		"交易时间,商户名称,交易说明,金额,收/付款方式,交易状态,收/支\n" +
		"2025-09-01 01:23:45,xxx,xxx,0.12,银行卡,xxxx,支出\n"
	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(data), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTypeInvalid.Message)
}

func TestJDComFinanceCsvFileImporterParseImportedData_UnknownMemoTransferTransaction(t *testing.T) {
	importer := JDComFinanceTransactionDataCsvFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	data := "导出信息：\n" +
		"京东账号名：xxxxxx\n" +
		"日期区间：2025-01-01 至 2025-09-01\n" +
		"\n" +
		"交易时间,商户名称,交易说明,金额,收/付款方式,交易状态,收/支\n" +
		"2025-09-01 01:23:45,xxx,xxx,0.12,银行卡,交易成功,不计收支\n"
	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(data), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTypeInvalid.Message)
}

func TestJDComFinanceCsvFileImporterParseImportedData_MissingFileHeader(t *testing.T) {
	importer := JDComFinanceTransactionDataCsvFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1,
		DefaultCurrency: "CNY",
	}

	data := "交易时间,商户名称,交易说明,金额,收/付款方式,交易状态,收/支\n" +
		"2025-09-01 01:23:45,xxx,xxx,0.12,银行卡,交易成功,支出\n"
	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(data), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidFileHeader.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(""), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidFileHeader.Message)
}

func TestJDComFinanceCsvFileImporterParseImportedData_MissingRequiredColumn(t *testing.T) {
	importer := JDComFinanceTransactionDataCsvFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1,
		DefaultCurrency: "CNY",
	}

	// Missing Time Column
	data1 := "导出信息：\n" +
		"京东账号名：xxxxxx\n" +
		"日期区间：2025-01-01 至 2025-09-01\n" +
		"\n" +
		"商户名称,交易说明,金额,收/付款方式,交易状态,收/支\n" +
		"xxx,xxx,0.12,银行卡,交易成功,支出\n"
	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(data1), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidFileHeader.Message)

	// Missing Merchant Name Column
	data2 := "导出信息：\n" +
		"京东账号名：xxxxxx\n" +
		"日期区间：2025-01-01 至 2025-09-01\n" +
		"\n" +
		"交易时间,交易说明,金额,收/付款方式,交易状态,收/支\n" +
		"2025-09-01 01:23:45,xxx,0.12,银行卡,交易成功,支出\n"
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(data2), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)

	// Missing Transaction Memo Column
	data3 := "导出信息：\n" +
		"京东账号名：xxxxxx\n" +
		"日期区间：2025-01-01 至 2025-09-01\n" +
		"\n" +
		"交易时间,商户名称,金额,收/付款方式,交易状态,收/支\n" +
		"2025-09-01 01:23:45,xxx,0.12,银行卡,交易成功,支出\n"
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(data3), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)

	// Missing Amount Column
	data4 := "导出信息：\n" +
		"京东账号名：xxxxxx\n" +
		"日期区间：2025-01-01 至 2025-09-01\n" +
		"\n" +
		"交易时间,商户名称,交易说明,收/付款方式,交易状态,收/支\n" +
		"2025-09-01 01:23:45,xxx,xxx,银行卡,交易成功,支出\n"
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(data4), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)

	// Missing Related Account Column
	data5 := "导出信息：\n" +
		"京东账号名：xxxxxx\n" +
		"日期区间：2025-01-01 至 2025-09-01\n" +
		"\n" +
		"交易时间,商户名称,交易说明,金额,交易状态,收/支\n" +
		"2025-09-01 01:23:45,xxx,xxx,0.12,交易成功,支出\n"
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(data5), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)

	// Missing Status Column
	data6 := "导出信息：\n" +
		"京东账号名：xxxxxx\n" +
		"日期区间：2025-01-01 至 2025-09-01\n" +
		"\n" +
		"交易时间,商户名称,交易说明,金额,收/付款方式,收/支\n" +
		"2025-09-01 01:23:45,xxx,xxx,0.12,银行卡,支出\n"
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(data6), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)

	// Missing Type Column
	data7 := "导出信息：\n" +
		"京东账号名：xxxxxx\n" +
		"日期区间：2025-01-01 至 2025-09-01\n" +
		"\n" +
		"交易时间,商户名称,交易说明,金额,收/付款方式,交易状态\n" +
		"2025-09-01 01:23:45,xxx,xxx,0.12,银行卡,交易成功\n"
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(data7), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingRequiredFieldInHeaderRow.Message)
}

func TestJDComFinanceCsvFileImporterParseImportedData_NoTransactionData(t *testing.T) {
	importer := JDComFinanceTransactionDataCsvFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	data := "导出信息：\n" +
		"京东账号名：xxxxxx\n" +
		"日期区间：2025-01-01 至 2025-09-01\n" +
		"\n" +
		"交易时间,商户名称,交易说明,金额,收/付款方式,交易状态,收/支\n"
	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(data), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrNotFoundTransactionDataInFile.Message)
}
