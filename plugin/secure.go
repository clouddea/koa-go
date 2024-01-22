package plugin

import "github.com/clouddea/koa-go/koa"

func NewSecure(CSRF bool, CORS bool, XSS bool) koa.PluginMultiArg {
	// TODO: implements
	// CSRF 不同源网站前端通过URL访问你的后端，会利用你已经认证后cookie状态做一些事 （浏览器是允许的）。解决办法，使用token
	// CORS 不同源网站前端访问你的后端，会因为CORS已开启，所以可以做CSRF
	// XSS 即用户的输入中包含了可以在前端执行的js代码 （类似SQL注入，国为XSS在前端执行，所以可以盗取用户cookie、token等）
	return nil
}
