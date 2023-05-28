// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
package tlnet

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	. "net/http"
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
	_handler   func(ResponseWriter, *Request)
	_processor thrift.TProcessor
	_ttype     TTYPE
}

func newStub(pattern, dir string, filter *Filter, handler func(ResponseWriter, *Request), processor thrift.TProcessor, ttype TTYPE) *stub {
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
// by ServeTLS and ListenAndServeTLS. Note that this value is
// cloned by ServeTLS and ListenAndServeTLS, so it's not
// possible to modify the configuration with methods like
// tls.Config.SetSessionTicketKeys. To use
// SetSessionTicketKeys, use Server.Serve with a TLS Listener
// instead.
func (this *Tlnet) TLSConfig(_TLSConfig *tls.Config) {
	this.Server.TLSConfig = _TLSConfig
}

// ReadTimeout is the maximum duration for reading the entire
// request, including the body. A zero or negative value means
// there will be no timeout.
//
// Because ReadTimeout does not let Handlers make per-request
// decisions on each request body's acceptable deadline or
// upload rate, most users will prefer to use
// ReadHeaderTimeout. It is valid to use them both.
func (this *Tlnet) ReadTimeout(_ReadTimeout time.Duration) {
	this.Server.ReadTimeout = _ReadTimeout
}

// ReadHeaderTimeout is the amount of time allowed to read
// request headers. The connection's read deadline is reset
// after reading the headers and the Handler can decide what
// is considered too slow for the body. If ReadHeaderTimeout
// is zero, the value of ReadTimeout is used. If both are
// zero, there is no timeout.
func (this *Tlnet) ReadHeaderTimeout(_ReadHeaderTimeout time.Duration) {
	this.Server.ReadHeaderTimeout = _ReadHeaderTimeout
}

// WriteTimeout is the maximum duration before timing out
// writes of the response. It is reset whenever a new
// request's header is read. Like ReadTimeout, it does not
// let Handlers make decisions on a per-request basis.
// A zero or negative value means there will be no timeout.
func (this *Tlnet) WriteTimeout(_WriteTimeout time.Duration) {
	this.Server.WriteTimeout = _WriteTimeout
}

// IdleTimeout is the maximum amount of time to wait for the
// next request when keep-alives are enabled. If IdleTimeout
// is zero, the value of ReadTimeout is used. If both are
// zero, there is no timeout.
func (this *Tlnet) IdleTimeout(_IdleTimeout time.Duration) {
	this.Server.IdleTimeout = _IdleTimeout
}

// MaxHeaderBytes controls the maximum number of bytes the
// server will read parsing the request header's keys and
// values, including the request line. It does not limit the
// size of the request body.
// If zero, DefaultMaxHeaderBytes is used.
func (this *Tlnet) MaxHeaderBytes(_MaxHeaderBytes int) {
	this.Server.MaxHeaderBytes = _MaxHeaderBytes
}

// 数据库文件路径
// func (this *tlnet) DBPath(dbPath string) {
// 	this._dbPath = dbPath
// }

// 设置请求body限制
func (this *Tlnet) SetMaxBytesReader(maxBytes int64) {
	this._maxBytes = maxBytes
}

// 处理thrift协议请求
func (this *Tlnet) AddProcessor(pattern string, processor thrift.TProcessor) {
	this._addProcessor(pattern, processor, JSON)
}

func (this *Tlnet) AddBinaryProcessor(pattern string, processor thrift.TProcessor) {
	this._addProcessor(pattern, processor, BINARY)
}

func (this *Tlnet) AddCompactProcessor(pattern string, processor thrift.TProcessor) {
	this._addProcessor(pattern, processor, COMPACT)
}

func (this *Tlnet) _addProcessor(pattern string, processor thrift.TProcessor, ttype TTYPE) {
	// this._processors = append(this._processors, &stub{_pattern: pattern, _processor: processor, _ttype: ttype})
	this._processors = append(this._processors, newStub(pattern, "", nil, nil, processor, ttype))
}

// 处理动态请求
func (this *Tlnet) AddHandlerFunc(pattern string, f *Filter, handlerFunc func(ResponseWriter, *Request)) {
	// this._handlers = append(this._handlers, &stub{_pattern: pattern, _filter: f, _handler: handlerFunc})
	this._handlers = append(this._handlers, newStub(pattern, "", f, handlerFunc, nil, 0))
}

// 处理静态页面
func (this *Tlnet) AddStaticHandler(pattern string, dir string, f *Filter, handlerFunc func(ResponseWriter, *Request)) {
	this._staticHandlers = append(this._staticHandlers, newStub(pattern, dir, f, handlerFunc, nil, 0))
}

/**http**/
func (this *Tlnet) HttpStart(addr string) (err error) {
	this._Handle()
	this.Server.Addr = addr
	err = this.Server.ListenAndServe()
	if err != nil {
		logger.Error("tlnet start error:", err.Error())
	}
	return
}

/**http tls**/
func (this *Tlnet) HttpStartTLS(addr string, certFile, keyFile string) (err error) {
	this._Handle()
	this.Server.Addr = addr
	err = this.Server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		logger.Error("tlnet startTLS error:", err.Error())
	}
	return
}

/**http tls**/
func (this *Tlnet) HttpStartTlsBytes(addr string, certBys, keyBys []byte) (err error) {
	this._Handle()
	this.Server.Addr = addr
	cfg := &tls.Config{}
	var cert tls.Certificate
	if cert, err = tls.X509KeyPair(certBys, keyBys); err == nil {
		cfg.NextProtos = append(cfg.NextProtos, "http/1.1")
		cfg.Certificates = append(cfg.Certificates, cert)
		this.TLSConfig(cfg)
		var ln net.Listener
		if ln, err = net.Listen("tcp", addr); err == nil {
			defer ln.Close()
			tlsListener := tls.NewListener(ln, cfg)
			err = this.Server.Serve(tlsListener)
		}
	}
	return
}

/**close service**/
func (this *Tlnet) Close() (err error) {
	if this.Server != nil {
		err = this.Server.Close()
		if err != nil {
			logger.Error("tlnet start error:", err.Error())
		}
	}
	return
}

func (this *Tlnet) _Handle() {
	for _, s := range this._processors {
		this._serverMux.Handle(s._pattern, http.StripPrefix(s._pattern, &httpHandler{_maxBytes: this._maxBytes, _stub: s, _tlnet: this}))
	}
	for _, s := range this._handlers {
		this._serverMux.Handle(s._pattern, http.StripPrefix(s._pattern, &httpHandler{_maxBytes: this._maxBytes, _stub: s, _tlnet: this}))
	}
	for _, s := range this._staticHandlers {
		this._serverMux.Handle(s._pattern, http.StripPrefix(s._pattern, &httpHandler{_maxBytes: this._maxBytes, _stub: s, _staticHandler: http.FileServer(http.Dir(s._dir)), _tlnet: this}))
	}
	if len(this._wss) > 0 {
		for _, s := range this._wss {
			this._serverMux.Handle(s._pattern, s._handler)
		}
	}
}

type httpHandler struct {
	_maxBytes      int64
	_stub          *stub
	_staticHandler http.Handler
	_tlnet         *Tlnet
}

// func _checkmethod(path string, m map[string]string) (methed string, ok bool) {
// 	path = path[:_getLastLetterIndex(path, "/")]
// 	if methed, ok = m[path]; len(path) > 0 && !ok {
// 		return _checkmethod(path[:len(path)-1], m)
// 	}
// 	return
// }

// func _getLastLetterIndex(s, _i string) (i int) {
// 	for i = len(s); i > 0; i-- {
// 		if s[i-1:i] == _i {
// 			return
// 		}
// 	}
// 	return
// }

func (this *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	path, uri, url_uri := r.URL.Path, r.RequestURI, r.URL.RequestURI()
	if this._maxBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, this._maxBytes)
	}
	if len(this._tlnet._methodpattern) > 0 {
		if url_uri[0] == '/' {
			url_uri = url_uri[1:]
		}
		var method string
		var ok bool
		if a, b := len(uri), len(url_uri); a != b {
			root := r.RequestURI[:(a - b)]
			method, ok = this._tlnet._methodpattern[root]
		}

		if ok && method != strings.ToUpper(r.Method) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}

	// static
	if this._staticHandler != nil && this._stub._filter != nil && this._stub._filter.notFoundhandler != nil {
		dir := this._stub._dir
		if dir != "" && dir[len(dir)-1:] != "/" {
			dir = fmt.Sprint(dir, "/")
		}
		if _, err := os.Stat(dir + path); os.IsNotExist(err) {
			if this._stub._filter.notFoundhandler != nil && this._stub._filter.notFoundhandler(w, r) {
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
	if this._stub._filter != nil {
		if this._stub._filter.suffixMap.Len() > 0 {
			if this._stub._filter._processSuffix(filterPath, w, r) {
				return
			}
		}
		if this._stub._filter.matchMap.Len() > 0 {
			if this._stub._filter._processGlobal(filterPath, w, r) {
				return
			}
		}
	}
	if this._stub._processor != nil {
		processorHandler(w, r, this._stub._processor, this._stub._ttype)
		return
	}
	if this._stub._handler != nil {
		this._stub._handler(w, r)
	}

	if this._staticHandler != nil {
		this._staticHandler.ServeHTTP(w, r)
	}
}

type wsHandler struct {
	httpContextFunc  func(hc *HttpContext)
	_Origin          string
	_OriginFunc      func(origin *url.URL) bool
	_MaxPayloadBytes int
	_OnError         func(self *Websocket)
}

func (this *wsHandler) checkOrigin(config *websocket.Config, req *http.Request) (err error) {
	config.Origin, err = websocket.Origin(config, req)
	if err == nil && config.Origin == nil {
		return fmt.Errorf("null origin")
	}
	if (this._Origin != "" && this._Origin != config.Origin.String()) || (this._OriginFunc != nil && !this._OriginFunc(config.Origin)) {
		return fmt.Errorf("error origin")
	}
	return err
}

// ServeHTTP implements the http.Handler interface for a WebSocket
func (this *wsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s := websocket.Server{Handler: this.wsConnFunc, Handshake: this.checkOrigin}
	s.ServeHTTP(w, req)
}

func (this *wsHandler) wsConnFunc(ws *websocket.Conn) {
	defer ws.Close()
	hc := newHttpContext(nil, ws.Request())
	ws.MaxPayloadBytes, hc.WS._OnError = this._MaxPayloadBytes, this._OnError
	hc.WS.Conn = ws
	for hc.WS.IsError == nil {
		var byt []byte
		if err := websocket.Message.Receive(ws, &byt); err != nil {
			hc.WS.IsError = err
			break
		}
		hc.WS._rbody = byt
		this.httpContextFunc(hc)
	}
	hc.WS._onErrorChan()
}

func processorHandler(w ResponseWriter, r *Request, processor thrift.TProcessor, _ttype TTYPE) {
	var protocolFactory thrift.TProtocolFactory
	switch _ttype {
	case JSON:
		protocolFactory = thrift.NewTJSONProtocolFactory()
	case BINARY:
		protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	case COMPACT:
		protocolFactory = thrift.NewTCompactProtocolFactory()
	}
	transport := thrift.NewStreamTransport(r.Body, w)
	ioProtocol := protocolFactory.GetProtocol(transport)
	hc := newHttpContext(w, r)
	_, err := processor.Process(context.WithValue(context.Background(), "HttpContext", hc), ioProtocol, ioProtocol)
	if err != nil {
		logger.Error("processorHandler Error:", err)
	}
}
