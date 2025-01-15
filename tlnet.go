// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import (
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"golang.org/x/net/websocket"
)

type stub struct {
	pattern    string
	dir        string
	filter     *Filter
	handler    func(http.ResponseWriter, *http.Request)
	tProcessor thrift.TProcessor
	tftype     tf_pco_type
	httpMethod httpMethod
}

func newStub(pattern, dir string, filter *Filter, handler func(http.ResponseWriter, *http.Request), processor thrift.TProcessor, ttype tf_pco_type, method httpMethod) *stub {
	return &stub{pattern, dir, filter, handler, processor, ttype, method}
}

type wsStub struct {
	_pattern string
	_handler *wsHandler
}

func newWsStub(pattern string, handler *wsHandler) *wsStub {
	return &wsStub{pattern, handler}
}

// NewTlnet creates and returns a new instance of Tlnet.
//
// This function initializes all necessary fields of the Tlnet structure and sets default values,
// preparing it to handle HTTP requests and WebSocket connections.
// The returned Tlnet instance can be used to configure routes, handle requests, and manage connections.
func NewTlnet() *Tlnet {
	t := new(Tlnet)
	t.processors = make([]*stub, 0)
	t.staticHandlers = make([]*stub, 0)
	t.handlers = make([]*stub, 0)
	t.Server = new(http.Server)
	t.wss = make([]*wsStub, 0)
	t.UseTlnetMux()
	return t
}

func (t *Tlnet) UseTlnetMux() *Tlnet {
	t.handlemux = NewTlnetMux()
	t.Server.Handler = t.handlemux
	t.mod = tlnetmod
	return t
}

func (t *Tlnet) UseTlnetMuxWithLimit(limit int) *Tlnet {
	t.handlemux = NewTlnetMuxWithLimit(limit)
	t.Server.Handler = t.handlemux
	t.mod = tlnetmod
	return t
}

// UseServeMux use net/http ServeMux as an HTTP request multiplexer
// https://pkg.go.dev/net/http#ServeMux
func (t *Tlnet) UseServeMux() *Tlnet {
	t.handlemux = http.NewServeMux()
	t.Server.Handler = t.handlemux
	t.mod = nativemod
	return t
}

type TlnetMux interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
	Handle(string, http.Handler)
	HandleFunc(string, func(http.ResponseWriter, *http.Request))
}

type Tlnet struct {
	maxBytes       int64
	processors     []*stub
	handlers       []*stub
	staticHandlers []*stub
	Server         *http.Server
	handlemux      TlnetMux
	wss            []*wsStub
	mod            muxmod
}

// TLSConfig optionally provides a TLS configuration for use
// by ServeTLS and ListenAndServeTLS. Note that t value is
// cloned by ServeTLS and ListenAndServeTLS, so it's not
// possible to modify the configuration with methods like
// tls.Config.SetSessionTicketKeys. To use
// SetSessionTicketKeys, use Server.Serve with a TLS Listener
// instead.
func (t *Tlnet) TLSConfig(tlsConfig *tls.Config) {
	t.Server.TLSConfig = tlsConfig
}

// ReadTimeout is the maximum duration for reading the entire
// request, including the body. A zero or negative value means
// there will be no timeout.
//
// Because ReadTimeout does not let Handlers make per-request
// decisions on each request body's acceptable deadline or
// upload rate, most users will prefer to use
// ReadHeaderTimeout. It is valid to use them both.
func (t *Tlnet) ReadTimeout(readTimeout time.Duration) {
	logger.Debug("[ReadTimeout]:", readTimeout)
	t.Server.ReadTimeout = readTimeout
}

// ReadHeaderTimeout is the amount of time allowed to read
// request headers. The connection's read deadline is reset
// after reading the headers and the Handler can decide what
// is considered too slow for the body. If ReadHeaderTimeout
// is zero, the value of ReadTimeout is used. If both are
// zero, there is no timeout.
func (t *Tlnet) ReadHeaderTimeout(readHeaderTimeout time.Duration) {
	logger.Debug("[ReadHeaderTimeout]:", readHeaderTimeout)
	t.Server.ReadHeaderTimeout = readHeaderTimeout
}

// WriteTimeout is the maximum duration before timing out
// writes of the response. It is reset whenever a new
// request's header is read. Like ReadTimeout, it does not
// let Handlers make decisions on a per-request basis.
// A zero or negative value means there will be no timeout.
func (t *Tlnet) WriteTimeout(writeTimeout time.Duration) {
	logger.Debug("[WriteTimeout]:", writeTimeout)
	t.Server.WriteTimeout = writeTimeout
}

// IdleTimeout is the maximum amount of time to wait for the
// next request when keep-alives are enabled. If IdleTimeout
// is zero, the value of ReadTimeout is used. If both are
// zero, there is no timeout.
func (t *Tlnet) IdleTimeout(idleTimeout time.Duration) {
	logger.Debug("[IdleTimeout]:", idleTimeout)
	t.Server.IdleTimeout = idleTimeout
}

// MaxHeaderBytes controls the maximum number of bytes the
// server will read parsing the request header's keys and
// values, including the request line. It does not limit the
// size of the request body.
// If zero, DefaultMaxHeaderBytes is used.
func (t *Tlnet) MaxHeaderBytes(maxHeaderBytes int) {
	logger.Debug("[MaxHeaderBytes]:", maxHeaderBytes)
	t.Server.MaxHeaderBytes = maxHeaderBytes
}

// SetMaxBytesReader
// Sets the request body size limit
func (t *Tlnet) SetMaxBytesReader(maxBytes int64) {
	logger.Debug("[SetMaxBytesReader]:", maxBytes)
	t.maxBytes = maxBytes
}

// HttpStart
//
// addr optionally specifies the TCP address for the server to listen on,
// in the form "host:port". If empty, ":http" (port 80) is used.
// The service names are defined in RFC 6335 and assigned by IANA.
// See net.Dial for details of the address format.
//
// e.g.
//
// 1. HttpStart(":8080")
//
// 2. HttpStart("127.0.0.1:8080")
func (t *Tlnet) HttpStart(addr string) (err error) {
	t.handle()
	t.Server.Addr = addr
	if err = t.Server.ListenAndServe(); err != nil {
		logger.Error("[Tlnet start failed]", err.Error())
	}
	return
}

// HttpsStart
//
// addr optionally specifies the TCP address for the server to listen on,
// in the form "host:port". If empty, ":http" (port 80) is used.
// The service names are defined in RFC 6335 and assigned by IANA.
// See net.Dial for details of the address format.
//
// e.g.
//
// 1. HttpsStart(":8080","server_crt","server_key")
//
// 2. HttpsStart("127.0.0.1:8080","server_crt","server_key")
func (t *Tlnet) HttpsStart(addr string, server_crt, server_key string) (err error) {
	t.handle()
	t.Server.Addr = addr
	if err = t.Server.ListenAndServeTLS(server_crt, server_key); err != nil {
		logger.Error("[Tlnet start TLS failed]", err.Error())

	}
	return
}

// HttpsStartWithBytes
//
// addr optionally specifies the TCP address for the server to listen on,
// in the form "host:port". If empty, ":http" (port 80) is used.
// The service names are defined in RFC 6335 and assigned by IANA.
// See net.Dial for details of the address format.
//
// e.g.
//
// 1. HttpsStartWithBytes(":8080",[]byte(crtBytes),[]byte(keyBytes))
//
// 2. HttpsStartWithBytes("127.0.0.1:8080",[]byte(crtBytes),[]byte(keyBytes))
func (t *Tlnet) HttpsStartWithBytes(addr string, crtBytes, keyBytes []byte) (err error) {
	t.handle()
	t.Server.Addr = addr
	cfg := &tls.Config{}
	var cert tls.Certificate
	if cert, err = tls.X509KeyPair(crtBytes, keyBytes); err == nil {
		cfg.NextProtos = append(cfg.NextProtos, "http/1.1")
		cfg.Certificates = append(cfg.Certificates, cert)
		t.TLSConfig(cfg)
		var ln net.Listener
		if ln, err = net.Listen("tcp", addr); err == nil {
			defer ln.Close()
			tlsListener := tls.NewListener(ln, cfg)
			err = t.Server.Serve(tlsListener)
		}
		if err != nil {
			logger.Error("[Tlnet start TLS by Bytes failed]", err.Error())
		}
	}
	return
}

// Close
// close tlnet service
func (t *Tlnet) Close() (err error) {
	if t.Server != nil {
		return t.Server.Close()
	}
	return
}

func (t *Tlnet) handle() {
	m := make(map[string]byte, 0)
	for _, s := range t.processors {
		t.handlemux.Handle(checkpattern(m, s.pattern), &httpHandler{maxBytes: t.maxBytes, stub: s})
	}
	for _, s := range t.handlers {
		t.handlemux.Handle(checkpattern(m, s.pattern), &httpHandler{maxBytes: t.maxBytes, stub: s})
	}
	for _, s := range t.staticHandlers {
		pattern := checkpattern(m, s.pattern)
		t.handlemux.Handle(pattern, http.StripPrefix(pattern, &httpHandler{maxBytes: t.maxBytes, stub: s, staticHandler: http.FileServer(http.Dir(s.dir))}))
	}
	if len(t.wss) > 0 {
		for _, s := range t.wss {
			if s._handler != nil {
				t.handlemux.Handle(checkpattern(m, s._pattern), s._handler)
			}
		}
	}
}

func checkpattern(m map[string]byte, pattern string) string {
	pattern = strings.ReplaceAll(pattern, " ", "")
	if pattern == "" {
		panic("pattern must not be empty")
	}
	if !strings.HasPrefix(pattern, "/") {
		panic("pattern is in an incorrect format and must start with a `/` : " + pattern)
	}
	if _, ok := m[pattern]; ok {
		panic("pattern cannot be defined repeatedly : " + pattern)
	}
	m[pattern] = 0
	return pattern
}

type httpHandler struct {
	maxBytes      int64
	stub          *stub
	staticHandler http.Handler
}

func (t *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	path := r.URL.Path
	if t.maxBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, t.maxBytes)
	}
	if t.stub != nil {
		if t.stub.httpMethod != defaultMethod {
			if !strings.EqualFold(string(t.stub.httpMethod), r.Method) {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
		}
		if t.staticHandler != nil && t.stub.filter != nil && t.stub.filter.notFoundhandler != nil {
			if _, err := os.Stat(filepath.Join(t.stub.dir, path)); os.IsNotExist(err) {
				if t.stub.filter.notFoundhandler != nil && t.stub.filter.notFoundhandler(w, r) {
					return
				}
			}
		}
		if t.stub.filter != nil {
			if t.stub.filter.suffixMap.Len() > 0 {
				if t.stub.filter.processSuffix(path, w, r) {
					return
				}
			}
			if t.stub.filter.matchMap.Len() > 0 {
				if t.stub.filter.processGlobal(path, w, r) {
					return
				}
			}
		}
		if t.stub.tProcessor != nil {
			processorHandler(w, r, t.stub.tProcessor, t.stub.tftype)
			return
		}
		if t.stub.handler != nil {
			t.stub.handler(w, r)
		}
	}

	if t.staticHandler != nil {
		t.staticHandler.ServeHTTP(w, r)
	}
}

type wsHandler struct {
	httpContextFunc func(hc *HttpContext)
	origin          string
	originFunc      func(origin *url.URL) bool
	maxPayloadBytes int
	onError         func(self *Websocket)
	onOpen          func(hc *HttpContext)
}

func (t *wsHandler) checkOrigin(config *websocket.Config, req *http.Request) (err error) {
	if config.Origin, err = websocket.Origin(config, req); err == nil && config.Origin == nil {
		return fmt.Errorf("null origin")
	}
	if (t.origin != "" && t.origin != config.Origin.String()) || (t.originFunc != nil && !t.originFunc(config.Origin)) {
		return fmt.Errorf("error origin")
	}
	return
}

// ServeHTTP implements the http.Handler interface for a WebSocket
func (t *wsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s := websocket.Server{Handler: t.handle, Handshake: t.checkOrigin}
	s.ServeHTTP(w, req)
}

func (t *wsHandler) handle(ws *websocket.Conn) {
	defer tlnetRecover(nil)
	defer ws.Close()
	hc := newHttpContextWithWebsocket(nil, ws.Request())
	ws.MaxPayloadBytes, hc.WS._OnError = t.maxPayloadBytes, t.onError
	hc.WS.Conn = ws
	if t.onOpen != nil {
		go t.onOpen(hc)
	}
	defer hc.WS._onErrorChan()
	for hc.WS.Error == nil {
		var byt []byte
		if err := websocket.Message.Receive(ws, &byt); err != nil {
			hc.WS.Error = err
			break
		}
		hc.WS._rbody = byt
		if t.httpContextFunc != nil {
			t.httpContextFunc(hc)
		}
	}
}

func ParseFormFile(file multipart.File, fileHeader *multipart.FileHeader, savePath, namePrefix string) (fileName string, err error) {
	defer file.Close()
	filepath := path.Join(savePath, namePrefix+fileHeader.Filename)
	os.MkdirAll(path.Dir(filepath), 0777)
	f, er := os.Create(filepath)
	err = er
	if err == nil {
		fileName = fileHeader.Filename
		defer f.Close()
		_, err = io.Copy(f, file)
	}
	return
}
