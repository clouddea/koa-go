package fastcgi

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/clouddea/koa-go/koa"
	"github.com/clouddea/koa-go/koa/util"
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
	requestReceiveBuffers := make(map[uint16]chan any)
	requestReceiveBuffersLock := sync.RWMutex{}
	startedReceiveResponse := false
	startedReceiveResponseLock := sync.Mutex{}

	// receive response unifiedly
	//go receiveResponse(conn, requestReceiveBuffers)

	return func(context *koa.Context, next func()) {
		defer func() {
			if err := recover(); err != nil {
				panic(err)
			}
		}()

		cuttedPath, ok := strings.CutPrefix(context.Req.URL.Path, prefix)
		if !ok {
			panic("invalid path for prefix: " + prefix)
		}

		requestIdLock.Lock()
		requestId += 1
		requestIdLock.Unlock()
		requestIdB1 := uint8(requestId >> 8 & 0xFF)
		requestIdB0 := uint8(requestId >> 0 & 0xFF)

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

		envs := map[string]string{
			"SCRIPT_FILENAME":   documentRoot + cuttedPath, // have no influence to php
			"REQUEST_METHOD":    "GET",
			"PATH_INFO":         context.Req.URL.Path, // will automately be replaced by extra path after php file
			"QUERY_STRING":      context.Req.URL.RawQuery,
			"DOCUMENT_ROOT":     documentRoot,
			"DOCUMENT_URI":      cuttedPath, // will add extra path after php file automately
			"REQUEST_URI":       cuttedPath, // will add extra path after php file automately
			"SCRIPT_NAME":       cuttedPath,
			"SERVER_PROTOCOL":   "HTTP/1.1",
			"GATEWAY_INTERFACE": "CGI/1.1",
			"CONTENT_TYPE":      contentType,
			"CONTENT_LENGTH":    fmt.Sprintf("%d", context.Req.ContentLength),
			"SERVER_NAME":       "localhost",
			"SERVER_PORT":       "8080",
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
			_, err = conn.Write(buf.Bytes())
			util.Assert(err, "can not write  param item request")
			_, err = conn.Write(util.AnyToBytes(envItem.NameLength))
			util.Assert(err, "can not write env item name length")
			_, err = conn.Write(util.AnyToBytes(envItem.ValueLength))
			util.Assert(err, "can not write env item value length")
			_, err = conn.Write(envItem.NameData)
			util.Assert(err, "can not write env item name data")
			_, err = conn.Write(envItem.ValueData)
			util.Assert(err, "can not write env item value data")

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
		err = binary.Write(buf, binary.BigEndian, paramEnd)
		if err != nil {
			log.Panicln("can not encode request", err)
		}
		_, err = conn.Write(buf.Bytes())
		if err != nil {
			log.Panicln("can not write request", err)
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
		err = binary.Write(buf, binary.BigEndian, inputEnd)
		if err != nil {
			log.Panicln("can not encode request", err)
		}
		_, err = conn.Write(buf.Bytes())
		if err != nil {
			log.Panicln("can not write request", err)
		}

		// receive own response
		responseBuffer := make(chan any, 1)
		requestReceiveBuffersLock.Lock()
		requestReceiveBuffers[requestId] = responseBuffer
		requestReceiveBuffersLock.Unlock()
		defer func() {
			requestReceiveBuffersLock.Lock()
			delete(requestReceiveBuffers, requestId)
			requestReceiveBuffersLock.Unlock()
		}()

		startedReceiveResponseLock.Lock()
		if !startedReceiveResponse {
			// receive response unifiedly
			go receiveResponse(conn, requestReceiveBuffers)
			startedReceiveResponse = true
		}
		startedReceiveResponseLock.Unlock()

		var headerLine = make([]byte, 0)
		var readHeaderFinish = false
		var charNL = 0
		var charCR = 0
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
								context.Response.Header().Set(
									strings.TrimSpace(headerLineKeyValue[0]),
									strings.TrimSpace(headerLineKeyValue[1]),
								)
							} else if b == '\n' {
								charNL += 1
							} else {
								charCR = 0
								charNL = 0
								headerLine = append(headerLine, b)
							}
							if charCR == 2 && charNL == 2 {
								readHeaderFinish = true
							}
						}

					}
				}
				if stderr, ok := result.(FCGIResponseStderr); ok {
					_, err = os.Stderr.Write(stderr.data)
					util.Assert(err, "unnable to write os stderr")
				}
				if _, ok := result.(FCGIResponseEndRequest); ok {
					log.Print("end of fascgi request: ",
						requestId)
					break loopReceiveResponseObject
				}
				if _, ok := result.(FCGIResponseUnknown); ok {
					continue
				}
			case <-time.After(30 * time.Second):
				panic("read response from fastcgi timeout: requestid: " + fmt.Sprintf("%v", requestId))
			}
		}
	}
}

func receiveResponse(conn net.Conn, bufs map[uint16]chan any) {

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
		contentBytes := make([]byte, contentLength)

		// read fcgi content
		if contentLength != 0 {
			_, err = conn.Read(contentBytes)
			util.Assert(err, "can not read response content")
			buf = &bytes.Buffer{}
			buf.Write(contentBytes)
		}
		if resp.Type == FCGI_STDOUT {
			bufs[requestId] <- FCGIResponseStdout{resp, contentBytes}
		} else if resp.Type == FCGI_STDERR {
			bufs[requestId] <- FCGIResponseStderr{resp, contentBytes}
		} else if resp.Type == FCGI_UNKNOWN_TYPE {
			unkonwBody := FCGI_UnknownTypeBody{}
			util.Assert(binary.Read(buf, binary.BigEndian, &unkonwBody), "can not parse unknown type body")
			bufs[requestId] <- FCGIResponseUnknown{resp, unkonwBody}
		} else if resp.Type == FCGI_END_REQUEST {
			endRequest := FCGI_EndRequestBody{}
			util.Assert(binary.Read(buf, binary.BigEndian, &endRequest), "can not parse end response")
			bufs[requestId] <- FCGIResponseEndRequest{resp, endRequest}
		}
		// read padding
		if paddingLength != 0 {
			paddingBytes := make([]byte, paddingLength)
			_, err := conn.Read(paddingBytes)
			util.Assert(err, "can not read padding bytes")
		}
	}
}
