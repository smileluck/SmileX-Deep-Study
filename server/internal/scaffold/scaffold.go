// Package scaffold 内嵌 harness 工作区脚手架文件（AGENTS.md、.agents/、
// .codebuddy/、roles/、prompts/），使打包后的单二进制在任意目录启动时，
// 能把这些契约文件铺到数据目录上一级，供 WorkBuddy 等 agent harness 使用。
//
// assets/ 由 scripts/sync-adapters.sh 从仓库根生成（make build/cross 自动调用），
// 勿手改。因 go:embed 不接受点开头路径，assets 内用 agents/ codebuddy/ 目录名，
// 写盘时还原为 .agents/ .codebuddy/。
package scaffold

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all:assets
var assetsFS embed.FS

var dotDirs = map[string]string{"agents": ".agents", "codebuddy": ".codebuddy"}

// Ensure 把内嵌脚手架写入 workspace，只创建缺失文件，绝不覆盖已存在文件。
// 返回新建与跳过的相对路径列表。
func Ensure(workspace string) (created, skipped []string, err error) {
	err = fs.WalkDir(assetsFS, "assets", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(p, "assets")
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			return nil
		}
		parts := strings.SplitN(rel, "/", 2)
		if dot, ok := dotDirs[parts[0]]; ok {
			parts[0] = dot
			rel = strings.Join(parts, "/")
		}
		dst := filepath.Join(workspace, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		if _, err := os.Stat(dst); err == nil {
			skipped = append(skipped, rel)
			return nil
		}
		data, err := assetsFS.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return err
		}
		created = append(created, rel)
		return nil
	})
	return created, skipped, err
}
