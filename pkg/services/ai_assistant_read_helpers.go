package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"strconv"
	"strings"
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
