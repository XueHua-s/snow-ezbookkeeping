package iif

import (
	"github.com/mayswind/ezbookkeeping/pkg/converters/converter"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestIIFTransactionDataFileParseImportedData_ParseYearMonthDayFormatTime(t *testing.T) {
	importer := IifTransactionDataFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t2024/09/01\tTest Account\t123.45\n"+
			"SPL\t2024/09/01\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"+
			"TRNS\t2024/09/2\tTest Account\t123.45\n"+
			"SPL\t2024/09/2\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"+
			"TRNS\t2024/9/03\tTest Account\t123.45\n"+
			"SPL\t2024/9/03\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"+
			"TRNS\t2024/9/4\tTest Account\t123.45\n"+
			"SPL\t2024/9/4\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 4, len(allNewTransactions))
	assert.Equal(t, int64(1725148800), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
	assert.Equal(t, int64(1725235200), utils.GetUnixTimeFromTransactionTime(allNewTransactions[1].TransactionTime))
	assert.Equal(t, int64(1725321600), utils.GetUnixTimeFromTransactionTime(allNewTransactions[2].TransactionTime))
	assert.Equal(t, int64(1725408000), utils.GetUnixTimeFromTransactionTime(allNewTransactions[3].TransactionTime))
}

func TestIIFTransactionDataFileParseImportedData_ParseShortMonthDayYearFormatTime(t *testing.T) {
	importer := IifTransactionDataFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t9/01/2024\tTest Account\t123.45\n"+
			"SPL\t9/01/2024\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"+
			"TRNS\t09/2/2024\tTest Account\t123.45\n"+
			"SPL\t09/2/2024\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"+
			"TRNS\t9/3/2024\tTest Account\t123.45\n"+
			"SPL\t9/3/2024\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 3, len(allNewTransactions))
	assert.Equal(t, int64(1725148800), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
	assert.Equal(t, int64(1725235200), utils.GetUnixTimeFromTransactionTime(allNewTransactions[1].TransactionTime))
	assert.Equal(t, int64(1725321600), utils.GetUnixTimeFromTransactionTime(allNewTransactions[2].TransactionTime))
}

func TestIIFTransactionDataFileParseImportedData_ParseShortMonthDayTwoDigitsYearFormatTime(t *testing.T) {
	importer := IifTransactionDataFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t9/01/24\tTest Account\t123.45\n"+
			"SPL\t9/01/24\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"+
			"TRNS\t09/2/24\tTest Account\t123.45\n"+
			"SPL\t09/2/24\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"+
			"TRNS\t24/9/3\tTest Account\t123.45\n"+
			"SPL\t24/9/3\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 3, len(allNewTransactions))
	assert.Equal(t, int64(1725148800), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
	assert.Equal(t, int64(1725235200), utils.GetUnixTimeFromTransactionTime(allNewTransactions[1].TransactionTime))
	assert.Equal(t, int64(1725321600), utils.GetUnixTimeFromTransactionTime(allNewTransactions[2].TransactionTime))
}

func TestIIFTransactionDataFileParseImportedData_ParseInvalidTime(t *testing.T) {
	importer := IifTransactionDataFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t09-01-2024\tTest Account\t123.45\n"+
			"SPL\t09-01-2024\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTimeInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t2024-09-01\tTest Account\t123.45\n"+
			"SPL\t2024-09-01\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTimeInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t9/24\tTest Account\t123.45\n"+
			"SPL\t9/24\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTimeInvalid.Message)
}

func TestIIFTransactionDataFileParseImportedData_ParseAmountWithThousandsSeparator(t *testing.T) {
	importer := IifTransactionDataFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t9/01/2024\tTest Account\t123,456.78\n"+
			"SPL\t9/01/2024\tTest Account2\t-123,456.78\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(12345678), allNewTransactions[0].Amount)
}

func TestIIFTransactionDataFileParseImportedData_ParseInvalidAmount(t *testing.T) {
	importer := IifTransactionDataFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\t123 45\n"+
			"SPL\t09/01/2024\tTest Account2\t-123.45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\t123.45\n"+
			"SPL\t09/01/2024\tTest Account2\t-123 45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)
}

func TestIIFTransactionDataFileParseImportedData_ParseDescription(t *testing.T) {
	importer := IifTransactionDataFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tNAME\tAMOUNT\tMEMO\n"+
			"!SPL\tDATE\tACCNT\tNAME\tAMOUNT\tMEMO\n"+
			"!ENDTRNS\t\t\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\t\"Test\"\t123.45\t\"foo    bar\t#test\"\n"+
			"SPL\t09/01/2024\tTest Account2\t\t-123.45\t\n"+
			"ENDTRNS\t\t\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, "foo    bar\t#test", allNewTransactions[0].Comment)

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tNAME\tAMOUNT\tMEMO\n"+
			"!SPL\tDATE\tACCNT\tNAME\tAMOUNT\tMEMO\n"+
			"!ENDTRNS\t\t\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\tTest\t123.45\t\n"+
			"SPL\t09/01/2024\tTest Account2\t\t-123.45\t\n"+
			"ENDTRNS\t\t\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, "Test", allNewTransactions[0].Comment)
}

func TestIIFTransactionDataFileParseImportedData_ParseSplitTransaction(t *testing.T) {
	importer := IifTransactionDataFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, allNewAccounts, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!ACCNT\tNAME\tACCNTTYPE\n"+
			"ACCNT\tTest Category\tINC\n"+
			"ACCNT\tTest Category2\tEXP\n"+
			"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\t123.45\n"+
			"SPL\t09/01/2024\tTest Category\t-23.45\n"+
			"SPL\t09/01/2024\tTest Account2\t-100.00\n"+
			"ENDTRNS\t\t\t\n"+
			"TRNS\t09/02/2024\tTest Account\t-100.00\n"+
			"SPL\t09/02/2024\tTest Category2\t30.00\n"+
			"SPL\t09/02/2024\tTest Account3\t20.00\n"+
			"SPL\t09/02/2024\tTest Account4\t50.00\n"+
			"ENDTRNS\t\t\t\n"+
			"TRNS\t09/03/2024\tTest Account\t100.00\n"+
			"SPL\t09/03/2024\tTest Account2\t-100.00\n"+
			"ENDTRNS\t\t\t\n"+
			"TRNS\t09/04/2024\tTest Category\t-100.00\n"+
			"SPL\t09/04/2024\tTest Account\t40.00\n"+
			"SPL\t09/04/2024\tTest Account2\t60.00\n"+
			"ENDTRNS\t\t\t\n"+
			"TRNS\t09/05/2024\tTest Category2\t100.00\n"+
			"SPL\t09/05/2024\tTest Account3\t-40.00\n"+
			"SPL\t09/05/2024\tTest Account4\t-60.00\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 10, len(allNewTransactions))
	assert.Equal(t, 4, len(allNewAccounts))

	assert.Equal(t, int64(1234567890), allNewTransactions[0].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_INCOME, allNewTransactions[0].Type)
	assert.Equal(t, int64(1725148800), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
	assert.Equal(t, int64(2345), allNewTransactions[0].Amount)
	assert.Equal(t, "Test Account", allNewTransactions[0].OriginalSourceAccountName)
	assert.Equal(t, "Test Category", allNewTransactions[0].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[1].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_TRANSFER_OUT, allNewTransactions[1].Type)
	assert.Equal(t, int64(1725148800), utils.GetUnixTimeFromTransactionTime(allNewTransactions[1].TransactionTime))
	assert.Equal(t, int64(10000), allNewTransactions[1].Amount)
	assert.Equal(t, "Test Account2", allNewTransactions[1].OriginalSourceAccountName)
	assert.Equal(t, "Test Account", allNewTransactions[1].OriginalDestinationAccountName)
	assert.Equal(t, "", allNewTransactions[1].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[2].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_EXPENSE, allNewTransactions[2].Type)
	assert.Equal(t, int64(1725235200), utils.GetUnixTimeFromTransactionTime(allNewTransactions[2].TransactionTime))
	assert.Equal(t, int64(3000), allNewTransactions[2].Amount)
	assert.Equal(t, "Test Account", allNewTransactions[2].OriginalSourceAccountName)
	assert.Equal(t, "Test Category2", allNewTransactions[2].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[3].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_TRANSFER_OUT, allNewTransactions[3].Type)
	assert.Equal(t, int64(1725235200), utils.GetUnixTimeFromTransactionTime(allNewTransactions[3].TransactionTime))
	assert.Equal(t, int64(2000), allNewTransactions[3].Amount)
	assert.Equal(t, "Test Account", allNewTransactions[3].OriginalSourceAccountName)
	assert.Equal(t, "Test Account3", allNewTransactions[3].OriginalDestinationAccountName)
	assert.Equal(t, "", allNewTransactions[3].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[4].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_TRANSFER_OUT, allNewTransactions[4].Type)
	assert.Equal(t, int64(1725235200), utils.GetUnixTimeFromTransactionTime(allNewTransactions[4].TransactionTime))
	assert.Equal(t, int64(5000), allNewTransactions[4].Amount)
	assert.Equal(t, "Test Account", allNewTransactions[4].OriginalSourceAccountName)
	assert.Equal(t, "Test Account4", allNewTransactions[4].OriginalDestinationAccountName)
	assert.Equal(t, "", allNewTransactions[4].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[5].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_TRANSFER_OUT, allNewTransactions[5].Type)
	assert.Equal(t, int64(1725321600), utils.GetUnixTimeFromTransactionTime(allNewTransactions[5].TransactionTime))
	assert.Equal(t, int64(10000), allNewTransactions[5].Amount)
	assert.Equal(t, "Test Account2", allNewTransactions[5].OriginalSourceAccountName)
	assert.Equal(t, "Test Account", allNewTransactions[5].OriginalDestinationAccountName)
	assert.Equal(t, "", allNewTransactions[5].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[6].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_INCOME, allNewTransactions[6].Type)
	assert.Equal(t, int64(1725408000), utils.GetUnixTimeFromTransactionTime(allNewTransactions[6].TransactionTime))
	assert.Equal(t, int64(4000), allNewTransactions[6].Amount)
	assert.Equal(t, "Test Account", allNewTransactions[6].OriginalSourceAccountName)
	assert.Equal(t, "Test Category", allNewTransactions[6].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[7].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_INCOME, allNewTransactions[7].Type)
	assert.Equal(t, int64(1725408000), utils.GetUnixTimeFromTransactionTime(allNewTransactions[7].TransactionTime))
	assert.Equal(t, int64(6000), allNewTransactions[7].Amount)
	assert.Equal(t, "Test Account2", allNewTransactions[7].OriginalSourceAccountName)
	assert.Equal(t, "Test Category", allNewTransactions[7].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[8].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_EXPENSE, allNewTransactions[8].Type)
	assert.Equal(t, int64(1725494400), utils.GetUnixTimeFromTransactionTime(allNewTransactions[8].TransactionTime))
	assert.Equal(t, int64(4000), allNewTransactions[8].Amount)
	assert.Equal(t, "Test Account3", allNewTransactions[8].OriginalSourceAccountName)
	assert.Equal(t, "Test Category2", allNewTransactions[8].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[9].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_EXPENSE, allNewTransactions[9].Type)
	assert.Equal(t, int64(1725494400), utils.GetUnixTimeFromTransactionTime(allNewTransactions[9].TransactionTime))
	assert.Equal(t, int64(6000), allNewTransactions[9].Amount)
	assert.Equal(t, "Test Account4", allNewTransactions[9].OriginalSourceAccountName)
	assert.Equal(t, "Test Category2", allNewTransactions[9].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewAccounts[0].Uid)
	assert.Equal(t, "Test Account", allNewAccounts[0].Name)
	assert.Equal(t, "CNY", allNewAccounts[0].Currency)

	assert.Equal(t, int64(1234567890), allNewAccounts[1].Uid)
	assert.Equal(t, "Test Account2", allNewAccounts[1].Name)
	assert.Equal(t, "CNY", allNewAccounts[1].Currency)

	assert.Equal(t, int64(1234567890), allNewAccounts[2].Uid)
	assert.Equal(t, "Test Account3", allNewAccounts[2].Name)
	assert.Equal(t, "CNY", allNewAccounts[2].Currency)

	assert.Equal(t, int64(1234567890), allNewAccounts[3].Uid)
	assert.Equal(t, "Test Account4", allNewAccounts[3].Name)
	assert.Equal(t, "CNY", allNewAccounts[3].Currency)
}

func TestIIFTransactionDataFileParseImportedData_ParseSplitTransactionDescription(t *testing.T) {
	importer := IifTransactionDataFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tNAME\tAMOUNT\tMEMO\n"+
			"!SPL\tDATE\tACCNT\tNAME\tAMOUNT\tMEMO\n"+
			"!ENDTRNS\t\t\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\t\"Test\"\t123.45\t\"foo    bar\t#test\"\n"+
			"SPL\t09/01/2024\tTest Account2\t\t-100.00\t\"foo\ttest#bar\"\n"+
			"SPL\t09/01/2024\tTest Account3\t\t-23.45\t\n"+
			"ENDTRNS\t\t\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(allNewTransactions))
	assert.Equal(t, "foo\ttest#bar", allNewTransactions[0].Comment)
	assert.Equal(t, "foo    bar\t#test", allNewTransactions[1].Comment)

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tNAME\tAMOUNT\tMEMO\n"+
			"!SPL\tDATE\tACCNT\tNAME\tAMOUNT\tMEMO\n"+
			"!ENDTRNS\t\t\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\tTest\t123.45\t\n"+
			"SPL\t09/01/2024\tTest Account2\t\t-100.00\t\"test\"\n"+
			"SPL\t09/01/2024\tTest Account3\tfoo\t-12.34\t\n"+
			"SPL\t09/01/2024\tTest Account4\t\t-11.11\t\n"+
			"ENDTRNS\t\t\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 3, len(allNewTransactions))
	assert.Equal(t, "test", allNewTransactions[0].Comment)
	assert.Equal(t, "foo", allNewTransactions[1].Comment)
	assert.Equal(t, "Test", allNewTransactions[2].Comment)
}

func TestIIFTransactionDataFileParseImportedData_NotSupportedSplitTransaction(t *testing.T) {
	importer := IifTransactionDataFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	// Opening balance transaction
	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"!TRNS\tTRNSTYPE\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tTRNSTYPE\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\t\n"+
			"TRNS\tBEGINBALCHECK\t09/01/2024\tTest Account\t123.45\n"+
			"SPL\tBEGINBALCHECK\t09/01/2024\tTest Account2\t-100.00\n"+
			"SPL\tBEGINBALCHECK\t09/01/2024\tTest Account3\t-23.45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrNotSupportedSplitTransactions.Message)

	// Transaction with invalid amount
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\t123 45\n"+
			"SPL\t09/01/2024\tTest Account2\t-100.00\n"+
			"SPL\t09/01/2024\tTest Account3\t-23.45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)

	// Transaction split data with invalid amount
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\t123.45\n"+
			"SPL\t09/01/2024\tTest Account2\t-100 00\n"+
			"SPL\t09/01/2024\tTest Account3\t-23.45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)

	// Transaction amount not equal to sum of split data amount
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"!TRNS\tDATE\tACCNT\tAMOUNT\n"+
			"!SPL\tDATE\tACCNT\tAMOUNT\n"+
			"!ENDTRNS\t\t\t\n"+
			"TRNS\t09/01/2024\tTest Account\t123.00\n"+
			"SPL\t09/01/2024\tTest Account2\t-100.00\n"+
			"SPL\t09/01/2024\tTest Account3\t-23.45\n"+
			"ENDTRNS\t\t\t\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrNotSupportedSplitTransactions.Message)
}
