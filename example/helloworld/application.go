package main

import (
	"fmt"
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
	configPlugin, cfg := plugin.NewConfig("./application.yaml", config.Config{})
	dbPlugin, db := plugin.NewSqlite(app, cfg.Database)
	auth := config.NewABACAuth()
	dao.DAOInit(db)
	app.Use(configPlugin)
	app.Use(dbPlugin)
	app.Use(plugin.NewJsonParser(cfg.JsonMaxSize))
	app.Use(auth)
	app.Use(logger)
	app.Use(router)
	fmt.Printf("server %v has started on port %v \n", cfg.Server.Name, cfg.Server.Port)
	app.Listen(cfg.Server.Port)
}
