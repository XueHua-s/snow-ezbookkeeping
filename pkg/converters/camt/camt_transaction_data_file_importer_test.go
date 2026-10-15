package camt

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

func TestCamt052TransactionDataFileParseImportedData_MinimumValidData(t *testing.T) {
	importer := Camt052TransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, allNewAccounts, allNewSubExpenseCategories, allNewSubIncomeCategories, allNewSubTransferCategories, allNewTags, err := importer.ParseImportedData(context, user, []byte(
		`<?xml version="1.0" encoding="UTF-8"?>
		<Document xmlns="urn:iso:std:iso:20022:tech:xsd:camt.052.001.02">
			<BkToCstmrAcctRpt>
				<Rpt>
					<Acct>
						<Id>
							<IBAN>123</IBAN>
						</Id>
						<Ccy>CNY</Ccy>
					</Acct>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T01:23:45+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="CNY">123.45</Amt>
					</Ntry>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T12:34:56+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>DBIT</CdtDbtInd>
						<Amt Ccy="CNY">0.12</Amt>
					</Ntry>
				</Rpt>
				<Rpt>
					<Acct>
						<Id>
							<Othr>
								<Id>456</Id>
							</Othr>
						</Id>
						<Ccy>USD</Ccy>
					</Acct>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T23:59:59+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="USD">1.23</Amt>
					</Ntry>
				</Rpt>
			</BkToCstmrAcctRpt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 3, len(allNewTransactions))
	assert.Equal(t, 2, len(allNewAccounts))
	assert.Equal(t, 1, len(allNewSubExpenseCategories))
	assert.Equal(t, 1, len(allNewSubIncomeCategories))
	assert.Equal(t, 0, len(allNewSubTransferCategories))
	assert.Equal(t, 0, len(allNewTags))

	assert.Equal(t, int64(1234567890), allNewTransactions[0].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_INCOME, allNewTransactions[0].Type)
	assert.Equal(t, int64(1725125025), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)
	assert.Equal(t, "123", allNewTransactions[0].OriginalSourceAccountName)
	assert.Equal(t, "CNY", allNewTransactions[0].OriginalSourceAccountCurrency)
	assert.Equal(t, "", allNewTransactions[0].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[1].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_EXPENSE, allNewTransactions[1].Type)
	assert.Equal(t, int64(1725165296), utils.GetUnixTimeFromTransactionTime(allNewTransactions[1].TransactionTime))
	assert.Equal(t, int64(12), allNewTransactions[1].Amount)
	assert.Equal(t, "123", allNewTransactions[1].OriginalSourceAccountName)
	assert.Equal(t, "CNY", allNewTransactions[1].OriginalSourceAccountCurrency)
	assert.Equal(t, "", allNewTransactions[1].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[2].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_INCOME, allNewTransactions[2].Type)
	assert.Equal(t, int64(1725206399), utils.GetUnixTimeFromTransactionTime(allNewTransactions[2].TransactionTime))
	assert.Equal(t, int64(123), allNewTransactions[2].Amount)
	assert.Equal(t, "456", allNewTransactions[2].OriginalSourceAccountName)
	assert.Equal(t, "USD", allNewTransactions[2].OriginalSourceAccountCurrency)
	assert.Equal(t, "", allNewTransactions[2].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewAccounts[0].Uid)
	assert.Equal(t, "123", allNewAccounts[0].Name)
	assert.Equal(t, "CNY", allNewAccounts[0].Currency)

	assert.Equal(t, int64(1234567890), allNewAccounts[1].Uid)
	assert.Equal(t, "456", allNewAccounts[1].Name)
	assert.Equal(t, "USD", allNewAccounts[1].Currency)

	assert.Equal(t, int64(1234567890), allNewSubExpenseCategories[0].Uid)
	assert.Equal(t, "", allNewSubExpenseCategories[0].Name)

	assert.Equal(t, int64(1234567890), allNewSubIncomeCategories[0].Uid)
	assert.Equal(t, "", allNewSubIncomeCategories[0].Name)
}

func TestCamt053TransactionDataFileParseImportedData_MinimumValidData(t *testing.T) {
	importer := Camt053TransactionDataImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "CNY",
	}

	allNewTransactions, allNewAccounts, allNewSubExpenseCategories, allNewSubIncomeCategories, allNewSubTransferCategories, allNewTags, err := importer.ParseImportedData(context, user, []byte(
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
							<DtTm>2024-09-01T01:23:45+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="CNY">123.45</Amt>
					</Ntry>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T12:34:56+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>DBIT</CdtDbtInd>
						<Amt Ccy="CNY">0.12</Amt>
					</Ntry>
				</Stmt>
				<Stmt>
					<Acct>
						<Id>
							<Othr>
								<Id>456</Id>
							</Othr>
						</Id>
						<Ccy>USD</Ccy>
					</Acct>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T23:59:59+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="USD">1.23</Amt>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)

	assert.Equal(t, 3, len(allNewTransactions))
	assert.Equal(t, 2, len(allNewAccounts))
	assert.Equal(t, 1, len(allNewSubExpenseCategories))
	assert.Equal(t, 1, len(allNewSubIncomeCategories))
	assert.Equal(t, 0, len(allNewSubTransferCategories))
	assert.Equal(t, 0, len(allNewTags))

	assert.Equal(t, int64(1234567890), allNewTransactions[0].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_INCOME, allNewTransactions[0].Type)
	assert.Equal(t, int64(1725125025), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
	assert.Equal(t, int64(12345), allNewTransactions[0].Amount)
	assert.Equal(t, "123", allNewTransactions[0].OriginalSourceAccountName)
	assert.Equal(t, "CNY", allNewTransactions[0].OriginalSourceAccountCurrency)
	assert.Equal(t, "", allNewTransactions[0].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[1].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_EXPENSE, allNewTransactions[1].Type)
	assert.Equal(t, int64(1725165296), utils.GetUnixTimeFromTransactionTime(allNewTransactions[1].TransactionTime))
	assert.Equal(t, int64(12), allNewTransactions[1].Amount)
	assert.Equal(t, "123", allNewTransactions[1].OriginalSourceAccountName)
	assert.Equal(t, "CNY", allNewTransactions[1].OriginalSourceAccountCurrency)
	assert.Equal(t, "", allNewTransactions[1].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewTransactions[2].Uid)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_INCOME, allNewTransactions[2].Type)
	assert.Equal(t, int64(1725206399), utils.GetUnixTimeFromTransactionTime(allNewTransactions[2].TransactionTime))
	assert.Equal(t, int64(123), allNewTransactions[2].Amount)
	assert.Equal(t, "456", allNewTransactions[2].OriginalSourceAccountName)
	assert.Equal(t, "USD", allNewTransactions[2].OriginalSourceAccountCurrency)
	assert.Equal(t, "", allNewTransactions[2].OriginalCategoryName)

	assert.Equal(t, int64(1234567890), allNewAccounts[0].Uid)
	assert.Equal(t, "123", allNewAccounts[0].Name)
	assert.Equal(t, "CNY", allNewAccounts[0].Currency)

	assert.Equal(t, int64(1234567890), allNewAccounts[1].Uid)
	assert.Equal(t, "456", allNewAccounts[1].Name)
	assert.Equal(t, "USD", allNewAccounts[1].Currency)

	assert.Equal(t, int64(1234567890), allNewSubExpenseCategories[0].Uid)
	assert.Equal(t, "", allNewSubExpenseCategories[0].Name)

	assert.Equal(t, int64(1234567890), allNewSubIncomeCategories[0].Uid)
	assert.Equal(t, "", allNewSubIncomeCategories[0].Name)
}

func TestCamt053TransactionDataFileParseImportedData_ParseValidTransactionTime(t *testing.T) {
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
							<Dt>2024-09-01</Dt>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="CNY">123.45</Amt>
					</Ntry>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-01T12:34:56+08:00</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="CNY">123.45</Amt>
					</Ntry>
					<Ntry>
						<BookgDt>
							<DtTm>2024-09-02T03:04:05Z</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="CNY">123.45</Amt>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 3, len(allNewTransactions))

	assert.Equal(t, int64(1725148800), utils.GetUnixTimeFromTransactionTime(allNewTransactions[0].TransactionTime))
	assert.Equal(t, int64(1725165296), utils.GetUnixTimeFromTransactionTime(allNewTransactions[1].TransactionTime))
	assert.Equal(t, int64(1725246245), utils.GetUnixTimeFromTransactionTime(allNewTransactions[2].TransactionTime))
}

func TestCamt053TransactionDataFileParseImportedData_ParseInvalidTransactionTime(t *testing.T) {
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
							<DtTm>2024</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="CNY">123.45</Amt>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTimeInvalid.Message)

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
							<DtTm>2024T1</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="CNY">123.45</Amt>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTimeInvalid.Message)

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
							<DtTm>2024-09-01 12:34:56</DtTm>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="CNY">123.45</Amt>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTimeInvalid.Message)

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
							<Dt>2024/09/01</Dt>
						</BookgDt>
						<CdtDbtInd>CRDT</CdtDbtInd>
						<Amt Ccy="CNY">123.45</Amt>
					</Ntry>
				</Stmt>
			</BkToCstmrStmt>
		</Document>`), time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.EqualError(t, err, errs.ErrTransactionTimeInvalid.Message)
}
