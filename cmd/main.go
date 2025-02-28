package main

import (
	"fmt"
	lorago "github.com/LorraineWen/lorago/router"
	"net/http"
	"os"
)

type User struct {
	Name string
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
	userGroup.Get("/index", func(context *lorago.Context) {
		context.String(http.StatusOK, "Hello World %s")
	})
	userGroup.Get("/login", func(context *lorago.Context) {
		context.Redirect(307, "/user/index")
	})
	engine.Run()
}
