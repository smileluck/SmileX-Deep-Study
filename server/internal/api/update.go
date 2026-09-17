package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"smilex-deep-study/server/internal/update"
)

// UpdateCheck 查询 GitHub 最新 release 并与当前版本比较。
// 网络失败也返回 200（带 error 与 manual_url），由前端提示手动下载。
func (a *API) UpdateCheck(c *gin.Context) {
	resp := gin.H{"current": a.Version, "has_update": false, "manual_url": update.ManualURL}
	if a.Version == "" || a.Version == "dev" {
		resp["reason"] = "dev build"
		c.JSON(http.StatusOK, resp)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	tag, notes, err := update.Latest(ctx)
	if err != nil {
		resp["error"] = err.Error()
		c.JSON(http.StatusOK, resp)
		return
	}
	resp["latest"] = tag
	resp["release_url"] = update.ManualURL
	if update.CompareVersions(a.Version, tag) < 0 {
		resp["has_update"] = true
		resp["notes"] = notes
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateApply 下载最新版本、SHA256 校验并原子替换当前二进制，重启后生效。
func (a *API) UpdateApply(c *gin.Context) {
	fail := func(msg string) {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": msg, "manual_url": update.ManualURL})
	}
	if a.Version == "" || a.Version == "dev" {
		fail("开发版本不支持自更新")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 6*time.Minute)
	defer cancel()
	tag, _, err := update.Latest(ctx)
	if err != nil {
		fail(err.Error())
		return
	}
	if update.CompareVersions(a.Version, tag) >= 0 {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "已是最新版本 " + tag})
		return
	}
	newBin, err := update.Download(ctx, tag)
	if err != nil {
		fail(err.Error())
		return
	}
	if err := update.SwapBinary(newBin); err != nil {
		fail(err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"need_restart":  true,
		"latest":        tag,
		"message":       "已更新到 " + tag + "，请重启 deep-study 生效",
	})
}
