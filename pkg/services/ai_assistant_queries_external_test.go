package services

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
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// External fixture creation is opt-in and refuses a non-test bootstrap name.
// Only a newly created, unique database is modified and dropped.
func assistantExternalServiceConfig(t *testing.T, kind string) *settings.DatabaseConfig {
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
	name := fmt.Sprintf("ebk_harness_test_query_%d", time.Now().UnixNano())
	config := &settings.DatabaseConfig{DatabaseType: kind, DatabaseName: name, MaxOpenConnection: 2, MaxIdleConnection: 2}
	quote := "`"
	if kind == settings.MySqlDbType {
		parsed, err := mysql.ParseDSN(dsn)
		if err != nil {
			t.Fatal("invalid MySQL test DSN")
		}
		if !strings.HasPrefix(parsed.DBName, "ebk_harness_test_") {
			t.Fatal("refusing non-test MySQL database")
		}
		config.DatabaseHost = parsed.Addr
		config.DatabaseUser = parsed.User
		config.DatabasePassword = parsed.Passwd
	} else {
		parsed, err := url.Parse(dsn)
		if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || parsed.User == nil {
			t.Fatal("invalid PostgreSQL test URL")
		}
		if !strings.HasPrefix(strings.TrimPrefix(parsed.Path, "/"), "ebk_harness_test_") {
			t.Fatal("refusing non-test PostgreSQL database")
		}
		config.DatabaseHost = parsed.Host
		config.DatabaseUser = parsed.User.Username()
		config.DatabasePassword, _ = parsed.User.Password()
		config.DatabaseSSLMode = parsed.Query().Get("sslmode")
		if config.DatabaseSSLMode == "" {
			config.DatabaseSSLMode = "disable"
		}
		quote = "\""
	}
	admin, err := sql.Open(kind, dsn)
	if err != nil {
		t.Fatal("open isolated test bootstrap")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if _, err = admin.ExecContext(ctx, "CREATE DATABASE "+quote+name+quote); err != nil {
		admin.Close()
		t.Fatal("create isolated database", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.ExecContext(ctx, "DROP DATABASE "+quote+name+quote); err != nil {
			t.Error("drop newly created test database", err)
		}
		admin.Close()
	})
	return config
}
func TestAssistantQueryFoundationExternal(t *testing.T) {
	for _, kind := range []string{settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			s, _ := assistantServiceFixtureConfig(t, assistantExternalServiceConfig(t, kind))
			s.handlers[data.ToolQueryTransactions] = assistantScanTestHandler
			for _, raw := range []string{`{"filter":{"keyword":"%"}}`, `{"filter":{"keyword":"_"}}`, `{"filter":{"keyword":"\\"}}`, `{"filter":{"keyword":"ABC"}}`} {
				r := assistantCall(s, data.ToolQueryTransactions, raw)
				if !r.OK || string(r.Data) != `["100"]` {
					t.Fatalf("literal query failed: %s %+v", r.Data, r.Error)
				}
			}
			r := assistantCall(s, data.ToolQueryTransactions, `{}`)
			if !r.OK || string(r.Data) != `["100","101","102","104"]` {
				t.Fatalf("full effective logical set: %s %+v", r.Data, r.Error)
			}
			r = assistantCall(s, data.ToolQueryTransactions, `{"filter":{"account_ids":["11","21"]}}`)
			if r.OK || r.Error.Code != data.ToolErrorNotFoundOrNotAccessible {
				t.Fatal("mixed ownership accepted", r)
			}
			r = assistantCall(s, data.ToolQueryTransactions, `{"filter":{"tag_ids":["41","42"],"tag_match":"all"}}`)
			if !r.OK || string(r.Data) != `["100"]` {
				t.Fatal("tag all mismatch", r)
			}
			r = assistantCall(s, data.ToolQueryTransactions, `{"filter":{"account_ids":["12"],"currency":"USD","amount_min":"30"}}`)
			if !r.OK || string(r.Data) != `[]` {
				t.Fatal("cross-side mismatch", r)
			}
		})
	}
}
