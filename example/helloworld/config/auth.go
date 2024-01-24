package config

import (
	"github.com/clouddea/koa-go/example/helloworld/dao"
	"github.com/clouddea/koa-go/koa"
	"github.com/clouddea/koa-go/plugin"
	"reflect"
)

const AUTH_OPERATION_URL_CREATE = "AUTH_OPERATION_URL_CREATE"
const AUTH_OPERATION_URL_QUERY = "AUTH_OPERATION_URL_QUERY"
const AUTH_OPERATION_URL_UPDATE = "AUTH_OPERATION_URL_UPDATE"
const AUTH_OPERATION_URL_DELETE = "AUTH_OPERATION_URL_DELETE"

func subjectFactory(ctx *koa.Context) any {
	user := dao.User{
		Id:       1,
		Role:     dao.USER_ROLE_TYPE_USER,
		Nickname: "test",
	}
	return map[string]any{
		"id":   user.Id,
		"role": user.Role,
		"name": user.Nickname,
	}
}

func URLPolicy(subject any, environment any, object any, operation any) bool {
	sub := subject.(map[string]any)
	if sub["role"] == dao.USER_ROLE_TYPE_ADMIN {
		return true
	}
	switch operation {
	case AUTH_OPERATION_URL_CREATE:
		return true
	case AUTH_OPERATION_URL_QUERY, AUTH_OPERATION_URL_UPDATE, AUTH_OPERATION_URL_DELETE:
		return sub["id"] == object.(dao.URL).Owner
	}
	return false
}

func UserPolicy(subject any, environment any, object any, operation any) bool {
	sub := subject.(map[string]any)
	return sub["role"] == dao.USER_ROLE_TYPE_ADMIN || sub["id"] == object.(dao.User).Id
}

func NewABACAuth() koa.PluginMultiArg {
	return plugin.NewAuth(subjectFactory,
		func(ctx *koa.Context) any {
			return nil
		},
		make(map[reflect.Type]plugin.AuthMap),
		map[reflect.Type]plugin.AuthPolicy{
			reflect.TypeOf(dao.URL{}):  URLPolicy,
			reflect.TypeOf(dao.User{}): UserPolicy,
		},
	)
}
