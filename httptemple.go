package tlnet

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

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
	tl.Handle("/del", del)
	tl.Handle("/backup", backup)
	tl.Handle("/load", load)
	logging.Debug("open data view server :", port)
	err := tl.HttpStart(port)
	logging.Error("err:", err.Error())
}

func search(hc *HttpContext) {
	var keys []string
	keys, _ = SimpleDB().GetKeys()
	searchId := strings.TrimSpace(hc.PostParam("searchId"))
	_showLine := hc.PostParam("showLine")
	transfer := hc.PostParam("transfer")
	_pageNumber := hc.PostParam("pageNumber")
	// logging.Debug("searchId:", searchId)
	// logging.Debug("showLine:", _showLine)
	// logging.Debug("pageNumber:", _pageNumber)
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
	showLine, _ := strconv.Atoi(_showLine)
	pageNumber, _ := strconv.Atoi(_pageNumber)
	s := "{"
	s = fmt.Sprint(s, `"searchId":`, `"`, searchId, `",`)
	s = fmt.Sprint(s, `"transfer":`, `"`, transfer, `",`)
	s = fmt.Sprint(s, `"showLine":`, `"`, showLine, `",`)
	s = fmt.Sprint(s, `"pageNumber":`, `"`, pageNumber, `",`)
	s = fmt.Sprint(s, `"totalcount":`, `"`, len(keys), `",`)
	s = fmt.Sprint(s, `"list":[`)
	max := len(keys)
	if max > showLine*(pageNumber+1) {
		max = showLine * (pageNumber + 1)
	}
	for i := showLine * pageNumber; i < max; i++ {
		// value, _ := SimpleDB().GetString([]byte(keys[i]))
		// logging.Debug("key", i+1, ":", keys[i], "===>", value)
		s = fmt.Sprint(s, `{"k":"`, keys[i], `","v":"`, "", `"}`)
		if i < max-1 {
			s = fmt.Sprint(s, ",")
		}
	}
	s = fmt.Sprint(s, "]}")
	htmlstring := loadfile("dataview.html")
	hc.ResponseString(0, strings.ReplaceAll(htmlstring, "#####", s))
}

func req(hc *HttpContext) {
	key := hc.PostParam("key")
	value, _ := SimpleDB().GetString([]byte(key))
	hc.ResponseString(0, value)
}

func del(hc *HttpContext) {
	key := hc.PostParam("key")
	if strings.HasPrefix(key, "0_") || strings.HasPrefix(key, "1_") || strings.HasPrefix(key, "pte_") || strings.HasPrefix(key, "idx_") {
		DeleteWithKey(key)
	} else {
		SimpleDB().Del([]byte(key))
	}
}

func backup(hc *HttpContext) {
	filename := fmt.Sprint("backup_", time.Now().Format("2006-01-02"), ".db")
	SimpleDB().BackupToDisk(filename, nil)
	file, err := os.Open(filename)
	if err == nil {
		defer func() {
			file.Close()
			os.Remove(filename)
		}()
		fileStat, _ := file.Stat()
		hc.Writer().Header().Set("Content-Disposition", "attachment; filename="+filename)
		hc.Writer().Header().Set("Content-Type", "application/octet-stream")
		hc.Writer().Header().Set("Content-Length", fmt.Sprint(fileStat.Size()))
		io.Copy(hc.Writer(), file)
	} else {
		hc.ResponseString(0, ackhtmlString("数据导出失败！"))
	}
}

func load(hc *HttpContext) {
	f, _, e := hc.FormFile("loadfile")
	if e == nil {
		var buf bytes.Buffer
		io.Copy(&buf, f)
		e = SimpleDB().LoadBytes(buf.Bytes())
	}
	if e == nil {
		hc.ResponseString(0, ackhtmlString("文件导入成功！"))
	} else {
		hc.ResponseString(0, ackhtmlString("文件导入失败！"))
	}

}

func loadfile(s string) (_r string) {
	bs, err := os.ReadFile(s)
	if err == nil {
		_r = string(bs)
	}
	return
}

func ackhtmlString(s string) string {
	return fmt.Sprint(`<html><body><p style="color: crimson;margin: 10px;">`, s, `</p> <a href="/">返回主页</a><body></html>`)
}

var htmlstring = `
<html>

<head>
    <title>tlnet data view</title>
</head>

<body style="background-color: beige;">
    <div style="margin-top: 20px;margin-left: 10px;">
        <input type="text" id="_searchId" style="width: 250px;height: 25px;" placeholder="key前缀" />&nbsp;
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
    <div style="margin: 10px; display:inline">
        <button onclick="backup()">导出数据到文件</button>
        <form id="backupForm" action="/backup" method="post" enctype="multipart/form-data">
        </form>
        <button onclick="load()" style="display:inline">文件数据导入</button>
        <form id="loadForm" action="/load" method="post" enctype="multipart/form-data" style="display:inline">
            <input type="file" id="loadfile" name="loadfile" style="width: 250px;height: 25px;">
        </form>
    </div>

    <hr>
    <div id="hint" style="color: crimson;margin: 10px;"></div>
    <table border="1" id="content">
    </table>
    <form id="formId" action="/" method="post">
        <input type="hidden" id="data" value='#####' />
        <input type="hidden" id="pageNumber" name="pageNumber" value="0" />
        <input type="hidden" id="searchId" name="searchId" />
        <input type="hidden" id="showLine" name="showLine" />
        <input type="hidden" id="transfer" name="transfer" />
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
        document.getElementById("hint").innerText = jsonObj.transfer;
        document.getElementById("currpage").innerText = parseInt(jsonObj.pageNumber) + 1;
        var content = document.getElementById("content");
        content.innerHTML = '<tr><th>key</th><th>value</th><th></th><th></th></tr>';
        for (var i = 0; i < list.length; i++) {
            let tr = document.createElement('tr');
            tr.id = "tr" + i;
            var key = list[i].k;
            tr.innerHTML = '<td>' + key + "</td>"
            let td2 = document.createElement('td');
            td2.innerHTML = '<textarea></textarea>';
            tr.appendChild(td2);
            let td3 = document.createElement('td');
            td3.innerHTML = '<button onclick="req(this)">获取value值</button>'
            tr.appendChild(td3);
            if (key.startsWith("0_") || (!key.startsWith("1_") && !key.startsWith("pte_") && !key.startsWith("idx_"))) {
                let td4 = document.createElement('td');
                td4.innerHTML = '<button onclick="del(this)">删除TableKey</button>'
                tr.appendChild(td4);
            }
            content.appendChild(tr);
        }
    }
    dataview();
    function search(page) {
        var search = document.getElementById("_searchId").value;
        document.getElementById("searchId").value = search;
        document.getElementById("showLine").value = document.getElementById("_showLine").value;
        var pageNumber = parseInt(document.getElementById("pageNumber").value) + page;
        if (pageNumber >= 0) {
            document.getElementById("pageNumber").value = pageNumber
            var form_ = document.getElementById("formId");
            form_.submit();
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

    function del(obj) {
        var xmlHttp = new XMLHttpRequest();
        var key = obj.parentNode.parentNode.cells[0].innerText;
        xmlHttp.onreadystatechange = function () {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200) {
                document.getElementById("transfer").value = "成功删除key: " + key;
                search(0);
            }
        }
        xmlHttp.open("POST", "/del", true);
        xmlHttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
        xmlHttp.send("key=" + key);
    }

    function backup() {
        var msg = "确定要导出所有数据吗？\n\n请确认！";
        if (confirm(msg) == true) {
            document.getElementById('backupForm').submit();
            return true;
        } else {
            return false;
        }
    }

    function load() {
        var file = document.getElementById('loadfile').files[0];
        if (!isEmpty(file)) {
            document.getElementById('loadForm').submit();
        } else {
            alert("未选择数据文件");
        }
    }

    function isEmpty(obj) {
        if (typeof obj == "undefined" || obj == null || obj == "") {
            return true;
        } else {
            return false;
        }
    }
</script>

</html>
`
