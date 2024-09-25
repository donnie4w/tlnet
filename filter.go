// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import (
	"errors"
	"github.com/donnie4w/gofer/hashmap"
	"net/http"
	"strings"
)

func NewFilter() *Filter {
	f := new(Filter)
	f.suffixMap = hashmap.NewMapL[string, int8]()
	f.matchMap = hashmap.NewMapL[string, func(http.ResponseWriter, *http.Request) bool]()
	return f
}

type Filter struct {
	notFoundhandler func(http.ResponseWriter, *http.Request) bool                        //uri not found bool为true时，执行func后，不再进行其他判断
	suffixMap       *hashmap.MapL[string, int8]                                          //map[string]int8  //suffix 后缀
	suffixHandler   func(http.ResponseWriter, *http.Request) bool                        //bool为true时，执行func后，不再进行其他判断
	matchMap        *hashmap.MapL[string, func(http.ResponseWriter, *http.Request) bool] //map[string]func(ResponseWriter, *Request) bool //正则匹配 bool为true时，执行func后，不再进行其他判断
}

// AddSuffixIntercept
//
// suffixs 后缀拦截，方法返回true则，则不执行Filter后的handlerFunc，直接返回
func (t *Filter) AddSuffixIntercept(suffixs []string, handler func(hc *HttpContext) bool) {
	for _, v := range suffixs {
		if strings.Contains(v, ".") {
			v = v[strings.LastIndex(v, ".")+1:]
		}
		t.suffixMap.Put(v, 1)
	}
	t.suffixHandler = func(w http.ResponseWriter, r *http.Request) bool {
		return handler(newHttpContext(w, r))
	}
}

// AddPageNotFoundIntercept
//
// url路径找不到 ，方法返回true则，则不执行Filter后的handlerFunc，直接返回
func (t *Filter) AddPageNotFoundIntercept(handler func(hc *HttpContext) bool) {
	t.notFoundhandler = func(w http.ResponseWriter, r *http.Request) bool {
		return handler(newHttpContext(w, r))
	}
}

// AddIntercept
//
// 定义拦截规则 pattern正则匹配
func (t *Filter) AddIntercept(pattern string, handler func(hc *HttpContext) bool) (err error) {
	if t.matchMap.Has(pattern) {
		logger.Error("Duplicate matching[", pattern, "]")
		err = errors.New("Duplicate matching[" + pattern + "]")
		return
	}
	t.matchMap.Put(pattern, func(w http.ResponseWriter, r *http.Request) bool {
		return handler(newHttpContext(w, r))
	})
	return
}

func (t *Filter) processSuffix(uri string, w http.ResponseWriter, r *http.Request) bool {
	if index := strings.LastIndex(uri, "."); index > -1 {
		suffix := uri[index+1:]
		if t.suffixMap.Has(suffix) {
			if t.suffixHandler(w, r) {
				return true
			}
		}
	}
	return false
}

func (t *Filter) processGlobal(path string, w http.ResponseWriter, r *http.Request) (_r bool) {
	t.matchMap.Range(func(pattern string, fun func(http.ResponseWriter, *http.Request) bool) bool {
		if matchString(pattern, path) {
			if fun(w, r) {
				_r = true
				return false
			}
		}
		return true
	})
	return
}
