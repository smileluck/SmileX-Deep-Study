// store 纯文件存储层：data/ 目录是唯一数据源，每次请求实时读盘。
package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Store struct {
	DataDir string
}

func New(dataDir string) *Store { return &Store{DataDir: dataDir} }

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
func writeAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
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
	if id == "" || strings.ContainsRune(id, '/') || strings.ContainsRune(id, filepath.Separator) {
		return nil, fmt.Errorf("非法卡片 id")
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
	keys := []string{"due", "stability", "difficulty", "elapsed_days",
		"scheduled_days", "reps", "lapses", "state", "last_review"}
	doc.UpdateMapping("fsrs", MapNode(keys, fsrsMap))
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
}

func (s *Store) CreateCard(p CreateCardParams, zeroFSRS map[string]any) (string, error) {
	if strings.TrimSpace(p.Front) == "" || strings.TrimSpace(p.Back) == "" {
		return "", fmt.Errorf("front/back 不能为空")
	}
	if p.Type == "" {
		p.Type = "basic"
	}
	id := fmt.Sprintf("manual-%s", time.Now().Format("20060102-150405"))
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\n", id)
	fmt.Fprintf(&b, "note: %s\n", p.Note)
	fmt.Fprintf(&b, "topic: %s\n", p.Topic)
	fmt.Fprintf(&b, "type: %s\n", p.Type)
	fmt.Fprintf(&b, "front: |-\n%s", indentBlock(p.Front))
	fmt.Fprintf(&b, "back: |-\n%s", indentBlock(p.Back))
	fmt.Fprintf(&b, "created: %s\n", time.Now().Format("2006-01-02"))
	b.WriteString("fsrs:\n")
	for _, k := range []string{"due", "stability", "difficulty", "elapsed_days",
		"scheduled_days", "reps", "lapses", "state", "last_review"} {
		fmt.Fprintf(&b, "  %s: %v\n", k, yamlLine(zeroFSRS[k]))
	}
	b.WriteString("---\n")
	path := filepath.Join(s.CardsDir(), id+".md")
	if err := writeAtomic(path, []byte(b.String())); err != nil {
		return "", err
	}
	return id, nil
}

func indentBlock(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	var b strings.Builder
	for _, l := range lines {
		if l == "" {
			b.WriteString("\n")
		} else {
			b.WriteString("  " + l + "\n")
		}
	}
	return b.String()
}

func yamlLine(v any) string {
	if v == nil {
		return "null"
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
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
	if id == "" || strings.ContainsRune(id, '/') {
		return nil, fmt.Errorf("非法笔记 id")
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
	if id == "" || strings.ContainsRune(id, '/') {
		return nil, fmt.Errorf("非法会话 id")
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

type TopicStat struct {
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Goal  string `json:"goal"`
	Notes int    `json:"notes"`
	Cards int    `json:"cards"`
}

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

func (s *Store) AppendMaterial(m MaterialMeta) error {
	list, err := s.MaterialsIndex()
	if err != nil {
		return err
	}
	list = append(list, m)
	b, _ := json.MarshalIndent(list, "", "  ")
	return writeAtomic(s.MaterialsPath(), append(b, '\n'))
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
	Card       string `json:"card"`
	Rating     int    `json:"rating"`
	TS         string `json:"ts"`
	StateBefore int   `json:"state_before"`
	StateAfter  int   `json:"state_after"`
}

func (s *Store) ReadReviewLog() ([]ReviewLogEntry, error) {
	raw, err := os.ReadFile(filepath.Join(s.DataDir, "progress", "review-log.jsonl"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []ReviewLogEntry
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var e ReviewLogEntry
		if json.Unmarshal([]byte(line), &e) == nil {
			out = append(out, e)
		}
	}
	return out, nil
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
