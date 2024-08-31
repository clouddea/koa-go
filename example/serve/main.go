package main

import (
	"fmt"
	"github.com/clouddea/koa-go/koa"
	"github.com/clouddea/koa-go/koa/util"
	"github.com/clouddea/koa-go/plugin"
	"os"
	"strconv"
	"strings"
)

func main() {
	var args map[string]string = make(map[string]string)
	for i := 1; i < len(os.Args); i++ {
		frags := strings.Split(os.Args[i], "=")
		if len(frags) == 1 {
			args[frags[0]] = ""
		} else {
			args[frags[0]] = frags[1]
		}
	}
	if _, ok := args["--help"]; ok {
		fmt.Println("USAGE: serve [--port=<port>] [--dir=<directory>] [--help]")
		return
	}
	dir := "."
	port := 8080
	if v, ok := args["--port"]; ok && v != "" {
		_port, err := strconv.Atoi(v)
		util.Assert(err, "port "+v+"is invalid")
		port = _port
	}
	if v, ok := args["--dir"]; ok && v != "" {
		dir = v
	}

	router := plugin.NewRouter(map[string]koa.Plugin{
		"/": plugin.NewStatic("/", dir),
	})
	logger := plugin.NewLogger(true)
	app := koa.NewKoa(nil)
	app.Use(logger)
	app.Use(router)
	fmt.Printf("served in port=%v, directory=%v, ...", port, dir)
	app.Listen(port)
}
