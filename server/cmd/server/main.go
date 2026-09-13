// deep-study 服务入口：API + 内嵌 SPA 静态文件。
package main

import (
	"flag"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"

	"smilex-deep-study/server/internal/api"
	"smilex-deep-study/server/internal/store"
	web "smilex-deep-study/web"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8788", "监听地址")
	dataDir := flag.String("data", "data", "数据目录（相对工作目录或绝对路径）")
	flag.Parse()

	st := store.New(*dataDir)
	if err := st.Ensure(); err != nil {
		log.Fatalf("初始化数据目录失败: %v", err)
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

	log.Printf("SmileX-Deep-Study 已启动: http://%s  (数据目录: %s)", *addr, *dataDir)
	if err := r.Run(*addr); err != nil {
		log.Fatal(err)
	}
}
