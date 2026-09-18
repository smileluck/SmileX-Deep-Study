package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// minimalPDF 生成带正确 xref 表的单页 PDF（Helvetica 文字），供提取测试用。
func minimalPDF(t *testing.T, text string) []byte {
	t.Helper()
	stream := fmt.Sprintf("BT /F1 24 Tf 100 700 Td (%s) Tj ET", text)
	objects := []string{
		"<</Type/Catalog/Pages 2 0 R>>",
		"<</Type/Pages/Kids[3 0 R]/Count 1>>",
		"<</Type/Page/Parent 2 0 R/MediaBox[0 0 612 792]/Contents 4 0 R/Resources<</Font<</F1 5 0 R>>>>>>",
		fmt.Sprintf("<</Length %d>>\nstream\n%s\nendstream", len(stream), stream),
		"<</Type/Font/Subtype/Type1/BaseFont/Helvetica>>",
	}
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects))
	for i, obj := range objects {
		offsets[i] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xref := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, off := range offsets {
		fmt.Fprintf(&buf, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&buf, "trailer\n<</Size %d/Root 1 0 R>>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return buf.Bytes()
}

func TestExtractText(t *testing.T) {
	r, st := setup(t)
	if err := os.WriteFile(filepath.Join(st.DataDir, "inbox", "hello.pdf"), minimalPDF(t, "Hello PDF"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, resp := doJSON(t, r, http.MethodGet, "/api/materials/extract-text?name=hello.pdf", nil)
	if code != 200 {
		t.Fatalf("extract-text status = %d, body = %s", code, resp)
	}
	var out struct {
		Pages int    `json:"pages"`
		Chars int    `json:"chars"`
		Text  string `json:"text"`
	}
	if err := json.Unmarshal(resp, &out); err != nil {
		t.Fatal(err)
	}
	if out.Pages != 1 {
		t.Errorf("pages = %d, want 1", out.Pages)
	}
	if !strings.Contains(out.Text, "Hello PDF") {
		t.Errorf("text 未包含写入的文字: %q", out.Text)
	}
	if out.Chars == 0 {
		t.Error("chars = 0，提取为空")
	}
}

func TestExtractTextRejectsBadInput(t *testing.T) {
	r, st := setup(t)
	if err := os.WriteFile(filepath.Join(st.DataDir, "inbox", "note.md"), []byte("# hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		want int
	}{
		{"", 400},                      // 缺参数
		{"../library/x.pdf", 400},      // 路径穿越
		{"note.md", 400},               // 非 PDF
		{"missing.pdf", 404},           // 不存在
	}
	for _, tc := range cases {
		code, _ := doJSON(t, r, http.MethodGet, "/api/materials/extract-text?name="+tc.name, nil)
		if code != tc.want {
			t.Errorf("name=%q: status = %d, want %d", tc.name, code, tc.want)
		}
	}
}
