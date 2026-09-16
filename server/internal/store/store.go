// store 纯文件存储层：data/ 目录是唯一数据源，每次请求实时读盘。
package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"smilex-deep-study/server/internal/fsrsx"
)

type Store struct {
	DataDir string
	mu      sync.Mutex
}

func New(dataDir string) *Store { return &Store{DataDir: dataDir} }

// Lock/Unlock 串行化跨多次文件操作的读-改-写流程（评分、改判、建卡、上传、
// recall 追加等）。gin 每个请求一个 goroutine，调用方在写路径入口持锁。
func (s *Store) Lock()   { s.mu.Lock() }
func (s *Store) Unlock() { s.mu.Unlock() }

var subdirs = []string{"inbox", "library", "topics", "notes", "cards", "sessions", "progress"}

// Ensure 创建目录骨架与空索引文件。
func (s *Store) Ensure() error {
	for _, d := range subdirs {
		if err := os.MkdirAll(filepath.Join(s.DataDir, d), 0o755); err != nil {
			return err
		}
	}
	if _, err := os.Stat(s.MaterialsPath()); os.IsNotExist(err) {
		if err := os.WriteFile(s.MaterialsPath(), []byte("[]\n"), 0o644); err != nil {
			return err
		}
	}
	if _, err := os.Stat(s.MasteryPath()); os.IsNotExist(err) {
		if err := os.WriteFile(s.MasteryPath(), []byte("{}\n"), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) MaterialsPath() string { return filepath.Join(s.DataDir, "library", "materials.json") }
func (s *Store) MasteryPath() string   { return filepath.Join(s.DataDir, "progress", "mastery.json") }

// ---------- 通用 ----------

func (s *Store) readDoc(path string) (*Doc, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseDoc(raw)
}

// writeAtomic 先写临时文件再替换，避免 agent 并发读到半截文件。
// 临时文件用同目录随机名，并发写同一路径互不污染。
func writeAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+"-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}

// ErrInvalidID 表示 id/slug 为空、为 ".." 或含路径分隔符；
// ErrNotFound 表示目标资源不存在。调用方用 errors.Is 区分 400/404。
var (
	ErrInvalidID = errors.New("非法 id")
	ErrNotFound  = errors.New("目标不存在")
)

// ValidID 校验 slug/id：拒绝空、".." 与路径分隔符（含 Windows 反斜杠），防目录逃逸。
func ValidID(id string) bool {
	if id == "" || id == ".." {
		return false
	}
	return !strings.ContainsRune(id, '/') &&
		!strings.ContainsRune(id, '\\') &&
		!strings.ContainsRune(id, os.PathSeparator)
}

// IDsIn 返回目录下全部 markdown 文件名（不含扩展名），按名排序。
func IDsIn(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") || strings.HasPrefix(name, ".") {
			continue
		}
		ids = append(ids, strings.TrimSuffix(name, ".md"))
	}
	sort.Strings(ids)
	return ids, nil
}

// ---------- 卡片 ----------

type Card struct {
	ID   string
	Path string
	FM   map[string]any
	Body string
}

func (s *Store) CardsDir() string { return filepath.Join(s.DataDir, "cards") }

// CardIDs / NoteIDs / SessionIDs 供校验器严格遍历（含解析失败的文件）。
func (s *Store) CardIDs() ([]string, error)    { return IDsIn(s.CardsDir()) }
func (s *Store) NoteIDs() ([]string, error)    { return IDsIn(s.NotesDir()) }
func (s *Store) SessionIDs() ([]string, error) { return IDsIn(s.SessionsDir()) }

func (s *Store) ListCards() ([]Card, error) {
	ids, err := IDsIn(s.CardsDir())
	if err != nil {
		return nil, err
	}
	cards := make([]Card, 0, len(ids))
	for _, id := range ids {
		c, err := s.GetCard(id)
		if err != nil {
			continue // 坏文件跳过而不是整体失败
		}
		cards = append(cards, *c)
	}
	return cards, nil
}

func (s *Store) GetCard(id string) (*Card, error) {
	if !ValidID(id) {
		return nil, fmt.Errorf("非法卡片 id %q: %w", id, ErrInvalidID)
	}
	path := filepath.Join(s.CardsDir(), id+".md")
	doc, err := s.readDoc(path)
	if err != nil {
		return nil, err
	}
	fm := doc.Map()
	fm["id"] = id
	return &Card{ID: id, Path: path, FM: fm, Body: doc.Body}, nil
}

// CardFSRSMap 提取 fsrs 块（缺失/为空返回 nil, nil 表示新卡）。
func (c *Card) CardFSRSMap() map[string]any {
	v, _ := c.FM["fsrs"].(map[string]any)
	return v
}

// WriteCardFSRS 定向重写某卡的 fsrs 块。
func (s *Store) WriteCardFSRS(id string, fsrsMap map[string]any) error {
	path := filepath.Join(s.CardsDir(), id+".md")
	doc, err := s.readDoc(path)
	if err != nil {
		return err
	}
	doc.UpdateMapping("fsrs", MapNode(fsrsx.Keys(), fsrsMap))
	b, err := doc.Bytes()
	if err != nil {
		return err
	}
	return writeAtomic(path, b)
}

// CreateCardParams UI 手动建卡参数。
type CreateCardParams struct {
	Front string
	Back  string
	Topic string
	Note  string
	Type  string // basic | cloze
	Hint  string // 可选：契约要求必填，UI 暂未提供时留空
}

func (s *Store) CreateCard(p CreateCardParams, zeroFSRS map[string]any) (string, error) {
	if strings.TrimSpace(p.Front) == "" || strings.TrimSpace(p.Back) == "" {
		return "", fmt.Errorf("front/back 不能为空")
	}
	if p.Type == "" {
		p.Type = "basic"
	}
	now := time.Now()
	// 随机后缀防同秒建卡互相覆盖
	id := fmt.Sprintf("manual-%s-%s", now.Format("20060102-150405"), randHex(2))

	// 标量统一走 yaml.Node 序列化：yaml.Marshal 按需加引号/转义，
	// note/topic 含特殊字符不会破坏或注入 frontmatter；due 强制双引号。
	strNode := func(v string, style yaml.Style) *yaml.Node {
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v, Style: style}
	}
	fm := &yaml.Node{Kind: yaml.MappingNode}
	add := func(k string, v *yaml.Node) {
		fm.Content = append(fm.Content, strNode(k, 0), v)
	}
	add("id", strNode(id, 0))
	add("note", strNode(p.Note, 0))
	add("topic", strNode(p.Topic, 0))
	add("type", strNode(p.Type, 0))
	add("front", strNode(p.Front, yaml.LiteralStyle))
	add("back", strNode(p.Back, yaml.LiteralStyle))
	if p.Hint != "" {
		add("hint", strNode(p.Hint, 0))
	}
	add("created", strNode(now.Format("2006-01-02"), 0))
	fsrsNode := &yaml.Node{Kind: yaml.MappingNode}
	for _, k := range fsrsx.Keys() {
		v := ScalarNode(zeroFSRS[k])
		if k == "due" {
			v.Style = yaml.DoubleQuotedStyle
		}
		fsrsNode.Content = append(fsrsNode.Content, strNode(k, 0), v)
	}
	add("fsrs", fsrsNode)
	out, err := yaml.Marshal(fm)
	if err != nil {
		return "", err
	}
	content := "---\n" + string(out) + "---\n"
	path := filepath.Join(s.CardsDir(), id+".md")
	if err := writeAtomic(path, []byte(content)); err != nil {
		return "", err
	}
	return id, nil
}

// randHex 返回 n 字节的十六进制串（2n 个字符）。
func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// ---------- 笔记 ----------

type Note struct {
	ID   string
	Path string
	FM   map[string]any
	Body string
}

func (s *Store) NotesDir() string { return filepath.Join(s.DataDir, "notes") }

func (s *Store) ListNotes() ([]Note, error) {
	ids, err := IDsIn(s.NotesDir())
	if err != nil {
		return nil, err
	}
	notes := make([]Note, 0, len(ids))
	for _, id := range ids {
		path := filepath.Join(s.NotesDir(), id+".md")
		doc, err := s.readDoc(path)
		if err != nil {
			continue
		}
		fm := doc.Map()
		fm["id"] = id
		notes = append(notes, Note{ID: id, Path: path, FM: fm, Body: doc.Body})
	}
	return notes, nil
}

func (s *Store) GetNote(id string) (*Note, error) {
	if !ValidID(id) {
		return nil, fmt.Errorf("非法笔记 id %q: %w", id, ErrInvalidID)
	}
	path := filepath.Join(s.NotesDir(), id+".md")
	doc, err := s.readDoc(path)
	if err != nil {
		return nil, err
	}
	fm := doc.Map()
	fm["id"] = id
	return &Note{ID: id, Path: path, FM: fm, Body: doc.Body}, nil
}

// ---------- 会话 ----------

type Session struct {
	ID   string
	Path string
	FM   map[string]any
	Body string
}

func (s *Store) SessionsDir() string { return filepath.Join(s.DataDir, "sessions") }

func (s *Store) ListSessions() ([]Session, error) {
	ids, err := IDsIn(s.SessionsDir())
	if err != nil {
		return nil, err
	}
	sessions := make([]Session, 0, len(ids))
	for _, id := range ids {
		path := filepath.Join(s.SessionsDir(), id+".md")
		doc, err := s.readDoc(path)
		if err != nil {
			continue
		}
		fm := doc.Map()
		fm["id"] = id
		sessions = append(sessions, Session{ID: id, Path: path, FM: fm, Body: doc.Body})
	}
	// 按日期倒序（id 首段是日期时直接按 id 倒序即可近似）
	sort.Slice(sessions, func(i, j int) bool {
		di, _ := sessions[i].FM["date"].(string)
		dj, _ := sessions[j].FM["date"].(string)
		if di != dj {
			return di > dj
		}
		return sessions[i].ID > sessions[j].ID
	})
	return sessions, nil
}

func (s *Store) GetSession(id string) (*Session, error) {
	if !ValidID(id) {
		return nil, fmt.Errorf("非法会话 id %q: %w", id, ErrInvalidID)
	}
	path := filepath.Join(s.SessionsDir(), id+".md")
	doc, err := s.readDoc(path)
	if err != nil {
		return nil, err
	}
	fm := doc.Map()
	fm["id"] = id
	return &Session{ID: id, Path: path, FM: fm, Body: doc.Body}, nil
}

// ---------- 主题 ----------

func (s *Store) TopicsDir() string { return filepath.Join(s.DataDir, "topics") }

func (s *Store) ListTopics() ([]map[string]any, error) {
	entries, err := os.ReadDir(s.TopicsDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out = []map[string]any{}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(s.TopicsDir(), e.Name(), "manifest.json"))
		if err != nil {
			continue
		}
		var m map[string]any
		if json.Unmarshal(raw, &m) == nil {
			m["slug"] = e.Name()
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		si, _ := out[i]["slug"].(string)
		sj, _ := out[j]["slug"].(string)
		return si < sj
	})
	return out, nil
}

// ---------- 学习计划 ----------

type Plan struct {
	Slug string // 主题计划为 slug；全局计划为 "master"
	Path string
	FM   map[string]any
	Body string
}

// MasterPlan 读取全局计划；文件不存在返回 nil, nil。
func (s *Store) MasterPlan() (*Plan, error) {
	path := filepath.Join(s.DataDir, "plans", "master.md")
	doc, err := s.readDoc(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return &Plan{Slug: "master", Path: path, FM: doc.Map(), Body: doc.Body}, nil
}

// TopicPlans 扫描 data/topics/*/plan.md，按 slug 排序。
func (s *Store) TopicPlans() ([]Plan, error) {
	entries, err := os.ReadDir(s.TopicsDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	plans := make([]Plan, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		path := filepath.Join(s.TopicsDir(), e.Name(), "plan.md")
		doc, err := s.readDoc(path)
		if err != nil {
			continue // 坏文件跳过而不是整体失败
		}
		plans = append(plans, Plan{Slug: e.Name(), Path: path, FM: doc.Map(), Body: doc.Body})
	}
	sort.Slice(plans, func(i, j int) bool { return plans[i].Slug < plans[j].Slug })
	return plans, nil
}

func (s *Store) GetTopicPlan(slug string) (*Plan, error) {
	if !ValidID(slug) {
		return nil, fmt.Errorf("非法主题 slug %q: %w", slug, ErrInvalidID)
	}
	path := filepath.Join(s.TopicsDir(), slug, "plan.md")
	doc, err := s.readDoc(path)
	if err != nil {
		return nil, err
	}
	return &Plan{Slug: slug, Path: path, FM: doc.Map(), Body: doc.Body}, nil
}

// SetTopicStatus 写 manifest.json 的 status 字段（active/paused，由调用方校验取值）。
// 主题不存在时返回包裹 ErrNotFound 的错误。
func (s *Store) SetTopicStatus(slug, status string) error {
	if !ValidID(slug) {
		return fmt.Errorf("非法主题 slug %q: %w", slug, ErrInvalidID)
	}
	path := filepath.Join(s.TopicsDir(), slug, "manifest.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("主题 %s 的 manifest 不存在: %w", slug, ErrNotFound)
		}
		return fmt.Errorf("主题 %s 的 manifest 读取失败: %w", slug, err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("主题 %s 的 manifest 解析失败: %w", slug, err)
	}
	m["status"] = status
	b, _ := json.MarshalIndent(m, "", "  ")
	return writeAtomic(path, append(b, '\n'))
}

// PausedTopics 返回 status=="paused" 的主题 slug 集合；读取出错时降级为空集合（不过滤）。
func (s *Store) PausedTopics() map[string]bool {
	paused := map[string]bool{}
	topics, err := s.ListTopics()
	if err != nil {
		return paused
	}
	for _, t := range topics {
		if status, _ := t["status"].(string); status == "paused" {
			if slug, _ := t["slug"].(string); slug != "" {
				paused[slug] = true
			}
		}
	}
	return paused
}

// ---------- 材料 ----------

type MaterialMeta struct {
	ID           int    `json:"id"`
	OriginalName string `json:"original_name"`
	StoredName   string `json:"stored_name"`
	Topic        string `json:"topic"`
	Status       string `json:"status"`
	ImportedAt   string `json:"imported_at"`
}

type FileInfo struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Mtime string `json:"mtime"`
}

func (s *Store) MaterialsIndex() ([]MaterialMeta, error) {
	raw, err := os.ReadFile(s.MaterialsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return []MaterialMeta{}, nil
		}
		return nil, err
	}
	list := []MaterialMeta{}
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("materials.json 损坏: %w", err)
	}
	return list, nil
}

func (s *Store) InboxList() ([]FileInfo, error) {
	return listFiles(filepath.Join(s.DataDir, "inbox"))
}

func (s *Store) LibraryList() ([]FileInfo, error) {
	return listFiles(filepath.Join(s.DataDir, "library"))
}

func listFiles(dir string) ([]FileInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []FileInfo{}, nil
		}
		return nil, err
	}
	out := []FileInfo{}
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") || e.Name() == "materials.json" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, FileInfo{
			Name:  e.Name(),
			Size:  info.Size(),
			Mtime: info.ModTime().Format(time.RFC3339),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Mtime > out[j].Mtime })
	return out, nil
}

// SaveInbox 保存上传文件（重名自动加序号前缀），返回最终文件名。
func (s *Store) SaveInbox(name string, r io.Reader) (string, error) {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "_")
	if name == "" || name == "." || name == ".." || name == "/" {
		name = fmt.Sprintf("upload-%d", time.Now().Unix())
	}
	dst := filepath.Join(s.DataDir, "inbox", name)
	if _, err := os.Stat(dst); err == nil {
		name = fmt.Sprintf("%d-%s", time.Now().Unix(), name)
		dst = filepath.Join(s.DataDir, "inbox", name)
	}
	f, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return name, nil
}

// ---------- 日志与掌握度 ----------

func (s *Store) AppendJSONL(rel string, v any) error {
	path := filepath.Join(s.DataDir, rel)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	b, _ := json.Marshal(v)
	_, err = f.Write(append(b, '\n'))
	return err
}

// ReviewLogEntry review-log.jsonl 的一行。
type ReviewLogEntry struct {
	Card        string         `json:"card"`
	Rating      int            `json:"rating"`
	TS          string         `json:"ts"`
	StateBefore int            `json:"state_before"`
	StateAfter  int            `json:"state_after"`
	// FSRSBefore 评分前完整 fsrs 状态，供改判（regrade）还原；旧日志行没有此字段
	FSRSBefore map[string]any `json:"fsrs_before,omitempty"`
	// Void 作废标记行：配合 VoidOf 作废一条历史评分（日志只追加，不改旧行）
	Void   bool   `json:"void,omitempty"`
	VoidOf string `json:"void_of,omitempty"`
}

// ReadReviewLog 读取复习日志；返回无法解析的行数（跳过但计数，供校验器告警）。
func (s *Store) ReadReviewLog() ([]ReviewLogEntry, int, error) {
	raw, err := os.ReadFile(filepath.Join(s.DataDir, "progress", "review-log.jsonl"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	var out []ReviewLogEntry
	bad := 0
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var e ReviewLogEntry
		if json.Unmarshal([]byte(line), &e) == nil {
			out = append(out, e)
		} else {
			bad++
		}
	}
	return out, bad, nil
}

func (s *Store) ReadMastery() (map[string]any, error) {
	raw, err := os.ReadFile(s.MasteryPath())
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("mastery.json 损坏: %w", err)
	}
	return m, nil
}
