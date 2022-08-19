package controllers

import (
	"encoding/json"
	"loveHome/models"
	"loveHome/utils"
	"path"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/beego/beego/orm"
)

type HouseController struct {
	beego.Controller
}

// 专门返回给前端的json数据函数
func (c *HouseController) RetData(resp map[string]interface{}) {
	c.Data["json"] = resp
	c.ServeJSON()
}

//请求首页房源
func (this *HouseController) GetHouseIndex() {
	resp:=make(map[string]interface{})
	resp["errno"]=models.RECODE_OK
	resp["errmsg"]=models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)
	var respData []interface{}
	beego.Debug("Index Houses....")
 
	//1 从缓存服务器中请求 "home_page_data" 字段,如果有值就直接返回
	//先从缓存中获取房屋数据,将缓存数据返回前端即可
	//连接redis需要的参数信息
	redis_config_map:=map[string]string{
	   "key":"lovehome",
	   "conn":utils.G_redis_addr+":"+utils.G_redis_port,
	   "dbNum":utils.G_redis_dbnum,
	}
	//把参数信息转成json格式
	redis_config,_:=json.Marshal(redis_config_map)
	//连接redis
	cache_conn,err:=cache.NewCache("redis",string(redis_config))
	if err!=nil{
	   resp["errno"]=models.RECODE_REQERR
	   resp["errmsg"]=models.RecodeText(models.RECODE_REQERR)
	   return
	}
	//设置key
	house_page_key:="home_page_data"
	//上传房源数据到指定key中
	house_page_value:=cache_conn.Get(house_page_key)
	//返回给前端json数据
	if house_page_value!=nil{
	   beego.Debug("======= get house page info  from CACHE!!! ========")
	   json.Unmarshal(house_page_value.([]byte),&respData)
	   resp["data"]=respData
	   return
	}
 
	//2 如果缓存没有,需要从数据库中查询到房屋列表
	//取出house对象
	houses:=[]models.House{}
	o:=orm.NewOrm()
	//查询数据库中所有房子信息
	if _,err:=o.QueryTable("house").Limit(models.HOME_PAGE_MAX_HOUSES).All(&houses);err==nil{
	   //循环遍历这些房子及关联表查询
	   for _,house:=range houses{
		  //o.LoadRelated(&house, "Area")
		  //o.LoadRelated(&house, "User")
		  //o.LoadRelated(&house, "Images")
		  //o.LoadRelated(&house, "Facilities")
		  //用下面方法查到的部分房子信息追加到respData数组中
		  respData=append(respData,house.To_house_info())
	   }
	}
	//将data存入缓存中
	house_page_value,_=json.Marshal(respData)
	cache_conn.Put(house_page_key,house_page_value,3600*time.Second)
 
	//返回前端data
	resp["data"]=respData
	return
 }

func (c *HouseController) PostHouseIndex() {
	resp := make(map[string]interface{})
	//每次结束了自动执行返回json数据
	defer c.RetData(resp)
	//从前端拿到数据
	reqData := make(map[string]interface{})
	json.Unmarshal(c.Ctx.Input.RequestBody, &reqData)
	//2.判断前端数据的合法性

	//3.插入数据到数据库
	house := models.House{}
	house.Title = reqData["title"].(string)
	price, _ := strconv.Atoi(reqData["price"].(string))
	house.Price = price
	house.Address = reqData["address"].(string)
	roomcount, _ := strconv.Atoi(reqData["room_count"].(string))
	house.Room_count = roomcount
	house.Unit = reqData["unit"].(string)
	house.Beds = reqData["beds"].(string)
	minday, _ := strconv.Atoi(reqData["min_days"].(string))
	maxday, _ := strconv.Atoi(reqData["max_days"].(string))
	house.Min_days = minday
	house.Max_days = maxday
	//处理facility设备  多对多
	facilities := []models.Facility{}

	for _, fid := range reqData["facility"].([]string) {
		f_id, _ := strconv.Atoi(fid)         //拿到facid
		fac := models.Facility{Id: f_id}     //通过他的id拿到他具体是什么
		facilities = append(facilities, fac) //然后放进设备数组
	}
	//facilities 放进 house.Facilities 多对多    QueryM2Mer Add
	//创建一个 orm对象
	o := orm.NewOrm()
	//把部分house数据插入数据库中，得到 上传的houseid用来返回
	house_id, err := o.Insert(&house)
	if err != nil {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}
	//第一个参数 要改的表 而且要有主键
	m2m := o.QueryM2M(&house, "Facilities")

	//得到M2m对象后，我们就可以把刚才获取到的用户设施数组 加到 faacilities_House中
	num, errM2m := m2m.Add(facilities)
	if errM2m != nil || num == 0 {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return
	}

	// 处理城区结构体 1对多
	area_id, _ := strconv.Atoi(reqData["area_id"].(string))
	//把area_id赋值到结构体id字段中
	area := models.Area{Id: area_id}
	//再把area结构体数据赋值给house结构体中的Area
	house.Area = &area

	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	respData := make(map[string]interface{})
	respData["house_id"] = house_id
	resp["data"] = respData

}

func (this *HouseController) GetDetailHouseData()  {

	//用来存json数据的
	resp:=make(map[string]interface{})
	resp["errno"]=models.RECODE_OK
	resp["errmsg"]=models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)
 
	//1.从session获取user_id
	user_id:=this.GetSession("user_id")
 
	//2.从请求的url中得到房屋id
	//Param中的id值可以随便换，但要是router中的对应
	house_id:=this.Ctx.Input.Param(":id")
	//转换一下interface{}转成int
	h_id,err:=strconv.Atoi(house_id)
	if err!=nil{
	   resp["errno"]=models.RECODE_REQERR
	   resp["errmsg"]=models.RecodeText(models.RECODE_REQERR)
	   return
	}
	//3.从redis缓存获取当前房屋的数据,如果有该房屋，则直接返回正确的json
	redis_config_map:=map[string]string{
	   "key":"lovehome",
	   "conn":utils.G_redis_addr+":"+utils.G_redis_port,
	   "dbNum":utils.G_redis_dbnum,
	}
	redis_config,_:=json.Marshal(redis_config_map)
	cache_conn, err := cache.NewCache("redis", string(redis_config))
	if err!=nil{
	   resp["errno"]=models.RECODE_REQERR
	   resp["errmsg"]=models.RecodeText(models.RECODE_REQERR)
	   return
	}
	//先把house有的东西返回给house的json
	respData:=make(map[string]interface{})
	//设置一个变量，每个房子插入redis不能一样，容易覆盖，所以用house_id做为key，比如lovehome:1,lovehome:2
	house_page_key:=house_id
	house_info_value:=cache_conn.Get(house_page_key)
	if house_info_value!=nil{
	   beego.Debug("======= get house info desc  from CACHE!!! ========")
	   //返回json的user_id
	   respData["user_id"]=user_id
	   //返回json的house信息
	   house_info:=make(map[string]interface{})
	   //解码json并存到house_info里
	   json.Unmarshal(house_info_value.([]byte),&house_info)
	   //将house_info的map返回json的house给前端
	   respData["house"]=house_info
	   resp["data"]=respData
	   return
	}
	//4.如果缓存没有房屋数据,那么从数据库中获取数据,再存入缓存中,然后返回给前端
	o:=orm.NewOrm()
	// --- 载入关系查询 -----
	house:=models.House{Id:h_id}
	//把房子信息读出来
	if err:= o.Read(&house);err!=nil{
	   resp["errno"]=models.RECODE_DBERR
	   resp["errmsg"]=models.RecodeText(models.RECODE_DBERR)
	   return
	}
	//5.关联查询area,user,images,facilities等表
	o.LoadRelated(&house,"Area")
	o.LoadRelated(&house,"User")
	o.LoadRelated(&house,"Images")
	o.LoadRelated(&house,"Facilities")
 
 
	//6.将房屋详细的json数据存放redis缓存中
	house_info_value,_=json.Marshal(house.To_one_house_desc())
	cache_conn.Put(house_page_key,house_info_value,3600*time.Second)
 
	//7.返回json数据给前端。
	respData["house"]=house.To_one_house_desc()
	respData["user_id"]=user_id
	resp["data"]=respData
 }

func (this *HouseController) UploadHouseImage()  {
	resp:=make(map[string]interface{})
	resp["errno"]=models.RECODE_OK
	resp["errmsg"]=models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)
 
	//1.从用户请求中获取到图片数据
	fileData,hd,err:=this.GetFile("house_image")
	defer fileData.Close() //获取完后等程序执行完后关掉连接
	beego.Info("========",fileData,hd,err)
	//没拿到图片
	if fileData==nil{
	   resp["errno"]=models.RECODE_REQERR
	   resp["errmsg"]=models.RecodeText(models.RECODE_REQERR)
	   return
	}
	if err!=nil{
	   resp["errno"]=models.RECODE_REQERR
	   resp["errmsg"]=models.RecodeText(models.RECODE_REQERR)
	   return
	}
	//2.将用户二进制数据存到fdfs中。得到fileid
	suffix:=path.Ext(hd.Filename)
	//判断上传文件的合法性
	if suffix!=".jpg"&&suffix!=".png"&&suffix!=".gif"&&suffix!=".jpeg"{
	   resp["errno"]=models.RECODE_REQERR
	   resp["errmsg"]=models.RecodeText(models.RECODE_REQERR)
	   return
	}
	//去掉.
	suffixStr:=suffix[1:]
	//创建hd.Size大小的[]byte数组用来存放fileData.Read读出来的[]byte数据
	fileBuffer:=make([]byte,hd.Size)
	//读出的数据存到[]byte数组中
	_,err=fileData.Read(fileBuffer)
	if err!=nil{
	   resp["errno"]=models.RECODE_IOERR
	   resp["errmsg"]=models.RecodeText(models.RECODE_IOERR)
	   return
	}
	//将图片上传到fdfs获取到fileid
	uploadResponse,err:=UploadByBuffer(fileBuffer,suffixStr)
	//3.从请求的url中获得house_id
	house_id:=this.Ctx.Input.Param(":id")
	//4.查看该房屋的index_image_url主图是否为空
	house:=models.House{} //打开house结构体
	//house结构体拿到houseid数据
	house.Id,_=strconv.Atoi(house_id)
	o:=orm.NewOrm() //创建orm
	errRead:=o.Read(&house) //读取house数据库where user.id
	if errRead!=nil{
	   resp["errno"]=models.RECODE_DBERR
	   resp["errmsg"]=models.RecodeText(models.RECODE_DBERR)
	   return
	}
	//查询index_image_url是否为空
	//为空则更将fileid路径赋值给index_image_url
	if house.Index_image_url==""{
	   house.Index_image_url=uploadResponse.RemoteFileId
	}
	//5.主图不为空，将该图片的fileid字段追加（关联查询）到houseimage字段中插入到house_image表中,并拿到了HouseImage，里面也有数据了
	//HouseImage功能就是如果主图有了，就追加其它图片的。
	house_image:=models.HouseImage{House:&house,Url:uploadResponse.RemoteFileId}
	//将house_image和house相关联,往house.Images里追加附加图片，可以追加多个
	house.Images=append(house.Images,&house_image)//向把HouseImage对象的数据添加到house.Images
	//将house_image入库，插入到house_image表中
	if _,err:=o.Insert(&house_image);err!=nil{
	   resp["errno"]=models.RECODE_DBERR
	   resp["errmsg"]=models.RecodeText(models.RECODE_DBERR)
	   return
	}
	//将house更新入库，插入到house中
	if _,err:=o.Update(&house);err!=nil{
	   resp["errno"]=models.RECODE_DBERR
	   resp["errmsg"]=models.RecodeText(models.RECODE_DBERR)
	   return
	}
	//6.拼接完整域名url_fileid
	respData:=make(map[string]string)
	respData["url"]=utils.AddDomain2Url(uploadResponse.RemoteFileId)
 
	//7.返回给前端json
	resp["data"]=respData
 }

 func (this *HouseController) GetHouseSearchData()  {
	resp:=make(map[string]interface{})
	resp["errno"]=models.RECODE_OK
	resp["errmsg"]=models.RecodeText(models.RECODE_OK)
	defer this.RetData(resp)
	//1.获取用户发来的参数，aid,sd,ed,sk,p
	var aid int
	this.Ctx.Input.Bind(&aid,"aid")
	var sd string
	this.Ctx.Input.Bind(&sd,"sd")
	var ed string
	this.Ctx.Input.Bind(&ed,"ed")
	var sk string
	this.Ctx.Input.Bind(&sk,"sk")
	var page int
	this.Ctx.Input.Bind(&page,"p")
	//2.检验开始时间一定要早于结束时间
	//将日期转成指定格式
	start_time,_:=time.Parse("2006-01-02 15:04:05",sd+" 00:00:00")
	end_time,_:=time.Parse("2006-01-02 15:04:05",ed+" 00:00:00")
	if end_time.Before(start_time){ //如果end在start之前,返回错误信息
		resp["errno"]=models.RECODE_REQERR
		resp["errmsg"]="结束时间必须在开始时间之前"
		return
	}
	
	//3.判断p的合法性，一定要大于0的整数
	if page<=0{
		resp["errno"]=models.RECODE_REQERR
		resp["errmsg"]="页数不能小于或等于0"
		return
	}
	//4.尝试从缓存中获取数据，返回查询结果json
	/定义一个key,注意这个存入redis中的key值拼接字符串，一定要用strconv.Itoa()转换，不要用string(),否则会出现\x01的效果,读取不了
	house_search_key:="house_search_"+strconv.Itoa(aid)
	//配置redis连接信息
	redis_config_map:=map[string]string{
		"key":"lovehome",
		"conn":utils.G_redis_addr+":"+utils.G_redis_port,
		"dbNum":utils.G_redis_dbnum,
	}
	//转成json
	redis_config,_:=json.Marshal(redis_config_map)
	//连接redis
	cache_conn, err := cache.NewCache("redis", string(redis_config))
	if err!=nil{
		resp["errno"]=models.RECODE_REQERR
		resp["errmsg"]=models.RecodeText(models.RECODE_REQERR)
		return
	}
	//从redis中拿到数据
	house_search_info:=cache_conn.Get(house_search_key)

	if house_search_info!=nil{
		beego.Debug("======= get house_search_info  from CACHE!!! ========")
		//存解码后的数据
		house_info:=[]map[string]interface{}{}
		//解码json数据
		json.Unmarshal(house_search_info.([]byte),&house_info)
		//把解码后的数据打包成json传给前端
		respData["houses"]=house_info
		respData["total_page"]=10
		respData["current_page"]=1
		resp["data"]=respData
		return
	}
	//5.如果缓存中没有数据，从数据库中查询
	
	//（此处过于复杂，可以暂时以发布时间顺序查询）
	//指定查询的表
	houses:=[]models.House{}
	o:=orm.NewOrm()
	//查询house表
	qs:=o.QueryTable("house")
	//查询指定城区的所有房源，按发布时间降序排列
	num,err:=qs.Filter("area_id",aid).OrderBy("-ctime").All(&houses)
	if err!=nil{
		resp["errno"]=models.RECODE_DBERR
		resp["errmsg"]=models.RecodeText(models.RECODE_DBERR)
		return
	}
	//求出所有分页
	total_page:=int(num)/models.HOUSE_LIST_PAGE_CAPACITY+1
	//起始页数
	house_page:=1
	//用来存遍历到的房屋数据
	var house_list []interface{}
	//遍历出上面查到的房屋数据，加到数组house_list中
	for _,house:=range houses  {
		o.LoadRelated(&house, "Area")
		o.LoadRelated(&house, "User")
		o.LoadRelated(&house, "Images")
		o.LoadRelated(&house, "Facilities")
		house_list=append(house_list,house.To_house_info())
	}
	//拿到了houst_list数据
	fmt.Println("========house_list======",house_list)
	//6.将查询条件存储到缓存
	houst_search_list,_:=json.Marshal(house_list)
	cache_conn.Put(house_search_key,houst_search_list,3600*time.Second)
	//7.返回查询结果json数据给前端
	respData["houses"]=house_list
	respData["total_page"]=total_page
	respData["current_page"]=house_page
	resp["data"]=respData
	return
 }