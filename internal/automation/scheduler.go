package automation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/xfengyin/narutoscript-next/internal/app"
	"github.com/xfengyin/narutoscript-next/internal/device"
	"github.com/xfengyin/narutoscript-next/internal/ocr"
	"github.com/xfengyin/narutoscript-next/internal/vision"
	"github.com/xfengyin/narutoscript-next/pkg/logger"
)

// Scheduler 任务调度器
type Scheduler struct {
	cron      *cron.Cron
	config    app.AutomationConfig
	log       *logger.Logger
	executor  *TaskExecutor
	ocr       *ocr.Service
	running   bool
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	app       *app.App
	taskQueue chan string
}

// NewScheduler 创建调度器
func NewScheduler(cfg app.AutomationConfig, appInst *app.App, log *logger.Logger) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Scheduler{
		cron:      cron.New(cron.WithSeconds()),
		config:    cfg,
		log:       log,
		app:       appInst,
		ocr:       ocr.NewService(),
		ctx:       ctx,
		cancel:    cancel,
		taskQueue: make(chan string, 100),
	}

	// 创建任务执行器
	if appInst.Device != nil && appInst.Vision != nil {
		s.executor = NewTaskExecutor(appInst.Device, appInst.Vision, appInst, log)
	}

	return s
}

// Run 启动调度器
func (s *Scheduler) Run() {
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	s.log.Info("任务调度器启动")

	// 注册定时任务
	s.registerCronJobs()

	// 启动 cron
	s.cron.Start()

	// 启动任务执行工作线程
	go s.taskWorker()

	// 等待停止信号
	<-s.ctx.Done()

	s.cron.Stop()
	s.log.Info("任务调度器停止")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		s.cancel()
		s.running = false
		close(s.taskQueue)
	}
}

// registerCronJobs 注册定时任务
func (s *Scheduler) registerCronJobs() {
	// 每天早上 8:00 执行所有日常和收获
	s.cron.AddFunc("0 0 8 * * *", func() {
		s.log.Info("触发定时任务: 日常+收获")
		s.EnqueueTask("daily_all")
		s.EnqueueTask("harvest_all")
	})

	// 每天中午 12:00 领取体力
	s.cron.AddFunc("0 0 12 * * *", func() {
		s.log.Info("触发定时任务: 午间体力")
		s.EnqueueTask("stamina_gift")
		s.EnqueueTask("mail")
	})

	// 每天下午 18:00 领取体力
	s.cron.AddFunc("0 0 18 * * *", func() {
		s.log.Info("触发定时任务: 晚间体力")
		s.EnqueueTask("stamina_gift")
		s.EnqueueTask("mail")
	})

	// 每周一 20:00 执行周常
	s.cron.AddFunc("0 0 20 * * 1", func() {
		s.log.Info("触发定时任务: 周常")
		s.EnqueueTask("weekly_all")
	})

	// 每小时检查活跃度宝箱
	s.cron.AddFunc("0 0 * * * *", func() {
		s.log.Info("触发定时任务: 活跃度宝箱")
		s.EnqueueTask("activity_box")
	})

	s.log.Info("定时任务已注册")
}

// EnqueueTask 添加任务到队列
func (s *Scheduler) EnqueueTask(taskName string) {
	select {
	case s.taskQueue <- taskName:
		s.log.Debug("任务已入队", "task", taskName)
	default:
		s.log.Warn("任务队列已满", "task", taskName)
	}
}

// taskWorker 任务执行工作线程
func (s *Scheduler) taskWorker() {
	for taskName := range s.taskQueue {
		s.log.Info("开始执行任务", "task", taskName)
		startTime := time.Now()

		err := s.executeTask(taskName)

		duration := time.Since(startTime)
		if err != nil {
			s.log.Error("任务失败", "task", taskName, "duration", duration, "error", err)
			s.app.UpdateTaskState(taskName, "failed", err.Error())
			s.app.AddLog("error", fmt.Sprintf("任务失败: %s (%v)", taskName, err), taskName)
		} else {
			s.log.Info("任务完成", "task", taskName, "duration", duration)
			s.app.UpdateTaskState(taskName, "success", fmt.Sprintf("耗时 %.1f 秒", duration.Seconds()))
			s.app.AddLog("success", fmt.Sprintf("任务完成: %s (%.1fs)", taskName, duration.Seconds()), taskName)
		}
	}
}

// executeTask 执行指定任务
func (s *Scheduler) executeTask(taskName string) error {
	if s.executor == nil {
		return fmt.Errorf("任务执行器未初始化")
	}

	// 检查设备连接
	if !s.app.Device.IsConnected() {
		s.log.Warn("设备未连接，尝试重连")
		if err := s.app.Device.Reconnect(); err != nil {
			return fmt.Errorf("设备连接失败: %v", err)
		}
	}

	// 执行任务
	switch taskName {
	// 日常任务
	case "team_raid":
		return s.executor.ExecuteTeamRaid()
	case "bounty_hall":
		return s.executor.ExecuteBountyHall()
	case "guild_pray":
		return s.executor.ExecuteGuildPray()
	case "survival":
		return s.executor.ExecuteSurvival()
	case "equipment":
		return s.executor.ExecuteEquipment()
	case "task_hall":
		return s.executor.ExecuteTaskHall()
	case "secret_realm":
		return s.executor.ExecuteSecretRealm()
	case "shop":
		return s.executor.ExecuteShop()

	// 收获任务
	case "lucky_money":
		return s.executor.ExecuteLuckyMoney()
	case "intel_agency":
		return s.executor.ExecuteIntelAgency()
	case "ranking":
		return s.executor.ExecuteRanking()
	case "activity_box":
		return s.executor.ExecuteActivityBox()
	case "monthly_sign":
		return s.executor.ExecuteMonthlySign()
	case "ninja_pass":
		return s.executor.ExecuteNinjaPass()
	case "share":
		return s.executor.ExecuteShare()
	case "mail":
		return s.executor.ExecuteMail()
	case "stamina_gift":
		return s.executor.ExecuteStaminaGift()
	case "recruit":
		return s.executor.ExecuteRecruit()

	// 周常任务
	case "practice":
		return s.executor.ExecutePractice()
	case "chase_akatsuki":
		return s.executor.ExecuteChaseAkatsuki()
	case "rebel":
		return s.executor.ExecuteRebel()
	case "guild_fortress":
		return s.executor.ExecuteGuildFortress()
	case "battlefield":
		return s.executor.ExecuteBattlefield()

	// 活动任务
	case "bbq":
		return s.executor.ExecuteBBQ()
	case "sakura":
		return s.executor.ExecuteSakura()

	// 批量任务
	case "daily_all":
		return s.executor.ExecuteAllDaily(s.ctx)
	case "harvest_all":
		return s.executor.ExecuteAllHarvest(s.ctx)
	case "weekly_all":
		return s.executor.ExecuteAllWeekly(s.ctx)

	default:
		return fmt.Errorf("未知任务: %s", taskName)
	}
}

// RunTaskNow 立即执行任务
func (s *Scheduler) RunTaskNow(taskName string) error {
	s.EnqueueTask(taskName)
	return nil
}

// GetTaskQueueLength 获取任务队列长度
func (s *Scheduler) GetTaskQueueLength() int {
	return len(s.taskQueue)
}

// ===== 任务注册接口（供外部调用）=====

// TaskInfo 任务信息
type TaskInfo struct {
	Name        string
	Category    string
	Description string
	Schedule    string
}

// GetAllTasks 获取所有任务信息
func GetAllTasks() []TaskInfo {
	return []TaskInfo{
		// 日常
		{Name: "team_raid", Category: "daily", Description: "小队突袭", Schedule: "每天8:00"},
		{Name: "bounty_hall", Category: "daily", Description: "丰饶之间", Schedule: "每天8:00"},
		{Name: "guild_pray", Category: "daily", Description: "组织祈福", Schedule: "每天8:00"},
		{Name: "survival", Category: "daily", Description: "生存试炼", Schedule: "每天8:00"},
		{Name: "equipment", Category: "daily", Description: "装备扫荡", Schedule: "每天8:00"},
		{Name: "task_hall", Category: "daily", Description: "任务集会所", Schedule: "每天8:00"},
		{Name: "secret_realm", Category: "daily", Description: "秘境挑战", Schedule: "每天8:00"},
		{Name: "shop", Category: "daily", Description: "商店购买", Schedule: "每天8:00"},
		// 收获
		{Name: "lucky_money", Category: "harvest", Description: "招财", Schedule: "每天8:00"},
		{Name: "intel_agency", Category: "harvest", Description: "情报社", Schedule: "每天8:00"},
		{Name: "ranking", Category: "harvest", Description: "排行榜", Schedule: "每天8:00"},
		{Name: "activity_box", Category: "harvest", Description: "活跃度宝箱", Schedule: "每小时"},
		{Name: "monthly_sign", Category: "harvest", Description: "每月签到", Schedule: "每天8:00"},
		{Name: "ninja_pass", Category: "harvest", Description: "忍法帖", Schedule: "每天8:00"},
		{Name: "share", Category: "harvest", Description: "每日分享", Schedule: "每天8:00"},
		{Name: "mail", Category: "harvest", Description: "邮件", Schedule: "每天12:00,18:00"},
		{Name: "stamina_gift", Category: "harvest", Description: "赠送体力", Schedule: "每天12:00,18:00"},
		{Name: "recruit", Category: "harvest", Description: "招募", Schedule: "每天8:00"},
		// 周常
		{Name: "practice", Category: "weekly", Description: "修行之路", Schedule: "每周一20:00"},
		{Name: "chase_akatsuki", Category: "weekly", Description: "追击晓组织", Schedule: "每周一20:00"},
		{Name: "rebel", Category: "weekly", Description: "叛忍来袭", Schedule: "每周一20:00"},
		{Name: "guild_fortress", Category: "weekly", Description: "组织要塞", Schedule: "每周一20:00"},
		{Name: "battlefield", Category: "weekly", Description: "天地战场", Schedule: "每周一20:00"},
		// 活动
		{Name: "bbq", Category: "event", Description: "丁次烤肉", Schedule: "活动期间"},
		{Name: "sakura", Category: "event", Description: "樱花季", Schedule: "活动期间"},
		// 批量
		{Name: "daily_all", Category: "batch", Description: "所有日常", Schedule: "手动"},
		{Name: "harvest_all", Category: "batch", Description: "所有收获", Schedule: "手动"},
		{Name: "weekly_all", Category: "batch", Description: "所有周常", Schedule: "手动"},
	}
}