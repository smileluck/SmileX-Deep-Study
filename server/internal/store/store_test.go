package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"smilex-deep-study/server/internal/fsrsx"
)

// zeroFSRS 契约要求的全零 fsrs 块（due 带引号语义由 CreateCard 保证）。
func zeroFSRS() map[string]any {
	return map[string]any{
		"due": "2026-09-16T00:00:00Z", "stability": 0, "difficulty": 0,
		"elapsed_days": 0, "scheduled_days": 0, "reps": 0, "lapses": 0,
		"state": 0, "last_review": nil,
	}
}

func TestCreateCardValidYAMLQuotedDue(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Ensure(); err != nil {
		t.Fatal(err)
	}
	// note/topic 含特殊字符：不应破坏 frontmatter 结构
	id, err := s.CreateCard(CreateCardParams{
		Front: "问题：什么是稳定性？\n第二行",
		Back:  "答案：R 衰减到阈值的天数",
		Topic: `demo: "x" #y`,
		Note:  "note-1",
	}, zeroFSRS())
	if err != nil {
		t.Fatalf("CreateCard: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(s.CardsDir(), id+".md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if !strings.Contains(text, `due: "2026-09-16T00:00:00Z"`) {
		t.Errorf("due 未带引号输出:\n%s", text)
	}
	doc, err := ParseDoc(raw)
	if err != nil {
		t.Fatalf("生成的卡片 frontmatter 无法解析: %v\n%s", err, text)
	}
	fm := doc.Map()
	if fm["id"] != id {
		t.Errorf("id = %v", fm["id"])
	}
	if fm["topic"] != `demo: "x" #y` {
		t.Errorf("topic 特殊字符往返失败: %v", fm["topic"])
	}
	if fm["note"] != "note-1" {
		t.Errorf("note = %v", fm["note"])
	}
	if fm["type"] != "basic" {
		t.Errorf("type 默认 basic, got %v", fm["type"])
	}
	fmFsrs, ok := fm["fsrs"].(map[string]any)
	if !ok {
		t.Fatalf("fsrs 块缺失: %v", fm)
	}
	if _, err := fsrsx.FromMap(fmFsrs); err != nil {
		t.Errorf("fsrs 块解析失败: %v", err)
	}
}

func TestCreateCardIDsUnique(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Ensure(); err != nil {
		t.Fatal(err)
	}
	id1, err := s.CreateCard(CreateCardParams{Front: "q1", Back: "a1"}, zeroFSRS())
	if err != nil {
		t.Fatal(err)
	}
	id2, err := s.CreateCard(CreateCardParams{Front: "q2", Back: "a2"}, zeroFSRS())
	if err != nil {
		t.Fatal(err)
	}
	if id1 == id2 {
		t.Errorf("同秒建两张卡 id 冲突: %s", id1)
	}
}

func TestCreateCardEmptyFrontBack(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Ensure(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateCard(CreateCardParams{Front: "  ", Back: "a"}, zeroFSRS()); err == nil {
		t.Error("空 front 应报错")
	}
	if _, err := s.CreateCard(CreateCardParams{Front: "q", Back: ""}, zeroFSRS()); err == nil {
		t.Error("空 back 应报错")
	}
}

func TestWriteAtomicConcurrent(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/shared.txt"
	const writers = 32
	var wg sync.WaitGroup
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// 内容足够长，交叉污染一定会破坏前缀/后缀
			data := []byte(fmt.Sprintf("begin-%03d-%s-end\n", i, strings.Repeat("x", 4096)))
			for j := 0; j < 20; j++ {
				if err := writeAtomic(path, data); err != nil {
					t.Errorf("writeAtomic: %v", err)
					return
				}
			}
		}(i)
	}
	wg.Wait()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	var i int
	if n, _ := fmt.Sscanf(text, "begin-%03d-", &i); n != 1 {
		t.Fatalf("最终文件内容损坏: %.40q", text)
	}
	want := fmt.Sprintf("begin-%03d-%s-end\n", i, strings.Repeat("x", 4096))
	if text != want {
		t.Fatalf("最终文件来自多个写者交叉污染: %.40q", text)
	}
}

func TestValidID(t *testing.T) {
	cases := []struct {
		id   string
		want bool
	}{
		{"", false},
		{"..", false},
		{"a/b", false},
		{"a\\b", false},
		{"/etc", false},
		{`..\\x`, false},
		{"fsrs-memory-model", true},
		{"20260913-tutor-fsrs", true},
		{"spaced-repetition", true},
	}
	for _, tc := range cases {
		if got := ValidID(tc.id); got != tc.want {
			t.Errorf("ValidID(%q) = %v, want %v", tc.id, got, tc.want)
		}
	}
}
