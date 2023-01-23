## tlnet
### tlnet is a web framework written in Go.
### 简单说明
- tlnet基于go内置http服务封装.使开发可以在一般编程习惯下进行.
- 定义了过滤器(拦截器)，和 业务逻辑的handler
- 引入thrift，thrift对多种语言包括js都有很好的支持，当然，可以不用它.
- 引入leveldb进行数据存储。如果业务简单，比如个人博客，可以依赖leveldb存储数据

### 一般用法：
	tlnet := NewTlnet() 
	 /*设置数据存储的文件路径，不用leveldb 可以不理会 */
	// tlnet.DBPath("test.db")    
	tlnet.SetMaxBytesReader((1 << 20) * 50) //设置请求body最大值，50M，无要求时不设置即可
	//tlnet设置请求超时，读取超时等方法类同
	//设置访问路径和逻辑处理方法
	tlnet.Handle("/qq", handleFunc)
	//设置静态资源访问
	tlnet.StaticDir("/s", "static/html", nil)
	tlnet.HttpStart(8080) 
	//tlnet.HttpStartTLS(port int32, certFile, keyFile string)
	
	//curl http://127.0.0.1:8080/qq?name=test
	//逻辑处理方法的使用用例：
	func handleFunc(hc *HttpContext) {
		logging.Debug(hc.ReqInfo)  //打印http一般信息
		/*GetParam()  获取GET请求的参数值*/
		logging.Debug(hc.GetParam("name"))  //打印test  
		logging.Debug(net.SplitHostPort(hc.ReqInfo.RemoteAddr))
		/*返回hello tlnet*/
		hc.ResponseString(http.StatusOK, "hello tlnet")
	}
打印结果：<br/>
[DEBUG]2023/01/23 23:22:14 tlnet_test.go:47: &{ /qq?name=test GET 127.0.0.1:8080 127.0.0.1:54236 curl/7.83.1 map[Accept:[*/*] User-Agent:[curl/7.83.1]]}<br/>
[DEBUG]2023/01/23 23:22:14 tlnet_test.go:48: test<br/>
[DEBUG]2023/01/23 23:22:14 tlnet_test.go:51: 127.0.0.154236<nil><br/>

### 方法说明：
	   //pattern请求的路径，handlerFunc实现逻辑处理 
	 1. Handle(pattern string, handlerFunc func(hc *HttpContext))
	
		 //Filter为拦截处理
	 2. HandleWithFilter(pattern string, _filter *Filter, handlerFunc func(hc *HttpContext))

	//Filter用法例如：
	 func httpFilter() *Filter {
		f := NewFilter()
		/*正则匹配请求uri，匹配成功后重定向网址*/
		f.AddIntercept(".html$", func(hc *HttpContext) {
		hc.Redirect("https://github.com")
		})
	}
	
	3.StaticDir(pattern, dir string, handlerFunc func(hc *HttpContext))
	//定义静态资源服务，dir为资源路径 
	
	4.StaticDirWithFilter(pattern, dir string, _filter *Filter, handlerFunc func(hc *HttpContext)) 
	
	HttpContext的方法：
		//返回数据 status默认0，返回200
	1. ResponseString(status int, _s string)
	2. GetParam(key string)  //获取Get请求参数值
	3. PostParam(key string) //获取Post请求参数值
	4. Redirect(path string)   //返回重定向
	5. FormFile(key string) (multipart.File, *multipart.FileHeader, error) //获取上传文件对象
	6.ParseFormFile(file multipart.File, fileHeader *multipart.FileHeader, savePath string) (fileName string, err error) //处理存储上传文件对象
	7.FormFiles(key string) *multipart.Form  //获取上传多个文件对象
	
### 以上是最上层方法，一般web开发时常用到
### 下面为上面方法的底层方法，同样可以使用：
	 //加入thrift协议
	 1.tlnet.AddProcessor(pattern string, processor thrift.TProcessor)
	 
	 2.tlnet.AddHandlerFunc(pattern string, f *Filter, handlerFunc func(ResponseWriter, *Request))
	 
	 3.AddStaticHandler(pattern string, dir string, f *Filter, handlerFunc func(ResponseWriter, *Request))
	 //后缀拦截器
	 4.filter.AddSuffixIntercept(suffixs []string, _handler func(ResponseWriter, *Request) bool)
	 //正则匹配拦截器
	 5.filter.AddGlobalIntercept(_pattern string, _handler func(ResponseWriter, *Request) bool)
	 //路径未找到拦截器
	 6.AddNotFoundPageIntercept(_handler func(ResponseWriter, *Request) bool)
	  
### 数据操作
	 tlnet.DBPath("test.db")
	 //存储可序列化对象
	 tlnetAddObject(e any, _idname, _tablename string)
	 
	 AddObjectWithTableIdName(e any, tableIdName string)
	 //更新对象信息
	 UpdateObject(e any, objId, _tablename string) 
	 //获取_tablename为前缀的所有存储对象
	 GetObjectByLike[T any](_tablename string) (ts []*T)
	 
	 GetAndSetId(_idname string)
	 GetObjectByOrder[T any](_tablename, _idname string, startId, endId int32) (ts []*T) 
	 AddValue(key string, value []byte)
	 GetValue(key string) (value []byte, err error)
	 DelKey(key string) (err error) 
	 //备份数据到硬盘
	 (this *DB) BackupToDisk(filename string, prefix []byte)
	 //获取备份的数据
	 RecoverBackup(filename string) (bs []*BakStub)
具体用例可以参考[tlnetDemo](https://github.com/donnie4w/tlnetDemo "tlnetDemo")
### End
