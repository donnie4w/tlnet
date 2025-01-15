// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import (
	"golang.org/x/net/websocket"
	"net/url"
	"sync"
)

type Websocket struct {
	Id           int64
	_rbody       []byte
	Conn         *websocket.Conn
	Error        error
	_OnError     func(self *Websocket)
	_mutex       *sync.Mutex
	_doErrorFunc bool
}

func newWebsocket(_id int64) *Websocket {
	return &Websocket{Id: _id, _mutex: new(sync.Mutex), _doErrorFunc: false}
}

func (t *Websocket) Send(v any) (err error) {
	defer tlnetRecover(&err)
	if t.Error == nil {
		if err = websocket.Message.Send(t.Conn, v); err != nil {
			t.Error = err
		}
		t._onErrorChan()
	}
	return t.Error
}

func (t *Websocket) Read() []byte {
	return t._rbody
}

func (t *Websocket) Close() (err error) {
	return t.Conn.Close()
}

func (t *Websocket) _onErrorChan() {
	if t.Error != nil && t._OnError != nil && !t._doErrorFunc {
		t._mutex.Lock()
		defer t._mutex.Unlock()
		if !t._doErrorFunc {
			t._doErrorFunc = true
			t._OnError(t)
		}
	}
}

type WebsocketConfig struct {
	Origin          string
	OriginFunc      func(origin *url.URL) bool
	MaxPayloadBytes int
	OnError         func(self *Websocket)
	OnOpen          func(hc *HttpContext)
}
