package tlnet

import (
	. "net/http"
	"strings"
)

func NewFilter() *Filter {
	f := new(Filter)
	f.suffixMap = make(map[string]int8, 0)
	f.matchMap = make(map[string]func(ResponseWriter, *Request) bool, 0)
	return f
}

type Filter struct {
	notFoundhandler func(ResponseWriter, *Request) bool            //uri not found bool为true时，执行func后，不再进行其他判断
	suffixMap       map[string]int8                                //suffix 后缀
	suffixHandler   func(ResponseWriter, *Request) bool            //bool为true时，执行func后，不再进行其他判断
	matchMap        map[string]func(ResponseWriter, *Request) bool //正则匹配 bool为true时，执行func后，不再进行其他判断
}

//suffixs 拦截器后缀数组，handler返回true则 不再进行其他流程判断，直接返回
func (this *Filter) AddSuffixIntercept(suffixs []string, _handler func(ResponseWriter, *Request) bool) {
	for _, v := range suffixs {
		if strings.Contains(v, ".") {
			v = v[strings.LastIndex(v, ".")+1:]
		}
		this.suffixMap[v] = 1
	}
	this.suffixHandler = _handler
}

//通用拦截器 ，handler返回true则 不再进行其他流程判断，直接返回
func (this *Filter) AddGlobalIntercept(_pattern string, _handler func(ResponseWriter, *Request) bool) {
	this.matchMap[_pattern] = _handler
}

//路径找不到拦截器 ，handler返回true则 不再进行其他流程判断，直接返回
func (this *Filter) AddNotFoundPageIntercept(_handler func(ResponseWriter, *Request) bool) {
	this.notFoundhandler = _handler
}

//过滤器增加拦截规则
func (this *Filter) AddIntercept(_pattern string, handlerFunc func(hc *HttpContext) bool) {
	this.matchMap[_pattern] = func(w ResponseWriter, r *Request) bool {
		return handlerFunc(newHttpContext(w, r))
	}
}

func (this *Filter) _processSuffix(uri string, w ResponseWriter, r *Request) bool {
	uri = strings.TrimSpace(uri)
	if strings.Contains(uri, ".") {
		suffix := uri[strings.LastIndex(uri, ".")+1:]
		if this.suffixMap[suffix] > 0 {
			if this.suffixHandler(w, r) {
				return true
			}
		}
	}
	return false
}

func (this *Filter) _processGlobal(path string, w ResponseWriter, r *Request) bool {
	path = strings.TrimSpace(path)
	for pattern, fun := range this.matchMap {
		if matchString(pattern, path) {
			if fun(w, r) {
				return true
			}
		}
	}
	return false
}
