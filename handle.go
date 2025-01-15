// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import "net/url"

func (t *Tlnet) Handle(pattern string, handlerctx func(hc *HttpContext)) {
	logger.Debug("[Handle] " + pattern)
	t.addcontextfunc(defaultMethod, pattern, handlerctx)
}

func (t *Tlnet) POST(pattern string, handlerctx func(hc *HttpContext)) {
	logger.Debug("[POST] " + pattern)
	t.addcontextfunc(HttpPost, pattern, handlerctx)
}

func (t *Tlnet) PATCH(pattern string, handlerctx func(hc *HttpContext)) {
	logger.Debug("[PATCH] " + pattern)
	t.addcontextfunc(HttpPatch, pattern, handlerctx)
}

func (t *Tlnet) PUT(pattern string, handlerctx func(hc *HttpContext)) {
	logger.Debug("[PUT] " + pattern)
	t.addcontextfunc(HttpPut, pattern, handlerctx)
}

func (t *Tlnet) DELETE(pattern string, handlerctx func(hc *HttpContext)) {
	logger.Debug("[DELETE] " + pattern)
	t.addcontextfunc(HttpDelete, pattern, handlerctx)
}

func (t *Tlnet) GET(pattern string, handlerctx func(hc *HttpContext)) {
	logger.Debug("[GET] " + pattern)
	t.addcontextfunc(HttpGet, pattern, handlerctx)
}

func (t *Tlnet) OPTIONS(pattern string, handlerctx func(hc *HttpContext)) {
	logger.Debug("[OPTIONS] " + pattern)
	t.addcontextfunc(HttpOptions, pattern, handlerctx)
}

func (t *Tlnet) HEAD(pattern string, handlerctx func(hc *HttpContext)) {
	logger.Debug("[HEAD] " + pattern)
	t.addcontextfunc(HttpHead, pattern, handlerctx)
}

func (t *Tlnet) TRACE(pattern string, handlerctx func(hc *HttpContext)) {
	logger.Debug("[TRACE] " + pattern)
	t.addcontextfunc(HttpTrace, pattern, handlerctx)
}

func (t *Tlnet) CONNECT(pattern string, handlerctx func(hc *HttpContext)) {
	logger.Debug("[CONNECT] " + pattern)
	t.addcontextfunc(HttpConnect, pattern, handlerctx)
}

func (t *Tlnet) HandleWebSocket(pattern string, handlerctx func(hc *HttpContext)) {
	logger.Debug("[HandleWebSocket] " + pattern)
	t.wss = append(t.wss, newWsStub(pattern, &wsHandler{httpContextFunc: handlerctx}))
}

func (t *Tlnet) HandleWebSocketBindOrigin(pattern, origin string, handlerctx func(hc *HttpContext)) {
	logger.Debug("[HandleWebSocketBindOrigin] " + pattern)
	t.wss = append(t.wss, newWsStub(pattern, &wsHandler{httpContextFunc: handlerctx, origin: origin}))
}

func (t *Tlnet) HandleWebSocketBindOriginFunc(pattern string, handlerctx func(hc *HttpContext), originFunc func(origin *url.URL) bool) {
	logger.Debug("[HandleWebSocketBindOrigin] " + pattern)
	t.wss = append(t.wss, newWsStub(pattern, &wsHandler{httpContextFunc: handlerctx, originFunc: originFunc}))
}

func (t *Tlnet) HandleWebSocketBindConfig(pattern string, handlerctx func(hc *HttpContext), config *WebsocketConfig) {
	logger.Debug("[HandleWebSocketBindConfig] " + pattern)
	t.wss = append(t.wss, newWsStub(pattern, &wsHandler{httpContextFunc: handlerctx, originFunc: config.OriginFunc, origin: config.Origin, maxPayloadBytes: config.MaxPayloadBytes, onError: config.OnError, onOpen: config.OnOpen}))
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
