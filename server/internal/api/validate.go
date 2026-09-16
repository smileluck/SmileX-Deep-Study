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

func isDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func (a *API) Validate(c *gin.Context) {
	errs := []validationIssue{}
	warns := []validationIssue{}
	checked := map[string]int{"cards": 0, "notes": 0, "sessions": 0, "mastery_topics": 0, "materials": 0, "plans": 0, "manifests": 0}

	// ---- 主题 manifest ----
	if entries, err := os.ReadDir(a.Store.TopicsDir()); err == nil {
		for _, e := range entries {
			if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			slug := e.Name()
			rel := "topics/" + slug + "/manifest.json"
			raw, err := os.ReadFile(filepath.Join(a.Store.TopicsDir(), slug, "manifest.json"))
			if err != nil {
				continue
			}
			checked["manifests"]++
			var m map[string]any
			if err := json.Unmarshal(raw, &m); err != nil {
				errs = append(errs, validationIssue{File: rel, Kind: "manifest",
					Issues: []string{"解析失败: " + err.Error()}})
				continue
			}
			var issues []string
			if v, _ := m["slug"].(string); v != slug {
				issues = append(issues, "slug 字段与目录名不一致")
			}
			if s, ok := m["status"]; ok {
				if v, _ := s.(string); v != "active" && v != "paused" {
					issues = append(issues, fmt.Sprintf("status 非法: %v（合法 active/paused，缺省视为 active）", s))
				}
			}
			if len(issues) > 0 {
				errs = append(errs, validationIssue{File: rel, Kind: "manifest", Issues: issues})
			}
		}
	}

	// ---- 卡片 ----
	if ids, err := a.Store.CardIDs(); err == nil {
		for _, id := range ids {
			checked["cards"]++
			card, err := a.Store.GetCard(id)
			if err != nil {
				errs = append(errs, validationIssue{File: "cards/" + id + ".md", Kind: "card",
					Issues: []string{"解析失败: " + err.Error()}})
				continue
			}
			var issues []string
			fm := card.FM
			if v, _ := fm["id"].(string); v != id {
				issues = append(issues, "frontmatter id 与文件名不一致")
			}
			for _, k := range []string{"topic", "front", "back", "created"} {
				if s, ok := fm[k].(string); !ok || strings.TrimSpace(s) == "" {
					issues = append(issues, "缺少或为空字段 "+k)
				}
			}
			if m, ok := fm["fsrs"].(map[string]any); !ok {
				issues = append(issues, "缺少 fsrs 块（建卡须原样复制全零模板）")
			} else if st, err := fsrsx.FromMap(m); err != nil {
				issues = append(issues, "fsrs 块非法: "+err.Error())
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
				errs = append(errs, validationIssue{File: "cards/" + id + ".md", Kind: "card", Issues: issues})
			}
		}
	}

	// ---- 笔记 ----
	if ids, err := a.Store.NoteIDs(); err == nil {
		// 主题内 order 统计：同 topic 重复、或部分有部分没有 → warning（非 error）
		type orderStat struct {
			byOrder map[int][]string // order → 笔记 id
			total   int
			ordered int
		}
		orderStats := map[string]*orderStat{}
		for _, id := range ids {
			checked["notes"]++
			note, err := a.Store.GetNote(id)
			if err != nil {
				errs = append(errs, validationIssue{File: "notes/" + id + ".md", Kind: "note",
					Issues: []string{"解析失败: " + err.Error()}})
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
			if v, _ := fm["id"].(string); v != id {
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
			if len(issues) > 0 {
				errs = append(errs, validationIssue{File: "notes/" + id + ".md", Kind: "note", Issues: issues})
			}
			if strings.TrimSpace(note.Body) == "" {
				warns = append(warns, validationIssue{File: "notes/" + id + ".md", Kind: "note",
					Issues: []string{"正文为空（原子笔记应有自己的话的阐述）"}})
			}
		}
		for topic, st := range orderStats {
			if st.ordered > 0 && st.ordered < st.total {
				warns = append(warns, validationIssue{File: "notes/", Kind: "note",
					Issues: []string{fmt.Sprintf("主题 %s：%d/%d 篇笔记有 order，其余缺失（同一主题内应全部带递进序号）", topic, st.ordered, st.total)}})
			}
			for o, noteIDs := range st.byOrder {
				if len(noteIDs) > 1 {
					warns = append(warns, validationIssue{File: "notes/", Kind: "note",
						Issues: []string{fmt.Sprintf("主题 %s：order=%d 重复（%s）", topic, o, strings.Join(noteIDs, ", "))}})
				}
			}
		}
	}

	// ---- 会话 ----
	if ids, err := a.Store.SessionIDs(); err == nil {
		for _, id := range ids {
			checked["sessions"]++
			sess, err := a.Store.GetSession(id)
			if err != nil {
				errs = append(errs, validationIssue{File: "sessions/" + id + ".md", Kind: "session",
					Issues: []string{"解析失败: " + err.Error()}})
				continue
			}
			var issues []string
			fm := sess.FM
			t, _ := fm["type"].(string)
			if !sessionTypes[t] {
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
				errs = append(errs, validationIssue{File: "sessions/" + id + ".md", Kind: "session", Issues: issues})
			}
			if strings.TrimSpace(sess.Body) == "" {
				warns = append(warns, validationIssue{File: "sessions/" + id + ".md", Kind: "session",
					Issues: []string{"正文为空（应记录对话要点/批改/报告）"}})
			}
		}
	}

	// ---- 掌握度 ----
	if mastery, err := a.Store.ReadMastery(); err != nil {
		errs = append(errs, validationIssue{File: "progress/mastery.json", Kind: "mastery",
			Issues: []string{"解析失败: " + err.Error()}})
	} else {
		for topic, v := range mastery {
			checked["mastery_topics"]++
			m, ok := v.(map[string]any)
			if !ok {
				errs = append(errs, validationIssue{File: "progress/mastery.json", Kind: "mastery",
					Issues: []string{topic + ": 条目不是对象"}})
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
				}
			}
			if len(issues) > 0 {
				errs = append(errs, validationIssue{File: "progress/mastery.json", Kind: "mastery",
					Issues: append([]string{topic + ":"}, issues...)})
			}
		}
	}

	// ---- 学习计划 ----
	planDates := func(fm map[string]any, issues []string) []string {
		for _, k := range []string{"created", "updated"} {
			if d, _ := fm[k].(string); d != "" && !isDate(d) {
				issues = append(issues, k+" 不是 YYYY-MM-DD")
			}
		}
		return issues
	}
	if entries, err := os.ReadDir(filepath.Join(a.Store.DataDir, "plans")); err == nil {
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".md") || strings.HasPrefix(name, ".") {
				continue
			}
			checked["plans"]++
			rel := "plans/" + name
			raw, err := os.ReadFile(filepath.Join(a.Store.DataDir, "plans", name))
			if err != nil {
				errs = append(errs, validationIssue{File: rel, Kind: "plan", Issues: []string{"读取失败: " + err.Error()}})
				continue
			}
			doc, err := store.ParseDoc(raw)
			if err != nil {
				errs = append(errs, validationIssue{File: rel, Kind: "plan", Issues: []string{"解析失败: " + err.Error()}})
				continue
			}
			fm := doc.Map()
			var issues []string
			for _, k := range []string{"id", "title", "created", "updated"} {
				if s, ok := fm[k].(string); !ok || strings.TrimSpace(s) == "" {
					issues = append(issues, "缺少或为空字段 "+k)
				}
			}
			issues = planDates(fm, issues)
			if len(issues) > 0 {
				errs = append(errs, validationIssue{File: rel, Kind: "plan", Issues: issues})
			}
		}
	}
	if entries, err := os.ReadDir(a.Store.TopicsDir()); err == nil {
		for _, e := range entries {
			if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			slug := e.Name()
			raw, err := os.ReadFile(filepath.Join(a.Store.TopicsDir(), slug, "plan.md"))
			if os.IsNotExist(err) {
				continue
			}
			checked["plans"]++
			rel := "topics/" + slug + "/plan.md"
			if err != nil {
				errs = append(errs, validationIssue{File: rel, Kind: "plan", Issues: []string{"读取失败: " + err.Error()}})
				continue
			}
			doc, err := store.ParseDoc(raw)
			if err != nil {
				errs = append(errs, validationIssue{File: rel, Kind: "plan", Issues: []string{"解析失败: " + err.Error()}})
				continue
			}
			fm := doc.Map()
			var issues []string
			for _, k := range []string{"topic", "goal", "created", "updated", "status"} {
				if s, ok := fm[k].(string); !ok || strings.TrimSpace(s) == "" {
					issues = append(issues, "缺少或为空字段 "+k)
				}
			}
			if v, _ := fm["topic"].(string); v != "" && v != slug {
				issues = append(issues, "frontmatter topic 与目录 slug 不一致")
			}
			if s, _ := fm["status"].(string); s != "" && s != "active" && s != "done" && s != "paused" {
				issues = append(issues, "status 非法: "+s)
			}
			issues = planDates(fm, issues)
			if len(issues) > 0 {
				errs = append(errs, validationIssue{File: rel, Kind: "plan", Issues: issues})
			}
		}
	}

	// ---- 材料索引 ----
	if _, err := a.Store.MaterialsIndex(); err != nil {
		errs = append(errs, validationIssue{File: "library/materials.json", Kind: "materials",
			Issues: []string{"解析失败: " + err.Error()}})
	} else {
		checked["materials"] = 1
	}

	c.JSON(http.StatusOK, validateResp{
		OK: len(errs) == 0, Errors: errs, Warnings: warns, Checked: checked,
	})
}
