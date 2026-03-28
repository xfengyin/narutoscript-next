package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/xfengyin/narutoscript-next/internal/app"
	"github.com/xfengyin/narutoscript-next/internal/server"
	"github.com/xfengyin/narutoscript-next/pkg/logger"
)

// 构建时注入的版本信息
var (
	Version   = "1.0.0"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

//go:embed all:internal/ui/dist
var staticFS embed.FS

func main() {
	// 初始化日志
	log := logger.New()
	defer log.Sync()

	// 打印启动信息
	printBanner(log)

	// 加载配置
	cfg, err := app.LoadConfig()
	if err != nil {
		log.Fatal("加载配置失败", "error", err)
	}

	// 创建应用
	application := app.New(cfg, log)

	// 启动服务
	srv := server.New(application, staticFS)

	// 启动 HTTP 服务
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		log.Info("启动服务", "addr", addr, "version", Version)
		if err := srv.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatal("服务启动失败", "error", err)
		}
	}()

	// 自动打开浏览器
	if cfg.Server.AutoOpenBrowser {
		go func() {
			time.Sleep(500 * time.Millisecond)
			url := fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
			if err := openBrowser(url); err != nil {
				log.Warn("无法自动打开浏览器", "error", err)
			}
		}()
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("正在关闭服务...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("服务关闭错误", "error", err)
	}

	// 停止任务调度器
	application.StopScheduler()

	log.Info("服务已关闭")
}

func printBanner(log *logger.Logger) {
	banner := `
 _   _                      _   _          _  _____           _ 
| \ | | _____  ___   _ ___ | |_| |__   ___| ||_   _|__   ___ | |
|  \| |/ _ \ \/ / | | / __|| __| '_ \ / _ \ __|| |/ _ \ / _ \| |
| |\  |  __/>  <| |_| \__ \| |_| | | |  __/ |_ | | (_) | (_) | |
|_| \_|\___/_/\_\\__,_|___/ \__|_| |_|\___|\__||_|\___/ \___/|_|
                                                                
`
	fmt.Println(banner)
	log.Info("NarutoScript Next 启动中...",
		"version", Version,
		"commit", GitCommit,
		"go", runtime.Version(),
		"os", runtime.GOOS+"/"+runtime.GOARCH,
	)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
	return cmd.Start()
}