package koa

import (
	eventbus "github.com/asaskevich/EventBus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	KOA_EVENT_START = "start" // 启动完毕
	KOA_EVENT_ERROR = "error" // 发生错误
	KOA_EVENT_CLOSE = "close" // 应用关闭
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
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		this.Emit(KOA_EVENT_CLOSE)
		os.Exit(0)
	}()
	http.HandleFunc("/", this.dispatcher)
	endpoint := "0.0.0.0:" + strconv.Itoa(port)
	this.Emit(KOA_EVENT_START)
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
	cookies := &koaCookies{}
	request := &KoaRequest{
		Body: req.Body,
	}
	response := &KoaResponse{
		status:        http.StatusOK,
		hasSentHeader: false,
	}
	context := &Context{
		Cookies:  cookies,
		State:    make(map[string]any),
		app:      this,
		Req:      req,
		Res:      writer,
		Request:  request,
		Response: response,
	}
	// global error processing
	defer func() {
		if err := recover(); err != nil {
			context.Throw(http.StatusInternalServerError)
			log.Println("Internal Sever Error")
			log.Printf("%v\n", err)
		}
	}()
	// emit request
	cookies.context = context
	request.context = context
	response.context = context
	this.dispatcher_dfs(context, 0)
	// write response
	if !response.hasSentHeader {
		response.hasSentHeader = true
		writer.WriteHeader(response.status)
		// write body
		if response.Body != nil {
			if body, ok := response.Body.([]byte); ok {
				_ = response.Write(body)
			} else {

			}
		}
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

type koaCookies struct {
	context *Context
}

func (this *koaCookies) Set(cookie *http.Cookie) {
	this.context.Req.AddCookie(cookie)
}

func (this *koaCookies) Get(name string) (*http.Cookie, error) {
	return this.context.Req.Cookie(name)
}
