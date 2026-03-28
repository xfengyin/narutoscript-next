package app

import (
	"context"
	"sync"
	"time"

	"github.com/xfengyin/narutoscript-next/internal/config"
	"github.com/xfengyin/narutoscript-next/internal/device"
	"github.com/xfengyin/narutoscript-next/internal/ocr"
	"github.com/xfengyin/narutoscript-next/internal/vision"
	"github.com/xfengyin/narutoscript-next/pkg/logger"
)

// Version 版本信息
var Version = "1.0.0"

// App 应用主结构
type App struct {
	Config   *config.Config
	Logger   *logger.Logger
	Device   *device.Controller
	Vision   *vision.Matcher
	OCR      *ocr.Service
	State    *AppState
	logStore *LogStore
	ctx      context.Context
	cancel   context.CancelFunc
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
	Name     string    `json:"name"`
	Display  string    `json:"display"`
	Category string    `json:"category"`
	Status   string    `json:"status"`
	Message  string    `json:"message"`
	LastRun  time.Time `json:"last_run"`
}

// GameStats 游戏数据
type GameStats struct {
	Gold       int64     `json:"gold"`
	Copper     int64     `json:"copper"`
	Stamina    int       `json:"stamina"`
	MaxStamina int       `json:"max_stamina"`
	TasksDone  int       `json:"tasks_done"`
	TasksTotal int       `json:"tasks_total"`
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
	return &LogStore{logs: make([]LogEntry, 0, maxSize), maxSize: maxSize}
}

// Add 添加日志
func (s *LogStore) Add(level, message, task string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append(s.logs, LogEntry{Time: time.Now(), Level: level, Message: message, Task: task})
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
	start := len(s.logs) - limit
	if start < 0 {
		start = 0
	}
	result := make([]LogEntry, limit)
	copy(result, s.logs[start:])
	return result
}

// New 创建应用实例
func New(cfg *config.Config, log *logger.Logger) *App {
	ctx, cancel := context.WithCancel(context.Background())
	state := &AppState{Tasks: make(map[string]*TaskState), Stats: &GameStats{MaxStamina: 180}, StartTime: time.Now()}
	initTaskStates(state)
	return &App{
		Config:   cfg,
		Logger:   log,
		Device:   device.NewController(cfg.Device, log),
		Vision:   vision.NewMatcher(log),
		OCR:      ocr.NewService(),
		State:    state,
		logStore: NewLogStore(1000),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func initTaskStates(state *AppState) {
	tasks := []struct{ name, display, category string }{
		{"team_raid", "小队突袭", "daily"}, {"bounty_hall", "丰饶之间", "daily"}, {"guild_pray", "组织祈福", "daily"},
		{"survival", "生存试炼", "daily"}, {"equipment", "装备扫荡", "daily"}, {"task_hall", "任务集会所", "daily"},
		{"secret_realm", "秘境挑战", "daily"}, {"shop", "商店购买", "daily"},
		{"lucky_money", "招财", "harvest"}, {"intel_agency", "情报社", "harvest"}, {"ranking", "排行榜", "harvest"},
		{"activity_box", "活跃度宝箱", "harvest"}, {"monthly_sign", "每月签到", "harvest"}, {"ninja_pass", "忍法帖", "harvest"},
		{"share", "每日分享", "harvest"}, {"mail", "邮件", "harvest"}, {"stamina_gift", "赠送体力", "harvest"}, {"recruit", "招募", "harvest"},
		{"practice", "修行之路", "weekly"}, {"chase_akatsuki", "追击晓组织", "weekly"}, {"rebel", "叛忍来袭", "weekly"},
		{"guild_fortress", "组织要塞", "weekly"}, {"battlefield", "天地战场", "weekly"},
		{"bbq", "丁次烤肉", "event"}, {"sakura", "樱花季", "event"},
	}
	state.Stats.TasksTotal = len(tasks)
	for _, t := range tasks {
		state.Tasks[t.name] = &TaskState{Name: t.name, Display: t.display, Category: t.category, Status: "idle"}
	}
}

// ConnectDevice 连接设备
func (a *App) ConnectDevice() error {
	if err := a.Device.Connect(); err != nil {
		a.State.mu.Lock()
		a.State.DeviceReady = false
		a.State.mu.Unlock()
		a.logStore.Add("error", "设备连接失败: "+err.Error(), "")
		return err
	}
	a.State.mu.Lock()
	a.State.DeviceName = a.Device.GetDeviceName()
	a.State.DeviceReady = true
	a.State.mu.Unlock()
	a.logStore.Add("info", "设备已连接: "+a.Device.GetDeviceName(), "")
	return nil
}

// GetState 获取状态
func (a *App) GetState() *AppState {
	a.State.mu.RLock()
	defer a.State.mu.RUnlock()
	tasks := make(map[string]*TaskState)
	for k, v := range a.State.Tasks {
		t := *v
		tasks[k] = &t
	}
	return &AppState{
		Running:     a.State.Running,
		DeviceName:  a.State.DeviceName,
		DeviceReady: a.State.DeviceReady,
		Tasks:       tasks,
		Stats:       a.State.Stats,
		StartTime:   a.State.StartTime,
	}
}

// UpdateTaskState 更新任务状态
func (a *App) UpdateTaskState(name, status, message string) {
	a.State.mu.Lock()
	defer a.State.mu.Unlock()
	if task, ok := a.State.Tasks[name]; ok {
		task.Status = status
		task.Message = message
		if status == "success" {
			a.State.Stats.TasksDone++
		}
	}
}

// GetLogs 获取日志
func (a *App) GetLogs(limit int) []LogEntry { return a.logStore.Get(limit) }

// AddLog 添加日志
func (a *App) AddLog(level, message, task string) { a.logStore.Add(level, message, task) }