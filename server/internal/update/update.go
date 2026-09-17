// Package update 基于 GitHub Releases 的自更新：版本检查、下载校验、原子替换。
// 稳定性原则：校验不过不替换、替换可回滚、网络失败优雅降级。
package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	owner = "smileluck"
	repo  = "SmileX-Deep-Study"
)

// ManualURL 所有网络路径都失败时给用户的手动下载页。
const ManualURL = "https://github.com/" + owner + "/" + repo + "/releases/latest"

var (
	apiLatestURL  = "https://api.github.com/repos/" + owner + "/" + repo + "/releases/latest"
	htmlLatestURL = "https://github.com/" + owner + "/" + repo + "/releases/latest"
	// downloadBaseURL 资产下载地址前缀，测试可替换为本地服务。
	downloadBaseURL = fmt.Sprintf("https://github.com/%s/%s/releases/download", owner, repo)
)

// mirrorPrefixes 下载/查询失败时依次尝试的镜像前缀（首个空串 = 直连）。
// 环境变量 DEEP_STUDY_UPDATE_MIRROR 可指定自定义镜像，替代内置镜像列表。
func mirrorPrefixes() []string {
	prefixes := []string{""}
	if m := strings.TrimRight(os.Getenv("DEEP_STUDY_UPDATE_MIRROR"), "/"); m != "" {
		return append(prefixes, m+"/")
	}
	return append(prefixes, "https://ghfast.top/", "https://gh-proxy.com/")
}

// CompareVersions 比较 vX.Y.Z 版本号：a<b 返回 -1，相等返回 0，a>b 返回 1。
// 容忍带不带 v 前缀与预发布后缀（如 v0.2.0-rc1 按 0.2.0 计）。
func CompareVersions(a, b string) int {
	pa, pb := parseVersion(a), parseVersion(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			if pa[i] < pb[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func parseVersion(s string) [3]int {
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	parts := strings.Split(s, ".")
	var v [3]int
	for i := 0; i < 3 && i < len(parts); i++ {
		num := parts[i]
		if j := strings.IndexAny(num, "-+"); j >= 0 {
			num = num[:j]
		}
		v[i], _ = strconv.Atoi(num)
	}
	return v
}

// AssetName 按发布流水线（release.yml）的命名约定给出指定平台的资产文件名。
func AssetName(tag, goos, goarch string) string {
	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("deep-study-%s-%s-%s.%s", tag, goos, goarch, ext)
}

func assetURL(tag, name string) string {
	return fmt.Sprintf("%s/%s/%s", downloadBaseURL, tag, name)
}

// Latest 查询最新 release 的 tag 与说明。API 直连失败时降级：
// 经直连/镜像请求 releases/latest 的重定向，从 Location 提取 tag（此时 notes 为空）。
func Latest(ctx context.Context) (tag, notes string, err error) {
	client := &http.Client{Timeout: 15 * time.Second}
	if req, rerr := http.NewRequestWithContext(ctx, http.MethodGet, apiLatestURL, nil); rerr == nil {
		req.Header.Set("Accept", "application/vnd.github+json")
		if resp, derr := client.Do(req); derr == nil {
			var payload struct {
				TagName string `json:"tag_name"`
				Body    string `json:"body"`
			}
			ok := resp.StatusCode == http.StatusOK &&
				json.NewDecoder(resp.Body).Decode(&payload) == nil && payload.TagName != ""
			resp.Body.Close()
			if ok {
				return payload.TagName, payload.Body, nil
			}
		}
	}

	noRedirect := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	for _, prefix := range mirrorPrefixes() {
		req, rerr := http.NewRequestWithContext(ctx, http.MethodGet, prefix+htmlLatestURL, nil)
		if rerr != nil {
			continue
		}
		resp, derr := noRedirect.Do(req)
		if derr != nil {
			continue
		}
		loc := resp.Header.Get("Location")
		status := resp.StatusCode
		resp.Body.Close()
		if status != http.StatusFound && status != http.StatusMovedPermanently {
			continue
		}
		if t := path.Base(strings.TrimSuffix(loc, "/")); strings.HasPrefix(t, "v") {
			return t, "", nil
		}
	}
	return "", "", errors.New("无法连接 GitHub（直连与镜像均失败）")
}

// Download 下载指定版本的当前平台资产，SHA256 校验通过后解压出新二进制，
// 返回其临时路径（与当前可执行文件同目录）。校验文件缺失或校验不符都会报错。
func Download(ctx context.Context, tag string) (string, error) {
	exe, err := executablePath()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(exe)

	name := AssetName(tag, runtime.GOOS, runtime.GOARCH)
	base := assetURL(tag, name)

	sumPath, err := fetchToTemp(ctx, dir, "deep-study-sum-*", base+".sha256")
	if err != nil {
		return "", fmt.Errorf("下载校验文件失败: %w", err)
	}
	defer os.Remove(sumPath)
	sumData, err := os.ReadFile(sumPath)
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(sumData))
	if len(fields) == 0 {
		return "", errors.New("校验文件为空")
	}
	want := fields[0]

	pkgPath, err := fetchToTemp(ctx, dir, "deep-study-pkg-*", base)
	if err != nil {
		return "", fmt.Errorf("下载安装包失败: %w", err)
	}
	defer os.Remove(pkgPath)

	f, err := os.Open(pkgPath)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	_, copyErr := io.Copy(h, f)
	f.Close()
	if copyErr != nil {
		return "", copyErr
	}
	if got := hex.EncodeToString(h.Sum(nil)); !strings.EqualFold(got, want) {
		return "", fmt.Errorf("SHA256 校验不通过（期望 %s，实际 %s）", want, got)
	}

	return extract(pkgPath, dir)
}

// fetchToTemp 按「直连 → 镜像」顺序下载 URL 到 dir 下的临时文件，返回路径。
func fetchToTemp(ctx context.Context, dir, pattern, rawURL string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Minute}
	var lastErr error
	for _, prefix := range mirrorPrefixes() {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, prefix+rawURL, nil)
		if err != nil {
			lastErr = err
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			continue
		}
		tmp, err := os.CreateTemp(dir, pattern)
		if err != nil {
			resp.Body.Close()
			return "", err
		}
		_, copyErr := io.Copy(tmp, resp.Body)
		closeErr := tmp.Close()
		resp.Body.Close()
		if copyErr != nil || closeErr != nil {
			os.Remove(tmp.Name())
			lastErr = fmt.Errorf("写入临时文件失败: %v / %v", copyErr, closeErr)
			continue
		}
		return tmp.Name(), nil
	}
	if lastErr == nil {
		lastErr = errors.New("无可用的下载路径")
	}
	return "", lastErr
}

// extract 从 tar.gz / zip 包中取出二进制，写入 dir 下的临时文件（0755）。
func extract(pkgPath, dir string) (string, error) {
	binName := "deep-study"
	if runtime.GOOS == "windows" {
		binName = "deep-study.exe"
	}

	out, err := os.CreateTemp(dir, "deep-study-new-*")
	if err != nil {
		return "", err
	}
	ok := false
	defer func() {
		out.Close()
		if !ok {
			os.Remove(out.Name())
		}
	}()
	if err := out.Chmod(0o755); err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		r, err := zip.OpenReader(pkgPath)
		if err != nil {
			return "", err
		}
		defer r.Close()
		for _, f := range r.File {
			if path.Base(f.Name) != binName {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			_, copyErr := io.Copy(out, rc)
			rc.Close()
			if copyErr != nil {
				return "", copyErr
			}
			ok = true
			return out.Name(), nil
		}
		return "", fmt.Errorf("压缩包中未找到 %s", binName)
	}

	pkg, err := os.Open(pkgPath)
	if err != nil {
		return "", err
	}
	defer pkg.Close()
	gz, err := gzip.NewReader(pkg)
	if err != nil {
		return "", err
	}
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if hdr.Typeflag != tar.TypeReg || path.Base(hdr.Name) != binName {
			continue
		}
		if _, err := io.Copy(out, tr); err != nil {
			return "", err
		}
		ok = true
		return out.Name(), nil
	}
	return "", fmt.Errorf("压缩包中未找到 %s", binName)
}

// SwapBinary 原子替换当前可执行文件：旧文件改名为 .old 备份，失败自动回滚。
// Windows 允许改名（挪走）运行中的 exe，但不能覆盖，因此三平台统一用 rename。
func SwapBinary(newBin string) error {
	exe, err := executablePath()
	if err != nil {
		return err
	}
	old := exe + ".old"
	_ = os.Remove(old)
	if err := os.Rename(exe, old); err != nil {
		return fmt.Errorf("备份当前版本失败: %w", err)
	}
	if err := os.Rename(newBin, exe); err != nil {
		_ = os.Rename(old, exe)
		return fmt.Errorf("写入新版本失败（已回滚）: %w", err)
	}
	_ = os.Remove(old) // Windows 上旧文件可能仍被占用，删不掉无碍
	return nil
}

func executablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	return exe, nil
}
