package web_frame

/*
 @Author: LorraineWen
 @Date: 2025/2/23 14:23:49
 该文件主要实现处理http请求的路由，包括路由分组，支持不同请求方法(不同路径和同一路径)
*/
import (
	"fmt"
	"net/http"
)

const (
	POST = http.MethodPost
	GET  = http.MethodGet
	ANY  = "ANY"
)

type Context struct {
	W http.ResponseWriter
	R *http.Request
}
type HandleFunc func(ctx *Context)

type router struct {
	groups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	g := &routerGroup{groupName: name, handlerMap: make(map[string]map[string]HandleFunc)}
	r.groups = append(r.groups, g)
	return g
}

type routerGroup struct {
	groupName  string
	handlerMap map[string]map[string]HandleFunc //路由对应的方法对应的处理函数
}

func (r *routerGroup) MethodHandle(name string, method string, handleFunc HandleFunc) {
	if _, ok := r.handlerMap[name]; !ok {
		r.handlerMap[name] = make(map[string]HandleFunc)
	}
	r.handlerMap[name][method] = handleFunc
}

// ANY类型的路由
func (r *routerGroup) Any(name string, handlerFunc HandleFunc) {
	r.MethodHandle(name, ANY, handlerFunc)
}

// GET类型的路由
func (r *routerGroup) Get(name string, handlerFunc HandleFunc) {
	r.MethodHandle(name, GET, handlerFunc)
}

// POST类型的路由
func (r *routerGroup) Post(name string, handlerFunc HandleFunc) {
	r.MethodHandle(name, POST, handlerFunc)
}

// 这里是直接嵌入了类型，所以Engine继承了router的方法和成员
type Engine struct {
	*router
}

func New() *Engine {
	return &Engine{&router{}}
}
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	groups := e.router.groups
	for _, group := range groups {
		for name, methodHandleFuncs := range group.handlerMap {
			url := "/" + group.groupName + name
			if r.RequestURI == url { //说明路由匹配
				//找到路由对应的请求方法
				ctx := &Context{R: r, W: w}
				_, ok := methodHandleFuncs[ANY]
				if ok {
					methodHandleFuncs[ANY](ctx)
					return
				}
				method := r.Method
				_, ok = methodHandleFuncs[method]
				if ok {
					methodHandleFuncs[method](ctx)
					return
				}
				//如果各个方法的路由中都找不到对应的路由就返回405
				w.WriteHeader(http.StatusMethodNotAllowed)
				fmt.Fprintln(w, method+"服务器暂不支持")
				return
			} else {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintln(w, r.RequestURI+"没有找到")
				return
			}
		}
	}
}
func (e *Engine) Run() {
	http.Handle("/", e)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
