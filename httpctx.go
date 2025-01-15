// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import (
	"bytes"
	"fmt"
	"github.com/donnie4w/gofer/util"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
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

// ResponseString
// return the data to the client
func (t *HttpContext) ResponseString(data string) (_r int, err error) {
	return t.ResponseBytes(http.StatusOK, []byte(data))
}

func (t *HttpContext) ResponseBytes(status int, bs []byte) (_r int, err error) {
	if status == 0 {
		status = http.StatusOK
	}
	t.w.WriteHeader(status)
	if len(bs) > 0 {
		_r, err = t.w.Write(bs)
	}
	return
}

// Error replies to the request with the specified error message and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// writes are done to w.
// The error message should be plain text.
func (t *HttpContext) Error(error string, code int) {
	http.Error(t.Writer(), error, code)
}

// GetParam gets the first value associated with the given key.
// If there are no values associated with the key, Get returns
// the empty string. To access multiple values, use the map
// directly.
func (t *HttpContext) GetParam(key string) (_r string) {
	_r = t.r.URL.Query().Get(key)
	if logger.IsVaild {
		logger.Debug("[GetParam] "+key, ", result:", _r)
	}
	return
}

// GetParamTrimSpace TrimSpace GetParam
func (t *HttpContext) GetParamTrimSpace(key string) (_r string) {
	_r = strings.TrimSpace(t.GetParam(key))
	if logger.IsVaild {
		logger.Debug("[GetParamTrimSpace] "+key, ", result:", _r)
	}
	return
}

// PostParam returns the first value for the named component of the query.
// POST and PUT body parameters take precedence over URL query string values.
// If key is not present, PostParam returns the empty string.
// To access multiple values of the same key, call PostParams
func (t *HttpContext) PostParam(key string) (_r string) {
	_r = t.r.FormValue(key)
	if logger.IsVaild {
		logger.Debug("[PostParam] "+key, ", result:", _r)
	}
	return
}

// PostParamTrimSpace TrimSpace PostParam
func (t *HttpContext) PostParamTrimSpace(key string) (_r string) {
	_r = strings.TrimSpace(t.PostParam(key))
	if logger.IsVaild {
		logger.Debug("[PostParamTrimSpace] "+key, ", result:", _r)
	}
	return
}

// PostParams
// multiple values of the same key
func (t *HttpContext) PostParams(key string) (_r []string) {
	t.r.ParseForm()
	_r = t.r.Form[key]
	if logger.IsVaild {
		logger.Debug("[PostParams] "+key, ", result:", _r)
	}
	return
}

// Redirect
// 重定向 302
func (t *HttpContext) Redirect(path string) {
	if logger.IsVaild {
		logger.Debug("[Redirect] " + path)
	}
	http.Redirect(t.w, t.r, path, http.StatusFound)
}

func (t *HttpContext) RedirectWithStatus(path string, status int) {
	if logger.IsVaild {
		logger.Debug("[RedirectWithStatus] "+path, ",status:", status)
	}
	http.Redirect(t.w, t.r, path, status)
}

func (t *HttpContext) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	if logger.IsVaild {
		logger.Debug("[FormFile] " + key)
	}
	return t.r.FormFile(key)
}

func (t *HttpContext) FormFiles(key string) *multipart.Form {
	if logger.IsVaild {
		logger.Debug("[FormFiles] " + key)
	}
	return t.r.MultipartForm
}

func (t *HttpContext) Request() *http.Request {
	return t.r
}

func (t *HttpContext) RequestBody() []byte {
	var buf bytes.Buffer
	io.Copy(&buf, t.r.Body)
	return buf.Bytes()
}

func (t *HttpContext) Writer() http.ResponseWriter {
	return t.w
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
