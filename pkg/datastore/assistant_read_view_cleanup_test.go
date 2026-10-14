package datastore

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
	xormcore "xorm.io/xorm/core"
)

// This proxy keeps the real SQLite engine/query cancellation but exposes the
// automatic sql.Tx rollback timing, which cannot be exercised deterministically
// with sleeps alone. SessionResetter+Validator prevent database/sql itself from
// discarding the connection, so the test proves our own discard path.
type assistantFaultDriver struct {
	opens            atomic.Int64
	closes           atomic.Int64
	restores         atomic.Int64
	failRestore      bool
	failBegin        bool
	failCommit       bool
	rollbackStarted  chan struct{}
	releaseRollback  chan struct{}
	rollbackFinished atomic.Bool
}

func (d *assistantFaultDriver) Open(dsn string) (driver.Conn, error) {
	c, err := (&sqlite3.SQLiteDriver{}).Open(dsn)
	if err != nil {
		return nil, err
	}
	d.opens.Add(1)
	return &assistantFaultConn{SQLiteConn: c.(*sqlite3.SQLiteConn), fault: d}, nil
}

type assistantFaultConnector struct {
	d   *assistantFaultDriver
	dsn string
}

func (c assistantFaultConnector) Connect(context.Context) (driver.Conn, error) {
	return c.d.Open(c.dsn)
}
func (c assistantFaultConnector) Driver() driver.Driver { return c.d }

type assistantFaultConn struct {
	*sqlite3.SQLiteConn
	fault *assistantFaultDriver
}

func (c *assistantFaultConn) Close() error                       { c.fault.closes.Add(1); return c.SQLiteConn.Close() }
func (c *assistantFaultConn) IsValid() bool                      { return true }
func (c *assistantFaultConn) ResetSession(context.Context) error { return nil }
func (c *assistantFaultConn) ExecContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
	if q == "PRAGMA query_only = 0" {
		c.fault.restores.Add(1)
		if c.fault.failRestore {
			return nil, errors.New("injected restore failure")
		}
	}
	return c.SQLiteConn.ExecContext(ctx, q, args)
}
func (c *assistantFaultConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if c.fault.failBegin {
		return nil, errors.New("injected begin failure")
	}
	tx, err := c.SQLiteConn.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &assistantFaultTx{Tx: tx, fault: c.fault}, nil
}

type assistantFaultTx struct {
	driver.Tx
	fault *assistantFaultDriver
}

func (tx *assistantFaultTx) Commit() error {
	if tx.fault.failCommit {
		return errors.New("injected commit failure")
	}
	return tx.Tx.Commit()
}
func (tx *assistantFaultTx) Rollback() error {
	if tx.fault.rollbackStarted != nil {
		close(tx.fault.rollbackStarted)
		<-tx.fault.releaseRollback
	}
	err := tx.Tx.Rollback()
	tx.fault.rollbackFinished.Store(true)
	return err
}
func assistantFaultFixture(t *testing.T, fault *assistantFaultDriver) (*Database, *sql.DB) {
	t.Helper()
	db, oldPool := assistantSQLiteFixture(t)
	master := db.engineGroup.Master()
	dsn := master.DataSourceName()
	dialect := master.Dialect()
	if err := oldPool.Close(); err != nil {
		t.Fatal(err)
	}
	pool := sql.OpenDB(assistantFaultConnector{d: fault, dsn: dsn})
	pool.SetMaxOpenConns(1)
	pool.SetMaxIdleConns(1)
	engine, err := xorm.NewEngineWithDialectAndDB("sqlite3", dsn, dialect, xormcore.FromDB(pool))
	if err != nil {
		t.Fatal(err)
	}
	group, err := xorm.NewEngineGroup(engine, []*xorm.Engine{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { group.Close() })
	db.engineGroup = group
	return db, pool
}

func TestAssistantReadViewSQLiteAutomaticRollbackDiscard(t *testing.T) {
	fault := &assistantFaultDriver{rollbackStarted: make(chan struct{}), releaseRollback: make(chan struct{})}
	db, pool := assistantFaultFixture(t, fault)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- db.WithAssistantReadView(assistantTestContext(ctx), func(v AssistantReadView) error {
			var count int
			if err := v.QueryRowContext(ctx, "SELECT COUNT(*) FROM fixture").Scan(&count); err != nil {
				return err
			}
			cancel()
			<-fault.rollbackStarted
			return ctx.Err()
		})
	}()
	select {
	case <-fault.rollbackStarted:
	case <-time.After(time.Second):
		t.Fatal("automatic rollback did not start")
	}
	// sql.Tx has set done=true, but the driver is still blocked in Rollback.
	select {
	case err := <-done:
		t.Fatalf("returned before rollback finished: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	if fault.restores.Load() != 0 {
		t.Fatal("restored query_only during rollback")
	}
	close(fault.releaseRollback)
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("cleanup did not finish")
	}
	if !fault.rollbackFinished.Load() || fault.closes.Load() != 1 || fault.restores.Load() != 0 {
		t.Fatalf("unsafe discard: finished=%v closed=%d restored=%d", fault.rollbackFinished.Load(), fault.closes.Load(), fault.restores.Load())
	}
	assertSQLiteWritable(t, pool)
	if fault.opens.Load() != 2 {
		t.Fatal("discarded connection reused")
	}
}

func TestAssistantReadViewSQLiteFailedCleanupDiscardsConnection(t *testing.T) {
	for _, stage := range []string{"restore", "commit", "begin"} {
		t.Run(stage, func(t *testing.T) {
			fault := &assistantFaultDriver{failRestore: stage == "restore", failCommit: stage == "commit", failBegin: stage == "begin"}
			db, pool := assistantFaultFixture(t, fault)
			err := db.WithAssistantReadView(nil, func(v AssistantReadView) error {
				var count int
				return v.QueryRowContext(nil, "SELECT COUNT(*) FROM fixture").Scan(&count)
			})
			if err == nil {
				t.Fatal("injected failure lost")
			}
			if stage == "begin" {
				if fault.closes.Load() != 0 || fault.restores.Load() != 1 {
					t.Fatal("begin failure must restore before reuse")
				}
			} else if fault.closes.Load() != 1 {
				t.Fatal("uncertain connection was not discarded")
			}
			assertSQLiteWritable(t, pool)
		})
	}
}
