package main

import (
	"fmt"
	lorago "github.com/LorraineWen/lorago/router"
)

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
	userGroup := engine.Group("user")
	userGroup.Get("/name", func(context *lorago.Context) {
		fmt.Fprintln(context.W, "get amie")
	}, m1)
	userGroup.Get("/name1", func(context *lorago.Context) {
		fmt.Fprintln(context.W, "get amie")
	})
	//userGroup.Use(func(next lorago.HandleFunc) lorago.HandleFunc {
	//	return func(ctx *lorago.Context) {
	//		//这里才是中间件在做的事情
	//		fmt.Fprintln(ctx.W, "Pre Handle1:")
	//		//这个next就是/user/name对应的处理函数
	//		next(ctx)
	//		fmt.Fprintln(ctx.W, "Post Handle1:")
	//		fmt.Fprintln(ctx.W, "Post Handle2:")
	//	}
	//})
	//userGroup.Use(func(next lorago.HandleFunc) lorago.HandleFunc {
	//	return func(ctx *lorago.Context) {
	//		//这里才是中间件在做的事情
	//		fmt.Fprintln(ctx.W, "Pre Handle2:")
	//		//这个next就是/user/name对应的处理函数
	//		next(ctx)
	//		fmt.Fprintln(ctx.W, "Post Handle3:")
	//		fmt.Fprintln(ctx.W, "Post Handle4:")
	//	}
	//})
	userGroup.Post("/name", func(context *lorago.Context) {
		fmt.Fprintln(context.W, "get amie")
	})
	userGroup.Delete("/name", func(context *lorago.Context) {
		context.W.Write([]byte("delete amie"))
	})
	userGroup.Put("/name", func(context *lorago.Context) {
		context.W.Write([]byte("put amie"))
	})
	engine.Run()
}
