package koa

import (
	eventbus "github.com/asaskevich/EventBus"
	"log"
	"net/http"
	"strconv"
)

const (
	KOA_EVENT_ERROR = "error"
)

type Koa struct {
	config  *KoaConfig
	plugins []Plugin
	bus     eventbus.Bus
}

func NewKoa(config *KoaConfig) *Koa {
	if config == nil {
		config = NewKoaConfig()
	}
	return &Koa{
		config:  config,
		plugins: make([]Plugin, 0),
		bus:     eventbus.New(),
	}
}

func (this *Koa) Listen(port int) {
	http.HandleFunc("/", this.dispatcher)
	endpoint := "0.0.0.0:" + strconv.Itoa(port)
	err := http.ListenAndServe(endpoint, nil)
	if err != nil {
		this.Emit(KOA_EVENT_ERROR, err)
		log.Fatalln("listen on "+endpoint+" error", err)
	}
}

func (this *Koa) Use(plugin Plugin) {
	this.plugins = append(this.plugins, plugin)
}

func (this *Koa) On(event string, handler func(arg ...any)) {
	err := this.bus.Subscribe(event, handler)
	if err != nil {
		return
	}
}

func (this *Koa) Emit(event string, args ...any) {
	this.bus.Publish(event, args...)
}

func (this *Koa) dispatcher(writer http.ResponseWriter, req *http.Request) {
	request := &KoaRequest{}
	response := &KoaResponse{
		status:        http.StatusOK,
		hasSentHeader: false,
	}
	context := &Context{
		app:      this,
		Req:      req,
		Res:      writer,
		Request:  request,
		Response: response,
	}
	response.context = context
	this.dispatcher_dfs(context, 0)
	// write response
	if !response.hasSentHeader {
		response.hasSentHeader = true
		writer.WriteHeader(response.status)
	}
}

/** 递归处理插件 */
func (this *Koa) dispatcher_dfs(context *Context, index int) {
	if index < 0 || index >= len(this.plugins) {
		return
	}
	plugin := this.plugins[index]
	switch plugin.(type) {
	case PluginSingleArg:
		plugin.Call(context)
	case PluginMultiArg:
		plugin.(PluginMultiArg)(context, func() {
			this.dispatcher_dfs(context, index+1)
		})
	}
}

type KoaConfig struct {
	env             string   // 默认是 NODE_ENV 或 "development"
	keys            []string //签名的 cookie 密钥数组
	proxy           bool     // 当真正的代理头字段将被信任时
	subdomains      []string // 忽略
	subdomainOffset int      // 偏移量，默认为 2
	proxyIpHeader   string   // 代理 ip 消息头, 默认为 X-Forwarded-For
	maxIpsCount     int      // 从代理 ip 消息头读取的最大 ips, 默认为 0 (代表无限)
}

func NewKoaConfig() *KoaConfig {
	return &KoaConfig{}
}

type KoaRequest struct {
}

type KoaResponse struct {
	context       *Context
	hasSentHeader bool
	status        int
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
