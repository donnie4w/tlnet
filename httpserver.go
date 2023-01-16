package tlnet

import (
	"context"
	"fmt"
	"os"
	"strings"

	"net/http"
	. "net/http"

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
	return t
}

type tlnet struct {
	_dbPath         string
	_maxBytes       int64
	_processors     []*stub
	_handlers       []*stub
	_staticHandlers []*stub
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

//设置端口
func (this *tlnet) HttpStart(port int32) {
	if this._dbPath != "" {
		InitDB(this._dbPath)
	}
	for _, s := range this._processors {
		logging.Debug("------------->>>>>_processors")
		http.Handle(s._pattern, http.StripPrefix(s._pattern, &httpHandler{_maxBytes: this._maxBytes, _stub: s}))
	}
	for _, s := range this._handlers {
		logging.Debug("------------->>>>>_handlers")
		http.Handle(s._pattern, http.StripPrefix(s._pattern, &httpHandler{_maxBytes: this._maxBytes, _stub: s}))
	}
	for _, s := range this._staticHandlers {
		logging.Debug("------------->>>>>_staticHandlers")
		http.Handle(s._pattern, http.StripPrefix(s._pattern, &httpHandler{_maxBytes: this._maxBytes, _stub: s, _h: http.FileServer(http.Dir(s._dir))}))
	}
	e := http.ListenAndServe(fmt.Sprint(":", port), nil)
	if e != nil {
		logging.Error("tlnet start error:", e.Error())
	}
}

type httpHandler struct {
	_maxBytes int64
	_stub     *stub
	_h        http.Handler
}

func (this *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	path := r.URL.Path
	logging.Debug("path：", path)
	if this._maxBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, this._maxBytes)
	}

	//只作用于静态页面上
	if this._h != nil && this._stub._filter != nil && this._stub._filter.notFoundhandler != nil {
		logging.Debug("----------------------->notFoundhandler")
		dir := this._stub._dir
		if !strings.HasSuffix(dir, "/") {
			dir = fmt.Sprint(dir, "/")
		}
		if _, err := os.Stat(dir + path); os.IsNotExist(err) {
			if this._stub._filter.notFoundhandler != nil && this._stub._filter.notFoundhandler(w, r) {
				logging.Debug("----------------------->2")
				return
			}
		}
	}

	if this._stub._filter != nil {
		if len(this._stub._filter.suffixMap) > 0 {
			logging.Debug("----------------------->0")
			if this._stub._filter._processSuffix(path, w, r) {
				return
			}
		}
		if len(this._stub._filter.matchMap) > 0 {
			logging.Debug("----------------------->1")
			if this._stub._filter._processGlobal(path, w, r) {
				return
			}
		}
	}

	if this._stub._processor != nil {
		logging.Debug("----------------------->3")
		processorHandler(w, r, this._stub._processor)
		return
	}
	if this._stub._handler != nil {
		logging.Debug("----------------------->4")
		this._stub._handler(w, r)
	}
	if this._h != nil {
		logging.Debug("----------------------->5")
		this._h.ServeHTTP(w, r)
	}
}

func processorHandler(w ResponseWriter, r *Request, processor thrift.TProcessor) {
	if "POST" == r.Method {
		protocolFactory := thrift.NewTJSONProtocolFactory()
		transport := thrift.NewStreamTransport(r.Body, w)
		ioProtocol := protocolFactory.GetProtocol(transport)
		htpo := &httpObj{w, r}
		s, err := processor.Process(context.WithValue(context.Background(), "httpObj", htpo), ioProtocol, ioProtocol)
		if !s {
			logging.Error("err：", err)
		}
	}
}
