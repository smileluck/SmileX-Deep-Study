package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// ---------- 学习计划 ----------

func (a *API) topicNames() map[string]string {
	topics, _ := a.Store.ListTopics()
	names := map[string]string{}
	for _, t := range topics {
		names[str(t["slug"])] = str(t["name"])
	}
	return names
}

func (a *API) GetPlans(c *gin.Context) {
	master, err := a.Store.MasterPlan()
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	plans, err := a.Store.TopicPlans()
	if err != nil {
		errJSON(c, 500, err)
		return
	}
	names := a.topicNames()
	topics := make([]map[string]any, 0, len(plans))
	for _, p := range plans {
		name := names[p.Slug]
		if name == "" {
			name = p.Slug
		}
		topics = append(topics, map[string]any{
			"slug": p.Slug, "topic_name": name, "fm": p.FM, "body": p.Body,
		})
	}
	var masterJSON any
	if master != nil {
		masterJSON = map[string]any{"fm": master.FM, "body": master.Body}
	}
	c.JSON(200, gin.H{"master": masterJSON, "topics": topics})
}

func (a *API) GetTopicPlan(c *gin.Context) {
	slug := c.Param("slug")
	plan, err := a.Store.GetTopicPlan(slug)
	if err != nil {
		errJSON(c, 404, fmt.Errorf("计划不存在: %s", slug))
		return
	}
	name := a.topicNames()[slug]
	if name == "" {
		name = slug
	}
	c.JSON(200, gin.H{"slug": slug, "topic_name": name, "fm": plan.FM, "body": plan.Body})
}
