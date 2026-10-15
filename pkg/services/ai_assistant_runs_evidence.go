package services

import (
	"encoding/base64"
	"encoding/json"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"math/big"
	"strconv"
	"strings"
	"time"
)

func assistantRunQueries(queries []data.QueryEvidence) []models.AIAssistantQuery {
	result := make([]models.AIAssistantQuery, 0, len(queries))
	for _, q := range queries {
		s := q.Scope
		scope := models.AIAssistantQueryScope{StartTime: s.StartTime, EndTime: s.EndTime, Timezone: s.Timezone, Currency: s.Currency, Types: s.Types, AccountIDs: s.AccountIDs, CategoryIDs: s.CategoryIDs, TagIDs: s.TagIDs, TagMatch: s.TagMatch, NoTags: s.NoTags, TransferSide: s.TransferSide, AmountMin: s.AmountMin, AmountMax: s.AmountMax, Keyword: s.Keyword, IncludesHidden: s.IncludesHidden, ExcludesDeleted: s.ExcludesDeleted}
		for _, p := range s.Periods {
			scope.Periods = append(scope.Periods, models.AIAssistantQueryPeriod{Name: p.Name, StartTime: p.StartTime, EndTime: p.EndTime, MatchedCount: p.MatchedCount})
		}
		warnings := q.Warnings
		if warnings == nil {
			warnings = []string{}
		}
		result = append(result, models.AIAssistantQuery{QueryID: q.QueryID, Tool: string(q.Tool), Scope: scope, Coverage: string(q.Coverage), QueriedAt: q.QueriedAt, Warnings: warnings})
	}
	return result
}

func assistantRunMinorAmount(amount string) (int64, error) {
	negative := strings.HasPrefix(amount, "-")
	digits := strings.TrimPrefix(amount, "-")
	parts := strings.Split(digits, ".")
	if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) != 2 {
		return 0, models.AssistantReasonProviderUnavailable
	}
	for _, r := range parts[0] + parts[1] {
		if r < '0' || r > '9' {
			return 0, models.AssistantReasonProviderUnavailable
		}
	}
	number, ok := new(big.Int).SetString(parts[0]+parts[1], 10)
	if !ok {
		return 0, models.AssistantReasonProviderUnavailable
	}
	if negative {
		number.Neg(number)
	}
	if !number.IsInt64() {
		return 0, models.AssistantReasonProviderUnavailable
	}
	return number.Int64(), nil
}

func assistantRunEvidence(run *models.AssistantMessage, result llm.AssistantRunResult) error {
	queries, err := json.Marshal(assistantRunQueries(result.Queries))
	if err != nil {
		return err
	}
	refs := make([]models.AssistantReferenceDTO, 0, len(result.References))
	for _, ref := range result.References {
		id, err := strconv.ParseInt(ref.ID, 10, 64)
		if err != nil || id <= 0 {
			return models.AssistantReasonProviderUnavailable
		}
		at, err := time.Parse(time.RFC3339, ref.Time)
		if err != nil {
			return models.AssistantReasonProviderUnavailable
		}
		amount, err := assistantRunMinorAmount(ref.SourceAmount)
		if err != nil {
			return err
		}
		destination := ""
		if ref.DestinationAmount != "" {
			n, err := assistantRunMinorAmount(ref.DestinationAmount)
			if err != nil {
				return err
			}
			destination = strconv.FormatInt(n, 10)
		}
		kind := models.TRANSACTION_TYPE_MODIFY_BALANCE
		switch ref.Type {
		case "income":
			kind = models.TRANSACTION_TYPE_INCOME
		case "expense":
			kind = models.TRANSACTION_TYPE_EXPENSE
		case "transfer":
			kind = models.TRANSACTION_TYPE_TRANSFER
		case "balance_modification":
		default:
			return models.AssistantReasonProviderUnavailable
		}
		refs = append(refs, models.AssistantReferenceDTO{Id: ref.ID, Time: at.Unix(), TimeText: at.In(time.FixedZone("", int(ref.TimezoneOffset)*60)).Format("2006-01-02 15:04:05"), Type: int(kind), CategoryName: ref.CategoryName, SourceAccountName: ref.SourceAccountName, DestinationAccountName: ref.DestinationAccountName, SourceAmountMinor: strconv.FormatInt(amount, 10), DestinationAmountMinor: destination, Currency: ref.Currency, DestinationCurrency: ref.DestinationCurrency, Comment: ref.Comment, HideAmount: ref.HideAmount})
	}
	encoded, err := json.Marshal(refs)
	if err != nil {
		return err
	}
	run.QueriesJSON = string(queries)
	run.ReferencesJSON = string(encoded)
	run.Continuation = ""
	run.ContinuationExpiresAt = 0
	run.ContinuationQueryId = ""
	run.ContinuationFilterJSON = ""
	if result.Continuation != nil {
		raw, err := json.Marshal(result.Continuation)
		if err != nil {
			return err
		}
		run.Continuation = string(raw)
		parts := strings.Split(result.Continuation.Cursor, ".")
		if len(parts) != 2 {
			return models.AssistantReasonCursorExpired
		}
		payload, err := base64.RawURLEncoding.DecodeString(parts[0])
		if err != nil {
			return models.AssistantReasonCursorExpired
		}
		var cursor assistantCursor
		if json.Unmarshal(payload, &cursor) != nil || cursor.UID != strconv.FormatInt(run.Uid, 10) || cursor.Tool != result.Continuation.Tool {
			return models.AssistantReasonCursorExpired
		}
		run.ContinuationExpiresAt = cursor.ExpiresAt * 1000
		for i := len(result.Queries) - 1; i >= 0; i-- {
			q := result.Queries[i]
			if q.Tool == result.Continuation.Tool {
				run.ContinuationQueryId = q.QueryID
				scope, err := json.Marshal(cursor.Args)
				if err != nil {
					return err
				}
				run.ContinuationFilterJSON = string(scope)
				break
			}
		}
	}
	return nil
}
