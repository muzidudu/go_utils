// Package routes 路由安装
package routes

import (
	"github.com/muzidudu/go_utils/fiber/bootstrap"
)

// InstallRouter 安装所有路由到应用
func InstallRouter(app *bootstrap.App) {
	InstallHTTPRoutes(app)
	InstallAPIRoutes(app)
}
