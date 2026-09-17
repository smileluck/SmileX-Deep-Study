package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v0.1.0", "v0.2.0", -1},
		{"0.1.0", "v0.1.0", 0},
		{"v0.10.0", "v0.9.9", 1},
		{"v1.0.0", "v1.0.0", 0},
		{"v0.2.0-rc1", "v0.2.0", 0},
		{"v1.0", "v1.0.0", 0},
		{"v2.0.0", "v1.9.9", 1},
	}
	for _, c := range cases {
		if got := CompareVersions(c.a, c.b); got != c.want {
			t.Errorf("CompareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestAssetName(t *testing.T) {
	cases := []struct {
		tag, goos, goarch, want string
	}{
		{"v0.2.0", "darwin", "arm64", "deep-study-v0.2.0-darwin-arm64.tar.gz"},
		{"v0.2.0", "linux", "amd64", "deep-study-v0.2.0-linux-amd64.tar.gz"},
		{"v0.2.0", "windows", "amd64", "deep-study-v0.2.0-windows-amd64.zip"},
	}
	for _, c := range cases {
		if got := AssetName(c.tag, c.goos, c.goarch); got != c.want {
			t.Errorf("AssetName(%q, %q, %q) = %q, want %q", c.tag, c.goos, c.goarch, got, c.want)
		}
	}
}

// TestDownload 用本地 HTTP 服务模拟 release 资产：校验通过则解出二进制，篡改则拒绝。
func TestDownload(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("本测试构造 tar.gz 包，Windows 走 zip 路径")
	}

	content := []byte("#!/bin/sh\necho fake-binary\n")
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "deep-study", Mode: 0o755, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gz.Close()
	pkg := buf.Bytes()

	name := AssetName("v9.9.9", runtime.GOOS, runtime.GOARCH)
	setup := func(t *testing.T, sumLine string) {
		t.Helper()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, ".sha256") {
				w.Write([]byte(sumLine))
				return
			}
			w.Write(pkg)
		}))
		old := downloadBaseURL
		downloadBaseURL = srv.URL
		t.Cleanup(func() {
			downloadBaseURL = old
			srv.Close()
		})
	}

	t.Run("校验通过", func(t *testing.T) {
		setup(t, fmt.Sprintf("%x  %s\n", sha256.Sum256(pkg), name))
		path, err := Download(context.Background(), "v9.9.9")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(path)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, content) {
			t.Errorf("解出的二进制内容不符")
		}
		if fi, _ := os.Stat(path); fi.Mode().Perm() != 0o755 {
			t.Errorf("权限 = %o, want 755", fi.Mode().Perm())
		}
	})

	t.Run("校验不通过则拒绝", func(t *testing.T) {
		setup(t, fmt.Sprintf("%x  %s\n", sha256.Sum256([]byte("tampered")), name))
		if _, err := Download(context.Background(), "v9.9.9"); err == nil ||
			!strings.Contains(err.Error(), "SHA256") {
			t.Errorf("期望 SHA256 校验失败错误，得到 %v", err)
		}
	})
}
