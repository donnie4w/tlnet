package tlnet

import (
	"fmt"
	"io"
	"net"
	"net/http"
	. "net/http"
	"testing"

	"github.com/donnie4w/simplelog/logging"
)

type TestObj struct {
	Id   int64
	Name string
	Age_ int32
}

func Test_DB(t *testing.T) {
	InitDB("test.db")
	err := Insert(&TestObj{Name: "wuxiaodong", Age_: 215})
	fmt.Println("————————————————————————————————————————————", err)
	err = Update(&TestObj{3, "aaaaaa", 111})
	Delete(TestObj{Id: 3})
	// Delete(&TestObj{Id: 2})
	ts := Selects[TestObj](0, 10)
	for i, v := range ts {
		logging.Debug(i+1, "----", v)
	}
	logging.Debug("max idx==>", GetIdSeqValue[TestObj]())

	fmt.Println("------------------------------------------------")
	ts = SelectByIdxName[TestObj]("age", "111")
	for i, v := range ts {
		logging.Debug(i+1, "=====", v)
	}
	fmt.Println("------------------------------------------------")
	ts = SelectByIdxNameLimit[TestObj]("age", []string{"215", "216", "333"}, 0, 2)
	for i, v := range ts {
		logging.Debug(i+1, "=========>", v)
	}
	fmt.Println("")
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	IterDB()
}

func IterDB() {
	keys, _ := SingleDB().GetKeys()
	for i, v := range keys {
		logging.Debug("key", i, "==", v)
		value, _ := SingleDB().GetString([]byte(v))
		logging.Debug(v, "==>", value)
	}
}

func _Test_snap(t *testing.T) {
	InitDB("test.db")
	SingleDB().Put([]byte("d"), []byte("3"))
	logging.Debug(SingleDB().GetKeys())
	er := SingleDB().BackupToDisk("snap.lb", []byte("d"))
	logging.Debug(er)
	logging.Debug(RecoverBackup("snap.lb"))
	for _, v := range RecoverBackup("snap.lb") {
		logging.Debug(string(v.Key), " == ", string(v.Value))
	}
}

func _Test_tlnet(t *testing.T) {
	tlnet := NewTlnet()
	// tlnet.DBPath("test.db")
	tlnet.SetMaxBytesReader((1 << 20) * 50)
	tlnet.AddHandlerFunc("/aaa", nil, aaa)
	tlnet.AddHandlerFunc("/bbb", notFoundFilter(), aaa)
	tlnet.AddProcessor("/ppp", nil)
	tlnet.AddStaticHandler("/", "test.db", staticFilter(), nil)
	tlnet.HttpStart(8080)
}

func _Test_tlnet2(t *testing.T) {
	tlnet := NewTlnet()
	// tlnet.DBPath("test.db")
	tlnet.SetMaxBytesReader((1 << 20) * 50)
	tlnet.Handle("/qq", handleFunc)
	tlnet.StaticDir("/s", "test.db", staticHandleFunc)
	tlnet.HttpStart(8080)
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
