package automation

import (
	"fmt"
)

// TaskFunc 任务执行函数类型
type TaskFunc func() error

// BaseTask 基础任务结构
type BaseTask struct {
	name     string
	category string
	execute  TaskFunc
}

func NewBaseTask(name, category string, fn TaskFunc) *BaseTask {
	return &BaseTask{
		name:     name,
		category: category,
		execute:  fn,
	}
}

func (t *BaseTask) Name() string {
	return t.name
}

func (t *BaseTask) Category() string {
	return t.category
}

func (t *BaseTask) Execute() error {
	if t.execute == nil {
		return fmt.Errorf("任务 %s 没有实现", t.name)
	}
	return t.execute()
}