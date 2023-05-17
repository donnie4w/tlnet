package test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/donnie4w/simplelog/logging"
	. "github.com/donnie4w/tlnet"
)

func TestTlnet(t *testing.T) {
	tlnet := NewTlnet()
	tlnet.ReadTimeout(10 * time.Second)
	tlnet.Handle("/1", func(hc *HttpContext) { hc.ResponseBytes(200, []byte("hello tlnet,return bytes")) })
	tlnet.Handle("/2", func(hc *HttpContext) { hc.ResponseString("hello tlnet, getParam(name) value:" + hc.GetParam("name")) })
	tlnet.HandleWithFilter("/3/1", interceptFilter("[a-z]"), func(hc *HttpContext) { hc.ResponseString("filter test,正则[a-z]： /3/1不拦截") })
	tlnet.HandleWithFilter("/3/a", interceptFilter("[a-z]$"), func(hc *HttpContext) { hc.ResponseString("filter test, /3/a 被拦截") })

	tlnet.POST("/post", func(hc *HttpContext) {
		hc.ResponseString("this is a post method,post param value:" + hc.PostParam("name"))
	})

	tlnet.GET("/get", func(hc *HttpContext) {
		hc.ResponseString("this is a get method,get param value:" + hc.GetParam("name"))
	})

	tlnet.GET("/baidu", func(hc *HttpContext) {
		//重定向
		hc.Redirect("https://baidu.com")
	})

	tlnet.HandleStaticWithFilter("/", "./", notAllow("(.go)$"), nil)

	/****************************************************************************/
	tlnet.SetMaxBytesReader((1 << 20) * 50) //websocket body最大限制 50m
	tlnet.HandleWebSocket("/ws", websocketFunc)
	// tlnet.HandleWebSocketBindConfig("/ws2", websocketFunc, &WebsocketConfig{})
	tlnet.HttpStart(":8000")
}

func post(hc *HttpContext) {
	hc.ResponseString("this is a post method")
}

// 自定义正则拦截
func interceptFilter(pattern string) (f *Filter) {
	f = NewFilter()
	f.AddIntercept(pattern, func(hc *HttpContext) bool {
		hc.ResponseString(hc.ReqInfo.Uri + " is matched：" + pattern)
		return true // 返回true则 HandleWithFilter中handlerFunc不执行，false则会执行
	})
	return
}

// 自定义正则拦截
func notAllow(pattern string) (f *Filter) {
	f = NewFilter()
	f.AddIntercept(pattern, func(hc *HttpContext) bool {
		hc.ResponseBytes(http.StatusMethodNotAllowed, []byte(hc.ReqInfo.Uri+" not allow"))
		return true //// 返回true则 HandleWithFilter中handlerFunc不执行，false则会执行
	})
	return
}

func notFoundFilter() *Filter {
	f := NewFilter()
	f.AddPageNotFoundIntercept(func(hc *HttpContext) bool {
		hc.ResponseString(hc.ReqInfo.Uri + " not found")
		return true
	})
	return f
}

/************************************************************************/
//tlnet websocket test
var wsmap = make(map[int64]*HttpContext, 0)

func websocketFunc(hc *HttpContext) {
	//每个websocket有一个id，连接识别id
	wsId := hc.WS.Id
	msg := string(hc.WS.Read())
	_, ok := wsmap[wsId]
	if !ok {
		if msg == "123456" {
			wsmap[wsId] = hc
			hc.WS.Send(fmt.Sprint("id=", wsId))
		}
		return
	}
	logging.Debug(hc.ReqInfo)
	//收到客户端信息
	logging.Debug(msg)
	hc.WS.Send("你发送的是：" + msg)
	for k, v := range wsmap {
		if k != wsId {
			v.WS.Send(fmt.Sprint("这是【", wsId, "】发送的信息:", msg))
		}
	}
}
