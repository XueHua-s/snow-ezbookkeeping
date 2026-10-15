package datastore

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"xorm.io/xorm"
)

// AssistantReadView is valid only inside WithAssistantReadView's callback.
// Queries must be trusted service templates; all user values are bound arguments.
// No transaction control or write executor is exposed.
type AssistantReadView interface {
	QueryContext(context.Context, string, ...any) (*AssistantRows, error)
	QueryRowContext(context.Context, string, ...any) *AssistantRow
	Dialect() AssistantSQLDialect
	Observations() AssistantReadObservations
}

// AssistantReadObservations contains no SQL, parameter values, or financial data.
// RowsRead counts rows actually delivered to the application (including aggregate
// rows), NOT the database optimizer's scanned rows or business result records.
// DatabaseDuration measures query/cursor I/O, excluding callback processing and
// transaction setup/cleanup; QueryCount excludes connection-management SQL.
type AssistantReadObservations struct {
	RequestID        string
	QueryCount       int64
	RowsRead         int64
	DatabaseDuration time.Duration
}

// AssistantSQLDialect centralizes identifier quoting, table mapping and binding.
type AssistantSQLDialect struct {
	databaseType string
	engine       *xorm.Engine
}

func (d AssistantSQLDialect) DatabaseType() string { return d.databaseType }

// QuoteIdentifier quotes a trusted identifier, optionally schema qualified.
func (d AssistantSQLDialect) QuoteIdentifier(identifier string) string {
	quote := "\""
	if d.databaseType == settings.MySqlDbType {
		quote = "`"
	}
	parts := strings.Split(identifier, ".")
	for i, part := range parts {
		parts[i] = quote + strings.ReplaceAll(part, quote, quote+quote) + quote
	}
	return strings.Join(parts, ".")
}

// TableName uses the existing engine's table mapper, including any prefix/schema.
func (d AssistantSQLDialect) TableName(bean any) string {
	return d.QuoteIdentifier(d.engine.TableName(bean, true))
}

// Placeholder is useful when building trusted SQL fragments with bound values.
func (d AssistantSQLDialect) Placeholder(index int) string {
	if d.databaseType == settings.PostgresDbType {
		return fmt.Sprintf("$%d", index)
	}
	return "?"
}

// Bind accepts portable '?' placeholders. Quoted strings/identifiers, comments,
// and PostgreSQL dollar-quoted strings are preserved. Services must not mix '?' and '$n'.
func (d AssistantSQLDialect) Bind(query string) string {
	if d.databaseType != settings.PostgresDbType {
		return query
	}
	var out strings.Builder
	index := 1
	for i := 0; i < len(query); {
		start := i
		switch {
		case query[i] == '\'' || query[i] == '"' || query[i] == '`':
			quote := query[i]
			i++
			for i < len(query) {
				if query[i] == quote {
					i++
					if i < len(query) && query[i] == quote {
						i++
						continue
					}
					break
				}
				i++
			}
		case strings.HasPrefix(query[i:], "--"):
			for i < len(query) && query[i] != '\n' {
				i++
			}
		case strings.HasPrefix(query[i:], "/*"):
			i += 2
			depth := 1
			for i < len(query) && depth > 0 {
				if strings.HasPrefix(query[i:], "/*") {
					depth++
					i += 2
				} else if strings.HasPrefix(query[i:], "*/") {
					depth--
					i += 2
				} else {
					i++
				}
			}
		case query[i] == '$':
			end := i + 1
			for end < len(query) && ((query[end] >= 'a' && query[end] <= 'z') || (query[end] >= 'A' && query[end] <= 'Z') || query[end] == '_' || (end > i+1 && query[end] >= '0' && query[end] <= '9')) {
				end++
			}
			if end < len(query) && query[end] == '$' {
				tag := query[i : end+1]
				if closeAt := strings.Index(query[end+1:], tag); closeAt >= 0 {
					i = end + 1 + closeAt + len(tag)
				} else {
					i = len(query)
				}
			} else {
				i++
			}
		case query[i] == '?':
			out.WriteString(d.Placeholder(index))
			index++
			i++
			continue
		default:
			i++
		}
		out.WriteString(query[start:i])
	}
	return out.String()
}

type assistantReadView struct {
	tx           *sql.Tx
	ctx          context.Context
	dialect      AssistantSQLDialect
	mu           sync.Mutex
	observations AssistantReadObservations
	rows         map[*AssistantRows]struct{}
	readErr      error
}

func (v *assistantReadView) Dialect() AssistantSQLDialect { return v.dialect }
func (v *assistantReadView) Observations() AssistantReadObservations {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.observations
}
func (v *assistantReadView) observe(start time.Time, rows int64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.observations.DatabaseDuration += time.Since(start)
	v.observations.RowsRead += rows
}

// finishRows closes only live cursors. Closed cursors are released immediately,
// so a full-history tool with many SQL batches does not retain one object/batch.
func (v *assistantReadView) finishRows() error {
	v.mu.Lock()
	live := make([]*AssistantRows, 0, len(v.rows))
	for rows := range v.rows {
		live = append(live, rows)
	}
	v.mu.Unlock()
	for _, rows := range live {
		_ = rows.Close()
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.readErr
}

func (v *assistantReadView) QueryContext(ctx context.Context, query string, args ...any) (*AssistantRows, error) {
	if ctx == nil {
		ctx = v.ctx
	}
	queryCtx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(v.ctx, cancel)
	cleanup := func() { stop(); cancel() }
	if err := v.ctx.Err(); err != nil {
		cleanup()
		return nil, err
	}
	start := time.Now()
	v.mu.Lock()
	v.observations.QueryCount++
	v.mu.Unlock()
	rows, err := v.tx.QueryContext(queryCtx, v.dialect.Bind(query), args...)
	v.observe(start, 0)
	if err != nil {
		cleanup()
		return nil, err
	}
	result := &AssistantRows{rows: rows, view: v, cleanup: cleanup}
	v.mu.Lock()
	if v.rows == nil {
		v.rows = make(map[*AssistantRows]struct{})
	}
	v.rows[result] = struct{}{}
	v.mu.Unlock()
	return result, nil
}

func (v *assistantReadView) QueryRowContext(ctx context.Context, query string, args ...any) *AssistantRow {
	rows, err := v.QueryContext(ctx, query, args...)
	return &AssistantRow{rows: rows, err: err}
}

// AssistantRows records database work while leaving application processing time out.
type AssistantRows struct {
	rows      *sql.Rows
	view      *assistantReadView
	cleanup   func()
	closeOnce sync.Once
	closeErr  error
	mu        sync.RWMutex
	closed    bool
	rowErr    error
}

func (r *AssistantRows) Next() bool {
	start := time.Now()
	next := r.rows.Next()
	var count int64
	if next {
		count = 1
	}
	r.view.observe(start, count)
	if !next {
		_ = r.Close()
	}
	return next
}
func (r *AssistantRows) Scan(dest ...any) error {
	start := time.Now()
	defer r.view.observe(start, 0)
	return r.rows.Scan(dest...)
}
func (r *AssistantRows) Err() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.closed {
		return r.rowErr
	}
	return r.rows.Err()
}
func (r *AssistantRows) Columns() ([]string, error) { return r.rows.Columns() }
func (r *AssistantRows) Close() error {
	r.closeOnce.Do(func() {
		start := time.Now()
		r.closeErr = r.rows.Close()
		// Capture the terminal read error before canceling our derived context.
		// database/sql's asynchronous watcher may otherwise report our own
		// cleanup cancellation after an explicit, successful early Close.
		r.mu.Lock()
		r.rowErr = r.rows.Err()
		r.closed = true
		r.mu.Unlock()
		r.cleanup()
		r.view.mu.Lock()
		delete(r.view.rows, r)
		if r.view.readErr == nil {
			r.view.readErr = errors.Join(r.closeErr, r.rowErr)
		}
		r.view.mu.Unlock()
		r.view.observe(start, 0)
	})
	return r.closeErr
}

// AssistantRow follows sql.Row's deferred error and single-row Scan semantics.
type AssistantRow struct {
	rows *AssistantRows
	err  error
}

func (r *AssistantRow) Err() error { return r.err }
func (r *AssistantRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	defer r.rows.Close()
	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}
	if err := r.rows.Scan(dest...); err != nil {
		return err
	}
	return r.rows.Close()
}

const assistantReadCleanupTimeout = 2 * time.Second

// WithAssistantReadView pins the selected user's primary pool for one tool call.
// The callback must complete before returning and must not retain the view/rows.
func (db *Database) WithAssistantReadView(c core.Context, fn func(AssistantReadView) error) (err error) {
	ctx := NewXOrmContextAdapter(c)
	conn, err := db.engineGroup.Master().DB().DB.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	isSQLite := db.databaseType == settings.Sqlite3DbType
	var previousQueryOnly int
	var tx *sql.Tx
	var view *assistantReadView
	pragmaChanged := false
	transactionFinished := true
	defer func() {
		if view != nil {
			err = errors.Join(err, view.finishRows())
		}
		if tx != nil && !transactionFinished {
			rollbackErr := tx.Rollback()
			// ErrTxDone does not mean automatic rollback has finished: sql.Tx
			// marks itself done BEFORE calling the driver. Never restore a pragma
			// on that path. Discarding via Raw waits for outstanding operations
			// (Conn's close lock) and prevents a polluted connection being reused.
			transactionFinished = rollbackErr == nil
			if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
				err = errors.Join(err, rollbackErr)
			}
		}
		if !transactionFinished {
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
			return
		}
		if pragmaChanged {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), assistantReadCleanupTimeout)
			defer cancel()
			_, restoreErr := conn.ExecContext(cleanupCtx, fmt.Sprintf("PRAGMA query_only = %d", previousQueryOnly))
			if restoreErr != nil {
				_ = conn.Raw(func(any) error { return driver.ErrBadConn })
				err = errors.Join(err, fmt.Errorf("restore assistant read connection: %w", restoreErr))
			}
		}
	}()

	if isSQLite {
		if err = conn.QueryRowContext(ctx, "PRAGMA query_only").Scan(&previousQueryOnly); err != nil {
			return err
		}
		// A failed Exec can still have changed the driver state; restore/discard.
		pragmaChanged = true
		if _, err = conn.ExecContext(ctx, "PRAGMA query_only = 1"); err != nil {
			return err
		}
	}
	opts := &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true}
	if isSQLite {
		opts = nil
	}
	tx, err = conn.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	transactionFinished = false
	view = &assistantReadView{tx: tx, ctx: ctx, dialect: AssistantSQLDialect{databaseType: db.databaseType, engine: db.engineGroup.Master()}, observations: AssistantReadObservations{RequestID: ctx.requestId}}
	if err = fn(view); err != nil {
		return err
	}
	if err = view.finishRows(); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	err = tx.Commit()
	transactionFinished = err == nil
	return err
}
