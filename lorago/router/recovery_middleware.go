package router

import (
	"errors"
	"fmt"
	"github.com/LorraineWen/lorago/router/lora_error"
	"net/http"
	"runtime"
	"strings"
)

/*
*@Author: LorraineWen
*定义panic捕获中间件，当路由产生panic时，在这里捕获，并且打印错误信息
*支持错误产生定位
 */
func RecoveryMiddleware(next HandleFunc) HandleFunc {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				if e := err.(error); e != nil {
					var loraError *lora_error.LoraError
					if errors.As(e, &loraError) {
						loraError.ExecResult()
					}
				}
				ctx.Logger.Error(detailMsg(err))
				ctx.Fail(http.StatusInternalServerError, "recovery:Internal Server Error")
			}
		}()
		next(ctx)
	}
}

// 打印报错代码的详细信息
func detailMsg(err any) string {
	var sb strings.Builder
	var pcs = make([]uintptr, 32)
	n := runtime.Callers(3, pcs) //关键函数就在这里
	sb.WriteString(fmt.Sprintf("%v\n", err))
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		sb.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return sb.String()
}
