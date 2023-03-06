package bootstrap

import (
	"github.com/gorilla/mux"
	"goblog/pkg/route"
	"goblog/routes"
)

// SetupRoute 路由初始化
func SetupRoute() *mux.Router {
	router := mux.NewRouter()
	// 注册网页相关路由
	routes.RegisterWebRoutes(router)
	// 设置路由实例，以供 route 包里面一些函数使用
	route.SetRoute(router)
	return router
}
