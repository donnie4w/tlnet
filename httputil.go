package tlnet

import (
	"net/http"
	"regexp"

	"github.com/donnie4w/simplelog/logging"
)

type HttpContext struct {
	w http.ResponseWriter
	r *http.Request
}

func (this *HttpContext) GetCookie(key string) (_r string, err error) {
	cookieValue, er := this.r.Cookie(key)
	if er == nil {
		_r = cookieValue.Value
	}
	err = er
	return
}
func (this *HttpContext) SetCookie(k, v string) {
	cookie := http.Cookie{Name: k, Value: v, Path: "/", MaxAge: 86400}
	http.SetCookie(this.w, &cookie)
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
