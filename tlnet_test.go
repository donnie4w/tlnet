// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
package tlnet

import (
	"fmt"
	"io"
	"net"
	. "net/http"
	"testing"
	"time"

	"github.com/donnie4w/simplelog/logging"
)

func Test_tlnet(t *testing.T) {
	tlnet := NewTlnet()
	tlnet.ReadTimeout(10 * time.Second)
	tlnet.SetMaxBytesReader((1 << 20) * 50)
	tlnet.HandleStaticWithFilter("/cccc/", "db", notFoundFilter(), nil)
	tlnet.AddHandlerFunc("/aaa", nil, aaa)
	tlnet.AddHandlerFunc("/bbb", notFoundFilter(), aaa)
	tlnet.AddProcessor("/ppp", nil)
	tlnet.POST("/notify", notify)
	tlnet.AddStaticHandler("/", "./", nil, nil)
	tlnet.HandleWebSocket("/ws", websocketFunc)
	tlnet.HandleWebSocketBindOrigin("/ws2", "http://tlnet/", websocketFunc)
	tlnet.HandleWebSocketBindConfig("/ws3", websocketFunc, newWebsocketConfig())
	tlnet.HttpStart(":8082")
}

func _Test_tlnet2(t *testing.T) {
	tlnet := NewTlnet()
	// tlnet.DBPath("test.db")
	tlnet.SetMaxBytesReader((1 << 20) * 50)
	tlnet.Handle("/qq", handleFunc)
	tlnet.HandleStatic("/s", "test.db", staticHandleFunc)
	tlnet.HttpStart(":8080")
}

func handleFunc(hc *HttpContext) {
	logging.Debug(hc.ReqInfo)
	logging.Debug(hc.GetParam("name"))
	// logging.Debug(hc.ReqInfo.Header.Get("X-Real-IP"))
	// logging.Debug(hc.ReqInfo.Header.Get("X-Forward-For"))
	logging.Debug(net.SplitHostPort(hc.ReqInfo.RemoteAddr))
	hc.ResponseString("hello tlnet")
}

func staticHandleFunc(hc *HttpContext) {
	logging.Debug(hc.ReqInfo)
	logging.Debug(hc.GetParam("name"))
	logging.Debug(net.SplitHostPort(hc.ReqInfo.RemoteAddr))
	// hc.Redirect("https://baidu.com")
}

func notFoundFilter() *Filter {
	f := NewFilter()
	f.AddPageNotFoundIntercept(notFound)
	return f
}

func aaa(w ResponseWriter, r *Request) {
	logging.Debug("aaa")
	logging.Debug(fmt.Sprint(r.Header))
	io.WriteString(w, "hello aaa 你访问成功了")
}

func notFound(hc *HttpContext) bool {
	logging.Debug("notFound")
	logging.Debug(hc.ReqInfo.Header)
	hc.ResponseString("not found")
	return true
}

// 后缀为.html的过滤器
func suffixFilter() *Filter {
	f := NewFilter()
	f.AddSuffixIntercept([]string{"html"}, suffixIntercept)
	return f
}

func httpFilter() *Filter {
	f := NewFilter()
	f.AddIntercept(".html$", func(hc *HttpContext) bool {
		hc.Redirect("https://github.com")
		return true
	})
	return f
}

func suffixIntercept(hc *HttpContext) bool {
	hc.ResponseString("html is not allowed")
	return true
}

func staticFilter() *Filter {
	f := NewFilter()
	f.AddPageNotFoundIntercept(notFound)
	f.AddIntercept("[ab]", globalIntercept)
	return f
}

func globalIntercept(hc *HttpContext) bool {
	hc.ResponseString("globalIntercept")
	return true
}

func notify(hc *HttpContext) {
	for _, v := range wsmap {
		v.WS.Send(fmt.Sprint("通知：", time.Now()))
	}
}

var wsmap = make(map[int64]*HttpContext, 0)

func newWebsocketConfig() *WebsocketConfig {
	wc := new(WebsocketConfig)
	wc.MaxPayloadBytes = 1 << 20 * 100
	wc.OnError = func(ws *Websocket) {
		logging.Error("err:", ws.IsError)
		err := ws.Close()
		logging.Error(err)
	}
	return wc
}

func websocketFunc(hc *HttpContext) {
	_, ok := wsmap[hc.WS.Id]
	if !ok {
		wsmap[hc.WS.Id] = hc
	}
	logging.Debug(hc.ReqInfo)
	logging.Debug(hc.ReqInfo.RemoteAddr)
	// msg := string(hc.WS.Read())
	logging.Debug(len(hc.WS.Read()))
	// logging.Debug("收到:", msg)
	// hc.WS.Send([]byte("好了"))
	// hc.WS.Send("你发送的是：" + string(hc.WS.Read()))
	// for k, v := range wsmap {
	// 	if k != hc.WS.Id {
	// 		v.WS.Send(fmt.Sprint("这是", hc.WS.Id, "发送的信息:", msg))
	// 	}
	// }

}
