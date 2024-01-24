package controller

import (
	"database/sql"
	"github.com/clouddea/koa-go/example/helloworld/dao"
	"github.com/clouddea/koa-go/example/helloworld/service"
	"github.com/clouddea/koa-go/koa"
	"github.com/clouddea/koa-go/plugin"
)

func User_Register_Controller(ctx *koa.Context) {
	db := ctx.Attr["sqlite3"].(*sql.DB)
	service.User_Register_Service(db, dao.User{
		Nickname: "test",
	})
}

func User_Update_Controller(ctx *koa.Context) {
	db := ctx.Attr["sqlite3"].(*sql.DB)
	auth := ctx.Attr["auth"].(plugin.Auth)
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
	"/user/update": koa.PluginSingleArg(User_Update_Controller),
}
