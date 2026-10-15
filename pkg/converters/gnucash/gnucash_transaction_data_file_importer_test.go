package gnucash

import (
	"bytes"
	"compress/gzip"
	"github.com/mayswind/ezbookkeeping/pkg/converters/converter"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const gnucashMinimumValidDataCase = "<?xml version=\"1.0\" encoding=\"utf-8\" ?>\n" +
	"<gnc-v2\n" +
	"     xmlns:gnc=\"http://www.gnucash.org/XML/gnc\"\n" +
	"     xmlns:act=\"http://www.gnucash.org/XML/act\"\n" +
	"     xmlns:book=\"http://www.gnucash.org/XML/book\"\n" +
	"     xmlns:cd=\"http://www.gnucash.org/XML/cd\"\n" +
	"     xmlns:cmdty=\"http://www.gnucash.org/XML/cmdty\"\n" +
	"     xmlns:slot=\"http://www.gnucash.org/XML/slot\"\n" +
	"     xmlns:split=\"http://www.gnucash.org/XML/split\"\n" +
	"     xmlns:trn=\"http://www.gnucash.org/XML/trn\">\n" +
	"<gnc:book version=\"2.0.0\">\n" +
	"<gnc:account version=\"2.0.0\">\n" +
	"  <act:name>Root Account</act:name>\n" +
	"  <act:id type=\"guid\">00000000000000000000000000000001</act:id>\n" +
	"  <act:type>ROOT</act:type>\n" +
	"</gnc:account>\n" +
	"<gnc:account version=\"2.0.0\">\n" +
	"  <act:name>Opening Balances</act:name>\n" +
	"  <act:id type=\"guid\">00000000000000000000000000000010</act:id>\n" +
	"  <act:type>EQUITY</act:type>\n" +
	"  <act:commodity>\n" +
	"    <cmdty:space>CURRENCY</cmdty:space>\n" +
	"    <cmdty:id>CNY</cmdty:id>\n" +
	"  </act:commodity>\n" +
	"  <act:slots>\n" +
	"    <slot>\n" +
	"      <slot:key>equity-type</slot:key>\n" +
	"      <slot:value type=\"string\">opening-balance</slot:value>\n" +
	"    </slot>\n" +
	"  </act:slots>\n" +
	"</gnc:account>\n" +
	"<gnc:account version=\"2.0.0\">\n" +
	"  <act:name>Test Category</act:name>\n" +
	"  <act:id type=\"guid\">00000000000000000000000000000100</act:id>\n" +
	"  <act:type>INCOME</act:type>\n" +
	"  <act:parent type=\"guid\">00000000000000000000000000000001</act:parent>\n" +
	"</gnc:account>\n" +
	"<gnc:account version=\"2.0.0\">\n" +
	"  <act:name>Test Category2</act:name>\n" +
	"  <act:id type=\"guid\">00000000000000000000000000000200</act:id>\n" +
	"  <act:type>EXPENSE</act:type>\n" +
	"  <act:parent type=\"guid\">00000000000000000000000000000001</act:parent>\n" +
	"</gnc:account>\n" +
	"<gnc:account version=\"2.0.0\">\n" +
	"  <act:name>Test Account</act:name>\n" +
	"  <act:id type=\"guid\">00000000000000000000000000001000</act:id>\n" +
	"  <act:type>BANK</act:type>\n" +
	"  <act:commodity>\n" +
	"    <cmdty:space>CURRENCY</cmdty:space>\n" +
	"    <cmdty:id>CNY</cmdty:id>\n" +
	"  </act:commodity>\n" +
	"  <act:parent type=\"guid\">00000000000000000000000000000001</act:parent>\n" +
	"</gnc:account>\n" +
	"<gnc:account version=\"2.0.0\">\n" +
	"  <act:name>Test Account2</act:name>\n" +
	"  <act:id type=\"guid\">00000000000000000000000000002000</act:id>\n" +
	"  <act:type>CASH</act:type>\n" +
	"  <act:commodity>\n" +
	"    <cmdty:space>CURRENCY</cmdty:space>\n" +
	"    <cmdty:id>CNY</cmdty:id>\n" +
	"  </act:commodity>\n" +
	"  <act:parent type=\"guid\">00000000000000000000000000000001</act:parent>\n" +
	"</gnc:account>\n" +
	"<gnc:transaction version=\"2.0.0\">\n" +
	"  <trn:date-posted>\n" +
	"    <ts:date>2024-09-01 00:00:00 +0000</ts:date>\n" +
	"  </trn:date-posted>\n" +
	"  <trn:splits>\n" +
	"    <trn:split>\n" +
	"      <split:quantity>12345/100</split:quantity>\n" +
	"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n" +
	"    </trn:split>\n" +
	"    <trn:split>\n" +
	"      <split:quantity>-12345/100</split:quantity>\n" +
	"      <split:account type=\"guid\">00000000000000000000000000000010</split:account>\n" +
	"    </trn:split>\n" +
	"  </trn:splits>\n" +
	"</gnc:transaction>\n" +
	"<gnc:transaction version=\"2.0.0\">\n" +
	"  <trn:date-posted>\n" +
	"    <ts:date>2024-09-01 01:23:45 +0000</ts:date>\n" +
	"  </trn:date-posted>\n" +
	"  <trn:splits>\n" +
	"    <trn:split>\n" +
	"      <split:quantity>12/100</split:quantity>\n" +
	"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n" +
	"    </trn:split>\n" +
	"    <trn:split>\n" +
	"      <split:quantity>-12/100</split:quantity>\n" +
	"      <split:account type=\"guid\">00000000000000000000000000000100</split:account>\n" +
	"    </trn:split>\n" +
	"  </trn:splits>\n" +
	"</gnc:transaction>\n" +
	"<gnc:transaction version=\"2.0.0\">\n" +
	"  <trn:date-posted>\n" +
	"    <ts:date>2024-09-01 12:34:56 +0000</ts:date>\n" +
	"  </trn:date-posted>\n" +
	"  <trn:splits>\n" +
	"    <trn:split>\n" +
	"      <split:quantity>100/100</split:quantity>\n" +
	"      <split:account type=\"guid\">00000000000000000000000000000200</split:account>\n" +
	"    </trn:split>\n" +
	"    <trn:split>\n" +
	"      <split:quantity>-100/100</split:quantity>\n" +
	"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n" +
	"    </trn:split>\n" +
	"  </trn:splits>\n" +
	"</gnc:transaction>\n" +
	"<gnc:transaction version=\"2.0.0\">\n" +
	"  <trn:date-posted>\n" +
	"    <ts:date>2024-09-01 23:59:59 +0000</ts:date>\n" +
	"  </trn:date-posted>\n" +
	"  <trn:splits>\n" +
	"    <trn:split>\n" +
	"      <split:quantity>5/100</split:quantity>\n" +
	"      <split:account type=\"guid\">00000000000000000000000000002000</split:account>\n" +
	"    </trn:split>\n" +
	"    <trn:split>\n" +
	"      <split:quantity>-5/100</split:quantity>\n" +
	"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n" +
	"    </trn:split>\n" +
	"  </trn:splits>\n" +
	"</gnc:transaction>\n" +
	"</gnc:book>\n" +
	"</gnc-v2>\n"

const gnucashCommonValidDataCaseHeader = "<?xml version=\"1.0\" encoding=\"utf-8\" ?>\n" +
	"<gnc-v2\n" +
	"     xmlns:gnc=\"http://www.gnucash.org/XML/gnc\"\n" +
	"     xmlns:act=\"http://www.gnucash.org/XML/act\"\n" +
	"     xmlns:book=\"http://www.gnucash.org/XML/book\"\n" +
	"     xmlns:cd=\"http://www.gnucash.org/XML/cd\"\n" +
	"     xmlns:cmdty=\"http://www.gnucash.org/XML/cmdty\"\n" +
	"     xmlns:slot=\"http://www.gnucash.org/XML/slot\"\n" +
	"     xmlns:split=\"http://www.gnucash.org/XML/split\"\n" +
	"     xmlns:trn=\"http://www.gnucash.org/XML/trn\">\n" +
	"<gnc:book version=\"2.0.0\">\n" +
	"<gnc:account version=\"2.0.0\">\n" +
	"  <act:name>Root Account</act:name>\n" +
	"  <act:id type=\"guid\">00000000000000000000000000000001</act:id>\n" +
	"  <act:type>ROOT</act:type>\n" +
	"</gnc:account>\n" +
	"<gnc:account version=\"2.0.0\">\n" +
	"  <act:name>Opening Balances</act:name>\n" +
	"  <act:id type=\"guid\">00000000000000000000000000000010</act:id>\n" +
	"  <act:type>EQUITY</act:type>\n" +
	"  <act:commodity>\n" +
	"    <cmdty:space>CURRENCY</cmdty:space>\n" +
	"    <cmdty:id>CNY</cmdty:id>\n" +
	"  </act:commodity>\n" +
	"  <act:slots>\n" +
	"    <slot>\n" +
	"      <slot:key>equity-type</slot:key>\n" +
	"      <slot:value type=\"string\">opening-balance</slot:value>\n" +
	"    </slot>\n" +
	"  </act:slots>\n" +
	"</gnc:account>\n" +
	"<gnc:account version=\"2.0.0\">\n" +
	"  <act:name>Test Account</act:name>\n" +
	"  <act:id type=\"guid\">00000000000000000000000000001000</act:id>\n" +
	"  <act:type>BANK</act:type>\n" +
	"  <act:commodity>\n" +
	"    <cmdty:space>CURRENCY</cmdty:space>\n" +
	"    <cmdty:id>CNY</cmdty:id>\n" +
	"  </act:commodity>\n" +
	"  <act:parent type=\"guid\">00000000000000000000000000000001</act:parent>\n" +
	"</gnc:account>\n"

const gnucashCommonValidDataCaseFooter = "</gnc:book>\n" +
	"</gnc-v2>\n"

func TestGnuCashTransactionDatabaseFileParseImportedData_MinimumValidData(t *testing.T) {
	importer := GnuCashTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, allNewAccounts, allNewSubExpenseCategories, allNewSubIncomeCategories, allNewSubTransferCategories, allNewTags, err := importer.ParseImportedData(context, user, []byte(gnucashMinimumValidDataCase), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	checkParsedMinimumValidData(t, allNewTransactions, allNewAccounts, allNewSubExpenseCategories, allNewSubIncomeCategories, allNewSubTransferCategories, allNewTags)
}

func TestGnuCashTransactionDatabaseFileParseImportedData_GzippedMinimumValidData(t *testing.T) {
	importer := GnuCashTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	_, err := gzipWriter.Write([]byte(gnucashMinimumValidDataCase))
	assert.Nil(t, err)

	err = gzipWriter.Close()
	assert.Nil(t, err)

	gzippedData := buffer.Bytes()
	allNewTransactions, allNewAccounts, allNewSubExpenseCategories, allNewSubIncomeCategories, allNewSubTransferCategories, allNewTags, err := importer.ParseImportedData(context, user, gzippedData, time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	checkParsedMinimumValidData(t, allNewTransactions, allNewAccounts, allNewSubExpenseCategories, allNewSubIncomeCategories, allNewSubTransferCategories, allNewTags)
}

func TestGnuCashTransactionDatabaseFileParseImportedData_MinimumValidDataWithReversedSplitOrder(t *testing.T) {
	importer := GnuCashTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, allNewAccounts, allNewSubExpenseCategories, allNewSubIncomeCategories, allNewSubTransferCategories, allNewTags, err := importer.ParseImportedData(context, user, []byte("<?xml version=\"1.0\" encoding=\"utf-8\" ?>\n"+
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
		"  <act:name>Test Category</act:name>\n"+
		"  <act:id type=\"guid\">00000000000000000000000000000100</act:id>\n"+
		"  <act:type>INCOME</act:type>\n"+
		"  <act:parent type=\"guid\">00000000000000000000000000000001</act:parent>\n"+
		"</gnc:account>\n"+
		"<gnc:account version=\"2.0.0\">\n"+
		"  <act:name>Test Category2</act:name>\n"+
		"  <act:id type=\"guid\">00000000000000000000000000000200</act:id>\n"+
		"  <act:type>EXPENSE</act:type>\n"+
		"  <act:parent type=\"guid\">00000000000000000000000000000001</act:parent>\n"+
		"</gnc:account>\n"+
		"<gnc:account version=\"2.0.0\">\n"+
		"  <act:name>Test Account</act:name>\n"+
		"  <act:id type=\"guid\">00000000000000000000000000001000</act:id>\n"+
		"  <act:type>BANK</act:type>\n"+
		"  <act:commodity>\n"+
		"    <cmdty:space>CURRENCY</cmdty:space>\n"+
		"    <cmdty:id>CNY</cmdty:id>\n"+
		"  </act:commodity>\n"+
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
		"    <ts:date>2024-09-01 00:00:00 +0000</ts:date>\n"+
		"  </trn:date-posted>\n"+
		"  <trn:splits>\n"+
		"    <trn:split>\n"+
		"      <split:quantity>-12345/100</split:quantity>\n"+
		"      <split:account type=\"guid\">00000000000000000000000000000010</split:account>\n"+
		"    </trn:split>\n"+
		"    <trn:split>\n"+
		"      <split:quantity>12345/100</split:quantity>\n"+
		"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n"+
		"    </trn:split>\n"+
		"  </trn:splits>\n"+
		"</gnc:transaction>\n"+
		"<gnc:transaction version=\"2.0.0\">\n"+
		"  <trn:date-posted>\n"+
		"    <ts:date>2024-09-01 01:23:45 +0000</ts:date>\n"+
		"  </trn:date-posted>\n"+
		"  <trn:splits>\n"+
		"    <trn:split>\n"+
		"      <split:quantity>-12/100</split:quantity>\n"+
		"      <split:account type=\"guid\">00000000000000000000000000000100</split:account>\n"+
		"    </trn:split>\n"+
		"    <trn:split>\n"+
		"      <split:quantity>12/100</split:quantity>\n"+
		"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n"+
		"    </trn:split>\n"+
		"  </trn:splits>\n"+
		"</gnc:transaction>\n"+
		"<gnc:transaction version=\"2.0.0\">\n"+
		"  <trn:date-posted>\n"+
		"    <ts:date>2024-09-01 12:34:56 +0000</ts:date>\n"+
		"  </trn:date-posted>\n"+
		"  <trn:splits>\n"+
		"    <trn:split>\n"+
		"      <split:quantity>-100/100</split:quantity>\n"+
		"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n"+
		"    </trn:split>\n"+
		"    <trn:split>\n"+
		"      <split:quantity>100/100</split:quantity>\n"+
		"      <split:account type=\"guid\">00000000000000000000000000000200</split:account>\n"+
		"    </trn:split>\n"+
		"  </trn:splits>\n"+
		"</gnc:transaction>\n"+
		"<gnc:transaction version=\"2.0.0\">\n"+
		"  <trn:date-posted>\n"+
		"    <ts:date>2024-09-01 23:59:59 +0000</ts:date>\n"+
		"  </trn:date-posted>\n"+
		"  <trn:splits>\n"+
		"    <trn:split>\n"+
		"      <split:quantity>-5/100</split:quantity>\n"+
		"      <split:account type=\"guid\">00000000000000000000000000001000</split:account>\n"+
		"    </trn:split>\n"+
		"    <trn:split>\n"+
		"      <split:quantity>5/100</split:quantity>\n"+
		"      <split:account type=\"guid\">00000000000000000000000000002000</split:account>\n"+
		"    </trn:split>\n"+
		"  </trn:splits>\n"+
		"</gnc:transaction>\n"+
		"</gnc:book>\n"+
		"</gnc-v2>\n"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	checkParsedMinimumValidData(t, allNewTransactions, allNewAccounts, allNewSubExpenseCategories, allNewSubIncomeCategories, allNewSubTransferCategories, allNewTags)
}
