package router

import (
	"html/template"
	"net/http"
)

/*
*@Author: LorraineWen
*@Date: 2025/2/26
*该文件主要提供上下文有关的接口
*由于http.ResponseWriter存放于Context中，所以Context应该提供接口进行模板渲染
 */
type Context struct {
	W http.ResponseWriter
	R *http.Request
	e *Engine //用于获取模板渲染函数
}

// 渲染HTML需要明确content-type和charset
func (ctx *Context) Html(status int, html string) error {
	ctx.W.WriteHeader(status)
	ctx.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := ctx.W.Write([]byte(html))
	if err != nil {
		return err
	}
	return nil
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

// 支持提前将模板加载到内存中
func (c *Context) Template(name string, data any) error {
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := c.e.htmlRender.Template.ExecuteTemplate(c.W, name, data)
	if err != nil {
		return err
	}
	return nil
}
