package services

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// Metadata is loaded once inside each tool's read view. Transaction records,
// associations and pictures remain bounded batches regardless of ledger age.
type assistantMetadata struct {
	Accounts      map[int64]*models.Account
	Categories    map[int64]*models.TransactionCategory
	Tags          map[int64]*models.TransactionTag
	TagGroups     map[int64]*models.TransactionTagGroup
	DeletedTagIDs map[int64]bool
}

func assistantNotAccessible() error {
	return &data.ToolError{Code: data.ToolErrorNotFoundOrNotAccessible, Message: "An entity was not found or is not accessible"}
}
func assistantInconsistent() error {
	return &data.ToolError{Code: data.ToolErrorDataInconsistent, Message: "The ledger contains a missing or inconsistent relation"}
}
func (e *assistantQueryExecution) table(bean any) string { return e.View.Dialect().TableName(bean) }
func (e *assistantQueryExecution) columns(names string) string {
	parts := strings.Split(names, ",")
	for i, p := range parts {
		parts[i] = e.View.Dialect().QuoteIdentifier(strings.TrimSpace(p))
	}
	return strings.Join(parts, ",")
}
func (e *assistantQueryExecution) loadMetadata() error {
	m := &assistantMetadata{Accounts: map[int64]*models.Account{}, Categories: map[int64]*models.TransactionCategory{}, Tags: map[int64]*models.TransactionTag{}, TagGroups: map[int64]*models.TransactionTagGroup{}}
	queries := []struct {
		bean    any
		columns string
		scan    func(interface{ Scan(...any) error }) error
	}{
		{&models.Account{}, "account_id,parent_account_id,name,type,category,currency,balance,hidden", func(r interface{ Scan(...any) error }) error {
			a := &models.Account{Uid: e.Service.uid}
			if err := r.Scan(&a.AccountId, &a.ParentAccountId, &a.Name, &a.Type, &a.Category, &a.Currency, &a.Balance, &a.Hidden); err != nil {
				return err
			}
			m.Accounts[a.AccountId] = a
			return nil
		}},
		{&models.TransactionCategory{}, "category_id,parent_category_id,name,type,hidden", func(r interface{ Scan(...any) error }) error {
			a := &models.TransactionCategory{Uid: e.Service.uid}
			if err := r.Scan(&a.CategoryId, &a.ParentCategoryId, &a.Name, &a.Type, &a.Hidden); err != nil {
				return err
			}
			m.Categories[a.CategoryId] = a
			return nil
		}},
		{&models.TransactionTag{}, "tag_id,tag_group_id,name,hidden", func(r interface{ Scan(...any) error }) error {
			a := &models.TransactionTag{Uid: e.Service.uid}
			if err := r.Scan(&a.TagId, &a.TagGroupId, &a.Name, &a.Hidden); err != nil {
				return err
			}
			m.Tags[a.TagId] = a
			return nil
		}},
		{&models.TransactionTagGroup{}, "tag_group_id,name", func(r interface{ Scan(...any) error }) error {
			a := &models.TransactionTagGroup{Uid: e.Service.uid}
			if err := r.Scan(&a.TagGroupId, &a.Name); err != nil {
				return err
			}
			m.TagGroups[a.TagGroupId] = a
			return nil
		}},
	}
	for _, q := range queries {
		rows, err := e.View.QueryContext(e.Context, "SELECT "+e.columns(q.columns)+" FROM "+e.table(q.bean)+" WHERE uid = ? AND deleted = ?", e.Service.uid, false)
		if err != nil {
			return err
		}
		for rows.Next() {
			if err = q.scan(rows); err != nil {
				rows.Close()
				return err
			}
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
	}
	// A deleted tag is legitimately absent from the effective ledger. An index
	// pointing to a missing/foreign tag is corruption, and must not turn a tagged
	// transaction into an apparently untagged one.
	m.DeletedTagIDs = map[int64]bool{}
	deletedTags, err := e.View.QueryContext(e.Context, "SELECT tag_id FROM "+e.table(&models.TransactionTag{})+" WHERE uid = ? AND deleted = ?", e.Service.uid, true)
	if err != nil {
		return err
	}
	for deletedTags.Next() {
		var id int64
		if err = deletedTags.Scan(&id); err != nil {
			deletedTags.Close()
			return err
		}
		m.DeletedTagIDs[id] = true
	}
	err = deletedTags.Err()
	deletedTags.Close()
	if err != nil {
		return err
	}
	e.Metadata = m
	return nil
}

// Expand keeps original parent choices in Args (and therefore the signature),
// while exposing a sorted, deduplicated effective set for scope and matching.
func assistantExpandIDs(ids []string, parents map[int64]int64) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	children := map[int64][]int64{}
	for id, parent := range parents {
		children[parent] = append(children[parent], id)
	}
	seen := map[int64]bool{}
	visiting := map[int64]bool{}
	var visit func(int64) error
	visit = func(id int64) error {
		if visiting[id] {
			return assistantInconsistent()
		}
		if seen[id] {
			return nil
		}
		visiting[id] = true
		seen[id] = true
		for _, child := range children[id] {
			if err := visit(child); err != nil {
				return err
			}
		}
		delete(visiting, id)
		return nil
	}
	for _, s := range ids {
		id, _ := assistantID(s)
		if _, ok := parents[id]; !ok {
			return nil, assistantNotAccessible()
		}
		if err := visit(id); err != nil {
			return nil, err
		}
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, strconv.FormatInt(id, 10))
	}
	return canonicalStrings(out), nil
}
func (e *assistantQueryExecution) prepare() error {
	if err := e.loadMetadata(); err != nil {
		return err
	}
	ap := map[int64]int64{}
	cp := map[int64]int64{}
	for id, a := range e.Metadata.Accounts {
		ap[id] = a.ParentAccountId
	}
	for id, c := range e.Metadata.Categories {
		cp[id] = c.ParentCategoryId
	}
	selected := e.Args.Filter.AccountIDs
	if e.Tool == data.ToolGetAccountBalances {
		selected = e.Args.AccountIDs
	}
	var err error
	e.AccountIDs, err = assistantExpandIDs(selected, ap)
	if err != nil {
		return err
	}
	e.CategoryIDs, err = assistantExpandIDs(e.Args.Filter.CategoryIDs, cp)
	if err != nil {
		return err
	}
	for _, s := range e.Args.Filter.TagIDs {
		id, _ := assistantID(s)
		if e.Metadata.Tags[id] == nil {
			return assistantNotAccessible()
		}
	}
	// Check explicit selections first, so inaccessible requested IDs always
	// receive the same non-disclosing error even if the ledger is also corrupt.
	for _, a := range e.Metadata.Accounts {
		if a.ParentAccountId != 0 && e.Metadata.Accounts[a.ParentAccountId] == nil {
			return assistantInconsistent()
		}
	}
	for _, c := range e.Metadata.Categories {
		if c.ParentCategoryId != 0 && e.Metadata.Categories[c.ParentCategoryId] == nil {
			return assistantInconsistent()
		}
	}
	for _, tag := range e.Metadata.Tags {
		if tag.TagGroupId != 0 && e.Metadata.TagGroups[tag.TagGroupId] == nil {
			return assistantInconsistent()
		}
	}
	e.HierarchyDigest = assistantHierarchyDigest(e.AccountIDs, e.CategoryIDs)
	if e.Cursor != nil && e.Cursor.HierarchyDigest != e.HierarchyDigest {
		return assistantCursorError()
	}
	f := e.Args.Filter
	e.Scope = &data.QueryScope{StartTime: f.StartTime, EndTime: f.EndTime, Timezone: e.Service.location.String(), Currency: f.Currency, Types: f.Types, AccountIDs: e.AccountIDs, CategoryIDs: e.CategoryIDs, TagIDs: f.TagIDs, TagMatch: f.TagMatch, NoTags: f.NoTags, TransferSide: f.TransferSide, AmountMin: f.AmountMin, AmountMax: f.AmountMax, Keyword: f.Keyword, IncludesHidden: true, ExcludesDeleted: true}
	if e.Tool == data.ToolListAccounts || e.Tool == data.ToolListCategories || e.Tool == data.ToolListTags {
		e.Scope.Keyword = e.Args.Keyword
		e.Scope.Types = nil
		if e.Tool == data.ToolListCategories && e.Args.Type != "" {
			e.Scope.Types = []string{e.Args.Type}
		}
	}
	for _, p := range e.Args.Periods {
		e.Scope.Periods = append(e.Scope.Periods, data.QueryPeriod{Name: p.Name, StartTime: p.StartTime, EndTime: p.EndTime})
	}
	return nil
}

type assistantTransactionSide struct {
	Account *models.Account
	Amount  int64
	Flow    string
	Side    string
}
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

// ScanTransactions reads the entire effective ledger in bounded storage batches.
// It validates incoming rows too, so an orphan or cross-UID source cannot vanish.
// The callback is invoked once per matching canonical transaction and must not
// retain every transaction when computing full-history aggregates.
func (e *assistantQueryExecution) ScanTransactions(high *assistantPageKey, visit func(*assistantLogicalTransaction) error) error {
	var last int64
	for {
		if err := e.Context.Err(); err != nil {
			return err
		}
		batch, err := e.readTransactionBatch("transaction_id > ? ORDER BY transaction_id ASC LIMIT ?", last, assistantTransactionBatchSize)
		if err != nil {
			return err
		}
		if len(batch) == 0 {
			return nil
		}
		logical, err := e.normalizeTransactions(batch)
		if err != nil {
			return err
		}
		canonicalInBatch := map[int64]bool{}
		for _, r := range batch {
			if r.Type != models.TRANSACTION_DB_TYPE_TRANSFER_IN {
				canonicalInBatch[r.TransactionId] = true
			}
		}
		for _, t := range logical {
			if !canonicalInBatch[t.Record.TransactionId] {
				continue
			}
			if high != nil && assistantCompareTimeKey(t.Key(), *high) > 0 {
				continue
			}
			if e.Match(t) {
				if err = visit(t); err != nil {
					return err
				}
			}
		}
		last = batch[len(batch)-1].TransactionId
		if len(batch) < assistantTransactionBatchSize {
			return nil
		}
	}
}
func assistantCompareTimeKey(a, b assistantPageKey) int {
	at, _ := strconv.ParseInt(a.Time, 10, 64)
	bt, _ := strconv.ParseInt(b.Time, 10, 64)
	if at < bt {
		return -1
	}
	if at > bt {
		return 1
	}
	ai, _ := strconv.ParseInt(a.ID, 10, 64)
	bi, _ := strconv.ParseInt(b.ID, 10, 64)
	if ai < bi {
		return -1
	}
	if ai > bi {
		return 1
	}
	return 0
}
func (e *assistantQueryExecution) HighWater() (assistantPageKey, error) {
	batch, err := e.readTransactionBatch("type <> ? ORDER BY transaction_time DESC, transaction_id DESC LIMIT 1", models.TRANSACTION_DB_TYPE_TRANSFER_IN)
	if err != nil {
		return assistantPageKey{}, err
	}
	if len(batch) == 0 {
		return assistantPageKey{}, nil
	}
	r := batch[0]
	return assistantPageKey{Time: strconv.FormatInt(r.TransactionTime, 10), ID: strconv.FormatInt(r.TransactionId, 10)}, nil
}
func (e *assistantQueryExecution) ReadTransaction(id string) (*assistantLogicalTransaction, error) {
	parsed, err := assistantID(id)
	if err != nil {
		return nil, err
	}
	batch, err := e.readTransactionBatch("transaction_id = ?", parsed)
	if err != nil {
		return nil, err
	}
	if len(batch) == 0 {
		return nil, assistantNotAccessible()
	}
	logical, err := e.normalizeTransactions(batch)
	if err != nil {
		return nil, err
	}
	if len(logical) != 1 {
		return nil, assistantInconsistent()
	}
	e.Match(logical[0])
	return logical[0], nil
}

// PageFits includes envelope/scope/cursor overhead, not just the record slice.
// Tool handlers use this before accepting a record and issue a cursor at the
// last accepted key. The execution wrapper independently checks the final size.
func (e *assistantQueryExecution) PageFits(result data.ToolResult) bool {
	result.Scope = e.Scope
	result.QueriedAt = e.Service.now().UTC().Format(time.RFC3339)
	result.QueryID = strings.Repeat("0", 32)
	if result.Warnings == nil {
		result.Warnings = []string{}
	}
	b, err := json.Marshal(result)
	return err == nil && len(b) <= e.Service.maxResultBytes
}
func assistantCount(n int64) string { return strconv.FormatInt(n, 10) }
func assistantJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("assistant projection is not JSON: %v", err))
	}
	return b
}
