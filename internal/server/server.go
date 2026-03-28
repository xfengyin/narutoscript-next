package server

import (
	"context"
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xfengyin/narutoscript-next/internal/app"
)

type Server struct {
	engine   *gin.Engine
	app      *app.App
	staticFS embed.FS
}

func New(application *app.App, staticFS embed.FS) *Server {
	gin.SetMode(gin.ReleaseMode)
	
	engine := gin.New()
	engine.Use(gin.Recovery())

	s := &Server{
		engine:   engine,
		app:      application,
		staticFS: staticFS,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// API 路由
	api := s.engine.Group("/api")
	{
		api.GET("/status", s.getStatus)
		api.GET("/stats", s.getStats)
		api.POST("/start", s.startTasks)
		api.POST("/stop", s.stopTasks)
		api.POST("/task/:name/run", s.runTask)
		api.GET("/tasks", s.getTasks)
		api.GET("/logs", s.getLogs)
		api.GET("/logs/stream", s.streamLogs)
	}

	// 静态文件服务
	s.serveStatic()
}

func (s *Server) serveStatic() {
	// 从 embed.FS 获取静态文件
	distFS, err := fs.Sub(s.staticFS, "internal/ui/dist")
	if err != nil {
		// 如果没有构建前端，返回简单的 HTML
		s.engine.NoRoute(func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"title": "NarutoScript Next",
			})
		})
		return
	}

	// 静态文件
	s.engine.StaticFS("/assets", http.FS(distFS))

	// 所有其他路由返回 index.html (SPA)
	s.engine.NoRoute(func(c *gin.Context) {
		c.FileFromFS("index.html", http.FS(distFS))
	})
}

func (s *Server) Start(addr string) error {
	return s.engine.Run(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

// API Handlers

func (s *Server) getStatus(c *gin.Context) {
	state := s.app.GetState()
	c.JSON(http.StatusOK, gin.H{
		"running": state.Running,
		"version": "1.0.0",
	})
}

func (s *Server) getStats(c *gin.Context) {
	state := s.app.GetState()
	c.JSON(http.StatusOK, state.Stats)
}

func (s *Server) startTasks(c *gin.Context) {
	if err := s.app.StartScheduler(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务已启动"})
}

func (s *Server) stopTasks(c *gin.Context) {
	s.app.StopScheduler()
	c.JSON(http.StatusOK, gin.H{"message": "任务已停止"})
}

func (s *Server) runTask(c *gin.Context) {
	taskName := c.Param("name")
	// TODO: 执行指定任务
	c.JSON(http.StatusOK, gin.H{"message": "任务已启动", "task": taskName})
}

func (s *Server) getTasks(c *gin.Context) {
	tasks := []gin.H{
		// 日常任务
		{"name": "team_raid", "display": "小队突袭", "category": "daily", "status": "idle"},
		{"name": "bounty_hall", "display": "丰饶之间", "category": "daily", "status": "idle"},
		{"name": "guild_pray", "display": "组织祈福", "category": "daily", "status": "idle"},
		{"name": "survival", "display": "生存试炼", "category": "daily", "status": "idle"},
		{"name": "equipment", "display": "装备扫荡", "category": "daily", "status": "idle"},
		{"name": "task_hall", "display": "任务集会所", "category": "daily", "status": "idle"},
		{"name": "secret_realm", "display": "秘境挑战", "category": "daily", "status": "idle"},
		{"name": "shop", "display": "商店购买", "category": "daily", "status": "idle"},
		// 收获系统
		{"name": "lucky_money", "display": "招财", "category": "harvest", "status": "idle"},
		{"name": "intel_agency", "display": "情报社", "category": "harvest", "status": "idle"},
		{"name": "ranking", "display": "排行榜", "category": "harvest", "status": "idle"},
		{"name": "activity_box", "display": "活跃度宝箱", "category": "harvest", "status": "idle"},
		{"name": "monthly_sign", "display": "每月签到", "category": "harvest", "status": "idle"},
		{"name": "ninja_pass", "display": "忍法帖", "category": "harvest", "status": "idle"},
		{"name": "share", "display": "每日分享", "category": "harvest", "status": "idle"},
		{"name": "mail", "display": "邮件", "category": "harvest", "status": "idle"},
		{"name": "stamina_gift", "display": "赠送体力", "category": "harvest", "status": "idle"},
		{"name": "recruit", "display": "招募", "category": "harvest", "status": "idle"},
		// 周常任务
		{"name": "practice", "display": "修行之路", "category": "weekly", "status": "idle"},
		{"name": "chase_akatsuki", "display": "追击晓组织", "category": "weekly", "status": "idle"},
		{"name": "rebel", "display": "叛忍来袭", "category": "weekly", "status": "idle"},
		{"name": "guild_fortress", "display": "组织要塞", "category": "weekly", "status": "idle"},
		{"name": "battlefield", "display": "天地战场", "category": "weekly", "status": "idle"},
		// 限时活动
		{"name": "bbq", "display": "丁次烤肉", "category": "event", "status": "idle"},
		{"name": "sakura", "display": "樱花季", "category": "event", "status": "idle"},
	}
	c.JSON(http.StatusOK, tasks)
}

func (s *Server) getLogs(c *gin.Context) {
	logs := []gin.H{
		{"time": "14:30:01", "level": "info", "message": "服务已启动"},
	}
	c.JSON(http.StatusOK, logs)
}