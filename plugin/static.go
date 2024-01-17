package plugin

import (
	"fmt"
	"github.com/clouddea/koa-go/koa"
	"github.com/gabriel-vasile/mimetype"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
)

var inlineTypes = map[string]bool{
	"text/plain":             true,
	"text/html":              true,
	"text/css":               true,
	"application/javascript": true,
	"image/jpeg":             true,
	"image/png":              true,
	"image/webp":             true,
	"image/gif":              true,
	"image/bmp":              true,
	"image/svg+xml":          true,
}

var homePageName = "index.html"

// prefix : path prefix
// dir: directory in host
func NewStatic(prefix string, dir string) koa.PluginSingleArg {
	return func(ctx *koa.Context) {
		// 判断路径是否正确
		relativeName, found := strings.CutPrefix(ctx.Req.URL.Path, prefix)
		if !found {
			ctx.Response.SetStatus(http.StatusBadRequest)
			return
		}
		if relativeName == "" || relativeName == "/" {
			relativeName = homePageName
		}
		filepath := path.Join(dir, relativeName)
		// 打开文件
		stat, err := os.Stat(filepath)
		if err != nil || stat.IsDir() {
			ctx.Response.SetStatus(http.StatusNotFound)
			return
		}
		file, err := os.Open(filepath)
		koa.Assert(err, "open file error"+filepath)
		defer file.Close()
		// 文件简单名
		_, filename := path.Split(filepath)

		// 获取 Range 头字段
		rangeHeader := ctx.Req.Header.Get("Range")
		supportRange := true
		if rangeHeader == "" {
			supportRange = false
		}

		// 解析 Range 头字段
		var start int64 = 0
		var end int64 = 0 // start 必须出现，end是可选的

		parts := strings.SplitN(rangeHeader, "=", 2)
		if len(parts) != 2 || parts[0] != "bytes" {
			supportRange = false
		} else {
			// 解析字节范围
			ranges := strings.Split(parts[1], "-") //Range: bytes=1-50   //不支持多个part，例如bytes=1-50, 100-150这种情况,因为比较复杂，也不常用
			if len(ranges) != 2 {
				supportRange = false
			} else {
				// 解析起始字节
				if ranges[0] != "" {
					start, err = strconv.ParseInt(ranges[0], 10, 64)
					if err != nil {
						supportRange = false
					}
				} else {
					supportRange = false
				}

				// 解析结束字节
				if ranges[1] != "" {
					end, err = strconv.ParseInt(ranges[1], 10, 64)
					if err != nil {
						supportRange = false
					}
				} else {
					koa.Assert(err, "stat file error")
					end = stat.Size() - 1
				}
			}
		}

		mime := readFileMIME(filepath)
		ctx.Response.Header().Add("Content-Type", mime)
		if _, ok := inlineTypes[strings.Split(mime, ";")[0]]; ok {
			ctx.Response.Header().Add("Content-Disposition", "inline")
		} else {
			ctx.Response.Header().Add("Content-Disposition", "attachment;filename="+url.QueryEscape(filename))
		}
		if !supportRange {
			koa.Assert(readFile(file, func(bytes []byte) {
				koa.Assert(ctx.Response.Write(bytes), "write to client error")
			}), "read file error")

		} else {
			if start > end || end >= stat.Size() {
				ctx.Response.SetStatus(http.StatusRequestedRangeNotSatisfiable)
				return
			}
			// 处理部分内容的逻辑
			fmt.Printf("range download file: %v-%v/%v", start, end, stat.Size())
			ctx.Response.Header().Add("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, stat.Size()))
			ctx.Response.Header().Add("Content-Length", fmt.Sprintf("%d", end-start+1))
			ctx.Response.SetStatus(http.StatusPartialContent)
			koa.Assert(readFileRange(file, start, end, func(bytes []byte) {
				koa.Assert(ctx.Response.Write(bytes), "write partial to client error")
			}), "read partial file error")
		}
	}
}

func readFile(file *os.File, callback func([]byte)) error {
	// 创建一个缓冲区，用于存储文件内容
	buffer := make([]byte, 1024)
	// 循环读取文件内容
	for {
		// 从文件中读取内容到缓冲区
		bytesRead, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break // 文件已经读取完毕
			}
			return err
		}

		// 输出读取的内容到标准输出
		callback(buffer[:bytesRead])
	}
	return nil
}

func readFileRange(file *os.File, start int64, end int64, callback func([]byte)) error {
	// 创建一个缓冲区，用于存储文件内容
	var bufSize int64 = 1024
	buffer := make([]byte, bufSize)
	_, err := file.Seek(start, 0)
	koa.Assert(err, "seek partial file error")
	// 循环读取文件内容
	for start <= end {
		// 从文件中读取内容到缓冲区
		_, err := file.Read(buffer)
		koa.Assert(err, "read partial file error")
		remains := end - start + 1
		if remains >= bufSize {
			remains = bufSize
		}
		start += remains
		// 输出读取的内容到标准输出
		callback(buffer[:remains])
	}
	return nil
}

func readFileMIME(filepath string) string {
	if strings.HasSuffix(filepath, ".css") {
		return "text/css"
	}
	mtype, err := mimetype.DetectFile(filepath)
	koa.Assert(err, "delete file mime type error")
	return mtype.String()
}
