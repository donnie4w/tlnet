// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
package tlnet

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"
)

type Websocket struct {
	Id           int64
	_rbody       []byte
	_wbody       interface{}
	Conn         *websocket.Conn
	IsError      error
	_OnError     func(self *Websocket)
	_mutex       *sync.Mutex
	_doErrorFunc bool
}

func NewWebsocket(_id int64) *Websocket {
	return &Websocket{Id: _id, _mutex: new(sync.Mutex), _doErrorFunc: false}
}

func (this *Websocket) Send(v interface{}) (err error) {
	if this.IsError == nil {
		if err = websocket.Message.Send(this.Conn, v); err != nil {
			this.IsError = err
		}
		this._onErrorChan()
	}
	return this.IsError
}

func (this *Websocket) Read() []byte {
	return this._rbody
}

func (this *Websocket) Close() (err error) {
	return this.Conn.Close()
}

func (this *Websocket) _onErrorChan() {
	if this.IsError != nil && this._OnError != nil && !this._doErrorFunc {
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
	_r = int64(_crc32(append(_int64ToBytes(int64(os.Getpid())), _int64ToBytes(time.Now().UnixNano())...)))
	_r = _r<<31 | int64(_seq&0x0000ffff)
	return
}

func _crc32(bs []byte) uint32 {
	return crc32.ChecksumIEEE(bs)
}

func _int64ToBytes(n int64) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes()
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

type Map[K any, V any] struct {
	m   sync.Map
	len int64
}

func newMap[K any, V any]() *Map[K, V] {
	return &Map[K, V]{m: sync.Map{}}
}

func (this *Map[K, V]) Put(key K, value V) {
	if _, ok := this.m.Swap(key, value); !ok {
		atomic.AddInt64(&this.len, 1)
	}
}

func (this *Map[K, V]) Get(key K) (t V, ok bool) {
	if v, ok := this.m.Load(key); ok {
		return v.(V), ok
	}
	return t, false
}

func (this *Map[K, V]) Has(key K) (ok bool) {
	_, ok = this.m.Load(key)
	return
}

func (this *Map[K, V]) Del(key K) {
	if _, ok := this.m.LoadAndDelete(key); ok {
		atomic.AddInt64(&this.len, -1)
	}
}

func (this *Map[K, V]) Range(f func(k K, v V) bool) {
	this.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}

func (this *Map[K, V]) Len() int64 {
	return this.len
}
