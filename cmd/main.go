package main

import (
	"fmt"
	lorago "github.com/LorraineWen/lorago/lora_router"
	"github.com/LorraineWen/lorago/lora_router/lora_auth"
	"log"
	"net/http"
	"time"
)

func main() {
	engine := lorago.New()
	userGroup := engine.Group("user")
	//创建basic验证实体
	base64 := lorago.BasicAuth("admin", "12345")
	fmt.Println(base64)
	basicAuth := lorago.BasicAuthEntity{
		UnAuthFunc: func(ctx *lorago.Context) {
			ctx.W.Write([]byte("验证失败"))
		},
		Users: make(map[string]string),
	}
	//这个就是数据库里面存放的用户和对应的密码
	basicAuth.Users["admin"] = "12345"
	//使用basic验证中间件
	userGroup.Use(basicAuth.BasicAuthMiddleware)
	userGroup.Get("/index2", func(context *lorago.Context) {
		user := make(map[string]string)
		user["name"] = "amie"
		user["age"] = "18"
		context.JsonResponseWrite(http.StatusOK, user)
	})
	jwtMiddleware := &lora_auth.JwtAuth{Key: []byte("123456")}
	engine.Use(jwtMiddleware.JwtAuthMiddleware)
	userGroup.Get("/login", func(ctx *lorago.Context) {

		jwt := &lora_auth.JwtAuth{}
		jwt.Key = []byte("123456")
		jwt.SendCookie = true
		jwt.TimeOut = 10 * time.Minute
		jwt.Authenticator = func(ctx *lorago.Context) (map[string]any, error) {
			data := make(map[string]any)
			data["userId"] = 1
			return data, nil
		}
		token, err := jwt.LoginHandler(ctx)
		if err != nil {
			log.Println(err)
			ctx.JsonResponseWrite(http.StatusOK, err.Error())
			return
		}
		ctx.JsonResponseWrite(http.StatusOK, token)
	})
	engine.Run()
}
