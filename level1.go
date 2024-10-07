// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import (
	"github.com/apache/thrift/lib/go/thrift"
	"golang.org/x/net/websocket"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
)

// HandleJsonProcessor   json protocol
func (t *Tlnet) HandleJsonProcessor(pattern string, processor thrift.TProcessor) {
	logger.Debug("[AddProcessor]:", pattern)
	t.addProcessor(pattern, processor, JSON)
}

// HandleBinaryProcessor binary protocol
func (t *Tlnet) HandleBinaryProcessor(pattern string, processor thrift.TProcessor) {
	logger.Debug("[HandleBinaryProcessor]:", pattern)
	t.addProcessor(pattern, processor, BINARY)
}

// HandleCompactProcessor compact protocol
func (t *Tlnet) HandleCompactProcessor(pattern string, processor thrift.TProcessor) {
	logger.Debug("[HandleCompactProcessor]:", pattern)
	t.addProcessor(pattern, processor, COMPACT)
}

// HandleFunc 处理动态请求
func (t *Tlnet) HandleFunc(pattern string, f *Filter, handlerFunc func(http.ResponseWriter, *http.Request)) {
	t.addhandlerFunc(defaultMethod, pattern, f, handlerFunc)
}

// HandleStaticFunc 处理静态资源
func (t *Tlnet) HandleStaticFunc(pattern string, dir string, f *Filter, handlerFunc func(http.ResponseWriter, *http.Request)) {
	t.addstatichandlerFunc("", pattern, dir, f, handlerFunc)
}

func (t *Tlnet) addcontextfunc(method httpMethod, pattern string, contextfunc func(hc *HttpContext)) {
	t.addhandlerFunc(method, pattern, nil, func(w http.ResponseWriter, r *http.Request) {
		contextfunc(newHttpContext(w, r))
	})
}

func (t *Tlnet) addhandlerFunc(method httpMethod, pattern string, f *Filter, handlerFunc func(http.ResponseWriter, *http.Request)) {
	if pattern == "" {
		panic("pattern must not be empty")
	}
	t.handlers = append(t.handlers, newStub(pattern, "", f, handlerFunc, nil, 0, method))
}

func (t *Tlnet) addstatichandlerFunc(method httpMethod, pattern string, dir string, f *Filter, handlerFunc func(http.ResponseWriter, *http.Request)) {
	if pattern == "" {
		panic("pattern must not be empty")
	}
	t.staticHandlers = append(t.staticHandlers, newStub(pattern, dir, f, handlerFunc, nil, 0, method))
}

func (t *Tlnet) addProcessor(pattern string, processor thrift.TProcessor, ttype tf_pco_type) {
	if pattern == "" {
		panic("pattern must not be empty")
	}
	t.processors = append(t.processors, newStub(pattern, "", nil, nil, processor, ttype, defaultMethod))
}

func (t *Tlnet) addhandlerctx(method httpMethod, pattern string, f *Filter, handlerctx func(hc *HttpContext)) {
	var handlerfunc func(http.ResponseWriter, *http.Request)
	if handlerctx != nil {
		handlerfunc = func(w http.ResponseWriter, r *http.Request) {
			handlerctx(newHttpContext(w, r))
		}
	}
	t.addhandlerFunc(method, pattern, f, handlerfunc)
}

func (t *Tlnet) addstatichandlerctx(method httpMethod, pattern string, dir string, f *Filter, handlerctx func(hc *HttpContext)) {
	var handlerfunc func(http.ResponseWriter, *http.Request)
	if handlerctx != nil {
		handlerfunc = func(w http.ResponseWriter, r *http.Request) {
			handlerctx(newHttpContext(w, r))
		}
	}
	t.addstatichandlerFunc(method, pattern, dir, f, handlerfunc)
}

type Websocket struct {
	Id           int64
	_rbody       []byte
	Conn         *websocket.Conn
	Error        error
	_OnError     func(self *Websocket)
	_mutex       *sync.Mutex
	_doErrorFunc bool
}

func newWebsocket(_id int64) *Websocket {
	return &Websocket{Id: _id, _mutex: new(sync.Mutex), _doErrorFunc: false}
}

func (t *Websocket) Send(v any) (err error) {
	defer tlnetRecover(&err)
	if t.Error == nil {
		if err = websocket.Message.Send(t.Conn, v); err != nil {
			t.Error = err
		}
		t._onErrorChan()
	}
	return t.Error
}

func (t *Websocket) Read() []byte {
	return t._rbody
}

func (t *Websocket) Close() (err error) {
	return t.Conn.Close()
}

func (t *Websocket) _onErrorChan() {
	if t.Error != nil && t._OnError != nil && !t._doErrorFunc {
		t._mutex.Lock()
		defer t._mutex.Unlock()
		if !t._doErrorFunc {
			t._doErrorFunc = true
			t._OnError(t)
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
