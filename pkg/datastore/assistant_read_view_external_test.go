package datastore

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"xorm.io/xorm"
)

// External tests require explicit opt-in AND a dedicated bootstrap database named
// ebk_harness_test_*. Each run creates/drops only its own uniquely named database.
// No existing schema/table is modified or cleaned by these tests.
func assistantExternalFixture(t *testing.T, kind string) (*Database, *sql.DB, *sql.DB) {
	t.Helper()
	if os.Getenv("EBK_ASSISTANT_DB_TEST") != "1" {
		t.Skip("external database not opted in: EBK_ASSISTANT_DB_TEST=1")
	}
	env := "EBK_ASSISTANT_MYSQL_TEST_DSN"
	if kind == settings.PostgresDbType {
		env = "EBK_ASSISTANT_POSTGRES_TEST_DSN"
	}
	dsn := os.Getenv(env)
	if dsn == "" {
		t.Skip("missing isolated test connection: " + env)
	}
	name := fmt.Sprintf("ebk_harness_test_%d", time.Now().UnixNano())
	var testDSN string
	if kind == settings.MySqlDbType {
		cfg, err := mysql.ParseDSN(dsn)
		if err != nil {
			t.Fatal("invalid MySQL test DSN")
		}
		if !strings.HasPrefix(cfg.DBName, "ebk_harness_test_") {
			t.Fatal("refusing non-test MySQL database")
		}
		cfg.DBName = name
		testDSN = cfg.FormatDSN()
	} else {
		parsed, err := url.Parse(dsn)
		if err != nil || !(parsed.Scheme == "postgres" || parsed.Scheme == "postgresql") {
			t.Fatal("PostgreSQL test DSN must be a URL")
		}
		if !strings.HasPrefix(strings.TrimPrefix(parsed.Path, "/"), "ebk_harness_test_") {
			t.Fatal("refusing non-test PostgreSQL database")
		}
		parsed.Path = "/" + name
		testDSN = parsed.String()
	}
	admin, err := sql.Open(kind, dsn)
	if err != nil {
		t.Fatal("open test bootstrap database failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	quote := "`"
	if kind == settings.PostgresDbType {
		quote = "\""
	}
	if _, err = admin.ExecContext(ctx, "CREATE DATABASE "+quote+name+quote); err != nil {
		admin.Close()
		t.Fatal("create isolated database:", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.ExecContext(ctx, "DROP DATABASE "+quote+name+quote); err != nil {
			t.Errorf("cleanup created test database %s: %v", name, err)
		}
		admin.Close()
	})
	group, err := xorm.NewEngineGroup(kind, []string{testDSN})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { group.Close() })
	pool := group.Master().DB().DB
	pool.SetMaxOpenConns(1)
	pool.SetMaxIdleConns(1)
	writer, err := sql.Open(kind, testDSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { writer.Close() })
	for _, q := range []string{"CREATE TABLE fixture (id BIGINT PRIMARY KEY, amount BIGINT)", "INSERT INTO fixture VALUES (1,10),(2,20),(3,30)"} {
		if _, err = pool.ExecContext(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	// Deliberately make the session default weaker; the read view must override it.
	isolation := "SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED"
	if kind == settings.PostgresDbType {
		isolation = "SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ COMMITTED"
	}
	if _, err = pool.ExecContext(ctx, isolation); err != nil {
		t.Fatal(err)
	}
	return &Database{databaseType: kind, engineGroup: group}, pool, writer
}

func TestAssistantReadViewExternal(t *testing.T) {
	for _, kind := range []string{settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			db, pool, writer := assistantExternalFixture(t, kind)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			err := db.WithAssistantReadView(assistantTestContext(ctx), func(v AssistantReadView) error {
				var count int
				if err := v.QueryRowContext(ctx, "SELECT COUNT(*) FROM fixture WHERE amount >= ?", 10).Scan(&count); err != nil {
					return err
				}
				if count != 3 {
					t.Fatal(count)
				}
				identityQuery := "SELECT CONNECTION_ID()"
				if kind == settings.PostgresDbType {
					identityQuery = "SELECT pg_backend_pid()"
				}
				var first, second int64
				if err := v.QueryRowContext(ctx, identityQuery).Scan(&first); err != nil {
					return err
				}
				if _, err := writer.ExecContext(ctx, "INSERT INTO fixture VALUES (4,40)"); err != nil {
					return err
				}
				if err := v.QueryRowContext(ctx, "SELECT COUNT(*) FROM fixture").Scan(&count); err != nil {
					return err
				}
				if count != 3 {
					t.Fatalf("READ COMMITTED leaked into view: %d", count)
				}
				if err := v.QueryRowContext(ctx, identityQuery).Scan(&second); err != nil {
					return err
				}
				if first != second {
					t.Fatal("connection changed")
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			err = db.WithAssistantReadView(assistantTestContext(ctx), func(v AssistantReadView) error {
				rows, err := v.QueryContext(ctx, "UPDATE fixture SET amount=99")
				if err != nil {
					return err
				}
				defer rows.Close()
				for rows.Next() {
				}
				return rows.Err()
			})
			if err == nil {
				t.Fatal("read-only transaction allowed write")
			}
			var sum int
			if err = pool.QueryRowContext(ctx, "SELECT SUM(amount) FROM fixture").Scan(&sum); err != nil || sum != 100 {
				t.Fatalf("write changed fixture: %d %v", sum, err)
			}
			cancelCtx, cancelQuery := context.WithTimeout(context.Background(), 30*time.Millisecond)
			defer cancelQuery()
			start := time.Now()
			err = db.WithAssistantReadView(assistantTestContext(cancelCtx), func(v AssistantReadView) error {
				q := "SELECT SLEEP(60)"
				if kind == settings.PostgresDbType {
					q = "SELECT pg_sleep(60)"
				}
				var result any
				return v.QueryRowContext(context.Background(), q).Scan(&result)
			})
			if err == nil || time.Since(start) > 2*time.Second {
				t.Fatalf("driver cancellation failed: duration=%v err=%v", time.Since(start), err)
			}
			if _, err = pool.ExecContext(ctx, "UPDATE fixture SET amount=amount WHERE id=1"); err != nil {
				t.Fatal("connection not reusable after cancellation:", err)
			}
			if pool.Stats().InUse != 0 {
				t.Fatal("connection leaked")
			}
		})
	}
}
