package tlnet

import (
	"net/http"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"

	"github.com/donnie4w/simplelog/logging"
)

type Websocket struct {
	Id           int64
	_rbody       []byte
	_wbody       interface{}
	Conn         *websocket.Conn
	IsError      error
	OnError      func(self *Websocket)
	_mutex       *sync.Mutex
	_doErrorFunc bool
}

func NewWebsocket(_id int64) *Websocket {
	return &Websocket{Id: _id, _mutex: new(sync.Mutex), _doErrorFunc: false}
}

func (this *Websocket) Send(v interface{}) (err error) {
	if this.IsError == nil {
		err = websocket.Message.Send(this.Conn, v)
		this.IsError = err
		this._onErrorChan()
		return
	} else {
		return this.IsError
	}
}

func (this *Websocket) Read() []byte {
	return this._rbody
}

func (this *Websocket) Close() (err error) {
	return this.Conn.Close()
}

func (this *Websocket) _onErrorChan() {
	if this.IsError != nil && this.OnError != nil && !this._doErrorFunc {
		this._mutex.Lock()
		this._mutex.Unlock()
		if !this._doErrorFunc {
			this._doErrorFunc = true
			this.OnError(this)
		}
	}
}

type HttpInfo struct {
	Path       string
	Uri        string
	Method     string
	Host       string
	RemoteAddr string
	UserAgent  string
	Referer    string
	Header     http.Header
}

type HttpContext struct {
	w       http.ResponseWriter
	r       *http.Request
	ReqInfo *HttpInfo
	WS      *Websocket
}

var _seqId int64

func newHttpContext(w http.ResponseWriter, r *http.Request) *HttpContext {
	hi := new(HttpInfo)
	hi.Header, hi.Host, hi.Method, hi.Path, hi.RemoteAddr, hi.Uri, hi.UserAgent, hi.Referer = r.Header, r.Host, r.Method, r.URL.Path, r.RemoteAddr, r.RequestURI, r.UserAgent(), r.Referer()
	atomic.AddInt64(&_seqId, 1)
	return &HttpContext{w, r, hi, NewWebsocket(time.Now().UnixNano() + _seqId)}
}

func (this *HttpContext) GetCookie(name string) (_r string, err error) {
	cookieValue, er := this.r.Cookie(name)
	if er == nil {
		_r = cookieValue.Value
	}
	err = er
	return
}
func (this *HttpContext) SetCookie(name, value, path string, maxAge int) {
	cookie := http.Cookie{Name: name, Value: value, Path: path, MaxAge: maxAge}
	http.SetCookie(this.w, &cookie)
}

func (this *HttpContext) SetCookie2(cookie *http.Cookie) {
	http.SetCookie(this.w, cookie)
}

func (this *HttpContext) MaxBytesReader(_max int64) {
	this.r.Body = http.MaxBytesReader(this.w, this.r.Body, _max)
}

func myRecover() {
	if err := recover(); err != nil {
		logging.Error(err)
	}
}

func matchString(pattern string, s string) bool {
	b, err := regexp.MatchString(pattern, s)
	if err != nil {
		b = false
	}
	return b
}
