package fastcgi

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/clouddea/koa-go/koa"
	"github.com/clouddea/koa-go/koa/util"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// cgi: https://datatracker.ietf.org/doc/html/rfc3875
// fastcgi: https://fastcgi-archives.github.io/FastCGI_Specification.html

func NewFastCGI(prefix string, documentRoot string, host string, port int) koa.PluginMultiArg {
	// run a tcp socket go routine
	// TODO: 连接断开重试
	conn, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal("err : ", err)
	}
	// do not need to close connection
	// defer conn.Close()

	// states
	requestId := uint16(1)
	requestIdLock := sync.Mutex{}
	requestReceiveBuffers := sync.Map{}

	startedReceiveResponse := false
	startedReceiveResponseLock := sync.Mutex{}

	// receive response unifiedly
	//go receiveResponse(conn, requestReceiveBuffers)

	return func(context *koa.Context, next func()) {

		cuttedPath, ok := strings.CutPrefix(context.Req.URL.Path, prefix)
		if !ok {
			panic("invalid path for prefix: " + prefix)
		}

		if strings.Contains(cuttedPath, ".") && !strings.HasSuffix(cuttedPath, ".php") {
			// 读取文件
			next()
			return
		}

		defer func() {
			if err := recover(); err != nil {
				panic(err)
			}
		}()

		requestIdLock.Lock()
		requestId += 1
		requestIdB1 := uint8(requestId >> 8 & 0xFF)
		requestIdB0 := uint8(requestId >> 0 & 0xFF)
		requestIdCurr := requestId
		requestIdLock.Unlock()

		// 在请求之前先建立channel
		responseBuffer := make(chan any, 1)
		requestReceiveBuffers.Store(requestIdCurr, responseBuffer)
		fmt.Printf("requestId: %v add channel\n", requestIdCurr)
		defer func() {
			requestReceiveBuffers.Delete(requestIdCurr)
			fmt.Printf("requestId: %v remove channel\n", requestIdCurr)
		}()

		// 开始发起请求

		request := FCGI_BeginRequestRecord{
			Header: FCGI_Header{
				Version:         FCGI_VERSION_1,
				Type:            FCGI_BEGIN_REQUEST,
				RequestIdB1:     requestIdB1,
				RequestIdB0:     requestIdB0,
				ContentLengthB1: 0,
				ContentLengthB0: 8,
				PaddingLength:   0,
				Reserved:        0,
			},
			Body: FCGI_BeginRequestBody{
				RoleB1:   0,
				RoleB0:   FCGI_RESPONDER,
				Flags:    FCGI_KEEP_CONN,
				Reserved: [5]uint8{},
			},
		}

		buf := &bytes.Buffer{}
		util.Assert(binary.Write(buf, binary.BigEndian, request), "can not encode request")
		_, err = conn.Write(buf.Bytes())
		util.Assert(err, "can not write request")

		//envs
		contentType := context.Req.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "text/html; charset=utf-8"
		}

		scriptFileName := cuttedPath
		if !strings.HasSuffix(scriptFileName, ".php") {
			scriptFileName += "/index.php"
		}

		envs := map[string]string{
			"SCRIPT_FILENAME":   documentRoot + scriptFileName, // have no influence to php
			"REQUEST_METHOD":    context.Req.Method,
			"PATH_INFO":         context.Req.URL.Path, // will automately be replaced by extra path after php file
			"QUERY_STRING":      context.Req.URL.RawQuery,
			"DOCUMENT_ROOT":     documentRoot,
			"DOCUMENT_URI":      cuttedPath, // will add extra path after php file automately
			"REQUEST_URI":       cuttedPath, // will add extra path after php file automately
			"SCRIPT_NAME":       cuttedPath,
			"SERVER_PROTOCOL":   context.Req.Proto, // "HTTP/1.1"
			"GATEWAY_INTERFACE": "CGI/1.1",
			"CONTENT_TYPE":      contentType,
			"CONTENT_LENGTH":    fmt.Sprintf("%d", context.Req.ContentLength),
			"SERVER_NAME":       "localhost",
			"SERVER_PORT":       "8080",
			"HTTP_HOST":         context.Req.Host, // for wordpress, dont kow how to set, from header: Host
		}

		// add request header
		// 暂时只支持1个同KEY
		for k, vs := range context.Req.Header {
			envKey := "HTTP_" + strings.ToUpper(strings.ReplaceAll(k, "-", "_"))
			envValue := vs[0]
			envs[envKey] = envValue
			//fmt.Printf("%v=%v\n", envKey, envValue)
		}

		for envKey, envValue := range envs {
			envKeyBytes := []byte(envKey)
			envValueBytes := []byte(envValue)
			envKeyLength := uint32(len(envKeyBytes))
			envValueLength := uint32(len(envValueBytes))
			envItem := FCGI_NameValuePair44{
				envKeyLength | 0x80000000,
				envValueLength | 0x80000000,
				envKeyBytes,
				envValueBytes,
			}
			contentLength := envKeyLength + envValueLength + 8

			paramHeader := FCGI_Header{
				Version:         FCGI_VERSION_1,
				Type:            FCGI_PARAMS,
				RequestIdB1:     requestIdB1,
				RequestIdB0:     requestIdB0,
				ContentLengthB1: uint8(contentLength >> 8 & 0xff),
				ContentLengthB0: uint8(contentLength & 0xff),
				PaddingLength:   0,
				Reserved:        0,
			}
			buf = &bytes.Buffer{}
			util.Assert(binary.Write(buf, binary.BigEndian, paramHeader), "can not encode param item request")
			buf.Write(util.AnyToBytes(envItem.NameLength))
			buf.Write(util.AnyToBytes(envItem.ValueLength))
			buf.Write(envItem.NameData)
			buf.Write(envItem.ValueData)
			_, err = conn.Write(buf.Bytes())
			util.Assert(err, fmt.Sprintf("can not write param item request"))

		}

		paramEnd := FCGI_Header{
			Version:         FCGI_VERSION_1,
			Type:            FCGI_PARAMS,
			RequestIdB1:     requestIdB1,
			RequestIdB0:     requestIdB0,
			ContentLengthB1: 0,
			ContentLengthB0: 0,
			PaddingLength:   0,
			Reserved:        0,
		}
		buf = &bytes.Buffer{}
		util.Assert(binary.Write(buf, binary.BigEndian, paramEnd), "can not encode request")
		_, err = conn.Write(buf.Bytes())
		util.Assert(err, "can not write request")

		for true {
			readBuffer := make([]byte, 1024)
			readBytesCount, err := context.Req.Body.Read(readBuffer)
			if readBytesCount != 0 {
				inputHeader := FCGI_Header{
					Version:         FCGI_VERSION_1,
					Type:            FCGI_STDIN,
					RequestIdB1:     requestIdB1,
					RequestIdB0:     requestIdB0,
					ContentLengthB1: uint8(readBytesCount >> 8 & 0xff),
					ContentLengthB0: uint8(readBytesCount & 0xff),
					PaddingLength:   0,
					Reserved:        0,
				}

				buf = &bytes.Buffer{}
				util.Assert(binary.Write(buf, binary.BigEndian, inputHeader), "can not encode request")
				_, err = conn.Write(buf.Bytes())
				util.Assert(err, "can not write request")
				_, err = conn.Write(readBuffer[:readBytesCount])
				util.Assert(err, "can not write request data")
				fmt.Printf("requestId: %v sent a request input frame. content length=%v \n", requestIdCurr, readBytesCount)
			}
			if err == io.EOF {
				break
			} else {
				util.Assert(err, "can not read request body")
			}
		}

		inputEnd := FCGI_Header{
			Version:         FCGI_VERSION_1,
			Type:            FCGI_STDIN,
			RequestIdB1:     requestIdB1,
			RequestIdB0:     requestIdB0,
			ContentLengthB1: 0,
			ContentLengthB0: 0,
			PaddingLength:   0,
			Reserved:        0,
		}

		buf = &bytes.Buffer{}
		util.Assert(binary.Write(buf, binary.BigEndian, inputEnd), "can not encode request")
		_, err = conn.Write(buf.Bytes())
		util.Assert(err, "can not write request")

		fmt.Printf("requestId: %d start: path=%v, query=%v, content length=%v \n",
			requestIdCurr, cuttedPath, context.Req.URL.RawQuery, context.Req.ContentLength)

		// receive own response
		startedReceiveResponseLock.Lock()
		if !startedReceiveResponse {
			// receive response unifiedly
			go receiveResponse(conn, &requestReceiveBuffers)
			startedReceiveResponse = true
		}
		startedReceiveResponseLock.Unlock()

		var headerLine = make([]byte, 0)
		var readHeaderFinish = false
		var charNL = 0
		var charCR = 0
		var headers = make(map[string]string)
	loopReceiveResponseObject:
		for true {
			select {
			case result := <-responseBuffer:
				if stdout, ok := result.(FCGIResponseStdout); ok {
					if readHeaderFinish {
						util.Assert(context.Response.Write(stdout.data), "can not write content to resp")
						continue
					}
					for _, b := range stdout.data {
						if readHeaderFinish {
							util.Assert(context.Response.Write([]byte{b}), "can not write content to resp")
						} else {
							if b == '\r' {
								charCR += 1
								headerLineString := string(headerLine)
								headerLine = make([]byte, 0)
								headerLineKeyValue := strings.SplitN(headerLineString, ":", 2)
								if len(headerLineKeyValue) != 2 {
									continue
								}
								headerLineKey := strings.TrimSpace(headerLineKeyValue[0])
								headerLineValue := strings.TrimSpace(headerLineKeyValue[1])
								headers[headerLineKey] = headerLineValue
								// set status
								if headerLineKey == "Status" {
									statusAndItsMessage := strings.SplitN(headerLineValue, " ", 2)
									statusCode, err := strconv.Atoi(statusAndItsMessage[0])
									util.Assert(err, "can not parse status code")
									context.Response.SetStatus(statusCode)
								}
							} else if b == '\n' {
								charNL += 1
							} else {
								charCR = 0
								charNL = 0
								headerLine = append(headerLine, b)
							}
							if charCR == 2 && charNL == 2 {
								readHeaderFinish = true
								for k, v := range headers {
									context.Response.Header().Set(k, v)
								}
							}
						}

					}
				}
				if stderr, ok := result.(FCGIResponseStderr); ok {
					_, err = os.Stderr.Write(stderr.data)
					util.Assert(err, "unnable to write os stderr")
				}
				if _, ok := result.(FCGIResponseEndRequest); ok {
					log.Printf("end of fascgi request: %v, received request id = %v",
						requestIdCurr,
						(uint16(result.(FCGIResponseEndRequest).RequestIdB1)<<8)+uint16(result.(FCGIResponseEndRequest).RequestIdB0))
					break loopReceiveResponseObject
				}
				if _, ok := result.(FCGIResponseUnknown); ok {
					continue
				}
			case <-time.After(300 * time.Second):
				panic("read response from fastcgi timeout: requestid: " + fmt.Sprintf("%v", requestIdCurr))
			}
		}
	}
}

func receiveResponse(conn net.Conn, bufs *sync.Map) {

	// receive response
	var buf = &bytes.Buffer{}
	for true {
		//read fci header
		bufBytes := make([]byte, 8)
		_, err := conn.Read(bufBytes)
		util.Assert(err, "can not read response")
		buf = &bytes.Buffer{}
		buf.Write(bufBytes)
		resp := FCGI_Header{}
		util.Assert(binary.Read(buf, binary.BigEndian, &resp), "can not parse response")

		requestId := uint16(resp.RequestIdB1)*256 + uint16(resp.RequestIdB0)
		contentLength := uint16(resp.ContentLengthB1)*256 + uint16(resp.ContentLengthB0)
		paddingLength := resp.PaddingLength

		fmt.Printf("requestId: %d get a response frame: type=%v,  content length=%v, padding=%v \n",
			requestId, resp.Type, contentLength, paddingLength)

		// read fcgi content
		if contentLength != 0 {
			remainsLength := contentLength
			buf = &bytes.Buffer{}
			for remainsLength > 0 {
				contentReadBuf := make([]byte, min(1024, remainsLength))
				contentReadBytesLength, err := conn.Read(contentReadBuf)
				util.Assert(err, "can not read response content")
				buf.Write(contentReadBuf[:contentReadBytesLength])
				remainsLength = remainsLength - uint16(contentReadBytesLength)
			}
			//fmt.Printf("requestId: %d frame read length: %v\n",
			//	requestId, contentReadBytesLength)
		}
		responseChannel, ok := bufs.Load(requestId)
		if !ok {
			panic(fmt.Sprintf("can not find response channel for request: %v", requestId))
		}
		var responseChannelTyped = responseChannel.(chan any)
		if resp.Type == FCGI_STDOUT {
			responseChannelTyped <- FCGIResponseStdout{resp, buf.Bytes()}
		} else if resp.Type == FCGI_STDERR {
			responseChannelTyped <- FCGIResponseStderr{resp, buf.Bytes()}
		} else if resp.Type == FCGI_UNKNOWN_TYPE {
			unkonwBody := FCGI_UnknownTypeBody{}
			util.Assert(binary.Read(buf, binary.BigEndian, &unkonwBody), "can not parse unknown type body")
			responseChannelTyped <- FCGIResponseUnknown{resp, unkonwBody}
		} else if resp.Type == FCGI_END_REQUEST {
			endRequest := FCGI_EndRequestBody{}
			fmt.Printf("get end of id : %v \n", (requestId))
			util.Assert(binary.Read(buf, binary.BigEndian, &endRequest), "can not parse end response")
			responseChannelTyped <- FCGIResponseEndRequest{resp, endRequest}
		}
		// read padding
		if paddingLength != 0 {
			paddingBytes := make([]byte, paddingLength)
			_, err := conn.Read(paddingBytes)
			util.Assert(err, "can not read padding bytes")
		}
	}
}
