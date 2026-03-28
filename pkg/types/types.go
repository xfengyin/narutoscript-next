package types

// Point 坐标点
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Rect 矩形区域
type Rect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// TaskResult 任务执行结果
type TaskResult struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GameData 游戏数据
type GameData struct {
	Gold       int64 `json:"gold"`
	Copper     int64 `json:"copper"`
	Stamina    int   `json:"stamina"`
	MaxStamina int   `json:"max_stamina"`
	Level      int   `json:"level"`
	VIP        int   `json:"vip"`
}

// LogLevel 日志级别
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogEntry 日志条目
type LogEntry struct {
	Time    string   `json:"time"`
	Level   LogLevel `json:"level"`
	Message string   `json:"message"`
}

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusIdle    TaskStatus = "idle"
	TaskStatusRunning TaskStatus = "running"
	TaskStatusSuccess TaskStatus = "success"
	TaskStatusFailed  TaskStatus = "failed"
	TaskStatusWaiting TaskStatus = "waiting"
)