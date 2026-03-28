package server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/xfengyin/narutoscript-next/internal/app"
	"github.com/xfengyin/narutoscript-next/internal/automation"
)

var Version = "1.0.0"

type Server struct {
	engine     *gin.Engine
	app        *app.App
	scheduler  *automation.Scheduler
	wsUpgrader websocket.Upgrader
	wsClients  map[*websocket.Conn]bool
	wsMu       sync.RWMutex
}

func New(application *app.App, scheduler *automation.Scheduler) *Server {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	s := &Server{
		engine:     engine,
		app:        application,
		scheduler:  scheduler,
		wsUpgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		wsClients:  make(map[*websocket.Conn]bool),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	api := s.engine.Group("/api")
	{
		api.GET("/status", s.getStatus)
		api.GET("/version", s.getVersion)
		api.GET("/stats", s.getStats)
		api.GET("/tasks", s.getTasks)
		api.POST("/start", s.startTasks)
		api.POST("/stop", s.stopTasks)
		api.POST("/task/:name/run", s.runTask)
		api.GET("/logs", s.getLogs)
		api.GET("/logs/stream", s.streamLogs)
		api.GET("/device", s.getDevice)
		api.POST("/device/connect", s.connectDevice)
		api.GET("/device/screenshot", s.getScreenshot)
		api.GET("/config", s.getConfig)
	}

	s.engine.NoRoute(func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(indexHTML))
	})
}

func (s *Server) Start(addr string) error {
	return s.engine.Run(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.wsMu.Lock()
	defer s.wsMu.Unlock()
	for client := range s.wsClients {
		client.Close()
		delete(s.wsClients, client)
	}
	return nil
}

func (s *Server) getStatus(c *gin.Context) {
	state := s.app.GetState()
	c.JSON(http.StatusOK, gin.H{
		"running":      false,
		"device_name":  state.DeviceName,
		"device_ready": state.DeviceReady,
		"uptime":       time.Since(state.StartTime).Seconds(),
		"queue_length": s.scheduler.GetTaskQueueLength(),
		"tasks_done":   state.Stats.TasksDone,
		"tasks_total":  state.Stats.TasksTotal,
	})
}

func (s *Server) getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": Version, "go": "1.22"})
}

func (s *Server) getStats(c *gin.Context) {
	c.JSON(http.StatusOK, s.app.GetState().Stats)
}

func (s *Server) getTasks(c *gin.Context) {
	state := s.app.GetState()
	tasks := make([]gin.H, 0)
	for _, t := range state.Tasks {
		tasks = append(tasks, gin.H{"name": t.Name, "display": t.Display, "category": t.Category, "status": t.Status, "message": t.Message})
	}
	c.JSON(http.StatusOK, tasks)
}

func (s *Server) startTasks(c *gin.Context) {
	if err := s.app.ConnectDevice(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "设备未连接", "message": err.Error()})
		return
	}
	go s.scheduler.Run()
	c.JSON(http.StatusOK, gin.H{"message": "任务已启动"})
}

func (s *Server) stopTasks(c *gin.Context) {
	s.scheduler.Stop()
	c.JSON(http.StatusOK, gin.H{"message": "任务已停止"})
}

func (s *Server) runTask(c *gin.Context) {
	taskName := c.Param("name")
	if !s.app.GetState().DeviceReady {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "设备未连接"})
		return
	}
	s.scheduler.RunTaskNow(taskName)
	c.JSON(http.StatusOK, gin.H{"message": "任务已入队", "task": taskName})
}

func (s *Server) getLogs(c *gin.Context) {
	c.JSON(http.StatusOK, s.app.GetLogs(100))
}

func (s *Server) streamLogs(c *gin.Context) {
	conn, err := s.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	s.wsMu.Lock()
	s.wsClients[conn] = true
	s.wsMu.Unlock()

	defer func() {
		s.wsMu.Lock()
		delete(s.wsClients, conn)
		s.wsMu.Unlock()
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		state := s.app.GetState()
		if err := conn.WriteJSON(gin.H{"type": "state", "data": state}); err != nil {
			return
		}
	}
}

func (s *Server) getDevice(c *gin.Context) {
	state := s.app.GetState()
	c.JSON(http.StatusOK, gin.H{"name": state.DeviceName, "ready": state.DeviceReady})
}

func (s *Server) connectDevice(c *gin.Context) {
	if err := s.app.ConnectDevice(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "设备已连接"})
}

func (s *Server) getScreenshot(c *gin.Context) {
	if !s.app.GetState().DeviceReady {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "设备未连接"})
		return
	}
	screen, err := s.app.Device.Screenshot()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "image/png", screen)
}

func (s *Server) getConfig(c *gin.Context) {
	c.JSON(http.StatusOK, s.app.Config)
}

const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NarutoScript Next</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%); min-height: 100vh; color: white; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        header { text-align: center; padding: 40px 0; }
        .logo { font-size: 60px; margin-bottom: 10px; }
        h1 { font-size: 28px; margin-bottom: 5px; }
        .subtitle { color: #888; }
        .stats { display: grid; grid-template-columns: repeat(4, 1fr); gap: 15px; margin: 30px 0; }
        .stat-card { background: rgba(255,255,255,0.1); padding: 20px; border-radius: 10px; text-align: center; }
        .stat-value { font-size: 24px; font-weight: bold; color: #ff6b35; }
        .stat-label { color: #888; font-size: 14px; margin-top: 5px; }
        .section { margin: 30px 0; }
        .section-title { font-size: 18px; margin-bottom: 15px; border-bottom: 1px solid #333; padding-bottom: 10px; }
        .tasks { display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 10px; }
        .task { background: rgba(255,255,255,0.05); padding: 15px; border-radius: 8px; display: flex; justify-content: space-between; align-items: center; }
        .task-name { color: #ccc; }
        .task-status { font-size: 12px; padding: 4px 10px; border-radius: 20px; }
        .status-idle { background: #333; color: #888; }
        .status-running { background: #1e40af; color: #93c5fd; }
        .status-success { background: #166534; color: #86efac; }
        .status-failed { background: #991b1b; color: #fca5a5; }
        .actions { display: flex; gap: 10px; margin: 30px 0; }
        button { padding: 12px 24px; border: none; border-radius: 8px; cursor: pointer; font-size: 14px; font-weight: 500; }
        .btn-primary { background: #ff6b35; color: white; }
        .btn-secondary { background: #333; color: #888; }
        .device-status { display: flex; align-items: center; gap: 10px; margin-bottom: 20px; }
        .dot { width: 10px; height: 10px; border-radius: 50%; }
        .dot-green { background: #22c55e; }
        .dot-red { background: #ef4444; }
        footer { text-align: center; padding: 30px; color: #555; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <div class="logo">🍥</div>
            <h1>NarutoScript Next</h1>
            <p class="subtitle">火影忍者手游自动化工具</p>
        </header>
        <div class="device-status">
            <div class="dot" id="deviceDot"></div>
            <span id="deviceText">检测设备中...</span>
        </div>
        <div class="stats" id="stats">
            <div class="stat-card"><div class="stat-value" id="gold">-</div><div class="stat-label">金币</div></div>
            <div class="stat-card"><div class="stat-value" id="copper">-</div><div class="stat-label">铜币</div></div>
            <div class="stat-card"><div class="stat-value" id="stamina">-</div><div class="stat-label">体力</div></div>
            <div class="stat-card"><div class="stat-value" id="tasks">0/25</div><div class="stat-label">任务完成</div></div>
        </div>
        <div class="actions">
            <button class="btn-primary" onclick="connectDevice()">连接设备</button>
            <button class="btn-secondary" onclick="loadTasks()">刷新状态</button>
        </div>
        <div class="section"><div class="section-title">日常任务</div><div class="tasks" id="dailyTasks"></div></div>
        <div class="section"><div class="section-title">收获系统</div><div class="tasks" id="harvestTasks"></div></div>
        <div class="section"><div class="section-title">周常任务</div><div class="tasks" id="weeklyTasks"></div></div>
    </div>
    <footer>NarutoScript Next v1.0.0</footer>
    <script>
        async function loadStatus() { try { const r = await fetch('/api/status'); const d = await r.json(); document.getElementById('deviceDot').className = 'dot ' + (d.device_ready ? 'dot-green' : 'dot-red'); document.getElementById('deviceText').textContent = d.device_ready ? '设备已连接: ' + d.device_name : '设备未连接'; document.getElementById('tasks').textContent = d.tasks_done + '/' + d.tasks_total; } catch(e) {} }
        async function loadTasks() { try { const r = await fetch('/api/tasks'); const t = await r.json(); const c = {daily: document.getElementById('dailyTasks'), harvest: document.getElementById('harvestTasks'), weekly: document.getElementById('weeklyTasks')}; for (const k in c) c[k].innerHTML = ''; t.forEach(x => { const d = document.createElement('div'); d.className = 'task'; d.innerHTML = '<span class="task-name">'+x.display+'</span><span class="task-status status-'+x.status+'">'+x.status+'</span>'; c[x.category]?.appendChild(d); }); } catch(e) {} }
        async function connectDevice() { try { const r = await fetch('/api/device/connect', {method: 'POST'}); const d = await r.json(); alert(r.ok ? '设备已连接' : '连接失败: ' + d.error); loadStatus(); } catch(e) { alert('连接失败'); } }
        loadStatus(); loadTasks(); setInterval(loadStatus, 2000);
    </script>
</body>
</html>`