// validate 是数据契约的硬校验：扫描 data/ 全部资产，
// 返回 schema 违规清单。agent 工作流收尾必须把它清零。
package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"smilex-deep-study/server/internal/fsrsx"
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
	"tutor": true, "feynman": true, "quiz": true, "diagnose": true, "import": true,
}

var evidenceKinds = map[string]bool{
	"quiz": true, "review": true, "feynman": true, "diagnose": true, "import": true,
}

func isDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func (a *API) Validate(c *gin.Context) {
	errs := []validationIssue{}
	warns := []validationIssue{}
	checked := map[string]int{"cards": 0, "notes": 0, "sessions": 0, "mastery_topics": 0, "materials": 0}

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
