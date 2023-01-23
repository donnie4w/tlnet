package tlnet

import (
	"net/http"
	"regexp"

	"github.com/donnie4w/simplelog/logging"
)

type HttpInfo struct {
	Path       string
	Uri        string
	Method     string
	Host       string
	RemoteAddr string
	UserAgent  string
	Header     http.Header
}

type HttpContext struct {
	w       http.ResponseWriter
	r       *http.Request
	ReqInfo *HttpInfo
}

func newHttpContext(w http.ResponseWriter, r *http.Request) *HttpContext {
	hi := new(HttpInfo)
	hi.Header, hi.Host, hi.Method, hi.Path, hi.RemoteAddr, hi.Uri, hi.UserAgent = r.Header, r.Host, r.Method, r.URL.Path, r.RemoteAddr, r.RequestURI, r.UserAgent()
	return &HttpContext{w, r, hi}
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

func myRecover() {
	if err := recover(); err != nil {
		logging.Error(err)
	}
}

func matchString(pattern string, s string) bool {
	b, err := regexp.MatchString(pattern, s)
	if err != nil {
		b = false
	}
	return b
}
