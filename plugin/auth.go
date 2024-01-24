package plugin

import (
	"github.com/clouddea/koa-go/koa"
	"reflect"
)

type Auth = func(object any, operation any) bool
type AuthMap = func(object any) any
type AuthPolicy = func(subject any, environment any, object any, operation any) bool

/** ABAC */
/** 该架构完美解耦游客和登录用户的权限，同时可以支持为游客设置权限，酷！ */
func NewAuth(
	subject func(ctx *koa.Context) any, // 如何获取登录用户
	environment func(ctx *koa.Context) any, // 如何获取环境对象
	objects map[reflect.Type]AuthMap, // 如何获取目标对象, 可以没有对应函数
	policies map[reflect.Type]AuthPolicy, // 如何进行决策，可以没有对应策略。其中object和operation都是可以模糊的，由应用层定义
) koa.PluginMultiArg {
	return func(ctx *koa.Context, next func()) {
		if subject == nil {
			panic("auth: subject factory can not be nil")
		}
		if environment == nil {
			panic("auth: environment factory can not be nil")
		}
		if objects == nil {
			panic("auth: object factory can not be nil")
		}
		if policies == nil {
			panic("auth: policies can not be nil")
		}
		sub := subject(ctx)
		env := environment(ctx)
		var auth Auth = func(object any, operation any) bool {
			obj := object // 默认不进行转换
			if factory, ok := objects[reflect.TypeOf(object)]; ok {
				obj = factory(object)
			}
			if policy, ok := policies[reflect.TypeOf(object)]; ok {
				if policy(sub, env, obj, operation) {
					return true
				}
			}
			return false
		}
		ctx.State["auth"] = auth
		next()
	}
}
