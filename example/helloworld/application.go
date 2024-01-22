package main

import (
	config "github.com/clouddea/koa-go/example/helloworld/config"
	"github.com/clouddea/koa-go/example/helloworld/controller"
	"github.com/clouddea/koa-go/example/helloworld/dao"
	"github.com/clouddea/koa-go/koa"
	"github.com/clouddea/koa-go/plugin"
	"maps"
)

func main() {
	app := koa.NewKoa(nil)
	var routers map[string]koa.Plugin = make(map[string]koa.Plugin)
	maps.Copy(routers, controller.TestController)
	maps.Copy(routers, controller.URLController)
	maps.Copy(routers, controller.UserController)
	router := plugin.NewRouter(routers)
	logger := plugin.NewLogger(true)
	configPlugin := plugin.NewConfig("./application.yaml", config.Config{})
	dbPlugin, db := plugin.NewSqlite(app, "./data/store.db?_synchronous=0")
	dao.DAOInit(db)
	app.Use(configPlugin)
	app.Use(dbPlugin)
	app.Use(logger)
	app.Use(router)
	app.Listen(8080)
}
