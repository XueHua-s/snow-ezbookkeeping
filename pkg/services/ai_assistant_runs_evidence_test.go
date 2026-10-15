package services

import (
	"encoding/json"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestAssistantRunEvidenceExactAmountsAndContinuation(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	cursor := assistantCursor{UID: "7", Tool: data.ToolQueryTransactions, IssuedAt: now.Unix(), ExpiresAt: now.Add(24 * time.Hour).Unix()}
	token, err := assistantSignCursor(cursor, []byte("isolated-test-key"))
	require.NoError(t, err)
	run := &models.AssistantMessage{Uid: 7}
	result := llm.AssistantRunResult{Queries: []data.QueryEvidence{{QueryID: "query-1", Tool: data.ToolQueryTransactions, QueriedAt: now.Format(time.RFC3339)}}, References: []data.TransactionReference{{ID: "8", Time: now.Format(time.RFC3339), Type: "transfer", SourceAmount: "-92233720368547758.08", DestinationAmount: "92233720368547758.07", Currency: "CNY", DestinationCurrency: "USD", HideAmount: true}}, Continuation: &data.ToolContinuation{Tool: data.ToolQueryTransactions, Cursor: token}}
	require.NoError(t, assistantRunEvidence(run, result))
	var references []models.AssistantReferenceDTO
	require.NoError(t, json.Unmarshal([]byte(run.ReferencesJSON), &references))
	require.Equal(t, "-9223372036854775808", references[0].SourceAmountMinor)
	require.Equal(t, "9223372036854775807", references[0].DestinationAmountMinor)
	require.True(t, references[0].HideAmount)
	require.Equal(t, cursor.ExpiresAt*1000, run.ContinuationExpiresAt)
	require.Equal(t, "query-1", run.ContinuationQueryId)
	require.False(t, strings.Contains(run.QueriesJSON, token))
	require.False(t, strings.Contains(run.ReferencesJSON, token))
}
