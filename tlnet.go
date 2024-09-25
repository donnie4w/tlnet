// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

func (t *Tlnet) Handle(pattern string, handlerFunc func(hc *HttpContext)) {
	logger.Debug("[Handle] " + pattern)
	t.handlerFunc(defaultMethod, pattern, handlerFunc)
}

func (t *Tlnet) POST(pattern string, handlerFunc func(hc *HttpContext)) {
	logger.Debug("[POST] " + pattern)
	t.handlerFunc(HttpPost, pattern, handlerFunc)
}

func (t *Tlnet) PATCH(pattern string, handlerFunc func(hc *HttpContext)) {
	logger.Debug("[PATCH] " + pattern)
	t.handlerFunc(HttpPatch, pattern, handlerFunc)
}

func (t *Tlnet) PUT(pattern string, handlerFunc func(hc *HttpContext)) {
	logger.Debug("[PUT] " + pattern)
	t.handlerFunc(HttpPut, pattern, handlerFunc)
}

func (t *Tlnet) DELETE(pattern string, handlerFunc func(hc *HttpContext)) {
	logger.Debug("[DELETE] " + pattern)
	t.handlerFunc(HttpDelete, pattern, handlerFunc)
}

func (t *Tlnet) GET(pattern string, handlerFunc func(hc *HttpContext)) {
	logger.Debug("[GET] " + pattern)
	t.handlerFunc(HttpGet, pattern, handlerFunc)
}

func (t *Tlnet) OPTIONS(pattern string, handlerFunc func(hc *HttpContext)) {
	logger.Debug("[OPTIONS] " + pattern)
	t.handlerFunc(HttpOptions, pattern, handlerFunc)
}

func (t *Tlnet) HEAD(pattern string, handlerFunc func(hc *HttpContext)) {
	logger.Debug("[HEAD] " + pattern)
	t.handlerFunc(HttpHead, pattern, handlerFunc)
}

func (t *Tlnet) TRACE(pattern string, handlerFunc func(hc *HttpContext)) {
	logger.Debug("[TRACE] " + pattern)
	t.handlerFunc(HttpTrace, pattern, handlerFunc)
}

func (t *Tlnet) CONNECT(pattern string, handlerFunc func(hc *HttpContext)) {
	logger.Debug("[CONNECT] " + pattern)
	t.handlerFunc(HttpConnect, pattern, handlerFunc)
}

func (t *Tlnet) HandleWebSocket(pattern string, handlerFunc func(hc *HttpContext)) {
	logger.Debug("[HandleWebSocket] " + pattern)
	t.wss = append(t.wss, newWsStub(pattern, &wsHandler{httpContextFunc: handlerFunc}))
}

func (t *Tlnet) HandleWebSocketBindOrigin(pattern, origin string, handlerFunc func(hc *HttpContext)) {
	logger.Debug("[HandleWebSocketBindOrigin] " + pattern)
	t.wss = append(t.wss, newWsStub(pattern, &wsHandler{httpContextFunc: handlerFunc, origin: origin}))
}

func (t *Tlnet) HandleWebSocketBindOriginFunc(pattern string, handlerFunc func(hc *HttpContext), originFunc func(origin *url.URL) bool) {
	logger.Debug("[HandleWebSocketBindOrigin] " + pattern)
	t.wss = append(t.wss, newWsStub(pattern, &wsHandler{httpContextFunc: handlerFunc, originFunc: originFunc}))
}

func (t *Tlnet) HandleWebSocketBindConfig(pattern string, handlerFunc func(hc *HttpContext), config *WebsocketConfig) {
	logger.Debug("[HandleWebSocketBindConfig] " + pattern)
	t.wss = append(t.wss, newWsStub(pattern, &wsHandler{httpContextFunc: handlerFunc, originFunc: config.OriginFunc, origin: config.Origin, maxPayloadBytes: config.MaxPayloadBytes, onError: config.OnError, onOpen: config.OnOpen}))
}

func (t *Tlnet) HandleWithFilter(pattern string, _filter *Filter, handlerctx func(hc *HttpContext)) {
	logger.Debug("[HandleWithFilter] " + pattern)
	t.addhandlerctx(defaultMethod, pattern, _filter, handlerctx)
}

func (t *Tlnet) HandleStatic(pattern, dir string, handlerctx func(hc *HttpContext)) {
	logger.Debug("[HandleStatic] " + pattern)
	t.addstatichandlerctx(defaultMethod, pattern, dir, nil, handlerctx)
}

func (t *Tlnet) HandleStaticWithFilter(pattern, dir string, _filter *Filter, handlerctx func(hc *HttpContext)) {
	logger.Debug("[HandleStaticWithFilter] " + pattern)
	t.addstatichandlerctx(defaultMethod, pattern, dir, _filter, handlerctx)
}

func SetLogger(on bool) {
	logger.SetLogger(on)
}

func SetLoggerLevel(l level) {
	logger.SetLoggerLevel(l)
}

// ResponseString
// return the data to the client
func (t *HttpContext) ResponseString(data string) (_r int, err error) {
	return t.ResponseBytes(http.StatusOK, []byte(data))
}

func (t *HttpContext) ResponseBytes(status int, bs []byte) (_r int, err error) {
	if status == 0 {
		status = http.StatusOK
	}
	t.w.WriteHeader(status)
	if len(bs) > 0 {
		_r, err = t.w.Write(bs)
	}
	return
}

// Error replies to the request with the specified error message and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// writes are done to w.
// The error message should be plain text.
func (t *HttpContext) Error(error string, code int) {
	http.Error(t.Writer(), error, code)
}

// GetParam gets the first value associated with the given key.
// If there are no values associated with the key, Get returns
// the empty string. To access multiple values, use the map
// directly.
func (t *HttpContext) GetParam(key string) (_r string) {
	_r = t.r.URL.Query().Get(key)
	if logger.IsVaild {
		logger.Debug("[GetParam] "+key, ", result:", _r)
	}
	return
}

// GetParamTrimSpace TrimSpace GetParam
func (t *HttpContext) GetParamTrimSpace(key string) (_r string) {
	_r = strings.TrimSpace(t.GetParam(key))
	if logger.IsVaild {
		logger.Debug("[GetParamTrimSpace] "+key, ", result:", _r)
	}
	return
}

// PostParam returns the first value for the named component of the query.
// POST and PUT body parameters take precedence over URL query string values.
// If key is not present, PostParam returns the empty string.
// To access multiple values of the same key, call PostParams
func (t *HttpContext) PostParam(key string) (_r string) {
	_r = t.r.FormValue(key)
	if logger.IsVaild {
		logger.Debug("[PostParam] "+key, ", result:", _r)
	}
	return
}

// PostParamTrimSpace TrimSpace PostParam
func (t *HttpContext) PostParamTrimSpace(key string) (_r string) {
	_r = strings.TrimSpace(t.PostParam(key))
	if logger.IsVaild {
		logger.Debug("[PostParamTrimSpace] "+key, ", result:", _r)
	}
	return
}

// PostParams
// multiple values of the same key
func (t *HttpContext) PostParams(key string) (_r []string) {
	t.r.ParseForm()
	_r = t.r.Form[key]
	if logger.IsVaild {
		logger.Debug("[PostParams] "+key, ", result:", _r)
	}
	return
}

// Redirect
// 重定向 302
func (t *HttpContext) Redirect(path string) {
	if logger.IsVaild {
		logger.Debug("[Redirect] " + path)
	}
	http.Redirect(t.w, t.r, path, http.StatusFound)
}

func (t *HttpContext) RedirectWithStatus(path string, status int) {
	if logger.IsVaild {
		logger.Debug("[RedirectWithStatus] "+path, ",status:", status)
	}
	http.Redirect(t.w, t.r, path, status)
}

func (t *HttpContext) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	if logger.IsVaild {
		logger.Debug("[FormFile] " + key)
	}
	return t.r.FormFile(key)
}

func (t *HttpContext) FormFiles(key string) *multipart.Form {
	if logger.IsVaild {
		logger.Debug("[FormFiles] " + key)
	}
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
