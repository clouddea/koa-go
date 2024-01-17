package main

import (
	"fmt"
	"github.com/clouddea/koa-go/koa"
	"github.com/clouddea/koa-go/plugin"
)

func main() {
	router := plugin.NewRouter(map[string]func(context *koa.Context){
		"/test": func(context *koa.Context) {
			_, err := fmt.Fprintln(context.Res, "test")
			if err != nil {
				panic(err)
			}
		},
		"/hello/world": func(context *koa.Context) {
			_, err := fmt.Fprintln(context.Res, "hello world")
			if err != nil {
				panic(err)
			}
		},
		"/static": plugin.NewStatic("/static", "./example/helloworld/www"),
	})
	logger := plugin.NewLogger()
	app := koa.NewKoa(nil)
	app.Use(logger)
	app.Use(router)
	app.Listen(8080)
}
