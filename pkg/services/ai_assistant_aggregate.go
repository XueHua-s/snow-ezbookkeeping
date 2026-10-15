package services

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"math/big"
	"strings"
	"time"
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
