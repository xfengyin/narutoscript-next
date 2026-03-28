# NarutoScript Next

> 火影忍者手游自动化脚本 - 下一代版本

🍥 基于 Go + React 构建的轻量级游戏自动化工具

## ✨ 特性

- 🚀 **极致轻量** - 打包后仅 ~20MB（原项目 1/10）
- ⚡ **快速启动** - 毫秒级启动，无需等待
- 🎨 **现代界面** - React + Tailwind CSS 简洁 UI
- 📦 **零配置** - 下载即用，无需安装 Python
- 🔒 **单文件** - 所有资源内嵌，一个 EXE 搞定
- 🌍 **跨平台** - 支持 Windows / macOS / Linux

## 📋 功能列表

### 日常任务（8个）
- ✅ 小队突袭
- ✅ 丰饶之间
- ✅ 组织祈福
- ✅ 生存试炼
- ✅ 装备扫荡
- ✅ 任务集会所
- ✅ 秘境挑战
- ✅ 商店购买

### 收获系统（10个）
- ✅ 招财、情报社、排行榜
- ✅ 活跃度宝箱、每月签到
- ✅ 忍法帖、每日分享
- ✅ 邮件、赠送体力、招募

### 周常任务（5个）
- ✅ 修行之路
- ✅ 追击晓组织
- ✅ 叛忍来袭
- ✅ 组织要塞
- ✅ 天地战场

### 限时活动（2个）
- ✅ 丁次烤肉
- ✅ 樱花季

## 🚀 快速开始

### 下载使用

1. 从 [Releases](https://github.com/xfengyin/narutoscript-next/releases) 下载最新版本
2. 解压到任意目录
3. 双击 `narutoscript-next.exe`
4. 自动打开浏览器访问 `http://localhost:8080`

### 系统要求

- Windows 10+ / macOS 10.15+ / Linux
- ADB 工具（用于连接模拟器/设备）
- 分辨率 1280x720（16:9 比例）

## 🛠️ 开发

### 环境要求

- Go 1.22+
- Node.js 18+
- ADB

### 本地开发

```bash
# 克隆仓库
git clone https://github.com/xfengyin/narutoscript-next.git
cd narutoscript-next

# 安装前端依赖
cd web
npm install

# 开发模式运行前端
npm run dev

# 另一个终端，运行后端
cd ..
go run ./cmd/main.go

# 构建前端
cd web && npm run build

# 构建后端（所有平台）
make build-all
```

### 项目结构

```
narutoscript-next/
├── cmd/main.go              # 入口
├── internal/
│   ├── app/                 # 应用核心
│   ├── server/              # Web 服务
│   ├── automation/          # 自动化任务
│   ├── device/              # 设备控制
│   ├── vision/              # 图像识别
│   ├── ocr/                 # 文字识别
│   └── ui/dist/             # 前端构建产物
├── web/                     # React 前端源码
├── pkg/                     # 公共包
└── .github/workflows/       # CI/CD
```

## 📄 许可证

GPLv3 - 详见 [LICENSE](LICENSE)

## 🙏 致谢

- 原项目 [NarutoScript](https://github.com/Elmyran/NarutoScript)
- [Alas](https://github.com/LmeSzinc/AzurLaneAutoScript) 自动化框架