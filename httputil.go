package tlnet

import (
	"net/http"
	"regexp"
	"time"

	"golang.org/x/net/websocket"

	"github.com/donnie4w/simplelog/logging"
)

type Websocket struct {
	Id      int64
	rbody   []byte
	wbody   interface{}
	ws      *websocket.Conn
	onError error
}

func (this *Websocket) Send(v interface{}) (err error) {
	if this.onError == nil {
		err = websocket.Message.Send(this.ws, v)
		this.onError = err
		return
	} else {
		return this.onError
	}
}

func (this *Websocket) Read() []byte {
	return this.rbody
}

type HttpInfo struct {
	Path       string
	Uri        string
	Method     string
	Host       string
	RemoteAddr string
	UserAgent  string
	Referer    string
	Header     http.Header
}

type HttpContext struct {
	w       http.ResponseWriter
	r       *http.Request
	ReqInfo *HttpInfo
	WS      *Websocket
}

func newHttpContext(w http.ResponseWriter, r *http.Request) *HttpContext {
	hi := new(HttpInfo)
	hi.Header, hi.Host, hi.Method, hi.Path, hi.RemoteAddr, hi.Uri, hi.UserAgent, hi.Referer = r.Header, r.Host, r.Method, r.URL.Path, r.RemoteAddr, r.RequestURI, r.UserAgent(), r.Referer()
	return &HttpContext{w, r, hi, &Websocket{Id: time.Now().UnixNano()}}
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

func (this *HttpContext) MaxBytesReader(_max int64) {
	this.r.Body = http.MaxBytesReader(this.w, this.r.Body, _max)
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
