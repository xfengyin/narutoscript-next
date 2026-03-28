package weekly

import (
	"github.com/xfengyin/narutoscript-next/internal/automation"
)

// RegisterAll 注册所有周常任务
func RegisterAll(scheduler *automation.Scheduler) {
	scheduler.AddTask("practice", automation.NewBaseTask("修行之路", "weekly", Practice))
	scheduler.AddTask("chase_akatsuki", automation.NewBaseTask("追击晓组织", "weekly", ChaseAkatsuki))
	scheduler.AddTask("rebel", automation.NewBaseTask("叛忍来袭", "weekly", Rebel))
	scheduler.AddTask("guild_fortress", automation.NewBaseTask("组织要塞", "weekly", GuildFortress))
	scheduler.AddTask("battlefield", automation.NewBaseTask("天地战场", "weekly", Battlefield))
}