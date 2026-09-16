// api 提供 HTTP 接口：上传、队列、评分、统计等。数据实时读盘。
package api

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"smilex-deep-study/server/internal/fsrsx"
	"smilex-deep-study/server/internal/store"
)

//go:embed workflows.json
var workflowsJSON []byte

type API struct {
	Store *store.Store
}

func Register(r *gin.Engine, st *store.Store) {
	a := &API{Store: st}
	r.Use(cors())
	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
		api.GET("/validate", a.Validate)
		api.GET("/workflows", a.GetWorkflows)

		api.POST("/materials/upload", a.UploadMaterials)
		api.GET("/materials", a.GetMaterials)

		api.GET("/topics", a.GetTopics)
		api.POST("/topics/:slug/status", a.SetTopicStatus)

		api.GET("/review/queue", a.ReviewQueue)
		api.POST("/review/grade", a.ReviewGrade)
		api.POST("/review/regrade", a.ReviewRegrade)
		api.POST("/review/recall", a.ReviewRecall)

		api.GET("/notes", a.ListNotes)
		api.GET("/notes/:id", a.GetNote)

		api.POST("/cards", a.CreateCard)

		api.GET("/sessions", a.ListSessions)
		api.GET("/sessions/:id", a.GetSession)

		api.GET("/plans", a.GetPlans)
		api.GET("/plans/:slug", a.GetTopicPlan)

		api.GET("/mastery", a.GetMastery)
		api.GET("/stats", a.GetStats)
		api.GET("/prompts/:name", a.GetPrompt)
	}
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func errJSON(c *gin.Context, code int, err error) {
	c.JSON(code, gin.H{"error": err.Error()})
}

// GetPrompt 返回 prompts/ 下的通用模板原文（工作流页一键复制用）。
// 优先读工作区（数据目录上一级，启动时已由 scaffold 铺出），其次回退到进程工作目录。
func (a *API) GetPrompt(c *gin.Context) {
	name := c.Param("name")
	if name == "" || strings.ContainsRune(name, '/') || strings.ContainsRune(name, '.') {
		errJSON(c, 400, fmt.Errorf("非法 prompt 名称"))
		return
	}
	rel := filepath.Join("prompts", name+".md")
	b, err := os.ReadFile(rel)
	if abs, aerr := filepath.Abs(a.Store.DataDir); aerr == nil {
		if wb, werr := os.ReadFile(filepath.Join(filepath.Dir(abs), rel)); werr == nil {
			b = wb
			err = nil
		}
	}
	if err != nil {
		errJSON(c, 404, fmt.Errorf("prompt 不存在（prompts/%s.md 缺失；二进制启动时会在数据目录上一级自动铺出）", name))
		return
	}
	c.Data(200, "text/markdown; charset=utf-8", b)
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

// ---------- workflows ----------

func (a *API) GetWorkflows(c *gin.Context) {
	c.Data(200, "application/json; charset=utf-8", workflowsJSON)
}

// ---------- 材料 ----------

func (a *API) UploadMaterials(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		errJSON(c, 400, fmt.Errorf("需要 multipart form: %w", err))
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		if len(form.File["file"]) > 0 {
			files = form.File["file"]
		}
	}
	if len(files) == 0 {
		errJSON(c, 400, fmt.Errorf("未收到文件（字段名 files 或 file）"))
		return
	}
	var saved []string
	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			continue
		}
		name, err := a.Store.SaveInbox(fh.Filename, f)
		f.Close()
		if err != nil {
			errJSON(c, 500, err)
			return
		}
		saved = append(saved, name)
	}
	c.JSON(200, gin.H{"saved": saved})
}

func (a *API) GetMaterials(c *gin.Context) {
	inbox, err := a.Store.InboxList()
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	lib, err := a.Store.LibraryList()
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	index, err := a.Store.MaterialsIndex()
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	c.JSON(200, gin.H{"inbox": inbox, "library": lib, "index": index})
}

// ---------- 主题 ----------

func (a *API) GetTopics(c *gin.Context) {
	topics, err := a.Store.ListTopics()
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	notes, _ := a.Store.ListNotes()
	cards, _ := a.Store.ListCards()
	counts := map[string][2]int{} // [notes, cards]
	for _, n := range notes {
		t := str(n.FM["topic"])
		cnt := counts[t]
		cnt[0]++
		counts[t] = cnt
	}
	for _, card := range cards {
		t := str(card.FM["topic"])
		cnt := counts[t]
		cnt[1]++
		counts[t] = cnt
	}
	for _, t := range topics {
		slug := str(t["slug"])
		cnt := counts[slug]
		t["notes"] = cnt[0]
		t["cards"] = cnt[1]
	}
	c.JSON(200, topics)
}

type topicStatusReq struct {
	Status string `json:"status" binding:"required"`
}

func (a *API) SetTopicStatus(c *gin.Context) {
	var req topicStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		errJSON(c, 400, err)
		return
	}
	if req.Status != "active" && req.Status != "paused" {
		errJSON(c, 400, fmt.Errorf("status 只接受 active/paused: %s", req.Status))
		return
	}
	slug := c.Param("slug")
	if err := a.Store.SetTopicStatus(slug, req.Status); err != nil {
		errJSON(c, 500, err)
		return
	}
	c.JSON(200, gin.H{"ok": true, "slug": slug, "status": req.Status})
}

// ---------- 复习 ----------

func cardDue(card store.Card) time.Time {
	m := card.CardFSRSMap()
	if m == nil {
		return time.Time{} // 无 fsrs 块 → 立即到期
	}
	due, _ := time.Parse(time.RFC3339, str(m["due"]))
	return due
}

// isNewCard 判定待学新卡：无 fsrs 块，或 fsrs 块解析成功且 reps==0（从未评过分）。
func isNewCard(card store.Card) bool {
	m := card.CardFSRSMap()
	if m == nil {
		return true
	}
	st, err := fsrsx.FromMap(m)
	return err == nil && st.Reps == 0
}

func (a *API) ReviewQueue(c *gin.Context) {
	cards, err := a.Store.ListCards()
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	topic := c.Query("topic")
	mode := c.Query("mode") // learn=只新卡 / review=只已学到期卡 / 其他=混合（兼容旧调用）
	paused := a.Store.PausedTopics()
	now := time.Now()
	var due []store.Card
	for _, card := range cards {
		if paused[str(card.FM["topic"])] {
			continue
		}
		if topic != "" && str(card.FM["topic"]) != topic {
			continue
		}
		switch mode {
		case "learn":
			// 新卡不按 due 过滤，永远可学
			if isNewCard(card) {
				due = append(due, card)
			}
		case "review":
			if isNewCard(card) {
				continue
			}
			if d := cardDue(card); d.IsZero() || !d.After(now) {
				due = append(due, card)
			}
		default:
			if d := cardDue(card); d.IsZero() || !d.After(now) {
				due = append(due, card)
			}
		}
	}
	if mode == "learn" {
		// 新卡无到期意义：组内按创建时间升序，再按主题交错
		sort.SliceStable(due, func(i, j int) bool {
			return str(due[i].FM["created"]) < str(due[j].FM["created"])
		})
	} else {
		// 主题内按到期时间排序，再按主题交错（检索练习：交错优先）
		sort.Slice(due, func(i, j int) bool { return cardDue(due[i]).Before(cardDue(due[j])) })
	}
	interleaved := interleaveByTopic(due)
	out := make([]map[string]any, 0, len(interleaved))
	for _, card := range interleaved {
		m := card.FM
		m["due"] = cardDue(card)
		if fm := card.CardFSRSMap(); fm != nil {
			if st, err := fsrsx.FromMap(fm); err == nil {
				m["state_name"] = fsrsx.StateName(st.State)
			}
		} else {
			m["state_name"] = "new"
		}
		out = append(out, m)
	}
	c.JSON(200, gin.H{"count": len(out), "cards": out})
}

func interleaveByTopic(cards []store.Card) []store.Card {
	var order []string
	byTopic := map[string][]store.Card{}
	for _, card := range cards {
		t := str(card.FM["topic"])
		if _, ok := byTopic[t]; !ok {
			order = append(order, t)
		}
		byTopic[t] = append(byTopic[t], card)
	}
	out := make([]store.Card, 0, len(cards))
	for len(out) < len(cards) {
		for _, t := range order {
			if len(byTopic[t]) > 0 {
				out = append(out, byTopic[t][0])
				byTopic[t] = byTopic[t][1:]
			}
		}
	}
	return out
}

type gradeReq struct {
	ID     string `json:"id" binding:"required"`
	Rating int    `json:"rating" binding:"required"`
}

func (a *API) ReviewGrade(c *gin.Context) {
	var req gradeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		errJSON(c, 400, err)
		return
	}
	card, err := a.Store.GetCard(req.ID)
	if err != nil {
		errJSON(c, 404, fmt.Errorf("卡片不存在: %s", req.ID))
		return
	}
	now := time.Now()
	var state *fsrsx.State
	if m := card.CardFSRSMap(); m != nil {
		state, err = fsrsx.FromMap(m)
		if err != nil {
			errJSON(c, 422, fmt.Errorf("卡片 %s 的 fsrs 块无法解析: %w", req.ID, err))
			return
		}
	} else {
		state = fsrsx.New(now)
	}
	before := state.State
	fsrsBefore := state.ToMap()
	next, err := fsrsx.Grade(state, req.Rating, now)
	if err != nil {
		errJSON(c, 400, err)
		return
	}
	if err := a.Store.WriteCardFSRS(req.ID, next.ToMap()); err != nil {
		errJSON(c, 500, err)
		return
	}
	_ = a.Store.AppendJSONL("progress/review-log.jsonl", store.ReviewLogEntry{
		Card: req.ID, Rating: req.Rating, TS: now.UTC().Format(time.RFC3339),
		StateBefore: before, StateAfter: next.State, FSRSBefore: fsrsBefore,
	})
	c.JSON(200, gin.H{"id": req.ID, "fsrs": next.ToMap(), "state_name": fsrsx.StateName(next.State)})
}

// ReviewRegrade 改判：恢复该卡最近一次评分的评分前状态，按新档位重算。
// 日志只追加：先追加作废行（void+void_of），再追加新评分行（带新 fsrs_before，可连续改判）。
func (a *API) ReviewRegrade(c *gin.Context) {
	var req gradeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		errJSON(c, 400, err)
		return
	}
	if _, err := a.Store.GetCard(req.ID); err != nil {
		errJSON(c, 404, fmt.Errorf("卡片不存在: %s", req.ID))
		return
	}
	log, err := a.Store.ReadReviewLog()
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	voided := voidedSet(log)
	// 找该卡最近一条有效评分记录（rating 1-4 且未被作废）
	last := -1
	for i := len(log) - 1; i >= 0; i-- {
		e := log[i]
		if e.Card == req.ID && e.Rating >= 1 && e.Rating <= 4 && !voided[e.TS] {
			last = i
			break
		}
	}
	if last < 0 {
		errJSON(c, 409, fmt.Errorf("卡片 %s 没有可改判的评分记录", req.ID))
		return
	}
	prev := log[last]
	if prev.FSRSBefore == nil {
		errJSON(c, 409, fmt.Errorf("该评分记录缺少 fsrs_before，不支持改判"))
		return
	}
	state, err := fsrsx.FromMap(prev.FSRSBefore)
	if err != nil {
		errJSON(c, 422, fmt.Errorf("评分前状态无法解析: %w", err))
		return
	}
	now := time.Now()
	before := state.State
	fsrsBefore := state.ToMap()
	next, err := fsrsx.Grade(state, req.Rating, now)
	if err != nil {
		errJSON(c, 400, err)
		return
	}
	if err := a.Store.WriteCardFSRS(req.ID, next.ToMap()); err != nil {
		errJSON(c, 500, err)
		return
	}
	nowTS := now.UTC().Format(time.RFC3339)
	_ = a.Store.AppendJSONL("progress/review-log.jsonl", store.ReviewLogEntry{
		Card: req.ID, TS: nowTS, Void: true, VoidOf: prev.TS,
	})
	_ = a.Store.AppendJSONL("progress/review-log.jsonl", store.ReviewLogEntry{
		Card: req.ID, Rating: req.Rating, TS: nowTS,
		StateBefore: before, StateAfter: next.State, FSRSBefore: fsrsBefore,
	})
	c.JSON(200, gin.H{"id": req.ID, "fsrs": next.ToMap(), "state_name": fsrsx.StateName(next.State)})
}

type recallReq struct {
	ID     string `json:"id" binding:"required"`
	Answer string `json:"answer" binding:"required"`
}

func (a *API) ReviewRecall(c *gin.Context) {
	var req recallReq
	if err := c.ShouldBindJSON(&req); err != nil {
		errJSON(c, 400, err)
		return
	}
	if _, err := a.Store.GetCard(req.ID); err != nil {
		errJSON(c, 404, fmt.Errorf("卡片不存在: %s", req.ID))
		return
	}
	err := a.Store.AppendJSONL("progress/recall-log.jsonl", map[string]any{
		"card": req.ID, "answer": req.Answer, "ts": time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	c.JSON(200, gin.H{"ok": true, "hint": "答案已记录，将在下次自测/诊断时由 harness 批改"})
}

// ---------- 笔记 ----------

// orderOf 提取笔记 frontmatter 的 order（YAML 可能解析为 int 或 float64）。
func orderOf(fm map[string]any) (int, bool) {
	switch v := fm["order"].(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	}
	return 0, false
}

func (a *API) ListNotes(c *gin.Context) {
	notes, err := a.Store.ListNotes()
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	cards, _ := a.Store.ListCards()
	cardCount := map[string]int{}
	for _, card := range cards {
		cardCount[str(card.FM["note"])]++
	}
	// 排序：topic 字母序优先；同 topic 内有 order 的升序在前，无 order 的排末尾按 id 字母序
	sort.SliceStable(notes, func(i, j int) bool {
		ti, tj := str(notes[i].FM["topic"]), str(notes[j].FM["topic"])
		if ti != tj {
			return ti < tj
		}
		oi, okI := orderOf(notes[i].FM)
		oj, okJ := orderOf(notes[j].FM)
		if okI != okJ {
			return okI
		}
		if okI && oi != oj {
			return oi < oj
		}
		return notes[i].ID < notes[j].ID
	})
	out := make([]map[string]any, 0, len(notes))
	for _, n := range notes {
		m := map[string]any{
			"id": n.ID, "title": str(n.FM["title"]), "topic": str(n.FM["topic"]),
			"tags": n.FM["tags"], "links": n.FM["links"], "created": str(n.FM["created"]),
			"cards": cardCount[n.ID],
			"gaps":  strings.Count(n.Body, "] "),
		}
		if o, ok := orderOf(n.FM); ok {
			m["order"] = o
		}
		out = append(out, m)
	}
	c.JSON(200, out)
}

func (a *API) GetNote(c *gin.Context) {
	id := c.Param("id")
	n, err := a.Store.GetNote(id)
	if err != nil {
		errJSON(c, 404, fmt.Errorf("笔记不存在: %s", id))
		return
	}
	notes, _ := a.Store.ListNotes()
	var backlinks []string
	for _, other := range notes {
		if other.ID == id {
			continue
		}
		if links, ok := other.FM["links"].([]any); ok {
			for _, l := range links {
				if str(l) == id {
					backlinks = append(backlinks, other.ID)
				}
			}
		}
	}
	cards, _ := a.Store.ListCards()
	var myCards []map[string]any
	for _, card := range cards {
		if str(card.FM["note"]) == id {
			m := card.FM
			m["due"] = cardDue(card)
			myCards = append(myCards, m)
		}
	}
	c.JSON(200, gin.H{"id": n.ID, "fm": n.FM, "body": n.Body,
		"backlinks": backlinks, "cards": myCards})
}

// ---------- 卡片 ----------

type createCardReq struct {
	Front string `json:"front" binding:"required"`
	Back  string `json:"back" binding:"required"`
	Topic string `json:"topic"`
	Note  string `json:"note"`
	Type  string `json:"type"`
}

func (a *API) CreateCard(c *gin.Context) {
	var req createCardReq
	if err := c.ShouldBindJSON(&req); err != nil {
		errJSON(c, 400, err)
		return
	}
	zero := fsrsx.New(time.Now()).ToMap()
	id, err := a.Store.CreateCard(store.CreateCardParams{
		Front: req.Front, Back: req.Back, Topic: req.Topic, Note: req.Note, Type: req.Type,
	}, zero)
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	c.JSON(200, gin.H{"id": id})
}

// ---------- 会话 ----------

func (a *API) ListSessions(c *gin.Context) {
	sessions, err := a.Store.ListSessions()
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	out := make([]map[string]any, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, map[string]any{
			"id": s.ID, "type": str(s.FM["type"]), "topic": str(s.FM["topic"]),
			"date": str(s.FM["date"]), "tool": str(s.FM["tool"]), "summary": str(s.FM["summary"]),
		})
	}
	c.JSON(200, out)
}

func (a *API) GetSession(c *gin.Context) {
	s, err := a.Store.GetSession(c.Param("id"))
	if err != nil {
		errJSON(c, 404, fmt.Errorf("会话不存在"))
		return
	}
	c.JSON(200, gin.H{"id": s.ID, "fm": s.FM, "body": s.Body})
}

// ---------- 掌握度与统计 ----------

// voidedSet 收集被作废评分记录的 ts（作废行 void_of 指向被作废记录的 ts）。
func voidedSet(log []store.ReviewLogEntry) map[string]bool {
	m := map[string]bool{}
	for _, e := range log {
		if e.Void && e.VoidOf != "" {
			m[e.VoidOf] = true
		}
	}
	return m
}

func (a *API) GetMastery(c *gin.Context) {
	mastery, err := a.Store.ReadMastery()
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	cards, _ := a.Store.ListCards()
	log, _ := a.Store.ReadReviewLog()
	voided := voidedSet(log)
	paused := a.Store.PausedTopics()
	cardTopic := map[string]string{}
	perTopic := map[string]map[string]any{}
	ensure := func(t string) map[string]any {
		if perTopic[t] == nil {
			perTopic[t] = map[string]any{"cards": 0, "due": 0, "new": 0, "reviews": 0, "again": 0}
		}
		return perTopic[t]
	}
	now := time.Now()
	for _, card := range cards {
		t := str(card.FM["topic"])
		cardTopic[card.ID] = t
		if paused[t] {
			continue
		}
		m := ensure(t)
		m["cards"] = m["cards"].(int) + 1
		if isNewCard(card) {
			m["new"] = m["new"].(int) + 1
			continue
		}
		// due 只算已学到期卡（口径与 stats 一致）
		d := cardDue(card)
		if d.IsZero() || !d.After(now) {
			m["due"] = m["due"].(int) + 1
		}
	}
	for _, e := range log {
		if e.Void || voided[e.TS] {
			continue
		}
		t := cardTopic[e.Card]
		if paused[t] {
			continue
		}
		m := ensure(t)
		m["reviews"] = m["reviews"].(int) + 1
		if e.Rating == 1 {
			m["again"] = m["again"].(int) + 1
		}
	}
	c.JSON(200, gin.H{"mastery": mastery, "per_topic": perTopic})
}

func (a *API) GetStats(c *gin.Context) {
	cards, err := a.Store.ListCards()
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	log, err := a.Store.ReadReviewLog()
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	mastery, _ := a.Store.ReadMastery()
	sessions, _ := a.Store.ListSessions()
	voided := voidedSet(log)
	paused := a.Store.PausedTopics()
	now := time.Now()
	today := now.Format("2006-01-02")

	due, newCards, reviewsToday, learnedToday := 0, 0, 0, 0
	dayCount := map[string]int{}
	dayLearned := map[string]int{}
	for _, card := range cards {
		if paused[str(card.FM["topic"])] {
			continue
		}
		if isNewCard(card) {
			newCards++
			continue
		}
		// due_now 只算已学到期卡；待学新卡由 new_cards 单独计数
		d := cardDue(card)
		if d.IsZero() || !d.After(now) {
			due++
		}
	}
	for _, e := range log {
		if e.Void || voided[e.TS] {
			continue
		}
		if t, err := time.Parse(time.RFC3339, e.TS); err == nil {
			key := t.Local().Format("2006-01-02")
			dayCount[key]++
			if key == today {
				reviewsToday++
			}
			// StateBefore==0（New 态）的首次评分计为新学；
			// 限 rating 1-4，跳过 W5 批改追加的 {"graded":true} 行（无 rating，解析为 0）
			if e.StateBefore == 0 && e.Rating >= 1 && e.Rating <= 4 {
				dayLearned[key]++
				if key == today {
					learnedToday++
				}
			}
		}
	}
	// 连续天数：从今天（无则从昨天）往回数有复习记录的日子
	streak := 0
	day := now
	if dayCount[today] == 0 {
		day = now.AddDate(0, 0, -1)
	}
	for {
		if dayCount[day.Format("2006-01-02")] == 0 {
			break
		}
		streak++
		day = day.AddDate(0, 0, -1)
	}
	// 90 天热力图
	var heatmap []map[string]any
	for i := 89; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		heatmap = append(heatmap, map[string]any{"date": d, "count": dayCount[d], "learned": dayLearned[d]})
	}
	// 最近会话
	recent := []map[string]any{}
	for i, s := range sessions {
		if i >= 5 {
			break
		}
		recent = append(recent, map[string]any{
			"id": s.ID, "type": str(s.FM["type"]), "topic": str(s.FM["topic"]),
			"date": str(s.FM["date"]), "summary": str(s.FM["summary"]),
		})
	}
	// 掌握度摘要
	masterySummary := map[string]any{}
	for k, v := range mastery {
		if m, ok := v.(map[string]any); ok {
			masterySummary[k] = map[string]any{"level": m["level"], "updated": m["updated"]}
		}
	}
	c.JSON(200, gin.H{
		"total_cards": len(cards), "due_now": due, "new_cards": newCards,
		"reviews_today": reviewsToday, "learned_today": learnedToday, "streak": streak,
		"heatmap": heatmap, "recent_sessions": recent, "mastery": masterySummary,
	})
}
