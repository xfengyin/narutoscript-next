# API 接口文档

## 基础信息

- 基础 URL: `http://localhost:8080/api`
- 内容类型: `application/json`

## 接口列表

### 获取状态

```
GET /api/status
```

响应:
```json
{
  "running": true,
  "version": "1.0.0"
}
```

### 获取统计数据

```
GET /api/stats
```

响应:
```json
{
  "gold": 12450,
  "copper": 89320,
  "stamina": 145,
  "tasks_done": 8,
  "tasks_total": 25
}
```

### 启动任务

```
POST /api/start
```

响应:
```json
{
  "message": "任务已启动"
}
```

### 停止任务

```
POST /api/stop
```

响应:
```json
{
  "message": "任务已停止"
}
```

### 获取任务列表

```
GET /api/tasks
```

响应:
```json
[
  {
    "name": "team_raid",
    "display": "小队突袭",
    "category": "daily",
    "status": "idle"
  },
  ...
]
```

### 执行指定任务

```
POST /api/task/:name/run
```

响应:
```json
{
  "message": "任务已启动",
  "task": "team_raid"
}
```

### 获取日志

```
GET /api/logs
```

响应:
```json
[
  {
    "time": "14:30:01",
    "level": "info",
    "message": "服务已启动"
  }
]
```

### 实时日志流

```
GET /api/logs/stream
```

WebSocket 连接，实时推送状态更新。

消息格式:
```json
{
  "type": "state",
  "data": {
    "running": true,
    "tasks": {...},
    "stats": {...}
  }
}
```