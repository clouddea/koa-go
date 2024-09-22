package controller

import (
	"database/sql"
	"github.com/clouddea/koa-go/example/helloworld/dao"
	"github.com/clouddea/koa-go/example/helloworld/service"
	"github.com/clouddea/koa-go/koa"
	"github.com/clouddea/koa-go/plugin"
)

func User_Register_Controller(ctx *koa.Context) {
	if input, ok := ctx.Request.Body.(map[string]any); ok {
		if _, ok := input["name"]; !ok {
			ctx.Response.Body = dao.ResponseError("请输入用户名")
			return
		}
		name := input["name"].(string)
		db := ctx.State["sqlite3"].(*sql.DB)
		service.User_Register_Service(db, dao.User{
			Nickname: name,
		})
		ctx.Response.Body = dao.ResponseSuccess()
	} else {
		ctx.Response.Body = dao.ResponseError("请输入数据")
	}
}

func User_Login_Controller(ctx *koa.Context) {
	if input, ok := ctx.Request.Body.(map[string]any); ok {
		if _, ok := input["name"]; !ok {
			ctx.Response.Body = dao.ResponseError("请输入用户名")
			return
		}
		var session = ctx.State["session"].(*plugin.Session)
		if _, ok := session.Val["logined"]; ok {
			ctx.Response.Body = dao.ResponseDataSuccess(map[string]any{
				"name":    input["name"].(string),
				"logined": session.Val["logined"],
			})
			return
		}

		ctx.Response.Body = dao.ResponseDataSuccess(map[string]any{
			"name":    input["name"].(string),
			"logined": false,
		})
		session.Val["logined"] = true

	} else {
		ctx.Response.Body = dao.ResponseError("请输入数据")
	}
}

func User_Update_Controller(ctx *koa.Context) {
	db := ctx.State["sqlite3"].(*sql.DB)
	auth := ctx.State["auth"].(plugin.Auth)
	if user, ok := dao.User_Query(db, 1); ok {
		if auth(user, nil) {
			service.User_Update_Service(db, dao.User{
				Id:       1,
				Nickname: "test3",
			})
		} else {
			ctx.Throw(403)
		}
	}
}

var UserController = map[string]koa.Plugin{
	"/register":    koa.PluginSingleArg(User_Register_Controller),
	"/login":       koa.PluginSingleArg(User_Login_Controller),
	"/user/update": koa.PluginSingleArg(User_Update_Controller),
}
