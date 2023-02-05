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
	// htmlstring := loadfile("dataview.html")
	hc.ResponseString(0, strings.ReplaceAll(htmlstring, "#####", s))
}

func req(hc *HttpContext) {
	key := hc.PostParam("key")
	value, _ := SimpleDB().GetString([]byte(key))
	// logging.Debug("req:", key, " = ", value)
	hc.ResponseString(0, value)
}

func loadfile(s string) (_r string) {
	bs, err := os.ReadFile(s)
	if err == nil {
		_r = string(bs)
	}
	return
}

var htmlstring = `
<html>

<head>
    <title>tlnet data view</title>
</head>

<body style="background-color: beige;">
    <div style="margin-top: 20px;margin-left: 10px;">
        <input type="text" id="_searchId" style="width: 250px;height: 25px;" />&nbsp;
        <button onclick="search(0)">搜索key前缀</button>
        <br style="margin: 100px;">
    </div>
    <div style="margin: 10px;">
        <button onclick="search(-1)">上一页</button>
        <button onclick="search(1)">下一页</button>
        当前页码:<span id="currpage" style="height: 25px;"></span>
        &nbsp;每页显示:<input type="text" id="_showLine" value="100" style="width: 100px;height: 25px;" />
        总数:<span id="totalcount" style="height: 25px;"></span>
    </div>
    <hr>
    <table border="1" id="content">
        <tr>
            <th>key</th>
            <th>value</th>
            <th></th>
        </tr>
    </table>
    <form id="formId" action="/" method="post">
        <input type="hidden" id="data" value='#####' />
        <input type="hidden" id="pageNumber" name="pageNumber" value="0" />
        <input type="hidden" id="searchId" name="searchId" />
        <input type="hidden" id="showLine" name="showLine" />
    </form>
</body>
<script>
    function dataview() {
        var data = document.getElementById("data").value;
        let jsonObj = JSON.parse(data);
        var list = jsonObj.list;
        document.getElementById("_searchId").value = jsonObj.searchId;
        document.getElementById("_showLine").value = jsonObj.showLine;
        document.getElementById("pageNumber").value = jsonObj.pageNumber;
        document.getElementById("totalcount").innerText = jsonObj.totalcount;
        document.getElementById("currpage").innerText = parseInt(jsonObj.pageNumber) + 1;
        for (var i = 0; i < list.length; i++) {
            let tr = document.createElement('tr');
            tr.id = "tr" + i;
            tr.innerHTML = '<td>' + list[i].k + "</td>"
            let td2 = document.createElement('td');
            td2.innerHTML = '<textarea>' + list[i].v + '</textarea>';
            tr.appendChild(td2);
            let td3 = document.createElement('td');
            td3.innerHTML = '<button onclick="req(this)">获取value值</button>'
            tr.appendChild(td3);
            document.getElementById("content").appendChild(tr);
        }
    }
    dataview();
    function search(page) {
        var search = document.getElementById("_searchId").value;
        if (search == "") {
            alert("搜索key为空");
        } else {
            document.getElementById("searchId").value = search;
            document.getElementById("showLine").value = document.getElementById("_showLine").value;
            var pageNumber = parseInt(document.getElementById("pageNumber").value) + page;
            if (pageNumber >= 0) {
                document.getElementById("pageNumber").value = pageNumber
                var form_ = document.getElementById("formId");
                form_.submit();
            }
        }
    }
    function req(obj) {
        var xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = function () {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200) {
                var res = xmlHttp.responseText;
                var tr = obj.parentNode.parentNode;
                tr.cells[1].childNodes[0].innerText = res;
            }
        }
        xmlHttp.open("POST", "/req", true);
        xmlHttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
        xmlHttp.send("key=" + obj.parentNode.parentNode.cells[0].innerText);
    }
</script>

</html>
`
