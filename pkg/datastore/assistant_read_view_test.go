package datastore

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

func assistantSQLiteFixture(t *testing.T) (*Database, *sql.DB) {
	t.Helper()
	group, err := xorm.NewEngineGroup("sqlite3", []string{"file:" + filepath.Join(t.TempDir(), "assistant.db") + "?_journal_mode=WAL&_busy_timeout=1000"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = group.Close() })
	db := &Database{databaseType: settings.Sqlite3DbType, engineGroup: group}
	pool := group.Master().DB().DB
	pool.SetMaxOpenConns(1)
	pool.SetMaxIdleConns(1)
	if _, err = pool.Exec("CREATE TABLE fixture (id INTEGER PRIMARY KEY, amount INTEGER); INSERT INTO fixture VALUES (1,10),(2,20),(3,30)"); err != nil {
		t.Fatal(err)
	}
	return db, pool
}

func assistantTestContext(ctx context.Context) core.Context {
	return core.NewAssistantContext(ctx, "test-request", "en", "127.0.0.1")
}

func assertSQLiteWritable(t *testing.T, pool *sql.DB) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var mode int
	if err := pool.QueryRowContext(ctx, "PRAGMA query_only").Scan(&mode); err != nil || mode != 0 {
		t.Fatalf("query_only leaked: mode=%d err=%v", mode, err)
	}
	if _, err := pool.ExecContext(ctx, "UPDATE fixture SET amount=amount WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	if pool.Stats().InUse != 0 {
		t.Fatal("connection leaked")
	}
}

func TestAssistantReadViewSQLiteSnapshotAndTelemetry(t *testing.T) {
	db, pool := assistantSQLiteFixture(t)
	// A separate writer uses the same isolated file, not the reader's pinned pool.
	writer, err := sql.Open("sqlite3", db.engineGroup.Master().DataSourceName())
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	err = db.WithAssistantReadView(assistantTestContext(context.Background()), func(v AssistantReadView) error {
		var count int
		if err := v.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM fixture").Scan(&count); err != nil {
			return err
		}
		if count != 3 {
			t.Fatal(count)
		}
		if _, err := writer.Exec("INSERT INTO fixture VALUES(4,40)"); err != nil {
			return err
		}
		rows, err := v.QueryContext(context.Background(), "SELECT amount FROM fixture WHERE amount >= ? ORDER BY id", 20)
		if err != nil {
			return err
		}
		defer rows.Close()
		sum := 0
		for rows.Next() {
			var amount int
			if err := rows.Scan(&amount); err != nil {
				return err
			}
			sum += amount
		}
		if err := rows.Err(); err != nil {
			return err
		}
		if sum != 50 {
			t.Fatalf("snapshot changed: %d", sum)
		}
		if err := v.QueryRowContext(context.Background(), "SELECT id FROM fixture WHERE id = ?", 99).Scan(&count); !errors.Is(err, sql.ErrNoRows) {
			t.Fatal(err)
		}
		obs := v.Observations()
		if obs.QueryCount != 3 || obs.RowsRead != 3 || obs.DatabaseDuration <= 0 || obs.RequestID != "test-request" {
			t.Fatalf("observations: %+v", obs)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err = pool.QueryRow("SELECT COUNT(*) FROM fixture").Scan(&count); err != nil || count != 4 {
		t.Fatalf("writer missing: %d %v", count, err)
	}
	assertSQLiteWritable(t, pool)
}

func TestAssistantReadViewSQLiteWriteProtectionAndRollback(t *testing.T) {
	for _, statement := range []string{"UPDATE fixture SET amount=99", "DELETE FROM fixture", "CREATE TABLE forbidden (id INT)"} {
		t.Run(statement, func(t *testing.T) {
			db, pool := assistantSQLiteFixture(t)
			err := db.WithAssistantReadView(assistantTestContext(context.Background()), func(v AssistantReadView) error {
				rows, err := v.QueryContext(context.Background(), statement)
				if err != nil {
					return err
				}
				defer rows.Close()
				for rows.Next() {
				}
				return rows.Err()
			})
			if err == nil {
				t.Fatal("write succeeded")
			}
			assertSQLiteWritable(t, pool)
			var sum int
			if err := pool.QueryRow("SELECT SUM(amount) FROM fixture").Scan(&sum); err != nil || sum != 60 {
				t.Fatalf("data changed: %d %v", sum, err)
			}
		})
	}
}

func TestAssistantReadViewSQLiteCancelLongQueryAndReuse(t *testing.T) {
	db, pool := assistantSQLiteFixture(t)
	for i := 0; i < 25; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		started := time.Now()
		err := db.WithAssistantReadView(assistantTestContext(ctx), func(v AssistantReadView) error {
			var sum int64
			return v.QueryRowContext(context.Background(), "WITH RECURSIVE n(x) AS (VALUES(1) UNION ALL SELECT x+1 FROM n WHERE x<1000000000) SELECT SUM(x) FROM n").Scan(&sum)
		})
		cancel()
		if err == nil {
			t.Fatal("long query was not canceled")
		}
		if time.Since(started) > time.Second {
			t.Fatalf("driver did not interrupt promptly: %v", time.Since(started))
		}
		assertSQLiteWritable(t, pool)
	}
}

func TestAssistantReadViewSQLiteCancelDuringRows(t *testing.T) {
	db, pool := assistantSQLiteFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var obs AssistantReadObservations
	err := db.WithAssistantReadView(assistantTestContext(ctx), func(v AssistantReadView) error {
		rows, err := v.QueryContext(ctx, "SELECT id FROM fixture ORDER BY id")
		if err != nil {
			return err
		}
		defer rows.Close()
		if !rows.Next() {
			return rows.Err()
		}
		var id int
		if err = rows.Scan(&id); err != nil {
			return err
		}
		cancel()
		<-ctx.Done()
		_ = rows.Close()
		obs = v.Observations()
		return ctx.Err()
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if obs.QueryCount != 1 || obs.RowsRead != 1 {
		t.Fatalf("actual read count: %+v", obs)
	}
	assertSQLiteWritable(t, pool)
}

func TestAssistantReadViewSQLiteCleanupOnErrorPanicAndPreviousState(t *testing.T) {
	db, pool := assistantSQLiteFixture(t)
	sentinel := errors.New("callback failed")
	if err := db.WithAssistantReadView(nil, func(v AssistantReadView) error {
		_, err := v.QueryContext(nil, "SELECT * FROM fixture")
		if err != nil {
			return err
		}
		return sentinel
	}); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	assertSQLiteWritable(t, pool)
	func() {
		defer func() {
			if recover() != "test panic" {
				t.Fatal("panic lost")
			}
		}()
		_ = db.WithAssistantReadView(nil, func(AssistantReadView) error { panic("test panic") })
	}()
	assertSQLiteWritable(t, pool)
	if _, err := pool.Exec("PRAGMA query_only=1"); err != nil {
		t.Fatal(err)
	}
	if err := db.WithAssistantReadView(nil, func(v AssistantReadView) error { return nil }); err != nil {
		t.Fatal(err)
	}
	var mode int
	if err := pool.QueryRow("PRAGMA query_only").Scan(&mode); err != nil || mode != 1 {
		t.Fatalf("previous query_only lost: %d %v", mode, err)
	}
}

func TestAssistantReadViewSQLDialect(t *testing.T) {
	db, _ := assistantSQLiteFixture(t)
	db.engineGroup.Master().SetTableMapper(names.NewPrefixMapper(names.SnakeMapper{}, "ebk_"))
	type Fixture struct{ ID int64 }
	d := AssistantSQLDialect{databaseType: settings.PostgresDbType, engine: db.engineGroup.Master()}
	if got := d.TableName(new(Fixture)); got != "\"ebk_fixture\"" {
		t.Fatal(got)
	}
	query := "SELECT '?', \"?\", $$?$$, $tag$?$tag$, ? /* ? /* ? */ */ -- ?\n, ?"
	want := "SELECT '?', \"?\", $$?$$, $tag$?$tag$, $1 /* ? /* ? */ */ -- ?\n, $2"
	if got := d.Bind(query); got != want {
		t.Fatalf("%s", got)
	}
	if got := d.QuoteIdentifier("odd\"name.column"); got != "\"odd\"\"name\".\"column\"" {
		t.Fatal(got)
	}
	d.databaseType = settings.MySqlDbType
	if got := d.Bind("SELECT ?"); got != "SELECT ?" {
		t.Fatal(got)
	}
	if !strings.HasPrefix(d.TableName(new(Fixture)), "`") {
		t.Fatal("mysql quoting")
	}
}

func TestAssistantReadViewAlwaysUsesPrimary(t *testing.T) {
	primary, primaryPool := assistantSQLiteFixture(t)
	replica, replicaPool := assistantSQLiteFixture(t)
	if _, err := replicaPool.Exec("UPDATE fixture SET amount=999"); err != nil {
		t.Fatal(err)
	}
	group, err := xorm.NewEngineGroup(primary.engineGroup.Master(), []*xorm.Engine{replica.engineGroup.Master()})
	if err != nil {
		t.Fatal(err)
	}
	db := &Database{databaseType: settings.Sqlite3DbType, engineGroup: group}
	for i := 0; i < 5; i++ {
		if err := db.WithAssistantReadView(nil, func(v AssistantReadView) error {
			var sum int
			if err := v.QueryRowContext(nil, "SELECT SUM(amount) FROM fixture").Scan(&sum); err != nil {
				return err
			}
			if sum != 60 {
				t.Fatalf("read replica instead of primary: %d", sum)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	assertSQLiteWritable(t, primaryPool)
}

func TestXOrmContextAdapterActuallyCancelsSQLiteSession(t *testing.T) {
	db, pool := assistantSQLiteFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	sess := db.NewSession(assistantTestContext(ctx))
	defer sess.Close()
	start := time.Now()
	_, err := sess.QueryString("WITH RECURSIVE n(x) AS (VALUES(1) UNION ALL SELECT x+1 FROM n WHERE x<1000000000) SELECT SUM(x) FROM n")
	if err == nil || time.Since(start) > time.Second {
		t.Fatalf("xorm query cancellation failed: %v duration=%v", err, time.Since(start))
	}
	sess.Close()
	assertSQLiteWritable(t, pool)
}

func TestAssistantReadViewReleasesClosedCursorsAndRetainsReadError(t *testing.T) {
	db, pool := assistantSQLiteFixture(t)
	err := db.WithAssistantReadView(nil, func(v AssistantReadView) error {
		if got := v.Observations(); got.QueryCount != 0 || got.RowsRead != 0 || got.DatabaseDuration != 0 {
			t.Fatalf("empty view metrics: %+v", got)
		}
		for i := 0; i < 100; i++ {
			var id int
			if err := v.QueryRowContext(nil, "SELECT id FROM fixture WHERE id=1").Scan(&id); err != nil {
				return err
			}
		}
		internal := v.(*assistantReadView)
		if len(internal.rows) != 0 {
			t.Fatal("closed cursors retained for every SQL batch")
		}
		rows, err := v.QueryContext(nil, "SELECT abs(-9223372036854775808)")
		if err != nil {
			return err
		}
		for rows.Next() {
		}
		if rows.Err() == nil {
			t.Fatal("fixture must fail during iteration")
		}
		// Even when a callback overlooks Err, a failed read cannot commit as success.
		return nil
	})
	if err == nil {
		t.Fatal("iteration error lost when closed cursor removed")
	}
	assertSQLiteWritable(t, pool)
}
