// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import (
	"net/http"
)

type tf_pco_type byte
type muxmod byte
type httpMethod string

const (
	_ tf_pco_type = iota
	JSON
	BINARY
	COMPACT
)

const (
	_ muxmod = iota
	tlnetmod
	nativemod
)

const (
	_SS = 1 << 18
	_MS = 1 << 20
	_LS = 1 << 24
)

const (
	HttpGet       httpMethod = http.MethodGet
	HttpHead      httpMethod = http.MethodHead
	HttpPost      httpMethod = http.MethodPost
	HttpPut       httpMethod = http.MethodPut
	HttpPatch     httpMethod = http.MethodPatch
	HttpDelete    httpMethod = http.MethodDelete
	HttpConnect   httpMethod = http.MethodConnect
	HttpOptions   httpMethod = http.MethodOptions
	HttpTrace     httpMethod = http.MethodTrace
	defaultMethod httpMethod = ""
)
