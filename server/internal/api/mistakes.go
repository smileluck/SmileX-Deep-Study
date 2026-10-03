package api

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"smilex-deep-study/server/internal/store"
)

type mistakeTopicCount struct {
	Active   int `json:"active"`
	Mastered int `json:"mastered"`
}

func (a *API) ListMistakes(c *gin.Context) {
	mistakes, err := a.Store.ListMistakes()
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	out := make([]map[string]any, 0, len(mistakes))
	active, mastered := 0, 0
	byTopic := map[string]*mistakeTopicCount{}
	for _, m := range mistakes {
		out = append(out, m.FM)
		topic := str(m.FM["topic"])
		tc := byTopic[topic]
		if tc == nil {
			tc = &mistakeTopicCount{}
			byTopic[topic] = tc
		}
		switch str(m.FM["status"]) {
		case "mastered":
			mastered++
			tc.Mastered++
		default:
			active++
			tc.Active++
		}
	}
	c.JSON(200, gin.H{"mistakes": out, "summary": gin.H{
		"active": active, "mastered": mastered, "by_topic": byTopic,
	}})
}

type mistakeStatusReq struct {
	Status string `json:"status" binding:"required"`
}

// GetMistakeAsset 提供错题原图（data/mistakes/assets/ 下的一级文件名），
// 路径穿越（../、子目录）一律 400，不存在 404。
func (a *API) GetMistakeAsset(c *gin.Context) {
	name := c.Query("name")
	if name == "" || filepath.Base(name) != name || strings.ContainsAny(name, `/\`) {
		errJSON(c, 400, fmt.Errorf("name 必须是 assets 内的文件名（不含路径）"))
		return
	}
	path := filepath.Join(a.Store.MistakeAssetsDir(), name)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		errJSON(c, 404, fmt.Errorf("assets 中不存在 %s", name))
		return
	}
	c.File(path)
}

func (a *API) SetMistakeStatus(c *gin.Context) {
	var req mistakeStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		errJSON(c, 400, err)
		return
	}
	if req.Status != "active" && req.Status != "mastered" {
		errJSON(c, 400, fmt.Errorf("status 只接受 active/mastered: %s", req.Status))
		return
	}
	id := c.Param("id")
	a.Store.Lock()
	defer a.Store.Unlock()
	var masteredAt *string
	if req.Status == "mastered" {
		d := time.Now().Format("2006-01-02")
		masteredAt = &d
	}
	if err := a.Store.SetMistakeStatus(id, req.Status, masteredAt); err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidID):
			errJSON(c, 400, err)
		case errors.Is(err, store.ErrNotFound):
			errJSON(c, 404, err)
		default:
			errJSON(c, 500, err)
		}
		return
	}
	m, err := a.Store.GetMistake(id)
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	c.JSON(200, m.FM)
}
