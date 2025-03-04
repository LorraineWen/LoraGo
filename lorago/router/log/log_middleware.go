package log

import (
	"fmt"
	"github.com/LorraineWen/lorago/router"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

/*
*@Author: LorraineWen
*支持日志中间件
*支持不同颜色的日志
*支持日志格式自定义
*支持分级日志，比如error级别的日志，info级别的日志，debug级别的日志
 */
//各种颜色
const (
	greenBg   = "\033[97;42m"
	whiteBg   = "\033[90;47m"
	yellowBg  = "\033[90;43m"
	redBg     = "\033[97;41m"
	blueBg    = "\033[97;44m"
	magentaBg = "\033[97;45m"
	cyanBg    = "\033[97;46m"
	green     = "\033[32m"
	white     = "\033[37m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	reset     = "\033[0m"
)

type LoggerConfig struct {
	Formatter LoggerFormatter //支持格式化输出
	out       io.Writer
}

// 支持日志格式化输出
var DefaultWriter io.Writer = os.Stdout

type LoggerFormatter func(params LogFormatterParams) string
type LogFormatterParams struct {
	Request    *http.Request
	TimeStamp  time.Time
	StatusCode int
	Latency    time.Duration
	ClientIP   net.IP
	Method     string
	Path       string
	isColorful bool //设置是否需要颜色，在控制台输出可以设置为true，如果是将日志输出到文件中，就要变为false,否则就会将颜色字符也写到文件里面
}

// 支持日志颜色
func (p *LogFormatterParams) StatusCodeColor() string {
	code := p.StatusCode
	switch code {
	case http.StatusOK:
		return green
	default:
		return red
	}
}
func (p *LogFormatterParams) ResetColor() string {
	return reset
}

var defaultLogFormatter = func(params LogFormatterParams) string {
	statusCodeColor := params.StatusCodeColor()
	resetColor := params.ResetColor()
	if params.Latency > time.Minute {
		params.Latency = params.Latency.Truncate(time.Second)
	}
	//启用颜色
	if params.isColorful {
		return fmt.Sprintf("%s [lorago] %s |%s %v %s| %s %3d %s |%s %13v %s| %15s  |%s %-7s %s %s %#v %s\n",
			yellow, resetColor, blue, params.TimeStamp.Format("2006/01/02 - 15:04:05"), resetColor,
			statusCodeColor, params.StatusCode, resetColor,
			red, params.Latency, resetColor,
			params.ClientIP,
			magenta, params.Method, resetColor,
			cyan, params.Path, resetColor,
		)
	} else {
		return fmt.Sprintf("[msgo] %v | %3d | %13v | %15s |%-7s %#v",
			params.TimeStamp.Format("2006/01/02 - 15:04:05"),
			params.StatusCode,
			params.Latency, params.ClientIP, params.Method, params.Path,
		)
	}

}

// 日志中间件
// 调用方式routerGroup.Use(lorago.LogMiddleware)
func LoggerWithConfig(conf LoggerConfig, next router.HandleFunc) router.HandleFunc {
	formatter := conf.Formatter
	if formatter == nil {
		formatter = defaultLogFormatter
	}
	out := conf.out
	var isColor bool = false
	//如果是标准输出，那么就启用颜色
	if out == nil {
		out = DefaultWriter
		isColor = true
	}
	return func(ctx *router.Context) {
		param := LogFormatterParams{
			Request: ctx.R,
		}
		// Start timer
		start := time.Now()
		path := ctx.R.URL.Path
		raw := ctx.R.URL.RawQuery
		//执行业务
		next(ctx)
		// stop timer
		stop := time.Now()
		latency := stop.Sub(start)
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(ctx.R.RemoteAddr))
		clientIP := net.ParseIP(ip)
		method := ctx.R.Method
		statusCode := ctx.StatusCode

		if raw != "" {
			path = path + "?" + raw
		}

		param.ClientIP = clientIP
		param.TimeStamp = stop
		param.Latency = latency
		param.StatusCode = statusCode
		param.Method = method
		param.Path = path
		param.isColorful = isColor
		fmt.Fprint(out, formatter(param))
	}
}
func LogMiddleware(next router.HandleFunc) router.HandleFunc {
	return LoggerWithConfig(LoggerConfig{}, next)
}
