// Package app 全局应用实例，供 repository、handlers 层调用
package app

import (
	"github.com/gofiber/fiber/v3"
	"github.com/muzidudu/go_utils/cache"
	"github.com/muzidudu/go_utils/fiber/bootstrap"
	"gorm.io/gorm"
)

var global *bootstrap.App

// Set 设置全局应用（bootstrap 启动时调用）
func Set(a *bootstrap.App) {
	global = a
}

// Get 获取全局应用
func Get() *bootstrap.App {
	return global
}

// Config 全局配置
func Config() *bootstrap.Config {
	if global == nil {
		return nil
	}
	return global.Config
}

// Cache 全局缓存
func Cache() cache.Cache {
	if global == nil {
		return nil
	}
	return global.Cache
}

// DB 全局数据库
func DB() *gorm.DB {
	if global == nil {
		return nil
	}
	return global.DB
}

// Fiber 全局 Fiber 实例
func Fiber() *fiber.App {
	if global == nil {
		return nil
	}
	return global.Fiber
}
