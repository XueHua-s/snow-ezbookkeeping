package beancount

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBeancountDataReaderRead(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""+
		"; Test Beancount Data\n"+
		"option \"name_assets\" \"AssetsAccount\"\n"+
		"option \"name_liabilities\" \"LiabilitiesAccount\"\n"+
		"option \"name_equity\" \"EquityAccount\"\n"+
		"option \"name_income\" \"IncomeAccount\"\n"+
		"option \"name_expenses\" \"ExpensesAccount\"\n"+
		"\n"+
		"2024-01-01 open AssetsAccount:TestAccount\n"+
		"2024-01-02 open LiabilitiesAccount:TestAccount2\n"+
		"2024-01-03 open EquityAccount:Opening-Balances\n"+
		"\n"+
		"; The following transactions with tag1 and tag2\n"+
		"pushtag #tag1\n"+
		"pushtag #tag2\n"+
		"\n"+
		"2024-01-05 * \"Payee Name\" \"Foo Bar\" #tag3 #tag4 ^test-link\n"+
		"  IncomeAccount:TestCategory -123.45 CNY\n"+
		"  AssetsAccount:TestAccount 123.45 CNY\n"+
		"; The following transactions with tag2\n"+
		"poptag #tag1\n"+
		"2024-01-06 * \"test\n#test2\" #tag5 #tag6 ^test-link2\n"+
		"  LiabilitiesAccount:TestAccount2 -0.12 USD\n"+
		"  ExpensesAccount:TestCategory2 0.12 USD\n"+
		"2024-01-07 close AssetsAccount:TestAccount\n"))
	assert.Nil(t, err)

	actualData, err := reader.read(context)
	assert.Nil(t, err)

	assert.Equal(t, 5, len(actualData.Accounts))
	assert.Equal(t, "AssetsAccount:TestAccount", actualData.Accounts["AssetsAccount:TestAccount"].Name)
	assert.Equal(t, beancountAssetsAccountType, actualData.Accounts["AssetsAccount:TestAccount"].AccountType)
	assert.Equal(t, "2024-01-01", actualData.Accounts["AssetsAccount:TestAccount"].OpenDate)
	assert.Equal(t, "2024-01-07", actualData.Accounts["AssetsAccount:TestAccount"].CloseDate)

	assert.Equal(t, "LiabilitiesAccount:TestAccount2", actualData.Accounts["LiabilitiesAccount:TestAccount2"].Name)
	assert.Equal(t, beancountLiabilitiesAccountType, actualData.Accounts["LiabilitiesAccount:TestAccount2"].AccountType)
	assert.Equal(t, "2024-01-02", actualData.Accounts["LiabilitiesAccount:TestAccount2"].OpenDate)

	assert.Equal(t, 2, len(actualData.Transactions))

	assert.Equal(t, "2024-01-05", actualData.Transactions[0].Date)
	assert.Equal(t, "Payee Name", actualData.Transactions[0].Payee)
	assert.Equal(t, "Foo Bar", actualData.Transactions[0].Narration)
	assert.Equal(t, 2, len(actualData.Transactions[0].Postings))
	assert.Equal(t, "IncomeAccount:TestCategory", actualData.Transactions[0].Postings[0].Account)
	assert.Equal(t, "-123.45", actualData.Transactions[0].Postings[0].Amount)
	assert.Equal(t, "CNY", actualData.Transactions[0].Postings[0].Commodity)
	assert.Equal(t, "AssetsAccount:TestAccount", actualData.Transactions[0].Postings[1].Account)
	assert.Equal(t, "123.45", actualData.Transactions[0].Postings[1].Amount)
	assert.Equal(t, "CNY", actualData.Transactions[0].Postings[1].Commodity)

	assert.Equal(t, 4, len(actualData.Transactions[0].Tags))
	assert.Equal(t, actualData.Transactions[0].Tags[0], "tag1")
	assert.Equal(t, actualData.Transactions[0].Tags[1], "tag2")
	assert.Equal(t, actualData.Transactions[0].Tags[2], "tag3")
	assert.Equal(t, actualData.Transactions[0].Tags[3], "tag4")

	assert.Equal(t, 1, len(actualData.Transactions[0].Links))
	assert.Equal(t, actualData.Transactions[0].Links[0], "test-link")

	assert.Equal(t, "2024-01-06", actualData.Transactions[1].Date)
	assert.Equal(t, "", actualData.Transactions[1].Payee)
	assert.Equal(t, "test\n#test2", actualData.Transactions[1].Narration)
	assert.Equal(t, 2, len(actualData.Transactions[1].Postings))
	assert.Equal(t, "LiabilitiesAccount:TestAccount2", actualData.Transactions[1].Postings[0].Account)
	assert.Equal(t, "-0.12", actualData.Transactions[1].Postings[0].Amount)
	assert.Equal(t, "USD", actualData.Transactions[1].Postings[0].Commodity)
	assert.Equal(t, "ExpensesAccount:TestCategory2", actualData.Transactions[1].Postings[1].Account)
	assert.Equal(t, "0.12", actualData.Transactions[1].Postings[1].Amount)
	assert.Equal(t, "USD", actualData.Transactions[1].Postings[1].Commodity)

	assert.Equal(t, 3, len(actualData.Transactions[1].Tags))
	assert.Equal(t, actualData.Transactions[1].Tags[0], "tag2")
	assert.Equal(t, actualData.Transactions[1].Tags[1], "tag5")
	assert.Equal(t, actualData.Transactions[1].Tags[2], "tag6")

	assert.Equal(t, 1, len(actualData.Transactions[1].Links))
	assert.Equal(t, actualData.Transactions[1].Links[0], "test-link2")
}

func TestBeancountDataReaderRead_EmptyContent(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""))
	assert.Nil(t, err)

	_, err = reader.read(context)
	assert.EqualError(t, err, errs.ErrNotFoundTransactionDataInFile.Message)
}

func TestBeancountDataReaderRead_UnsupportedInclude(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte("include \"other.beancount\""))
	assert.Nil(t, err)

	_, err = reader.read(context)
	assert.EqualError(t, err, errs.ErrBeancountFileNotSupportInclude.Message)
}

func TestBeancountDataReaderRead_SkipUnsupportedDirective(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""+
		"plugin \"beancount.plugins.plugin_name\"\n"+
		"unknown directive\n"+
		"2024-01-01 commodity USD\n"+
		"2024-01-01 price USD 1.08 CAD\n"+
		"2024-01-01 note Assets:Test \"some text\"\n"+
		"2024-01-01 document Assets:Test \"scheme://path\"\n"+
		"2024-01-01 event \"location\" \"address\"\n"+
		"2024-01-01 balance Assets:Test 100.00 USD\n"+
		"2024-01-01 pad Assets:Test Equity:Opening-Balances\n"+
		"2024-01-01 query \"Name\" \"\nSELECT FIELDS FROM TABLE\"\n"+
		"2024-01-01 custom \"Type\" \"Value\"\n"+
		"2024-01-01 unknown directive\n"))
	assert.Nil(t, err)

	_, err = reader.read(context)
	assert.Nil(t, err)
}

func TestBeancountDataReaderReadAndSetOption_AccountTypeName(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""+
		"option \"name_assets\" \"A\"\n"+
		"option \"name_liabilities\" \"L\"\n"+
		"option \"name_equity\" \"E\"\n"+
		"\n"+
		"2024-01-01 open A:TestAccount\n"+
		"2024-01-02 open L:TestAccount2\n"+
		"2024-01-03 open E:Opening-Balances\n"))
	assert.Nil(t, err)

	actualData, err := reader.read(context)
	assert.Nil(t, err)

	assert.Equal(t, 3, len(actualData.Accounts))

	assert.Equal(t, "A:TestAccount", actualData.Accounts["A:TestAccount"].Name)
	assert.Equal(t, beancountAssetsAccountType, actualData.Accounts["A:TestAccount"].AccountType)

	assert.Equal(t, "L:TestAccount2", actualData.Accounts["L:TestAccount2"].Name)
	assert.Equal(t, beancountLiabilitiesAccountType, actualData.Accounts["L:TestAccount2"].AccountType)

	assert.Equal(t, "E:Opening-Balances", actualData.Accounts["E:Opening-Balances"].Name)
	assert.Equal(t, beancountEquityAccountType, actualData.Accounts["E:Opening-Balances"].AccountType)
	assert.True(t, actualData.Accounts["E:Opening-Balances"].isOpeningBalanceEquityAccount())
}

func TestBeancountDataReaderReadAndSetOption_InvalidLineOrUnsupportedOption(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""+
		"option \"test\" \"Test\" \"Test2\"\n"+
		"option \"test\" \"Test\"\n"+
		"option \"test\"\n"+
		"option \n"+
		"option\n"))
	assert.Nil(t, err)

	_, err = reader.read(context)
	assert.Nil(t, err)
}

func TestBeancountDataReaderReadAndSetTags(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""+
		"pushtag #tag1\n"+
		"pushtag #tag2\n"+
		"pushtag #tag2\n"+
		"pushtag #tag1\n"+
		"\n"+
		"2024-01-01 * #tag3 #tag4\n"+
		"poptag #tag1\n"+
		"poptag #tag2\n"+
		"pushtag\n"+
		"pushtag    \n"+
		"pushtag tag\n"+
		"2024-01-02 * #tag5 #tag6\n"+
		"poptag #tag1\n"+
		"poptag #tag2\n"+
		"poptag\n"+
		"poptag \n"+
		"2024-01-03 * #tag5 #tag6\n"+
		"pushtag #tag3\n"+
		"pushtag #tag6\n"+
		"2024-01-04 * #tag5 #tag6\n"+
		"2024-01-05 * #tag5 #tag6 #tag6 #tag5\n"))
	assert.Nil(t, err)

	actualData, err := reader.read(context)
	assert.Nil(t, err)

	assert.Equal(t, 5, len(actualData.Transactions))

	assert.Equal(t, 4, len(actualData.Transactions[0].Tags))
	assert.Equal(t, actualData.Transactions[0].Tags[0], "tag1")
	assert.Equal(t, actualData.Transactions[0].Tags[1], "tag2")
	assert.Equal(t, actualData.Transactions[0].Tags[2], "tag3")
	assert.Equal(t, actualData.Transactions[0].Tags[3], "tag4")

	assert.Equal(t, 2, len(actualData.Transactions[1].Tags))
	assert.Equal(t, actualData.Transactions[1].Tags[0], "tag5")
	assert.Equal(t, actualData.Transactions[1].Tags[1], "tag6")

	assert.Equal(t, 2, len(actualData.Transactions[2].Tags))
	assert.Equal(t, actualData.Transactions[2].Tags[0], "tag5")
	assert.Equal(t, actualData.Transactions[2].Tags[1], "tag6")

	assert.Equal(t, 3, len(actualData.Transactions[3].Tags))
	assert.Equal(t, actualData.Transactions[3].Tags[0], "tag3")
	assert.Equal(t, actualData.Transactions[3].Tags[1], "tag6")
	assert.Equal(t, actualData.Transactions[3].Tags[2], "tag5")

	assert.Equal(t, 3, len(actualData.Transactions[4].Tags))
	assert.Equal(t, actualData.Transactions[4].Tags[0], "tag3")
	assert.Equal(t, actualData.Transactions[4].Tags[1], "tag6")
	assert.Equal(t, actualData.Transactions[4].Tags[2], "tag5")
}

func TestBeancountDataReaderReadAccountLine_InvalidLine(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""+
		"2024-01-01 open\n"))
	assert.Nil(t, err)

	actualData, err := reader.read(context)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(actualData.Accounts))
}

func TestBeancountDataReaderReadAccountLine_InvalidAccountType(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""+
		"2024-01-01 open Test:TestAccount\n"))
	assert.Nil(t, err)

	_, err = reader.read(context)
	assert.EqualError(t, err, errs.ErrInvalidBeancountFile.Message)

	reader, err = createNewBeancountDataReader(context, []byte(""+
		"option \"name_assets\" \"A\"\n"+
		"\n"+
		"2024-01-01 open Assets:TestAccount\n"))
	assert.Nil(t, err)

	_, err = reader.read(context)
	assert.EqualError(t, err, errs.ErrInvalidBeancountFile.Message)
}
