package plugin

import (
	"github.com/clouddea/koa-go/koa"
	"github.com/clouddea/koa-go/koa/util"
	"io"
	"log"
	"maps"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func NewProxy(mappings map[string][]string) koa.PluginMultiArg {

	trie := util.NewTrie()
	// 因为是只读的，所以不用加锁
	// ""  =>  [""]
	// "/" => ["", ""]

	for k, _ := range mappings {
		word := strings.Split(strings.TrimSpace(k), "/")
		trie.Insert(word, k) // 记录下PATH的值
	}
	return func(ctx *koa.Context, next func()) {
		path := ctx.Req.URL.Path
		word := strings.Split(path, "/")
		value, _ := trie.Search(word)
		if value == nil {
			next()
			return
		}
		prefix := value.(string)
		upstreams := mappings[prefix]
		upstream := upstreams[rand.Intn(len(upstreams))]
		extraPath, ok := strings.CutPrefix(path, prefix)
		if !ok {
			ctx.Response.SetStatus(http.StatusInternalServerError)
			return
		}
		// 发送请求
		var client = &http.Client{
			Timeout: time.Second * 60,
		}
		upstream = strings.ReplaceAll(upstream, "$path", path)
		upstream = strings.ReplaceAll(upstream, "$extra", extraPath)
		if ctx.Req.URL.RawQuery != "" {
			upstream += "?" + ctx.Req.URL.RawQuery
		}

		rqst, err := http.NewRequest(ctx.Req.Method, upstream, ctx.Req.Body)
		if err != nil {
			log.Println("New request failed:", err)
			ctx.Response.SetStatus(http.StatusInternalServerError)
			return
		}
		rqst.Header = ctx.Req.Header

		rsps, err := client.Do(rqst)
		if err != nil {
			log.Println("Request failed:", err)
			ctx.Response.SetStatus(http.StatusBadGateway)
			return
		}
		// 设置响应状态
		ctx.Response.SetStatus(rsps.StatusCode)
		// 设置响应头
		maps.Copy(ctx.Res.Header(), rsps.Header)
		// 循环读取文件内容
		defer rsps.Body.Close()
		buffer := make([]byte, 1024)
		for {
			// 从文件中读取内容到缓冲区
			bytesRead, err := rsps.Body.Read(buffer)
			if bytesRead > 0 {
				// 输出读取的内容到标准输出
				err = ctx.Response.Write(buffer[:bytesRead])
				if err != nil {
					log.Println("Write response failed:", err)
					return
				}
			}
			if err != nil {
				if err == io.EOF {
					break // 文件已经读取完毕
				}
				log.Println("Read response failed:", err)
				return
			}
		}
		return
	}
}
