package controllers

import (
	"loveHome/models"

	"github.com/astaxie/beego"
)

// 两种方法 1.main入口设置 beego.BConfig.WebConfig.Session.SessionOn = true
//
//	2.配置文件 conf/app.conf 配置  sessionon = true
type SessionController struct {
	beego.Controller
}

// 专门返回给前端的json数据函数
func (c *SessionController) RetDate(resp map[string]interface{}) {
	c.Data["json"] = resp
	c.ServeJSON()
}

// 退出登录
func (c *SessionController) DeleteSessionData() {
	resp := make(map[string]interface{})
	defer c.RetDate(resp)
	c.DelSession("name")
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
}

func (c *SessionController) GetSessionData() {
	resp := make(map[string]interface{})
	//每次结束了自动执行返回json数据
	defer c.RetDate(resp)
	//获取user结构体对象
	user := models.User{}

	resp["errno"] = models.RECODE_DBERR
	resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)

	//获取session
	name := c.GetSession("name")
	//判断如果获取到session需要做什么
	if name != nil {
		user.Name = name.(string)
		resp["errno"] = models.RECODE_OK
		resp["errmsg"] = models.RecodeText(models.RECODE_OK)
		//把user结构体数据传给data
		resp["data"] = user
	}
}
