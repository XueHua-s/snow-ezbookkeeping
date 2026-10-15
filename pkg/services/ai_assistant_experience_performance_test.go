package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// Opt-in: real isolated databases, deterministic history, no model requests.
// The fixture has 1,000 conversations, 50 with 200 rounds and two answer
// versions per round. Logical storage is near 100 MiB, with nearly 2 MiB per full chat.
func TestAssistantExperienceHistoryPerformance(t *testing.T) {
	if os.Getenv("EBK_ASSISTANT_HISTORY_PERF") != "1" {
		t.Skip("set EBK_ASSISTANT_HISTORY_PERF=1 for near-quota history verification")
	}
	t.Logf("environment GOOS=%s GOARCH=%s CPUs=%d Go=%s", runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), runtime.Version())
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			storage, engine := assistantStorageFixture(t, kind)
			s := NewAssistantConversationService(storage, AssistantConversationOptions{Secret: []byte("history-performance-fixture")})
			s.now = storage.now
			ctx := core.NewNullContext()
			now := storage.now().UnixMilli()
			var contentBytes, logicalBytes int64
			for c := 1; c <= 1000; c++ {
				turns := 1
				size := 32
				if c <= 50 {
					turns = 200
					size = 4300
				}
				rows := make([]models.AssistantMessage, 0, turns*3)
				for turn := 1; turn <= turns; turn++ {
					id := int64(c*10000 + turn*3)
					q := models.AssistantMessage{Id: id, Uid: 7, ConversationId: int64(c), TurnSequence: int64(turn), Role: "user", UserMessageId: id, Content: strings.Repeat("q", 200), SelectedAssistantMessageId: id + 2, CreatedAt: now, SchemaVersion: 1}
					rows = append(rows, q)
					for v := 1; v <= 2; v++ {
						body := strings.Repeat("a", size)
						if c == 1000 {
							if v == 1 {
								body = "仅旧版本命中"
							} else {
								body = "中文稀有标记%_\\"
							}
						}
						rows = append(rows, models.AssistantMessage{Id: id + int64(v), Uid: 7, ConversationId: int64(c), TurnSequence: int64(turn), Role: "assistant", UserMessageId: id, AttemptNumber: v, Content: body, CreatedAt: now, CompletedAt: now, RunState: models.AssistantRunCompleted, SnapshotRevision: 1, SchemaVersion: 1, QueriesJSON: "[]", ReferencesJSON: "[]"})
					}
				}
				var used int64
				for i := range rows {
					m := &rows[i]
					contentBytes += int64(len(m.Content))
					var err error
					if m.Role == "user" {
						m.UsedBytes, err = AssistantStorageBytes(assistantQuestionDTO(m))
					} else {
						m.UsedBytes, err = AssistantAttemptStorageBytes(&rows[i-m.AttemptNumber], m)
						used += m.UsedBytes
					}
					if err != nil {
						t.Fatal(err)
					}
				}
				conversation := &models.AssistantConversation{Id: int64(c), Uid: 7, Title: fmt.Sprintf("History %04d", c), CreatedAt: now, ActivityAt: now, NextTurnSequence: int64(turns + 1), TurnCount: turns, UsedBytes: used, ContextVersion: 1, MetadataVersion: 1, SourceVersion: 1}
				header, err := AssistantStorageBytes(struct {
					Conversation                                       models.AssistantConversationDTO
					Summary, SummarySources, Seed, SeedSources, Frozen string
				}{Conversation: assistantConversationDTO(conversation)})
				if err != nil {
					t.Fatal(err)
				}
				conversation.HeaderUsedBytes = header
				conversation.UsedBytes += header
				logicalBytes += conversation.UsedBytes
				if conversation.UsedBytes >= 2<<20 {
					t.Fatal("fixture exceeds per-chat quota", conversation.UsedBytes)
				}
				if c == 1 {
					t.Logf("full_chat_logical_bytes=%d", conversation.UsedBytes)
				}
				if _, err := engine.Insert(conversation); err != nil {
					t.Fatal(err)
				}
				for offset := 0; offset < len(rows); offset += 30 {
					end := offset + 30
					if end > len(rows) {
						end = len(rows)
					}
					if _, err := engine.Insert(rows[offset:end]); err != nil {
						t.Fatal(err)
					}
				}
			}
			if _, err := engine.Insert(&models.AssistantUserState{Uid: 7, UsedBytes: logicalBytes, ConversationCount: 1000, Version: 1}); err != nil {
				t.Fatal(err)
			}
			t.Logf("fixture conversations=1000 full_chats=50 rounds_per_full_chat=200 versions=2 content_bytes=%d logical_bytes=%d", contentBytes, logicalBytes)
			if logicalBytes < 97<<20 || logicalBytes >= 100<<20 {
				t.Fatal("fixture outside near-quota range", logicalBytes)
			}
			// Every logical question is visited exactly once, with the selected answer.
			seen := map[string]bool{}
			listCursor := ""
			totalTurns := 0
			byteLimited := false
			for {
				list, err := s.List(ctx, 7, models.AssistantListRequest{Limit: 100, Cursor: listCursor})
				if err != nil {
					t.Fatal(err)
				}
				if len(list.Items) > 100 {
					t.Fatal("list exceeds limit")
				}
				for _, c := range list.Items {
					if seen[c.Id] {
						t.Fatal("duplicate conversation", c.Id)
					}
					seen[c.Id] = true
					before := int64(0)
					turns := map[int64]bool{}
					for {
						page, err := s.Get(ctx, 7, models.AssistantGetConversationRequest{Id: c.Id, Limit: 100, BeforeTurnSequence: before})
						if err != nil {
							t.Fatal(err)
						}
						raw, _ := json.Marshal(page)
						if len(raw) > 256<<10 || len(page.Messages) > 100 {
							t.Fatal("page exceeds dual budget", len(raw), len(page.Messages))
						}
						if page.HasMore && len(page.Messages) < 100 {
							byteLimited = true
						}
						if len(page.Messages) == 0 {
							t.Fatal("non-progressing page", c.Id)
						}
						for _, m := range page.Messages {
							if turns[m.TurnSequence] {
								t.Fatal("duplicate turn")
							}
							turns[m.TurnSequence] = true
							cid, _ := strconv.Atoi(c.Id)
							want := strconv.Itoa(cid*10000 + int(m.TurnSequence)*3 + 2)
							if m.SelectedAnswer == nil || m.SelectedAnswer.RunId != want {
								t.Fatal("wrong selected version", m.Id)
							}
						}
						if !page.HasMore {
							break
						}
						before, err = strconv.ParseInt(page.NextCursor, 10, 64)
						if err != nil {
							t.Fatal(err)
						}
					}
					want := 1
					cid, _ := strconv.Atoi(c.Id)
					if cid <= 50 {
						want = 200
					}
					if len(turns) != want {
						t.Fatal("missing turns", c.Id, len(turns), want)
					}
					for i := 1; i <= want; i++ {
						if !turns[int64(i)] {
							t.Fatal("missing turn", i)
						}
					}
					totalTurns += len(turns)
				}
				if list.NextCursor == "" {
					break
				}
				listCursor = list.NextCursor
			}
			if len(seen) != 1000 || totalTurns != 10950 || !byteLimited {
				t.Fatal("incomplete fixture walk", len(seen), totalTurns, byteLimited)
			}
			countPage, err := s.Get(ctx, 7, models.AssistantGetConversationRequest{Id: "1", Limit: 7})
			if err != nil || len(countPage.Messages) != 7 || !countPage.HasMore {
				t.Fatal("requested count limit not enforced", err, len(countPage.Messages))
			}
			if countPage.Messages[0].TurnSequence != 194 || countPage.Messages[6].TurnSequence != 200 {
				t.Fatal("first page is not chronological tail")
			}
			searchSeen := map[string]bool{}
			searchCursor := ""
			for {
				page, err := s.Search(ctx, 7, models.AssistantSearchRequest{Query: "History", Limit: 37, Cursor: searchCursor})
				if err != nil {
					t.Fatal(err)
				}
				if len(page.Items) > 37 {
					t.Fatal("search exceeds count budget")
				}
				for _, match := range page.Items {
					if searchSeen[match.ConversationId] || match.MessageId != "" {
						t.Fatal("search duplicated or returned non-title match", match)
					}
					searchSeen[match.ConversationId] = true
				}
				if page.NextCursor == "" {
					break
				}
				if page.NextCursor == searchCursor {
					t.Fatal("search cursor did not advance")
				}
				searchCursor = page.NextCursor
			}
			if len(searchSeen) != 1000 {
				t.Fatal("search omitted conversations", len(searchSeen))
			}
			measure := func(name string, budget time.Duration, call func()) {
				samples := make([]time.Duration, 20)
				for i := range samples {
					start := time.Now()
					call()
					samples[i] = time.Since(start)
				}
				sort.Slice(samples, func(i, j int) bool { return samples[i] < samples[j] })
				p95 := samples[18]
				t.Logf("%s p95=%s samples=20", name, p95)
				if p95 > budget {
					t.Errorf("%s p95 %s exceeds %s", name, p95, budget)
				}
			}
			measure("history_first_page", 300*time.Millisecond, func() {
				_, err := s.Get(ctx, 7, models.AssistantGetConversationRequest{Id: "1"})
				if err != nil {
					t.Fatal(err)
				}
			})
			measure("rare_search", time.Second, func() {
				r, err := s.Search(ctx, 7, models.AssistantSearchRequest{Query: "中文稀有标记"})
				if err != nil {
					t.Fatal(err)
				}
				if len(r.Items) != 1 || r.Items[0].ConversationId != "1000" || r.NextCursor != "" {
					t.Fatal("search omitted rare tail result", r)
				}
			})
			for _, query := range []string{"中文稀有标记", "%", "_", "\\", "仅旧版本命中"} {
				for _, versions := range []bool{false, true} {
					matches := []models.AssistantSearchMatch{}
					cursor := ""
					for pages := 0; ; pages++ {
						if pages > 1000 {
							t.Fatal("search does not terminate")
						}
						r, err := s.Search(ctx, 7, models.AssistantSearchRequest{Query: query, IncludeVersions: versions, Cursor: cursor})
						if err != nil {
							t.Fatal(err)
						}
						matches = append(matches, r.Items...)
						if r.NextCursor == "" {
							break
						}
						if r.NextCursor == cursor {
							t.Fatal("search did not advance")
						}
						cursor = r.NextCursor
					}
					want := 1
					if query == "仅旧版本命中" && !versions {
						want = 0
					}
					if len(matches) != want {
						t.Fatal("literal/version search differs", query, versions, matches)
					}
				}
			}
			runtime.GC()
			var baseline runtime.MemStats
			runtime.ReadMemStats(&baseline)
			writer := &assistantHistoryCountingWriter{baseline: baseline.HeapAlloc}
			start := time.Now()
			exportParent, cancelExport := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancelExport()
			exportContext := core.NewAssistantContext(exportParent, "history-performance-export", "en", "")
			if err := s.Export(exportContext, 7, models.AssistantExportRequest{All: true, Format: "json", IncludeVersions: true}, writer); err != nil {
				t.Fatal(err)
			}
			if elapsed := time.Since(start); elapsed >= 60*time.Second {
				t.Errorf("export exceeded default HTTP deadline: %s", elapsed)
			}
			t.Logf("all_export bytes=%d max_write=%d peak_heap_growth=%d elapsed=%s", writer.bytes, writer.maxWrite, writer.peak, time.Since(start))
			if writer.bytes < contentBytes || writer.maxWrite > 256<<10 || writer.peak > 64<<20 {
				t.Error("export is incomplete or exceeds bounded-memory envelope")
			}
		})
	}
}

type assistantHistoryCountingWriter struct {
	bytes          int64
	maxWrite       int
	baseline, peak uint64
	calls          int
}

func (w *assistantHistoryCountingWriter) Write(p []byte) (int, error) {
	w.bytes += int64(len(p))
	w.calls++
	if len(p) > w.maxWrite {
		w.maxWrite = len(p)
	}
	if w.calls%100 == 0 {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		if m.HeapAlloc > w.baseline && m.HeapAlloc-w.baseline > w.peak {
			w.peak = m.HeapAlloc - w.baseline
		}
	}
	return len(p), nil
}
