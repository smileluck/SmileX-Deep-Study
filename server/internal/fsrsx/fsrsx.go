// fsrsx 封装 go-fsrs：frontmatter fsrs 块 ⇄ fsrs.Card 的转换与评分。
package fsrsx

import (
	"encoding/json"
	"fmt"
	"time"

	gofsrs "github.com/open-spaced-repetition/go-fsrs"
)

// RequestRetention 目标记忆率（FSRS 调度核心参数）。
const RequestRetention = 0.90

// State 对应卡片 frontmatter 的 fsrs 块。
type State struct {
	Due           time.Time  `json:"due"`
	Stability     float64    `json:"stability"`
	Difficulty    float64    `json:"difficulty"`
	ElapsedDays   uint64     `json:"elapsed_days"`
	ScheduledDays uint64     `json:"scheduled_days"`
	Reps          uint64     `json:"reps"`
	Lapses        uint64     `json:"lapses"`
	State         int        `json:"state"`
	LastReview    *time.Time `json:"last_review"`
}

// FromMap 从 frontmatter 的 map 形式解析（缺字段容忍：新卡可能只有零值）。
func FromMap(m map[string]any) (*State, error) {
	if m == nil {
		return nil, fmt.Errorf("fsrs 块缺失")
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var st State
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, fmt.Errorf("fsrs 块格式错误: %w", err)
	}
	return &st, nil
}

// New 生成新卡的零状态（due=now）。
func New(now time.Time) *State {
	return &State{Due: now.UTC()}
}

// Keys 返回写回时的固定键序。
func Keys() []string {
	return []string{"due", "stability", "difficulty", "elapsed_days",
		"scheduled_days", "reps", "lapses", "state", "last_review"}
}

// ToMap 转为写回 frontmatter 的 map（时间 RFC3339 UTC，未复习为 null）。
func (s *State) ToMap() map[string]any {
	m := map[string]any{
		"due":            s.Due.UTC().Format(time.RFC3339),
		"stability":      s.Stability,
		"difficulty":     s.Difficulty,
		"elapsed_days":   s.ElapsedDays,
		"scheduled_days": s.ScheduledDays,
		"reps":           s.Reps,
		"lapses":         s.Lapses,
		"state":          s.State,
		"last_review":    nil,
	}
	if s.LastReview != nil {
		m["last_review"] = s.LastReview.UTC().Format(time.RFC3339)
	}
	return m
}

func (s *State) toCard() gofsrs.Card {
	c := gofsrs.Card{
		Due:           s.Due,
		Stability:     s.Stability,
		Difficulty:    s.Difficulty,
		ElapsedDays:   s.ElapsedDays,
		ScheduledDays: s.ScheduledDays,
		Reps:          s.Reps,
		Lapses:        s.Lapses,
		State:         gofsrs.State(s.State),
	}
	if s.LastReview != nil {
		c.LastReview = *s.LastReview
	}
	return c
}

func fromCard(c gofsrs.Card) *State {
	s := &State{
		Due:           c.Due,
		Stability:     c.Stability,
		Difficulty:    c.Difficulty,
		ElapsedDays:   c.ElapsedDays,
		ScheduledDays: c.ScheduledDays,
		Reps:          c.Reps,
		Lapses:        c.Lapses,
		State:         int(c.State),
	}
	if !c.LastReview.IsZero() {
		t := c.LastReview
		s.LastReview = &t
	}
	return s
}

// Grade 执行一次评分（rating: 1=Again 2=Hard 3=Good 4=Easy），
// 返回新状态。这是全系统唯一写调度状态的入口。
func Grade(st *State, rating int, now time.Time) (*State, error) {
	if rating < 1 || rating > 4 {
		return nil, fmt.Errorf("rating 必须是 1-4，收到 %d", rating)
	}
	p := gofsrs.DefaultParam()
	p.RequestRetention = RequestRetention
	info, ok := p.Repeat(st.toCard(), now)[gofsrs.Rating(rating)]
	if !ok {
		return nil, fmt.Errorf("FSRS 未返回 rating=%d 的结果", rating)
	}
	return fromCard(info.Card), nil
}

// StateName 状态数值的可读名。
func StateName(s int) string {
	switch gofsrs.State(s) {
	case gofsrs.New:
		return "new"
	case gofsrs.Learning:
		return "learning"
	case gofsrs.Review:
		return "review"
	case gofsrs.Relearning:
		return "relearning"
	}
	return "unknown"
}
