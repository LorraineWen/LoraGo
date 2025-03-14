package lora_router

import (
	"fmt"
	"github.com/LorraineWen/lorago/lora_conf"
	"github.com/LorraineWen/lorago/lora_log"
	"github.com/LorraineWen/lorago/lora_render"
	"github.com/LorraineWen/lorago/lora_util"
	"html/template"
	"log"
	"net/http"
	"sync"
)

/*
*@Author: LorraineWen
*@Date: 2025/2/23
*该文件主要实现处理http请求的路由，包括路由分组，支持不同请求方法(不同路径和同一路径)
*支持动态路由(/user/get/:id，支持通配符/static/**
*支持前置中间件和后置中间件，并将两者合并
*支持组路由中间件和单个路由中间件
 */
const (
	POST    = http.MethodPost
	GET     = http.MethodGet
	ANY     = "ANY"
	DELETE  = http.MethodDelete
	PUT     = http.MethodPut
	PATCH   = http.MethodPatch
	OPTIONS = http.MethodOptions
	HEAD    = http.MethodHead
)

// 定义路由回调函数类型
type HandleFunc func(ctx *Context)

// 定义路由类型，该路由是按照路由组进行注册的，所以只支持路由组方式注册路由
type router struct {
	routerGroups []*routerGroup
	engine       *Engine //通过处理器为每个路由组设置中间件
}

// 获取路由组对象
// 调用方式:userGroup:=Group("user")
func (r *router) Group(name string) *routerGroup {
	routerGroup := &routerGroup{
		groupName:          name,
		handlerMap:         make(map[string]map[string]HandleFunc),
		middlewaresFuncMap: make(map[string]map[string][]MiddlewareFunc),
		trieNode: &trieNode{
			name:     "/",
			children: make([]*trieNode, 0)},
	}
	routerGroup.Use(r.engine.MiddlewareFuncs...)
	r.routerGroups = append(r.routerGroups, routerGroup)
	return routerGroup
}

// 定义中间件回调函数类型
type MiddlewareFunc func(handleFunc HandleFunc) HandleFunc

// 定义路由组对象类型
type routerGroup struct {
	groupName  string
	handlerMap map[string]map[string]HandleFunc //路由对应的方法对应的处理函数
	//getname:post:postnamefunc
	//getname:get:getnamefunc
	//getname:delete:deletenamefunc
	middlewaresFuncMap map[string]map[string][]MiddlewareFunc //适用于单个路由的中间件
	MiddleWare         []MiddlewareFunc                       //适用于组路由中间件
	trieNode           *trieNode
}

// Get，Post等函数的内部实现函数
func (r *routerGroup) MethodHandle(name string, method string, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	if _, ok := r.handlerMap[name]; !ok {
		r.handlerMap[name] = make(map[string]HandleFunc)
		r.middlewaresFuncMap[name] = make(map[string][]MiddlewareFunc)
	}
	if _, ok := r.handlerMap[name][method]; ok {
		panic("该路由已经注册过了")
	}
	r.handlerMap[name][method] = handleFunc
	r.middlewaresFuncMap[name][method] = append(r.middlewaresFuncMap[name][method], middlewareFunc...)
	r.trieNode.put(name)
}

// ANY类型的路由
func (r *routerGroup) Any(name string, handlerFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	r.MethodHandle(name, ANY, handlerFunc, middlewareFunc...)
}

// GET类型的路由
func (r *routerGroup) Get(name string, handlerFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	r.MethodHandle(name, GET, handlerFunc, middlewareFunc...)
}

// POST类型的路由
func (r *routerGroup) Post(name string, handlerFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	r.MethodHandle(name, POST, handlerFunc, middlewareFunc...)
}

// DELETE类型的路由
func (r *routerGroup) Delete(name string, handlerFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	r.MethodHandle(name, DELETE, handlerFunc, middlewareFunc...)
}

// PUT类型的路由
func (r *routerGroup) Put(name string, handlerFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	r.MethodHandle(name, PUT, handlerFunc, middlewareFunc...)
}

// PATCH类型的路由
func (r *routerGroup) Patch(name string, handlerFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	r.MethodHandle(name, PATCH, handlerFunc, middlewareFunc...)
}

// HEAD类型的路由
func (r *routerGroup) Head(name string, handlerFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	r.MethodHandle(name, HEAD, handlerFunc, middlewareFunc...)
}

// OPTIONS类型的路由
func (r *routerGroup) Options(name string, handlerFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	r.MethodHandle(name, OPTIONS, handlerFunc, middlewareFunc...)
}

// 中间件注册函数:路由组级别注册
func (r *routerGroup) Use(middlewares ...MiddlewareFunc) {
	r.MiddleWare = append(r.MiddleWare, middlewares...)
}
func (r *routerGroup) MiddlewareHandleFunc(ctx *Context, name, method string, hanldefunc HandleFunc) {
	//路由组级别的中间件
	if r.MiddleWare != nil {
		for _, middlewareFunc := range r.MiddleWare {
			//一开始handlefunc就是/user/name的请求处理函数，获取第一个中间件的处理函数之后
			//handlefunc就变成了第一个中间件的处理函数，接着变成第二个中间件的处理函数
			hanldefunc = middlewareFunc(hanldefunc)
		}
	}
	//路由级别的中间件
	if r.middlewaresFuncMap[name][method] != nil {
		for _, middlewareFunc := range r.middlewaresFuncMap[name][method] {
			hanldefunc = middlewareFunc(hanldefunc)
		}
	}
	hanldefunc(ctx) //这个handlefunc调用的是最后一个注册的中间件的处理函数，最后一个中间件的处理函数的next调用的则是倒数第二个中间件的处理函数
}

// 这里是直接嵌入了类型，所以Engine继承了router的方法和成员
type Engine struct {
	*router
	funcMap         template.FuncMap               //设置html模板渲染时所需要的函数
	htmlRender      lora_render.HtmlTemplateRender //在内存中存放html模板
	pool            sync.Pool                      //存放context对象，避免context对象的多次重复创建，导致多次重复释放内存和分配内存
	Logger          *lora_log.Logger               //初始化context里面的日志对象
	MiddlewareFuncs []MiddlewareFunc               //初始化的处理器的时候就需要注册的中间件
	errHandler      ErrorHandler                   //支持code和status
}

// 直接初始化引擎
func New() *Engine {
	engine := &Engine{router: &router{}, funcMap: nil, htmlRender: lora_render.HtmlTemplateRender{}, Logger: lora_log.NewLogger()}
	engine.pool.New = func() any {

		return engine.allocateContext()
	}
	engine.Use(RecoveryMiddleware, LogMiddleware) //自动使用日志和panic捕获的中间件
	engine.router.engine = engine
	return engine
}

// 通过配置文件初始化引擎
func Default() *Engine {
	engine := New()
	logPath, ok := lora_conf.TomlConf.Log["path"]
	if ok {
		engine.Logger.SetLogPath(logPath.(string))
	}
	engine.Use(RecoveryMiddleware, LogMiddleware)
	return engine
}
func (e *Engine) Use(middlewareFunc ...MiddlewareFunc) {
	e.MiddlewareFuncs = append(e.MiddlewareFuncs, middlewareFunc...)
}

// 由于context对象会存在许多的属性，所以单独抽取出一个函数来进行context的初始化
func (e *Engine) allocateContext() any {
	return &Context{engine: e}
}

// 以下三个函数都是在渲染html模板时，需要调用的函数
// 通过配置文件加载模板
func (e *Engine) LoadTemplateGlobByConf() {
	pattern, ok := lora_conf.TomlConf.Template["pattern"]
	if !ok {
		panic("config pattern not exist")
	}
	e.LoadTemplate(pattern.(string))
}

// 设置html需要的一些函数
func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

// 将html模板加载到内存中
// 调用方式:engine.LoadTemplate("../test/template/*.html")
func (e *Engine) LoadTemplate(pattern string) {
	t := template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
	e.htmlRender = lora_render.HtmlTemplateRender{Template: t}
}

// Engine需要实现ServeHTTP函数，才能实现Hanler接口，Engine才能成为一个自定义的路由处理器
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	ctx := e.pool.Get().(*Context)
	ctx.W = w
	ctx.R = r
	ctx.Logger = e.Logger
	for _, group := range e.routerGroups {
		//判断请求中的URL里面是否包含分组路径
		routerName := lora_util.SubStringLast(r.URL.Path, "/"+group.groupName) //如果url中包含分组路径，那么就返回url中分组路径后面的请求路径，/user/getname，返回/getname
		node := group.trieNode.get(routerName)
		//对于/user/getname/1,routerName=/getname/1
		//node.routerName=/get/name/:id，这也是我们实际注册的路由，所应该应该使用node.routerName来索引得到处理routerName的函数
		if node != nil && node.isEnd {
			handle, ok := group.handlerMap[node.routerName][ANY]
			if ok {
				group.MiddlewareHandleFunc(ctx, node.routerName, method, handle)
				return
			}
			handle, ok = group.handlerMap[node.routerName][method]
			if ok {
				group.MiddlewareHandleFunc(ctx, node.routerName, method, handle)
				return
			}
			//如果各个方法的路由中都找不到对应的路由就返回405
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintln(w, method+"服务器暂不支持")
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, r.RequestURI+"没有找到")
	return
}

// 开启https验证
func (e *Engine) RunTLS(addr, certFile, keyFile string) {
	err := http.ListenAndServeTLS(addr, certFile, keyFile, e)
	if err != nil {
		log.Fatal("ListenAndServeTLS: ", err)
	}
}

type ErrorHandler func(err error) (int, any)

func (e *Engine) RegisterErrorHandler(err ErrorHandler) {
	e.errHandler = err
}
func (e *Engine) Run() {
	//e是一个自定义的路由处理器
	err := http.ListenAndServe(":8080", e)
	if err != nil {
		panic(err)
	}
}
