package services

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
)

func newAssistantAggregateTools(s *AssistantQueryService) map[data.ToolName]assistantToolHandler {
	return map[data.ToolName]assistantToolHandler{data.ToolAggregateTransactions: assistantAggregateTransactions}
}

// Aggregation retains sufficient statistics per observed group, never ledger rows.
// Empty calendar buckets are virtual: pagination examines only the next page of
// each observed dimension combination, including when ordering by a metric.
type assistantAggregateStats struct {
	Count    int64
	Sum      big.Int
	Min, Max *big.Int
}

func (s *assistantAggregateStats) add(amount int64) {
	n := big.NewInt(amount)
	s.Count++
	s.Sum.Add(&s.Sum, n)
	if s.Min == nil || n.Cmp(s.Min) < 0 {
		s.Min = new(big.Int).Set(n)
	}
	if s.Max == nil || n.Cmp(s.Max) > 0 {
		s.Max = new(big.Int).Set(n)
	}
}
func (s *assistantAggregateStats) metric(metric string) *big.Int {
	switch metric {
	case "count":
		return big.NewInt(s.Count)
	case "sum":
		return &s.Sum
	case "min":
		return s.Min
	case "max":
		return s.Max
	}
	return nil
}
func (s *assistantAggregateStats) project(metrics []string) map[string]any {
	out := map[string]any{}
	for _, metric := range metrics {
		n := s.metric(metric)
		if n == nil {
			out[metric] = nil
		} else if metric == "count" {
			out[metric] = n.String()
		} else {
			out[metric] = assistantFormatAmount(n)
		}
	}
	return out
}

type assistantAggregateDimension struct {
	Currency string `json:"currency"`
	Flow     string `json:"flow"`
	Account  string `json:"account,omitempty"`
	Category string `json:"category,omitempty"`
}
type assistantAggregateGroup struct {
	Dimension assistantAggregateDimension
	Period    int
	Ordinal   int
	Stats     assistantAggregateStats
}
type assistantAggregateRange struct {
	Name       string
	Start, End *time.Time
	Count      int64
}
type assistantAggregatePosition struct {
	Key      string  `json:"key"`
	Currency string  `json:"currency"`
	Value    *string `json:"value"`
}
type assistantAggregation struct {
	e            *assistantQueryExecution
	dimension    string
	ranges       []assistantAggregateRange
	groups       map[string]*assistantAggregateGroup
	totals       map[string]*assistantAggregateGroup
	combinations map[assistantAggregateDimension]bool
	matched      int64
	additive     bool
}

// A local midnight may be missing (midnight DST) or an entire civil date may
// be skipped (a dateline change). time.Date can normalize either backwards.
// Find the first real instant on or after the requested civil date instead.
func assistantCalendarBoundary(civil time.Time, loc *time.Location) time.Time {
	target := civil.Unix() / 86400
	low, high := civil.Unix()-48*3600, civil.Unix()+48*3600
	for low < high {
		middle := low + (high-low)/2
		y, m, d := time.Unix(middle, 0).In(loc).Date()
		index := time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Unix() / 86400
		if index < target {
			low = middle + 1
		} else {
			high = middle
		}
	}
	return time.Unix(low, 0).In(loc)
}

// Calendar indexes use civil dates, so DST's 23/25-hour days remain one bucket.
func assistantCalendarIndex(t time.Time, dimension string) int {
	y, m, d := t.Date()
	switch dimension {
	case "year":
		return y
	case "month":
		return y*12 + int(m) - 1
	default:
		return int(time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Unix() / 86400)
	}
}
func (a *assistantAggregation) ordinal(p int, t time.Time) int {
	if a.dimension == "" {
		return 0
	}
	index := assistantCalendarIndex(t, a.dimension)
	if a.ranges[p].Start != nil {
		index -= assistantCalendarIndex(*a.ranges[p].Start, a.dimension)
	}
	return index
}
func (a *assistantAggregation) bucketCount(p int) int {
	r := a.ranges[p]
	if a.dimension == "" {
		return 1
	}
	if r.Start == nil || r.End == nil {
		return 0
	}
	return a.ordinal(p, r.End.Add(-time.Nanosecond)) + 1
}
func (a *assistantAggregation) key(g *assistantAggregateGroup) string {
	// Encode dimension strings to prevent delimiter collisions. Ordinal is last
	// and fixed-width, making each virtual series monotonic for binary seeking.
	return string(assistantJSON([]string{a.ranges[g.Period].Name, g.Dimension.Account, g.Dimension.Category, g.Dimension.Flow})) + fmt.Sprintf("/%012d", g.Ordinal+100000000)
}
func (a *assistantAggregation) storageKey(g *assistantAggregateGroup) string {
	return g.Dimension.Currency + "/" + a.key(g)
}
func (a *assistantAggregation) position(g *assistantAggregateGroup) assistantAggregatePosition {
	p := assistantAggregatePosition{Key: a.key(g), Currency: g.Dimension.Currency}
	if n := g.Stats.metric(a.e.Args.OrderBy.Field); n != nil {
		v := n.String()
		p.Value = &v
	}
	return p
}
func (a *assistantAggregation) compare(x, y assistantAggregatePosition) int {
	if c := strings.Compare(x.Currency, y.Currency); c != 0 {
		return c
	}
	order := a.e.Args.OrderBy
	if order.Field != "group_key" {
		c := 0
		if x.Value == nil && y.Value != nil {
			c = -1
		} else if x.Value != nil && y.Value == nil {
			c = 1
		} else if x.Value != nil {
			nx, _ := new(big.Int).SetString(*x.Value, 10)
			ny, _ := new(big.Int).SetString(*y.Value, 10)
			c = nx.Cmp(ny)
		}
		if c != 0 {
			if order.Direction == "desc" {
				c = -c
			}
			return c
		}
	}
	c := strings.Compare(x.Key, y.Key)
	if order.Field == "group_key" && order.Direction == "desc" {
		c = -c
	}
	return c
}

// RFC3339 cannot encode historical timezone offsets with seconds. Preserve
// the exact boundary instant by using UTC for those historical buckets.
func assistantAggregateTime(t time.Time) string {
	if _, offset := t.Zone(); offset%60 != 0 {
		t = t.UTC()
	}
	return t.Format(time.RFC3339)
}
func (a *assistantAggregation) project(g *assistantAggregateGroup, total bool) map[string]any {
	out := g.Stats.project(a.e.Args.Metrics)
	out["currency"] = g.Dimension.Currency
	out["flow"] = g.Dimension.Flow
	if len(a.e.Args.Periods) > 0 {
		out["period"] = a.ranges[g.Period].Name
	}
	if total {
		return out
	}
	if containsString(a.e.Args.GroupBy, "account") {
		out["account_id"] = g.Dimension.Account
		id, _ := strconv.ParseInt(g.Dimension.Account, 10, 64)
		out["account_name"] = a.e.Metadata.Accounts[id].Name
	}
	if containsString(a.e.Args.GroupBy, "category") {
		out["category_id"] = nil
		out["category_name"] = "Uncategorized"
		if g.Dimension.Category != "" {
			out["category_id"] = g.Dimension.Category
			id, _ := strconv.ParseInt(g.Dimension.Category, 10, 64)
			out["category_name"] = a.e.Metadata.Categories[id].Name
		}
	}
	if a.dimension != "" {
		start := a.bucketStart(g.Period, g.Ordinal)
		end := a.bucketStart(g.Period, g.Ordinal+1)
		r := a.ranges[g.Period]
		if r.Start != nil && start.Before(*r.Start) {
			start = *r.Start
		}
		if r.End != nil && end.After(*r.End) {
			end = *r.End
		}
		out["time_bucket"] = map[string]any{"unit": a.dimension, "start_time": assistantAggregateTime(start), "end_time": assistantAggregateTime(end), "ordinal": strconv.Itoa(g.Ordinal)}
	}
	return out
}
func (a *assistantAggregation) bucketStart(p, ordinal int) time.Time {
	if r := a.ranges[p]; r.Start != nil {
		ordinal += assistantCalendarIndex(*r.Start, a.dimension)
	}
	var civil time.Time
	switch a.dimension {
	case "year":
		civil = time.Date(ordinal, 1, 1, 0, 0, 0, 0, time.UTC)
	case "month":
		civil = time.Date(ordinal/12, time.Month(ordinal%12+1), 1, 0, 0, 0, 0, time.UTC)
	default:
		civil = time.Unix(int64(ordinal)*86400, 0).UTC()
	}
	return assistantCalendarBoundary(civil, a.e.Service.location)
}
func (a *assistantAggregation) validBucket(p, ordinal int) bool {
	return a.dimension == "" || ordinal < a.bucketCount(p) && a.bucketStart(p, ordinal).Before(a.bucketStart(p, ordinal+1))
}
func (a *assistantAggregation) scan() error {
	return a.e.ScanTransactions(nil, func(tx *assistantLogicalTransaction) error {
		matched := false
		members := 0
		for p := range a.ranges {
			r := &a.ranges[p]
			timestamp := time.UnixMilli(tx.Record.TransactionTime).In(a.e.Service.location)
			if r.Start != nil && timestamp.Before(*r.Start) || r.End != nil && !timestamp.Before(*r.End) {
				continue
			}
			matched = true
			members++
			r.Count++
			for _, side := range tx.MatchedSides {
				d := assistantAggregateDimension{Currency: side.Account.Currency, Flow: side.Flow}
				total := &assistantAggregateGroup{Period: p, Dimension: d}
				key := a.storageKey(total)
				if a.totals[key] == nil {
					a.totals[key] = total
				}
				a.totals[key].Stats.add(side.Amount)
				if containsString(a.e.Args.GroupBy, "account") {
					d.Account = strconv.FormatInt(side.Account.AccountId, 10)
				}
				if containsString(a.e.Args.GroupBy, "category") && tx.Category != nil {
					d.Category = strconv.FormatInt(tx.Category.CategoryId, 10)
				}
				g := &assistantAggregateGroup{Period: p, Dimension: d, Ordinal: a.ordinal(p, timestamp)}
				key = a.storageKey(g)
				if a.groups[key] == nil {
					a.groups[key] = g
				}
				a.groups[key].Stats.add(side.Amount)
				a.combinations[d] = true
			}
		}
		if matched {
			a.matched++
			if len(tx.MatchedSides) > 1 || members > 1 {
				a.additive = false
			}
		}
		return nil
	})
}

// Missing periods inherit only observed currency/flow/dimension combinations.
// An explicitly requested currency permits an empty ungrouped statistic; without
// that evidence an entirely empty ledger must not invent a currency.
func (a *assistantAggregation) fillEmpty() error {
	if len(a.combinations) == 0 && a.e.Args.Filter.Currency != "" && !containsString(a.e.Args.GroupBy, "account") && !containsString(a.e.Args.GroupBy, "category") {
		types := a.e.Args.Filter.Types
		if len(types) == 0 {
			types = []string{"income", "expense", "transfer", "balance_modification"}
		}
		for _, kind := range types {
			flows := []string{kind}
			if kind == "transfer" {
				flows = nil
				if a.e.Args.Filter.TransferSide != "destination" {
					flows = append(flows, "transfer_out")
				}
				if a.e.Args.Filter.TransferSide != "source" {
					flows = append(flows, "transfer_in")
				}
			}
			for _, flow := range flows {
				a.combinations[assistantAggregateDimension{Currency: a.e.Args.Filter.Currency, Flow: flow}] = true
			}
		}
	}
	flows := map[assistantAggregateDimension]bool{}
	for d := range a.combinations {
		flows[assistantAggregateDimension{Currency: d.Currency, Flow: d.Flow}] = true
	}
	for d := range flows {
		for p := range a.ranges {
			if err := a.e.Context.Err(); err != nil {
				return err
			}
			total := &assistantAggregateGroup{Period: p, Dimension: d}
			if key := a.storageKey(total); a.totals[key] == nil {
				a.totals[key] = total
			}
		}
	}
	return nil
}

func (a *assistantAggregation) page() ([]*assistantAggregateGroup, error) {
	limit := a.e.Args.PageSize + 1
	var last *assistantAggregatePosition
	if a.e.Cursor != nil {
		last = &assistantAggregatePosition{}
		if err := json.Unmarshal([]byte(a.e.Cursor.Last.Group), last); err != nil {
			return nil, assistantCursorError()
		}
	}
	page := make([]*assistantAggregateGroup, 0, limit)
	offer := func(g *assistantAggregateGroup) {
		pos := a.position(g)
		if last != nil && a.compare(pos, *last) <= 0 {
			return
		}
		i := sort.Search(len(page), func(i int) bool { return a.compare(a.position(page[i]), pos) > 0 })
		if i >= limit {
			return
		}
		page = append(page, nil)
		copy(page[i+1:], page[i:])
		page[i] = g
		if len(page) > limit {
			page = page[:limit]
		}
	}
	for _, g := range a.groups {
		if err := a.e.Context.Err(); err != nil {
			return nil, err
		}
		offer(g)
	}
	// Only observed non-time combinations participate. Each series contributes at
	// most page_size+1 empty candidates; binary search skips arbitrarily deep pages.
	for d := range a.combinations {
		for p := range a.ranges {
			count := a.bucketCount(p)
			if count == 0 {
				continue
			}
			reverse := a.e.Args.OrderBy.Field == "group_key" && a.e.Args.OrderBy.Direction == "desc"
			groupAt := func(i int) *assistantAggregateGroup {
				if reverse {
					i = count - 1 - i
				}
				return &assistantAggregateGroup{Period: p, Dimension: d, Ordinal: i}
			}
			start := 0
			if last != nil {
				start = sort.Search(count, func(i int) bool { return a.compare(a.position(groupAt(i)), *last) > 0 })
			}
			accepted := 0
			for i := start; i < count && accepted < limit; i++ {
				if err := a.e.Context.Err(); err != nil {
					return nil, err
				}
				g := groupAt(i)
				if !a.validBucket(p, g.Ordinal) {
					continue
				}
				if a.groups[a.storageKey(g)] != nil {
					continue
				}
				offer(g)
				accepted++
			}
		}
	}
	return page, nil
}

// Ratios are rounded in integer hundredths of a percent, including negative
// deltas. Neither database SUM nor binary floating point participates.
func assistantAggregatePercentage(delta, baseline *big.Int) (*string, string) {
	if baseline.Sign() <= 0 {
		return nil, "non_positive_baseline"
	}
	numerator := new(big.Int).Mul(delta, big.NewInt(10000))
	sign := numerator.Sign()
	numerator.Abs(numerator)
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, baseline, remainder)
	if remainder.Lsh(remainder, 1).Cmp(baseline) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if sign < 0 {
		quotient.Neg(quotient)
	}
	value := assistantFormatAmount(quotient)
	return &value, ""
}
func (a *assistantAggregation) comparison(g *assistantAggregateGroup) map[string]any {
	c := a.e.Args.Comparison
	if c == nil {
		return nil
	}
	baseline, current := -1, -1
	for p, r := range a.ranges {
		if r.Name == c.Baseline {
			baseline = p
		}
		if r.Name == c.Current {
			current = p
		}
	}
	if g.Period != baseline && g.Period != current {
		return nil
	}
	out := map[string]any{"group": a.project(g, false), "baseline_period": c.Baseline, "current_period": c.Current}
	values := map[string]any{}
	paired := a.validBucket(baseline, g.Ordinal) && a.validBucket(current, g.Ordinal)
	for _, metric := range a.e.Args.Metrics {
		pair := map[string]any{"baseline": nil, "current": nil, "delta": nil, "percentage": nil}
		numbers := make([]*big.Int, 2)
		for i, p := range []int{baseline, current} {
			if !a.validBucket(p, g.Ordinal) {
				continue
			}
			other := &assistantAggregateGroup{Period: p, Dimension: g.Dimension, Ordinal: g.Ordinal}
			if actual := a.groups[a.storageKey(other)]; actual != nil {
				other = actual
			}
			numbers[i] = other.Stats.metric(metric)
			if numbers[i] != nil {
				if metric == "count" {
					pair[[]string{"baseline", "current"}[i]] = numbers[i].String()
				} else {
					pair[[]string{"baseline", "current"}[i]] = assistantFormatAmount(numbers[i])
				}
			}
		}
		if !paired {
			pair["reason"] = "unmatched_time_bucket"
		} else if numbers[0] == nil || numbers[1] == nil {
			pair["reason"] = "empty_metric"
		} else {
			delta := new(big.Int).Sub(numbers[1], numbers[0])
			if metric == "count" {
				pair["delta"] = delta.String()
			} else {
				pair["delta"] = assistantFormatAmount(delta)
			}
			percentage, reason := assistantAggregatePercentage(delta, numbers[0])
			pair["percentage"] = percentage
			if reason != "" {
				pair["reason"] = reason
			}
		}
		values[metric] = pair
	}
	out["paired"] = paired
	out["metrics"] = values
	return out
}
func assistantAggregateTransactions(e *assistantQueryExecution) (data.ToolResult, error) {
	a := &assistantAggregation{e: e, groups: map[string]*assistantAggregateGroup{}, totals: map[string]*assistantAggregateGroup{}, combinations: map[assistantAggregateDimension]bool{}, additive: true}
	for _, dimension := range e.Args.GroupBy {
		if dimension == "day" || dimension == "month" || dimension == "year" {
			a.dimension = dimension
		}
	}
	if len(e.Args.Periods) > 0 {
		for _, p := range e.Args.Periods {
			start, end := time.UnixMilli(p.StartMillis).In(e.Service.location), time.UnixMilli(p.EndMillis).In(e.Service.location)
			a.ranges = append(a.ranges, assistantAggregateRange{Name: p.Name, Start: &start, End: &end})
		}
	} else {
		r := assistantAggregateRange{}
		if e.Args.Filter.StartMillis != nil {
			v := time.UnixMilli(*e.Args.Filter.StartMillis).In(e.Service.location)
			r.Start = &v
		}
		if e.Args.Filter.EndMillis != nil {
			v := time.UnixMilli(*e.Args.Filter.EndMillis).In(e.Service.location)
			r.End = &v
		}
		a.ranges = []assistantAggregateRange{r}
	}
	if err := a.scan(); err != nil {
		return data.ToolResult{}, err
	}
	if err := a.fillEmpty(); err != nil {
		return data.ToolResult{}, err
	}
	totals := make([]*assistantAggregateGroup, 0, len(a.totals))
	for _, g := range a.totals {
		totals = append(totals, g)
	}
	sort.Slice(totals, func(i, j int) bool { return a.storageKey(totals[i]) < a.storageKey(totals[j]) })
	projectedTotals := make([]map[string]any, 0, len(totals))
	for _, g := range totals {
		projectedTotals = append(projectedTotals, a.project(g, true))
	}
	result := data.ToolResult{OK: true, MatchedCount: assistantCount(a.matched), Coverage: data.QueryCoveragePage, TotalsCoverage: data.QueryCoverageAllMatching, Totals: assistantJSON(projectedTotals), CountsAdditive: &a.additive, Warnings: []string{}}
	for p, r := range a.ranges {
		if len(e.Args.Periods) > 0 {
			result.Periods = append(result.Periods, data.QueryPeriod{Name: r.Name, StartTime: e.Args.Periods[p].StartTime, EndTime: e.Args.Periods[p].EndTime, MatchedCount: assistantCount(r.Count)})
		}
	}
	page, err := a.page()
	if err != nil {
		return data.ToolResult{}, err
	}
	high, err := e.HighWater()
	if err != nil {
		return data.ToolResult{}, err
	}
	rows := make([]map[string]any, 0, e.Args.PageSize)
	comparisons := make([]map[string]any, 0, e.Args.PageSize)
	result.Data = assistantJSON(rows)
	more := false
	result.GroupsHasMore = &more
	if e.Args.Comparison != nil {
		result.Comparison = assistantJSON(comparisons)
	}
	for i, g := range page {
		if i >= e.Args.PageSize {
			break
		}
		nextRows := append(rows, a.project(g, false))
		nextComparisons := comparisons
		if comparison := a.comparison(g); comparison != nil {
			nextComparisons = append(comparisons, comparison)
		}
		candidate := result
		candidate.Data = assistantJSON(nextRows)
		candidate.ReturnedCount = len(nextRows)
		if e.Args.Comparison != nil {
			candidate.Comparison = assistantJSON(nextComparisons)
		}
		candidate.HasMore = i+1 < len(page)
		candidate.GroupsHasMore = &candidate.HasMore
		candidate.NextCursor = ""
		if candidate.HasMore {
			candidate.NextCursor, err = e.NextCursor(high, assistantPageKey{Group: string(assistantJSON(a.position(g)))})
			if err != nil {
				return data.ToolResult{}, err
			}
		}
		if !e.PageFits(candidate) {
			if len(rows) == 0 {
				return data.ToolResult{}, &data.ToolError{Code: data.ToolErrorResultTooLarge, Message: "Aggregate item and complete totals exceed the result limit"}
			}
			break
		}
		result = candidate
		rows = nextRows
		comparisons = nextComparisons
	}
	if !e.PageFits(result) {
		return data.ToolResult{}, &data.ToolError{Code: data.ToolErrorResultTooLarge, Message: "Complete aggregate totals exceed the result limit"}
	}
	return result, nil
}
