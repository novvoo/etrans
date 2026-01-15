package handlers

import (
	"etrans/middleware"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type SessionInfo struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	TaskCount int       `json:"taskCount"`
}

// GetSessionsHandler 获取所有可用会话
func GetSessionsHandler(c *gin.Context) {
	usersDir := filepath.Join("data", "users")
	entries, err := os.ReadDir(usersDir)
	if err != nil {
		// 如果目录不存在，可能还没有任何会话
		if os.IsNotExist(err) {
			c.JSON(http.StatusOK, gin.H{
				"currentSessionId": middleware.GetSessionID(c),
				"sessions":         []SessionInfo{},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取会话列表"})
		return
	}

	var sessions []SessionInfo
	currentSessionID := middleware.GetSessionID(c)

	for _, entry := range entries {
		if entry.IsDir() {
			info, err := entry.Info()
			createdAt := time.Now()
			if err == nil {
				createdAt = info.ModTime() // Use ModTime as proxy
			}

			// 简单的任务计数
			taskCount := 0
			tasksDir := filepath.Join(usersDir, entry.Name(), "tasks")
			if taskEntries, err := os.ReadDir(tasksDir); err == nil {
				for _, te := range taskEntries {
					if filepath.Ext(te.Name()) == ".json" {
						taskCount++
					}
				}
			}

			sessions = append(sessions, SessionInfo{
				ID:        entry.Name(),
				CreatedAt: createdAt,
				TaskCount: taskCount,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"currentSessionId": currentSessionID,
		"sessions":         sessions,
	})
}

// SwitchSessionHandler 切换会话
func SwitchSessionHandler(c *gin.Context) {
	var req struct {
		TargetSessionID string `json:"targetSessionId"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效请求"})
		return
	}

	if req.TargetSessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "目标会话ID不能为空"})
		return
	}

	// 验证会话是否存在
	userDir := filepath.Join("data", "users", req.TargetSessionID)
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "目标会话不存在"})
		return
	}

	// 设置 Cookie
	// 根据请求协议设置secure标志
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetCookie(
		middleware.SessionCookieName,
		req.TargetSessionID,
		int(middleware.SessionTimeout.Seconds()),
		"/",
		"",
		isSecure,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"message": "会话已切换", "sessionId": req.TargetSessionID})
}
