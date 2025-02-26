package router

/*
*@Author: LorraineWen
*@Date: 2025/2/23 14:23:49
*该文件主要实现处理http请求的路由，包括路由分组，支持不同请求方法(不同路径和同一路径)
*支持动态路由(/user/get/:id，支持通配符/static/**
*支持前置中间件和后置中间件
 */
import (
	"fmt"
	"github.com/LorraineWen/lorago/util"
	"net/http"
)

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

type Context struct {
	W http.ResponseWriter
	R *http.Request
}
type HandleFunc func(ctx *Context)

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	g := &routerGroup{
		groupName:          name,
		handlerMap:         make(map[string]map[string]HandleFunc),
		middlewaresFuncMap: make(map[string]map[string][]MiddlewareFunc),
		trieNode: &trieNode{
			name:     "/",
			children: make([]*trieNode, 0)},
	}
	r.routerGroups = append(r.routerGroups, g)
	return g
}

type MiddlewareFunc func(handleFunc HandleFunc) HandleFunc
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
}

func New() *Engine {
	return &Engine{&router{}}
}
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	for _, group := range e.routerGroups {
		//判断请求中的URL里面是否包含分组路径
		routerName := util.SubStringLast(r.RequestURI, "/"+group.groupName) //如果url中包含分组路径，那么就返回url中分组路径后面的请求路径，/user/getname，返回/getname
		node := group.trieNode.get(routerName)
		//对于/user/getname/1,routerName=/getname/1
		//node.routerName=/get/name/:id，这也是我们实际注册的路由，所应该应该使用node.routerName来索引得到处理routerName的函数
		if node != nil && node.isEnd {
			ctx := &Context{R: r, W: w}
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
func (e *Engine) Run() {
	http.Handle("/", e)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
