package store

import (
	"strings"
	"testing"
)

func TestParseDocUpdateMappingRoundTrip(t *testing.T) {
	raw := `---
id: card-1
topic: demo
front: 问题
fsrs:
  due: "2026-09-13T00:00:00Z"
  stability: 0
---
正文内容
`
	doc, err := ParseDoc([]byte(raw))
	if err != nil {
		t.Fatalf("ParseDoc: %v", err)
	}
	doc.UpdateMapping("fsrs", MapNode([]string{"due", "stability"}, map[string]any{
		"due": "2026-09-20T00:00:00Z", "stability": 3.5,
	}))
	out, err := doc.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	// 正文与既有键保留
	if !strings.Contains(string(out), "正文内容") {
		t.Errorf("正文丢失: %s", out)
	}
	if !strings.Contains(string(out), "front: 问题") {
		t.Errorf("既有键 front 丢失: %s", out)
	}
	doc2, err := ParseDoc(out)
	if err != nil {
		t.Fatalf("二次 ParseDoc: %v", err)
	}
	m := doc2.Map()
	fsrs, ok := m["fsrs"].(map[string]any)
	if !ok {
		t.Fatalf("fsrs 块缺失: %v", m)
	}
	if fsrs["due"] != "2026-09-20T00:00:00Z" {
		t.Errorf("due = %v", fsrs["due"])
	}
	if fsrs["stability"] != 3.5 {
		t.Errorf("stability = %v", fsrs["stability"])
	}
}

func TestNormalizeTimes(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		key  string
		want any
	}{
		// 裸写零点整时间：被 YAML 解析成 time.Time 后降级为纯日期（契约已知坑）
		{"midnight-utc bare", "due: 2026-09-13T00:00:00Z\n", "due", "2026-09-13"},
		// 裸写非零点时间：规范化为 RFC3339 UTC
		{"unquoted time", "due: 2026-09-13T08:30:00Z\n", "due", "2026-09-13T08:30:00Z"},
		// 带引号的时间串保持字符串原样
		{"quoted time", "due: \"2026-09-13T00:00:00Z\"\n", "due", "2026-09-13T00:00:00Z"},
		// 裸日期：解析为 time.Time 后回写成同一日期串
		{"date-only", "created: 2026-09-13\n", "created", "2026-09-13"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := ParseDoc([]byte("---\n" + tc.raw + "---\n"))
			if err != nil {
				t.Fatalf("ParseDoc: %v", err)
			}
			if got := doc.Map()[tc.key]; got != tc.want {
				t.Errorf("got %v (%T), want %v", got, got, tc.want)
			}
		})
	}
}

func TestParseDocCRLF(t *testing.T) {
	raw := "---\r\nid: card-1\r\ncreated: 2026-09-13\r\n---\r\n正文\r\n"
	doc, err := ParseDoc([]byte(raw))
	if err != nil {
		t.Fatalf("CRLF ParseDoc: %v", err)
	}
	m := doc.Map()
	if m["id"] != "card-1" {
		t.Errorf("id = %v", m["id"])
	}
	if m["created"] != "2026-09-13" {
		t.Errorf("created = %v", m["created"])
	}
}

func TestParseDocNoFrontmatter(t *testing.T) {
	doc, err := ParseDoc([]byte("只有正文，没有 frontmatter\n"))
	if err != nil {
		t.Fatalf("ParseDoc: %v", err)
	}
	if len(doc.Map()) != 0 {
		t.Errorf("无 frontmatter 时应为空 map: %v", doc.Map())
	}
	if !strings.Contains(doc.Body, "只有正文") {
		t.Errorf("Body = %q", doc.Body)
	}
}
