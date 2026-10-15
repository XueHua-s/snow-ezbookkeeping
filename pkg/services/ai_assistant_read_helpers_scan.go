package services

import (
	"encoding/json"
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"strconv"
	"strings"
	"time"
)

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
	return err == nil && len(b) <= e.Service.maxResultBytes && (e.Service.resultFits == nil || e.Service.resultFits(result))
}

func assistantCount(n int64) string { return strconv.FormatInt(n, 10) }

func assistantJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("assistant projection is not JSON: %v", err))
	}
	return b
}
