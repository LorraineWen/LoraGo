package main

import (
	"fmt"
	lorago "github.com/LorraineWen/lorago/router"
	"net/http"
	"os"
)

type User struct {
	Name     string `json:"name" binding:"required"`
	Age      int    `json:"age"`
	Password string `json:"password"`
}
type RespError struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

func m1(next lorago.HandleFunc) lorago.HandleFunc {
	return func(ctx *lorago.Context) {
		//这里才是中间件在做的事情
		fmt.Fprintln(ctx.W, "Pre Handle1:")
		//这个next就是/user/name对应的处理函数
		next(ctx)
		fmt.Fprintln(ctx.W, "Post Handle1:")
		fmt.Fprintln(ctx.W, "Post Handle2:")
	}
}
func main() {
	engine := lorago.New()
	fmt.Println(os.Getwd())
	userGroup := engine.Group("user")
	//直接传入模板名称和数据就可以了
	engine.LoadTemplate("../test/template/*.html")
	userGroup.Post("/index2", func(context *lorago.Context) {
		user := []User{}
		context.Validate = true
		context.DisallowUnknownFields = true
		err := context.BindJson(&user)
		if err != nil {
			context.JsonResponseWrite(http.StatusInternalServerError, &RespError{Code: http.StatusInternalServerError, Error: err.Error()})
			return
		}
		context.JsonResponseWrite(http.StatusOK, user)
	})
	userGroup.Post("/index1", func(context *lorago.Context) {
		res := context.DefaultFormQuery("id", "1")
		fmt.Println(res)
		context.JsonResponseWrite(http.StatusOK, res)
	})
	engine.Run()
}
