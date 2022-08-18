package controllers

import (
	"encoding/json"
	"fdfs_client"
	"loveHome/models"
	"path"

	"loveHome/fdfs_client"

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
	c.SetSession("user_id", user.Id)
	c.SetSession("mobile", user.Mobile)
	//打包发给前端

}

func (c *UserController) PostAvatar() {
	resp := make(map[string]interface{})
	defer c.RetDate(resp)
	//1.获取上传的文件
	fileData, hd, err := c.GetFile("avatar")
	if err != nil {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		return
	}
	//2.得到文件后缀
	suffix := path.Ext(hd.Filename)
	//3.存储文件到fdfs上
	fdfsClient, err := fdfs_client.NewFdfsClient("../conf/client.conf")
	if err != nil {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		return
	}
	fileBuffer := make([]byte, hd.Size)
	_, errBuffer := fileData.Read(fileBuffer)
	if errBuffer != nil {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		return
	}

	DataResponse, err := fdfsClient.UploadAppenderByBuffer(fileBuffer, suffix[1:])
	if err != nil {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		return
	}

	//4.查从session拿user_id
	user_id := c.GetSession("user_id")
	//5.更新用户数据库内容
	var user models.User
	o := orm.NewOrm()
	qs := o.QueryTable("user")
	qs.Filter("Id", user_id).One(&user)
	user.Avatar_url = DataResponse.RemoteFileId

	_, errUpdate := o.Update(&user)
	if errUpdate != nil {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		return
	}
	urlMap := make(map[string]string)
	urlMap["avatar_url"] = "192.168.58.129:8080/" + DataResponse.RemoteFileId
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	resp["data"] = urlMap
}

func (c *UserController) GetUserData() {
	resp := make(map[string]interface{})
	defer c.RetDate(resp)
	//1.从session获取userid
	user_id := c.GetSession("user_id")
	//2.从数据库中拿到userid的信息
	user := models.User{Id: user_id.(int)}
	o := orm.NewOrm()
	err := o.Read(&user)
	if err != nil {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}

	resp["data"] = &user
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
}

func (c *UserController) UpdateUserName() {
	resp := make(map[string]interface{})
	defer c.RetDate(resp)

	//1.获得session中的userid
	user_id := c.GetSession("user_id")
	//2.获取前端传的数据
	UserName := make(map[string]string)
	json.Unmarshal(c.Ctx.Input.RequestBody, &UserName)
	//3.更新userid 对应的 name
	o := orm.NewOrm()
	user := models.User{Id: user_id.(int)}
	if err := o.Read(&user); err != nil {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	user.Name = UserName["name"]

	_, err := o.Update(&user)
	if err != nil {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	//4.把session中的name字段更新
	c.SetSession("name", user.Name)
	//5.把数据打包
	resp["data"] = UserName
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
}

func (c *UserController) PostRealName() {
	resp := make(map[string]interface{})
	defer c.RetDate(resp)

	//1.获得session中的userid
	user_id := c.GetSession("user_id")
	//2.获取前端传的数据
	RealName := make(map[string]string)
	json.Unmarshal(c.Ctx.Input.RequestBody, &RealName)
	//3.更新数据库userid对应信息 先读后更改
	o := orm.NewOrm()
	user := models.User{Id: user_id.(int)}
	if err := o.Read(&user); err != nil {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}

	user.Real_name = RealName["real_name"]
	user.Id_card = RealName["id_card"]
	_, err := o.Update(&user)
	if err != nil {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	c.SetSession("user_id", user.Id)
	//4.打包
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
}
