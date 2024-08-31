package plugin

import (
	"github.com/clouddea/koa-go/koa"
	"github.com/clouddea/koa-go/koa/util"
	"gopkg.in/yaml.v3"
	"os"
)

/** 读取yaml配置文件 */
func NewConfig[T any](filename string, config T) (koa.PluginMultiArg, *T) {
	bytes, err := os.ReadFile(filename)
	util.Assert(err, "open config file error")
	err = yaml.Unmarshal(bytes, &config)
	util.Assert(err, "read yaml file error")
	return func(ctx *koa.Context, next func()) {
		ctx.State["config"] = config
		next()
	}, &config
}
