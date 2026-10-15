package beancount

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBeancountDataReaderReadTransactionLine(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""+
		"2024-01-01 *\n"+
		"2024-01-02 * \"test\ttest2\ntest3\" ; some comment\n"+
		"2024-01-03 ! \"test\" \"test2\"\n"+
		"2024-01-04 P \"test\" #tag #tag2 ; some comment\n"+
		"2024-01-05 txn \"test\" ^scheme://path/to/test/link ; some comment\n"+
		"2024-01-06 txn ; \"test\" \"test2\" #tag ^link\n"))
	assert.Nil(t, err)

	actualData, err := reader.read(context)
	assert.Nil(t, err)

	assert.Equal(t, 6, len(actualData.Transactions))

	assert.Equal(t, "2024-01-01", actualData.Transactions[0].Date)
	assert.Equal(t, beancountDirectiveCompletedTransaction, actualData.Transactions[0].Directive)
	assert.Equal(t, "", actualData.Transactions[0].Payee)
	assert.Equal(t, "", actualData.Transactions[0].Narration)

	assert.Equal(t, "2024-01-02", actualData.Transactions[1].Date)
	assert.Equal(t, beancountDirectiveCompletedTransaction, actualData.Transactions[1].Directive)
	assert.Equal(t, "", actualData.Transactions[1].Payee)
	assert.Equal(t, "test\ttest2\ntest3", actualData.Transactions[1].Narration)

	assert.Equal(t, "2024-01-03", actualData.Transactions[2].Date)
	assert.Equal(t, beancountDirectiveInCompleteTransaction, actualData.Transactions[2].Directive)
	assert.Equal(t, "test", actualData.Transactions[2].Payee)
	assert.Equal(t, "test2", actualData.Transactions[2].Narration)

	assert.Equal(t, "2024-01-04", actualData.Transactions[3].Date)
	assert.Equal(t, beancountDirectivePaddingTransaction, actualData.Transactions[3].Directive)
	assert.Equal(t, "", actualData.Transactions[3].Payee)
	assert.Equal(t, "test", actualData.Transactions[3].Narration)

	assert.Equal(t, 2, len(actualData.Transactions[3].Tags))
	assert.Equal(t, actualData.Transactions[3].Tags[0], "tag")
	assert.Equal(t, actualData.Transactions[3].Tags[1], "tag2")

	assert.Equal(t, "2024-01-05", actualData.Transactions[4].Date)
	assert.Equal(t, beancountDirectiveTransaction, actualData.Transactions[4].Directive)
	assert.Equal(t, "", actualData.Transactions[4].Payee)
	assert.Equal(t, "test", actualData.Transactions[4].Narration)

	assert.Equal(t, 1, len(actualData.Transactions[4].Links))
	assert.Equal(t, actualData.Transactions[4].Links[0], "scheme://path/to/test/link")

	assert.Equal(t, "2024-01-06", actualData.Transactions[5].Date)
	assert.Equal(t, beancountDirectiveTransaction, actualData.Transactions[5].Directive)
	assert.Equal(t, "", actualData.Transactions[5].Payee)
	assert.Equal(t, "", actualData.Transactions[5].Narration)
}

func TestBeancountDataReaderReadTransactionPostingLine(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""+
		"2024-01-01 *\n"+
		"  Income:TestCategory -123.45 CNY ; some comment\n"+
		"  Assets:TestAccount 123.45   CNY\n"+
		"2024-01-02 *\n"+
		"  Liabilities:TestAccount2 -0.23 USD ; some comment\n"+
		"  Expenses:TestCategory2    0.12 USD @@ 0.84 CNY\n"+
		"  Expenses:TestCategory3    0.11 USD @ 7.12 CNY\n"+
		"  ! Expenses:TestCategory4  0.00 USD {0.00 CNY}\n"+
		"  Expenses:TestCategory5 \n"))
	assert.Nil(t, err)

	actualData, err := reader.read(context)
	assert.Nil(t, err)

	assert.Equal(t, 2, len(actualData.Transactions))

	assert.Equal(t, "2024-01-01", actualData.Transactions[0].Date)
	assert.Equal(t, 2, len(actualData.Transactions[0].Postings))
	assert.Equal(t, "Income:TestCategory", actualData.Transactions[0].Postings[0].Account)
	assert.Equal(t, "-123.45", actualData.Transactions[0].Postings[0].Amount)
	assert.Equal(t, "CNY", actualData.Transactions[0].Postings[0].Commodity)

	assert.Equal(t, "Assets:TestAccount", actualData.Transactions[0].Postings[1].Account)
	assert.Equal(t, "123.45", actualData.Transactions[0].Postings[1].Amount)
	assert.Equal(t, "CNY", actualData.Transactions[0].Postings[1].Commodity)

	assert.Equal(t, "2024-01-02", actualData.Transactions[1].Date)
	assert.Equal(t, 4, len(actualData.Transactions[1].Postings))

	assert.Equal(t, "Liabilities:TestAccount2", actualData.Transactions[1].Postings[0].Account)
	assert.Equal(t, "-0.23", actualData.Transactions[1].Postings[0].Amount)
	assert.Equal(t, "USD", actualData.Transactions[1].Postings[0].Commodity)
	assert.Equal(t, "Expenses:TestCategory2", actualData.Transactions[1].Postings[1].Account)

	assert.Equal(t, "0.12", actualData.Transactions[1].Postings[1].Amount)
	assert.Equal(t, "USD", actualData.Transactions[1].Postings[1].Commodity)
	assert.Equal(t, "0.84", actualData.Transactions[1].Postings[1].TotalCost)
	assert.Equal(t, "CNY", actualData.Transactions[1].Postings[1].TotalCostCommodity)
	assert.Equal(t, "Expenses:TestCategory3", actualData.Transactions[1].Postings[2].Account)

	assert.Equal(t, "0.11", actualData.Transactions[1].Postings[2].Amount)
	assert.Equal(t, "USD", actualData.Transactions[1].Postings[2].Commodity)
	assert.Equal(t, "7.12", actualData.Transactions[1].Postings[2].Price)
	assert.Equal(t, "CNY", actualData.Transactions[1].Postings[2].PriceCommodity)

	assert.Equal(t, "0.00", actualData.Transactions[1].Postings[3].Amount)
	assert.Equal(t, "USD", actualData.Transactions[1].Postings[3].Commodity)
}

func TestBeancountDataReaderReadTransactionPostingLine_AmountExpression(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""+
		"2024-01-01 *\n"+
		"  Income:TestCategory (1.2-3.4) * 5.6 / 7.8 CNY\n"+
		"  Assets:TestAccount 1.2 * 3.4/-5.6 - 7.8 CNY\n"))
	assert.Nil(t, err)

	actualData, err := reader.read(context)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(actualData.Transactions))

	assert.Equal(t, "2024-01-01", actualData.Transactions[0].Date)
	assert.Equal(t, 2, len(actualData.Transactions[0].Postings))
	assert.Equal(t, "Income:TestCategory", actualData.Transactions[0].Postings[0].Account)
	assert.Equal(t, "(1.2-3.4) * 5.6 / 7.8", actualData.Transactions[0].Postings[0].OriginalAmount)
	assert.Equal(t, "-1.57", actualData.Transactions[0].Postings[0].Amount)
	assert.Equal(t, "CNY", actualData.Transactions[0].Postings[0].Commodity)

	assert.Equal(t, "Assets:TestAccount", actualData.Transactions[0].Postings[1].Account)
	assert.Equal(t, "1.2 * 3.4/-5.6 - 7.8", actualData.Transactions[0].Postings[1].OriginalAmount)
	assert.Equal(t, "-8.52", actualData.Transactions[0].Postings[1].Amount)
	assert.Equal(t, "CNY", actualData.Transactions[0].Postings[1].Commodity)
}

func TestBeancountDataReaderReadTransactionPostingLine_InvalidAmountExpression(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""+
		"2024-01-01 *\n"+
		"  Income:TestCategory (1.2-3.4)*5.6/0 CNY\n"))
	assert.Nil(t, err)

	_, err = reader.read(context)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)

	reader, err = createNewBeancountDataReader(context, []byte(""+
		"2024-01-01 *\n"+
		"  Assets:TestAccount abc CNY\n"))
	assert.Nil(t, err)

	_, err = reader.read(context)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)

	reader, err = createNewBeancountDataReader(context, []byte(""+
		"2024-01-01 *\n"+
		"  Assets:TestAccount 9999999999999.99*2 CNY\n"))
	assert.Nil(t, err)

	_, err = reader.read(context)
	assert.EqualError(t, err, errs.ErrNumericOverflow.Message)
}

func TestBeancountDataReaderReadTransactionPostingLine_InvalidAccountType(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""+
		"2024-01-01 *\n"+
		"  Income:TestCategory -123.45 CNY\n"+
		"  Test:TestAccount 123.45 CNY\n"))
	assert.Nil(t, err)

	_, err = reader.read(context)
	assert.EqualError(t, err, errs.ErrInvalidBeancountFile.Message)
}

func TestBeancountDataReaderReadTransactionPostingLine_InvalidCommodity(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""+
		"2024-01-01 *\n"+
		"  Income:TestCategory -123.45 cny\n"+
		"  Assets:TestAccount 123.45 cny\n"))
	assert.Nil(t, err)

	_, err = reader.read(context)
	assert.EqualError(t, err, errs.ErrInvalidBeancountFile.Message)
}

func TestBeancountDataReaderReadTransactionPostingLine_MissingAmount(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""+
		"2024-01-01 *\n"+
		"  Assets:TestAccount\n"))
	assert.Nil(t, err)

	actualData, err := reader.read(context)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actualData.Transactions))
	assert.Equal(t, 0, len(actualData.Transactions[0].Postings))

	reader, err = createNewBeancountDataReader(context, []byte(""+
		"2024-01-01 *\n"+
		"  Assets:TestAccount \n"))
	assert.Nil(t, err)

	actualData, err = reader.read(context)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actualData.Transactions))
	assert.Equal(t, 0, len(actualData.Transactions[0].Postings))
}

func TestBeancountDataReaderReadTransactionPostingLine_MissingCommodity(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""+
		"2024-01-01 *\n"+
		"  Assets:TestAccount 123.45\n"))
	assert.Nil(t, err)

	_, err = reader.read(context)
	assert.EqualError(t, err, errs.ErrInvalidBeancountFile.Message)

	reader, err = createNewBeancountDataReader(context, []byte(""+
		"2024-01-01 *\n"+
		"  Assets:TestAccount 123.45 \n"))
	assert.Nil(t, err)

	_, err = reader.read(context)
	assert.EqualError(t, err, errs.ErrInvalidBeancountFile.Message)
}

func TestBeancountDataReaderReadTransactionMetadataLine(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewBeancountDataReader(context, []byte(""+
		"2024-01-01 *\n"+
		"  key: value\n"+
		"  key2: \"value 2\"\n"+
		"  key3: \n"+
		"  key4: \"\"\n"+
		"  key5 : \"\"\n"+
		"  key2: \"new value\"\n"+
		"  Income:TestCategory -123.45 CNY\n"+
		"  Assets:TestAccount 123.45 CNY\n"+
		"2024-01-02 *\n"+
		"  Liabilities:TestAccount2 -0.23 USD\n"+
		"    key6: value6\n"+
		"    key7: \"value 7\"\n"+
		"    key8: \n"+
		"    key9: \"\"\n"+
		"    key0 : \"\"\n"+
		"    key6: \"new value\"\n"+
		"  Expenses:TestCategory2 0.12 USD\n"))
	assert.Nil(t, err)

	actualData, err := reader.read(context)
	assert.Nil(t, err)

	assert.Equal(t, 2, len(actualData.Transactions))

	assert.Equal(t, "2024-01-01", actualData.Transactions[0].Date)
	assert.Equal(t, 2, len(actualData.Transactions[0].Postings))
	assert.Equal(t, 2, len(actualData.Transactions[0].Metadata))
	assert.Equal(t, "value", actualData.Transactions[0].Metadata["key"])
	assert.Equal(t, "value 2", actualData.Transactions[0].Metadata["key2"])

	assert.Equal(t, "2024-01-02", actualData.Transactions[1].Date)
	assert.Equal(t, 2, len(actualData.Transactions[1].Postings))
	assert.Equal(t, 2, len(actualData.Transactions[1].Postings[0].Metadata))
	assert.Equal(t, "value6", actualData.Transactions[1].Postings[0].Metadata["key6"])
	assert.Equal(t, "value 7", actualData.Transactions[1].Postings[0].Metadata["key7"])
	assert.Equal(t, 0, len(actualData.Transactions[1].Postings[1].Metadata))
}
