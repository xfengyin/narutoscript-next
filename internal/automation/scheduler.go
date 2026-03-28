package automation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/xfengyin/narutoscript-next/internal/device"
	"github.com/xfengyin/narutoscript-next/internal/vision"
	"github.com/xfengyin/narutoscript-next/pkg/logger"
)

// Scheduler 任务调度器
type Scheduler struct {
	cron      *cron.Cron
	log       *logger.Logger
	executor  *TaskExecutor
	running   bool
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	updater   TaskStateUpdater
	taskQueue chan string
}

// NewScheduler 创建调度器
func NewScheduler(dev *device.Controller, vis *vision.Matcher, log *logger.Logger, updater TaskStateUpdater) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		cron:      cron.New(cron.WithSeconds()),
		log:       log,
		executor:  NewTaskExecutor(dev, vis, log),
		ctx:       ctx,
		cancel:    cancel,
		updater:   updater,
		taskQueue: make(chan string, 100),
	}
}

// Run 启动调度器
func (s *Scheduler) Run() {
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	s.log.Info("任务调度器启动")
	s.registerCronJobs()
	s.cron.Start()
	go s.taskWorker()

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

func (s *Scheduler) registerCronJobs() {
	s.cron.AddFunc("0 0 8 * * *", func() {
		s.EnqueueTask("daily_all")
		s.EnqueueTask("harvest_all")
	})
	s.cron.AddFunc("0 0 12 * * *", func() {
		s.EnqueueTask("stamina_gift")
		s.EnqueueTask("mail")
	})
	s.cron.AddFunc("0 0 18 * * *", func() {
		s.EnqueueTask("stamina_gift")
		s.EnqueueTask("mail")
	})
	s.cron.AddFunc("0 0 20 * * 1", func() {
		s.EnqueueTask("weekly_all")
	})
}

// EnqueueTask 添加任务到队列
func (s *Scheduler) EnqueueTask(taskName string) {
	select {
	case s.taskQueue <- taskName:
	default:
	}
}

func (s *Scheduler) taskWorker() {
	for taskName := range s.taskQueue {
		s.log.Info("执行任务", "task", taskName)
		startTime := time.Now()
		err := s.executeTask(taskName)
		duration := time.Since(startTime)
		if err != nil {
			s.log.Error("任务失败", "task", taskName, "error", err)
			if s.updater != nil {
				s.updater.UpdateTaskState(taskName, "failed", err.Error())
			}
		} else {
			s.log.Info("任务完成", "task", taskName, "duration", duration)
			if s.updater != nil {
				s.updater.UpdateTaskState(taskName, "success", fmt.Sprintf("%.1fs", duration.Seconds()))
			}
		}
	}
}

func (s *Scheduler) executeTask(taskName string) error {
	coords := s.executor.GetCoords()
	switch taskName {
	case "team_raid":
		return s.runSimpleTask("team_raid", coords.Daily.TeamRaid)
	case "bounty_hall":
		return s.runSimpleTask("bounty_hall", coords.Daily.BountyHall)
	case "guild_pray":
		return s.runSimpleTask("guild_pray", coords.Daily.GuildPray)
	case "survival":
		return s.runSimpleTask("survival", coords.Daily.Survival)
	case "equipment":
		return s.runSimpleTask("equipment", coords.Daily.Equipment)
	case "task_hall":
		return s.runSimpleTask("task_hall", coords.Daily.TaskHall)
	case "secret_realm":
		return s.runSimpleTask("secret_realm", coords.Daily.SecretRealm)
	case "shop":
		return s.runSimpleTask("shop", coords.Daily.Shop)
	case "lucky_money":
		return s.runSimpleTask("lucky_money", coords.Harvest.LuckyMoney)
	case "intel_agency":
		return s.runSimpleTask("intel_agency", coords.Harvest.IntelAgency)
	case "ranking":
		return s.runSimpleTask("ranking", coords.Harvest.Ranking)
	case "activity_box":
		return s.runSimpleTask("activity_box", coords.Harvest.ActivityBox)
	case "monthly_sign":
		return s.runSimpleTask("monthly_sign", coords.Harvest.MonthlySign)
	case "ninja_pass":
		return s.runSimpleTask("ninja_pass", coords.Harvest.NinjaPass)
	case "share":
		return s.runSimpleTask("share", coords.Harvest.Share)
	case "mail":
		return s.runSimpleTask("mail", coords.Harvest.Mail)
	case "stamina_gift":
		return s.runSimpleTask("stamina_gift", coords.Harvest.StaminaGift)
	case "recruit":
		return s.runSimpleTask("recruit", coords.Harvest.Recruit)
	case "practice":
		return s.runSimpleTask("practice", coords.Weekly.Practice)
	case "chase_akatsuki":
		return s.runSimpleTask("chase_akatsuki", coords.Weekly.ChaseAkatsuki)
	case "rebel":
		return s.runSimpleTask("rebel", coords.Weekly.Rebel)
	case "guild_fortress":
		return s.runSimpleTask("guild_fortress", coords.Weekly.GuildFortress)
	case "battlefield":
		return s.runSimpleTask("battlefield", coords.Weekly.Battlefield)
	case "bbq":
		return s.runSimpleTask("bbq", coords.Event.BBQ)
	case "sakura":
		return s.runSimpleTask("sakura", coords.Event.Sakura)
	case "daily_all":
		for _, t := range []string{"team_raid", "bounty_hall", "guild_pray", "survival", "equipment", "task_hall", "secret_realm", "shop"} {
			s.executeTask(t)
		}
		return nil
	case "harvest_all":
		for _, t := range []string{"lucky_money", "intel_agency", "ranking", "activity_box", "monthly_sign", "ninja_pass", "share", "mail", "stamina_gift", "recruit"} {
			s.executeTask(t)
		}
		return nil
	case "weekly_all":
		for _, t := range []string{"practice", "chase_akatsuki", "rebel", "guild_fortress", "battlefield"} {
			s.executeTask(t)
		}
		return nil
	default:
		return fmt.Errorf("未知任务: %s", taskName)
	}
}

func (s *Scheduler) runSimpleTask(name string, coords TaskCoords) error {
	steps := []TaskStep{
		StepTap("进入", coords.Enter),
		StepWait("等待", s.executor.GetCoords().WaitTimes.PageLoad),
		StepTap("开始", coords.Start),
		StepWait("等待", s.executor.GetCoords().WaitTimes.BattleEnd),
		StepTap("领取", coords.Claim),
		StepBack("返回"),
	}
	return s.executor.ExecuteTask(name, steps, s.updater)
}

// RunTaskNow 立即执行任务
func (s *Scheduler) RunTaskNow(taskName string) error {
	s.EnqueueTask(taskName)
	return nil
}

// GetTaskQueueLength 获取队列长度
func (s *Scheduler) GetTaskQueueLength() int {
	return len(s.taskQueue)
}