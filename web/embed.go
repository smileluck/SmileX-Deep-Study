// Package web 内嵌前端构建产物（dist/），
// 使 go build 产出单文件部署二进制。
package web

import "embed"

//go:embed all:dist
var DistFS embed.FS
