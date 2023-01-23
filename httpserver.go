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
)

type stub struct {
	_pattern   string
	_dir       string
	_filter    *Filter
	_handler   func(ResponseWriter, *Request)
	_processor thrift.TProcessor
}

func NewTlnet() *tlnet {
	t := new(tlnet)
	t._processors = make([]*stub, 0)
	t._staticHandlers = make([]*stub, 0)
	t._handlers = make([]*stub, 0)
	t._server = new(http.Server)
	return t
}

type tlnet struct {
	_dbPath         string
	_maxBytes       int64
	_processors     []*stub
	_handlers       []*stub
	_staticHandlers []*stub
	_server         *http.Server
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
	this._processors = append(this._processors, &stub{_pattern: pattern, _processor: processor})
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
func (this *tlnet) HttpStart(port int32) {
	this._Handle()
	this._server.Addr = fmt.Sprint(":", port)
	e := this._server.ListenAndServe()
	if e != nil {
		logging.Error("tlnet start error:", e.Error())
	}
}

/**http tls**/
func (this *tlnet) HttpStartTLS(port int32, certFile, keyFile string) {
	this._Handle()
	this._server.Addr = fmt.Sprint(":", port)
	e := this._server.ListenAndServeTLS(certFile, keyFile)
	if e != nil {
		logging.Error("tlnet startTLS error:", e.Error())
	}
}

func (this *tlnet) _Handle() {
	if this._dbPath != "" {
		InitDB(this._dbPath)
	}
	for _, s := range this._processors {
		http.Handle(s._pattern, http.StripPrefix(s._pattern, &httpHandler{_maxBytes: this._maxBytes, _stub: s}))
	}
	for _, s := range this._handlers {
		http.Handle(s._pattern, http.StripPrefix(s._pattern, &httpHandler{_maxBytes: this._maxBytes, _stub: s}))
	}
	for _, s := range this._staticHandlers {
		http.Handle(s._pattern, http.StripPrefix(s._pattern, &httpHandler{_maxBytes: this._maxBytes, _stub: s, _staticHandler: http.FileServer(http.Dir(s._dir))}))
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
	//只作用于静态页面上
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
		processorHandler(w, r, this._stub._processor)
		return
	}
	if this._stub._handler != nil {
		this._stub._handler(w, r)
	}

	if this._staticHandler != nil {
		this._staticHandler.ServeHTTP(w, r)
	}
}

func processorHandler(w ResponseWriter, r *Request, processor thrift.TProcessor) {
	if "POST" == strings.ToUpper(r.Method) {
		protocolFactory := thrift.NewTJSONProtocolFactory()
		transport := thrift.NewStreamTransport(r.Body, w)
		ioProtocol := protocolFactory.GetProtocol(transport)
		hc := newHttpContext(w, r)
		s, err := processor.Process(context.WithValue(context.Background(), "HttpContext", hc), ioProtocol, ioProtocol)
		if !s {
			logging.Error("err：", err)
		}
	}
}
