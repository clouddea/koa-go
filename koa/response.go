package koa

import "net/http"

type KoaResponse struct {
	context       *Context
	hasSentHeader bool
	status        int
	Body          any
}

func (this *KoaResponse) SetStatus(status int) {
	this.status = status
}

func (this *KoaResponse) GetStatus() int {
	return this.status
}

func (this *KoaResponse) Write(bytes []byte) error {
	if !this.hasSentHeader {
		this.hasSentHeader = true
		this.context.Res.WriteHeader(this.status)
	}
	_, err := this.context.Res.Write(bytes)
	return err
}

func (this *KoaResponse) Header() http.Header {
	return this.context.Res.Header()
}
