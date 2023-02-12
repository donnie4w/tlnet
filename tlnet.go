package tlnet

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func (this *tlnet) Handle(pattern string, handlerFunc func(hc *HttpContext)) {
	this.AddHandlerFunc(pattern, nil, func(w http.ResponseWriter, r *http.Request) {
		handlerFunc(newHttpContext(w, r))
	})
}

func (this *tlnet) Post(pattern string, handlerFunc func(hc *HttpContext)) {
	this._methodpattern[pattern] = http.MethodPost
	this.Handle(pattern, handlerFunc)
}

func (this *tlnet) Patch(pattern string, handlerFunc func(hc *HttpContext)) {
	this._methodpattern[pattern] = http.MethodPatch
	this.Handle(pattern, handlerFunc)
}

func (this *tlnet) Put(pattern string, handlerFunc func(hc *HttpContext)) {
	this._methodpattern[pattern] = http.MethodPut
	this.Handle(pattern, handlerFunc)
}

func (this *tlnet) Delete(pattern string, handlerFunc func(hc *HttpContext)) {
	this._methodpattern[pattern] = http.MethodDelete
	this.Handle(pattern, handlerFunc)
}

func (this *tlnet) Get(pattern string, handlerFunc func(hc *HttpContext)) {
	this._methodpattern[pattern] = http.MethodGet
	this.Handle(pattern, handlerFunc)
}

func (this *tlnet) Options(pattern string, handlerFunc func(hc *HttpContext)) {
	this._methodpattern[pattern] = http.MethodOptions
	this.Handle(pattern, handlerFunc)
}

func (this *tlnet) Head(pattern string, handlerFunc func(hc *HttpContext)) {
	this._methodpattern[pattern] = http.MethodHead
	this.Handle(pattern, handlerFunc)
}

func (this *tlnet) Trace(pattern string, handlerFunc func(hc *HttpContext)) {
	this._methodpattern[pattern] = http.MethodTrace
	this.Handle(pattern, handlerFunc)
}

func (this *tlnet) Connect(pattern string, handlerFunc func(hc *HttpContext)) {
	this._methodpattern[pattern] = http.MethodConnect
	this.Handle(pattern, handlerFunc)
}

func (this *tlnet) WebSocketHandle(pattern string, handlerFunc func(hc *HttpContext)) {
	this._wss = append(this._wss, newWsStub(pattern, &wsHandler{httpContextFunc: handlerFunc}))
}

func (this *tlnet) HandleWithFilter(pattern string, _filter *Filter, handlerFunc func(hc *HttpContext)) {
	this.AddHandlerFunc(pattern, _filter, func(w http.ResponseWriter, r *http.Request) {
		handlerFunc(newHttpContext(w, r))
	})
}

func (this *tlnet) StaticHandle(pattern, dir string, handlerFunc func(hc *HttpContext)) {
	if handlerFunc != nil {
		this.AddStaticHandler(pattern, dir, nil, func(w http.ResponseWriter, r *http.Request) {
			handlerFunc(newHttpContext(w, r))
		})
	} else {
		this.AddStaticHandler(pattern, dir, nil, nil)
	}
}

func (this *tlnet) StaticHandleWithFilter(pattern, dir string, _filter *Filter, handlerFunc func(hc *HttpContext)) {
	if handlerFunc != nil {
		this.AddStaticHandler(pattern, dir, _filter, func(w http.ResponseWriter, r *http.Request) {
			handlerFunc(newHttpContext(w, r))
		})
	} else {
		this.AddStaticHandler(pattern, dir, _filter, nil)
	}
}

func (this *HttpContext) ResponseString(status int, _s string) (err error) {
	defer myRecover()
	if status == 0 {
		status = http.StatusOK
	}
	this.w.WriteHeader(status)
	_, err = this.w.Write([]byte(_s))
	return
}

// GetParam gets the first value associated with the given key.
// If there are no values associated with the key, Get returns
// the empty string. To access multiple values, use the map
// directly.
func (this *HttpContext) GetParam(key string) (_r string) {
	defer myRecover()
	_r = this.r.URL.Query().Get(key)
	return
}

// PostParam returns the first value for the named component of the query.
// POST and PUT body parameters take precedence over URL query string values.
// If key is not present, PostParam returns the empty string.
// To access multiple values of the same key, call PostParams
func (this *HttpContext) PostParam(key string) (_r string) {
	defer myRecover()
	_r = this.r.FormValue(key)
	return
}

/*multiple values of the same key*/
func (this *HttpContext) PostParams(key string) (_r []string) {
	defer myRecover()
	this.r.ParseForm()
	return this.r.Form[key]
}

func (this *HttpContext) Redirect(path string) (err error) {
	defer myRecover()
	http.Redirect(this.w, this.r, path, http.StatusTemporaryRedirect)
	return
}

func (this *HttpContext) RedirectWithStatus(path string, status int) (err error) {
	defer myRecover()
	http.Redirect(this.w, this.r, path, http.StatusTemporaryRedirect)
	return
}

func (this *HttpContext) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	defer myRecover()
	return this.r.FormFile(key)
}

func (this *HttpContext) FormFiles(key string) *multipart.Form {
	defer myRecover()
	return this.r.MultipartForm
}

func (this *HttpContext) Request() *http.Request {
	return this.r
}

func (this *HttpContext) Writer() http.ResponseWriter {
	return this.w
}

func ParseFormFile(file multipart.File, fileHeader *multipart.FileHeader, savePath, namePrefix string) (fileName string, err error) {
	defer myRecover()
	defer file.Close()
	f, er := os.Create(fmt.Sprint(savePath, "/", namePrefix, fileHeader.Filename))
	err = er
	if err == nil {
		fileName = fileHeader.Filename
		defer f.Close()
		_, err = io.Copy(f, file)
	}
	return
}
