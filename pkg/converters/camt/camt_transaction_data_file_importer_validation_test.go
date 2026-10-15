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

func TestCamt053TransactionDataFileParseImportedData_MissingAccountNode(t *testing.T) {
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
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T12:34:56+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="CNY">123.45</Amt>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingAccountData.Message)
}

func TestCamt053TransactionDataFileParseImportedData_MissingTransactionRequiredNode(t *testing.T) {
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
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="CNY">123.45</Amt>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrMissingTransactionTime.Message)

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
						<Amt Ccy="CNY">123.45</Amt>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTypeInvalid.Message)

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
	assert.EqualError(t, err, errs.ErrAccountCurrencyInvalid.Message)
}
