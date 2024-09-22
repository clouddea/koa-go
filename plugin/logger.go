package plugin

import (
	"github.com/clouddea/koa-go/koa"
	"log"
	"log/slog"
	"net/http"
	"os"
)

/** 最简日志实现 */
func NewLogger(debug bool) koa.PluginMultiArg {
	return func(context *koa.Context, next func()) {
		defer func() {
			err := recover()
			if err != nil {
				context.Throw(http.StatusInternalServerError)
				log.Printf("[ERROR] %v", err)
			}
		}()
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		context.State["log"] = logger
		if !debug {
			next()
			return
		}
		next()
		logger.Info("[ACCESS]",
			"method", context.Req.Method,
			"status", context.Response.GetStatus(),
			"path", context.Req.URL.Path,
			"query", context.Req.URL.RawQuery)
	}
}
