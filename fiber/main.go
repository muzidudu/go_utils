// Package main Fiber v3 快速启动框架
// 集成: configmgr, cache, gorm postgresql, 优雅关闭, bootstrap 启动
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/muzidudu/go_utils/fiber/bootstrap"
	"github.com/muzidudu/go_utils/fiber/internal/routes"
)

func main() {
	// 1. Bootstrap 启动：加载配置、缓存、数据库
	app, err := bootstrap.New("config/config.yaml")
	if err != nil {
		log.Fatalf("bootstrap: %v", err)
	}
	defer app.Close()

	// 2. 安装路由
	routes.InstallRouter(app)

	// 3. 监听地址与 Listen 配置
	port := 3000
	if app.Config != nil && app.Config.Server.Port > 0 {
		port = app.Config.Server.Port
	}
	addr := fmt.Sprintf(":%d", port)
	listenCfg := fiber.ListenConfig{}
	if app.Config != nil && app.Config.Server.Debug {
		listenCfg.EnablePrintRoutes = true
	}

	// 4. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Gracefully shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := app.Fiber.ShutdownWithContext(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("Server listening on %s", addr)
	if err := app.Fiber.Listen(addr, listenCfg); err != nil {
		log.Fatalf("listen: %v", err)
	}
	log.Println("Server stopped")
}
