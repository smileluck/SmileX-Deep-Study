// frontmatter 解析与序列化：经 yaml.Node 定向更新，
// 保留 agent 写入的任意额外字段与键序。
package store

import (
	"bytes"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Doc 是一个带 frontmatter 的 markdown 文档。
type Doc struct {
	FMNode *yaml.Node // MappingNode；无 frontmatter 时为空 mapping
	Body   string
}

// ParseDoc 解析 `---\n…\n---\nbody` 结构。
func ParseDoc(raw []byte) (*Doc, error) {
	lines := bytes.Split(raw, []byte("\n"))
	if len(lines) > 0 && bytes.Equal(bytes.TrimSpace(lines[0]), []byte("---")) {
		for i := 1; i < len(lines); i++ {
			if bytes.Equal(bytes.TrimSpace(lines[i]), []byte("---")) {
				fmBytes := bytes.Join(lines[1:i], []byte("\n"))
				var doc yaml.Node
				if err := yaml.Unmarshal(fmBytes, &doc); err != nil {
					return nil, fmt.Errorf("frontmatter 不是合法 YAML: %w", err)
				}
				var mapping *yaml.Node
				if len(doc.Content) > 0 && doc.Content[0].Kind == yaml.MappingNode {
					mapping = doc.Content[0]
				} else {
					mapping = emptyMapping()
				}
				return &Doc{FMNode: mapping, Body: string(bytes.Join(lines[i+1:], []byte("\n")))}, nil
			}
		}
		return nil, fmt.Errorf("frontmatter 未闭合（缺少第二个 ---）")
	}
	return &Doc{FMNode: emptyMapping(), Body: string(raw)}, nil
}

func emptyMapping() *yaml.Node {
	n := &yaml.Node{Kind: yaml.MappingNode}
	return n
}

// Map 把 frontmatter 转为普通 map（用于 JSON 输出）。
// YAML 会把 2026-09-13 这类裸标量解析成 time.Time，这里统一规范化为字符串。
func (d *Doc) Map() map[string]any {
	m := map[string]any{}
	if d.FMNode == nil || d.FMNode.Kind != yaml.MappingNode {
		return m
	}
	if err := d.FMNode.Decode(&m); err != nil {
		return map[string]any{}
	}
	normalizeTimes(m)
	return m
}

func normalizeTimes(m map[string]any) {
	for k, v := range m {
		switch t := v.(type) {
		case time.Time:
			if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() == 0 {
				m[k] = t.Format("2006-01-02")
			} else {
				m[k] = t.UTC().Format(time.RFC3339)
			}
		case map[string]any:
			normalizeTimes(t)
		case []any:
			for i, e := range t {
				if tt, ok := e.(time.Time); ok {
					if tt.Hour() == 0 && tt.Minute() == 0 && tt.Second() == 0 && tt.Nanosecond() == 0 {
						t[i] = tt.Format("2006-01-02")
					} else {
						t[i] = tt.UTC().Format(time.RFC3339)
					}
				}
			}
		}
	}
}

// Get 返回 frontmatter 顶层键的值。
func (d *Doc) Get(key string) any {
	if d.FMNode == nil || d.FMNode.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(d.FMNode.Content); i += 2 {
		if d.FMNode.Content[i].Value == key {
			var v any
			if err := d.FMNode.Content[i+1].Decode(&v); err == nil {
				return v
			}
		}
	}
	return nil
}

// UpdateMapping 定向替换/插入一个顶层键的值（保序，其余键不动）。
func (d *Doc) UpdateMapping(key string, valueNode *yaml.Node) {
	if d.FMNode.Kind != yaml.MappingNode {
		d.FMNode = emptyMapping()
	}
	for i := 0; i+1 < len(d.FMNode.Content); i += 2 {
		if d.FMNode.Content[i].Value == key {
			d.FMNode.Content[i+1] = valueNode
			return
		}
	}
	kn := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
	d.FMNode.Content = append(d.FMNode.Content, kn, valueNode)
}

// Bytes 重新拼装完整文件内容。
func (d *Doc) Bytes() ([]byte, error) {
	out, err := yaml.Marshal(d.FMNode)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(out)
	buf.WriteString("---\n")
	buf.WriteString(d.Body)
	return buf.Bytes(), nil
}

// ScalarNode 把 Go 值包装为 yaml 标量节点（nil → null）。
func ScalarNode(v any) *yaml.Node {
	n := &yaml.Node{Kind: yaml.ScalarNode}
	switch t := v.(type) {
	case nil:
		n.Tag = "!!null"
		n.Value = "null"
	case string:
		n.Tag = "!!str"
		n.Value = t
	case bool:
		n.Tag = "!!bool"
		n.Value = fmt.Sprintf("%v", t)
	case int:
		n.Tag = "!!int"
		n.Value = fmt.Sprintf("%d", t)
	case uint64:
		n.Tag = "!!int"
		n.Value = fmt.Sprintf("%d", t)
	case float64:
		n.Tag = "!!float"
		n.Value = fmt.Sprintf("%g", t)
	default:
		n.Tag = "!!str"
		n.Value = fmt.Sprintf("%v", t)
	}
	return n
}

// MapNode 按给定键序构造 mapping 节点。
func MapNode(keys []string, vals map[string]any) *yaml.Node {
	n := &yaml.Node{Kind: yaml.MappingNode}
	for _, k := range keys {
		n.Content = append(n.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: k},
			ScalarNode(vals[k]))
	}
	return n
}
