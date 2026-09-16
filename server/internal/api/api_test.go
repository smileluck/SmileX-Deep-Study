package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"smilex-deep-study/server/internal/fsrsx"
	"smilex-deep-study/server/internal/store"
)

func setup(t *testing.T) (*gin.Engine, *store.Store) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	st := store.New(t.TempDir())
	if err := st.Ensure(); err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	Register(r, st)
	return r, st
}

func writeManifest(t *testing.T, st *store.Store, slug, status string) {
	t.Helper()
	dir := filepath.Join(st.TopicsDir(), slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	m := map[string]any{"slug": slug, "name": slug}
	if status != "" {
		m["status"] = status
	}
	b, _ := json.Marshal(m)
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeCardFile 直接落盘一张契约格式的卡片（全零 fsrs 块，due 带引号）。
func writeCardFile(t *testing.T, st *store.Store, id, topic string) {
	t.Helper()
	content := fmt.Sprintf(`---
id: %s
note: ""
topic: %s
type: basic
front: q
back: a
hint: 这是一个十五字以上的回忆抓手提示
created: 2026-09-16
fsrs:
  due: "2026-09-16T00:00:00Z"
  stability: 0
  difficulty: 0
  elapsed_days: 0
  scheduled_days: 0
  reps: 0
  lapses: 0
  state: 0
  last_review: null
---
`, id, topic)
	if err := os.WriteFile(filepath.Join(st.CardsDir(), id+".md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func doJSON(t *testing.T, r *gin.Engine, method, path string, body any) (int, []byte) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code, w.Body.Bytes()
}

func TestReviewGradeAppendsLogAndUpdatesCard(t *testing.T) {
	r, st := setup(t)
	writeManifest(t, st, "t1", "")
	writeCardFile(t, st, "card-a", "t1")

	code, resp := doJSON(t, r, http.MethodPost, "/api/review/grade",
		map[string]any{"id": "card-a", "rating": 3})
	if code != 200 {
		t.Fatalf("grade status = %d, body = %s", code, resp)
	}
	log, bad, err := st.ReadReviewLog()
	if err != nil || bad != 0 {
		t.Fatalf("ReadReviewLog: err=%v bad=%d", err, bad)
	}
	if len(log) != 1 {
		t.Fatalf("review-log 行数 = %d, want 1", len(log))
	}
	e := log[0]
	if e.Card != "card-a" || e.Rating != 3 || e.FSRSBefore == nil {
		t.Errorf("日志条目不符: %+v", e)
	}
	card, err := st.GetCard("card-a")
	if err != nil {
		t.Fatal(err)
	}
	st2, err := fsrsx.FromMap(card.CardFSRSMap())
	if err != nil {
		t.Fatal(err)
	}
	if st2.Reps != 1 {
		t.Errorf("评分后 reps = %d, want 1", st2.Reps)
	}
}

func TestRegradeVoidsOnlyTargetCardRecord(t *testing.T) {
	r, st := setup(t)
	writeManifest(t, st, "t1", "")
	writeCardFile(t, st, "card-a", "t1")
	writeCardFile(t, st, "card-b", "t1")

	// 两张卡在同一秒各有一条评分记录
	ts := "2026-09-16T08:00:00Z"
	before := fsrsx.New(time.Date(2026, 9, 16, 7, 0, 0, 0, time.UTC)).ToMap()
	for _, id := range []string{"card-a", "card-b"} {
		if err := st.AppendJSONL("progress/review-log.jsonl", store.ReviewLogEntry{
			Card: id, Rating: 3, TS: ts, StateBefore: 0, StateAfter: 1, FSRSBefore: before,
		}); err != nil {
			t.Fatal(err)
		}
	}

	code, resp := doJSON(t, r, http.MethodPost, "/api/review/regrade",
		map[string]any{"id": "card-a", "rating": 4})
	if code != 200 {
		t.Fatalf("regrade status = %d, body = %s", code, resp)
	}
	log, _, err := st.ReadReviewLog()
	if err != nil {
		t.Fatal(err)
	}
	voided := voidedSet(log)
	if !voided[voidKey{"card-a", ts}] {
		t.Error("card-a 的同秒记录应被作废")
	}
	if voided[voidKey{"card-b", ts}] {
		t.Error("card-b 的同秒记录被误伤作废")
	}
	// 有效评分记录数：card-b 1 条 + card-a 新评分 1 条
	valid := 0
	for _, e := range log {
		if !e.Void && e.Rating >= 1 && !voided[voidKey{e.Card, e.TS}] {
			valid++
		}
	}
	if valid != 2 {
		t.Errorf("有效评分记录 = %d, want 2", valid)
	}
}

func TestReviewQueueExcludesArchivedAndPaused(t *testing.T) {
	r, st := setup(t)
	writeManifest(t, st, "active-x", "")
	writeManifest(t, st, "paused-x", "paused")
	writeCardFile(t, st, "c-active", "active-x")
	writeCardFile(t, st, "c-paused", "paused-x")
	writeCardFile(t, st, "c-archived", "_archived")

	code, resp := doJSON(t, r, http.MethodGet, "/api/review/queue", nil)
	if code != 200 {
		t.Fatalf("queue status = %d", code)
	}
	var q struct {
		Count int `json:"count"`
		Cards []struct {
			ID string `json:"id"`
		} `json:"cards"`
	}
	if err := json.Unmarshal(resp, &q); err != nil {
		t.Fatal(err)
	}
	if q.Count != 1 || q.Cards[0].ID != "c-active" {
		t.Errorf("queue = %+v, want 只有 c-active", q)
	}

	code, resp = doJSON(t, r, http.MethodGet, "/api/stats", nil)
	if code != 200 {
		t.Fatalf("stats status = %d", code)
	}
	var stats struct {
		TotalCards int `json:"total_cards"`
	}
	if err := json.Unmarshal(resp, &stats); err != nil {
		t.Fatal(err)
	}
	if stats.TotalCards != 1 {
		t.Errorf("total_cards = %d, want 1（排除 paused 与 _archived）", stats.TotalCards)
	}
}

func TestValidateCardMissingHint(t *testing.T) {
	r, st := setup(t)
	writeManifest(t, st, "t1", "")
	content := `---
id: nohint
note: ""
topic: t1
type: basic
front: q
back: a
created: 2026-09-16
fsrs:
  due: "2026-09-16T00:00:00Z"
  stability: 0
  difficulty: 0
  elapsed_days: 0
  scheduled_days: 0
  reps: 0
  lapses: 0
  state: 0
  last_review: null
---
`
	if err := os.WriteFile(filepath.Join(st.CardsDir(), "nohint.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	code, resp := doJSON(t, r, http.MethodGet, "/api/validate", nil)
	if code != 200 {
		t.Fatalf("validate status = %d", code)
	}
	var out validateResp
	if err := json.Unmarshal(resp, &out); err != nil {
		t.Fatal(err)
	}
	if out.OK {
		t.Fatal("缺 hint 的卡片应使 validate 失败")
	}
	found := false
	for _, e := range out.Errors {
		if e.File == "cards/nohint.md" {
			for _, issue := range e.Issues {
				if strings.Contains(issue, "hint") {
					found = true
				}
			}
		}
	}
	if !found {
		t.Errorf("validate errors 中未找到 nohint 卡片的 hint 问题: %+v", out.Errors)
	}
}
