### tlnet  轻量级高性能http服务框架 [[English document]](https://github.com/donnie4w/tlnet/blob/main/README.md "[English document]")

### 功能说明

1. **易用性**：使用与内置http的使用风格相似，但更加简易
2. **高效性**：基于缓存与树的路由算法，提供了极高的性能
3. **基础路由**：支持常见的HTTP方法（GET,POST,PUT,DELETE,PATCH等）
4. **拦截器**：支持拦截器，包括正则匹配拦截url等拦截功能
5. **websocket**：支持websocket,使用websocket服务与普通http服务同样简易
6. **静态文件服务**：支持设置路径来提供静态文件服务，静态服务同样支持拦截器
7. **restful**：支持构建RESTful API服务，如  /path/:id/:name/order
8. **通配符**：支持使用通配符来定义路由，如 /path/*/order
9. **thrift**：支持thrift http协议服务
10. **轻量级**：启动快，极少资源占用，无冗余功能

### [Github 地址](http://github.com/donnie4w/tlnet "Github 地址")


### 适用场景
1. 若你寻找一个比`gin`，`echo`或原生http等更高效更轻量的http框架，`tlnet`将是一个很好的选择，它提供基础性http服务功能，具备良好的扩展性，稳定性，与高并发支持。由于`tlnet`的基于缓存与树的实现，绝大多数情况下路由时间复杂度为`O(1)`,具备比树更高效的性能。
2. `tlnet` 非常适合构建高性能、轻量级的 Web应用程序和 API 或微服务
3. `tlnet`没有内置完整的全栈开发的功能支持；对于需要大量内置功能、复杂业务逻辑、全栈开发支持需求的场景，可能`Beego`或者`Revel`等框架的功能更全面。

------

# 快速使用

### 安装

```bash
go get github.com/donnie4w/tlnet
```

####  示例1：
```go
tl := NewTlnet()
tl.GET("/g", getHandle)
tl.HttpStart(":8000")  

func getHandle(hc *HttpContext) {
    hc.ResponseString("hello tlnet") 
}

```


#### 示例2：
```go
tl := NewTlnet()
tl.POST("/p", postHandle)   //post请求
tl.HandleStatic("/js/", "/html/js", nil) //设置静态资源服务
tl.HttpStart(":8000")  

func postHandle(hc *HttpContext) {
    hc.ResponseString("hello tlnet") 
}
```

#### 示例3：
```go
tl := NewTlnet()
tl.GET("/user/:id/:name/order", userorderhandler)   //restful API
tl.HttpStart(":8000")  

func userorderhandler(hc *HttpContext) {
    hc.ResponseString(hc.GetParam("id") + ":" + hc.GetParam("name"))
}
```


### 拦截器使用用例：

```go
tl := NewTlnet()
tl.HandleWithFilter("/",interceptFilter(),func(hc *HttpContext) {hc.ResponseString("uri:"+hc.ReqInfo.Uri)) })   //http通用处理并定义拦截器
tl.HandleStaticWithFilter("/js/", "/html/js", interceptFilter(), nil)  
tl.HttpStart(":8000")

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
```

### websocket的基础使用：

```go
func TestWebsocket(t *testing.T) {
    tl := NewTlnet()
    tl.HandleWebSocket("/ws", websocketHandler)
    tl.HttpStart(":8000")
}
func websocketHandler(hc *HttpContext) {
    wsId := hc.WS.Id //每个websocket链接都会生成一个识别id
    msg := string(hc.WS.Read())                                 //读取客户端请求数据
    hc.WS.Send("服务ID" + fmt.Sprint(wsId) + "已经收到你发送的信息:" + msg) //向客户端回复数据
}
```
	

### tlnet的其他函数 使用示例说明：

```go
SetLogger(true) // 开启日志打印，默认关闭

tl := NewTlnet() //  创建Tlnet示例对象

tl.SetMaxBytesReader(1 << 20)  //设置http请求数据 最大最大限制 1M

tl.HandleWebSocketBindOrigin("/ws", "http://tlnet.top/", websocketFunc)  //定义Origin为 http://tlnet.top/

tl.HandleWebSocketBindConfig("/ws", websocketFunc, newWebsocketConfig()) //定义 config，实现websocket的错误处理，连接成功 等函数处理

tl.ReadTimeout(10 * time.Second) //设置读取超时

tl.HttpsStart(":8000",certFile, keyFile)  // 启动 https 服务  certFile为crt文件路径，keyFile为key文件路径

tl.HttpsStartWithBytes(":8000",certFileBs, keyFileBs) // 启动 https 服务  certFileBs为crt文件数据，keyFileBs为key文件数据

tl.Close()  //关闭tlnet实例对象服务
```

------------

### Tlnet 性能

#### `tlnet` + `原生http` + `gin` + `echo`

[《Tlnet—性能压测数据 tlnet+gin+echo+原生http》](https://tlnet.top/article/22425190)

[压测程序地址](https://github.com/donnie4w/test/tree/main/httpserver)


##### 已注册URL处理器，即能找到URL的场景
```text
pkg: test/httpserver
cpu: Intel(R) Core(TM) i5-1035G1 CPU @ 1.00GHz
BenchmarkGin
BenchmarkGin-4           2356797               507.6 ns/op          1040 B/op          9 allocs/op
BenchmarkGin-8           2428094               500.5 ns/op          1040 B/op          9 allocs/op
BenchmarkHttp
BenchmarkHttp-4          3405583               346.5 ns/op           344 B/op          9 allocs/op
BenchmarkHttp-8          3551180               330.7 ns/op           344 B/op          9 allocs/op
BenchmarkTlnet
BenchmarkTlnet-4         6224007               187.6 ns/op           288 B/op          5 allocs/op
BenchmarkTlnet-8         6570204               184.3 ns/op           288 B/op          5 allocs/op
BenchmarkEcho
BenchmarkEcho-4          2333074               535.3 ns/op          1024 B/op         10 allocs/op
BenchmarkEcho-8          2366791               528.6 ns/op          1024 B/op         10 allocs/op
PASS
```

```text
goarch: amd64
pkg: test/httpserver
cpu: Intel(R) Core(TM) i5-1035G1 CPU @ 1.00GHz
BenchmarkGin
BenchmarkGin-4           2290322               494.4 ns/op          1040 B/op          9 allocs/op
BenchmarkGin-8           2491639               487.8 ns/op          1040 B/op          9 allocs/op
BenchmarkGin-16          2560693               493.1 ns/op          1041 B/op          9 allocs/op
BenchmarkHttp
BenchmarkHttp-4          3565180               323.4 ns/op           344 B/op          9 allocs/op
BenchmarkHttp-8          3544458               339.9 ns/op           344 B/op          9 allocs/op
BenchmarkHttp-16         3596947               307.0 ns/op           344 B/op          9 allocs/op
BenchmarkTlnet
BenchmarkTlnet-4         5572468               189.7 ns/op           288 B/op          5 allocs/op
BenchmarkTlnet-8         6420810               189.5 ns/op           288 B/op          5 allocs/op
BenchmarkTlnet-16        5862798               197.4 ns/op           288 B/op          5 allocs/op
BenchmarkEcho
BenchmarkEcho-4          2356166               504.5 ns/op          1024 B/op         10 allocs/op
BenchmarkEcho-8          2238975               540.4 ns/op          1024 B/op         10 allocs/op
BenchmarkEcho-16         1740646               639.0 ns/op          1024 B/op         10 allocs/op
PASS
```

```text
goarch: amd64
pkg: test/httpserver
cpu: Intel(R) Core(TM) i5-1035G1 CPU @ 1.00GHz
BenchmarkGin
BenchmarkGin-4           2445027               484.4 ns/op          1040 B/op          9 allocs/op
BenchmarkGin-8           2465703               489.2 ns/op          1040 B/op          9 allocs/op
BenchmarkGin-16          2567120               462.3 ns/op          1040 B/op          9 allocs/op
BenchmarkHttp
BenchmarkHttp-4          3700851               332.5 ns/op           344 B/op          9 allocs/op
BenchmarkHttp-8          3507562               385.1 ns/op           344 B/op          9 allocs/op
BenchmarkHttp-16         3458373               318.5 ns/op           344 B/op          9 allocs/op
BenchmarkTlnet
BenchmarkTlnet-4         7283329               166.9 ns/op           288 B/op          6 allocs/op
BenchmarkTlnet-8         7045941               163.3 ns/op           288 B/op          6 allocs/op
BenchmarkTlnet-16        7091187               162.7 ns/op           288 B/op          6 allocs/op
BenchmarkEcho
BenchmarkEcho-4          2395107               501.3 ns/op          1024 B/op         10 allocs/op
BenchmarkEcho-8          2230350               541.2 ns/op          1024 B/op         10 allocs/op
BenchmarkEcho-16         2118523               571.2 ns/op          1025 B/op         10 allocs/op
```
------

##### 为注册URL处理器，即能找到URL的场景

```text
goarch: amd64
pkg: test/httpserver
cpu: Intel(R) Core(TM) i5-1035G1 CPU @ 1.00GHz
BenchmarkGin
BenchmarkGin-4           2695962               436.6 ns/op           992 B/op          8 allocs/op
BenchmarkGin-8           2682855               448.6 ns/op           992 B/op          8 allocs/op
BenchmarkGin-16          2692042               456.7 ns/op           993 B/op          8 allocs/op
BenchmarkHttp
BenchmarkHttp-4          1462479               806.5 ns/op          1200 B/op         19 allocs/op
BenchmarkHttp-8          1378512               868.7 ns/op          1200 B/op         19 allocs/op
BenchmarkHttp-16         1359558               898.3 ns/op          1201 B/op         19 allocs/op
BenchmarkTlnet
BenchmarkTlnet-4         6343191               182.1 ns/op           288 B/op          5 allocs/op
BenchmarkTlnet-8         6371780               193.3 ns/op           288 B/op          5 allocs/op
BenchmarkTlnet-16        6292598               182.6 ns/op           288 B/op          5 allocs/op
BenchmarkEcho
BenchmarkEcho-4          1387812               854.2 ns/op          1489 B/op         16 allocs/op
BenchmarkEcho-8          1000000              1031 ns/op            1489 B/op         16 allocs/op
BenchmarkEcho-16          871554              1220 ns/op            1491 B/op         16 allocs/op
```

```text
goarch: amd64
pkg: test/httpserver
cpu: Intel(R) Core(TM) i5-1035G1 CPU @ 1.00GHz
BenchmarkGin
BenchmarkGin-4           2695962               436.6 ns/op           992 B/op          8 allocs/op
BenchmarkGin-8           2682855               448.6 ns/op           992 B/op          8 allocs/op
BenchmarkGin-16          2692042               456.7 ns/op           993 B/op          8 allocs/op
BenchmarkHttp
BenchmarkHttp-4          1462479               806.5 ns/op           1200 B/op         19 allocs/op
BenchmarkHttp-8          1378512               868.7 ns/op           1200 B/op         19 allocs/op
BenchmarkHttp-16         1359558               898.3 ns/op           1201 B/op         19 allocs/op
BenchmarkTlnet
BenchmarkTlnet-4         6343191               182.1 ns/op           288 B/op          5 allocs/op
BenchmarkTlnet-8         6371780               193.3 ns/op           288 B/op          5 allocs/op
BenchmarkTlnet-16        6292598               182.6 ns/op           288 B/op          5 allocs/op
BenchmarkEcho
BenchmarkEcho-4          1387812               854.2 ns/op           1489 B/op         16 allocs/op
BenchmarkEcho-8          1000000               1031 ns/op            1489 B/op         16 allocs/op
BenchmarkEcho-16          871554               1220 ns/op            1491 B/op         16 allocs/op
```

```text
goarch: amd64
pkg: test/httpserver
cpu: Intel(R) Core(TM) i5-1035G1 CPU @ 1.00GHz
BenchmarkGin
BenchmarkGin-4           2758132               436.3 ns/op           992 B/op          8 allocs/op
BenchmarkGin-8           2667070               447.5 ns/op           992 B/op          8 allocs/op
BenchmarkGin-16          2778464               423.7 ns/op           992 B/op          8 allocs/op
BenchmarkHttp
BenchmarkHttp-4          1484661               809.1 ns/op          1200 B/op         19 allocs/op
BenchmarkHttp-8          1470441               831.8 ns/op          1200 B/op         19 allocs/op
BenchmarkHttp-16         1466318               841.5 ns/op          1201 B/op         19 allocs/op
BenchmarkTlnet
BenchmarkTlnet-4         6568359               178.0 ns/op           288 B/op          5 allocs/op
BenchmarkTlnet-8         6523093               185.5 ns/op           288 B/op          5 allocs/op
BenchmarkTlnet-16        6733190               172.9 ns/op           288 B/op          5 allocs/op
BenchmarkEcho
BenchmarkEcho-4          1412888               828.2 ns/op          1488 B/op         16 allocs/op
BenchmarkEcho-8          1404442               892.6 ns/op          1489 B/op         16 allocs/op
BenchmarkEcho-16         1306354               932.2 ns/op          1491 B/op         16 allocs/op
```

```text
goarch: amd64
pkg: test/httpserver
cpu: Intel(R) Core(TM) i5-1035G1 CPU @ 1.00GHz
BenchmarkGin
BenchmarkGin-4           2715367               441.7 ns/op           992 B/op          8 allocs/op
BenchmarkGin-8           2552917               463.5 ns/op           992 B/op          8 allocs/op
BenchmarkGin-16          2607180               448.3 ns/op           993 B/op          8 allocs/op
BenchmarkHttp
BenchmarkHttp-4          1500954               808.5 ns/op          1200 B/op         19 allocs/op
BenchmarkHttp-8          1417261               861.2 ns/op          1200 B/op         19 allocs/op
BenchmarkHttp-16         1334204               899.6 ns/op          1201 B/op         19 allocs/op
BenchmarkTlnet
BenchmarkTlnet-4         6459314               182.8 ns/op           288 B/op          5 allocs/op
BenchmarkTlnet-8         5938984               195.5 ns/op           288 B/op          5 allocs/op
BenchmarkTlnet-16        6510511               191.6 ns/op           288 B/op          5 allocs/op
BenchmarkEcho
BenchmarkEcho-4          1414153               840.8 ns/op          1489 B/op         16 allocs/op
BenchmarkEcho-8          1344892               896.2 ns/op          1489 B/op         16 allocs/op
BenchmarkEcho-16         1244558              1139 ns/op            1490 B/op         16 allocs/op
```


