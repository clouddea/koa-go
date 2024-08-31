package plugin

import (
	"encoding/json"
	"github.com/clouddea/koa-go/koa"
	"github.com/clouddea/koa-go/koa/util"
	"io"
	"net/http"
	"strconv"
)

/** JSONParser */
func NewJsonParser(maxLenBytes int) koa.PluginMultiArg {
	return func(context *koa.Context, next func()) {
		if context.Req.Header.Get("Content-Type") == "application/json" {
			length, err := strconv.Atoi(context.Req.Header.Get("Content-Length"))
			if err != nil {
				context.Throw(http.StatusBadRequest)
				return
			}
			if length > maxLenBytes {
				context.Throw(http.StatusNotAcceptable)
				return
			}
			switch context.Request.Body.(type) {
			case io.ReadCloser:
				data := make([]byte, length)
				buf := make([]byte, 1024)
				size := 0
				for size <= maxLenBytes {
					readBytes, err := context.Request.Body.(io.ReadCloser).Read(buf)
					if err != nil && err != io.EOF {
						context.Throw(http.StatusBadRequest)
						return
					}
					if readBytes == 0 {
						break
					}
					copy(data[size:size+readBytes], buf[:readBytes])
				}
				var body map[string]any = make(map[string]any)
				err := json.Unmarshal(data, &body)
				if err != nil {
					context.Throw(http.StatusBadRequest)
					return
				}
				context.Request.Body = body
			}
		}
		next()
		if context.Response.Body != nil {
			if _, ok := context.Response.Body.([]byte); !ok {
				bytes, err := json.Marshal(context.Response.Body)
				util.Assert(err, "can not parse obj to json")
				context.Response.Header().Add("Content-Length", strconv.Itoa(len(bytes)))
				context.Response.Body = bytes
			}
		}
	}
}
