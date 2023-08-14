### tlnet: minimalist http service framework for go  [[中文文档]](https://github.com/donnie4w/tlnet/blob/main/README_zh.md "[中文文档]")

1. The main purpose is to make http development simpler and the code cleaner
2. Provides basic http functions. Including GET, PUT, POST, etc
3. Implement interceptor, including regular matching intercept url and other interception functions
4. Supports websocket, and implementing websocket services is extremely simple
5. Tlnet follows the usage logic of native gohttp, so migrating native http to tlnet is extremely easy
6. Tlnet follows the go native route and the performance is determined by the current go version of http

### [Official Website](http://tlnet.top/tlnet "official website")

### [Tlnet Api](http://tlnet.top/tlnetfunc "tlnet api")

### Basic use cases：

	tlnet := NewTlnet() 
	tlnet.GET("/g", getFunc)       //http get request
	tlnet.POST("/p", postFunc)   //http post request
	tlnet.HandleStatic("/js", "/html/js", nil) //static resource access
	tlnet.HttpStart(":8000")  

	func getFunc(hc *HttpContext) {
		logging.Debug(hc.ReqInfo)  //Print http request information
		logging.Debug(hc.GetParam("name"))  //Prints the value of the parameter `name`
		hc.ResponseString("hello tlnet") //Return string to client：hello tlnet
	}

	func postFunc(hc *HttpContext) {
		logging.Debug(hc.ReqInfo)  
		logging.Debug(hc.PostParam("name"))
		hc.ResponseString("hello tlnet") 
	}

------------


### Interceptor use cases

	tlnet := NewTlnet() 
	tlnet.GET("/g", getFunc)       //http get request
	tlnet.POST("/p", postFunc)   //http post request
	tlnet.Handle("/", func(hc *HttpContext) {hc.ResponseString("hi,123")) })   //tlnet general processing method
	tlnet.HandleWithFilter("/",interceptFilter(`[a-z]`),func(hc *HttpContext) {hc.ResponseString("hi,123")) })   //interceptFilter() is Interceptor
	//Static resource access service
	tlnet.HandleStatic("/js", "/html/js", nil)
	tlnet.HandleStaticWithFilter("/", "./", notAllow("(.go)$"), nil)  //notAllow() block go file access
	tlnet.HttpStart(":8000") 

	func getFunc(hc *HttpContext) {
		logging.Debug(hc.ReqInfo)  
		logging.Debug(hc.GetParam("name"))
		hc.ResponseString("hello tlnet")
	}

	func postFunc(hc *HttpContext) {
		logging.Debug(hc.ReqInfo)  
		logging.Debug(hc.PostParam("name"))
		hc.ResponseString("hello tlnet") 
	}
	
	//interceptor
	func interceptFilter(pattern string) (f *Filter) {
		f = NewFilter()
		f.AddIntercept(pattern, func(hc *HttpContext) bool {
		hc.ResponseString("be intercepted")
		return true // If true is returned, handlerFunc in HandleWithFilter is not executed, and if false is executed
		})
		return
	}
	
	// This custom interceptor is designed to disallow access to go files: (.go)$
	func notAllow(pattern string) (f *Filter) {
		f = NewFilter()
		f.AddIntercept(pattern, func(hc *HttpContext) bool {
		hc.ResponseBytes(http.StatusMethodNotAllowed, []byte(hc.ReqInfo.Uri+" not allow"))
		return true // If true is returned, handlerFunc in HandleWithFilter is not executed, and if false is executed
		})
		return
	}
	

### websocket usage：

	tlnet := NewTlnet() 
	tlnet.HandleWebSocket("/ws", websocketFunc)  
	tlnet.HttpStart(8080) 
	
	var wsmap = make(map[int64]*HttpContext, 0)
	func websocketFunc(hc *HttpContext) {
	//Each websocket has an id, a unique identifier
		wsId := hc.WS.Id
		msg := string(hc.WS.Read())  //  hc.WS.Read() Read client request data
		_, ok := wsmap[wsId]
		if !ok {
			if msg == "123456" {
				wsmap[wsId] = hc
				hc.WS.Send(fmt.Sprint("id=", wsId))  //Send data to the client
			}
			return
		}
		logging.Debug(hc.ReqInfo)
	//receive client message
		logging.Debug(msg)
		hc.WS.Send("What you sent was：" + msg)
		for k, v := range wsmap {
			if k != wsId {
				v.WS.Send(fmt.Sprint("this is a message from【", wsId, "】:", msg))
			}
		}
	}
	


------------


### Advanced use cases for tlnet：

	tlnet.SetMaxBytesReader((1 << 20) * 50) //means:websocket body size maximum limit: 50m
	tlnet.HandleWebSocketBindOrigin("/ws", "http://tlnet/", websocketFunc)  //set Origin
	tlnet.HandleWebSocketBindConfig("/ws3", websocketFunc, newWebsocketConfig()) //Implement a websocket config
	tlnet.AddHandlerFunc("/aaa", nil, func(w ResponseWriter, r *Request){})  // interface to go http native handler
	tlnet.AddStaticHandler("/", "./", nil, nil)//Static Resources interface to go http native handler
	tlnet.ReadTimeout(10 * time.Second) //read timeout
	tlnet.HttpStartTLS(":8000",certFile, keyFile)  //TLS by file
	tlnet.HttpStartTlsBytes(":8000",certFileBs, keyFileBs) //TLS by bytes
	tlnet.Close()  //close tlnet service

------------
