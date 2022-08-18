package routers

import (
	"loveHome/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{}, "get:Get")
	beego.Router("/api/v1.0/areas", &controllers.AreaController{}, "get:GetArea")
	beego.Router("/api/v1.0/houses/index", &controllers.HouseIndexController{}, "get:GetHouseIndex")
	beego.Router("/api/v1.0/session", &controllers.SessionController{}, "get:GetSessionData;delete:DeleteSessionData")
	beego.Router("/api/v1.0/users", &controllers.UserController{}, "post:Reg")
	beego.Router("/api/v1.0/sessions", &controllers.SessionController{}, "post:Login")
	beego.Router("/api/v1.0/user/avatar", &controllers.UserController{}, "post:PostAvatar")
	beego.Router("/api/v1.0/user", &controllers.UserController{}, "get:GetUserData")
	beego.Router("/api/v1.0/user/name", &controllers.UserController{}, "put:UpdateUserName")
	beego.Router("/api/v1.0/user/auth", &controllers.UserController{}, "get:GetUserData;post:PostRealName")
	beego.Router("/api/v1.0/user/houses", &controllers.UserController{}, "get:GetHouseData")

}
