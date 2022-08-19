package controllers

import (
	"encoding/json"
	"loveHome/models"

	"github.com/astaxie/beego"
	"github.com/beego/beego/orm"
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

// 注册完的登录状态
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

// 登录
func (c *SessionController) Login() {
	//1.得到用户信息
	resp := make(map[string]interface{})
	defer c.RetDate(resp)
	//获取前端传过来的json数据
	json.Unmarshal(c.Ctx.Input.RequestBody, &resp)
	//2.判断是否合法
	if resp["mobile"] == nil || resp["password"] == nil {
		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)
		return
	}
	// else if len(resp["mobile"].(string)) != 11 {
	// 	resp["errno"] = models.RECODE_DATAERR
	// 	resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)
	// 	return
	// } //简单的判断是否为空和长度

	//3.与数据库匹配，判断账号密码是否正确
	o := orm.NewOrm()
	user := models.User{Name: resp["mobile"].(string)}
	//查询user表
	//因为名字会改 所以修改为按手机查询
	qs := o.QueryTable("mobile")
	//过滤只查询mobile == user.name的，one（&user）返回数据到user结构体中，记得用地址
	err := qs.Filter("mobile", user.Mobile).One(&user)
	if err != nil {
		beego.Info("o.Read(&user) err=====", err)
		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)
		return
	}
	//判断密码是否一致
	if user.Password_hash != resp["password"] {

		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)
		return
	}
	//4.成功后添加session 保持登录状态
	c.SetSession("name", user.Name)
	c.SetSession("mobile", resp["mobile"])
	c.SetSession("user_id", user.Id)

	//5.返回json数据给前端
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
}
