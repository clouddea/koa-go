package plugin

import (
	"github.com/clouddea/koa-go/koa"
	"gopkg.in/yaml.v3"
	"os"
)

/** 读取yaml配置文件 */
func NewConfig[T any](filename string, config T) koa.PluginMultiArg {
	bytes, err := os.ReadFile(filename)
	koa.Assert(err, "open config file error")
	err = yaml.Unmarshal(bytes, &config)
	koa.Assert(err, "read yaml file error")
	return func(ctx *koa.Context, next func()) {
		ctx.Attr["config"] = config
		next()
	}
}
