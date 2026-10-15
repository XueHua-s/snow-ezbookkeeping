package ofx

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateNewOFXFileReader_OFX1WithEndElement(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewOFXFileReader(context, []byte(
		"OFXHEADER:100\n"+
			"DATA:OFXSGML\n"+
			"VERSION:103\n"+
			"SECURITY:NONE\n"+
			"ENCODING:USASCII\n"+
			"CHARSET:1252\n"+
			"COMPRESSION:NONE\n"+
			"OLDFILEUID:NONE\n"+
			"NEWFILEUID:NONE\n"+
			"\n"+
			"<OFX>\n"+
			"<BANKMSGSRSV1>\n"+
			"<STMTTRNRS>\n"+
			"<STMTRS>\n"+
			"<CURDEF>CNY</CURDEF>\n"+
			"<BANKACCTFROM>\n"+
			"<ACCTID>123</ACCTID>\n"+
			"</BANKACCTFROM>\n"+
			"<BANKTRANLIST>\n"+
			"<STMTTRN>\n"+
			"<TRNTYPE>DEP</TRNTYPE>\n"+
			"<DTPOSTED>20240901012345.000[+8:CST]</DTPOSTED>\n"+
			"<TRNAMT>123.45</TRNAMT>\n"+
			"</STMTTRN>\n"+
			"</BANKTRANLIST>\n"+
			"</STMTRS>\n"+
			"</STMTTRNRS>\n"+
			"</BANKMSGSRSV1>\n"+
			"</OFX>"))

	assert.Nil(t, err)

	ofxFile, err := reader.read(context)
	assert.Nil(t, err)
	assert.NotNil(t, ofxFile)

	assert.NotNil(t, ofxFile.FileHeader)
	assert.Equal(t, ofxVersion1, ofxFile.FileHeader.OFXDeclarationVersion)
	assert.Equal(t, "103", ofxFile.FileHeader.OFXDataVersion)
	assert.Equal(t, "NONE", ofxFile.FileHeader.Security)
	assert.Equal(t, "NONE", ofxFile.FileHeader.OldFileUid)
	assert.Equal(t, "NONE", ofxFile.FileHeader.NewFileUid)

	assert.NotNil(t, ofxFile.BankMessageResponseV1)
	assert.NotNil(t, ofxFile.BankMessageResponseV1.StatementTransactionResponse)
	assert.NotNil(t, ofxFile.BankMessageResponseV1.StatementTransactionResponse.StatementResponse)

	assert.Equal(t, "CNY", ofxFile.BankMessageResponseV1.StatementTransactionResponse.StatementResponse.DefaultCurrency)

	assert.NotNil(t, ofxFile.BankMessageResponseV1.StatementTransactionResponse.StatementResponse.AccountFrom)
	assert.Equal(t, "123", ofxFile.BankMessageResponseV1.StatementTransactionResponse.StatementResponse.AccountFrom.AccountId)

	assert.Equal(t, 1, len(ofxFile.BankMessageResponseV1.StatementTransactionResponse.StatementResponse.TransactionList.StatementTransactions))
	assert.Equal(t, ofxDepositTransaction, ofxFile.BankMessageResponseV1.StatementTransactionResponse.StatementResponse.TransactionList.StatementTransactions[0].TransactionType)
	assert.Equal(t, "20240901012345.000[+8:CST]", ofxFile.BankMessageResponseV1.StatementTransactionResponse.StatementResponse.TransactionList.StatementTransactions[0].PostedDate)
	assert.Equal(t, "123.45", ofxFile.BankMessageResponseV1.StatementTransactionResponse.StatementResponse.TransactionList.StatementTransactions[0].Amount)
}

func TestCreateNewOFXFileReader_OFX1WithBlanklinesInHeader(t *testing.T) {
	context := core.NewNullContext()
	reader, err := createNewOFXFileReader(context, []byte(
		"\n"+
			"\n"+
			"OFXHEADER:100\n"+
			"DATA:OFXSGML\n"+
			"VERSION:103\n"+
			"SECURITY:NONE\n"+
			"ENCODING:USASCII\n"+
			"CHARSET:1252\n"+
			"COMPRESSION:NONE\n"+
			"OLDFILEUID:NONE\n"+
			"NEWFILEUID:NONE\n"+
			"\n"+
			"<OFX>"+
			"<BANKMSGSRSV1>"+
			"<STMTTRNRS>"+
			"<STMTRS>"+
			"<CURDEF>CNY"+
			"<BANKACCTFROM>"+
			"<ACCTID>123"+
			"</BANKACCTFROM>"+
			"<BANKTRANLIST>"+
			"<STMTTRN>"+
			"<TRNTYPE>DEP"+
			"<DTPOSTED>20240901012345.000[+8:CST]"+
			"<TRNAMT>123.45"+
			"</STMTTRN>"+
			"</BANKTRANLIST>"+
			"</STMTRS>"+
			"</STMTTRNRS>"+
			"</BANKMSGSRSV1>"+
			"</OFX>"))

	assert.Nil(t, err)

	ofxFile, err := reader.read(context)
	assert.Nil(t, err)
	assert.NotNil(t, ofxFile)

	assert.NotNil(t, ofxFile.FileHeader)
	assert.Equal(t, ofxVersion1, ofxFile.FileHeader.OFXDeclarationVersion)
	assert.Equal(t, "103", ofxFile.FileHeader.OFXDataVersion)
	assert.Equal(t, "NONE", ofxFile.FileHeader.Security)
	assert.Equal(t, "NONE", ofxFile.FileHeader.OldFileUid)
	assert.Equal(t, "NONE", ofxFile.FileHeader.NewFileUid)

	assert.NotNil(t, ofxFile.BankMessageResponseV1)
	assert.NotNil(t, ofxFile.BankMessageResponseV1.StatementTransactionResponse)
	assert.NotNil(t, ofxFile.BankMessageResponseV1.StatementTransactionResponse.StatementResponse)

	assert.Equal(t, "CNY", ofxFile.BankMessageResponseV1.StatementTransactionResponse.StatementResponse.DefaultCurrency)

	assert.NotNil(t, ofxFile.BankMessageResponseV1.StatementTransactionResponse.StatementResponse.AccountFrom)
	assert.Equal(t, "123", ofxFile.BankMessageResponseV1.StatementTransactionResponse.StatementResponse.AccountFrom.AccountId)

	assert.Equal(t, 1, len(ofxFile.BankMessageResponseV1.StatementTransactionResponse.StatementResponse.TransactionList.StatementTransactions))
	assert.Equal(t, ofxDepositTransaction, ofxFile.BankMessageResponseV1.StatementTransactionResponse.StatementResponse.TransactionList.StatementTransactions[0].TransactionType)
	assert.Equal(t, "20240901012345.000[+8:CST]", ofxFile.BankMessageResponseV1.StatementTransactionResponse.StatementResponse.TransactionList.StatementTransactions[0].PostedDate)
	assert.Equal(t, "123.45", ofxFile.BankMessageResponseV1.StatementTransactionResponse.StatementResponse.TransactionList.StatementTransactions[0].Amount)
}

func TestCreateNewOFXFileReader_OFX1WithoutCharset(t *testing.T) {
	context := core.NewNullContext()

	reader, err := createNewOFXFileReader(context, []byte(
		"OFXHEADER:100\n"+
			"DATA:OFXSGML\n"+
			"VERSION:103\n"+
			"SECURITY:NONE\n"+
			"ENCODING:USASCII\n"+
			"CHARSET:\n"+
			"COMPRESSION:NONE\n"+
			"OLDFILEUID:NONE\n"+
			"NEWFILEUID:NONE\n"+
			"FOO:BAR\n"+
			"\n"+
			"<OFX>\n"+
			"</OFX>"))
	assert.Nil(t, err)

	ofxFile, err := reader.read(context)
	assert.NotNil(t, ofxFile)

	assert.NotNil(t, ofxFile.FileHeader)
	assert.Equal(t, ofxVersion1, ofxFile.FileHeader.OFXDeclarationVersion)
	assert.Equal(t, "103", ofxFile.FileHeader.OFXDataVersion)
	assert.Equal(t, "NONE", ofxFile.FileHeader.Security)
	assert.Equal(t, "NONE", ofxFile.FileHeader.OldFileUid)
	assert.Equal(t, "NONE", ofxFile.FileHeader.NewFileUid)
}

func TestCreateNewOFXFileReader_OFX1WithInvalidHeaderVersion(t *testing.T) {
	context := core.NewNullContext()
	_, err := createNewOFXFileReader(context, []byte(
		"OFXHEADER:200\n"+
			"DATA:OFXSGML\n"+
			"VERSION:103\n"+
			"SECURITY:NONE\n"+
			"ENCODING:USASCII\n"+
			"CHARSET:1252\n"+
			"COMPRESSION:NONE\n"+
			"OLDFILEUID:NONE\n"+
			"NEWFILEUID:NONE\n"+
			"\n"+
			"<OFX>\n"+
			"</OFX>"))

	assert.EqualError(t, err, errs.ErrInvalidOFXFile.Message)
}

func TestCreateNewOFXFileReader_OFX1WithInvalidHeader(t *testing.T) {
	context := core.NewNullContext()
	_, err := createNewOFXFileReader(context, []byte(
		"OFXHEADER:100\n"+
			"DATA:XML\n"+
			"VERSION:103\n"+
			"SECURITY:NONE\n"+
			"ENCODING:USASCII\n"+
			"CHARSET:1252\n"+
			"COMPRESSION:NONE\n"+
			"OLDFILEUID:NONE\n"+
			"NEWFILEUID:NONE\n"+
			"\n"+
			"<OFX>\n"+
			"</OFX>"))

	assert.EqualError(t, err, errs.ErrInvalidOFXFile.Message)
}

func TestCreateNewOFXFileReader_OFX1WithUnknownHeader(t *testing.T) {
	context := core.NewNullContext()

	reader, err := createNewOFXFileReader(context, []byte(
		"OFXHEADER:100\n"+
			"DATA:OFXSGML\n"+
			"VERSION:103\n"+
			"SECURITY:NONE\n"+
			"ENCODING:USASCII\n"+
			"CHARSET:1252\n"+
			"COMPRESSION:NONE\n"+
			"OLDFILEUID:NONE\n"+
			"NEWFILEUID:NONE\n"+
			"FOO:BAR\n"+
			"\n"+
			"<OFX>\n"+
			"</OFX>"))
	assert.Nil(t, err)

	ofxFile, err := reader.read(context)
	assert.NotNil(t, ofxFile)

	assert.NotNil(t, ofxFile.FileHeader)
	assert.Equal(t, ofxVersion1, ofxFile.FileHeader.OFXDeclarationVersion)
	assert.Equal(t, "103", ofxFile.FileHeader.OFXDataVersion)
	assert.Equal(t, "NONE", ofxFile.FileHeader.Security)
	assert.Equal(t, "NONE", ofxFile.FileHeader.OldFileUid)
	assert.Equal(t, "NONE", ofxFile.FileHeader.NewFileUid)
}
