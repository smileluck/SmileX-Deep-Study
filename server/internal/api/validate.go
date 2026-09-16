// validate 是数据契约的硬校验：扫描 data/ 全部资产，
// 返回 schema 违规清单。agent 工作流收尾必须把它清零。
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"smilex-deep-study/server/internal/fsrsx"
	"smilex-deep-study/server/internal/store"
)

type validationIssue struct {
	File   string   `json:"file"`
	Kind   string   `json:"kind"`
	Issues []string `json:"issues"`
}

type validateResp struct {
	OK       bool              `json:"ok"`
	Errors   []validationIssue `json:"errors"`
	Warnings []validationIssue `json:"warnings"`
	Checked  map[string]int    `json:"checked"`
}

var sessionTypes = map[string]bool{
	"tutor": true, "feynman": true, "quiz": true, "diagnose": true, "import": true, "plan": true, "merge": true,
}

var evidenceKinds = map[string]bool{
	"quiz": true, "review": true, "feynman": true, "diagnose": true, "import": true, "merge": true,
}

var cardTypes = map[string]bool{"basic": true, "cloze": true}

func isDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

// validator 一次校验的上下文：累计 errors/warnings/checked，
// 并缓存跨资产交叉引用所需的集合（主题 slug、笔记 id）。
type validator struct {
	st      *store.Store
	errs    []validationIssue
	warns   []validationIssue
	checked map[string]int
	topics  map[string]bool // 有 manifest 的主题 slug
	noteIDs map[string]bool // 全部笔记 id（含解析失败的文件）
}

func (v *validator) err(file, kind string, issues []string) {
	v.errs = append(v.errs, validationIssue{File: file, Kind: kind, Issues: issues})
}

func (v *validator) warn(file, kind string, issues ...string) {
	v.warns = append(v.warns, validationIssue{File: file, Kind: kind, Issues: issues})
}

func (a *API) Validate(c *gin.Context) {
	v := &validator{
		st:   a.Store,
		errs: []validationIssue{}, warns: []validationIssue{},
		checked: map[string]int{"cards": 0, "notes": 0, "sessions": 0, "mastery_topics": 0, "materials": 0, "plans": 0, "manifests": 0},
		topics:  map[string]bool{},
		noteIDs: map[string]bool{},
	}
	// 顺序有依赖：主题/笔记先验，卡片交叉引用其结果
	v.validateManifests()
	v.validateNotes()
	v.validateCards()
	v.validateSessions()
	v.validateMastery()
	v.validatePlans()
	v.validateMaterials()
	v.validateReviewLog()
	c.JSON(http.StatusOK, validateResp{
		OK: len(v.errs) == 0, Errors: v.errs, Warnings: v.warns, Checked: v.checked,
	})
}

// dirReadErr 目录读取失败时记 warning（不存在视为空工作区，跳过不告警）。
func (v *validator) dirReadErr(rel, kind string, err error) {
	if err != nil && !os.IsNotExist(err) {
		v.warn(rel, kind, "目录读取失败，该段校验被跳过: "+err.Error())
	}
}

// ---------- 主题 manifest ----------

func (v *validator) validateManifests() {
	entries, err := os.ReadDir(v.st.TopicsDir())
	if err != nil {
		v.dirReadErr("topics/", "manifest", err)
		return
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		slug := e.Name()
		v.topics[slug] = true
		rel := "topics/" + slug + "/manifest.json"
		raw, err := os.ReadFile(filepath.Join(v.st.TopicsDir(), slug, "manifest.json"))
		if err != nil {
			if !os.IsNotExist(err) {
				v.errs = append(v.errs, validationIssue{File: rel, Kind: "manifest",
					Issues: []string{"读取失败: " + err.Error()}})
			}
			continue
		}
		v.checked["manifests"]++
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			v.err(rel, "manifest", []string{"解析失败: " + err.Error()})
			continue
		}
		var issues []string
		if s, _ := m["slug"].(string); s != slug {
			issues = append(issues, "slug 字段与目录名不一致")
		}
		if s, ok := m["status"]; ok {
			if v2, _ := s.(string); v2 != "active" && v2 != "paused" {
				issues = append(issues, fmt.Sprintf("status 非法: %v（合法 active/paused，缺省视为 active）", s))
			}
		}
		if len(issues) > 0 {
			v.err(rel, "manifest", issues)
		}
	}
}

// topicKnown 校验笔记/卡片的 topic 是否有对应主题 manifest。
// 空 topic 与 _archived（废弃卡）不校验。
func (v *validator) topicKnown(topic string) bool {
	return topic == "" || topic == "_archived" || v.topics[topic]
}

// ---------- 卡片 ----------

func (v *validator) validateCards() {
	ids, err := v.st.CardIDs()
	if err != nil {
		v.dirReadErr("cards/", "card", err)
		return
	}
	for _, id := range ids {
		v.checked["cards"]++
		rel := "cards/" + id + ".md"
		card, err := v.st.GetCard(id)
		if err != nil {
			v.err(rel, "card", []string{"解析失败: " + err.Error()})
			continue
		}
		var issues []string
		fm := card.FM
		if s, _ := fm["id"].(string); s != id {
			issues = append(issues, "frontmatter id 与文件名不一致")
		}
		for _, k := range []string{"topic", "front", "back", "created"} {
			if s, ok := fm[k].(string); !ok || strings.TrimSpace(s) == "" {
				issues = append(issues, "缺少或为空字段 "+k)
			}
		}
		if d, _ := fm["created"].(string); d != "" && !isDate(d) {
			issues = append(issues, "created 不是 YYYY-MM-DD")
		}
		if t, _ := fm["type"].(string); !cardTypes[t] {
			issues = append(issues, fmt.Sprintf("type 非法: %q（合法 basic/cloze）", t))
		}
		hint, _ := fm["hint"].(string)
		if strings.TrimSpace(hint) == "" {
			issues = append(issues, "缺少或为空字段 hint（回忆抓手，建卡必填）")
		} else if n := utf8.RuneCountInString(hint); n < 15 || n > 40 {
			issues = append(issues, fmt.Sprintf("hint 长度 %d 字，契约要求 15-40 字", n))
		}
		if topic, _ := fm["topic"].(string); !v.topicKnown(topic) {
			issues = append(issues, "topic 无对应主题 manifest: "+topic)
		}
		if note, _ := fm["note"].(string); note != "" && !v.noteIDs[note] {
			issues = append(issues, "note 引用的笔记不存在: "+note)
		}
		if m, ok := fm["fsrs"].(map[string]any); !ok {
			issues = append(issues, "缺少 fsrs 块（建卡须原样复制全零模板）")
		} else if st, err := fsrsx.FromMap(m); err != nil {
			msg := "fsrs 块非法: " + err.Error()
			if d, ok := m["due"].(string); ok && isDate(d) {
				msg += "（due 是裸日期；必须写成带引号的完整时间戳，如 due: \"2026-09-13T00:00:00Z\"，" +
					"否则 YAML 把它解析成时间对象后会被 normalizeTimes 降级为纯日期，FSRS 永远解析失败）"
			}
			issues = append(issues, msg)
		} else {
			if st.Due.IsZero() {
				issues = append(issues, "fsrs.due 为零值")
			}
			if st.State < 0 || st.State > 3 {
				issues = append(issues, fmt.Sprintf("fsrs.state 越界: %d（合法 0-3）", st.State))
			}
			if st.LastReview == nil && st.Reps > 0 {
				issues = append(issues, "fsrs.reps > 0 但 last_review 为空")
			}
		}
		if len(issues) > 0 {
			v.err(rel, "card", issues)
		}
	}
}

// ---------- 笔记 ----------

func (v *validator) validateNotes() {
	ids, err := v.st.NoteIDs()
	if err != nil {
		v.dirReadErr("notes/", "note", err)
		return
	}
	for _, id := range ids {
		v.noteIDs[id] = true
	}
	// 主题内 order 统计：同 topic 重复、或部分有部分没有 → warning（非 error）
	type orderStat struct {
		byOrder map[int][]string // order → 笔记 id
		total   int
		ordered int
	}
	orderStats := map[string]*orderStat{}
	for _, id := range ids {
		v.checked["notes"]++
		rel := "notes/" + id + ".md"
		note, err := v.st.GetNote(id)
		if err != nil {
			v.err(rel, "note", []string{"解析失败: " + err.Error()})
			continue
		}
		var issues []string
		fm := note.FM
		topic, _ := fm["topic"].(string)
		if topic != "" {
			st := orderStats[topic]
			if st == nil {
				st = &orderStat{byOrder: map[int][]string{}}
				orderStats[topic] = st
			}
			st.total++
			if o, ok := orderOf(fm); ok {
				st.ordered++
				st.byOrder[o] = append(st.byOrder[o], id)
			}
		}
		if s, _ := fm["id"].(string); s != id {
			issues = append(issues, "frontmatter id 与文件名不一致")
		}
		for _, k := range []string{"title", "topic", "created"} {
			if s, ok := fm[k].(string); !ok || strings.TrimSpace(s) == "" {
				issues = append(issues, "缺少或为空字段 "+k)
			}
		}
		if d, _ := fm["created"].(string); d != "" && !isDate(d) {
			issues = append(issues, "created 不是 YYYY-MM-DD")
		}
		if !v.topicKnown(topic) {
			issues = append(issues, "topic 无对应主题 manifest: "+topic)
		}
		if len(issues) > 0 {
			v.err(rel, "note", issues)
		}
		if links, ok := fm["links"].([]any); ok {
			var dangling []string
			for _, l := range links {
				if s := str(l); s != "" && !v.noteIDs[s] {
					dangling = append(dangling, s)
				}
			}
			if len(dangling) > 0 {
				v.warn(rel, "note", "links 引用的笔记不存在: "+strings.Join(dangling, ", "))
			}
		}
		if strings.TrimSpace(note.Body) == "" {
			v.warn(rel, "note", "正文为空（原子笔记应有自己的话的阐述）")
		}
	}
	for topic, st := range orderStats {
		if st.ordered > 0 && st.ordered < st.total {
			v.warn("notes/", "note",
				fmt.Sprintf("主题 %s：%d/%d 篇笔记有 order，其余缺失（同一主题内应全部带递进序号）", topic, st.ordered, st.total))
		}
		for o, noteIDs := range st.byOrder {
			if len(noteIDs) > 1 {
				v.warn("notes/", "note",
					fmt.Sprintf("主题 %s：order=%d 重复（%s）", topic, o, strings.Join(noteIDs, ", ")))
			}
		}
	}
}

// ---------- 会话 ----------

func (v *validator) validateSessions() {
	ids, err := v.st.SessionIDs()
	if err != nil {
		v.dirReadErr("sessions/", "session", err)
		return
	}
	for _, id := range ids {
		v.checked["sessions"]++
		rel := "sessions/" + id + ".md"
		sess, err := v.st.GetSession(id)
		if err != nil {
			v.err(rel, "session", []string{"解析失败: " + err.Error()})
			continue
		}
		var issues []string
		fm := sess.FM
		if s, _ := fm["id"].(string); s != id {
			issues = append(issues, "frontmatter id 与文件名不一致")
		}
		if t, _ := fm["type"].(string); !sessionTypes[t] {
			issues = append(issues, "type 非法: "+t)
		}
		for _, k := range []string{"topic", "date", "summary"} {
			if s, ok := fm[k].(string); !ok || strings.TrimSpace(s) == "" {
				issues = append(issues, "缺少或为空字段 "+k)
			}
		}
		if d, _ := fm["date"].(string); d != "" && !isDate(d) {
			issues = append(issues, "date 不是 YYYY-MM-DD")
		}
		if len(issues) > 0 {
			v.err(rel, "session", issues)
		}
		if strings.TrimSpace(sess.Body) == "" {
			v.warn(rel, "session", "正文为空（应记录对话要点/批改/报告）")
		}
	}
}

// ---------- 掌握度 ----------

func (v *validator) validateMastery() {
	mastery, err := v.st.ReadMastery()
	if err != nil {
		v.err("progress/mastery.json", "mastery", []string{"解析失败: " + err.Error()})
		return
	}
	for topic, val := range mastery {
		v.checked["mastery_topics"]++
		m, ok := val.(map[string]any)
		if !ok {
			v.err("progress/mastery.json", "mastery", []string{topic + ": 条目不是对象"})
			continue
		}
		var issues []string
		lvl, ok := m["level"].(float64)
		if !ok || lvl < 0 || lvl > 5 || lvl != float64(int(lvl)) {
			issues = append(issues, "level 必须是 0-5 整数")
		}
		if d, _ := m["updated"].(string); !isDate(d) {
			issues = append(issues, "updated 缺失或不是 YYYY-MM-DD")
		}
		if ev, ok := m["evidence"].([]any); !ok || len(ev) == 0 {
			issues = append(issues, "evidence 缺失或为空（调整掌握度必须附证据）")
		} else {
			for i, e := range ev {
				em, ok := e.(map[string]any)
				if !ok {
					issues = append(issues, fmt.Sprintf("evidence[%d] 不是对象", i))
					continue
				}
				if k, _ := em["kind"].(string); !evidenceKinds[k] {
					issues = append(issues, fmt.Sprintf("evidence[%d].kind 非法: %v", i, em["kind"]))
				}
				if d, _ := em["date"].(string); !isDate(d) {
					issues = append(issues, fmt.Sprintf("evidence[%d].date 不是 YYYY-MM-DD", i))
				}
				if d, _ := em["detail"].(string); strings.TrimSpace(d) == "" {
					issues = append(issues, fmt.Sprintf("evidence[%d].detail 为空", i))
				}
				delta, ok := em["delta"].(float64)
				if !ok {
					issues = append(issues, fmt.Sprintf("evidence[%d].delta 缺失或不是数字", i))
				} else if delta < -2 || delta > 2 || delta != float64(int(delta)) {
					issues = append(issues, fmt.Sprintf("evidence[%d].delta 越界: %v（合法 -2..+2 整数）", i, delta))
				}
			}
		}
		if len(issues) > 0 {
			v.err("progress/mastery.json", "mastery", append([]string{topic + ":"}, issues...))
		}
	}
}

// ---------- 学习计划 ----------

func planDates(fm map[string]any, issues []string) []string {
	for _, k := range []string{"created", "updated"} {
		if d, _ := fm[k].(string); d != "" && !isDate(d) {
			issues = append(issues, k+" 不是 YYYY-MM-DD")
		}
	}
	return issues
}

func (v *validator) validatePlans() {
	// 全局计划 data/plans/*.md
	entries, err := os.ReadDir(filepath.Join(v.st.DataDir, "plans"))
	if err != nil {
		v.dirReadErr("plans/", "plan", err)
	} else {
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".md") || strings.HasPrefix(name, ".") {
				continue
			}
			v.checked["plans"]++
			rel := "plans/" + name
			fm, ok := v.readPlanDoc(rel, filepath.Join(v.st.DataDir, "plans", name))
			if !ok {
				continue
			}
			var issues []string
			for _, k := range []string{"id", "title", "created", "updated"} {
				if s, ok := fm[k].(string); !ok || strings.TrimSpace(s) == "" {
					issues = append(issues, "缺少或为空字段 "+k)
				}
			}
			issues = planDates(fm, issues)
			if len(issues) > 0 {
				v.err(rel, "plan", issues)
			}
		}
	}
	// 主题计划 data/topics/<slug>/plan.md
	tentries, err := os.ReadDir(v.st.TopicsDir())
	if err != nil {
		v.dirReadErr("topics/", "plan", err)
		return
	}
	for _, e := range tentries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		slug := e.Name()
		raw, err := os.ReadFile(filepath.Join(v.st.TopicsDir(), slug, "plan.md"))
		if os.IsNotExist(err) {
			continue
		}
		v.checked["plans"]++
		rel := "topics/" + slug + "/plan.md"
		if err != nil {
			v.err(rel, "plan", []string{"读取失败: " + err.Error()})
			continue
		}
		doc, err := store.ParseDoc(raw)
		if err != nil {
			v.err(rel, "plan", []string{"解析失败: " + err.Error()})
			continue
		}
		fm := doc.Map()
		var issues []string
		for _, k := range []string{"topic", "goal", "created", "updated", "status"} {
			if s, ok := fm[k].(string); !ok || strings.TrimSpace(s) == "" {
				issues = append(issues, "缺少或为空字段 "+k)
			}
		}
		if s, _ := fm["topic"].(string); s != "" && s != slug {
			issues = append(issues, "frontmatter topic 与目录 slug 不一致")
		}
		if s, _ := fm["status"].(string); s != "" && s != "active" && s != "done" && s != "paused" {
			issues = append(issues, "status 非法: "+s)
		}
		issues = planDates(fm, issues)
		if len(issues) > 0 {
			v.err(rel, "plan", issues)
		}
	}
}

// readPlanDoc 读取并解析计划文件，失败记入 errors，ok=false。
func (v *validator) readPlanDoc(rel, path string) (map[string]any, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		v.err(rel, "plan", []string{"读取失败: " + err.Error()})
		return nil, false
	}
	doc, err := store.ParseDoc(raw)
	if err != nil {
		v.err(rel, "plan", []string{"解析失败: " + err.Error()})
		return nil, false
	}
	return doc.Map(), true
}

// ---------- 材料索引 ----------

func (v *validator) validateMaterials() {
	list, err := v.st.MaterialsIndex()
	if err != nil {
		v.err("library/materials.json", "materials", []string{"解析失败: " + err.Error()})
		return
	}
	v.checked["materials"] = len(list)
	for i, m := range list {
		var issues []string
		if m.ID == 0 {
			issues = append(issues, "id 缺失或为 0")
		}
		if strings.TrimSpace(m.OriginalName) == "" {
			issues = append(issues, "original_name 为空")
		}
		if strings.TrimSpace(m.StoredName) == "" {
			issues = append(issues, "stored_name 为空")
		}
		if strings.TrimSpace(m.Topic) == "" {
			issues = append(issues, "topic 为空")
		}
		if strings.TrimSpace(m.Status) == "" {
			issues = append(issues, "status 为空")
		}
		if strings.TrimSpace(m.ImportedAt) == "" {
			issues = append(issues, "imported_at 为空")
		}
		if len(issues) > 0 {
			v.err("library/materials.json", "materials",
				append([]string{fmt.Sprintf("条目[%d]:", i)}, issues...))
		}
	}
}

// ---------- 复习日志 ----------

func (v *validator) validateReviewLog() {
	_, bad, err := v.st.ReadReviewLog()
	if err != nil {
		v.dirReadErr("progress/review-log.jsonl", "review-log", err)
		return
	}
	if bad > 0 {
		v.warn("progress/review-log.jsonl", "review-log",
			fmt.Sprintf("%d 行无法解析（统计与改判已跳过这些行，请人工修复）", bad))
	}
}
