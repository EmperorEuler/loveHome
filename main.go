package main

import (
	_ "loveHome/routers"
	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	beego.Run()
}

