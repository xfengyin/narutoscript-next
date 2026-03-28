package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xfengyin/narutoscript-next/internal/app"
	"github.com/xfengyin/narutoscript-next/internal/server"
	"github.com/xfengyin/narutoscript-next/pkg/logger"
)

//go:embed all:internal/ui/dist
var staticFS embed.FS

func main() {
	// 初始化日志
	log := logger.New()
	defer log.Sync()

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
		log.Info("启动服务", "addr", addr)
		if err := srv.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatal("服务启动失败", "error", err)
		}
	}()

	// 自动打开浏览器
	if cfg.Server.AutoOpenBrowser {
		go func() {
			time.Sleep(500 * time.Millisecond)
			openBrowser(fmt.Sprintf("http://localhost:%d", cfg.Server.Port))
		}()
	}

	log.Info("NarutoScript Next 已启动", "version", "1.0.0")

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

	log.Info("服务已关闭")
}

func openBrowser(url string) {
	// Windows
	exec.Command("cmd", "/c", "start", url).Start()
}