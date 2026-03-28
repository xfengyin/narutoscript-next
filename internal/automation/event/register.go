package event

import (
	"github.com/xfengyin/narutoscript-next/internal/automation"
)

// RegisterAll 注册所有活动任务
func RegisterAll(scheduler *automation.Scheduler) {
	scheduler.AddTask("bbq", automation.NewBaseTask("丁次烤肉", "event", BBQ))
	scheduler.AddTask("sakura", automation.NewBaseTask("樱花季", "event", Sakura))
}