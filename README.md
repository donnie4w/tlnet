### tlnet   Lightweight high-performance http service framework [[中文文档]](https://github.com/donnie4w/tlnet/blob/main/README_zh.md "[中文文档]")

### Function Description

1. **Ease of Use**: Similar in style to the built-in `http`, but simpler to use.
2. **Efficiency**: Provides extremely high performance thanks to a routing algorithm based on caching and trees.
3. **Basic Routing**: Supports common HTTP methods (GET, POST, PUT, DELETE, PATCH, etc.).
4. **Interceptors**: Supports interceptors, including URL interception based on regex matching.
5. **WebSocket**: Supports WebSocket with the same simplicity as regular HTTP services.
6. **Static File Service**: Allows setting paths to serve static files, with interceptor support for static services.
7. **RESTful**: Supports building RESTful API services, such as `/path/:id/:name/order`.
8. **Wildcards**: Supports using wildcards to define routes, such as `/path/*/order`.
9. **Thrift**: Supports Thrift HTTP protocol services.
10. **Lightweight**: Quick to start, minimal resource usage, no redundant features.

### [GitHub Repository](http://github.com/donnie4w/tlnet "GitHub Repository")

### Applicable Scenarios

1. If you are looking for a more efficient and lightweight HTTP framework compared to `gin`, `echo`, or native `http`, `tlnet` is a great choice. It provides fundamental HTTP service functionality with good scalability, stability, and high concurrency support. Due to `tlnet`’s cache and tree-based implementation, routing has an average time complexity of `O(1)`, offering better performance than traditional tree structures.
2. `tlnet` is well-suited for building high-performance, lightweight web applications, APIs, or microservices.
3. `tlnet` does not include full-stack development features. For scenarios that require a lot of built-in features, complex business logic, and full-stack development support, frameworks like `Beego` or `Revel` might offer more comprehensive functionality.

------

# Quick Start

### Installation

```bash
go get github.com/donnie4w/tlnet
```

#### Example 1:

```go
tl := NewTlnet()
tl.GET("/g", getHandle)
tl.HttpStart(":8000")

func getHandle(hc *HttpContext) {
    hc.ResponseString("hello tlnet")
}
```

#### Example 2:

```go
tl := NewTlnet()
tl.POST("/p", postHandle)   // POST request
tl.HandleStatic("/js/", "/html/js", nil) // Set up static resource service
tl.HttpStart(":8000")

func postHandle(hc *HttpContext) {
    hc.ResponseString("hello tlnet")
}
```

#### Example 3:

```go
tl := NewTlnet()
tl.GET("/user/:id/:name/order", userOrderHandler)   // RESTful API
tl.HttpStart(":8000")

func userOrderHandler(hc *HttpContext) {
    hc.ResponseString(hc.GetParam("id") + ":" + hc.GetParam("name"))
}
```

### Interceptor Use Case:

```go
tl := NewTlnet()
tl.HandleWithFilter("/", interceptFilter(), func(hc *HttpContext) {hc.ResponseString("uri:"+hc.ReqInfo.Uri)})  // General HTTP handling with an interceptor
tl.HandleStaticWithFilter("/js/", "/html/js", interceptFilter(), nil)
tl.HttpStart(":8000")

func interceptFilter() *Filter {
    f := NewFilter()
    // Built-in suffix interceptor
    f.AddSuffixIntercept([]string{".jpg"}, func(hc *HttpContext) bool {
        logging.Debug("suffix path:" + hc.ReqInfo.Path)
        return false
    })
    // Built-in "page not found" interceptor
    f.AddPageNotFoundIntercept(func(hc *HttpContext) bool {
        hc.Error("page not found:" + hc.ReqInfo.Uri, http.StatusNotFound)
        return true
    })
    // Custom URL regex matching interceptor
    f.AddIntercept(".*?", func(hc *HttpContext) bool {
        logging.Debug("intercept:", hc.ReqInfo.Uri)
        if hc.ReqInfo.Path == "" {
            hc.Error("path is empty:" + hc.ReqInfo.Uri, http.StatusForbidden)
            return true
        }
        return false
    })
    return f
}
```

### Basic WebSocket Usage:

```go
func TestWebsocket(t *testing.T) {
    tl := NewTlnet()
    tl.HandleWebSocket("/ws", websocketHandler)
    tl.HttpStart(":8000")
}

func websocketHandler(hc *HttpContext) {
    wsId := hc.WS.Id // Each WebSocket connection generates a unique ID
    msg := string(hc.WS.Read()) // Read client request data
    hc.WS.Send("Service ID " + fmt.Sprint(wsId) + " received your message: " + msg) // Respond to client
}
```

### Additional `tlnet` Functionality Examples:

```go
SetLogger(true) // Enable logging, disabled by default

tl := NewTlnet() // Create a Tlnet instance

tl.SetMaxBytesReader(1 << 20)  // Set the maximum HTTP request data limit to 1 MB

tl.HandleWebSocketBindOrigin("/ws", "http://tlnet.top/", websocketFunc)  // Define Origin as http://tlnet.top/

tl.HandleWebSocketBindConfig("/ws", websocketFunc, newWebsocketConfig()) // Define config for WebSocket error handling, connection success, etc.

tl.ReadTimeout(10 * time.Second) // Set read timeout

tl.HttpsStart(":8000", certFile, keyFile)  // Start HTTPS service with certFile (certificate path) and keyFile (key path)

tl.HttpsStartWithBytes(":8000", certFileBs, keyFileBs) // Start HTTPS service with certFileBs (certificate data) and keyFileBs (key data)

tl.Close()  // Close the Tlnet instance service
```

------------

### Tlnet Performance

#### `tlnet` vs. `native http` vs. `gin` vs. `echo`

[《Tlnet Performance Benchmark: tlnet vs gin vs echo vs native http》](https://tlnet.top/article/22425190)

[Benchmark Program Repository](https://github.com/donnie4w/test/tree/main/httpserver)


##### URL processor has been registered, that is, the URL can be found
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

##### URL processor is not registered, that is, the URL can be found

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

