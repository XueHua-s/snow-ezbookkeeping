package services

import (
	"encoding/json"
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"math"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type assistantFilter struct {
	StartTime    *string  `json:"start_time,omitempty"`
	EndTime      *string  `json:"end_time,omitempty"`
	Types        []string `json:"types,omitempty"`
	AccountIDs   []string `json:"account_ids,omitempty"`
	CategoryIDs  []string `json:"category_ids,omitempty"`
	TagIDs       []string `json:"tag_ids,omitempty"`
	TagMatch     string   `json:"tag_match"`
	NoTags       bool     `json:"no_tags"`
	Currency     string   `json:"currency,omitempty"`
	TransferSide string   `json:"transfer_side"`
	AmountMin    *string  `json:"amount_min,omitempty"`
	AmountMax    *string  `json:"amount_max,omitempty"`
	Keyword      string   `json:"keyword,omitempty"`
	StartMillis  *int64   `json:"-"`
	EndMillis    *int64   `json:"-"`
	MinMinor     *int64   `json:"-"`
	MaxMinor     *int64   `json:"-"`
}

type assistantOrder struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

type assistantPeriod struct {
	Name        string `json:"name"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	StartMillis int64  `json:"-"`
	EndMillis   int64  `json:"-"`
}

type assistantComparison struct {
	Baseline string `json:"baseline"`
	Current  string `json:"current"`
}

type assistantQueryArgs struct {
	Filter        assistantFilter      `json:"filter"`
	Keyword       string               `json:"keyword,omitempty"`
	Type          string               `json:"type,omitempty"`
	AccountIDs    []string             `json:"account_ids,omitempty"`
	TransactionID string               `json:"transaction_id,omitempty"`
	OrderBy       assistantOrder       `json:"order_by"`
	PageSize      int                  `json:"page_size"`
	GroupBy       []string             `json:"group_by,omitempty"`
	Metrics       []string             `json:"metrics,omitempty"`
	Periods       []assistantPeriod    `json:"periods,omitempty"`
	Comparison    *assistantComparison `json:"comparison,omitempty"`
}

func containsString(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}

func canonicalStrings(a []string) []string {
	if len(a) == 0 {
		return nil
	}
	sort.Strings(a)
	out := a[:0]
	for _, v := range a {
		if len(out) == 0 || out[len(out)-1] != v {
			out = append(out, v)
		}
	}
	return out
}

func assistantID(s string) (int64, error) {
	if s == "" || len(s) > 19 || strings.Trim(s, "0123456789") != "" {
		return 0, assistantInvalid("id")
	}
	v, e := strconv.ParseInt(s, 10, 64)
	if e != nil || v <= 0 {
		return 0, assistantInvalid("id")
	}
	return v, nil
}

var assistantAmountPattern = regexp.MustCompile(`^-?[0-9]+(?:\.[0-9]{1,2})?$`)

func assistantAmount(s string) (int64, string, error) {
	if len(s) > 23 || !assistantAmountPattern.MatchString(s) {
		return 0, "", assistantInvalid("amount")
	}
	negative := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	p := strings.Split(s, ".")
	frac := "00"
	if len(p) == 2 {
		frac = p[1] + strings.Repeat("0", 2-len(p[1]))
	}
	n := new(big.Int)
	n.SetString(p[0]+frac, 10)
	if negative {
		n.Neg(n)
	}
	if !n.IsInt64() {
		return 0, "", assistantInvalid("amount")
	}
	return n.Int64(), assistantFormatAmount(n), nil
}

func assistantFormatAmount(n *big.Int) string {
	abs := new(big.Int).Abs(n)
	q, r := new(big.Int), new(big.Int)
	q.QuoRem(abs, big.NewInt(100), r)
	sign := ""
	if n.Sign() < 0 {
		sign = "-"
	}
	return fmt.Sprintf("%s%s.%02d", sign, q.String(), r.Int64())
}

func assistantTime(s string) (string, int64, error) {
	// time.Parse accepts fractional seconds even without a fractional layout.
	if strings.ContainsAny(s, ".,") {
		return "", 0, assistantInvalid("time")
	}
	t, e := time.Parse(time.RFC3339, s)
	if e != nil {
		return "", 0, assistantInvalid("time")
	}
	sec := t.Unix()
	if sec > math.MaxInt64/1000 || sec < math.MinInt64/1000 {
		return "", 0, assistantInvalid("time")
	}
	return t.UTC().Format(time.RFC3339), sec * 1000, nil
}

func normalizeAssistantArgs(name data.ToolName, m map[string]any) (assistantQueryArgs, error) {
	var a assistantQueryArgs
	raw, _ := json.Marshal(m)
	if err := json.Unmarshal(raw, &a); err != nil {
		return a, assistantInvalid("arguments")
	}
	a.PageSize = 50
	if v := m["page_size"]; v != nil {
		n, _ := v.(json.Number).Int64()
		if n < 1 || n > 200 {
			return a, assistantInvalid("page_size")
		}
		a.PageSize = int(n)
	}
	if a.Filter.TagMatch == "" {
		a.Filter.TagMatch = "any"
	}
	if a.Filter.TransferSide == "" {
		a.Filter.TransferSide = "either"
	}
	if a.OrderBy.Field == "" {
		a.OrderBy = assistantOrder{Field: "id", Direction: "asc"}
		if name == data.ToolQueryTransactions {
			a.OrderBy = assistantOrder{Field: "time", Direction: "desc"}
		}
		if name == data.ToolAggregateTransactions {
			a.OrderBy.Field = "group_key"
		}
	}
	fm, _ := m["filter"].(map[string]any)
	for _, entry := range []struct {
		m      map[string]any
		key    string
		target *[]string
		ids    bool
	}{{m, "account_ids", &a.AccountIDs, true}, {fm, "account_ids", &a.Filter.AccountIDs, true}, {fm, "category_ids", &a.Filter.CategoryIDs, true}, {fm, "tag_ids", &a.Filter.TagIDs, true}, {fm, "types", &a.Filter.Types, false}} {
		if v := entry.m[entry.key]; v != nil && len(*entry.target) == 0 {
			return a, assistantInvalid(entry.key)
		}
		if entry.ids {
			for i, s := range *entry.target {
				id, e := assistantID(s)
				if e != nil {
					return a, e
				}
				(*entry.target)[i] = strconv.FormatInt(id, 10)
			}
		}
		*entry.target = canonicalStrings(*entry.target)
	}
	if len(a.Filter.TagIDs) == 0 && fm["tag_match"] != nil {
		return a, assistantInvalid("tag_match")
	}
	if a.Filter.NoTags && len(a.Filter.TagIDs) > 0 {
		return a, assistantInvalid("no_tags")
	}
	if len(a.Filter.Types) == 0 {
		a.Filter.Types = []string{"balance_modification", "expense", "income", "transfer"}
	}
	if name == data.ToolGetTransaction {
		id, e := assistantID(a.TransactionID)
		if e != nil {
			return a, e
		}
		a.TransactionID = strconv.FormatInt(id, 10)
	}
	for _, bound := range []struct {
		s **string
		n **int64
	}{{&a.Filter.StartTime, &a.Filter.StartMillis}, {&a.Filter.EndTime, &a.Filter.EndMillis}} {
		if *bound.s != nil {
			s, n, e := assistantTime(**bound.s)
			if e != nil {
				return a, e
			}
			*bound.s = &s
			*bound.n = &n
		}
	}
	if a.Filter.StartMillis != nil && a.Filter.EndMillis != nil && *a.Filter.StartMillis >= *a.Filter.EndMillis {
		return a, assistantInvalid("end_time")
	}
	if a.Filter.Currency != "" {
		a.Filter.Currency = strings.ToUpper(a.Filter.Currency)
		if len(a.Filter.Currency) != 3 || strings.Trim(a.Filter.Currency, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "" {
			return a, assistantInvalid("currency")
		}
	}
	for _, bound := range []struct {
		s **string
		n **int64
	}{{&a.Filter.AmountMin, &a.Filter.MinMinor}, {&a.Filter.AmountMax, &a.Filter.MaxMinor}} {
		if *bound.s != nil {
			n, s, e := assistantAmount(**bound.s)
			if e != nil {
				return a, e
			}
			*bound.s = &s
			*bound.n = &n
		}
	}
	if (a.Filter.MinMinor != nil || a.Filter.MaxMinor != nil) && a.Filter.Currency == "" {
		return a, assistantInvalid("currency")
	}
	if a.Filter.MinMinor != nil && a.Filter.MaxMinor != nil && *a.Filter.MinMinor > *a.Filter.MaxMinor {
		return a, assistantInvalid("amount_max")
	}
	a.Filter.Keyword = strings.ToLower(a.Filter.Keyword)
	a.Keyword = strings.ToLower(a.Keyword)
	if a.OrderBy.Field == "amount" && (a.Filter.Currency == "" || ((len(a.Filter.Types) == 0 || containsString(a.Filter.Types, "transfer")) && a.Filter.TransferSide == "either")) {
		return a, assistantInvalid("order_by")
	}
	if name == data.ToolAggregateTransactions {
		seen := map[string]bool{}
		times := 0
		for _, d := range a.GroupBy {
			if seen[d] {
				return a, assistantInvalid("group_by")
			}
			seen[d] = true
			if containsString([]string{"day", "month", "year"}, d) {
				times++
			}
		}
		if times > 1 {
			return a, assistantInvalid("group_by")
		}
		a.GroupBy = canonicalStrings(a.GroupBy)
		if m["metrics"] == nil {
			a.Metrics = []string{"count", "sum"}
		} else if len(a.Metrics) == 0 {
			return a, assistantInvalid("metrics")
		}
		a.Metrics = canonicalStrings(a.Metrics)
		if a.OrderBy.Field != "group_key" && !containsString(a.Metrics, a.OrderBy.Field) {
			return a, assistantInvalid("order_by")
		}
		if m["periods"] != nil && len(a.Periods) == 0 {
			return a, assistantInvalid("periods")
		}
		if len(a.Periods) > 0 && (a.Filter.StartTime != nil || a.Filter.EndTime != nil) {
			return a, assistantInvalid("periods")
		}
		names := map[string]bool{}
		for i := range a.Periods {
			p := &a.Periods[i]
			if strings.TrimSpace(p.Name) == "" || names[p.Name] {
				return a, assistantInvalid("periods.name")
			}
			names[p.Name] = true
			var e error
			p.StartTime, p.StartMillis, e = assistantTime(p.StartTime)
			if e != nil {
				return a, e
			}
			p.EndTime, p.EndMillis, e = assistantTime(p.EndTime)
			if e != nil {
				return a, e
			}
			if p.StartMillis >= p.EndMillis {
				return a, assistantInvalid("periods")
			}
		}
		if a.Comparison != nil && (!names[a.Comparison.Baseline] || !names[a.Comparison.Current] || a.Comparison.Baseline == a.Comparison.Current) {
			return a, assistantInvalid("comparison")
		}
	}
	return a, nil
}

// Matching in Go gives identical literal, Unicode-aware case-insensitive
// semantics on all supported databases, including %, _ and backslash.
func assistantLiteralContains(value, keyword string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(keyword))
}
