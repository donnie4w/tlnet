package tlnet

import (
	"fmt"
	"io"
	"net"
	"net/http"
	. "net/http"
	"testing"
	"time"

	"github.com/donnie4w/simplelog/logging"
	. "github.com/donnie4w/tlnet/db"
)

type TestObj struct {
	Id   int64
	Name string `idx`
	Age_ int
}

func init() {
	UseSimpleDB("tl.db")
}

func _Test_DB(t *testing.T) {
	var err error
	for i := 0; i < 10; i++ {
		err = Insert(&TestObj{Name: "wuxiaodong", Age_: 10 + i})
		SimpleDB().Put([]byte(fmt.Sprint("www", i)), []byte(fmt.Sprint("xxxx", i)))
	}

	var s string
	// err, s = BuildIndex[TestObj]()
	fmt.Println("————————————————————————————————————————————", err)
	fmt.Println("————————————————————————————————————————————", s)

	// err = Update(&TestObj{3, "wuxiaodong", 222})
	// Delete(TestObj{Id: 3})
	// Delete(&TestObj{Id: 3})
	time.Sleep(3 * time.Second)
	ts := Selects[TestObj](0, 10)
	for i, v := range ts {
		logging.Debug(i+1, "----", v)
	}
	logging.Debug("max idx==>", GetIdSeqValue[TestObj]())

	fmt.Println("------------------------------------------------")
	ts = SelectByIdxName[TestObj]("name", "wuxiaodong")
	for i, v := range ts {
		logging.Debug(i+1, "=====", v)
	}
	fmt.Println("------------------------------------------------")
	ts = SelectByIdxNameLimit[TestObj]("age_", []string{"215", "216", "333"}, 2, 2)
	for i, v := range ts {
		logging.Debug(i+1, "=========>", v)
	}
	o := SelectOneByIdxName[TestObj]("name", "dongdong")
	logging.Debug("o==>", o)
	fmt.Println("")
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	// IterDB()
}

// func Benchmark_Alloc(b *testing.B) {
// 	var i int
// 	for i = 0; i < b.N; i++ {
// 		fmt.Sprintf("%d", i)
// 		// Insert(&TestObj{Name_: "wuxiaodong", Age_: i})
// 		ts := Selects[TestObj](0, 10)
// 		for i, v := range ts {
// 			logging.Debug(i+1, "----", v)
// 		}
// 		ts = SelectByIdxName[TestObj]("Age_", "3370")
// 		for i, v := range ts {
// 			logging.Debug(i+1, "=====", v)
// 		}
// 		ts = SelectByIdxNameLimit[TestObj]("age", []string{"215", "216", "333"}, 0, 2)
// 		for i, v := range ts {
// 			logging.Debug(i+1, "=========>", v)
// 		}
// 	}
// 	logging.Debug("i===>", i)
// }

// func IterDB() {
// 	keys, _ := SimpleDB().GetKeys()
// 	for i, v := range keys {
// 		logging.Debug("key", i+1, "==", v)
// 		value, _ := SimpleDB().GetString([]byte(v))
// 		logging.Debug(v, "==>", value)
// 	}
// }

// func _Test_snap(t *testing.T) {
// 	SimpleDB().Put([]byte("d"), []byte("3"))
// 	logging.Debug(SimpleDB().GetKeys())
// 	er := SimpleDB().BackupToDisk("snap.lb", []byte("d"))
// 	logging.Debug(er)
// 	logging.Debug(RecoverBackup("snap.lb"))
// 	for _, v := range RecoverBackup("snap.lb") {
// 		logging.Debug(string(v.Key), " == ", string(v.Value))
// 	}
// }

func Test_tlnet(t *testing.T) {
	tlnet := NewTlnet()
	tlnet.DBPath("tl.db")
	tlnet.SetMaxBytesReader((1 << 20) * 50)
	tlnet.AddHandlerFunc("/aaa", nil, aaa)
	tlnet.AddHandlerFunc("/bbb", notFoundFilter(), aaa)
	tlnet.AddProcessor("/ppp", nil)
	tlnet.Handle("/notify", notify)
	tlnet.AddStaticHandler("/", "./", nil, nil)
	tlnet.WebSocketHandle("/ws", websocketFunc)
	OpenView(3434)
	tlnet.HttpStart(":8082")
}

func _Test_tlnet2(t *testing.T) {
	tlnet := NewTlnet()
	// tlnet.DBPath("test.db")
	tlnet.SetMaxBytesReader((1 << 20) * 50)
	tlnet.Handle("/qq", handleFunc)
	tlnet.StaticDir("/s", "test.db", staticHandleFunc)
	tlnet.HttpStart(":8080")
}

func handleFunc(hc *HttpContext) {
	logging.Debug(hc.ReqInfo)
	logging.Debug(hc.GetParam("name"))
	// logging.Debug(hc.ReqInfo.Header.Get("X-Real-IP"))
	// logging.Debug(hc.ReqInfo.Header.Get("X-Forward-For"))
	logging.Debug(net.SplitHostPort(hc.ReqInfo.RemoteAddr))
	hc.ResponseString(http.StatusOK, "hello tlnet")
}

func staticHandleFunc(hc *HttpContext) {
	logging.Debug(hc.ReqInfo)
	logging.Debug(hc.GetParam("name"))
	logging.Debug(net.SplitHostPort(hc.ReqInfo.RemoteAddr))
	// hc.Redirect("https://baidu.com")
}

func notFoundFilter() *Filter {
	f := NewFilter()
	f.AddNotFoundPageIntercept(notFound)
	return f
}

func aaa(w ResponseWriter, r *Request) {
	logging.Debug("aaa")
	logging.Debug(fmt.Sprint(r.Header))
	io.WriteString(w, "hello aaa 你访问成功了")
}

func notFound(w ResponseWriter, r *Request) bool {
	logging.Debug("notFound")
	logging.Debug(fmt.Sprint(r.Header))
	io.WriteString(w, "not found 404")
	return true
}

//后缀为.html的过滤器
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

func suffixIntercept(w ResponseWriter, r *Request) bool {
	io.WriteString(w, "html is not allowed")
	return true
}

func staticFilter() *Filter {
	f := NewFilter()
	f.AddNotFoundPageIntercept(permission)
	f.AddGlobalIntercept("[ab]", globalIntercept)
	return f
}

func globalIntercept(w ResponseWriter, r *Request) bool {
	io.WriteString(w, "globalIntercept")
	return true
}

func permission(w ResponseWriter, r *Request) bool {
	logging.Debug("permission")
	logging.Debug(fmt.Sprint(r.Header))
	err := r.Body.Close()
	logging.Debug(err)
	return true
}

func notify(hc *HttpContext) {
	for _, v := range wsmap {
		v.WS.Send(fmt.Sprint("通知：", time.Now()))
	}
}

var wsmap = make(map[int64]*HttpContext, 0)

func websocketFunc(hc *HttpContext) {
	_, ok := wsmap[hc.WS.Id]
	if !ok {
		wsmap[hc.WS.Id] = hc
	}
	logging.Debug(hc.ReqInfo)
	logging.Debug(hc.ReqInfo.RemoteAddr)
	msg := string(hc.WS.Read())
	logging.Debug("收到:", msg)
	// hc.WS.Send([]byte("好了"))
	// hc.WS.Send("你发送的是：" + string(hc.WS.Read()))

	for k, v := range wsmap {
		if k != hc.WS.Id {
			v.WS.Send(fmt.Sprint("这是", hc.WS.Id, "发送的信息:", msg))
		}
	}

}
