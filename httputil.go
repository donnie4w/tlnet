package tlnet

import (
	"errors"
	"fmt"
	"io"
	"regexp"

	// "math/rand"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/donnie4w/simplelog/logging"
)

type httpObj struct {
	w http.ResponseWriter
	r *http.Request
}

func (this *httpObj) GetCookie(key string) (_r string, err error) {
	cookieValue, er := this.r.Cookie(key)
	if er == nil {
		_r = cookieValue.Value
	}
	err = er
	return
}
func (this *httpObj) SetCookie(k, v string) {
	cookie := http.Cookie{Name: k, Value: v, Path: "/", MaxAge: 86400}
	http.SetCookie(this.w, &cookie)
}
func getIP(r *http.Request) (string, error) {
	ip := r.Header.Get("X-Real-IP")
	if net.ParseIP(ip) != nil {
		return ip, nil
	}
	ip = r.Header.Get("X-Forward-For")
	for _, i := range strings.Split(ip, ",") {
		if net.ParseIP(i) != nil {
			return i, nil
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	if net.ParseIP(ip) != nil {
		return ip, nil
	}
	return "", errors.New("no valid ip found")
}

func FormFile(r *http.Request, prefix, fileId, saveFilePath string) (filename string, err error) {
	defer myRecover()
	file, fileHeader, er := r.FormFile(fileId)
	if er != nil {
		err = er
		logging.Error(er.Error())
		return
	}
	defer file.Close()
	// rand.Seed(time.Now().Unix())
	filename = fmt.Sprint(prefix, "_", fileHeader.Filename)
	f, er := os.Create(fmt.Sprint(saveFilePath, "/", filename))
	err = er
	if er == nil {
		defer f.Close()
		_, err = io.Copy(f, file)
	}
	return
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
