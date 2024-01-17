package koa

type PluginSingleArg func(ctx *Context)
type PluginMultiArg func(ctx *Context, next func())

type Plugin interface {
	Call(ctx *Context)
}

func (this PluginSingleArg) Call(ctx *Context) { this(ctx) }
func (this PluginMultiArg) Call(ctx *Context)  { this(ctx, nil) }
