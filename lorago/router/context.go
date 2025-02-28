package router

import (
	"fmt"
	"github.com/LorraineWen/lorago/router/render"
	"github.com/LorraineWen/lorago/util"
	"html/template"
	"net/http"
	"net/url"
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
	W http.ResponseWriter
	R *http.Request
	e *Engine //用于获取模板渲染函数
}

func (c *Context) Render(code int, r render.Render) error {
	err := r.Render(c.W)
	c.W.WriteHeader(code)
	return err
}

// 支持html格式响应
func (ctx *Context) Html(status int, data any) error {
	err := ctx.Render(status, &render.HtmlRender{IsTemplate: false, Data: data})
	return err
}

// 支持json格式响应
func (ctx *Context) Json(status int, data any) error {
	err := ctx.Render(status, &render.JsonRender{Data: data})
	return err
}

// 支持xml格式响应
func (ctx *Context) Xml(status int, data any) error {
	err := ctx.Render(status, &render.XmlRender{Data: data})
	return err
}

// 支持格式化String格式响应
func (ctx *Context) String(status int, format string, data ...any) (err error) {
	err = ctx.Render(status, &render.StringRender{
		Format: format,
		Data:   data,
	})
	return
}

// 支持提前将模板加载到内存中
func (ctx *Context) Template(status int, name string, data any) error {
	err := ctx.Render(status, &render.HtmlRender{
		IsTemplate: true,
		Name:       name,
		Data:       data,
		Template:   ctx.e.htmlRender.Template,
	})
	return err
}

// 支持HTML模板，动态渲染HTML，对ParseFiles函数的封装
func (ctx *Context) HtmlTemplate(name string, funcMap template.FuncMap, data any, fileName ...string) error {
	t := template.New(name)
	t.Funcs(funcMap)
	t, err := t.ParseFiles(fileName...)
	if err != nil {
		return err
	}
	ctx.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.Execute(ctx.W, data)
	if err != nil {
		return err
	}
	return nil
}

// 支持HTML模板，使用通配符对一个目录下的html文件进行动态渲染，对ParseGlob函数的封装
func (ctx *Context) HtmlTemplateGlob(name string, funcMap template.FuncMap, pattern string, data any) error {
	t := template.New(name)
	t.Funcs(funcMap)
	t, err := t.ParseGlob(pattern)
	if err != nil {
		return err
	}
	ctx.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.Execute(ctx.W, data)
	if err != nil {
		return err
	}
	return nil
}

// 支持文件下载
func (ctx *Context) File(filePath string) {
	http.ServeFile(ctx.W, ctx.R, filePath)
}

// 支持自定义文件名称下载，下载好的文件名称自动变为filename
func (c *Context) FileAttachment(filepath, filename string) {
	if util.IsASCII(filename) {
		c.W.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	} else {
		c.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.W, c.R, filepath)
}

// 从本地文件系统下载文件，fileSystem实际上就是一个本地的目录http.Dir("../test/template")
// filePath就是以template作为根目录了
func (c *Context) FileFromFileSystem(filePath string, fileSystem http.FileSystem) {
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
