package controllers

import (
	"encoding/json"
	"loveHome/models"
	"time"

	"github.com/astaxie/beego"
	"github.com/beego/beego/orm"

	"github.com/astaxie/beego/cache"
	_ "github.com/astaxie/beego/cache/redis"
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

	//连接redis
	cache_conn, err := cache.NewCache("redis", `{"key":"lovehome","conn":":6379","dbNum":"0"}`)
	// //put一个key：aaa，value:bbb到redis，失效时间为3600s
	// errCache := cache_conn.Put("aaa", "bbb", time.Second*3600)
	// if errCache != nil {
	// 	beego.Error("cache err ===", errCache)
	// }
	if err != nil {
		beego.Error("cache_conn err===", err)
		resp["errno"] = models.RECODE_DATAERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DATAERR)
	}
	// //从redis上get信息， 返回值是【】byte数组bbb为【989898】
	// beego.Info("cache_conn.aaa ===", cache_conn.Get("aaa"))

	//1.从redis拿数据
	if areaDate := cache_conn.Get("area"); areaDate != nil {
		beego.Info("get from redis")
		resp["errno"] = models.RECODE_OK
		resp["errmsg"] = models.RecodeText(models.RECODE_OK)
		// !!!!要把从redis中取来的数据必须先解码才能在前台显示
		var areas_info interface{}
		//!!!!!解码数据    且     存入info中
		json.Unmarshal(areaDate.([]byte), &areas_info)
		resp["data"] = areas_info
		return
	}

	//2.(如果redis没有数据）从数据库(mysql)拿数据

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

	//把数据转换成json存入缓存（redis）
	json_str, err := json.Marshal(areas)
	if err != nil {
		beego.Info("encoding err")
		return
	}
	cache_conn.Put("area", json_str, time.Second*3600)

	//把拿到的数据打包成json返回前端 defer做到了

}
