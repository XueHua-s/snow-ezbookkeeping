package services

import (
	"container/heap"
	"sort"
	"strconv"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
)

func newAssistantTransactionTools(_ *AssistantQueryService) map[data.ToolName]assistantToolHandler {
	return map[data.ToolName]assistantToolHandler{data.ToolQueryTransactions: assistantQueryTransactions, data.ToolGetTransaction: assistantGetTransaction}
}

// Only this business projection crosses the tool boundary; ORM records and
// private ordering keys remain inside the read view and signed cursor.
type assistantTransactionProjection struct {
	data.TransactionReference
	Tags            []assistantTransactionTag `json:"tags"`
	AttachmentCount string                    `json:"attachment_count"`
	MatchedSides    []string                  `json:"matched_sides"`
}
type assistantTransactionTag struct {
	ID   string `json:"tag_id"`
	Name string `json:"name"`
}

func assistantProjectTransaction(t *assistantLogicalTransaction) assistantTransactionProjection {
	p := assistantTransactionProjection{TransactionReference: t.Reference(), Tags: []assistantTransactionTag{}, AttachmentCount: assistantCount(t.AttachmentCount), MatchedSides: []string{}}
	if t.Category == nil {
		p.CategoryName = "Uncategorized"
	}
	for _, tag := range t.Tags {
		p.Tags = append(p.Tags, assistantTransactionTag{ID: strconv.FormatInt(tag.TagId, 10), Name: tag.Name})
	}
	for _, side := range t.MatchedSides {
		p.MatchedSides = append(p.MatchedSides, side.Side)
	}
	return p
}
func assistantTransactionResultTooLarge() error {
	return &data.ToolError{Code: data.ToolErrorResultTooLarge, Message: "A complete transaction cannot fit within the result limit"}
}
func assistantGetTransaction(e *assistantQueryExecution) (data.ToolResult, error) {
	t, err := e.ReadTransaction(e.Args.TransactionID)
	if err != nil {
		return data.ToolResult{}, err
	}
	result := data.ToolResult{OK: true, MatchedCount: "1", ReturnedCount: 1, Coverage: data.QueryCoveragePage, Data: assistantJSON(assistantProjectTransaction(t)), References: []data.TransactionReference{t.Reference()}}
	if !e.PageFits(result) {
		return data.ToolResult{}, assistantTransactionResultTooLarge()
	}
	return result, nil
}

type assistantTransactionCandidate struct {
	transaction *assistantLogicalTransaction
	key         assistantPageKey
}

// The root is the LAST candidate in requested order. Thus a whole-history
// scan keeps at most page_size+1 rows, independent of ledger size or page depth.
type assistantTransactionPageHeap struct {
	items []assistantTransactionCandidate
	order assistantOrder
}

func (h assistantTransactionPageHeap) Len() int { return len(h.items) }
func (h assistantTransactionPageHeap) Less(i, j int) bool {
	return assistantCompareTransactionKeys(h.items[i].key, h.items[j].key, h.order) > 0
}
func (h assistantTransactionPageHeap) Swap(i, j int) { h.items[i], h.items[j] = h.items[j], h.items[i] }
func (h *assistantTransactionPageHeap) Push(v any) {
	h.items = append(h.items, v.(assistantTransactionCandidate))
}
func (h *assistantTransactionPageHeap) Pop() any {
	last := len(h.items) - 1
	v := h.items[last]
	h.items[last] = assistantTransactionCandidate{}
	h.items = h.items[:last]
	return v
}
func assistantCompareTransactionKeys(a, b assistantPageKey, order assistantOrder) int {
	comparison := 0
	if order.Field == "amount" {
		av, _ := strconv.ParseInt(a.Amount, 10, 64)
		bv, _ := strconv.ParseInt(b.Amount, 10, 64)
		if av < bv {
			comparison = -1
		} else if av > bv {
			comparison = 1
		}
	}
	if comparison == 0 {
		comparison = assistantCompareTimeKey(a, b)
	}
	if order.Direction == "desc" {
		comparison = -comparison
	}
	return comparison
}
func assistantTransactionOrderKey(t *assistantLogicalTransaction, order assistantOrder) assistantPageKey {
	key := t.Key()
	if order.Field == "amount" {
		// Validation requires a currency and an explicit transfer side, so shared
		// matching has exactly one side for every amount-ordered transaction.
		key.Amount = strconv.FormatInt(t.MatchedSides[0].Amount, 10)
	}
	return key
}
func assistantQueryTransactions(e *assistantQueryExecution) (data.ToolResult, error) {
	page := &assistantTransactionPageHeap{order: e.Args.OrderBy}
	var high assistantPageKey
	var boundary *assistantPageKey
	if e.Cursor != nil {
		high = e.Cursor.HighWater
		boundary = &high
	}
	var matched int64
	err := e.ScanTransactions(boundary, func(t *assistantLogicalTransaction) error {
		matched++
		if e.Cursor == nil && (high.ID == "" || assistantCompareTimeKey(t.Key(), high) > 0) {
			high = t.Key()
		}
		key := assistantTransactionOrderKey(t, e.Args.OrderBy)
		if e.Cursor != nil && assistantCompareTransactionKeys(key, e.Cursor.Last, e.Args.OrderBy) <= 0 {
			return nil
		}
		candidate := assistantTransactionCandidate{transaction: t, key: key}
		if page.Len() < e.Args.PageSize+1 {
			heap.Push(page, candidate)
		} else if assistantCompareTransactionKeys(key, page.items[0].key, e.Args.OrderBy) < 0 {
			page.items[0] = candidate
			heap.Fix(page, 0)
		}
		return nil
	})
	if err != nil {
		return data.ToolResult{}, err
	}
	sort.Slice(page.items, func(i, j int) bool {
		return assistantCompareTransactionKeys(page.items[i].key, page.items[j].key, e.Args.OrderBy) < 0
	})
	count := page.Len()
	if count > e.Args.PageSize {
		count = e.Args.PageSize
	}
	// Test the exact envelope, including the cursor at the actual last emitted
	// row. Removing suffixes preserves every omitted row for the following page.
	for {
		result := data.ToolResult{OK: true, MatchedCount: assistantCount(matched), ReturnedCount: count, HasMore: page.Len() > count, Coverage: data.QueryCoveragePage, Warnings: []string{"Each page reads current data; edits or backfilled transactions during pagination may change results."}}
		records := make([]assistantTransactionProjection, 0, count)
		for _, candidate := range page.items[:count] {
			records = append(records, assistantProjectTransaction(candidate.transaction))
			if len(result.References) < 8 {
				result.References = append(result.References, candidate.transaction.Reference())
			}
		}
		result.Data = assistantJSON(records)
		if result.HasMore {
			result.NextCursor, err = e.NextCursor(high, page.items[count-1].key)
			if err != nil {
				return data.ToolResult{}, err
			}
		}
		if e.PageFits(result) {
			return result, nil
		}
		if count <= 1 {
			return data.ToolResult{}, assistantTransactionResultTooLarge()
		}
		count--
	}
}
