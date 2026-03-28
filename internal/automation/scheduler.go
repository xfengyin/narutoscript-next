package automation

import (
	"context"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/xfengyin/narutoscript-next/internal/app"
	"github.com/xfengyin/narutoscript-next/pkg/logger"
)

type Scheduler struct {
	cron     *cron.Cron
	config   app.AutomationConfig
	log      *logger.Logger
	tasks    map[string]Task
	running  bool
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
}

type Task interface {
	Name() string
	Execute() error
}

func NewScheduler(cfg app.AutomationConfig, log *logger.Logger) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		cron:   cron.New(),
		config: cfg,
		log:    log,
		tasks:  make(map[string]Task),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Scheduler) Run() {
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	// 添加定时任务
	// 每天 8:00 执行日常任务
	s.cron.AddFunc("0 8 * * *", func() {
		s.executeDailyTasks()
	})

	// 每周一 20:00 执行周常任务
	s.cron.AddFunc("0 20 * * 1", func() {
		s.executeWeeklyTasks()
	})

	s.cron.Start()
	s.log.Info("任务调度器已启动")

	<-s.ctx.Done()
	s.cron.Stop()
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.running {
		s.cancel()
		s.running = false
		s.log.Info("任务调度器已停止")
	}
}

func (s *Scheduler) executeDailyTasks() {
	s.log.Info("开始执行日常任务")
	// TODO: 执行所有日常任务
}

func (s *Scheduler) executeWeeklyTasks() {
	s.log.Info("开始执行周常任务")
	// TODO: 执行所有周常任务
}

func (s *Scheduler) AddTask(name string, task Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[name] = task
}

func (s *Scheduler) RunTask(name string) error {
	s.mu.Lock()
	task, ok := s.tasks[name]
	s.mu.Unlock()

	if !ok {
		return nil // 任务不存在
	}

	return task.Execute()
}