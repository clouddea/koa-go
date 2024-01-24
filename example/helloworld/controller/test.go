package controller

import (
	"database/sql"
	"github.com/clouddea/koa-go/example/helloworld/config"
	"github.com/clouddea/koa-go/example/helloworld/dao"
	"github.com/clouddea/koa-go/koa"
	"github.com/clouddea/koa-go/plugin"
)

var TestController = map[string]koa.Plugin{
	"/test/benchmark": koa.PluginSingleArg(func(context *koa.Context) {
		_ = context.Response.Write([]byte("APP NAME is " + context.State["config"].(config.Config).Server.Name))
	}),
	"/test/hello/world": koa.PluginSingleArg(func(context *koa.Context) {
		_ = context.Response.Write([]byte("hello world"))
	}),
	"/test/sqlite3": koa.PluginSingleArg(func(context *koa.Context) {
		context.Assert(dao.Test_DAO_Add(context.State["sqlite3"].(*sql.DB)) == nil, 500)
		_ = context.Response.Write([]byte("add success"))
	}),
	"/": plugin.NewStatic("/", "./www"),
}
