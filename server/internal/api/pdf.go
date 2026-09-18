package api

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ledongthuc/pdf"
	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// ExtractText 提取 inbox 中 PDF 的文本层。纯 Go 实现（CGO_ENABLED=0 可交叉编译），
// 随二进制分发，用户无需安装 poppler 等外部工具。
// 扫描件没有文本层，返回的 text 为空或近空——契约中由 agent 走
// "页转图 + 视觉阅读/OCR"降级路径，本端点如实返回即可。
func (a *API) ExtractText(c *gin.Context) {
	name := c.Query("name")
	if name == "" || filepath.Base(name) != name {
		errJSON(c, 400, fmt.Errorf("name 必须是 inbox 内的文件名（不含路径）"))
		return
	}
	if !strings.EqualFold(filepath.Ext(name), ".pdf") {
		errJSON(c, 400, fmt.Errorf("仅支持 .pdf 文件"))
		return
	}
	path := filepath.Join(a.Store.DataDir, "inbox", name)
	data, err := os.ReadFile(path)
	if err != nil {
		errJSON(c, 404, fmt.Errorf("inbox 中不存在 %s", name))
		return
	}

	text, pages, failed, err := extractPDFText(data)
	if err != nil {
		errJSON(c, 422, fmt.Errorf("PDF 解析失败: %v", err))
		return
	}
	c.JSON(200, gin.H{
		"name":         name,
		"pages":        pages,
		"failed_pages": failed,
		"chars":        len([]rune(text)),
		"text":         text,
	})
}

// extractPDFText 先直接解析；遇加密配置不支持（如 V=4 + RC4 流，常见于"能打开
// 但带空密码"的国产 PDF）时，先用 pdfcpu 以空密码解密到内存再解析。
func extractPDFText(data []byte) (text string, pages, failed int, err error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		var plain bytes.Buffer
		conf := model.NewDefaultConfiguration()
		conf.ValidationMode = model.ValidationRelaxed
		if derr := pdfcpuapi.Decrypt(bytes.NewReader(data), &plain, conf); derr != nil {
			return "", 0, 0, fmt.Errorf("%v（且空密码解密失败: %v；若 PDF 设有打开密码请先在阅读器中另存为无密码版本）", err, derr)
		}
		r, err = pdf.NewReader(bytes.NewReader(plain.Bytes()), int64(plain.Len()))
		if err != nil {
			return "", 0, 0, err
		}
	}

	var sb strings.Builder
	pages = r.NumPage()
	for i := 1; i <= pages; i++ {
		t, perr := r.Page(i).GetPlainText(nil)
		if perr != nil {
			failed++
			continue
		}
		sb.WriteString(t)
		sb.WriteString("\n\n")
	}
	return sb.String(), pages, failed, nil
}
