package app

import (
	"context"
	"sync"
	"time"

	"github.com/xfengyin/narutoscript-next/internal/automation"
	"github.com/xfengyin/narutoscript-next/internal/device"
	"github.com/xfengyin/narutoscript-next/internal/vision"
	"github.com/xfengyin/narutoscript-next/pkg/logger"
)

// Version 版本信息
var Version = "1.0.0"

// App 应用主结构
type App struct {
	Config    *Config
	Logger    *logger.Logger
	Device    *device.Controller
	Vision    *vision.Matcher
	Scheduler *automation.Scheduler
	State     *AppState
	logStore  *LogStore
	ctx       context.Context
	cancel    context.CancelFunc
}

// AppState 应用状态
type AppState struct {
	mu          sync.RWMutex
	Running     bool
	DeviceName  string
	DeviceReady bool
	Tasks       map[string]*TaskState
	Stats       *GameStats
	StartTime   time.Time
}

// TaskState 任务状态
type TaskState struct {
	Name       string    `json:"name"`
	Display    string    `json:"display"`
	Category   string    `json:"category"`
	Status     string    `json:"status"` // idle, running, success, failed, waiting
	Progress   int       `json:"progress"`
	Message    string    `json:"message"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Duration   int64     `json:"duration_ms"`
	RetryCount int       `json:"retry_count"`
}

// GameStats 游戏数据
type GameStats struct {
	Gold       int64 `json:"gold"`
	Copper     int64 `json:"copper"`
	Stamina    int   `json:"stamina"`
	MaxStamina int   `json:"max_stamina"`
	TasksDone  int   `json:"tasks_done"`
	TasksTotal int   `json:"tasks_total"`
	LastUpdate time.Time `json:"last_update"`
}

// LogStore 日志存储
type LogStore struct {
	mu      sync.RWMutex
	logs    []LogEntry
	maxSize int
}

// LogEntry 日志条目
type LogEntry struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
	Task    string    `json:"task,omitempty"`
}

// NewLogStore 创建日志存储
func NewLogStore(maxSize int) *LogStore {
	return &LogStore{
		logs:    make([]LogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add 添加日志
func (s *LogStore) Add(level, message, task string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Message: message,
		Task:    task,
	}

	s.logs = append(s.logs, entry)

	// 超出最大数量时删除旧日志
	if len(s.logs) > s.maxSize {
		s.logs = s.logs[len(s.logs)-s.maxSize:]
	}
}

// Get 获取日志
func (s *LogStore) Get(limit int) []LogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.logs) {
		limit = len(s.logs)
	}

	// 返回最新的日志
	start := len(s.logs) - limit
	if start < 0 {
		start = 0
	}

	result := make([]LogEntry, limit)
	copy(result, s.logs[start:])
	return result
}

// New 创建应用实例
func New(cfg *Config, log *logger.Logger) *App {
	ctx, cancel := context.WithCancel(context.Background())

	state := &AppState{
		Running:   false,
		Tasks:     make(map[string]*TaskState),
		Stats:     &GameStats{MaxStamina: 180},
		StartTime: time.Now(),
	}

	// 初始化任务状态
	initTaskStates(state)

	return &App{
		Config:   cfg,
		Logger:   log,
		Device:   device.NewController(cfg.Device, log),
		Vision:   vision.NewMatcher(log),
		State:    state,
		logStore: NewLogStore(1000),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// initTaskStates 初始化任务状态
func initTaskStates(state *AppState) {
	// 日常任务
	dailyTasks := []struct{ name, display string }{
		{"team_raid", "小队突袭"},
		{"bounty_hall", "丰饶之间"},
		{"guild_pray", "组织祈福"},
		{"survival", "生存试炼"},
		{"equipment", "装备扫荡"},
		{"task_hall", "任务集会所"},
		{"secret_realm", "秘境挑战"},
		{"shop", "商店购买"},
	}

	// 收获任务
	harvestTasks := []struct{ name, display string }{
		{"lucky_money", "招财"},
		{"intel_agency", "情报社"},
		{"ranking", "排行榜"},
		{"activity_box", "活跃度宝箱"},
		{"monthly_sign", "每月签到"},
		{"ninja_pass", "忍法帖"},
		{"share", "每日分享"},
		{"mail", "邮件"},
		{"stamina_gift", "赠送体力"},
		{"recruit", "招募"},
	}

	// 周常任务
	weeklyTasks := []struct{ name, display string }{
		{"practice", "修行之路"},
		{"chase_akatsuki", "追击晓组织"},
		{"rebel", "叛忍来袭"},
		{"guild_fortress", "组织要塞"},
		{"battlefield", "天地战场"},
	}

	// 活动任务
	eventTasks := []struct{ name, display string }{
		{"bbq", "丁次烤肉"},
		{"sakura", "樱花季"},
	}

	allTasks := make([]struct {
		name, display, category string
	}, 0)

	for _, t := range dailyTasks {
		allTasks = append(allTasks, struct {
			name, display, category string
		}{t.name, t.display, "daily"})
	}
	for _, t := range harvestTasks {
		allTasks = append(allTasks, struct {
			name, display, category string
		}{t.name, t.display, "harvest"})
	}
	for _, t := range weeklyTasks {
		allTasks = append(allTasks, struct {
			name, display, category string
		}{t.name, t.display, "weekly"})
	}
	for _, t := range eventTasks {
		allTasks = append(allTasks, struct {
			name, display, category string
		}{t.name, t.display, "event"})
	}

	state.Stats.TasksTotal = len(allTasks)

	for _, t := range allTasks {
		state.Tasks[t.name] = &TaskState{
			Name:     t.name,
			Display:  t.display,
			Category: t.category,
			Status:   "idle",
		}
	}
}

// ConnectDevice 连接设备
func (a *App) ConnectDevice() error {
	if err := a.Device.Connect(); err != nil {
		a.State.mu.Lock()
		a.State.DeviceReady = false
		a.State.mu.Unlock()
		return err
	}

	a.State.mu.Lock()
	a.State.DeviceName = a.Device.GetDeviceName()
	a.State.DeviceReady = true
	a.State.mu.Unlock()

	a.logStore.Add("info", "设备已连接: "+a.Device.GetDeviceName(), "")
	return nil
}

// StartScheduler 启动调度器
func (a *App) StartScheduler() error {
	a.State.mu.Lock()
	defer a.State.mu.Unlock()

	if a.State.Running {
		return nil
	}

	a.Scheduler = automation.NewScheduler(a.Config.Automation, a.Logger)
	a.State.Running = true
	a.logStore.Add("info", "任务调度器已启动", "")

	go a.Scheduler.Run()

	return nil
}

// StopScheduler 停止调度器
func (a *App) StopScheduler() {
	a.State.mu.Lock()
	defer a.State.mu.Unlock()

	if a.Scheduler != nil {
		a.Scheduler.Stop()
	}
	a.State.Running = false
	a.logStore.Add("info", "任务调度器已停止", "")
}

// GetState 获取状态
func (a *App) GetState() *AppState {
	a.State.mu.RLock()
	defer a.State.mu.RUnlock()

	// 返回副本
	tasks := make(map[string]*TaskState)
	for k, v := range a.State.Tasks {
		taskCopy := *v
		tasks[k] = &taskCopy
	}

	statsCopy := *a.State.Stats

	return &AppState{
		Running:     a.State.Running,
		DeviceName:  a.State.DeviceName,
		DeviceReady: a.State.DeviceReady,
		Tasks:       tasks,
		Stats:       &statsCopy,
		StartTime:   a.State.StartTime,
	}
}

// UpdateTaskState 更新任务状态
func (a *App) UpdateTaskState(name string, status, message string) {
	a.State.mu.Lock()
	defer a.State.mu.Unlock()

	if task, ok := a.State.Tasks[name]; ok {
		task.Status = status
		task.Message = message

		if status == "running" {
			task.StartTime = time.Now()
		} else if status == "success" || status == "failed" {
			task.EndTime = time.Now()
			task.Duration = task.EndTime.Sub(task.StartTime).Milliseconds()
		}
	}
}

// GetLogs 获取日志
func (a *App) GetLogs(limit int) []LogEntry {
	return a.logStore.Get(limit)
}

// AddLog 添加日志
func (a *App) AddLog(level, message, task string) {
	a.logStore.Add(level, message, task)
}