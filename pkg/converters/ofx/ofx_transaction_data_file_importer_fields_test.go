package ofx

import (
	"github.com/mayswind/ezbookkeeping/pkg/converters/converter"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestOFXTransactionDataFileParseImportedData_ParseAmount_CommaAsDecimalPoint(t *testing.T) {
	importer := OFXTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"<OFX>\n"+
			"  <BANKMSGSRSV1>\n"+
			"    <STMTTRNRS>\n"+
			"      <STMTRS>\n"+
			"        <CURDEF>CNY</CURDEF>\n"+
			"        <BANKACCTFROM>\n"+
			"          <ACCTID>123</ACCTID>\n"+
			"        </BANKACCTFROM>\n"+
			"        <BANKTRANLIST>\n"+
			"          <STMTTRN>\n"+
			"            <TRNTYPE>DEP</TRNTYPE>\n"+
			"            <DTPOSTED>20240901012345.000[+8:CST]</DTPOSTED>\n"+
			"            <TRNAMT>123,45</TRNAMT>\n"+
			"          </STMTTRN>\n"+
			"        </BANKTRANLIST>\n"+
			"      </STMTRS>\n"+
			"    </STMTTRNRS>\n"+
			"  </BANKMSGSRSV1>\n"+
			"</OFX>"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)
}

func TestOFXTransactionDataFileParseImportedData_ParseInvalidAmount(t *testing.T) {
	importer := OFXTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"<OFX>\n"+
			"  <BANKMSGSRSV1>\n"+
			"    <STMTTRNRS>\n"+
			"      <STMTRS>\n"+
			"        <CURDEF>CNY</CURDEF>\n"+
			"        <BANKACCTFROM>\n"+
			"          <ACCTID>123</ACCTID>\n"+
			"        </BANKACCTFROM>\n"+
			"        <BANKTRANLIST>\n"+
			"          <STMTTRN>\n"+
			"            <TRNTYPE>DEP</TRNTYPE>\n"+
			"            <DTPOSTED>20240901012345.000[+8:CST]</DTPOSTED>\n"+
			"            <TRNAMT>123 45</TRNAMT>\n"+
			"          </STMTTRN>\n"+
			"        </BANKTRANLIST>\n"+
			"      </STMTRS>\n"+
			"    </STMTTRNRS>\n"+
			"  </BANKMSGSRSV1>\n"+
			"</OFX>"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)
}

func TestOFXTransactionDataFileParseImportedData_ParseTransactionCurrency(t *testing.T) {
	importer := OFXTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"<OFX>\n"+
			"  <BANKMSGSRSV1>\n"+
			"    <STMTTRNRS>\n"+
			"      <STMTRS>\n"+
			"        <CURDEF>CNY</CURDEF>\n"+
			"        <BANKACCTFROM>\n"+
			"          <ACCTID>123</ACCTID>\n"+
			"        </BANKACCTFROM>\n"+
			"        <BANKTRANLIST>\n"+
			"        <STMTTRN>\n"+
			"            <TRNTYPE>DEP</TRNTYPE>\n"+
			"            <DTPOSTED>20240901012345.000[+8:CST]</DTPOSTED>\n"+
			"            <TRNAMT>123.45</TRNAMT>\n"+
			"            <CURRENCY>USD</CURRENCY>\n"+
			"          </STMTTRN>\n"+
			"        </BANKTRANLIST>\n"+
			"      </STMTRS>\n"+
			"    </STMTTRNRS>\n"+
			"  </BANKMSGSRSV1>\n"+
			"</OFX>"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, "USD", allNewTransactions[0].OriginalSourceAccountCurrency)
}

func TestOFXTransactionDataFileParseImportedData_ParseDescription(t *testing.T) {
	importer := OFXTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"<OFX>\n"+
			"  <BANKMSGSRSV1>\n"+
			"    <STMTTRNRS>\n"+
			"      <STMTRS>\n"+
			"        <CURDEF>CNY</CURDEF>\n"+
			"        <BANKACCTFROM>\n"+
			"          <ACCTID>123</ACCTID>\n"+
			"        </BANKACCTFROM>\n"+
			"        <BANKTRANLIST>\n"+
			"          <STMTTRN>\n"+
			"            <TRNTYPE>DEP</TRNTYPE>\n"+
			"            <DTPOSTED>20240901012345.000[+8:CST]</DTPOSTED>\n"+
			"            <TRNAMT>123.45</TRNAMT>\n"+
			"            <NAME>Test</NAME>\n"+
			"            <MEMO>foo    bar\t#test</MEMO>\n"+
			"          </STMTTRN>\n"+
			"        </BANKTRANLIST>\n"+
			"      </STMTRS>\n"+
			"    </STMTTRNRS>\n"+
			"  </BANKMSGSRSV1>\n"+
			"</OFX>"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, "foo    bar\t#test", allNewTransactions[0].Comment)

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"<OFX>\n"+
			"  <BANKMSGSRSV1>\n"+
			"    <STMTTRNRS>\n"+
			"      <STMTRS>\n"+
			"        <CURDEF>CNY</CURDEF>\n"+
			"        <BANKACCTFROM>\n"+
			"          <ACCTID>123</ACCTID>\n"+
			"        </BANKACCTFROM>\n"+
			"        <BANKTRANLIST>\n"+
			"          <STMTTRN>\n"+
			"            <TRNTYPE>DEP</TRNTYPE>\n"+
			"            <DTPOSTED>20240901012345.000[+8:CST]</DTPOSTED>\n"+
			"            <TRNAMT>123.45</TRNAMT>\n"+
			"            <NAME>Test</NAME>\n"+
			"          </STMTTRN>\n"+
			"        </BANKTRANLIST>\n"+
			"      </STMTRS>\n"+
			"    </STMTTRNRS>\n"+
			"  </BANKMSGSRSV1>\n"+
			"</OFX>"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, "Test", allNewTransactions[0].Comment)

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"<OFX>\n"+
			"  <BANKMSGSRSV1>\n"+
			"    <STMTTRNRS>\n"+
			"      <STMTRS>\n"+
			"        <CURDEF>CNY</CURDEF>\n"+
			"        <BANKACCTFROM>\n"+
			"          <ACCTID>123</ACCTID>\n"+
			"        </BANKACCTFROM>\n"+
			"        <BANKTRANLIST>\n"+
			"          <STMTTRN>\n"+
			"            <TRNTYPE>DEP</TRNTYPE>\n"+
			"            <DTPOSTED>20240901012345.000[+8:CST]</DTPOSTED>\n"+
			"            <TRNAMT>123.45</TRNAMT>\n"+
			"            <PAYEE>\n"+
			"              <NAME>Test</NAME>\n"+
			"            </PAYEE>\n"+
			"          </STMTTRN>\n"+
			"        </BANKTRANLIST>\n"+
			"      </STMTRS>\n"+
			"    </STMTTRNRS>\n"+
			"  </BANKMSGSRSV1>\n"+
			"</OFX>"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, "Test", allNewTransactions[0].Comment)
}

func TestOFXTransactionDataFileParseImportedData_MissingAccountFromNode(t *testing.T) {
	importer := OFXTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1,
		DefaultCurrency: "CNY",
	}

	// Missing Posted Date Node
	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"<OFX>\n"+
			"  <BANKMSGSRSV1>\n"+
			"    <STMTTRNRS>\n"+
			"      <STMTRS>\n"+
			"        <CURDEF>CNY</CURDEF>\n"+
			"        <BANKTRANLIST>\n"+
			"          <STMTTRN>\n"+
			"            <TRNTYPE>DEP</TRNTYPE>\n"+
			"            <DTPOSTED>20240901012345.000[+8:CST]</DTPOSTED>\n"+
			"            <TRNAMT>123.45</TRNAMT>\n"+
			"          </STMTTRN>\n"+
			"        </BANKTRANLIST>\n"+
			"      </STMTRS>\n"+
			"    </STMTTRNRS>\n"+
			"  </BANKMSGSRSV1>\n"+
			"</OFX>"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingAccountData.Message)
}

func TestOFXTransactionDataFileParseImportedData_MissingCurrencyNode(t *testing.T) {
	importer := OFXTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1,
		DefaultCurrency: "CNY",
	}

	// Missing Default Currency Node
	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"<OFX>\n"+
			"    <BANKMSGSRSV1>\n"+
			"    <STMTTRNRS>\n"+
			"      <STMTRS>\n"+
			"        <BANKACCTFROM>\n"+
			"          <ACCTID>123</ACCTID>\n"+
			"        </BANKACCTFROM>\n"+
			"        <BANKTRANLIST>\n"+
			"          <STMTTRN>\n"+
			"            <TRNTYPE>DEP</TRNTYPE>\n"+
			"            <DTPOSTED>20240901012345.000[+8:CST]</DTPOSTED>\n"+
			"            <TRNAMT>123.45</TRNAMT>\n"+
			"          </STMTTRN>\n"+
			"        </BANKTRANLIST>\n"+
			"      </STMTRS>\n"+
			"    </STMTTRNRS>\n"+
			"  </BANKMSGSRSV1>\n"+
			"</OFX>"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAccountCurrencyInvalid.Message)
}

func TestOFXTransactionDataFileParseImportedData_MissingTransactionRequiredNode(t *testing.T) {
	importer := OFXTransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1,
		DefaultCurrency: "CNY",
	}

	// Missing Posted Date Node
	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		"<OFX>\n"+
			"    <BANKMSGSRSV1>\n"+
			"    <STMTTRNRS>\n"+
			"      <STMTRS>\n"+
			"        <CURDEF>CNY</CURDEF>\n"+
			"        <BANKACCTFROM>\n"+
			"          <ACCTID>123</ACCTID>\n"+
			"        </BANKACCTFROM>\n"+
			"        <BANKTRANLIST>\n"+
			"          <STMTTRN>\n"+
			"            <TRNTYPE>DEP</TRNTYPE>\n"+
			"            <TRNAMT>123.45</TRNAMT>\n"+
			"          </STMTTRN>\n"+
			"        </BANKTRANLIST>\n"+
			"      </STMTRS>\n"+
			"    </STMTTRNRS>\n"+
			"  </BANKMSGSRSV1>\n"+
			"</OFX>"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingTransactionTime.Message)

	// Missing Transaction Type Node
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"<OFX>\n"+
			"    <BANKMSGSRSV1>\n"+
			"    <STMTTRNRS>\n"+
			"      <STMTRS>\n"+
			"        <CURDEF>CNY</CURDEF>\n"+
			"        <BANKACCTFROM>\n"+
			"          <ACCTID>123</ACCTID>\n"+
			"        </BANKACCTFROM>\n"+
			"        <BANKTRANLIST>\n"+
			"          <STMTTRN>\n"+
			"            <DTPOSTED>20240901012345.000[+8:CST]</DTPOSTED>\n"+
			"            <TRNAMT>123.45</TRNAMT>\n"+
			"          </STMTTRN>\n"+
			"        </BANKTRANLIST>\n"+
			"      </STMTRS>\n"+
			"    </STMTTRNRS>\n"+
			"  </BANKMSGSRSV1>\n"+
			"</OFX>"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTypeInvalid.Message)

	// Missing Amount Node
	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		"<OFX>\n"+
			"    <BANKMSGSRSV1>\n"+
			"    <STMTTRNRS>\n"+
			"      <STMTRS>\n"+
			"        <CURDEF>CNY</CURDEF>\n"+
			"        <BANKACCTFROM>\n"+
			"          <ACCTID>123</ACCTID>\n"+
			"        </BANKACCTFROM>\n"+
			"        <BANKTRANLIST>\n"+
			"          <STMTTRN>\n"+
			"            <TRNTYPE>DEP</TRNTYPE>\n"+
			"            <DTPOSTED>20240901012345.000[+8:CST]</DTPOSTED>\n"+
			"          </STMTTRN>\n"+
			"        </BANKTRANLIST>\n"+
			"      </STMTRS>\n"+
			"    </STMTTRNRS>\n"+
			"  </BANKMSGSRSV1>\n"+
			"</OFX>"), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)
}
