package controllers

import { 

}

type OrderController struct {
	beego.Controller
}

type OrderRequest struct {
	House_id   string `json:"house_id"`   //下单的房源id
	Start_date string `json:"start_date"` //订单开始时间
	End_date   string `json:"end_date"`   //订单结束时间
 }

// 专门返回给前端的json数据函数
func (c *OrderController) RetData(resp map[string]interface{}) {
	c.Data["json"] = resp
	c.ServeJSON()
}

func (this *OrderController)GetOrderData(){
	resp:=make(map[string]interface{})
	resp["errno"]=models.RECODE_OK
	resp["errmsg"]=models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)

	//1.根据session获取userid
	user_id:=this.GetSession("user_id").(int)
	//2.根据url获得当前操作角色
	role:=this.GetString("role")

	//如果role为空，则返回json错误信息
	if role==""{
		resp["errno"]=models.RECODE_ROLEERR
		resp["errmsg"]=models.RecodeText(models.RECODE_ROLEERR)
		return
	}
	if role=="landlord"{
		beego.Info("我是房东")
		landLordHouses:=[]models.House{}
		//把房东的房源信息全都查出来
		o.QueryTable("house").Filter("user__id",user_id).All(&landLordHouses)
		//int数组专门存房源id
		housesIds:=[]int{}
		//遍历所以房源信息
		for _,house:=range landLordHouses{
		   //把房子id追加到houseIds数组中，拿到houseIds数组
		   housesIds=append(housesIds,house.Id)
		}
		//在订单表中找到房东house_id为自己房源的id的所有房子
		//house__id__in相当于select  *  from  order_house  where  house.id  in('1','2','3');
		//意思就是查询house.id=1，2，3的都查出来。
		//切记，Filter里一定要用双下划线.
		o.QueryTable("order_house").Filter("house__id__in",housesIds).OrderBy("-ctime").All(&orders)
	 	}else{
		//如果是租客，查询房屋订单表找到自己发布的全部订单
			beego.Info("我是租客")
			o.QueryTable("order_house").Filter("user__id",user_id).OrderBy("-ctime").All(&orders)
		 }
	 //关联house和user表，查询出所有订单数据，追加到order_list数组中，需要返回给前端的数据
	 	for _,order:=range orders{
		 	o.LoadRelated(&order,"House")
		 	o.LoadRelated(&order,"User")
			order_list=append(order_list,order.To_order_info())
		}
	


}

func (this *OrderController) PostOrderHouseData()  {
	resp:=make(map[string]interface{})
	resp["errno"]=models.RECODE_OK
	resp["errmsg"]=models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)
	//1.根据session得到用户id
	user_id:=this.GetSession("user_id")
	//2.得到用户请求的json数据，检测合法性，不合法就返回错误json数据
	var req OrderRequest
	if err:=json.Unmarshal(this.Ctx.Input.RequestBody,&req);err!=nil{
		resp["errno"]=models.RECODE_REQERR
		resp["errmsg"]=models.RecodeText(models.RECODE_REQERR)
		return
	}
	fmt.Printf("req = %+v\n", req)
	//用户请求参数做合法判断
	if req.House_id==""||req.Start_date==""||req.End_date==""{
		resp["errno"]=models.RECODE_REQERR
		resp["errmsg"]="请求参数为空"
		return
	}
	//3.确定退房时间end_date必须在订房时间start_date之后
	//格式化日期时间
start_date_time,_:=time.Parse("2006-01-02 15:04:05",req.Start_date+" 00:00:00")
end_date_time,_:=time.Parse("2006-01-02 15:04:05",req.End_date+" 00:00:00")
//确保end_date 在 start_date之后
if end_date_time.Before(start_date_time){ //如果end在start之前,返回错误信息
   resp["errno"]=models.RECODE_REQERR
   resp["errmsg"]="结束时间在开始时间之前"
   return
}
fmt.Printf("start_date_time = %v,end_date_time = %v",start_date_time,end_date_time)
	//可以用一个函数
 
	/*
	a1:=time.Now()
	a2:=time.Now()
	a1的时候在a2之前吗，返回bool
	a1.Before(a2)
	*/
 
	//3.得到一共入住的天数
	//意思就是获取end_date_time距离start_date_time共几天
	days:=end_date_time.Sub(start_date_time).Hours()/24+1
	fmt.Printf("days = %f\n",days)
	//4.根据order_id获取关联的房源信息
	house_id,_:=strconv.Atoi(req.House_id)
	//操作house表，where house_id
	house:=models.House{Id:house_id}
	o:=orm.NewOrm()
	//读取house表中的当前房源信息
	if err:=o.Read(&house);err!=nil{
	resp["errno"]=models.RECODE_NODATA
	resp["errmsg"]=models.RecodeText(models.RECODE_NODATA)
	return
	}
	//5.确保当前的user_id不是房源信息所关联的user_id（就是说房东不能自己预定自己房子）
	//房东不能预定自己的房子
	if user_id==house.User.Id{
		resp["errno"]=models.RECODE_ROLEERR
		resp["errmsg"]=models.RecodeText(models.RECODE_ROLEERR)
		return
	}
	//6.确保用户选择的房屋未被预定，日期没有冲突，如果已经被人预定返回错误信息
 
	//（此处逻辑过于复杂，可以先不判断）
 
	//7.封装完整的order订单信息
	//给orderHouse结构体数据
	//这里获取user,house结构体，才能从这两个表中取数据，前面已经关联好数据了，所以这两个结构体里是有数据的
	amount:=days * float64(house.Price) //天数 *钱数,返回订单总金额

	order:=models.OrderHouse{} //操作房源订单表
	user:=models.User{Id:user_id.(int)} //操作User数据

	order.House=&house //order获取到house数据
	order.User=&user //order获得user数据
	order.Begin_data=start_date_time //起始时间
	order.End_data=end_date_time //结束时间
	order.Days=int(days) //总天数
	order.House_price=house.Price //房间间价
	order.Amount=int(amount) //订单总金额
	order.Status=models.ORDER_STATUS_WAIT_ACCEPT //订单状态
	fmt.Printf("order = %+v\n",order)
	//8.将订单信息写入表中
	if _,err:=o.Insert(&order);err!=nil{
		resp["errno"]=models.RECODE_DBERR
		resp["errmsg"]=models.RecodeText(models.RECODE_DBERR)
		return
	 }
	//9.返回order_id的json给前端
	this.SetSession("user_id",user_id)
	respData:=make(map[string]interface{})
	respData["order_id"]=order.Id
	resp["data"]=respData
	return
}

func (this *OrderController) OrderStatus()  {
	resp:=make(map[string]interface{})
	resp["errno"]=models.RECODE_OK
	resp["errmsg"]=models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)
	//1.根据session拿到user_id
	user_id:=this.GetSession("user_id")
	//2.通过当前url参数得到当前订单id
	order_id:=this.Ctx.Input.Param(":id")
	//3.解析客户端请求的json数据，得到action数据
	//用来存获取到的请求数据action,reason
	var req map[string]interface{}
	if err:=json.Unmarshal(this.Ctx.Input.RequestBody,&req);err!=nil{
		resp["errno"]=models.RECODE_REQERR
		resp["errmsg"]=models.RecodeText(models.RECODE_REQERR)
		return
	}
	//4.检验action是否合法，不合法，返回json错误信息
	action:=req["action"]
	if action!="accept" && action!="reject"{
		resp["errno"]=models.RECODE_REQERR
		resp["errmsg"]=models.RecodeText(models.RECODE_REQERR)
		return
	}
	//5.查找订单表，找到该订单并确定当前的订单状态是WAIT_ACCEPT
	o:=orm.NewOrm()
	order:=models.OrderHouse{}
	//查到订单状态数据是否为待接单，如果是就查询出这个房源的订单信息
	if err:=o.QueryTable("order_house").Filter("id",order_id).Filter("status",models.ORDER_STATUS_WAIT_ACCEPT).One(&order);err!=nil{
	   resp["errno"]=models.RECODE_DBERR
	   resp["errmsg"]=models.RecodeText(models.RECODE_DBERR)
	   return
	}
	//6.检验该订单的user_id是否是当前用户的user_id
/*
把order表和house表做关联，才能检验订单,必须做关联，不做关联查询不到数据，会报错
[C] [asm_amd64.s:573] Handler crashed with error runtime error: invalid memory address or nil pointer dereference
解释：程序在运行时候出错：内存地址无效，空指针被回收
*/
	if _,err:=o.LoadRelated(&order,"House");err!=nil{
		resp["errno"]=models.RECODE_DATAERR
		resp["errmsg"]=models.RecodeText(models.RECODE_DATAERR)
		return
	}
	house:=order.House
	fmt.Printf("house = %+v\n", house)
	fmt.Printf("house.user_id = %d\n", house.User.Id)
	if house.User.Id!=user_id{
		resp["errno"]=models.RECODE_DATAERR
		resp["errmsg"]="订单用户不匹配,操作无效"
		return
	}
	//7.判断action
	if action=="accept"{
		//如果action为accept(接单)，更换该订单status为WAIT_COMMENT等待用户评价
		order.Status=models.ORDER_STATUS_WAIT_COMMENT
		beego.Debug("action = accpet!")
		beego.Debug("order.Status = ",order.Status)
	 }else if action=="reject"{
		//如果action为reject(拒单),更换订单status为REJECT，然后从url参数获取reason参数字段 ，将reason字段的value添加到order的评价comment字段中
		order.Status=models.ORDER_STATUS_REJECTED
		reason:=req["reason"]
		order.Comment=reason.(string)
		beego.Debug("action = reject!, reason is ", reason)
		beego.Debug("order.Comment = ",order.Comment)
	 }
	//上面都完成后，更新该订单数据到数据库中。o.Update()
	if _,err:=o.Update(&order);err!=nil{
		resp["errno"]=models.RECODE_DBERR
		resp["errmsg"]=models.RecodeText(models.RECODE_DBERR)
		return
	 }
	//8.返回正确json给前端

}

func (this *OrderController) OrderComment()  {
	resp:=make(map[string]interface{})
	resp["errno"]=models.RECODE_OK
	resp["errmsg"]=models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)
	//1.通过session获得用户id
	user_id:=this.GetSession("user_id").(int)
	//2.通过请求url获取order_id.
	order_id:=this.Ctx.Input.Param(":id")
	//3.获取前端发过来的请求参数的value
	var req map[string]interface{}
	if err:=json.Unmarshal(this.Ctx.Input.RequestBody,&req);err!=nil{
		resp["errno"]=models.RECODE_REQERR
		resp["errmsg"]=models.RecodeText(models.RECODE_REQERR)
		return
	}
	//4.检验评价信息是否合法，确保不为空
	comment:=req["comment"].(string)
	if comment==""{
		resp["errno"]=models.RECODE_PARAMERR
		resp["errmsg"]=models.RecodeText(models.RECODE_PARAMERR)
		return
	}
	//5.查询数据库，订单必须存在，订单状态必须为WAI_COMMENT待评价状态。
	order:=models.OrderHouse{} //获取orderhouse表，把查到的数据存这里
	o:=orm.NewOrm()
	if err:=o.QueryTable("order_house").Filter("id",order_id).Filter("status",models.ORDER_STATUS_WAIT_COMMENT).One(&order);err!=nil{
		resp["errno"]=models.RECODE_DBERR
		resp["errmsg"]=models.RecodeText(models.RECODE_DBERR)
		return
	}
	//6.确保订单所关联的用户和该用户是同一个人
	// 关联查询order订单所关联的user信息
	if _,err:=o.LoadRelated(&order,"User");err!=nil{
		resp["errno"]=models.RECODE_DBERR
		resp["errmsg"]=models.RecodeText(models.RECODE_DBERR)
		return
	}
	//如果不是本人，就报错
	if user_id!=order.User.Id{
		resp["errno"]=models.RECODE_DATAERR
		resp["errmsg"]="该订单并不属于本人"
		return
	}
	//7.关联查询order订单所关联的House信息
	if _,err:=o.LoadRelated(&order,"House");err!=nil{
		resp["errno"]=models.RECODE_DBERR
		resp["errmsg"]=models.RecodeText(models.RECODE_DBERR)
		return
	}

	//拿到order订单对应的House所有数据
	house:=order.House
	//8.更新order的status为COMPLETE,更新order的comment信息
	order.Status=models.ORDER_STATUS_COMPLETE
	order.Comment=comment
	//9.订房成功，将房屋订单成交量+1
	house.Order_count++
	//10.将order和house完整数据更新到数据库中,指定只更新status,comment字段的数据
	//cols意思是，只更新status和comment的数据，其它字段不更新
	if _,err:=o.Update(&order,"status","comment");err!=nil{
		beego.Error("update order status, comment error, err = ", err)
		resp["errno"]=models.RECODE_DBERR
		resp["errmsg"]=models.RecodeText(models.RECODE_DBERR)
		return
	}
	//更新house表的时候不需要指针，因为不是传的对象,上面是对象，所以需要传指针
	if _,err:=o.Update(house,"order_count");err!=nil{
		beego.Error("update house order_count error, err = ", err)
		resp["errno"]=models.RECODE_DBERR
		resp["errmsg"]=models.RecodeText(models.RECODE_DBERR)
		return
	}
	//11.将house_info_[house_id]存的的redis的key删除（因为已经修改了订单数量）
	//设置redis配置信息

	redis_config_map:=map[string]string{
		"key":"lovehome",
		"conn":utils.G_redis_addr+":"+utils.G_redis_port,
		"dbNum":utils.G_redis_dbnum,
	}
	//把redis信息打包成json
	redis_config,_:=json.Marshal(redis_config_map)
	//连接redis
	cache_conn,err:=cache.NewCache("redis",string(redis_config))
	if err!=nil{
		beego.Debug("connect cache error")
	}
	//key指定为house.Id,因为redis存的时候用的就是house.Id
	house_info_key:=strconv.Itoa(house.Id)
	//删除
	if err:=cache_conn.Delete(house_info_key);err!=nil{
		beego.Error("delete",house_info_key,"error,err = ",err)
	}
	return
 }