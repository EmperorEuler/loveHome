package controllers

import (
	"encoding/json"
	"loveHome/models"

	"github.com/astaxie/beego"
	"github.com/beego/beego/orm"
)

type UserController struct {
	beego.Controller
}

// 专门返回给前端的json数据函数
func (c *UserController) RetDate(resp map[string]interface{}) {
	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *UserController) Reg() {
	resp := make(map[string]interface{})
	defer c.RetDate(resp)
	//获取前端接受到的账号密码数据（json）转换成想要的类型（map）要先设配置
	json.Unmarshal(c.Ctx.Input.RequestBody, &resp)
	beego.Info("========", resp["mobile"])
	beego.Info("========", resp["password"])
	beego.Info("========", resp["sms_code"])
	//将数据插入数据库
	o := orm.NewOrm()
	user := models.User{} //获取user数据库结构体，啥数据不用传
	user.Password_hash = resp["password"].(string)
	user.Name = resp["mobile"].(string)
	user.Mobile = resp["mobile"].(string)
	////插入
	id, err := o.Insert(&user)
	if err != nil {
		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)
		return
	}
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	beego.Info("reg succee,id===", id)

	//注册完显示一个名字出来 =》设置一个session
	c.SetSession("name", user.Name)
	//打包发给前端

}
