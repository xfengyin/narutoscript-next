package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/xfengyin/narutoscript-next/internal/app"
	"github.com/xfengyin/narutoscript-next/internal/automation"
	"github.com/xfengyin/narutoscript-next/internal/server"
	"github.com/xfengyin/narutoscript-next/pkg/logger"
)

var (
	Version   = "1.0.0"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func main() {
	log := logger.New()
	defer log.Sync()

	printBanner(log)

	cfg, err := app.LoadConfig()
	if err != nil {
		log.Fatal("加载配置失败", "error", err)
	}

	application := app.New(cfg, log)
	scheduler := automation.NewScheduler(application.Device, application.Vision, log, application)
	srv := server.New(application, scheduler)

	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		log.Info("启动服务", "addr", addr, "version", Version)
		if err := srv.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatal("服务启动失败", "error", err)
		}
	}()

	if cfg.Server.AutoOpenBrowser {
		go func() {
			time.Sleep(500 * time.Millisecond)
			url := fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
			openBrowser(url)
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("正在关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv.Shutdown(ctx)
	scheduler.Stop()

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
	log.Info("NarutoScript Next 启动中...", "version", Version, "go", runtime.Version(), "os", runtime.GOOS+"/"+runtime.GOARCH)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return
	}
	cmd.Start()
}