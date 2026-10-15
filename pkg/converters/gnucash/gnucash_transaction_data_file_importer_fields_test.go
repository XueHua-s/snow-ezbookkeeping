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

func TestGnuCashTransactionDatabaseFileParseImportedData_ParseInvalidTime(t *testing.T) {
	importer := GnuCashTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		gnucashCommonValidDataCaseHeader+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:date-posted>\n"+
			"    <ts:date>2024-09-01 00:00:00</ts:date>\n"+
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
	assert.EqualError(t, err, errs.ErrTransactionTimeInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		gnucashCommonValidDataCaseHeader+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:date-posted>\n"+
			"    <ts:date>2024-09-01T00:00:00+00:00</ts:date>\n"+
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
	assert.EqualError(t, err, errs.ErrTransactionTimeInvalid.Message)
}

func TestGnuCashTransactionDatabaseFileParseImportedData_ParseValidTimezone(t *testing.T) {
	importer := GnuCashTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		gnucashCommonValidDataCaseHeader+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:date-posted>\n"+
			"    <ts:date>2024-09-01 12:34:56 -1000</ts:date>\n"+
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
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725230096), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		gnucashCommonValidDataCaseHeader+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:date-posted>\n"+
			"    <ts:date>2024-09-01 12:34:56 +1245</ts:date>\n"+
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
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1725148196), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
}

func TestGnuCashTransactionDatabaseFileParseImportedData_ParseValidAccountCurrency(t *testing.T) {
	importer := GnuCashTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, allNewAccounts, _, _, _, _, err := importer.ParseImportedData(context, user, []byte("<?xml version=\"1.0\" encoding=\"utf-8\" ?>\n"+
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
		"  <act:commodity>\n"+
		"    <cmdty:space>CURRENCY</cmdty:space>\n"+
		"    <cmdty:id>USD</cmdty:id>\n"+
		"  </act:commodity>\n"+
		"  <act:parent type=\"guid\">00000000000000000000000000000001</act:parent>\n"+
		"</gnc:account>\n"+
		"<gnc:account version=\"2.0.0\">\n"+
		"  <act:name>Test Account2</act:name>\n"+
		"  <act:id type=\"guid\">00000000000000000000000000002000</act:id>\n"+
		"  <act:type>CASH</act:type>\n"+
		"  <act:commodity>\n"+
		"    <cmdty:space>CURRENCY</cmdty:space>\n"+
		"    <cmdty:id>EUR</cmdty:id>\n"+
		"  </act:commodity>\n"+
		"  <act:parent type=\"guid\">00000000000000000000000000000001</act:parent>\n"+
		"</gnc:account>\n"+
		"<gnc:transaction version=\"2.0.0\">\n"+
		"  <trn:date-posted>\n"+
		"    <ts:date>2024-09-01 01:23:45 +0000</ts:date>\n"+
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
		"<gnc:transaction version=\"2.0.0\">\n"+
		"  <trn:date-posted>\n"+
		"    <ts:date>2024-09-01 12:34:56 +0000</ts:date>\n"+
		"  </trn:date-posted>\n"+
		"  <trn:splits>\n"+
		"    <trn:split>\n"+
		"      <split:quantity>5/100</split:quantity>\n"+
		"      <split:account type=\"guid\">00000000000000000000000000002000</split:account>\n"+
		"    </trn:split>\n"+
		"    <trn:split>\n"+
		"      <split:quantity>-5/100</split:quantity>\n"+
		"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n"+
		"    </trn:split>\n"+
		"  </trn:splits>\n"+
		"</gnc:transaction>\n"+
		"</gnc:book>\n"+
		"</gnc-v2>\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 2, len(allNewTransactions))
	assert.Equal(t, 2, len(allNewAccounts))

	assert.Equal(t, int64(1234567890), allNewAccounts[0].Uid)
	assert.Equal(t, "Test Account", allNewAccounts[0].Name)
	assert.Equal(t, "USD", allNewAccounts[0].Currency)

	assert.Equal(t, int64(1234567890), allNewAccounts[1].Uid)
	assert.Equal(t, "Test Account2", allNewAccounts[1].Name)
	assert.Equal(t, "EUR", allNewAccounts[1].Currency)
}

func TestGnuCashTransactionDatabaseFileParseImportedData_ParseValidAmount(t *testing.T) {
	importer := GnuCashTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		gnucashCommonValidDataCaseHeader+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:date-posted>\n"+
			"    <ts:date>2024-09-01 12:34:56 +0000</ts:date>\n"+
			"  </trn:date-posted>\n"+
			"  <trn:splits>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>12345/1</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n"+
			"    </trn:split>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>-12345/1</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000000010</split:account>\n"+
			"    </trn:split>\n"+
			"  </trn:splits>\n"+
			"</gnc:transaction>\n"+
			gnucashCommonValidDataCaseFooter), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1234500), allNewTransactions[0].Amount)

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		gnucashCommonValidDataCaseHeader+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:date-posted>\n"+
			"    <ts:date>2024-09-01 12:34:56 +0000</ts:date>\n"+
			"  </trn:date-posted>\n"+
			"  <trn:splits>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>12345/1000</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n"+
			"    </trn:split>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>-12345/1000</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000000010</split:account>\n"+
			"    </trn:split>\n"+
			"  </trn:splits>\n"+
			"</gnc:transaction>\n"+
			gnucashCommonValidDataCaseFooter), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(1234), allNewTransactions[0].Amount)
}

func TestGnuCashTransactionDatabaseFileParseImportedData_ParseInvalidAmount(t *testing.T) {
	importer := GnuCashTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		gnucashCommonValidDataCaseHeader+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:date-posted>\n"+
			"    <ts:date>2024-09-01 12:34:56 +0000</ts:date>\n"+
			"  </trn:date-posted>\n"+
			"  <trn:splits>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>/</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n"+
			"    </trn:split>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>/</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000000010</split:account>\n"+
			"    </trn:split>\n"+
			"  </trn:splits>\n"+
			"</gnc:transaction>\n"+
			gnucashCommonValidDataCaseFooter), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		gnucashCommonValidDataCaseHeader+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:date-posted>\n"+
			"    <ts:date>2024-09-01 12:34:56 +0000</ts:date>\n"+
			"  </trn:date-posted>\n"+
			"  <trn:splits>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>12345</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n"+
			"    </trn:split>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>-12345</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000000010</split:account>\n"+
			"    </trn:split>\n"+
			"  </trn:splits>\n"+
			"</gnc:transaction>\n"+
			gnucashCommonValidDataCaseFooter), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		gnucashCommonValidDataCaseHeader+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:date-posted>\n"+
			"    <ts:date>2024-09-01 12:34:56 +0000</ts:date>\n"+
			"  </trn:date-posted>\n"+
			"  <trn:splits>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>12345/</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n"+
			"    </trn:split>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>-12345/</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000000010</split:account>\n"+
			"    </trn:split>\n"+
			"  </trn:splits>\n"+
			"</gnc:transaction>\n"+
			gnucashCommonValidDataCaseFooter), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)
}

func TestGnuCashTransactionDatabaseFileParseImportedData_ParseDescription(t *testing.T) {
	importer := GnuCashTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		gnucashCommonValidDataCaseHeader+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:date-posted>\n"+
			"    <ts:date>2024-09-01 12:34:56 +0000</ts:date>\n"+
			"  </trn:date-posted>\n"+
			"  <trn:description>foo    bar\t#test</trn:description>\n"+
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

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, "foo    bar\t#test", allNewTransactions[0].Comment)
}

func TestGnuCashTransactionDatabaseFileParseImportedData_SkipZeroAmountTransaction(t *testing.T) {
	importer := GnuCashTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		gnucashCommonValidDataCaseHeader+
			"<gnc:transaction version=\"2.0.0\">\n"+
			"  <trn:date-posted>\n"+
			"    <ts:date>2024-09-01 12:34:56 +0000</ts:date>\n"+
			"  </trn:date-posted>\n"+
			"  <trn:splits>\n"+
			"    <trn:split>\n"+
			"      <split:quantity>0/100</split:quantity>\n"+
			"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n"+
			"    </trn:split>\n"+
			"  </trn:splits>\n"+
			"</gnc:transaction>\n"+
			gnucashCommonValidDataCaseFooter), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrNotFoundTransactionDataInFile.Message)
}
