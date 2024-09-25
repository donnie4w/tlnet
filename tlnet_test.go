// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import (
	"fmt"
	"github.com/donnie4w/simplelog/logging"
	"net/http"
	_ "net/http/pprof"
	"testing"
)

func Test_tlnet(t *testing.T) {
	tl := NewTlnet()
	SetLogger(true)
	tl.GET("/a/", func(hc *HttpContext) { hc.ResponseString("path:" + hc.ReqInfo.Path) })
	tl.GET("/b/:id/:name/order", func(hc *HttpContext) { hc.ResponseString(hc.GetParam("id") + ":" + hc.GetParam("name")) })
	tl.GET("/test", func(hc *HttpContext) { hc.ResponseString("uri:/test") })
	tl.HandleWithFilter("/filter/", interceptFilter(), func(hc *HttpContext) { hc.ResponseString("filter test") })
	//tl.GET("/", func(hc *HttpContext) { hc.ResponseString("root path:" + hc.ReqInfo.Path) })
	tl.HandleStaticWithFilter("/img/", "./html/img", interceptFilter(), nil)
	tl.HttpStart(":8080")
}

func interceptFilter() *Filter {
	f := NewFilter()
	//内置后缀拦截器
	f.AddSuffixIntercept([]string{".jpg"}, func(hc *HttpContext) bool {
		logging.Debug("suffix  path:" + hc.ReqInfo.Path)
		return false
	})
	//内置路径未找到拦截器
	f.AddPageNotFoundIntercept(func(hc *HttpContext) bool {
		hc.Error("page not found:"+hc.ReqInfo.Uri, http.StatusNotFound)
		return true
	})
	//自定义URL正则匹配规则拦截器
	f.AddIntercept(".*?", func(hc *HttpContext) bool {
		logging.Debug("intercept:", hc.ReqInfo.Uri)
		if hc.ReqInfo.Path == "" {
			hc.Error("path is empty:"+hc.ReqInfo.Uri, http.StatusForbidden)
			return true
		}
		return false
	})
	return f
}

func TestWebsocket(t *testing.T) {
	tl := NewTlnet()
	tl.HandleWebSocket("/ws", websocketHandler)
	tl.HttpStart(":8080")

}
func websocketHandler(hc *HttpContext) {
	wsId := hc.WS.Id                                            //每个websocket链接都会生成一个识别id
	msg := string(hc.WS.Read())                                 //读取客户端请求数据
	hc.WS.Send("服务ID" + fmt.Sprint(wsId) + "已经收到你发送的信息:" + msg) //向客户端回复数据
}
