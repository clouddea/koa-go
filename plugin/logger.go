package plugin

import (
	"github.com/clouddea/koa-go/koa"
	"log"
)

/** 最简日志实现 */
func NewLogger() koa.PluginMultiArg {
	return func(context *koa.Context, next func()) {
		defer func() {
			err := recover()
			if err != nil {
				log.Printf("[ERROR] %v", err)
			}
		}()
		next()
		log.Printf("[%v] [%v] %v %v",
			context.Req.Method,
			context.Response.GetStatus(),
			context.Req.URL.Path,
			context.Req.URL.RawQuery)
	}
}
