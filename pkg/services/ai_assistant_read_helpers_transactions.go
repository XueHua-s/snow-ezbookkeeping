package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"
)

type assistantLogicalTransaction struct {
	Record          *models.Transaction // canonical source; never serialized directly
	Source          assistantTransactionSide
	Destination     *assistantTransactionSide
	Category        *models.TransactionCategory
	Tags            []*models.TransactionTag
	AttachmentCount int64
	MatchedSides    []assistantTransactionSide
}

func (t *assistantLogicalTransaction) Type() string {
	switch t.Record.Type {
	case models.TRANSACTION_DB_TYPE_INCOME:
		return "income"
	case models.TRANSACTION_DB_TYPE_EXPENSE:
		return "expense"
	case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE:
		return "balance_modification"
	default:
		return "transfer"
	}
}

func (t *assistantLogicalTransaction) Key() assistantPageKey {
	return assistantPageKey{ID: strconv.FormatInt(t.Record.TransactionId, 10), Time: strconv.FormatInt(t.Record.TransactionTime, 10)}
}

func (t *assistantLogicalTransaction) Reference() data.TransactionReference {
	r := data.TransactionReference{ID: strconv.FormatInt(t.Record.TransactionId, 10), Time: time.Unix(t.Record.TransactionTime/1000, 0).UTC().Format(time.RFC3339), TimezoneOffset: t.Record.TimezoneUtcOffset, Type: t.Type(), SourceAccountID: strconv.FormatInt(t.Source.Account.AccountId, 10), SourceAccountName: t.Source.Account.Name, SourceAmount: assistantFormatAmount(big.NewInt(t.Source.Amount)), Currency: t.Source.Account.Currency, Comment: t.Record.Comment, HideAmount: t.Record.HideAmount}
	if t.Category != nil {
		id := strconv.FormatInt(t.Category.CategoryId, 10)
		r.CategoryID = &id
		r.CategoryName = t.Category.Name
	}
	if t.Destination != nil {
		d := t.Destination
		r.DestinationAccountID = strconv.FormatInt(d.Account.AccountId, 10)
		r.DestinationAccountName = d.Account.Name
		r.DestinationAmount = assistantFormatAmount(big.NewInt(d.Amount))
		r.DestinationCurrency = d.Account.Currency
	}
	return r
}

// Match applies predicates only after transfer normalization. Account, currency
// and amount predicates are conjoined on the SAME side, never independently.
func (e *assistantQueryExecution) Match(t *assistantLogicalTransaction) bool {
	f := e.Args.Filter
	r := t.Record
	if f.StartMillis != nil && r.TransactionTime < *f.StartMillis || f.EndMillis != nil && r.TransactionTime >= *f.EndMillis {
		return false
	}
	if len(f.Types) > 0 && !containsString(f.Types, t.Type()) {
		return false
	}
	if len(e.CategoryIDs) > 0 && !containsString(e.CategoryIDs, strconv.FormatInt(r.CategoryId, 10)) {
		return false
	}
	if !assistantLiteralContains(r.Comment, f.Keyword) {
		return false
	}
	if f.NoTags && len(t.Tags) > 0 {
		return false
	}
	if len(f.TagIDs) > 0 {
		n := 0
		for _, id := range f.TagIDs {
			for _, tag := range t.Tags {
				if strconv.FormatInt(tag.TagId, 10) == id {
					n++
					break
				}
			}
		}
		if f.TagMatch == "all" && n != len(f.TagIDs) || f.TagMatch == "any" && n == 0 {
			return false
		}
	}
	t.MatchedSides = nil
	matchSide := func(s assistantTransactionSide) {
		if len(e.AccountIDs) > 0 && !containsString(e.AccountIDs, strconv.FormatInt(s.Account.AccountId, 10)) {
			return
		}
		if f.Currency != "" && s.Account.Currency != f.Currency {
			return
		}
		if f.MinMinor != nil && s.Amount < *f.MinMinor || f.MaxMinor != nil && s.Amount > *f.MaxMinor {
			return
		}
		t.MatchedSides = append(t.MatchedSides, s)
	}
	if t.Destination == nil {
		matchSide(t.Source)
	} else {
		if f.TransferSide != "destination" {
			matchSide(t.Source)
		}
		if f.TransferSide != "source" {
			matchSide(*t.Destination)
		}
	}
	return len(t.MatchedSides) > 0
}

const assistantTransactionBatchSize = 256

const assistantTransactionColumns = "transaction_id,type,category_id,account_id,transaction_time,timezone_utc_offset,amount,related_id,related_account_id,related_account_amount,hide_amount,comment"

func (e *assistantQueryExecution) readTransactionBatch(where string, args ...any) ([]*models.Transaction, error) {
	params := []any{e.Service.uid, false}
	params = append(params, args...)
	rows, err := e.View.QueryContext(e.Context, "SELECT "+e.columns(assistantTransactionColumns)+" FROM "+e.table(&models.Transaction{})+" WHERE uid = ? AND deleted = ? AND "+where, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*models.Transaction{}
	for rows.Next() {
		r := &models.Transaction{Uid: e.Service.uid}
		if err = rows.Scan(&r.TransactionId, &r.Type, &r.CategoryId, &r.AccountId, &r.TransactionTime, &r.TimezoneUtcOffset, &r.Amount, &r.RelatedId, &r.RelatedAccountId, &r.RelatedAccountAmount, &r.HideAmount, &r.Comment); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func assistantIN(ids []int64) (string, []any) {
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return "(" + strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",") + ")", args
}

func (e *assistantQueryExecution) normalizeTransactions(batch []*models.Transaction) ([]*assistantLogicalTransaction, error) {
	records := map[int64]*models.Transaction{}
	needed := map[int64]bool{}
	for _, r := range batch {
		records[r.TransactionId] = r
	}
	for _, r := range batch {
		if r.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN || r.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			if r.RelatedId <= 0 {
				return nil, assistantInconsistent()
			}
			if records[r.RelatedId] == nil {
				needed[r.RelatedId] = true
			}
		}
	}
	if len(needed) > 0 {
		ids := make([]int64, 0, len(needed))
		for id := range needed {
			ids = append(ids, id)
		}
		in, args := assistantIN(ids)
		related, err := e.readTransactionBatch("transaction_id IN "+in, args...)
		if err != nil {
			return nil, err
		}
		for _, r := range related {
			records[r.TransactionId] = r
		}
	}
	out := make([]*assistantLogicalTransaction, 0, len(batch))
	seen := map[int64]bool{}
	for _, original := range batch {
		r := original
		var destination *models.Transaction
		if r.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			source := records[r.RelatedId]
			if source == nil || source.Type != models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
				return nil, assistantInconsistent()
			}
			r = source
			destination = original
		} else if r.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			destination = records[r.RelatedId]
		}
		if r.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			if destination == nil || destination.Type != models.TRANSACTION_DB_TYPE_TRANSFER_IN || destination.RelatedId != r.TransactionId || r.RelatedId != destination.TransactionId || r.RelatedAccountId != destination.AccountId || destination.RelatedAccountId != r.AccountId || r.RelatedAccountAmount != destination.Amount || destination.RelatedAccountAmount != r.Amount {
				return nil, assistantInconsistent()
			}
		}
		if r.Type < models.TRANSACTION_DB_TYPE_MODIFY_BALANCE || r.Type > models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			return nil, assistantInconsistent()
		}
		account := e.Metadata.Accounts[r.AccountId]
		if account == nil || account.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
			return nil, assistantInconsistent()
		}
		t := &assistantLogicalTransaction{Record: r, Source: assistantTransactionSide{Account: account, Amount: r.Amount, Side: "source"}}
		t.Source.Flow = t.Type()
		if r.CategoryId != 0 {
			t.Category = e.Metadata.Categories[r.CategoryId]
			if t.Category == nil {
				return nil, assistantInconsistent()
			}
		}
		if destination != nil {
			account := e.Metadata.Accounts[destination.AccountId]
			if account == nil || account.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
				return nil, assistantInconsistent()
			}
			t.Source.Flow = "transfer_out"
			t.Destination = &assistantTransactionSide{Account: account, Amount: destination.Amount, Side: "destination", Flow: "transfer_in"}
		}
		if !seen[r.TransactionId] {
			seen[r.TransactionId] = true
			out = append(out, t)
		}
	}
	if err := e.attachTransactionMetadata(out); err != nil {
		return nil, err
	}
	return out, nil
}

func (e *assistantQueryExecution) attachTransactionMetadata(transactions []*assistantLogicalTransaction) error {
	if len(transactions) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(transactions))
	byID := map[int64]*assistantLogicalTransaction{}
	for _, t := range transactions {
		ids = append(ids, t.Record.TransactionId)
		byID[t.Record.TransactionId] = t
	}
	in, args := assistantIN(ids)
	params := []any{e.Service.uid, false}
	params = append(params, args...)
	rows, err := e.View.QueryContext(e.Context, "SELECT transaction_id,tag_id FROM "+e.table(&models.TransactionTagIndex{})+" WHERE uid = ? AND deleted = ? AND transaction_id IN "+in, params...)
	if err != nil {
		return err
	}
	seen := map[[2]int64]bool{}
	for rows.Next() {
		var transactionID, tagID int64
		if err = rows.Scan(&transactionID, &tagID); err != nil {
			rows.Close()
			return err
		}
		tag := e.Metadata.Tags[tagID]
		if tag == nil && !e.Metadata.DeletedTagIDs[tagID] {
			rows.Close()
			return assistantInconsistent()
		}
		key := [2]int64{transactionID, tagID}
		if tag != nil && !seen[key] {
			byID[transactionID].Tags = append(byID[transactionID].Tags, tag)
			seen[key] = true
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	rows, err = e.View.QueryContext(e.Context, "SELECT transaction_id,COUNT(*) FROM "+e.table(&models.TransactionPictureInfo{})+" WHERE uid = ? AND deleted = ? AND transaction_id IN "+in+" GROUP BY transaction_id", params...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, count int64
		if err = rows.Scan(&id, &count); err != nil {
			return err
		}
		byID[id].AttachmentCount = count
	}
	if err = rows.Err(); err != nil {
		return err
	}
	for _, t := range transactions {
		sort.Slice(t.Tags, func(i, j int) bool { return t.Tags[i].TagId < t.Tags[j].TagId })
	}
	return nil
}
