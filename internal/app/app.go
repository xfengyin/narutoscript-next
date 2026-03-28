package app

import (
	"sync"

	"github.com/xfengyin/narutoscript-next/internal/automation"
	"github.com/xfengyin/narutoscript-next/internal/device"
	"github.com/xfengyin/narutoscript-next/internal/vision"
	"github.com/xfengyin/narutoscript-next/pkg/logger"
)

type App struct {
	Config    *Config
	Logger    *logger.Logger
	Device    *device.Controller
	Vision    *vision.Matcher
	Scheduler *automation.Scheduler
	State     *AppState
}

type AppState struct {
	mu       sync.RWMutex
	Running  bool
	Tasks    map[string]*TaskState
	Stats    *GameStats
}

type TaskState struct {
	Name      string
	Status    string // idle, running, success, failed
	Progress  int
	Message   string
	StartTime int64
	EndTime   int64
}

type GameStats struct {
	Gold      int64
	Copper    int64
	Stamina   int
	TasksDone int
	TasksTotal int
}

func New(cfg *Config, log *logger.Logger) *App {
	state := &AppState{
		Running: false,
		Tasks:   make(map[string]*TaskState),
		Stats:   &GameStats{},
	}

	return &App{
		Config: cfg,
		Logger: log,
		Device: device.NewController(cfg.Device, log),
		Vision: vision.NewMatcher(log),
		State:  state,
	}
}

func (a *App) StartScheduler() error {
	a.State.mu.Lock()
	defer a.State.mu.Unlock()

	if a.State.Running {
		return nil
	}

	a.Scheduler = automation.NewScheduler(a.Config.Automation, a.Logger)
	a.State.Running = true

	go a.Scheduler.Run()

	return nil
}

func (a *App) StopScheduler() {
	a.State.mu.Lock()
	defer a.State.mu.Unlock()

	if a.Scheduler != nil {
		a.Scheduler.Stop()
	}
	a.State.Running = false
}

func (a *App) GetState() *AppState {
	a.State.mu.RLock()
	defer a.State.mu.RUnlock()

	// 返回副本
	tasks := make(map[string]*TaskState)
	for k, v := range a.State.Tasks {
		tasks[k] = v
	}

	return &AppState{
		Running: a.State.Running,
		Tasks:   tasks,
		Stats:   a.State.Stats,
	}
}

func (a *App) UpdateTaskState(name string, state *TaskState) {
	a.State.mu.Lock()
	defer a.State.mu.Unlock()
	a.State.Tasks[name] = state
}

func (a *App) UpdateStats(stats *GameStats) {
	a.State.mu.Lock()
	defer a.State.mu.Unlock()
	a.State.Stats = stats
}