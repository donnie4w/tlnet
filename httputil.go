// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/tlnet
package tlnet

import (
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/donnie4w/gofer/util"
	"golang.org/x/net/websocket"
)

type Websocket struct {
	Id           int64
	_rbody       []byte
	_wbody       interface{}
	Conn         *websocket.Conn
	Error        error
	_OnError     func(self *Websocket)
	_OnOpen      func(hc *HttpContext)
	_mutex       *sync.Mutex
	_doErrorFunc bool
}

func NewWebsocket(_id int64) *Websocket {
	return &Websocket{Id: _id, _mutex: new(sync.Mutex), _doErrorFunc: false}
}

func (this *Websocket) Send(v interface{}) (err error) {
	defer myRecover()
	if this.Error == nil {
		if err = websocket.Message.Send(this.Conn, v); err != nil {
			this.Error = err
		}
		this._onErrorChan()
	}
	return this.Error
}

func (this *Websocket) Read() []byte {
	return this._rbody
}

func (this *Websocket) Close() (err error) {
	return this.Conn.Close()
}

func (this *Websocket) _onErrorChan() {
	defer myRecover()
	if this.Error != nil && this._OnError != nil && !this._doErrorFunc {
		this._mutex.Lock()
		defer this._mutex.Unlock()
		if !this._doErrorFunc {
			this._doErrorFunc = true
			this._OnError(this)
		}
	}
}

type WebsocketConfig struct {
	Origin          string
	OriginFunc      func(origin *url.URL) bool
	MaxPayloadBytes int
	OnError         func(self *Websocket)
	OnOpen          func(hc *HttpContext)
}

/*http头信息*/
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
	return &HttpContext{w, r, hi, NewWebsocket(wsId(atomic.AddInt64(&_seqId, 1)))}
}

func wsId(_seq int64) (_r int64) {
	b := make([]byte, 16)
	copy(b[0:8], util.Int64ToBytes(util.RandId()))
	copy(b[8:], util.Int64ToBytes(time.Now().UnixNano()))
	_r = int64(util.CRC32(b))
	_r = _r<<32 | int64(int32(_seq))
	if _r < 0 {
		_r = -_r
	}
	return
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
		logger.Error(err)
	}
}

func matchString(pattern string, s string) bool {
	b, err := regexp.MatchString(pattern, s)
	if err != nil {
		b = false
	}
	return b
}

type mapl[K any, V any] struct {
	m   sync.Map
	len int64
	mux *sync.Mutex
}

func newMap[K any, V any]() *mapl[K, V] {
	return &mapl[K, V]{m: sync.Map{}, mux: &sync.Mutex{}}
}

func (this *mapl[K, V]) Put(key K, value V) {
	if _, ok := this.m.Swap(key, value); !ok {
		atomic.AddInt64(&this.len, 1)
	}
}

func (this *mapl[K, V]) Get(key K) (t V, ok bool) {
	if v, ok := this.m.Load(key); ok {
		return v.(V), ok
	}
	return t, false
}

func (this *mapl[K, V]) Has(key K) (ok bool) {
	_, ok = this.m.Load(key)
	return
}

func (this *mapl[K, V]) Del(key K) {
	this.mux.Lock()
	defer this.mux.Unlock()
	if _, ok := this.m.LoadAndDelete(key); ok {
		atomic.AddInt64(&this.len, -1)
	}
}

func (this *mapl[K, V]) Range(f func(k K, v V) bool) {
	this.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}

func (this *mapl[K, V]) Len() int64 {
	return this.len
}
