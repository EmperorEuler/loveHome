package controllers

import (
	"loveHome/models"

	"github.com/astaxie/beego"
	"github.com/beego/beego/orm"
)

type AreaController struct {
	beego.Controller
}

// 专门返回给前端的json数据函数
func (c *AreaController) RetDate(resp map[string]interface{}) {
	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *AreaController) GetArea() {
	resp := make(map[string]interface{})
	//每次结束了自动执行返回json数据
	defer c.RetDate(resp)
	//声明并初始化一个数组用来存从数据库中查询到的所有城区数据
	var areas []models.Area
	//从session拿数据

	//从数据库(mysql)拿数据

	o := orm.NewOrm()
	num, err := o.QueryTable("area").All(&areas)
	//查询失败
	if err != nil {
		beego.Info("数据错误")
		//前端要求返回json包
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}

	//查询成功但无数据 num==0
	if num == 0 {
		resp["errno"] = models.RECODE_NODATA
		resp["errmsg"] = models.RecodeText(models.RECODE_NODATA)
		return
	}
	//查询成功返回数据
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	resp["data"] = areas

	//把拿到的数据打包成json返回前端 defer做到了

}
