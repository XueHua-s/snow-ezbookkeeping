package camt

import (
	"github.com/mayswind/ezbookkeeping/pkg/converters/converter"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCamt053TransactionDataFileParseImportedData_ParseTransactionValidAmountAndCurrency(t *testing.T) {
	importer := Camt053TransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		`<?xml version="1.0" encoding="UTF-8"?>
		<Document xmlns="urn:iso:std:iso:20022:tech:xsd:camt.053.001.02">
			<BkToCstmrStmt>
				<Stmt>
					<Acct>
						<Id>
							<IBAN>123</IBAN>
						</Id>
						<Ccy>CNY</Ccy>
					</Acct>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T12:34:56+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="USD">123.45</Amt>
						<NtryDtls>
							<TxDtls>
								<AmtDtls>
									<TxAmt>
										<Amt Ccy="USD">100.23</Amt>
									</TxAmt>
								</AmtDtls>
							</TxDtls>
							<TxDtls>
								<AmtDtls>
									<TxAmt>
										<Amt Ccy="USD">23.22</Amt>
									</TxAmt>
								</AmtDtls>
							</TxDtls>
						</NtryDtls>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(allNewTransactions))
	assert.Equal(t, "USD", allNewTransactions[0].OriginalSourceAccountCurrency)
	assert.Equal(t, int64(2322), allNewTransactions[0].Amount)
	assert.Equal(t, "USD", allNewTransactions[1].OriginalSourceAccountCurrency)
	assert.Equal(t, int64(10023), allNewTransactions[1].Amount)

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		`<?xml version="1.0" encoding="UTF-8"?>
		<Document xmlns="urn:iso:std:iso:20022:tech:xsd:camt.053.001.02">
			<BkToCstmrStmt>
				<Stmt>
					<Acct>
						<Id>
							<IBAN>123</IBAN>
						</Id>
						<Ccy>CNY</Ccy>
					</Acct>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T12:34:56+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="USD">123.45</Amt>
						<NtryDtls>
							<TxDtls>
								<AmtDtls>
									<InstdAmt>
										<Amt Ccy="USD">99.99</Amt>
									</InstdAmt>
									<TxAmt>
										<Amt Ccy="USD">100.23</Amt>
									</TxAmt>
								</AmtDtls>
							</TxDtls>
							<TxDtls>
								<AmtDtls>
									<InstdAmt>
										<Amt Ccy="USD">23.46</Amt>
									</InstdAmt>
									<TxAmt>
										<Amt Ccy="USD">23.22</Amt>
									</TxAmt>
								</AmtDtls>
							</TxDtls>
						</NtryDtls>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(allNewTransactions))
	assert.Equal(t, "USD", allNewTransactions[0].OriginalSourceAccountCurrency)
	assert.Equal(t, int64(2346), allNewTransactions[0].Amount)
	assert.Equal(t, "USD", allNewTransactions[1].OriginalSourceAccountCurrency)
	assert.Equal(t, int64(9999), allNewTransactions[1].Amount)

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		`<?xml version="1.0" encoding="UTF-8"?>
		<Document xmlns="urn:iso:std:iso:20022:tech:xsd:camt.053.001.02">
			<BkToCstmrStmt>
				<Stmt>
					<Acct>
						<Id>
							<IBAN>123</IBAN>
						</Id>
						<Ccy>CNY</Ccy>
					</Acct>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T12:34:56+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="USD">123.45</Amt>
						<NtryDtls>
							<TxDtls>
								<AmtDtls>
									<TxAmt>
										<Amt Ccy="USD">123.45</Amt>
									</TxAmt>
								</AmtDtls>
							</TxDtls>
						</NtryDtls>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, "USD", allNewTransactions[0].OriginalSourceAccountCurrency)
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		`<?xml version="1.0" encoding="UTF-8"?>
		<Document xmlns="urn:iso:std:iso:20022:tech:xsd:camt.053.001.02">
			<BkToCstmrStmt>
				<Stmt>
					<Acct>
						<Id>
							<IBAN>123</IBAN>
						</Id>
						<Ccy>CNY</Ccy>
					</Acct>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T12:34:56+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt>123.45</Amt>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, "CNY", allNewTransactions[0].OriginalSourceAccountCurrency)
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)
}

func TestCamt053TransactionDataFileParseImportedData_ParseTransactionInvalidAmountAndCurrency(t *testing.T) {
	importer := Camt053TransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	_, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		`<?xml version="1.0" encoding="UTF-8"?>
		<Document xmlns="urn:iso:std:iso:20022:tech:xsd:camt.053.001.02">
			<BkToCstmrStmt>
				<Stmt>
					<Acct>
						<Id>
							<IBAN>123</IBAN>
						</Id>
						<Ccy>CNY</Ccy>
					</Acct>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T12:34:56+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt>123 45</Amt>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		`<?xml version="1.0" encoding="UTF-8"?>
		<Document xmlns="urn:iso:std:iso:20022:tech:xsd:camt.053.001.02">
			<BkToCstmrStmt>
				<Stmt>
					<Acct>
						<Id>
							<IBAN>123</IBAN>
						</Id>
						<Ccy>CNY</Ccy>
					</Acct>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T12:34:56+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="USD">123.45</Amt>
						<NtryDtls>
							<TxDtls>
								<AmtDtls>
								</AmtDtls>
							</TxDtls>
							<TxDtls>
							</TxDtls>
						</NtryDtls>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)

	_, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		`<?xml version="1.0" encoding="UTF-8"?>
		<Document xmlns="urn:iso:std:iso:20022:tech:xsd:camt.053.001.02">
			<BkToCstmrStmt>
				<Stmt>
					<Acct>
						<Id>
							<IBAN>123</IBAN>
						</Id>
						<Ccy>CNY</Ccy>
					</Acct>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T12:34:56+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="USD">123.45</Amt>
						<NtryDtls>
							<TxDtls>
							</TxDtls>
							<TxDtls>
								<AmtDtls>
								</AmtDtls>
							</TxDtls>
						</NtryDtls>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrAmountInvalid.Message)
}

func TestCamt053TransactionDataFileParseImportedData_ParseDescription(t *testing.T) {
	importer := Camt053TransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, _, _, _, _, _, err := importer.ParseImportedData(context, user, []byte(
		`<?xml version="1.0" encoding="UTF-8"?>
		<Document xmlns="urn:iso:std:iso:20022:tech:xsd:camt.053.001.02">
			<BkToCstmrStmt>
				<Stmt>
					<Acct>
						<Id>
							<IBAN>123</IBAN>
						</Id>
						<Ccy>CNY</Ccy>
					</Acct>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T12:34:56+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="CNY">123.45</Amt>
						<AddtlNtryInf>Test Entry</AddtlNtryInf>
						<NtryDtls>
							<TxDtls>
								<AddtlTxInf>Test Transaction</AddtlTxInf>
								<RmtInf>
									<Ustrd>Test Line 1</Ustrd>
									<Ustrd>Test Line 2</Ustrd>
								</RmtInf>
							</TxDtls>
						</NtryDtls>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, "Test Transaction", allNewTransactions[0].Comment)

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		`<?xml version="1.0" encoding="UTF-8"?>
		<Document xmlns="urn:iso:std:iso:20022:tech:xsd:camt.053.001.02">
			<BkToCstmrStmt>
				<Stmt>
					<Acct>
						<Id>
							<IBAN>123</IBAN>
						</Id>
						<Ccy>CNY</Ccy>
					</Acct>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T12:34:56+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="CNY">123.45</Amt>
						<AddtlNtryInf>Test Entry</AddtlNtryInf>
						<NtryDtls>
							<TxDtls>
								<RmtInf>
									<Ustrd>Test Line 1</Ustrd>
									<Ustrd>Test Line 2</Ustrd>
								</RmtInf>
							</TxDtls>
						</NtryDtls>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, "Test Line 1\nTest Line 2", allNewTransactions[0].Comment)

	allNewTransactions, _, _, _, _, _, err = importer.ParseImportedData(context, user, []byte(
		`<?xml version="1.0" encoding="UTF-8"?>
		<Document xmlns="urn:iso:std:iso:20022:tech:xsd:camt.053.001.02">
			<BkToCstmrStmt>
				<Stmt>
					<Acct>
						<Id>
							<IBAN>123</IBAN>
						</Id>
						<Ccy>CNY</Ccy>
					</Acct>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T12:34:56+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="CNY">123.45</Amt>
						<AddtlNtryInf>Test Entry</AddtlNtryInf>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(allNewTransactions))
	assert.Equal(t, "Test Entry", allNewTransactions[0].Comment)
}
