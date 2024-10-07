// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import (
	"fmt"
	"github.com/donnie4w/gofer/util"
	"net/http"
	"regexp"
)

// HttpInfo is a struct that represents the header information of an HTTP request.
type HttpInfo struct {
	// Path represents the target path of the request (e.g., "/index.html").
	Path string

	// Uri represents the full Uniform Resource Identifier of the request.
	Uri string

	// Method indicates the type of HTTP method used for the request (e.g., "GET", "POST").
	Method string

	// Host specifies the hostname or IP address of the server being requested, possibly including the port number.
	Host string

	// RemoteAddr specifies the network address of the client making the request, useful for logging or security purposes.
	RemoteAddr string

	// UserAgent describes the software (such as a browser version) that made the request.
	UserAgent string

	// Referer is the URL of the page that linked to the current page being requested.
	Referer string

	// Header is a map containing all the headers sent with the request, allowing multiple values per key.
	Header http.Header
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
	return &HttpContext{w, r, hi, nil}
}

func newHttpContextWithWebsocket(w http.ResponseWriter, r *http.Request) *HttpContext {
	hi := new(HttpInfo)
	hi.Header, hi.Host, hi.Method, hi.Path, hi.RemoteAddr, hi.Uri, hi.UserAgent, hi.Referer = r.Header, r.Host, r.Method, r.URL.Path, r.RemoteAddr, r.RequestURI, r.UserAgent(), r.Referer()
	return &HttpContext{w, r, hi, newWebsocket(util.UUID64())}
}

func (t *HttpContext) GetCookie(name string) (_r string, err error) {
	cookieValue, er := t.r.Cookie(name)
	if er == nil {
		_r = cookieValue.Value
	}
	err = er
	return
}

func (t *HttpContext) SetCookie(name, value, path string, maxAge int) {
	cookie := http.Cookie{Name: name, Value: value, Path: path, MaxAge: maxAge}
	http.SetCookie(t.w, &cookie)
}

func (t *HttpContext) SetCookie2(cookie *http.Cookie) {
	http.SetCookie(t.w, cookie)
}

func (t *HttpContext) MaxBytesReader(_max int64) {
	t.r.Body = http.MaxBytesReader(t.w, t.r.Body, _max)
}

func tlnetRecover(err *error) {
	if e := recover(); e != nil {
		if err != nil {
			er := fmt.Errorf("%v", e)
			err = &er
		}
	}
}

func matchString(pattern string, s string) bool {
	if b, err := regexp.MatchString(pattern, s); err != nil {
		return false
	} else {
		return b
	}
}
