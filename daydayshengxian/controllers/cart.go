package controllers

import (
	"github.com/astaxie/beego"
	"github.com/gomodule/redigo/redis"
	"github.com/astaxie/beego/orm"
	"daydayshengxian/models"
	"strconv"
)

type Cart struct {
	beego.Controller
}

func (c *Cart)Handleaddcart()  {
	skuid,err1:=c.GetInt("skuid")
	count,err2:=c.GetInt("count")
	resp:=make(map[string]interface{})
	defer c.ServeJSON()
	if err1!=nil||err2!=nil {
		resp["code"]=1
		resp["msg"]="json数据错误"
		c.Data["json"]=resp
		return
	}
	username:=c.GetSession("username")
	if username==nil {
		resp["code"]=2
		resp["msg"]="未登陆"
		c.Data["json"]=resp
		return
	}
	o:=orm.NewOrm()
	var user models.User
	user.Name=username.(string)
	o.Read(&user,"Name")
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil {
		beego.Info("redis连接失败")
		return
	}
	precount,err:=redis.Int(conn.Do("hget","cart_"+strconv.Itoa(user.Id),skuid))
	conn.Do("hset","cart_"+strconv.Itoa(user.Id),skuid,count+precount)
	rep,err:=conn.Do("hlen","cart_"+strconv.Itoa(user.Id))
	cartcount,_:=redis.Int(rep,err)
	resp["code"]=5
	resp["msg"]="ok"
	resp["cartcount"]=cartcount
	c.Data["json"]=resp
	c.ServeJSON()
}

func Getcartcount(c *beego.Controller)int  {
	username:=c.GetSession("username")
	if username==nil {
		return 0
	}
	o:=orm.NewOrm()
	var user models.User
	user.Name=username.(string)
	o.Read(&user,"Name")
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil {
		return 0
	}
	defer conn.Close()
	rep,err:=conn.Do("hlen","cart_"+strconv.Itoa(user.Id))
	cartcount,_:=redis.Int(rep,err)
	return cartcount
}

func (c *Cart)Showcart()  {
	username:=Getuser(&c.Controller)
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil {
		return
	}
	defer conn.Close()
	o:=orm.NewOrm()
	var user models.User
	user.Name=username
	o.Read(&user,"Name")
	goodsmap,_:=redis.IntMap(conn.Do("hgetall","cart_"+strconv.Itoa(user.Id)))
	goods:=make([]map[string]interface{},len(goodsmap))
	index:=0
	totalprice:=0
	totalcount:=0
	for i,v:=range goodsmap{
		skuid,_:=strconv.Atoi(i)
		var goodssku models.GoodsSKU
		goodssku.Id=skuid
		o.Read(&goodssku)
		temp:=make(map[string]interface{})
		temp["goodssku"]=goodssku
		temp["count"]=v
		totalprice+=goodssku.Price*v
		totalcount+=v
		temp["addprice"]=goodssku.Price*v
		goods[index]=temp
		index+=1
	}
	c.Data["goods"]=goods
	c.Data["totalprice"]=totalprice
	c.Data["totalcount"]=totalcount
	c.TplName="cart.html"
}

func (c *Cart)Handleupdatecart()  {
	skuid,err1:=c.GetInt("skuid")
	count,err2:=c.GetInt("count")
	resp:=make(map[string]interface{})
	defer c.ServeJSON()
	if err1!=nil||err2!=nil {
		resp["code"]=1
		resp["msg"]="请求错误"
		c.Data["json"]=resp
		return
	}
	username:=c.GetSession("username")
	if username==nil {
		resp["code"]=3
		resp["msg"]="未登陆"
		c.Data["json"]=resp
		return
	}
	o:=orm.NewOrm()
	var user models.User
	user.Name=username.(string)
	o.Read(&user,"Name")
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil {
		resp["code"]=2
		resp["msg"]="redis无法连接"
		c.Data["json"]=resp
		return
	}
	defer conn.Close()
	conn.Do("hset","cart_"+strconv.Itoa(user.Id),skuid,count)
	resp["code"]=5
	resp["msg"]="ok"
	c.Data["json"]=resp
}

func (c *Cart)Deletecart()  {
	skuid,err:=c.GetInt("skuid")
	resp:=make(map[string]interface{})
	defer c.ServeJSON()
	if err!=nil {
		resp["code"]=1
		resp["msg"]="数据错误"
		c.Data["json"]=resp
		return
	}
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	defer conn.Close()
	if err!=nil {
		resp["code"]=2
		resp["msg"]="redis连接失败"
		c.Data["json"]=resp
		return
	}
	o:=orm.NewOrm()
	var user models.User
	username:=c.GetSession("username")
	user.Name=username.(string)
	o.Read(&user,"Name")
	conn.Do("hdel","cart_"+strconv.Itoa(user.Id),skuid)
	resp["code"]=5
	resp["msg"]="ok"
	c.Data["json"]=resp
}