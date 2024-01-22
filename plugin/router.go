package plugin

import (
	"github.com/clouddea/koa-go/koa"
	"net/http"
	"strings"
)

/** 最简路由实现 */
func NewRouter(mapping map[string]koa.Plugin) koa.PluginMultiArg {
	return func(context *koa.Context, next func()) {
		items := strings.Split(context.Req.URL.Path, "/")
		var candidates []string
		var path = ""
		for i := 0; i < len(items); i++ {
			if i > 0 {
				path = path + "/"
				candidates = append(candidates, path)
			}
			path = path + items[i]
			candidates = append(candidates, path)
		}
		// 从后往前，最长匹配
		for i := len(candidates) - 1; i >= 0; i-- {
			if v, ok := mapping[candidates[i]]; ok {
				switch v.(type) {
				case koa.PluginSingleArg:
					v.(koa.PluginSingleArg)(context)
				case koa.PluginMultiArg:
					v.(koa.PluginMultiArg)(context, next)
				}
				return
			}
		}
		// 没有匹配，响应404
		context.Response.SetStatus(http.StatusNotFound)
		next()
	}
}
