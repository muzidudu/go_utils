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
	"github.com/muzidudu/go_utils/fiber/internal/app"
	"github.com/muzidudu/go_utils/fiber/internal/models"
	"github.com/muzidudu/go_utils/fiber/internal/repository"
	"github.com/muzidudu/go_utils/fiber/internal/routes"
)

func main() {
	// 1. Bootstrap 启动：加载配置、缓存、数据库
	a, err := bootstrap.New("config/config.yaml")
	if err != nil {
		log.Fatalf("bootstrap: %v", err)
	}
	defer a.Close()

	// 2. 设置全局 app（供 repository、handlers、middlewares 调用）
	app.Set(a)

	// 4. 数据库自动迁移
	if err := a.Migrate(&models.User{}, &models.Site{}); err != nil {
		log.Printf("warn: migrate: %v", err)
	}

	// 4.1 若无站点则创建默认站点
	seedDefaultSite(a)

	// 5. 安装路由
	routes.InstallRouter(a)

	// 6. 监听地址与 Listen 配置
	port := 3000
	if a.Config != nil && a.Config.Server.Port > 0 {
		port = a.Config.Server.Port
	}
	addr := fmt.Sprintf(":%d", port)
	listenCfg := fiber.ListenConfig{}
	if a.Config != nil && a.Config.Server.Debug {
		listenCfg.EnablePrintRoutes = true
	}

	// 7. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Gracefully shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.Fiber.ShutdownWithContext(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("Server listening on %s", addr)
	if err := a.Fiber.Listen(addr, listenCfg); err != nil {
		log.Fatalf("listen: %v", err)
	}
	log.Println("Server stopped")
}

// seedDefaultSite 若无站点则创建默认站点
func seedDefaultSite(a *bootstrap.App) {
	if a.DB == nil {
		return
	}
	list, err := repository.Site.List()
	if err != nil || len(list) > 0 {
		return
	}
	defaultSite := &models.Site{
		Name:      "default",
		Domain:    "localhost",
		Template:  "default",
		IsDefault: true,
		Status:    1,
	}
	if err := repository.Site.Create(defaultSite); err != nil {
		log.Printf("warn: seed default site: %v", err)
		return
	}
	log.Println("seeded default site: localhost")
}
