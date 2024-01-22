package plugin

import "github.com/clouddea/koa-go/koa"

/** ABAC */
func NewAuth(login string, logout string) koa.PluginMultiArg {
	// TODO: implement
	return func(ctx *koa.Context, next func()) {

	}
}
