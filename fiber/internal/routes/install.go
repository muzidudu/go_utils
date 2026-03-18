package routes

import "github.com/muzidudu/go_utils/fiber/bootstrap"

// InstallRouter 安装所有路由到应用
func InstallRouter(app *bootstrap.App) {
	setup(app, NewAPIRoute(), NewHTTPRoute())
}

// setup 安装路由
func setup(app *bootstrap.App, router ...Router) {
	for _, r := range router {
		r.InstallRouter(app)
	}
}
