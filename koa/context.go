package koa

import "net/http"

/** 应用上下文，存储例如db，request等数据 */
type Context struct {
	app      *Koa
	Req      *http.Request
	Res      http.ResponseWriter
	Request  *KoaRequest
	Response *KoaResponse
	cookies  any // TODO: 实现
}

func (this *Context) Throw(code int) {

}

func (this *Context) Assert(state bool, code int) {

}
