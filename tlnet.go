// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/tlnet
package tlnet

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/donnie4w/simplelog/logging"
)

var logger = logging.NewLogger().SetFormat(logging.FORMAT_DATE | logging.FORMAT_TIME | logging.FORMAT_MICROSECNDS).SetLevel(logging.LEVEL_ERROR)

func (t *Tlnet) Handle(pattern string, handlerFunc func(hc *HttpContext)) {
	t.AddHandlerFunc(pattern, nil, func(w http.ResponseWriter, r *http.Request) {
		handlerFunc(newHttpContext(w, r))
	})
}

func (t *Tlnet) POST(pattern string, handlerFunc func(hc *HttpContext)) {
	t._methodpattern[pattern] = http.MethodPost
	t.Handle(pattern, handlerFunc)
}

func (t *Tlnet) PATCH(pattern string, handlerFunc func(hc *HttpContext)) {
	t._methodpattern[pattern] = http.MethodPatch
	t.Handle(pattern, handlerFunc)
}

func (t *Tlnet) PUT(pattern string, handlerFunc func(hc *HttpContext)) {
	t._methodpattern[pattern] = http.MethodPut
	t.Handle(pattern, handlerFunc)
}

func (t *Tlnet) DELETE(pattern string, handlerFunc func(hc *HttpContext)) {
	t._methodpattern[pattern] = http.MethodDelete
	t.Handle(pattern, handlerFunc)
}

func (t *Tlnet) GET(pattern string, handlerFunc func(hc *HttpContext)) {
	t._methodpattern[pattern] = http.MethodGet
	t.Handle(pattern, handlerFunc)
}

func (t *Tlnet) OPTIONS(pattern string, handlerFunc func(hc *HttpContext)) {
	t._methodpattern[pattern] = http.MethodOptions
	t.Handle(pattern, handlerFunc)
}

func (t *Tlnet) HEAD(pattern string, handlerFunc func(hc *HttpContext)) {
	t._methodpattern[pattern] = http.MethodHead
	t.Handle(pattern, handlerFunc)
}

func (t *Tlnet) TRACE(pattern string, handlerFunc func(hc *HttpContext)) {
	t._methodpattern[pattern] = http.MethodTrace
	t.Handle(pattern, handlerFunc)
}

func (t *Tlnet) CONNECT(pattern string, handlerFunc func(hc *HttpContext)) {
	t._methodpattern[pattern] = http.MethodConnect
	t.Handle(pattern, handlerFunc)
}

func (t *Tlnet) HandleWebSocket(pattern string, handlerFunc func(hc *HttpContext)) {
	t._wss = append(t._wss, newWsStub(pattern, &wsHandler{httpContextFunc: handlerFunc}))
}

func (t *Tlnet) HandleWebSocketBindOrigin(pattern, origin string, handlerFunc func(hc *HttpContext)) {
	t._wss = append(t._wss, newWsStub(pattern, &wsHandler{httpContextFunc: handlerFunc, _Origin: origin}))
}

func (t *Tlnet) HandleWebSocketBindOriginFunc(pattern string, handlerFunc func(hc *HttpContext), originFunc func(origin *url.URL) bool) {
	t._wss = append(t._wss, newWsStub(pattern, &wsHandler{httpContextFunc: handlerFunc, _OriginFunc: originFunc}))
}

func (t *Tlnet) HandleWebSocketBindConfig(pattern string, handlerFunc func(hc *HttpContext), config *WebsocketConfig) {
	t._wss = append(t._wss, newWsStub(pattern, &wsHandler{httpContextFunc: handlerFunc, _OriginFunc: config.OriginFunc, _Origin: config.Origin, _MaxPayloadBytes: config.MaxPayloadBytes, _OnError: config.OnError, _OnOpen: config.OnOpen}))
}

func (t *Tlnet) HandleWithFilter(pattern string, _filter *Filter, handlerFunc func(hc *HttpContext)) {
	if handlerFunc != nil {
		t.AddHandlerFunc(pattern, _filter, func(w http.ResponseWriter, r *http.Request) {
			handlerFunc(newHttpContext(w, r))
		})
	} else {
		t.AddHandlerFunc(pattern, _filter, nil)
	}
}

func (t *Tlnet) HandleStatic(pattern, dir string, handlerFunc func(hc *HttpContext)) {
	if handlerFunc != nil {
		t.AddStaticHandler(pattern, dir, nil, func(w http.ResponseWriter, r *http.Request) {
			handlerFunc(newHttpContext(w, r))
		})
	} else {
		t.AddStaticHandler(pattern, dir, nil, nil)
	}
}

func (t *Tlnet) HandleStaticWithFilter(pattern, dir string, _filter *Filter, handlerFunc func(hc *HttpContext)) {
	if handlerFunc != nil {
		t.AddStaticHandler(pattern, dir, _filter, func(w http.ResponseWriter, r *http.Request) {
			handlerFunc(newHttpContext(w, r))
		})
	} else {
		t.AddStaticHandler(pattern, dir, _filter, nil)
	}
}

func SetLogOFF() {
	logger.SetLevel(logging.LEVEL_OFF)
}

// 数据返回客户端
// return the data to the client
func (t *HttpContext) ResponseString(data string) (_r int, err error) {
	return t.ResponseBytes(http.StatusOK, []byte(data))
}

func (t *HttpContext) ResponseBytes(status int, bs []byte) (_r int, err error) {
	defer myRecover()
	if status == 0 {
		status = http.StatusOK
	}
	t.w.WriteHeader(status)
	if t.w.Header().Get("Content-Length") == "" {
		t.w.Header().Add("Content-Length", fmt.Sprint(len(bs)))
	}
	_r, err = t.w.Write(bs)
	return
}

// GetParam gets the first value associated with the given key.
// If there are no values associated with the key, Get returns
// the empty string. To access multiple values, use the map
// directly.
func (t *HttpContext) GetParam(key string) (_r string) {
	defer myRecover()
	_r = t.r.URL.Query().Get(key)
	return
}

// TrimSpace GetParam
func (t *HttpContext) GetParamTrimSpace(key string) (_r string) {
	return strings.TrimSpace(t.GetParam(key))
}

// PostParam returns the first value for the named component of the query.
// POST and PUT body parameters take precedence over URL query string values.
// If key is not present, PostParam returns the empty string.
// To access multiple values of the same key, call PostParams
func (t *HttpContext) PostParam(key string) (_r string) {
	defer myRecover()
	_r = t.r.FormValue(key)
	return
}

// TrimSpace PostParam
func (t *HttpContext) PostParamTrimSpace(key string) (_r string) {
	return strings.TrimSpace(t.PostParam(key))
}

/*multiple values of the same key*/
func (t *HttpContext) PostParams(key string) (_r []string) {
	defer myRecover()
	t.r.ParseForm()
	return t.r.Form[key]
}

// 重定向302
func (t *HttpContext) Redirect(path string) {
	defer myRecover()
	http.Redirect(t.w, t.r, path, http.StatusFound)
}

func (t *HttpContext) RedirectWithStatus(path string, status int) {
	defer myRecover()
	http.Redirect(t.w, t.r, path, status)
}

func (t *HttpContext) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	defer myRecover()
	return t.r.FormFile(key)
}

func (t *HttpContext) FormFiles(key string) *multipart.Form {
	defer myRecover()
	return t.r.MultipartForm
}

func (t *HttpContext) Request() *http.Request {
	return t.r
}

func (t *HttpContext) RequestBody() []byte {
	var buf bytes.Buffer
	io.Copy(&buf, t.r.Body)
	return buf.Bytes()
}

func (t *HttpContext) Writer() http.ResponseWriter {
	return t.w
}

func ParseFormFile(file multipart.File, fileHeader *multipart.FileHeader, savePath, namePrefix string) (fileName string, err error) {
	defer myRecover()
	defer file.Close()
	f, er := os.Create(fmt.Sprint(savePath, "/", namePrefix, fileHeader.Filename))
	err = er
	if err == nil {
		fileName = fileHeader.Filename
		defer f.Close()
		_, err = io.Copy(f, file)
	}
	return
}
