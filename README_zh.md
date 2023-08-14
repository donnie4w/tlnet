### tlnet  极简的 go http 服务框架 [[English document]](https://github.com/donnie4w/tlnet/blob/main/README.md "[English document]")
###### tlnet is go http service framework
### 功能说明

1. tlnet主要目的是让http开发更加简单，代码更加洁简
2. tlnet具备http基本功能.包括GET，PUT，POST等
3. tlnet实现拦截器，包括正则匹配拦截url等拦截功能
4. tlnet支持websocket,实现websocket服务变得极为简单
5. tlnet遵循原生gohttp的使用逻辑，所以极容易将原生http迁移到tlnet
6. tlnet遵循go原生路由，性能由当前go版本http性能决定

### [官方网站](http://tlnet.top/tlnet "官方网站")

### [Tlnet Api文档](http://tlnet.top/tlnetfunc "Tlnet Api文档")

### 基础使用用例：

	tlnet := NewTlnet() 
	tlnet.GET("/g", getFunc)       //http get请求
	tlnet.POST("/p", postFunc)   //http post请求
	tlnet.HandleStatic("/js", "/html/js", nil) //设置静态资源访问
	tlnet.HttpStart(":8000")  

	func getFunc(hc *HttpContext) {
		logging.Debug(hc.ReqInfo)  //打印http信息
		logging.Debug(hc.GetParam("name"))  //打印test
		hc.ResponseString("hello tlnet")  /*返回hello tlnet*/
	}

	func postFunc(hc *HttpContext) {
		logging.Debug(hc.ReqInfo)  //打印http信息
		logging.Debug(hc.PostParam("name"))  
		hc.ResponseString("hello tlnet")  /*返回hello tlnet*/
	}

### 拦截器使用用例：

	tlnet := NewTlnet() 
	tlnet.GET("/g", getFunc)       //http get请求
	tlnet.POST("/p", postFunc)   //http post请求
	tlnet.Handle("/", func(hc *HttpContext) {hc.ResponseString("hi,123")) })   //http通用处理
	tlnet.HandleWithFilter("/",interceptFilter(`[a-z]`),func(hc *HttpContext) {hc.ResponseString("hi,123")) })   //http通用处理并定义拦截器
	//设置静态资源访问
	tlnet.HandleStatic("/js", "/html/js", nil)
	tlnet.HandleStaticWithFilter("/", "./", notAllow("(.go)$"), nil)  //静态资源并定义拦截器，拦截go文件访问权限
	tlnet.HttpStart(":8000") 

	func getFunc(hc *HttpContext) {
		logging.Debug(hc.ReqInfo)  //打印http一般信息
		logging.Debug(hc.GetParam("name"))  //打印test
		hc.ResponseString("hello tlnet")  /*返回hello tlnet*/
	}

	func postFunc(hc *HttpContext) {
		logging.Debug(hc.ReqInfo)  //打印http一般信息
		logging.Debug(hc.PostParam("name"))  
		hc.ResponseString("hello tlnet")  /*返回hello tlnet*/
	}
	
	//拦截器
	func interceptFilter(pattern string) (f *Filter) {
		f = NewFilter()
		f.AddIntercept(pattern, func(hc *HttpContext) bool {
		hc.ResponseString("被拦截了")
		return true // 返回true则 HandleWithFilter中handlerFunc不执行，false则会执行
		})
		return
	}
	
	// 定义拦截器，不允许访问go文件: (.go)$
	func notAllow(pattern string) (f *Filter) {
		f = NewFilter()
		f.AddIntercept(pattern, func(hc *HttpContext) bool {
		hc.ResponseBytes(http.StatusMethodNotAllowed, []byte(hc.ReqInfo.Uri+" not allow"))
		return true // 返回true则 HandleWithFilter中handlerFunc不执行，false则会执行
		})
		return
	}
	

### websocket的基础使用：

	tlnet := NewTlnet() 
	tlnet.HandleWebSocket("/ws", websocketFunc)  
	tlnet.HttpStart(8080) 
	
	var wsmap = make(map[int64]*HttpContext, 0)
	func websocketFunc(hc *HttpContext) {
	//每个websocket有一个id，链接识别id
		wsId := hc.WS.Id
		msg := string(hc.WS.Read())  //  hc.WS.Read()读取客户端请求数据
		_, ok := wsmap[wsId]
		if !ok {
			if msg == "123456" {
				wsmap[wsId] = hc
				hc.WS.Send(fmt.Sprint("id=", wsId))  //向客户端发送数据
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
	

### tlnet的高级使用用例：

	tlnet.SetMaxBytesReader((1 << 20) * 50) //websocket body最大限制 50m
	tlnet.HandleWebSocketBindOrigin("/ws", "http://tlnet/", websocketFunc)  //定义Origin
	tlnet.HandleWebSocketBindConfig("/ws3", websocketFunc, newWebsocketConfig()) //定义 config，实现websocket的错误处理方法
	tlnet.AddHandlerFunc("/aaa", nil, func(w ResponseWriter, r *Request){})  // 对接go http原生handler
	tlnet.AddStaticHandler("/", "./", nil, nil)//静态资源对接原生handler
	tlnet.ReadTimeout(10 * time.Second) //设置读取超时
	tlnet.HttpStartTLS(":8000",certFile, keyFile)  //TLS by file
	tlnet.HttpStartTlsBytes(":8000",certFileBs, keyFileBs) //TLS by bytes
	tlnet.Close()  //关闭连接

------------
