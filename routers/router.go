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
	beego.Router("/api/v1.0/houses", &controllers.HouseController{}, "post:PostHouseData")
	beego.Router("/api/v1.0/houses/?:id/images", &controllers.HouseController{}, "post:UploadHouseImage")
	beego.Router("/api/v1.0/houses/?:id", &controllers.HouseController{}, "get:GetDetailHouseData")
	beego.Router("/api/v1.0/user/orders", &controllers.OrderController{}, "get:GetOrderData")
	beego.Router("/api/v1.0/orders", &controllers.OrderController{}, "post:PostOrderHouseData")
	beego.Router("/api/v1.0/orders/:id/status", &controllers.OrderController{}, "put:OrderStatus")
	beego.Router("/api/v1.0/orders/:id/comment", &controllers.OrderController{}, "put:OrderComment")
}
