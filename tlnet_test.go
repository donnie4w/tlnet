package tlnet

import (
	"fmt"
	"io"
	. "net/http"
	"testing"

	"github.com/donnie4w/simplelog/logging"
)

func Test_tlnet(t *testing.T) {
	tlnet := NewTlnet()
	tlnet.DBPath("test.db")
	tlnet.SetMaxBytesReader((1 << 20) * 50)
	tlnet.AddHandlerFunc("/aaa", nil, aaa)
	tlnet.AddHandlerFunc("/bbb", notFoundFilter(), aaa)
	tlnet.AddProcessor("/ppp", nil)
	tlnet.AddStaticHandler("/", "test.db", staticFilter(), nil)
	tlnet.HttpStart(8080)
}

func notFoundFilter() *Filter {
	f := NewFilter()
	f.AddNotFoundPageIntercept(notFound)
	return f
}

func aaa(w ResponseWriter, r *Request) {
	logging.Debug("aaa")
	logging.Debug(fmt.Sprint(r.Header))
	io.WriteString(w, "hello aaa 你访问成功了")
}

func notFound(w ResponseWriter, r *Request) bool {
	logging.Debug("notFound")
	logging.Debug(fmt.Sprint(r.Header))
	io.WriteString(w, "not found 404")
	return true
}

//后缀为.html的过滤器
func suffixFilter() *Filter {
	f := NewFilter()
	f.AddSuffixIntercept([]string{"html"}, suffixIntercept)
	return f
}

func suffixIntercept(w ResponseWriter, r *Request) bool {
	io.WriteString(w, "html is not allowed")
	return true
}

func staticFilter() *Filter {
	f := NewFilter()
	f.AddNotFoundPageIntercept(permission)
	f.AddGlobalIntercept("[ab]", globalIntercept)
	return f
}

func globalIntercept(w ResponseWriter, r *Request) bool {
	io.WriteString(w, "globalIntercept")
	return true
}

func permission(w ResponseWriter, r *Request) bool {
	logging.Debug("permission")
	logging.Debug(fmt.Sprint(r.Header))
	err := r.Body.Close()
	logging.Debug(err)
	return true
}
