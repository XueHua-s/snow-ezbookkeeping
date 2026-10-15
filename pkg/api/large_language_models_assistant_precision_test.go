package api

import (
	"encoding/json"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/stretchr/testify/require"
)

func TestAssistantLegacyReferencePreservesExactMinorAmounts(t *testing.T) {
	references, err := assistantReferences([]data.TransactionReference{{
		ID: "9223372036854775807", Time: "2026-10-15T08:00:00Z", Type: "transfer",
		SourceAmount: "-92233720368547758.08", DestinationAmount: "92233720368547758.07",
		Currency: "USD", DestinationCurrency: "CNY", HideAmount: true,
	}})
	require.NoError(t, err)
	require.Equal(t, "-9223372036854775808", references[0].SourceAmountMinor)
	require.Equal(t, "9223372036854775807", references[0].DestinationAmountMinor)
	require.True(t, references[0].HideAmount)
	encoded, err := json.Marshal(references)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"sourceAmountMinor":"-9223372036854775808"`)
	require.Contains(t, string(encoded), `"destinationAmountMinor":"9223372036854775807"`)
}

func TestAssistantReferenceDistinguishesMissingAndZeroDestination(t *testing.T) {
	for _, amount := range []string{"", "0.00"} {
		references, err := assistantReferences([]data.TransactionReference{{ID: "1", Time: "2026-10-15T08:00:00Z", Type: "expense", SourceAmount: "0.00", DestinationAmount: amount}})
		require.NoError(t, err)
		require.Equal(t, "0", references[0].SourceAmountMinor)
		if amount == "" {
			require.Empty(t, references[0].DestinationAmountMinor)
		} else {
			require.Equal(t, "0", references[0].DestinationAmountMinor)
		}
	}
}
