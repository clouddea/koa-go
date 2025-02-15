package fastcgi

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/clouddea/koa-go/koa"
	"github.com/clouddea/koa-go/koa/util"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
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

	requestId := uint16(250)
	requestIdLock := sync.Mutex{}

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
		err := binary.Write(buf, binary.BigEndian, request)
		if err != nil {
			log.Panicln("can not encode request", err)
		}
		_, err = conn.Write(buf.Bytes())
		if err != nil {
			log.Panicln("can not write request", err)
		}

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
			err = binary.Write(buf, binary.BigEndian, paramHeader)
			if err != nil {
				log.Panicln("can not encode param item request", err)
			}
			_, err = conn.Write(buf.Bytes())
			if err != nil {
				log.Panicln("can not write  param item request", err)
			}
			_, err = conn.Write(util.AnyToBytes(envItem.NameLength))
			if err != nil {
				log.Panicln("can not write env item name length", err)
			}
			_, err = conn.Write(util.AnyToBytes(envItem.ValueLength))
			if err != nil {
				log.Panicln("can not write env item value length", err)
			}
			_, err = conn.Write(envItem.NameData)
			if err != nil {
				log.Panicln("can not write env item name data", err)
			}
			_, err = conn.Write(envItem.ValueData)
			if err != nil {
				log.Panicln("can not write env item value data", err)
			}

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

		// receive response
		var headerLine []byte = make([]byte, 0)
		var readHeaderFinish = false
		var charNL = 0
		var charCR = 0
		for true {
			bufBytes := make([]byte, 8)
			_, err = conn.Read(bufBytes)
			if err != nil {
				log.Panicln("can not read response", err)
			}

			buf = &bytes.Buffer{}
			buf.Write(bufBytes)

			resp := FCGI_Header{}
			err = binary.Read(buf, binary.BigEndian, &resp)
			if err != nil {
				log.Panicln("can not parse response", err)
			}
			contentLength := int(resp.ContentLengthB1)*256 + int(resp.ContentLengthB0)
			paddingLength := resp.PaddingLength
			contentBytes := make([]byte, contentLength)
			if contentLength != 0 {
				// read content
				_, err = conn.Read(contentBytes)
				if err != nil {
					log.Panicln("can not read response content", err)
				}
			}

			buf = &bytes.Buffer{}
			buf.Write(bufBytes)

			if resp.Type == FCGI_STDOUT || resp.Type == FCGI_STDERR {

				if readHeaderFinish {
					err := context.Response.Write(contentBytes)
					if err != nil {
						log.Panicln("can not write content to resp ", err)
					}
				} else {
					for _, b := range contentBytes {
						if readHeaderFinish {
							err := context.Response.Write([]byte{b})
							if err != nil {
								log.Panicln("can not write content to resp ", err)
							}
						} else {
							if b == '\r' {
								charCR += 1
								headerLineString := string(headerLine)
								headerLineKeyValue := strings.SplitN(headerLineString, ":", 2)
								if len(headerLineKeyValue) != 2 {
									continue
								}
								context.Response.Header().Set(
									strings.TrimSpace(headerLineKeyValue[0]),
									strings.TrimSpace(headerLineKeyValue[1]),
								)
								headerLine = make([]byte, 0)
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
			} else if resp.Type == FCGI_UNKNOWN_TYPE {
				unkonwBody := FCGI_UnknownTypeBody{}
				err = binary.Read(buf, binary.BigEndian, &unkonwBody)
				if err != nil {
					log.Panicln("can not parse unknow body", err)
				}
			} else if resp.Type == FCGI_END_REQUEST {
				endRequest := FCGI_EndRequestBody{}
				err = binary.Read(buf, binary.BigEndian, &endRequest)
				if err != nil {
					log.Panicln("can not parse end response", err)
				}
				log.Print("end of fascgi request: ",
					(int(resp.RequestIdB1)<<8)|int(resp.RequestIdB0),
					endRequest.ProtocolStatus)
				break
			}
			// read padding
			if paddingLength != 0 {
				paddingBytes := make([]byte, paddingLength)
				_, err := conn.Read(paddingBytes)
				if err != nil {
					log.Panicln("can not read padding bytes", err)
				}
			}
		}
	}
}
