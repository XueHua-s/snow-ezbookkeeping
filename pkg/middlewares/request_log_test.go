package middlewares

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

func TestAssistantRequestLogDoesNotPersistQueryCanaries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "request.log")
	if err := log.SetLoggerConfiguration(&settings.Config{EnableFileLog: true, EnableRequestLog: true, EnableQueryLog: true, FileLogPath: path}, true); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = log.SetLoggerConfiguration(&settings.Config{EnableRequestLog: true, EnableQueryLog: true}, true)
	})
	gin.SetMode(gin.TestMode)
	for _, endpoint := range []string{"conversations/search.json", "conversations/export.json", "runs/stream.json"} {
		gc, _ := gin.CreateTestContext(httptest.NewRecorder())
		gc.Request = httptest.NewRequest("GET", "/api/v1/llm/assistant/"+endpoint+"?query=PRIVATE_FINANCIAL_CANARY&cursor=PRIVATE_CURSOR_CANARY", nil)
		c := core.WrapWebContext(gc, nil)
		c.SetContextId("controlled-request-id")
		RequestLog(c)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	output := string(raw)
	if strings.Contains(output, "CANARY") || strings.Contains(output, "?query") {
		t.Fatalf("private query persisted: %s", output)
	}
	if !strings.Contains(output, "/conversations/search.json") || !strings.Contains(output, "controlled-request-id") {
		t.Fatal("controlled operational metadata missing")
	}
}
