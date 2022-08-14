package controllers

import "github.com/astaxie/beego"

type HouseIndexController struct {
	beego.Controller
}

// 专门返回给前端的json数据函数
func (c *HouseIndexController) RetDate(resp map[string]interface{}) {
	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *HouseIndexController) GetHouseIndex() {
	resp := make(map[string]interface{})
	//每次结束了自动执行返回json数据
	defer c.RetDate(resp)
	resp["errno"] = 4001
	resp["errmsg"] = "查询失败"
}
