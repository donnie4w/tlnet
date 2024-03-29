// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
package tlnet

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"golang.org/x/net/websocket"
)

type TTYPE int

const (
	_ TTYPE = iota
	JSON
	BINARY
	COMPACT
)

type stub struct {
	_pattern   string
	_dir       string
	_filter    *Filter
	_handler   func(http.ResponseWriter, *http.Request)
	_processor thrift.TProcessor
	_ttype     TTYPE
}

func newStub(pattern, dir string, filter *Filter, handler func(http.ResponseWriter, *http.Request), processor thrift.TProcessor, ttype TTYPE) *stub {
	if pattern[0] != '/' {
		panic(fmt.Sprint("pattern error,[", pattern, "] must begin with '/'"))
	}
	return &stub{pattern, dir, filter, handler, processor, ttype}
}

type wsStub struct {
	_pattern string
	_handler *wsHandler
}

func newWsStub(pattern string, handler *wsHandler) *wsStub {
	if pattern[0] != '/' {
		panic(fmt.Sprint("pattern error,[", pattern, "] must begin with '/'"))
	}
	return &wsStub{pattern, handler}
}

func NewTlnet() *Tlnet {
	t := new(Tlnet)
	t._processors = make([]*stub, 0)
	t._staticHandlers = make([]*stub, 0)
	t._handlers = make([]*stub, 0)
	t.Server = new(http.Server)
	t._serverMux = http.NewServeMux()
	t.Server.Handler = t._serverMux
	t._wss = make([]*wsStub, 0)
	t._methodpattern = make(map[string]string, 0)
	return t
}

type Tlnet struct {
	_maxBytes       int64
	_processors     []*stub
	_handlers       []*stub
	_staticHandlers []*stub
	Server          *http.Server
	_serverMux      *http.ServeMux
	_wss            []*wsStub
	_methodpattern  map[string]string
}

// TLSConfig optionally provides a TLS configuration for use
// by ServeTLS and ListenAndServeTLS. Note that t value is
// cloned by ServeTLS and ListenAndServeTLS, so it's not
// possible to modify the configuration with methods like
// tls.Config.SetSessionTicketKeys. To use
// SetSessionTicketKeys, use Server.Serve with a TLS Listener
// instead.
func (t *Tlnet) TLSConfig(_TLSConfig *tls.Config) {
	t.Server.TLSConfig = _TLSConfig
}

// ReadTimeout is the maximum duration for reading the entire
// request, including the body. A zero or negative value means
// there will be no timeout.
//
// Because ReadTimeout does not let Handlers make per-request
// decisions on each request body's acceptable deadline or
// upload rate, most users will prefer to use
// ReadHeaderTimeout. It is valid to use them both.
func (t *Tlnet) ReadTimeout(_ReadTimeout time.Duration) {
	t.Server.ReadTimeout = _ReadTimeout
}

// ReadHeaderTimeout is the amount of time allowed to read
// request headers. The connection's read deadline is reset
// after reading the headers and the Handler can decide what
// is considered too slow for the body. If ReadHeaderTimeout
// is zero, the value of ReadTimeout is used. If both are
// zero, there is no timeout.
func (t *Tlnet) ReadHeaderTimeout(_ReadHeaderTimeout time.Duration) {
	t.Server.ReadHeaderTimeout = _ReadHeaderTimeout
}

// WriteTimeout is the maximum duration before timing out
// writes of the response. It is reset whenever a new
// request's header is read. Like ReadTimeout, it does not
// let Handlers make decisions on a per-request basis.
// A zero or negative value means there will be no timeout.
func (t *Tlnet) WriteTimeout(_WriteTimeout time.Duration) {
	t.Server.WriteTimeout = _WriteTimeout
}

// IdleTimeout is the maximum amount of time to wait for the
// next request when keep-alives are enabled. If IdleTimeout
// is zero, the value of ReadTimeout is used. If both are
// zero, there is no timeout.
func (t *Tlnet) IdleTimeout(_IdleTimeout time.Duration) {
	t.Server.IdleTimeout = _IdleTimeout
}

// MaxHeaderBytes controls the maximum number of bytes the
// server will read parsing the request header's keys and
// values, including the request line. It does not limit the
// size of the request body.
// If zero, DefaultMaxHeaderBytes is used.
func (t *Tlnet) MaxHeaderBytes(_MaxHeaderBytes int) {
	t.Server.MaxHeaderBytes = _MaxHeaderBytes
}

// 数据库文件路径
// func (t *tlnet) DBPath(dbPath string) {
// 	t._dbPath = dbPath
// }

// 设置请求body限制
func (t *Tlnet) SetMaxBytesReader(maxBytes int64) {
	t._maxBytes = maxBytes
}

// 处理thrift协议请求
func (t *Tlnet) AddProcessor(pattern string, processor thrift.TProcessor) {
	t._addProcessor(pattern, processor, JSON)
}

func (t *Tlnet) AddBinaryProcessor(pattern string, processor thrift.TProcessor) {
	t._addProcessor(pattern, processor, BINARY)
}

func (t *Tlnet) AddCompactProcessor(pattern string, processor thrift.TProcessor) {
	t._addProcessor(pattern, processor, COMPACT)
}

func (t *Tlnet) _addProcessor(pattern string, processor thrift.TProcessor, ttype TTYPE) {
	// t._processors = append(t._processors, &stub{_pattern: pattern, _processor: processor, _ttype: ttype})
	t._processors = append(t._processors, newStub(pattern, "", nil, nil, processor, ttype))
}

// 处理动态请求
func (t *Tlnet) AddHandlerFunc(pattern string, f *Filter, handlerFunc func(http.ResponseWriter, *http.Request)) {
	// t._handlers = append(t._handlers, &stub{_pattern: pattern, _filter: f, _handler: handlerFunc})
	t._handlers = append(t._handlers, newStub(pattern, "", f, handlerFunc, nil, 0))
}

// 处理静态页面
func (t *Tlnet) AddStaticHandler(pattern string, dir string, f *Filter, handlerFunc func(http.ResponseWriter, *http.Request)) {
	t._staticHandlers = append(t._staticHandlers, newStub(pattern, dir, f, handlerFunc, nil, 0))
}

/**http**/
func (t *Tlnet) HttpStart(addr string) (err error) {
	t._Handle()
	t.Server.Addr = addr
	err = t.Server.ListenAndServe()
	if err != nil {
		logger.Error("tlnet start error:", err.Error())
	}
	return
}

/**http tls**/
func (t *Tlnet) HttpStartTLS(addr string, certFile, keyFile string) (err error) {
	t._Handle()
	t.Server.Addr = addr
	err = t.Server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		logger.Error("tlnet startTLS error:", err.Error())
	}
	return
}

/**http tls**/
func (t *Tlnet) HttpStartTlsBytes(addr string, certBys, keyBys []byte) (err error) {
	t._Handle()
	t.Server.Addr = addr
	cfg := &tls.Config{}
	var cert tls.Certificate
	if cert, err = tls.X509KeyPair(certBys, keyBys); err == nil {
		cfg.NextProtos = append(cfg.NextProtos, "http/1.1")
		cfg.Certificates = append(cfg.Certificates, cert)
		t.TLSConfig(cfg)
		var ln net.Listener
		if ln, err = net.Listen("tcp", addr); err == nil {
			defer ln.Close()
			tlsListener := tls.NewListener(ln, cfg)
			err = t.Server.Serve(tlsListener)
		}
	}
	return
}

/**close service**/
func (t *Tlnet) Close() (err error) {
	if t.Server != nil {
		return t.Server.Close()
	}
	return
}

func (t *Tlnet) _Handle() {
	for _, s := range t._processors {
		t._serverMux.Handle(s._pattern, http.StripPrefix(s._pattern, &httpHandler{_maxBytes: t._maxBytes, _stub: s, _tlnet: t}))
	}
	for _, s := range t._handlers {
		t._serverMux.Handle(s._pattern, http.StripPrefix(s._pattern, &httpHandler{_maxBytes: t._maxBytes, _stub: s, _tlnet: t}))
	}
	for _, s := range t._staticHandlers {
		t._serverMux.Handle(s._pattern, http.StripPrefix(s._pattern, &httpHandler{_maxBytes: t._maxBytes, _stub: s, _staticHandler: http.FileServer(http.Dir(s._dir)), _tlnet: t}))
	}
	if len(t._wss) > 0 {
		for _, s := range t._wss {
			if s._handler != nil {
				t._serverMux.Handle(s._pattern, s._handler)
			}
		}
	}
}

type httpHandler struct {
	_maxBytes      int64
	_stub          *stub
	_staticHandler http.Handler
	_tlnet         *Tlnet
}

func (t *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	path, uri, url_uri := r.URL.Path, r.RequestURI, r.URL.RequestURI()
	if t._maxBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, t._maxBytes)
	}
	if len(t._tlnet._methodpattern) > 0 {
		if url_uri[0] == '/' {
			url_uri = url_uri[1:]
		}
		method, ok := "", false
		if a, b := len(uri), len(url_uri); a != b {
			root := r.RequestURI[:(a - b)]
			method, ok = t._tlnet._methodpattern[root]
		}
		if ok && !strings.EqualFold(method, r.Method) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}

	// static
	if t._staticHandler != nil && t._stub._filter != nil && t._stub._filter.notFoundhandler != nil {
		dir := t._stub._dir
		if dir != "" && dir[len(dir)-1:] != "/" {
			dir = dir + "/"
		}
		if _, err := os.Stat(dir + path); os.IsNotExist(err) {
			if t._stub._filter.notFoundhandler != nil && t._stub._filter.notFoundhandler(w, r) {
				return
			}
		}
	}
	var filterPath string
	if i := strings.LastIndex(uri, "?"); i > 0 {
		filterPath = uri[:i]
	} else {
		filterPath = uri
	}
	if t._stub._filter != nil {
		if t._stub._filter.suffixMap.Len() > 0 {
			if t._stub._filter._processSuffix(filterPath, w, r) {
				return
			}
		}
		if t._stub._filter.matchMap.Len() > 0 {
			if t._stub._filter._processGlobal(filterPath, w, r) {
				return
			}
		}
	}
	if t._stub._processor != nil {
		processorHandler(w, r, t._stub._processor, t._stub._ttype)
		return
	}
	if t._stub._handler != nil {
		t._stub._handler(w, r)
	}

	if t._staticHandler != nil {
		t._staticHandler.ServeHTTP(w, r)
	}
}

type wsHandler struct {
	httpContextFunc  func(hc *HttpContext)
	_Origin          string
	_OriginFunc      func(origin *url.URL) bool
	_MaxPayloadBytes int
	_OnError         func(self *Websocket)
	_OnOpen          func(hc *HttpContext)
}

func (t *wsHandler) checkOrigin(config *websocket.Config, req *http.Request) (err error) {
	config.Origin, err = websocket.Origin(config, req)
	if err == nil && config.Origin == nil {
		return fmt.Errorf("null origin")
	}
	if (t._Origin != "" && t._Origin != config.Origin.String()) || (t._OriginFunc != nil && !t._OriginFunc(config.Origin)) {
		return fmt.Errorf("error origin")
	}
	return err
}

// ServeHTTP implements the http.Handler interface for a WebSocket
func (t *wsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s := websocket.Server{Handler: t.wsConnFunc, Handshake: t.checkOrigin}
	s.ServeHTTP(w, req)
}

func (t *wsHandler) wsConnFunc(ws *websocket.Conn) {
	defer myRecover()
	defer ws.Close()
	hc := newHttpContext(nil, ws.Request())
	ws.MaxPayloadBytes, hc.WS._OnError = t._MaxPayloadBytes, t._OnError
	hc.WS.Conn = ws
	if t._OnOpen != nil {
		go t._OnOpen(hc)
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

func processorHandler(w http.ResponseWriter, r *http.Request, processor thrift.TProcessor, _ttype TTYPE) {
	var protocolFactory thrift.TProtocolFactory
	switch _ttype {
	case JSON:
		protocolFactory = thrift.NewTJSONProtocolFactory()
	case BINARY:
		protocolFactory = thrift.NewTBinaryProtocolFactoryConf(&thrift.TConfiguration{})
	case COMPACT:
		protocolFactory = thrift.NewTCompactProtocolFactoryConf(&thrift.TConfiguration{})
	}
	transport := thrift.NewStreamTransport(r.Body, w)
	ioProtocol := protocolFactory.GetProtocol(transport)
	hc := newHttpContext(w, r)
	_, err := processor.Process(context.WithValue(context.Background(), "HttpContext", hc), ioProtocol, ioProtocol)
	if err != nil {
		logger.Error("processorHandler Error:", err)
	}
}
