// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
package tlnet

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	. "net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	. "github.com/donnie4w/tlnet/db"
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

func NewTlnet() *tlnet {
	t := new(tlnet)
	t._processors = make([]*stub, 0)
	t._staticHandlers = make([]*stub, 0)
	t._handlers = make([]*stub, 0)
	t._server = new(http.Server)
	t._serverMux = http.NewServeMux()
	t._server.Handler = t._serverMux
	t._wss = make([]*wsStub, 0)
	t._methodpattern = make(map[string]string, 0)
	return t
}

type tlnet struct {
	_dbPath         string
	_maxBytes       int64
	_processors     []*stub
	_handlers       []*stub
	_staticHandlers []*stub
	_server         *http.Server
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
func (this *tlnet) TLSConfig(_TLSConfig *tls.Config) {
	this._server.TLSConfig = _TLSConfig
}

// ReadTimeout is the maximum duration for reading the entire
// request, including the body. A zero or negative value means
// there will be no timeout.
//
// Because ReadTimeout does not let Handlers make per-request
// decisions on each request body's acceptable deadline or
// upload rate, most users will prefer to use
// ReadHeaderTimeout. It is valid to use them both.
func (this *tlnet) ReadTimeout(_ReadTimeout time.Duration) {
	this._server.ReadTimeout = _ReadTimeout
}

// ReadHeaderTimeout is the amount of time allowed to read
// request headers. The connection's read deadline is reset
// after reading the headers and the Handler can decide what
// is considered too slow for the body. If ReadHeaderTimeout
// is zero, the value of ReadTimeout is used. If both are
// zero, there is no timeout.
func (this *tlnet) ReadHeaderTimeout(_ReadHeaderTimeout time.Duration) {
	this._server.ReadHeaderTimeout = _ReadHeaderTimeout
}

// WriteTimeout is the maximum duration before timing out
// writes of the response. It is reset whenever a new
// request's header is read. Like ReadTimeout, it does not
// let Handlers make decisions on a per-request basis.
// A zero or negative value means there will be no timeout.
func (this *tlnet) WriteTimeout(_WriteTimeout time.Duration) {
	this._server.WriteTimeout = _WriteTimeout
}

// IdleTimeout is the maximum amount of time to wait for the
// next request when keep-alives are enabled. If IdleTimeout
// is zero, the value of ReadTimeout is used. If both are
// zero, there is no timeout.
func (this *tlnet) IdleTimeout(_IdleTimeout time.Duration) {
	this._server.IdleTimeout = _IdleTimeout
}

// MaxHeaderBytes controls the maximum number of bytes the
// server will read parsing the request header's keys and
// values, including the request line. It does not limit the
// size of the request body.
// If zero, DefaultMaxHeaderBytes is used.
func (this *tlnet) MaxHeaderBytes(_MaxHeaderBytes int) {
	this._server.MaxHeaderBytes = _MaxHeaderBytes
}

// ?????????????????????
func (this *tlnet) DBPath(dbPath string) {
	this._dbPath = dbPath
}

// ????????????body??????
func (this *tlnet) SetMaxBytesReader(maxBytes int64) {
	this._maxBytes = maxBytes
}

// ??????thrift????????????
func (this *tlnet) AddProcessor(pattern string, processor thrift.TProcessor) {
	this._addProcessor(pattern, processor, JSON)
}

func (this *tlnet) AddBinaryProcessor(pattern string, processor thrift.TProcessor) {
	this._addProcessor(pattern, processor, BINARY)
}

func (this *tlnet) AddCompactProcessor(pattern string, processor thrift.TProcessor) {
	this._addProcessor(pattern, processor, COMPACT)
}

func (this *tlnet) _addProcessor(pattern string, processor thrift.TProcessor, ttype TTYPE) {
	// this._processors = append(this._processors, &stub{_pattern: pattern, _processor: processor, _ttype: ttype})
	this._processors = append(this._processors, newStub(pattern, "", nil, nil, processor, ttype))
}

// ??????????????????
func (this *tlnet) AddHandlerFunc(pattern string, f *Filter, handlerFunc func(ResponseWriter, *Request)) {
	// this._handlers = append(this._handlers, &stub{_pattern: pattern, _filter: f, _handler: handlerFunc})
	this._handlers = append(this._handlers, newStub(pattern, "", f, handlerFunc, nil, 0))
}

// ??????????????????
func (this *tlnet) AddStaticHandler(pattern string, dir string, f *Filter, handlerFunc func(ResponseWriter, *Request)) {
	this._staticHandlers = append(this._staticHandlers, newStub(pattern, dir, f, handlerFunc, nil, 0))
}

/**http**/
func (this *tlnet) HttpStart(addr string) (err error) {
	this._Handle()
	this._server.Addr = addr
	err = this._server.ListenAndServe()
	if err != nil {
		logger.Error("tlnet start error:", err.Error())
	}
	return
}

/**http tls**/
func (this *tlnet) HttpStartTLS(addr string, certFile, keyFile string) (err error) {
	this._Handle()
	this._server.Addr = addr
	err = this._server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		logger.Error("tlnet startTLS error:", err.Error())
	}
	return
}

func (this *tlnet) _Handle() {
	if this._dbPath != "" {
		UseSimpleDB(this._dbPath)
	}
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
	_tlnet         *tlnet
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
	path := r.URL.Path
	if this._maxBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, this._maxBytes)
	}
	if len(this._tlnet._methodpattern) > 0 {
		uri := r.RequestURI
		url_uri := r.URL.RequestURI()
		if url_uri[0] == '/' {
			url_uri = url_uri[1:]
		}
		var root string
		var method string
		var ok bool
		if a, b := len(uri), len(url_uri); a != b {
			root = r.RequestURI[:(a - b)]
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
		if dir[len(dir)-1:] != "/" {
			dir = fmt.Sprint(dir, "/")
		}
		if _, err := os.Stat(dir + path); os.IsNotExist(err) {
			if this._stub._filter.notFoundhandler != nil && this._stub._filter.notFoundhandler(w, r) {
				return
			}
		}
	}
	if this._stub._filter != nil {
		if len(this._stub._filter.suffixMap) > 0 {
			if this._stub._filter._processSuffix(path, w, r) {
				return
			}
		}
		if len(this._stub._filter.matchMap) > 0 {
			if this._stub._filter._processGlobal(path, w, r) {
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
