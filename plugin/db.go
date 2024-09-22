package plugin

import (
	"database/sql"
	"github.com/clouddea/koa-go/koa"
	"github.com/clouddea/koa-go/koa/util"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func NewSqlite(app *koa.Koa, filename string) (koa.PluginMultiArg, *sql.DB) {
	sqlite3, err := sql.Open("sqlite3", filename)
	util.Assert(err, "open sqlite3 error")
	log.Println("sqlite3 opened")
	app.On(koa.KOA_EVENT_CLOSE, func(arg ...any) {
		_ = sqlite3.Close()
		log.Println("sqlite3 closed")
	})
	return func(ctx *koa.Context, next func()) {
		ctx.State["sqlite3"] = sqlite3
		next()
	}, sqlite3
}
