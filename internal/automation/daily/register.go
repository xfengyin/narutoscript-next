package daily

import (
	"github.com/xfengyin/narutoscript-next/internal/automation"
)

// RegisterAll 注册所有日常任务
func RegisterAll(scheduler *automation.Scheduler) {
	scheduler.AddTask("team_raid", automation.NewBaseTask("小队突袭", "daily", TeamRaid))
	scheduler.AddTask("bounty_hall", automation.NewBaseTask("丰饶之间", "daily", BountyHall))
	scheduler.AddTask("guild_pray", automation.NewBaseTask("组织祈福", "daily", GuildPray))
	scheduler.AddTask("survival", automation.NewBaseTask("生存试炼", "daily", Survival))
	scheduler.AddTask("equipment", automation.NewBaseTask("装备扫荡", "daily", Equipment))
	scheduler.AddTask("task_hall", automation.NewBaseTask("任务集会所", "daily", TaskHall))
	scheduler.AddTask("secret_realm", automation.NewBaseTask("秘境挑战", "daily", SecretRealm))
	scheduler.AddTask("shop", automation.NewBaseTask("商店购买", "daily", Shop))
}