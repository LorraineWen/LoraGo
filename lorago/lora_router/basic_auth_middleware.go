package lora_router

import (
	"encoding/base64"
	"net/http"
)

/*
*@Author: LorraineWen
*实现basic验证，调用的底层函数是http.Request.BasicAuthEntity
 */
type BasicAuthEntity struct {
	UnAuthFunc HandleFunc        //验证失败的处理函数，用户可以传入进来
	Users      map[string]string //存放用户:密码
}

func (basicAuth *BasicAuthEntity) BasicAuthMiddleware(next HandleFunc) HandleFunc {
	return func(ctx *Context) {
		//通过请求头中的base64推测出username和password
		requestUserName, requestPassword, ok := ctx.R.BasicAuth()
		if !ok {
			basicAuth.unAuthHandler(ctx)
			return
		}
		password, ok := basicAuth.Users[requestUserName]
		if !ok {
			basicAuth.unAuthHandler(ctx)
			return
		}
		if password != requestPassword {
			basicAuth.unAuthHandler(ctx)
			return
		}
		//通过验证
		ctx.BasicSet("username", requestUserName)
		next(ctx)
	}
}
func (basicAuth *BasicAuthEntity) unAuthHandler(ctx *Context) {
	if basicAuth.UnAuthFunc != nil {
		basicAuth.UnAuthFunc(ctx)
	} else {
		ctx.W.WriteHeader(http.StatusUnauthorized)
	}
}

// 这个函数是用来获取base64
func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
