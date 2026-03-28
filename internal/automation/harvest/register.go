package harvest

import (
	"github.com/xfengyin/narutoscript-next/internal/automation"
)

// RegisterAll 注册所有收获任务
func RegisterAll(scheduler *automation.Scheduler) {
	scheduler.AddTask("lucky_money", automation.NewBaseTask("招财", "harvest", LuckyMoney))
	scheduler.AddTask("intel_agency", automation.NewBaseTask("情报社", "harvest", IntelAgency))
	scheduler.AddTask("ranking", automation.NewBaseTask("排行榜", "harvest", Ranking))
	scheduler.AddTask("activity_box", automation.NewBaseTask("活跃度宝箱", "harvest", ActivityBox))
	scheduler.AddTask("monthly_sign", automation.NewBaseTask("每月签到", "harvest", MonthlySign))
	scheduler.AddTask("ninja_pass", automation.NewBaseTask("忍法帖", "harvest", NinjaPass))
	scheduler.AddTask("share", automation.NewBaseTask("每日分享", "harvest", Share))
	scheduler.AddTask("mail", automation.NewBaseTask("邮件", "harvest", Mail))
	scheduler.AddTask("stamina_gift", automation.NewBaseTask("赠送体力", "harvest", StaminaGift))
	scheduler.AddTask("recruit", automation.NewBaseTask("招募", "harvest", Recruit))
}