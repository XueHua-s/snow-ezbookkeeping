package models

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"math/big"
)

const MinimumTransactionAmount = -999999999999999

const MaximumTransactionAmount = 999999999999999

const MaximumTagsCountOfTransaction = 10

const MaximumPicturesCountOfTransaction = 10

// TransactionType represents transaction type
type TransactionType byte

// Transaction types
const (
	TRANSACTION_TYPE_MODIFY_BALANCE TransactionType = 1
	TRANSACTION_TYPE_INCOME         TransactionType = 2
	TRANSACTION_TYPE_EXPENSE        TransactionType = 3
	TRANSACTION_TYPE_TRANSFER       TransactionType = 4
)

// ToTransactionDbType returns the transaction db type for this enum
func (t TransactionType) ToTransactionDbType() (TransactionDbType, error) {
	if t == TRANSACTION_TYPE_MODIFY_BALANCE {
		return TRANSACTION_DB_TYPE_MODIFY_BALANCE, nil
	} else if t == TRANSACTION_TYPE_EXPENSE {
		return TRANSACTION_DB_TYPE_EXPENSE, nil
	} else if t == TRANSACTION_TYPE_INCOME {
		return TRANSACTION_DB_TYPE_INCOME, nil
	} else if t == TRANSACTION_TYPE_TRANSFER {
		return TRANSACTION_DB_TYPE_TRANSFER_OUT, nil
	} else {
		return 0, errs.ErrTransactionTypeInvalid
	}
}

// TransactionRelatedAccountType represents related account type in transaction
type TransactionRelatedAccountType byte

// Transaction relation types
const (
	TRANSACTION_RELATED_ACCOUNT_TYPE_TRANSFER_FROM TransactionRelatedAccountType = 1
	TRANSACTION_RELATED_ACCOUNT_TYPE_TRANSFER_TO   TransactionRelatedAccountType = 2
)

// TransactionDbType represents transaction type in database
type TransactionDbType byte

// Transaction db types
const (
	TRANSACTION_DB_TYPE_MODIFY_BALANCE TransactionDbType = 1
	TRANSACTION_DB_TYPE_INCOME         TransactionDbType = 2
	TRANSACTION_DB_TYPE_EXPENSE        TransactionDbType = 3
	TRANSACTION_DB_TYPE_TRANSFER_OUT   TransactionDbType = 4
	TRANSACTION_DB_TYPE_TRANSFER_IN    TransactionDbType = 5
)

// String returns a textual representation of the transaction types for db enum
func (t TransactionDbType) String() string {
	switch t {
	case TRANSACTION_DB_TYPE_MODIFY_BALANCE:
		return "Modify Balance"
	case TRANSACTION_DB_TYPE_INCOME:
		return "Income"
	case TRANSACTION_DB_TYPE_EXPENSE:
		return "Expense"
	case TRANSACTION_DB_TYPE_TRANSFER_OUT:
		return "Transfer Out"
	case TRANSACTION_DB_TYPE_TRANSFER_IN:
		return "Transfer In"
	default:
		return fmt.Sprintf("Invalid(%d)", int(t))
	}
}

// ToTransactionType returns the transaction type for this db enum
func (t TransactionDbType) ToTransactionType() (TransactionType, error) {
	if t == TRANSACTION_DB_TYPE_MODIFY_BALANCE {
		return TRANSACTION_TYPE_MODIFY_BALANCE, nil
	} else if t == TRANSACTION_DB_TYPE_EXPENSE {
		return TRANSACTION_TYPE_EXPENSE, nil
	} else if t == TRANSACTION_DB_TYPE_INCOME {
		return TRANSACTION_TYPE_INCOME, nil
	} else if t == TRANSACTION_DB_TYPE_TRANSFER_OUT {
		return TRANSACTION_TYPE_TRANSFER, nil
	} else if t == TRANSACTION_DB_TYPE_TRANSFER_IN {
		return TRANSACTION_TYPE_TRANSFER, nil
	} else {
		return 0, errs.ErrTransactionTypeInvalid
	}
}

// ToTransactionRelatedAccountType returns the related account type for this db enum
func (t TransactionDbType) ToTransactionRelatedAccountType() (TransactionRelatedAccountType, error) {
	if t == TRANSACTION_DB_TYPE_TRANSFER_OUT {
		return TRANSACTION_RELATED_ACCOUNT_TYPE_TRANSFER_TO, nil
	} else if t == TRANSACTION_DB_TYPE_TRANSFER_IN {
		return TRANSACTION_RELATED_ACCOUNT_TYPE_TRANSFER_FROM, nil
	} else {
		return 0, errs.ErrTransactionTypeInvalid
	}
}

// TransactionTagFilterValue represents transaction tag filter value for no tag
const TransactionNoTagFilterValue = "none"

// TransactionTagFilterType represents transaction tag filter type
type TransactionTagFilterType byte

// Transaction tag filter types
const (
	TRANSACTION_TAG_FILTER_HAS_ANY     TransactionTagFilterType = 0
	TRANSACTION_TAG_FILTER_HAS_ALL     TransactionTagFilterType = 1
	TRANSACTION_TAG_FILTER_NOT_HAS_ANY TransactionTagFilterType = 2
	TRANSACTION_TAG_FILTER_NOT_HAS_ALL TransactionTagFilterType = 3
)

// Transaction represents transaction data stored in database
type Transaction struct {
	TransactionId        int64             `xorm:"PK"`
	Uid                  int64             `xorm:"UNIQUE(UQE_transaction_uid_time) INDEX(IDX_transaction_uid_deleted_time) INDEX(IDX_transaction_uid_deleted_type_time) INDEX(IDX_transaction_uid_deleted_type_account_id_time) INDEX(IDX_transaction_uid_deleted_category_id_time) INDEX(IDX_transaction_uid_deleted_account_id_time) INDEX(IDX_transaction_uid_deleted_time_longitude_latitude) NOT NULL"`
	Deleted              bool              `xorm:"INDEX(IDX_transaction_uid_deleted_time) INDEX(IDX_transaction_uid_deleted_type_time) INDEX(IDX_transaction_uid_deleted_type_account_id_time) INDEX(IDX_transaction_uid_deleted_category_id_time) INDEX(IDX_transaction_uid_deleted_account_id_time) INDEX(IDX_transaction_uid_deleted_time_longitude_latitude) NOT NULL"`
	Type                 TransactionDbType `xorm:"INDEX(IDX_transaction_uid_deleted_type_time) INDEX(IDX_transaction_uid_deleted_type_account_id_time) NOT NULL"`
	CategoryId           int64             `xorm:"INDEX(IDX_transaction_uid_deleted_category_id_time) NOT NULL"`
	AccountId            int64             `xorm:"INDEX(IDX_transaction_uid_deleted_account_id_time) INDEX(IDX_transaction_uid_deleted_type_account_id_time) NOT NULL"`
	TransactionTime      int64             `xorm:"UNIQUE(UQE_transaction_uid_time) INDEX(IDX_transaction_uid_deleted_time) INDEX(IDX_transaction_uid_deleted_type_time) INDEX(IDX_transaction_uid_deleted_type_account_id_time) INDEX(IDX_transaction_uid_deleted_category_id_time) INDEX(IDX_transaction_uid_deleted_account_id_time) NOT NULL"`
	TimezoneUtcOffset    int16             `xorm:"NOT NULL"`
	Amount               int64             `xorm:"NOT NULL"`
	RelatedId            int64             `xorm:"NOT NULL"`
	RelatedAccountId     int64             `xorm:"NOT NULL"`
	RelatedAccountAmount int64             `xorm:"NOT NULL"`
	HideAmount           bool              `xorm:"NOT NULL"`
	Comment              string            `xorm:"VARCHAR(255) NOT NULL"`
	GeoLongitude         float64           `xorm:"INDEX(IDX_transaction_uid_deleted_time_longitude_latitude)"`
	GeoLatitude          float64           `xorm:"INDEX(IDX_transaction_uid_deleted_time_longitude_latitude)"`
	CreatedIp            string            `xorm:"VARCHAR(39)"`
	ScheduledCreated     bool
	CreatedUnixTime      int64
	UpdatedUnixTime      int64
	DeletedUnixTime      int64
}

// TransactionWithAccountBalance represents a transaction item with account balance
type TransactionWithAccountBalance struct {
	*Transaction
	AccountOpeningBalance *big.Int
	AccountClosingBalance *big.Int
}
