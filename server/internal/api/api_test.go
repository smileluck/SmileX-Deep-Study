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
	Register(r, st, "dev")
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

// writeMistakeFile 直接落盘一条契约格式的错题。
func writeMistakeFile(t *testing.T, st *store.Store, id, topic, status, masteredAt string) {
	t.Helper()
	if err := os.MkdirAll(st.MistakesDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	content := fmt.Sprintf(`---
id: %s
topic: %s
source: quiz
session: 20261003-quiz-t1
question: 什么是稳定性？
answer: |
  R 衰减到阈值的天数。
my_answer: 不知道
analysis: 概念没记住
status: %s
created: 2026-10-03
mastered_at: %s
---
`, id, topic, status, masteredAt)
	if err := os.WriteFile(filepath.Join(st.MistakesDir(), id+".md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestListMistakesEmpty(t *testing.T) {
	r, _ := setup(t)
	code, resp := doJSON(t, r, http.MethodGet, "/api/mistakes", nil)
	if code != 200 {
		t.Fatalf("status = %d, body = %s", code, resp)
	}
	if !strings.Contains(string(resp), `"mistakes":[]`) {
		t.Errorf("空错题库应返回空数组而非 null: %s", resp)
	}
	var out struct {
		Mistakes []any `json:"mistakes"`
		Summary  struct {
			Active   int `json:"active"`
			Mastered int `json:"mastered"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(resp, &out); err != nil {
		t.Fatal(err)
	}
	if out.Summary.Active != 0 || out.Summary.Mastered != 0 {
		t.Errorf("summary = %+v, want 全零", out.Summary)
	}
}

func TestListMistakesSummary(t *testing.T) {
	r, st := setup(t)
	writeManifest(t, st, "t1", "")
	writeManifest(t, st, "t2", "")
	writeMistakeFile(t, st, "m-1", "t1", "active", "null")
	writeMistakeFile(t, st, "m-2", "t1", "mastered", "2026-10-02")
	writeMistakeFile(t, st, "m-3", "t2", "active", "null")

	code, resp := doJSON(t, r, http.MethodGet, "/api/mistakes", nil)
	if code != 200 {
		t.Fatalf("status = %d, body = %s", code, resp)
	}
	var out struct {
		Mistakes []map[string]any `json:"mistakes"`
		Summary  struct {
			Active   int `json:"active"`
			Mastered int `json:"mastered"`
			ByTopic  map[string]struct {
				Active   int `json:"active"`
				Mastered int `json:"mastered"`
			} `json:"by_topic"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(resp, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Mistakes) != 3 {
		t.Fatalf("len(mistakes) = %d, want 3", len(out.Mistakes))
	}
	if out.Summary.Active != 2 || out.Summary.Mastered != 1 {
		t.Errorf("summary = %+v, want active=2 mastered=1", out.Summary)
	}
	if out.Summary.ByTopic["t1"].Active != 1 || out.Summary.ByTopic["t1"].Mastered != 1 {
		t.Errorf("by_topic[t1] = %+v, want active=1 mastered=1", out.Summary.ByTopic["t1"])
	}
	if out.Summary.ByTopic["t2"].Active != 1 {
		t.Errorf("by_topic[t2] = %+v, want active=1", out.Summary.ByTopic["t2"])
	}
	for _, m := range out.Mistakes {
		for _, k := range []string{"id", "topic", "source", "session", "question", "answer", "status", "created"} {
			if _, ok := m[k]; !ok {
				t.Errorf("条目缺少字段 %s: %+v", k, m)
			}
		}
	}
}

func TestSetMistakeStatusAPI(t *testing.T) {
	r, st := setup(t)
	writeManifest(t, st, "t1", "")
	writeMistakeFile(t, st, "m-1", "t1", "active", "null")

	code, resp := doJSON(t, r, http.MethodPost, "/api/mistakes/m-1/status",
		map[string]any{"status": "mastered"})
	if code != 200 {
		t.Fatalf("status = %d, body = %s", code, resp)
	}
	var entry map[string]any
	if err := json.Unmarshal(resp, &entry); err != nil {
		t.Fatal(err)
	}
	if entry["status"] != "mastered" {
		t.Errorf("返回条目 status = %v", entry["status"])
	}
	today := time.Now().Format("2006-01-02")
	if entry["mastered_at"] != today {
		t.Errorf("mastered_at = %v, want %s", entry["mastered_at"], today)
	}

	code, resp = doJSON(t, r, http.MethodPost, "/api/mistakes/m-1/status",
		map[string]any{"status": "active"})
	if code != 200 {
		t.Fatalf("status = %d, body = %s", code, resp)
	}
	if err := json.Unmarshal(resp, &entry); err != nil {
		t.Fatal(err)
	}
	if entry["mastered_at"] != nil {
		t.Errorf("恢复 active 后 mastered_at = %v, want null", entry["mastered_at"])
	}

	code, _ = doJSON(t, r, http.MethodPost, "/api/mistakes/m-1/status",
		map[string]any{"status": "bogus"})
	if code != 400 {
		t.Errorf("非法 status 应 400, got %d", code)
	}
	code, _ = doJSON(t, r, http.MethodPost, "/api/mistakes/ghost/status",
		map[string]any{"status": "mastered"})
	if code != 404 {
		t.Errorf("不存在的错题应 404, got %d", code)
	}
}

func TestStatsActiveMistakes(t *testing.T) {
	r, st := setup(t)
	writeManifest(t, st, "t1", "")
	writeMistakeFile(t, st, "m-1", "t1", "active", "null")
	writeMistakeFile(t, st, "m-2", "t1", "mastered", "2026-10-02")

	code, resp := doJSON(t, r, http.MethodGet, "/api/stats", nil)
	if code != 200 {
		t.Fatalf("stats status = %d", code)
	}
	var stats struct {
		ActiveMistakes int `json:"active_mistakes"`
	}
	if err := json.Unmarshal(resp, &stats); err != nil {
		t.Fatal(err)
	}
	if stats.ActiveMistakes != 1 {
		t.Errorf("active_mistakes = %d, want 1", stats.ActiveMistakes)
	}
}

func TestValidateMistakes(t *testing.T) {
	r, st := setup(t)
	writeManifest(t, st, "t1", "")
	if err := os.MkdirAll(st.MistakesDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	bad := `---
id: bad-mistake
topic: ghost-topic
source: typo
status: bogus
created: 10/03/2026
---
`
	if err := os.WriteFile(filepath.Join(st.MistakesDir(), "bad-mistake.md"), []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}
	writeMistakeFile(t, st, "ok-mistake", "t1", "active", "null")

	code, resp := doJSON(t, r, http.MethodGet, "/api/validate", nil)
	if code != 200 {
		t.Fatalf("validate status = %d", code)
	}
	var out validateResp
	if err := json.Unmarshal(resp, &out); err != nil {
		t.Fatal(err)
	}
	if out.OK {
		t.Fatal("违规错题应使 validate 失败")
	}
	if out.Checked["mistakes"] != 2 {
		t.Errorf("checked[mistakes] = %d, want 2", out.Checked["mistakes"])
	}
	var issues []string
	for _, e := range out.Errors {
		if e.File == "mistakes/bad-mistake.md" {
			issues = append(issues, e.Issues...)
		}
	}
	joined := strings.Join(issues, ";")
	for _, want := range []string{"topic", "source", "status", "created", "question"} {
		if !strings.Contains(joined, want) {
			t.Errorf("报错中应包含 %q 问题: %s", want, joined)
		}
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

func TestGetMistakeAsset(t *testing.T) {
	r, st := setup(t)
	if err := os.MkdirAll(st.MistakeAssetsDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	png := []byte{0x89, 'P', 'N', 'G'}
	if err := os.WriteFile(filepath.Join(st.MistakeAssetsDir(), "m-1.png"), png, 0o644); err != nil {
		t.Fatal(err)
	}

	code, resp := doJSON(t, r, http.MethodGet, "/api/mistakes/asset?name=m-1.png", nil)
	if code != 200 {
		t.Fatalf("asset status = %d, body = %s", code, resp)
	}
	if !bytes.Equal(resp, png) {
		t.Errorf("asset 内容不一致: %q", resp)
	}

	code, _ = doJSON(t, r, http.MethodGet, "/api/mistakes/asset?name=..%2Fm-1.md", nil)
	if code != 400 {
		t.Errorf("路径穿越应 400, got %d", code)
	}
	code, _ = doJSON(t, r, http.MethodGet, "/api/mistakes/asset?name=ghost.png", nil)
	if code != 404 {
		t.Errorf("不存在应 404, got %d", code)
	}
	code, _ = doJSON(t, r, http.MethodGet, "/api/mistakes/asset", nil)
	if code != 400 {
		t.Errorf("缺 name 应 400, got %d", code)
	}
}

func TestValidateMistakeImage(t *testing.T) {
	r, st := setup(t)
	writeManifest(t, st, "t1", "")
	content := `---
id: m-img
topic: t1
source: image
session: ""
question: 图中第 3 题
answer: |
  42
status: active
created: 2026-10-03
mastered_at: null
image: assets/m-img.png
---
`
	if err := os.MkdirAll(st.MistakesDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(st.MistakesDir(), "m-img.md"), []byte(content), 0o644); err != nil {
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
	found := false
	for _, e := range out.Errors {
		if e.File == "mistakes/m-img.md" {
			for _, issue := range e.Issues {
				if strings.Contains(issue, "image") {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatalf("image 指向缺失原图应报错: %+v", out.Errors)
	}

	if err := os.MkdirAll(st.MistakeAssetsDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(st.MistakeAssetsDir(), "m-img.png"), []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, resp = doJSON(t, r, http.MethodGet, "/api/validate", nil)
	out = validateResp{}
	if err := json.Unmarshal(resp, &out); err != nil {
		t.Fatal(err)
	}
	if !out.OK {
		t.Errorf("补齐原图后 validate 应通过: %+v", out.Errors)
	}
}
