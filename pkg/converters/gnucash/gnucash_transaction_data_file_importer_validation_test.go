package gnucash

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

func TestGnuCashTransactionDatabaseFileParseImportedData_NotSupportedToParseSplitTransaction(t *testing.T) {
	importer := GnuCashTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		gnucashCommonValidDataCaseHeader+
			"<gnc:account version=\"2.0.0\">\n"+
			"  <act:name>Test Category2</act:name>\n"+
			"  <act:id type=\"guid\">00000000000000000000000000000200</act:id>\n"+
			"  <act:type>EXPENSE</act:type>\n"+
			"  <act:parent type=\"guid\">00000000000000000000000000000001</act:parent>\n"+
			"</gnc:account>\n"+
			"<gnc:account version=\"2.0.0\">\n"+
			"  <act:name>Test Account2</act:name>\n"+
			"  <act:id type=\"guid\">00000000000000000000000000002000</act:id>\n"+
			"  <act:type>CASH</act:type>\n"+
			"  <act:commodity>\n"+
			"    <cmdty:space>CURRENCY</cmdty:space>\n"+
			"    <cmdty:id>CNY</cmdty:id>\n"+
			"  </act:commodity>\n"+
			"  <act:parent type=\"guid\">00000000000000000000000000000001</act:parent>\n"+
			"</gnc:account>\n"+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:date-posted>\n"+
			"    <ts:date>2024-09-01 12:34:56 +0000</ts:date>\n"+
			"  </trn:date-posted>\n"+
			"  <trn:splits>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>100/100</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000000200</split:account>\n"+
			"    </trn:split>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>200/100</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000002000</split:account>\n"+
			"    </trn:split>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>-300/100</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n"+
			"    </trn:split>\n"+
			"  </trn:splits>\n"+
			"</gnc:transaction>\n"+
			gnucashCommonValidDataCaseFooter), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrNotSupportedSplitTransactions.Message)
}

func TestGnuCashTransactionDatabaseFileParseImportedData_MissingAccountRequiredNode(t *testing.T) {
	importer := GnuCashTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1,
		DefaultCurrency: "CNY",
	}

	// Missing Account Currency Node
	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"<?xml version=\"1.0\" encoding=\"utf-8\" ?>\n"+
			"<gnc-v2\n"+
			"     xmlns:gnc=\"http://www.gnucash.org/XML/gnc\"\n"+
			"     xmlns:act=\"http://www.gnucash.org/XML/act\"\n"+
			"     xmlns:book=\"http://www.gnucash.org/XML/book\"\n"+
			"     xmlns:cd=\"http://www.gnucash.org/XML/cd\"\n"+
			"     xmlns:cmdty=\"http://www.gnucash.org/XML/cmdty\"\n"+
			"     xmlns:slot=\"http://www.gnucash.org/XML/slot\"\n"+
			"     xmlns:split=\"http://www.gnucash.org/XML/split\"\n"+
			"     xmlns:trn=\"http://www.gnucash.org/XML/trn\">\n"+
			"<gnc:book version=\"2.0.0\">\n"+
			"<gnc:account version=\"2.0.0\">\n"+
			"  <act:name>Root Account</act:name>\n"+
			"  <act:id type=\"guid\">00000000000000000000000000000001</act:id>\n"+
			"  <act:type>ROOT</act:type>\n"+
			"</gnc:account>\n"+
			"<gnc:account version=\"2.0.0\">\n"+
			"  <act:name>Opening Balances</act:name>\n"+
			"  <act:id type=\"guid\">00000000000000000000000000000010</act:id>\n"+
			"  <act:type>EQUITY</act:type>\n"+
			"  <act:commodity>\n"+
			"    <cmdty:space>CURRENCY</cmdty:space>\n"+
			"    <cmdty:id>CNY</cmdty:id>\n"+
			"  </act:commodity>\n"+
			"  <act:slots>\n"+
			"    <slot>\n"+
			"      <slot:key>equity-type</slot:key>\n"+
			"      <slot:value type=\"string\">opening-balance</slot:value>\n"+
			"    </slot>\n"+
			"  </act:slots>\n"+
			"</gnc:account>\n"+
			"<gnc:account version=\"2.0.0\">\n"+
			"  <act:name>Test Account</act:name>\n"+
			"  <act:id type=\"guid\">00000000000000000000000000001000</act:id>\n"+
			"  <act:type>BANK</act:type>\n"+
			"  <act:parent type=\"guid\">00000000000000000000000000000001</act:parent>\n"+
			"</gnc:account>\n"+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:date-posted>\n"+
			"    <ts:date>2024-09-01 00:00:00 +0000</ts:date>\n"+
			"  </trn:date-posted>\n"+
			"  <trn:splits>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>12345/100</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n"+
			"    </trn:split>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>-12345/100</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000000010</split:account>\n"+
			"    </trn:split>\n"+
			"  </trn:splits>\n"+
			"</gnc:transaction>\n"+
			gnucashCommonValidDataCaseFooter), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAccountCurrencyInvalid.Message)
}

func TestGnuCashTransactionDatabaseFileParseImportedData_MissingTransactionRequiredNode(t *testing.T) {
	importer := GnuCashTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1,
		DefaultCurrency: "CNY",
	}

	// Missing Transaction Time Node
	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		gnucashCommonValidDataCaseHeader+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:splits>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>12345/100</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n"+
			"    </trn:split>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>-12345/100</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000000010</split:account>\n"+
			"    </trn:split>\n"+
			"  </trn:splits>\n"+
			"</gnc:transaction>\n"+
			gnucashCommonValidDataCaseFooter), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingTransactionTime.Message)

	// Missing Transaction Splits Node
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		gnucashCommonValidDataCaseHeader+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:date-posted>\n"+
			"    <ts:date>2024-09-01 00:00:00 +0000</ts:date>\n"+
			"  </trn:date-posted>\n"+
			"</gnc:transaction>\n"+
			gnucashCommonValidDataCaseFooter), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrInvalidGnuCashFile.Message)

	// Missing Transaction Split Quantity Node
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		gnucashCommonValidDataCaseHeader+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:date-posted>\n"+
			"    <ts:date>2024-09-01 00:00:00 +0000</ts:date>\n"+
			"  </trn:date-posted>\n"+
			"  <trn:splits>\n"+
			"    <trn:split>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n"+
			"    </trn:split>\n"+
			"    <trn:split>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000000010</split:account>\n"+
			"    </trn:split>\n"+
			"  </trn:splits>\n"+
			"</gnc:transaction>\n"+
			gnucashCommonValidDataCaseFooter), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)

	// Missing Transaction Split Account Node
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		gnucashCommonValidDataCaseHeader+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:date-posted>\n"+
			"    <ts:date>2024-09-01 00:00:00 +0000</ts:date>\n"+
			"  </trn:date-posted>\n"+
			"  <trn:splits>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>12345/100</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n"+
			"    </trn:split>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>-12345/100</split:quantity>\n"+
			"    </trn:split>\n"+
			"  </trn:splits>\n"+
			"</gnc:transaction>\n"+
			gnucashCommonValidDataCaseFooter), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingAccountData.Message)
}

func checkParsedMinimumValidData(t *testing.T, allNewTransactions models.ImportedTransactionSlice, allNewAccounts []*models.Account, allNewSubExpenseCategories []*models.TransactionCategory, allNewSubIncomeCategories []*models.TransactionCategory, allNewSubTransferCategories []*models.TransactionCategory, allNewTags []*models.TransactionTag) {
	assert.Equal(t, 4, len(allNewTransactions))
	assert.Equal(t, 2, len(allNewAccounts))
	assert.Equal(t, 1, len(allNewSubExpenseCategories))
	assert.Equal(t, 1, len(allNewSubIncomeCategories))
	assert.Equal(t, 1, len(allNewSubTransferCategories))
	assert.Equal(t, 0, len(allNewTags))

	assert.Equal(t, int64(1234567890), allNewTransactions[0].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_MODIFY_BALANCE, allNewTransactions[0].Type)
	assert.Equal(t, int64(1725148800), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)
	assert.Equal(t, "Test Account", allNewTransactions[0].OriginalSourceAccountName)
	assert.Equal(t, "", allNewTransactions[0].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[1].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_INCOME, allNewTransactions[1].Type)
	assert.Equal(t, int64(1725153825), utils.GetUnixTimeFromTransactionTime(allNewTransactions[1].TransactionTime))
	assert.Equal(t, int64(12), allNewTransactions[1].Amount)
	assert.Equal(t, "Test Account", allNewTransactions[1].OriginalSourceAccountName)
	assert.Equal(t, "Test Category", allNewTransactions[1].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[2].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_EXPENSE, allNewTransactions[2].Type)
	assert.Equal(t, int64(1725194096), utils.GetUnixTimeFromTransactionTime(allNewTransactions[2].TransactionTime))
	assert.Equal(t, int64(100), allNewTransactions[2].Amount)
	assert.Equal(t, "Test Account", allNewTransactions[2].OriginalSourceAccountName)
	assert.Equal(t, "Test Category2", allNewTransactions[2].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[3].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_TRANSFER_OUT, allNewTransactions[3].Type)
	assert.Equal(t, int64(1725235199), utils.GetUnixTimeFromTransactionTime(allNewTransactions[3].TransactionTime))
	assert.Equal(t, int64(5), allNewTransactions[3].Amount)
	assert.Equal(t, "Test Account", allNewTransactions[3].OriginalSourceAccountName)
	assert.Equal(t, "Test Account2", allNewTransactions[3].OriginalDestinationAccountName)
	assert.Equal(t, "", allNewTransactions[3].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewAccounts[0].Uid)
	assert.Equal(t, "Test Account", allNewAccounts[0].Name)
	assert.Equal(t, "CNY", allNewAccounts[0].Currency)

	assert.Equal(t, int64(1234567890), allNewAccounts[1].Uid)
	assert.Equal(t, "Test Account2", allNewAccounts[1].Name)
	assert.Equal(t, "CNY", allNewAccounts[1].Currency)

	assert.Equal(t, int64(1234567890), allNewSubExpenseCategories[0].Uid)
	assert.Equal(t, "Test Category2", allNewSubExpenseCategories[0].Name)

	assert.Equal(t, int64(1234567890), allNewSubIncomeCategories[0].Uid)
	assert.Equal(t, "Test Category", allNewSubIncomeCategories[0].Name)

	assert.Equal(t, int64(1234567890), allNewSubTransferCategories[0].Uid)
	assert.Equal(t, "", allNewSubTransferCategories[0].Name)
}
