// deep-study 服务入口：API + 内嵌 SPA 静态文件。
package main

import (
	"context"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"smilex-deep-study/server/internal/api"
	"smilex-deep-study/server/internal/scaffold"
	"smilex-deep-study/server/internal/store"
	web "smilex-deep-study/web"
)

// version 由发布流水线通过 -ldflags "-X main.version=<tag>" 注入，本地构建为 dev。
var version = "dev"

func main() {
	addr := flag.String("addr", "127.0.0.1:5574", "监听地址")
	dataDir := flag.String("data", "data", "数据目录（相对工作目录或绝对路径）")
	withScaffold := flag.Bool("scaffold", true, "启动时铺出 harness 工作区文件（AGENTS.md/.agents/.codebuddy/roles/prompts，只建缺失不覆盖）")
	showVersion := flag.Bool("version", false, "打印版本号并退出")
	flag.Parse()

	if *showVersion {
		println(version)
		return
	}

	data, workspace := resolveDirs(*dataDir, flagPassed("data"))

	st := store.New(data)
	if err := st.Ensure(); err != nil {
		log.Fatalf("初始化数据目录失败: %v", err)
	}

	if *withScaffold {
		created, skipped, err := scaffold.Ensure(workspace)
		if err != nil {
			log.Printf("铺出 harness 工作区文件失败（不影响服务）: %v", err)
		} else {
			log.Printf("harness 工作区: %s（新建 %d 个文件，跳过已存在 %d 个）", workspace, len(created), len(skipped))
		}
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api.Register(r, st)

	dist, err := fs.Sub(web.DistFS, "dist")
	if err != nil {
		log.Fatalf("内嵌前端资源异常: %v", err)
	}
	fileServer := http.FileServer(http.FS(dist))
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		if strings.HasPrefix(p, "/api/") {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		clean := strings.TrimPrefix(path.Clean("/"+p), "/")
		if clean == "" {
			clean = "index.html"
		}
		if f, err := dist.Open(clean); err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		index, err := fs.ReadFile(dist, "index.html")
		if err != nil {
			c.String(500, "index.html 缺失")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})

	srv := &http.Server{Addr: *addr, Handler: r}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		log.Printf("SmileX-Deep-Study 已启动: http://%s  (数据目录: %s)", *addr, data)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("优雅退出失败: %v", err)
	}
}

func flagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// resolveDirs 确定数据目录与工作区。显式传了 -data 时：数据目录原样使用，
// 工作区取其上一级。未传 -data 时检测当前目录：
//  1. 已是工作区（有 data/ 或 AGENTS.md）或只有二进制本身 → 原地作为工作区；
//  2. 已有 deepstudy/ 子目录 → 复用它；
//  3. 否则（非空目录）→ 新建 deepstudy/ 把工作区与数据收进去；
//     二进制本身留在当前目录，二次运行命令不变。
func resolveDirs(dataFlag string, explicit bool) (dataDir, workspace string) {
	if explicit {
		dataDir = dataFlag
		if abs, err := filepath.Abs(dataFlag); err == nil {
			workspace = filepath.Dir(abs)
		} else {
			workspace = dataFlag
		}
		return dataDir, workspace
	}

	cwd, err := os.Getwd()
	if err != nil {
		return dataFlag, "."
	}
	// /tmp 等符号链接路径统一成真实路径，保证与 exe 路径可比较
	if real, rerr := filepath.EvalSymlinks(cwd); rerr == nil {
		cwd = real
	}
	entries, _ := os.ReadDir(cwd)

	exe, _ := os.Executable()
	exe, _ = filepath.EvalSymlinks(exe)

	if has(entries, "data") || has(entries, "AGENTS.md") {
		return dataFlag, cwd
	}

	nested := filepath.Join(cwd, "deepstudy")
	if has(entries, "deepstudy") {
		return nestedData(nested), nested
	}

	meaningful := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if exe != "" && filepath.Join(cwd, e.Name()) == exe {
			continue
		}
		meaningful++
	}
	if meaningful == 0 {
		return dataFlag, cwd
	}

	if err := os.MkdirAll(nested, 0o755); err != nil {
		log.Printf("创建 deepstudy/ 失败，退回当前目录: %v", err)
		return dataFlag, cwd
	}
	log.Printf("检测到非空目录，工作区已收进 %s（二进制留在当前目录，二次运行命令不变）", nested)
	return nestedData(nested), nested
}

func nestedData(nested string) string { return filepath.Join(nested, "data") }

func has(entries []os.DirEntry, name string) bool {
	for _, e := range entries {
		if e.Name() == name {
			return true
		}
	}
	return false
}
