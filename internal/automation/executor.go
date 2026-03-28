package automation

import (
	"fmt"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/xfengyin/narutoscript-next/internal/app"
	"github.com/xfengyin/narutoscript-next/internal/device"
	"github.com/xfengyin/narutoscript-next/internal/vision"
	"github.com/xfengyin/narutoscript-next/pkg/logger"
)

// TaskExecutor 任务执行器
type TaskExecutor struct {
	device *device.Controller
	vision *vision.Matcher
	coords *CoordinateConfig
	log    *logger.Logger
	app    *app.App
	mu     sync.Mutex
}

// CoordinateConfig 坐标配置
type CoordinateConfig struct {
	MainMenu  MainMenuConfig  `mapstructure:"main_menu"`
	Daily     DailyConfig     `mapstructure:"daily"`
	Harvest   HarvestConfig   `mapstructure:"harvest"`
	Weekly    WeeklyConfig    `mapstructure:"weekly"`
	Event     EventConfig     `mapstructure:"event"`
	WaitTimes WaitTimesConfig `mapstructure:"wait_times"`
}

// 各模块配置结构
type MainMenuConfig struct {
	BottomNav    map[string][]int `mapstructure:"bottom_nav"`
	TopResources map[string][]int `mapstructure:"top_resources"`
	Buttons      map[string][]int `mapstructure:"buttons"`
}

type DailyConfig struct {
	Entries     map[string][]int `mapstructure:"entries"`
	TeamRaid    TaskCoords       `mapstructure:"team_raid"`
	BountyHall  TaskCoords       `mapstructure:"bounty_hall"`
	GuildPray   TaskCoords       `mapstructure:"guild_pray"`
	Survival    TaskCoords       `mapstructure:"survival"`
	Equipment   TaskCoords       `mapstructure:"equipment"`
	TaskHall    TaskCoords       `mapstructure:"task_hall"`
	SecretRealm TaskCoords       `mapstructure:"secret_realm"`
	Shop        TaskCoords       `mapstructure:"shop"`
}

type HarvestConfig struct {
	LuckyMoney  TaskCoords `mapstructure:"lucky_money"`
	IntelAgency TaskCoords `mapstructure:"intel_agency"`
	Ranking     TaskCoords `mapstructure:"ranking"`
	ActivityBox TaskCoords `mapstructure:"activity_box"`
	MonthlySign TaskCoords `mapstructure:"monthly_sign"`
	NinjaPass   TaskCoords `mapstructure:"ninja_pass"`
	Share       TaskCoords `mapstructure:"share"`
	Mail        TaskCoords `mapstructure:"mail"`
	StaminaGift TaskCoords `mapstructure:"stamina_gift"`
	Recruit     TaskCoords `mapstructure:"recruit"`
}

type WeeklyConfig struct {
	Practice      TaskCoords `mapstructure:"practice"`
	ChaseAkatsuki TaskCoords `mapstructure:"chase_akatsuki"`
	Rebel         TaskCoords `mapstructure:"rebel"`
	GuildFortress TaskCoords `mapstructure:"guild_fortress"`
	Battlefield   TaskCoords `mapstructure:"battlefield"`
}

type EventConfig struct {
	BBQ    TaskCoords `mapstructure:"bbq"`
	Sakura TaskCoords `mapstructure:"sakura"`
}

type TaskCoords struct {
	Enter          []int `mapstructure:"enter"`
	Start          []int `mapstructure:"start"`
	Confirm        []int `mapstructure:"confirm"`
	Claim          []int `mapstructure:"claim"`
	Auto           []int `mapstructure:"auto"`
	RepeatTimes    int   `mapstructure:"repeat_times"`
	SelectType     []int `mapstructure:"select_type"`
	SelectLevel    []int `mapstructure:"select_level"`
	SelectStage    []int `mapstructure:"select_stage"`
	Sweep          []int `mapstructure:"sweep"`
	SweepTimes     []int `mapstructure:"sweep_times"`
	Next           []int `mapstructure:"next"`
	ClaimAll       []int `mapstructure:"claim_all"`
	Refresh        []int `mapstructure:"refresh"`
	FreeClaim      []int `mapstructure:"free_claim"`
	PremiumClaim   []int `mapstructure:"premium_claim"`
	Join           []int `mapstructure:"join"`
	Battle         []int `mapstructure:"battle"`
	Collect        []int `mapstructure:"collect"`
	Exchange       []int `mapstructure:"exchange"`
	Feed           []int `mapstructure:"feed"`
	Gift           []int `mapstructure:"gift"`
	RecruitFree    []int `mapstructure:"free_recruit"`
	RecruitPremium []int `mapstructure:"premium_recruit"`
}

type WaitTimesConfig struct {
	PageLoad    int `mapstructure:"page_load"`
	BattleStart int `mapstructure:"battle_start"`
	BattleEnd   int `mapstructure:"battle_end"`
	RewardClaim int `mapstructure:"reward_claim"`
	Animation   int `mapstructure:"animation"`
	RetryDelay  int `mapstructure:"retry_delay"`
}

// NewTaskExecutor 创建任务执行器
func NewTaskExecutor(dev *device.Controller, vis *vision.Matcher, appInst *app.App, log *logger.Logger) *TaskExecutor {
	coords := loadCoordinateConfig()
	return &TaskExecutor{
		device: dev,
		vision: vis,
		coords: coords,
		log:    log,
		app:    appInst,
	}
}

func loadCoordinateConfig() *CoordinateConfig {
	viper.SetConfigFile("assets/coords.yaml")
	viper.SetConfigType("yaml")

	var coords CoordinateConfig
	if err := viper.ReadInConfig(); err != nil {
		return getDefaultCoords()
	}
	if err := viper.Unmarshal(&coords); err != nil {
		return getDefaultCoords()
	}
	return &coords
}

func getDefaultCoords() *CoordinateConfig {
	return &CoordinateConfig{
		MainMenu: MainMenuConfig{
			BottomNav: map[string][]int{
				"home":   {640, 680},
				"ninja":  {320, 680},
				"battle": {960, 680},
				"guild":  {160, 680},
				"shop":   {1120, 680},
			},
			Buttons: map[string][]int{
				"close":   {1200, 50},
				"back":    {80, 50},
				"confirm": {640, 500},
			},
		},
		WaitTimes: WaitTimesConfig{
			PageLoad:    2000,
			BattleStart: 3000,
			BattleEnd:   5000,
			RewardClaim: 1000,
			Animation:   500,
			RetryDelay:  1000,
		},
	}
}

// Tap 点击坐标
func (e *TaskExecutor) Tap(x, y int) error {
	e.log.Debug("点击", "x", x, "y", y)
	return e.device.Tap(x, y)
}

// TapCoord 点击配置中的坐标
func (e *TaskExecutor) TapCoord(coord []int) error {
	if len(coord) < 2 {
		return fmt.Errorf("无效坐标")
	}
	return e.Tap(coord[0], coord[1])
}

// Wait 等待
func (e *TaskExecutor) Wait(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// WaitPage 等待页面加载
func (e *TaskExecutor) WaitPage() {
	e.Wait(e.coords.WaitTimes.PageLoad)
}

// WaitBattle 等待战斗结束
func (e *TaskExecutor) WaitBattle() {
	e.Wait(e.coords.WaitTimes.BattleEnd)
}

// GoBack 返回主界面
func (e *TaskExecutor) GoBack() error {
	for i := 0; i < 3; i++ {
		e.TapCoord(e.coords.MainMenu.Buttons["back"])
		e.Wait(e.coords.WaitTimes.Animation)
	}
	return nil
}

// GoHome 回到主页
func (e *TaskExecutor) GoHome() error {
	return e.TapCoord(e.coords.MainMenu.BottomNav["home"])
}

// ClosePopup 关闭弹窗
func (e *TaskExecutor) ClosePopup() error {
	return e.TapCoord(e.coords.MainMenu.Buttons["close"])
}

// Confirm 确认操作
func (e *TaskExecutor) Confirm() error {
	return e.TapCoord(e.coords.MainMenu.Buttons["confirm"])
}

// TaskStep 任务步骤
type TaskStep struct {
	Desc   string
	Action func(e *TaskExecutor) error
}

// ExecuteTask 执行任务
func (e *TaskExecutor) ExecuteTask(taskName string, steps []TaskStep) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.app.UpdateTaskState(taskName, "running", "开始执行")
	e.log.Info("开始任务", "task", taskName)

	for i, step := range steps {
		e.app.UpdateTaskState(taskName, "running", fmt.Sprintf("步骤 %d/%d: %s", i+1, len(steps), step.Desc))
		if err := step.Action(e); err != nil {
			e.app.UpdateTaskState(taskName, "failed", fmt.Sprintf("步骤 %d 失败: %v", i+1, err))
			return err
		}
	}

	e.app.UpdateTaskState(taskName, "success", "任务完成")
	return nil
}

// StepTap 点击步骤
func StepTap(desc string, coord []int) TaskStep {
	return TaskStep{
		Desc: desc,
		Action: func(e *TaskExecutor) error {
			return e.TapCoord(coord)
		},
	}
}

// StepWait 等待步骤
func StepWait(desc string, ms int) TaskStep {
	return TaskStep{
		Desc: desc,
		Action: func(e *TaskExecutor) error {
			e.Wait(ms)
			return nil
		},
	}
}

// StepBack 返回步骤
func StepBack(desc string) TaskStep {
	return TaskStep{
		Desc: desc,
		Action: func(e *TaskExecutor) error {
			return e.GoBack()
		},
	}
}

// StepConfirm 确认步骤
func StepConfirm(desc string) TaskStep {
	return TaskStep{
		Desc: desc,
		Action: func(e *TaskExecutor) error {
			return e.Confirm()
		},
	}
}

// RepeatSteps 重复执行步骤
func RepeatSteps(steps []TaskStep, times int) []TaskStep {
	result := make([]TaskStep, 0, len(steps)*times)
	for i := 0; i < times; i++ {
		result = append(result, steps...)
	}
	return result
}