package router

import (
	"fmt"
	"github.com/LorraineWen/lorago/router/render"
	"github.com/LorraineWen/lorago/util"
	"net/http"
	"net/url"
	"strings"
)

/*
*@Author: LorraineWen
*@Date: 2025/2/26
*该文件主要提供上下文有关的接口
*由于http.ResponseWriter存放于Context中，所以Context应该提供接口进行html模板渲染
*还需要支持json，xml等格式的返回
*支持下载文件的需求，可以自定义下载的文件的名称
 */
type Context struct {
	W          http.ResponseWriter
	R          *http.Request
	e          *Engine    //用于获取模板渲染函数
	queryCache url.Values //用于获取请求路径中的参数，实际上就是map[string[]string
}

// 一个多态函数，htmlRender等结构体实现了Render函数，因此可以传入htmlRender等接口体，调用它们自己的Render函数，编码html等响应格式
func (c *Context) Render(status int, r render.Render) error {
	err := r.Render(c.W)
	//避免重复写入
	if status != http.StatusOK { //如果是200，各种格式的Render函数编码响应的时候，会调用Write，ExecuteTemplate等函数，它们会自动将200写入响应头部中
		c.W.WriteHeader(status)
	}
	return err
}

// 支持html格式响应
func (ctx *Context) HtmlResponseWrite(status int, data any) error {
	err := ctx.Render(status, &render.HtmlRender{IsTemplate: false, Data: data})
	return err
}

// 支持json格式响应
// 调用方式context.JsonResponseWrite(http.StatusOK, &User{Name: "amie"})
func (ctx *Context) JsonResponseWrite(status int, data any) error {
	err := ctx.Render(status, &render.JsonRender{Data: data})
	return err
}

// 支持xml格式响应
// 调用方式context.Xml(http.StatusOK, &User{Name: "amie"})
func (ctx *Context) XmlResponseWrite(status int, data any) error {
	err := ctx.Render(status, &render.XmlRender{Data: data})
	return err
}

// 支持格式化String格式响应
// 调用方式context.StringResponseWrite(http.StatusOK, "你好 %s", "amie")
func (ctx *Context) StringResponseWrite(status int, format string, data ...any) (err error) {
	err = ctx.Render(status, &render.StringRender{
		Format: format,
		Data:   data,
	})
	return
}

// 配合e.LoadTemplate函数使用，e.LoadTemplate初始化 ctx.e.htmlRender.Template，Template只需要传递该模板需要的数据
// 调用方式:
// engine.LoadTemplate("../test/template/*.html")
//
//	userGroup.Get("/index", func(context *lorago.Context) {
//		context.TemplateResponseWrite(http.StatusOK, "index.html", &User{Name: "amie"})
//	})
//
// name是../test/template目录下的具体html文件的名称
func (ctx *Context) TemplateResponseWrite(status int, name string, data any) error {
	err := ctx.Render(status, &render.HtmlRender{
		IsTemplate: true,
		Name:       name,
		Data:       data,
		Template:   ctx.e.htmlRender.Template,
	})
	return err
}

// 支持文件下载
func (ctx *Context) FileResponseWrite(filePath string) {
	http.ServeFile(ctx.W, ctx.R, filePath)
}

// 支持自定义文件名称下载，下载好的文件名称自动变为filename
func (c *Context) FileAttachmentResponseWrite(filepath, filename string) {
	if util.IsASCII(filename) {
		c.W.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	} else {
		c.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.W, c.R, filepath)
}

// 从本地文件系统下载文件，fileSystem实际上就是一个本地的目录http.Dir("../test/template")
// filePath就是以template作为根目录了
func (c *Context) FileFromFileSystemResponseWrite(filePath string, fileSystem http.FileSystem) {
	defer func(old string) {
		c.R.URL.Path = old
	}(c.R.URL.Path)

	c.R.URL.Path = filePath

	http.FileServer(fileSystem).ServeHTTP(c.W, c.R)
}

// 路由重定向支持
func (c *Context) Redirect(status int, location string) {
	//由于http.Redirect的重定向只对部分状态码有效果，因此需要对状态码进行判断
	if (status < http.StatusMultipleChoices || status > http.StatusPermanentRedirect) && status != http.StatusCreated {
		panic(fmt.Sprintf("在该状态下无法进行重定向 %d", status))
	}
	http.Redirect(c.W, c.R, location, status)
}

// 将请求路径中的参数，按照map[string][]string的格式存储到c.queryCache中
func (c *Context) initQueryCache() {
	if c.R != nil {
		c.queryCache = c.R.URL.Query() //底层方法就是这个
	} else {
		c.queryCache = url.Values{}
	}
}

// 获取请求参数/adduser?id=1&name=amie
// 调用方式:GetQuery("id")
func (c *Context) GetQuery(key string) string {
	c.initQueryCache() //加载路径参数
	return c.queryCache.Get(key)
}

// 获取数组类型的请求参数，一个key对应多个value
// 一个key多个value:/adduser?id=1&id=2&id=3
// 调用方式:GetQueryArray("id")
func (c *Context) GetQueryArray(key string) (values []string, ok bool) {
	c.initQueryCache() //加载路径参数
	values, ok = c.queryCache[key]
	return
}

// 获取map格式的请求参数/adduser?user[id]=1&user[name]=amie
// 目前只能获取string:string类型的，可以进行数据转换支持更多类型
// 调用方式:GetQueryMap("user")
func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	c.initQueryCache() //加载路径参数
	res := make(map[string]string)
	exist := false
	for requestKey, value := range c.queryCache { //requestKey="user[id]"而不是"id"
		if index1 := strings.IndexByte(requestKey, '['); index1 >= 1 { //大于等于1，是为了避免[出现在user前面
			if index2 := strings.IndexByte(requestKey, ']'); index2 >= index1 {
				if requestKey[0:index1] == key { //key="user"而不是"id"
					exist = true
					res[requestKey[index1+1:index2]] = value[len(value)-1] //一个key对应多个value的情况下，只能获取最后一个值，只是map本身的特性，后面的覆盖前面的
				}
			}
		}
	}
	return res, exist
}

// 为请求参数设置一个默认值
// 调用方式DefaultQuery("id","1")
func (c *Context) DefaultQuery(key, defaultValue string) string {
	array, ok := c.GetQueryArray(key)
	if !ok {
		return defaultValue
	}
	return array[0]
}
