package server

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/xfengyin/narutoscript-next/internal/app"
)

// Version 版本信息
var Version = "1.0.0"

// Server HTTP 服务器
type Server struct {
	engine     *gin.Engine
	app        *app.App
	staticFS   embed.FS
	wsUpgrader websocket.Upgrader
	wsClients  map[*websocket.Conn]bool
}

// New 创建服务器实例
func New(application *app.App, staticFS embed.FS) *Server {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Recovery())

	// CORS 配置
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	s := &Server{
		engine:   engine,
		app:      application,
		staticFS: staticFS,
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		wsClients: make(map[*websocket.Conn]bool),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// API 路由
	api := s.engine.Group("/api")
	{
		// 状态
		api.GET("/status", s.getStatus)
		api.GET("/version", s.getVersion)
		api.GET("/stats", s.getStats)

		// 任务
		api.GET("/tasks", s.getTasks)
		api.POST("/start", s.startTasks)
		api.POST("/stop", s.stopTasks)
		api.POST("/task/:name/run", s.runTask)

		// 日志
		api.GET("/logs", s.getLogs)
		api.GET("/logs/stream", s.streamLogs)

		// 设备
		api.GET("/device", s.getDevice)
		api.POST("/device/connect", s.connectDevice)

		// 配置
		api.GET("/config", s.getConfig)
		api.PUT("/config", s.updateConfig)
	}

	// 静态文件服务
	s.serveStatic()
}

func (s *Server) serveStatic() {
	distFS, err := fs.Sub(s.staticFS, "internal/ui/dist")
	if err != nil {
		s.engine.GET("/", func(c *gin.Context) {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(indexHTML))
		})
		return
	}

	s.engine.GET("/assets/*filepath", func(c *gin.Context) {
		c.FileFromFS(c.Request.URL.Path, http.FS(distFS))
	})

	s.engine.NoRoute(func(c *gin.Context) {
		c.FileFromFS("index.html", http.FS(distFS))
	})
}

func (s *Server) Start(addr string) error {
	return s.engine.Run(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	// 关闭所有 WebSocket 连接
	for client := range s.wsClients {
		client.Close()
	}
	return nil
}

// ===== API Handlers =====

func (s *Server) getStatus(c *gin.Context) {
	state := s.app.GetState()
	c.JSON(http.StatusOK, gin.H{
		"running":      state.Running,
		"device_name":  state.DeviceName,
		"device_ready": state.DeviceReady,
		"uptime":       time.Since(state.StartTime).Seconds(),
	})
}

func (s *Server) getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": Version,
		"go":      "1.22",
	})
}

func (s *Server) getStats(c *gin.Context) {
	state := s.app.GetState()
	c.JSON(http.StatusOK, state.Stats)
}

func (s *Server) getTasks(c *gin.Context) {
	state := s.app.GetState()
	category := c.Query("category")

	tasks := make([]gin.H, 0)
	for _, task := range state.Tasks {
		if category == "" || task.Category == category {
			tasks = append(tasks, gin.H{
				"name":       task.Name,
				"display":    task.Display,
				"category":   task.Category,
				"status":     task.Status,
				"message":    task.Message,
				"progress":   task.Progress,
				"duration":   task.Duration,
				"retry_count": task.RetryCount,
			})
		}
	}

	c.JSON(http.StatusOK, tasks)
}

func (s *Server) startTasks(c *gin.Context) {
	if err := s.app.ConnectDevice(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "设备未连接",
			"message": err.Error(),
		})
		return
	}

	if err := s.app.StartScheduler(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "任务已启动",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func (s *Server) stopTasks(c *gin.Context) {
	s.app.StopScheduler()
	c.JSON(http.StatusOK, gin.H{
		"message": "任务已停止",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func (s *Server) runTask(c *gin.Context) {
	taskName := c.Param("name")

	state := s.app.GetState()
	if _, ok := state.Tasks[taskName]; !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	if !state.DeviceReady {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "设备未连接"})
		return
	}

	// 更新任务状态
	s.app.UpdateTaskState(taskName, "running", "正在执行...")
	s.app.AddLog("info", "开始执行任务: "+taskName, taskName)

	c.JSON(http.StatusOK, gin.H{
		"message": "任务已启动",
		"task":    taskName,
		"time":    time.Now().Format(time.RFC3339),
	})
}

func (s *Server) getLogs(c *gin.Context) {
	limit := 100
	if l := c.Query("limit"); l != "" {
		if n, err := parseInt(l); err == nil && n > 0 {
			limit = n
		}
	}

	logs := s.app.GetLogs(limit)
	c.JSON(http.StatusOK, logs)
}

func (s *Server) streamLogs(c *gin.Context) {
	conn, err := s.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	s.wsClients[conn] = true
	defer delete(s.wsClients, conn)

	// 发送初始状态
	state := s.app.GetState()
	conn.WriteJSON(gin.H{
		"type": "state",
		"data": state,
	})

	// 定期推送状态更新
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			state := s.app.GetState()
			err := conn.WriteJSON(gin.H{
				"type": "state",
				"data": state,
			})
			if err != nil {
				return
			}
		}
	}
}

func (s *Server) getDevice(c *gin.Context) {
	state := s.app.GetState()
	c.JSON(http.StatusOK, gin.H{
		"name":    state.DeviceName,
		"ready":   state.DeviceReady,
		"width":   s.app.Config.Device.ScreenWidth,
		"height":  s.app.Config.Device.ScreenHeight,
	})
}

func (s *Server) connectDevice(c *gin.Context) {
	if err := s.app.ConnectDevice(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "连接失败",
			"message": err.Error(),
		})
		return
	}

	state := s.app.GetState()
	c.JSON(http.StatusOK, gin.H{
		"message":     "设备已连接",
		"device_name": state.DeviceName,
	})
}

func (s *Server) getConfig(c *gin.Context) {
	c.JSON(http.StatusOK, s.app.Config)
}

func (s *Server) updateConfig(c *gin.Context) {
	// TODO: 实现配置更新
	c.JSON(http.StatusOK, gin.H{"message": "配置已更新"})
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// 回退 HTML
const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NarutoScript Next</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
        }
        .container { text-align: center; padding: 40px; }
        .logo { font-size: 80px; margin-bottom: 20px; }
        h1 { font-size: 32px; margin-bottom: 10px; }
        p { color: #888; margin-bottom: 30px; }
        .info { background: rgba(255,255,255,0.1); padding: 20px; border-radius: 10px; }
        code { color: #ff6b35; }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">🍥</div>
        <h1>NarutoScript Next</h1>
        <p>前端未构建，请运行构建命令</p>
        <div class="info">
            <code>cd web && npm install && npm run build</code>
        </div>
    </div>
</body>
</html>`