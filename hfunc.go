// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import (
	"github.com/apache/thrift/lib/go/thrift"
	"net/http"
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

func (t *Tlnet) addcontextfunc(method httpMethod, pattern string, handlerctx func(hc *HttpContext)) {
	t.addhandlerFunc(method, pattern, nil, func(w http.ResponseWriter, r *http.Request) {
		handlerctx(newHttpContext(w, r))
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
