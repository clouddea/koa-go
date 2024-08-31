package plugin

import (
	"database/sql"
	"fmt"
	"github.com/clouddea/koa-go/koa"
	"github.com/clouddea/koa-go/koa/util"
	_ "github.com/mattn/go-sqlite3"
)

func NewSqlite(app *koa.Koa, filename string) (koa.PluginMultiArg, *sql.DB) {
	sqlite3, err := sql.Open("sqlite3", filename)
	util.Assert(err, "open sqlite3 error")
	fmt.Println("sqlite3 opened")
	app.On(koa.KOA_EVENT_CLOSE, func(arg ...any) {
		_ = sqlite3.Close()
		fmt.Println("sqlite3 closed")
	})
	return func(ctx *koa.Context, next func()) {
		ctx.State["sqlite3"] = sqlite3
		next()
	}, sqlite3
}
