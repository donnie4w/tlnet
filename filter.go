// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
package tlnet

import (
	"errors"
	. "net/http"
	"strings"
)

func NewFilter() *Filter {
	f := new(Filter)
	f.suffixMap = newMap[string, int8]()                               //make(map[string]int8, 0)
	f.matchMap = newMap[string, func(ResponseWriter, *Request) bool]() //make(map[string]func(ResponseWriter, *Request) bool, 0)
	return f
}

type Filter struct {
	notFoundhandler func(ResponseWriter, *Request) bool               //uri not found bool为true时，执行func后，不再进行其他判断
	suffixMap       *Map[string, int8]                                //map[string]int8  //suffix 后缀
	suffixHandler   func(ResponseWriter, *Request) bool               //bool为true时，执行func后，不再进行其他判断
	matchMap        *Map[string, func(ResponseWriter, *Request) bool] //map[string]func(ResponseWriter, *Request) bool //正则匹配 bool为true时，执行func后，不再进行其他判断
}

// suffixs 后缀拦截，方法返回true则，则不执行Filter后的handlerFunc，直接返回
func (this *Filter) AddSuffixIntercept(suffixs []string, handlerFunc func(hc *HttpContext) bool) {
	for _, v := range suffixs {
		if strings.Contains(v, ".") {
			v = v[strings.LastIndex(v, ".")+1:]
		}
		this.suffixMap.Put(v, 1)
	}
	this.suffixHandler = func(w ResponseWriter, r *Request) bool {
		return handlerFunc(newHttpContext(w, r))
	}
}

// url路径找不到 ，方法返回true则，则不执行Filter后的handlerFunc，直接返回
func (this *Filter) AddPageNotFoundIntercept(handlerFunc func(hc *HttpContext) bool) {
	this.notFoundhandler = func(w ResponseWriter, r *Request) bool {
		return handlerFunc(newHttpContext(w, r))
	}
}

// 自定义拦截规则 pattern正则匹配
func (this *Filter) AddIntercept(_pattern string, handlerFunc func(hc *HttpContext) bool) (err error) {
	if this.matchMap.Has(_pattern) {
		logger.Fatal("Duplicate matching[", _pattern, "]")
		err = errors.New("Duplicate matching[" + _pattern + "]")
		return
	}
	this.matchMap.Put(_pattern, func(w ResponseWriter, r *Request) bool {
		return handlerFunc(newHttpContext(w, r))
	})
	return
}

func (this *Filter) _processSuffix(uri string, w ResponseWriter, r *Request) bool {
	uri = strings.TrimSpace(uri)
	if strings.Contains(uri, ".") {
		suffix := uri[strings.LastIndex(uri, ".")+1:]
		if this.suffixMap.Has(suffix) {
			if this.suffixHandler(w, r) {
				return true
			}
		}
	}
	return false
}

func (this *Filter) _processGlobal(path string, w ResponseWriter, r *Request) (_r bool) {
	path = strings.TrimSpace(path)
	this.matchMap.Range(func(pattern string, fun func(ResponseWriter, *Request) bool) bool {
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
