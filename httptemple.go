package tlnet

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/donnie4w/simplelog/logging"
	. "github.com/donnie4w/tlnet/db"
)

func OpenView(port int32) {
	go data_view(port)
}

func data_view(port int32) {
	myRecover()
	tl := NewTlnet()
	tl.Handle("/", search)
	tl.Handle("/req", req)
	logging.Debug("open data view server :", port)
	err := tl.HttpStart(port)
	logging.Error("err:", err.Error())
}

func search(hc *HttpContext) {
	var keys []string
	keys, _ = SimpleDB().GetKeys()
	searchId := hc.PostParam("searchId")
	_showLine := hc.PostParam("showLine")
	_pageNumber := hc.PostParam("pageNumber")
	logging.Debug("searchId:", searchId)
	logging.Debug("showLine:", _showLine)
	logging.Debug("pageNumber:", _pageNumber)
	if _showLine == "" {
		_showLine = "100"
	}
	if _pageNumber == "" {
		_pageNumber = "0"
	}
	if searchId == "" {
		keys, _ = SimpleDB().GetKeys()
	} else {
		keys, _ = SimpleDB().GetKeysPrefix([]byte(searchId))
	}
	showLine, _ := strconv.ParseInt(_showLine, 10, 0)
	pageNumber, _ := strconv.ParseInt(_pageNumber, 10, 0)
	s := "{"
	s = fmt.Sprint(s, `"searchId":`, `"`, searchId, `",`)
	s = fmt.Sprint(s, `"showLine":`, `"`, showLine, `",`)
	s = fmt.Sprint(s, `"pageNumber":`, `"`, pageNumber, `",`)
	s = fmt.Sprint(s, `"totalcount":`, `"`, len(keys), `",`)
	s = fmt.Sprint(s, `"list":[`)
	max := len(keys)
	if max > int(showLine*(pageNumber+1)) {
		max = int(showLine * (pageNumber + 1))
	}
	for i := int(showLine * pageNumber); i < max; i++ {
		// value, _ := SimpleDB().GetString([]byte(keys[i]))
		// logging.Debug("key", i+1, ":", keys[i], "===>", value)
		s = fmt.Sprint(s, `{"k":"`, keys[i], `","v":"`, "", `"}`)
		if i < max-1 {
			s = fmt.Sprint(s, ",")
		}
	}
	s = fmt.Sprint(s, "]}")
	// logging.Debug(s)
	htmlstring := loadfile("dataview.html")
	hc.ResponseString(0, strings.ReplaceAll(htmlstring, "#####", s))
}

func req(hc *HttpContext) {
	key := hc.PostParam("key")
	value, _ := SimpleDB().GetString([]byte(key))
	logging.Debug("req:", key, " = ", value)
	hc.ResponseString(0, value)
}

func loadfile(s string) (_r string) {
	bs, err := os.ReadFile(s)
	if err == nil {
		_r = string(bs)
	}
	return
}
