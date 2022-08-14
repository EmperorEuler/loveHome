package main

import (
	_ "loveHome/models"
	_ "loveHome/routers"
	"net/http"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func main() {
	ignoreStaticPath()
	beego.BConfig.WebConfig.Session.SessionOn = true
	beego.Run()
}

// 把原始网址补全
func ignoreStaticPath() {
	//透明static
	beego.InsertFilter("/", beego.BeforeRouter, TransparentStatic)
	beego.InsertFilter("/*", beego.BeforeRouter, TransparentStatic)
}

// 传进来一个ctx参数 获取当前窗口句柄
func TransparentStatic(ctx *context.Context) {
	//获取当前窗口路径  如/index.html
	orpath := ctx.Request.URL.Path
	beego.Debug("request url:", orpath)
	//如果请求url还有api字段，说明指令应该取消静态资源路径重定向
	if strings.Index(orpath, "api") >= 0 {
		return
	}
	//访问的应该是static/html/index.html
	http.ServeFile(ctx.ResponseWriter, ctx.Request, "static/html/"+ctx.Request.URL.Path)
}
