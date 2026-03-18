package routes

import "github.com/muzidudu/go_utils/fiber/bootstrap"

type Router interface {
	InstallRouter(app *bootstrap.App)
}
