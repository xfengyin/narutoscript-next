package automation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"github.com/xfengyin/narutoscript-next/internal/app"
	"github.com/xfengyin/narutoscript-next/internal/device"
	"github.com/xfengyin/narutoscript-next/internal/vision"
	"github.com/xfengyin/narutoscript-next/pkg/logger"
)

// TaskExecutor 任务执行器
type TaskExecutor struct {
	device  *device.Controller
	vision  *vision.Matcher
	coords  *CoordinateConfig
	log     *logger.Logger
	app     *app.App
	mu      sync.Mutex
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
	BottomNav   map[string][]int `mapstructure:"bottom_nav"`
	TopResources map[string][]int `mapstructure:"top_resources"`
	Buttons     map[string][]int `mapstructure:"buttons"`
}

type DailyConfig struct {
	Entries    map[string][]int `mapstructure:"entries"`
	TeamRaid   TaskCoords       `mapstructure:"team_raid"`
	BountyHall TaskCoords       `mapstructure:"bounty_hall"`
	GuildPray  TaskCoords       `mapstructure:"guild_pray"`
	Survival   TaskCoords       `mapstructure:"survival"`
	Equipment  TaskCoords       `mapstructure:"equipment"`
	TaskHall   TaskCoords       `mapstructure:"task_hall"`
	SecretRealm TaskCoords      `mapstructure:"secret_realm"`
	Shop       TaskCoords       `mapstructure:"shop"`
}

type HarvestConfig struct {
	LuckyMoney   TaskCoords `mapstructure:"lucky_money"`
	IntelAgency  TaskCoords `mapstructure:"intel_agency"`
	Ranking      TaskCoords `mapstructure:"ranking"`
	ActivityBox  TaskCoords `mapstructure:"activity_box"`
	MonthlySign  TaskCoords `mapstructure:"monthly_sign"`
	NinjaPass    TaskCoords `mapstructure:"ninja_pass"`
	Share        TaskCoords `mapstructure:"share"`
	Mail         TaskCoords `mapstructure:"mail"`
	StaminaGift  TaskCoords `mapstructure:"stamina_gift"`
	Recruit      TaskCoords `mapstructure:"recruit"`
}

type WeeklyConfig struct {
	Practice       TaskCoords `mapstructure:"practice"`
	ChaseAkatsuki  TaskCoords `mapstructure:"chase_akatsuki"`
	Rebel          TaskCoords `mapstructure:"rebel"`
	GuildFortress  TaskCoords `mapstructure:"guild_fortress"`
	Battlefield    TaskCoords `mapstructure:"battlefield"`
}

type EventConfig struct {
	BBQ    TaskCoords `mapstructure:"bbq"`
	Sakura TaskCoords `mapstructure:"sakura"`
}

type TaskCoords struct {
	Enter        []int `mapstructure:"enter"`
	Start        []int `mapstructure:"start"`
	Confirm      []int `mapstructure:"confirm"`
	Claim        []int `mapstructure:"claim"`
	Auto         []int `mapstructure:"auto"`
	RepeatTimes  int   `mapstructure:"repeat_times"`
	SelectType   []int `mapstructure:"select_type"`
	SelectLevel  []int `mapstructure:"select_level"`
	SelectStage  []int `mapstructure:"select_stage"`
	Sweep        []int `mapstructure:"sweep"`
	SweepTimes   []int `mapstructure:"sweep_times"`
	Next         []int `mapstructure:"next"`
	ClaimAll     []int `mapstructure:"claim_all"`
	Refresh      []int `mapstructure:"refresh"`
	FreeClaim    []int `mapstructure:"free_claim"`
	PremiumClaim []int `mapstructure:"premium_claim"`
	Join         []int `mapstructure:"join"`
	Battle       []int `mapstructure:"battle"`
	Collect      []int `mapstructure:"collect"`
	Exchange     []int `mapstructure:"exchange"`
	Feed         []int `mapstructure:"feed"`
	Gift         []int `mapstructure:"gift"`
	RecruitFree  []int `mapstructure:"free_recruit"`
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
	// 加载坐标配置
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
		// 使用默认坐标
		return getDefaultCoords()
	}

	if err := viper.Unmarshal(&coords); err != nil {
		return getDefaultCoords()
	}

	return &coords
}

func getDefaultCoords() *CoordinateConfig {
	// 默认坐标（1280x720）
	return &CoordinateConfig{
		MainMenu: MainMenuConfig{
			BottomNav: map[string][]int{
				"home":  {640, 680},
				"ninja": {320, 680},
				"battle": {960, 680},
				"guild": {160, 680},
				"shop":  {1120, 680},
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

// ===== 基础操作方法 =====

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
	// 多次点击返回确保回到主界面
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

// ===== 任务执行框架 =====

// ExecuteTask 执行任务（通用框架）
func (e *TaskExecutor) ExecuteTask(taskName string, steps []TaskStep) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.app.UpdateTaskState(taskName, "running", "开始执行")
	e.log.Info("开始任务", "task", taskName)

	for i, step := range steps {
		progress := (i + 1) * 100 / len(steps)
		e.app.UpdateTaskState(taskName, "running", fmt.Sprintf("步骤 %d/%d: %s", i+1, len(steps), step.Desc))
		e.log.Debug("执行步骤", "task", taskName, "step", i+1, "desc", step.Desc)

		if err := step.Action(e); err != nil {
			e.log.Error("步骤失败", "task", taskName, "step", i+1, "error", err)
			e.app.UpdateTaskState(taskName, "failed", fmt.Sprintf("步骤 %d 失败: %v", i+1, err))
			return err
		}
	}

	e.app.UpdateTaskState(taskName, "success", "任务完成")
	e.log.Info("任务完成", "task", taskName)
	return nil
}

// TaskStep 任务步骤
type TaskStep struct {
	Desc   string
	Action func(e *TaskExecutor) error
}

// ===== 常用步骤生成器 =====

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

// ===== 具体任务实现 =====

// ExecuteTeamRaid 小队突袭
func (e *TaskExecutor) ExecuteTeamRaid() error {
	steps := []TaskStep{
		StepTap("进入小队突袭", e.coords.Daily.TeamRaid.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("选择难度", e.coords.Daily.TeamRaid.SelectType),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("开始战斗", e.coords.Daily.TeamRaid.Start),
		StepWait("等待战斗开始", e.coords.WaitTimes.BattleStart),
		StepTap("开启自动", e.coords.Daily.TeamRaid.Auto),
		StepWait("等待战斗结束", e.coords.WaitTimes.BattleEnd),
		StepTap("领取奖励", e.coords.Daily.TeamRaid.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
	}

	// 重复 3 次
	allSteps := RepeatSteps(steps, e.coords.Daily.TeamRaid.RepeatTimes)
	allSteps = append(allSteps, StepBack("返回主界面"))

	return e.ExecuteTask("team_raid", allSteps)
}

// ExecuteBountyHall 丰饶之间
func (e *TaskExecutor) ExecuteBountyHall() error {
	steps := []TaskStep{
		StepTap("进入丰饶之间", e.coords.Daily.BountyHall.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("选择类型", e.coords.Daily.BountyHall.SelectType),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("开始", e.coords.Daily.BountyHall.Start),
		StepWait("等待战斗开始", e.coords.WaitTimes.BattleStart),
		StepTap("开启自动", e.coords.Daily.BountyHall.Auto),
		StepWait("等待战斗结束", e.coords.WaitTimes.BattleEnd),
		StepTap("领取奖励", e.coords.Daily.BountyHall.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
	}

	allSteps := RepeatSteps(steps, e.coords.Daily.BountyHall.RepeatTimes)
	allSteps = append(allSteps, StepBack("返回主界面"))

	return e.ExecuteTask("bounty_hall", allSteps)
}

// ExecuteGuildPray 组织祈福
func (e *TaskExecutor) ExecuteGuildPray() error {
	steps := []TaskStep{
		StepTap("进入组织", e.coords.MainMenu.BottomNav["guild"]),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("点击祈福", e.coords.Daily.GuildPray.Enter),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("选择等级", e.coords.Daily.GuildPray.SelectLevel),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepConfirm("确认祈福"),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("guild_pray", steps)
}

// ExecuteSurvival 生存试炼
func (e *TaskExecutor) ExecuteSurvival() error {
	steps := []TaskStep{
		StepTap("进入生存试炼", e.coords.Daily.Survival.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("开始挑战", e.coords.Daily.Survival.Start),
		StepWait("等待战斗开始", e.coords.WaitTimes.BattleStart),
		StepTap("开启自动", e.coords.Daily.Survival.Auto),
		StepWait("等待战斗结束", e.coords.WaitTimes.BattleEnd),
		StepTap("领取奖励", e.coords.Daily.Survival.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
	}

	// 生存试炼可能有多层，这里简化处理
	allSteps := append(steps, StepBack("返回主界面"))

	return e.ExecuteTask("survival", allSteps)
}

// ExecuteEquipment 装备扫荡
func (e *TaskExecutor) ExecuteEquipment() error {
	steps := []TaskStep{
		StepTap("进入装备", e.coords.Daily.Equipment.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("选择关卡", e.coords.Daily.Equipment.SelectStage),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("扫荡", e.coords.Daily.Equipment.Sweep),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("选择次数", e.coords.Daily.Equipment.SweepTimes),
		StepConfirm("确认扫荡"),
		StepWait("等待", e.coords.WaitTimes.BattleEnd),
		StepTap("领取奖励", e.coords.Daily.Equipment.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("equipment", steps)
}

// ExecuteTaskHall 任务集会所
func (e *TaskExecutor) ExecuteTaskHall() error {
	steps := []TaskStep{
		StepTap("进入任务集会所", e.coords.Daily.TaskHall.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("领取全部", e.coords.Daily.TaskHall.ClaimAll),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("task_hall", steps)
}

// ExecuteSecretRealm 秘境挑战
func (e *TaskExecutor) ExecuteSecretRealm() error {
	steps := []TaskStep{
		StepTap("进入秘境", e.coords.Daily.SecretRealm.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("选择秘境", e.coords.Daily.SecretRealm.SelectType),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("开始挑战", e.coords.Daily.SecretRealm.Start),
		StepWait("等待战斗开始", e.coords.WaitTimes.BattleStart),
		StepTap("开启自动", e.coords.Daily.SecretRealm.Auto),
		StepWait("等待战斗结束", e.coords.WaitTimes.BattleEnd),
		StepTap("领取奖励", e.coords.Daily.SecretRealm.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
	}

	allSteps := RepeatSteps(steps, e.coords.Daily.SecretRealm.RepeatTimes)
	allSteps = append(allSteps, StepBack("返回主界面"))

	return e.ExecuteTask("secret_realm", allSteps)
}

// ExecuteShop 商店购买
func (e *TaskExecutor) ExecuteShop() error {
	steps := []TaskStep{
		StepTap("进入商店", e.coords.MainMenu.BottomNav["shop"]),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("日常商店", e.coords.Daily.Shop.Enter),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("全部购买", e.coords.Daily.Shop.ClaimAll),
		StepConfirm("确认购买"),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("shop", steps)
}

// ===== 收获系统任务 =====

// ExecuteLuckyMoney 招财
func (e *TaskExecutor) ExecuteLuckyMoney() error {
	steps := []TaskStep{
		StepTap("进入招财", e.coords.Harvest.LuckyMoney.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("免费招财", e.coords.Harvest.LuckyMoney.FreeClaim),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("领取", e.coords.Harvest.LuckyMoney.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("lucky_money", steps)
}

// ExecuteIntelAgency 情报社
func (e *TaskExecutor) ExecuteIntelAgency() error {
	steps := []TaskStep{
		StepTap("进入情报社", e.coords.Harvest.IntelAgency.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("领取情报", e.coords.Harvest.IntelAgency.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("intel_agency", steps)
}

// ExecuteRanking 排行榜奖励
func (e *TaskExecutor) ExecuteRanking() error {
	steps := []TaskStep{
		StepTap("进入排行榜", e.coords.Harvest.Ranking.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("领取奖励", e.coords.Harvest.Ranking.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("ranking", steps)
}

// ExecuteActivityBox 活跃度宝箱
func (e *TaskExecutor) ExecuteActivityBox() error {
	// 活跃度宝箱需要根据当前活跃度判断
	steps := []TaskStep{
		// 点击各个宝箱位置
		StepTap("检查宝箱", []int{200, 200}),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("检查宝箱", []int{400, 200}),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("检查宝箱", []int{600, 200}),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("检查宝箱", []int{800, 200}),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("检查宝箱", []int{1000, 200}),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
	}

	return e.ExecuteTask("activity_box", steps)
}

// ExecuteMonthlySign 每月签到
func (e *TaskExecutor) ExecuteMonthlySign() error {
	steps := []TaskStep{
		StepTap("进入签到", e.coords.Harvest.MonthlySign.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("领取今日奖励", e.coords.Harvest.MonthlySign.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("monthly_sign", steps)
}

// ExecuteNinjaPass 忍法帖
func (e *TaskExecutor) ExecuteNinjaPass() error {
	steps := []TaskStep{
		StepTap("进入忍法帖", e.coords.Harvest.NinjaPass.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("领取等级奖励", e.coords.Harvest.NinjaPass.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("ninja_pass", steps)
}

// ExecuteShare 每日分享
func (e *TaskExecutor) ExecuteShare() error {
	steps := []TaskStep{
		StepTap("进入分享", e.coords.Harvest.Share.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("分享", e.coords.Harvest.Share.Enter),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepConfirm("确认分享"),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("share", steps)
}

// ExecuteMail 邮件领取
func (e *TaskExecutor) ExecuteMail() error {
	steps := []TaskStep{
		StepTap("进入邮件", e.coords.Harvest.Mail.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("领取全部", e.coords.Harvest.Mail.ClaimAll),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("mail", steps)
}

// ExecuteStaminaGift 赠送体力
func (e *TaskExecutor) ExecuteStaminaGift() error {
	steps := []TaskStep{
		StepTap("进入体力赠送", e.coords.Harvest.StaminaGift.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("赠送好友", e.coords.Harvest.StaminaGift.Gift),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("领取赠送", e.coords.Harvest.StaminaGift.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("stamina_gift", steps)
}

// ExecuteRecruit 招募
func (e *TaskExecutor) ExecuteRecruit() error {
	steps := []TaskStep{
		StepTap("进入招募", e.coords.Harvest.Recruit.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("免费招募", e.coords.Harvest.Recruit.RecruitFree),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepConfirm("确认招募"),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("recruit", steps)
}

// ===== 周常任务 =====

// ExecutePractice 修行之路
func (e *TaskExecutor) ExecutePractice() error {
	steps := []TaskStep{
		StepTap("进入修行之路", e.coords.Weekly.Practice.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("开始章节", e.coords.Weekly.Practice.Start),
		StepWait("等待战斗开始", e.coords.WaitTimes.BattleStart),
		StepTap("开启自动", e.coords.Weekly.Practice.Auto),
		StepWait("等待战斗结束", e.coords.WaitTimes.BattleEnd),
		StepTap("领取奖励", e.coords.Weekly.Practice.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("practice", steps)
}

// ExecuteChaseAkatsuki 追击晓组织
func (e *TaskExecutor) ExecuteChaseAkatsuki() error {
	steps := []TaskStep{
		StepTap("进入追击晓组织", e.coords.Weekly.ChaseAkatsuki.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("选择成员", e.coords.Weekly.ChaseAkatsuki.SelectType),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("开始战斗", e.coords.Weekly.ChaseAkatsuki.Start),
		StepWait("等待战斗开始", e.coords.WaitTimes.BattleStart),
		StepTap("开启自动", e.coords.Weekly.ChaseAkatsuki.Auto),
		StepWait("等待战斗结束", e.coords.WaitTimes.BattleEnd),
		StepTap("领取奖励", e.coords.Weekly.ChaseAkatsuki.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("chase_akatsuki", steps)
}

// ExecuteRebel 叛忍来袭
func (e *TaskExecutor) ExecuteRebel() error {
	steps := []TaskStep{
		StepTap("进入叛忍来袭", e.coords.Weekly.Rebel.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("加入战斗", e.coords.Weekly.Rebel.Join),
		StepWait("等待战斗开始", e.coords.WaitTimes.BattleStart),
		StepTap("开启自动", e.coords.Weekly.Rebel.Auto),
		StepWait("等待战斗结束", e.coords.WaitTimes.BattleEnd),
		StepTap("领取奖励", e.coords.Weekly.Rebel.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("rebel", steps)
}

// ExecuteGuildFortress 组织要塞
func (e *TaskExecutor) ExecuteGuildFortress() error {
	steps := []TaskStep{
		StepTap("进入组织", e.coords.MainMenu.BottomNav["guild"]),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("进入要塞", e.coords.Weekly.GuildFortress.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("加入战斗", e.coords.Weekly.GuildFortress.Join),
		StepWait("等待战斗开始", e.coords.WaitTimes.BattleStart),
		StepTap("开启自动", e.coords.Weekly.GuildFortress.Auto),
		StepWait("等待战斗结束", e.coords.WaitTimes.BattleEnd),
		StepTap("领取奖励", e.coords.Weekly.GuildFortress.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("guild_fortress", steps)
}

// ExecuteBattlefield 天地战场
func (e *TaskExecutor) ExecuteBattlefield() error {
	steps := []TaskStep{
		StepTap("进入天地战场", e.coords.Weekly.Battlefield.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("选择阵营", e.coords.Weekly.Battlefield.SelectType),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("加入战斗", e.coords.Weekly.Battlefield.Join),
		StepWait("等待战斗开始", e.coords.WaitTimes.BattleStart),
		StepTap("开启自动", e.coords.Weekly.Battlefield.Auto),
		StepWait("等待战斗结束", e.coords.WaitTimes.BattleEnd),
		StepTap("领取奖励", e.coords.Weekly.Battlefield.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("battlefield", steps)
}

// ===== 活动任务 =====

// ExecuteBBQ 丁次烤肉
func (e *TaskExecutor) ExecuteBBQ() error {
	steps := []TaskStep{
		StepTap("进入丁次烤肉", e.coords.Event.BBQ.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("开始烤肉", e.coords.Event.BBQ.Start),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("喂肉", e.coords.Event.BBQ.Feed),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("领取奖励", e.coords.Event.BBQ.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("bbq", steps)
}

// ExecuteSakura 樱花季
func (e *TaskExecutor) ExecuteSakura() error {
	steps := []TaskStep{
		StepTap("进入樱花季", e.coords.Event.Sakura.Enter),
		StepWait("等待加载", e.coords.WaitTimes.PageLoad),
		StepTap("收集花瓣", e.coords.Event.Sakura.Collect),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("兑换奖励", e.coords.Event.Sakura.Exchange),
		StepWait("等待", e.coords.WaitTimes.Animation),
		StepTap("领取", e.coords.Event.Sakura.Claim),
		StepWait("等待", e.coords.WaitTimes.RewardClaim),
		StepBack("返回主界面"),
	}

	return e.ExecuteTask("sakura", steps)
}

// ===== 执行所有日常 =====

// ExecuteAllDaily 执行所有日常任务
func (e *TaskExecutor) ExecuteAllDaily(ctx context.Context) error {
	dailyTasks := []func() error{
		e.ExecuteTeamRaid,
		e.ExecuteBountyHall,
		e.ExecuteGuildPray,
		e.ExecuteSurvival,
		e.ExecuteEquipment,
		e.ExecuteTaskHall,
		e.ExecuteSecretRealm,
		e.ExecuteShop,
	}

	for _, task := range dailyTasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := task(); err != nil {
				e.log.Error("日常任务失败", "error", err)
				// 继续执行下一个任务
			}
			e.Wait(e.coords.WaitTimes.Animation)
		}
	}

	return nil
}

// ExecuteAllHarvest 执行所有收获任务
func (e *TaskExecutor) ExecuteAllHarvest(ctx context.Context) error {
	harvestTasks := []func() error{
		e.ExecuteLuckyMoney,
		e.ExecuteIntelAgency,
		e.ExecuteRanking,
		e.ExecuteActivityBox,
		e.ExecuteMonthlySign,
		e.ExecuteNinjaPass,
		e.ExecuteShare,
		e.ExecuteMail,
		e.ExecuteStaminaGift,
		e.ExecuteRecruit,
	}

	for _, task := range harvestTasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := task(); err != nil {
				e.log.Error("收获任务失败", "error", err)
			}
			e.Wait(e.coords.WaitTimes.Animation)
		}
	}

	return nil
}

// ExecuteAllWeekly 执行所有周常任务
func (e *TaskExecutor) ExecuteAllWeekly(ctx context.Context) error {
	weeklyTasks := []func() error{
		e.ExecutePractice,
		e.ExecuteChaseAkatsuki,
		e.ExecuteRebel,
		e.ExecuteGuildFortress,
		e.ExecuteBattlefield,
	}

	for _, task := range weeklyTasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := task(); err != nil {
				e.log.Error("周常任务失败", "error", err)
			}
			e.Wait(e.coords.WaitTimes.Animation)
		}
	}

	return nil
}