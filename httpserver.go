package tlnet

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	. "net/http"
	"os"
	"strings"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/donnie4w/simplelog/logging"
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

type wsStub struct {
	_pattern string
	_handler *wsHandler
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

//数据库文件路径
func (this *tlnet) DBPath(dbPath string) {
	this._dbPath = dbPath
}

//设置请求body限制
func (this *tlnet) SetMaxBytesReader(maxBytes int64) {
	this._maxBytes = maxBytes
}

//处理thrift协议请求
func (this *tlnet) AddProcessor(pattern string, processor thrift.TProcessor) {
	this._processors = append(this._processors, &stub{_pattern: pattern, _processor: processor, _ttype: JSON})
}

func (this *tlnet) AddBinaryProcessor(pattern string, processor thrift.TProcessor) {
	this._processors = append(this._processors, &stub{_pattern: pattern, _processor: processor, _ttype: BINARY})
}

func (this *tlnet) AddCompactProcessor(pattern string, processor thrift.TProcessor) {
	this._processors = append(this._processors, &stub{_pattern: pattern, _processor: processor, _ttype: COMPACT})
}

//处理动态请求
func (this *tlnet) AddHandlerFunc(pattern string, f *Filter, handlerFunc func(ResponseWriter, *Request)) {
	this._handlers = append(this._handlers, &stub{_pattern: pattern, _filter: f, _handler: handlerFunc})
}

//处理静态页面
func (this *tlnet) AddStaticHandler(pattern string, dir string, f *Filter, handlerFunc func(ResponseWriter, *Request)) {
	if !strings.HasSuffix(pattern, "/") {
		pattern = fmt.Sprint(pattern, "/")
	}
	this._staticHandlers = append(this._staticHandlers, &stub{_pattern: pattern, _dir: dir, _filter: f, _handler: handlerFunc})
}

/**http**/
func (this *tlnet) HttpStart(port int32) (err error) {
	this._Handle()
	this._server.Addr = fmt.Sprint(":", port)
	err = this._server.ListenAndServe()
	if err != nil {
		logging.Error("tlnet start error:", err.Error())
	}
	return
}

/**http tls**/
func (this *tlnet) HttpStartTLS(port int32, certFile, keyFile string) (err error) {
	this._Handle()
	this._server.Addr = fmt.Sprint(":", port)
	err = this._server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		logging.Error("tlnet startTLS error:", err.Error())
	}
	return
}

func (this *tlnet) _Handle() {
	if this._dbPath != "" {
		UseSimpleDB(this._dbPath)
	}
	for _, s := range this._processors {
		this._serverMux.Handle(s._pattern, http.StripPrefix(s._pattern, &httpHandler{_maxBytes: this._maxBytes, _stub: s}))
	}
	for _, s := range this._handlers {
		this._serverMux.Handle(s._pattern, http.StripPrefix(s._pattern, &httpHandler{_maxBytes: this._maxBytes, _stub: s}))
	}
	for _, s := range this._staticHandlers {
		this._serverMux.Handle(s._pattern, http.StripPrefix(s._pattern, &httpHandler{_maxBytes: this._maxBytes, _stub: s, _staticHandler: http.FileServer(http.Dir(s._dir))}))
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
}

func (this *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	path := r.URL.Path
	if this._maxBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, this._maxBytes)
	}
	//作用于静态页面
	if this._staticHandler != nil && this._stub._filter != nil && this._stub._filter.notFoundhandler != nil {
		dir := this._stub._dir
		if !strings.HasSuffix(dir, "/") {
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
	httpContextFunc func(hc *HttpContext)
}

func checkOrigin(config *websocket.Config, req *http.Request) (err error) {
	config.Origin, err = websocket.Origin(config, req)
	if err == nil && config.Origin == nil {
		return fmt.Errorf("null origin")
	}
	return err
}

// ServeHTTP implements the http.Handler interface for a WebSocket
func (this wsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s := websocket.Server{Handler: this.wsConnFunc, Handshake: checkOrigin}
	s.ServeHTTP(w, req)
}

func (this wsHandler) wsConnFunc(ws *websocket.Conn) {
	hc := newHttpContext(nil, ws.Request())
	hc.WS.ws = ws
	for hc.WS.onError == nil {
		var byt []byte
		if err := websocket.Message.Receive(ws, &byt); err != nil {
			hc.WS.onError = err
			break
		}
		hc.WS.rbody = byt
		this.httpContextFunc(hc)
	}
}

func processorHandler(w ResponseWriter, r *Request, processor thrift.TProcessor, _ttype TTYPE) {
	if "POST" == strings.ToUpper(r.Method) {
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
			logging.Error("processorHandler Error:", err)
		}
	}
}
