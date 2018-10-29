package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"daydayshengxian/models"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"math"
)

type Goods struct {
	beego.Controller
}

func Getuser(c *beego.Controller)string  {
	username:=c.GetSession("username")
	if username==nil {
		c.Data["username"]=""
		return ""
	}else {
		c.Data["username"]=username.(string)
		return username.(string)
	}

}

func (g *Goods)Showindex()  {
	Getuser(&g.Controller)
	o:=orm.NewOrm()
	var goodstype []models.GoodsType
	o.QueryTable("GoodsType").All(&goodstype)
	g.Data["goodstype"]=goodstype
	var goodsbanner []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&goodsbanner)
	g.Data["goodsbanner"]=goodsbanner
	var promotiongoods []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&promotiongoods)
	g.Data["promotiongoods"]=promotiongoods
	goods:=make([]map[string]interface{},len(goodstype))
	for i,v:=range goodstype{
		temp:=make(map[string]interface{})
		temp["type"]=v
		goods[i]=temp
	}
	for _,v:= range goods {
		var textgoods []models.IndexTypeGoodsBanner
		var imagegoods []models.IndexTypeGoodsBanner
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSKU").OrderBy("Index").Filter("GoodsType",v["type"]).Filter("DisplayType",0).All(&textgoods)
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSKU").OrderBy("Index").Filter("GoodsType",v["type"]).Filter("DisplayType",1).All(&imagegoods)
		v["textgoods"]=textgoods
		v["imagegoods"]=imagegoods
	}
	g.Data["goods"]=goods
	g.TplName="index.html"
}

func Showlayout(c *beego.Controller)  {
	o:=orm.NewOrm()
	var types []models.GoodsType
	o.QueryTable("GoodsType").All(&types)
	c.Data["types"]=types
	Getuser(c)
	c.Layout="goodslayout.html"
}

func (g *Goods)Showgoodsdetail()  {
	id,err:=g.GetInt("id")
	if err!=nil {
		beego.Error("请求错误")
		g.Redirect("/",302)
		return
	}
	o:=orm.NewOrm()
	var goodssku models.GoodsSKU
	goodssku.Id=id
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType","Goods").Filter("Id",id).One(&goodssku)
	var newgoods []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType",goodssku.GoodsType).OrderBy("Time").Limit(2,0).All(&newgoods)
	g.Data["newgoods"]=newgoods
	g.Data["goodssku"]=goodssku
	username:=g.GetSession("username")
	if username!=nil {
		o:=orm.NewOrm()
		var user models.User
		user.Name=username.(string)
		o.Read(&user,"Name")
		conn,err:=redis.Dial("tcp","127.0.0.1:6379")
		defer conn.Close()
		if err!=nil {
			beego.Info("redis连接错误")
		}
		conn.Do("lrem","history_"+strconv.Itoa(user.Id),0,id)
		conn.Do("lpush","history_"+strconv.Itoa(user.Id),id)
	}
	Showlayout(&g.Controller)
	cartcount:=Getcartcount(&g.Controller)
	g.Data["cartcount"]=cartcount
	g.TplName="detail.html"
}

func Pagetool(pagecount int,pageindex int)[]int  {
	var pages []int
	if pagecount<5 {
		pages=make([]int,pagecount)
		for i,_:=range pages{
			pages[i]=i+1
		}
	}else if pageindex<3 {
		pages=[]int{1,2,3,4,5}
	}else if pageindex>=pagecount-3 {
		pages=[]int{pagecount-4,pagecount-3,pagecount-2,pagecount-1,pagecount}
	}else {
		pages=[]int{pageindex-2,pageindex-1,pageindex,pageindex+1,pageindex+2}
	}
	return pages
}

func (g *Goods)Showlist()  {
	id,err:=g.GetInt("typeid")
	beego.Info(id)
	if err!=nil {
		g.Redirect("/",302)
		return
	}
	Showlayout(&g.Controller)
	o:=orm.NewOrm()
	var goodsnew []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",id).OrderBy("Time").Limit(2,0).All(&goodsnew)
	g.Data["goodsnew"]=goodsnew
	var goods []models.GoodsSKU
	count,_:=o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",id).Count()
	pagesize:=1
	pagecount:=math.Ceil(float64(count)/float64(pagesize))
	pageindex,err:=g.GetInt("pageindex")
	if err!=nil {
		pageindex=1
	}
	pages:=Pagetool(int(pagecount),pageindex)
	g.Data["pages"]=pages
	g.Data["typeid"]=id
	g.Data["pageindex"]=pageindex
	start:=(pageindex-1)*pagesize
	perpage:=pageindex-1
	if perpage<=1 {
		perpage=1
	}
	g.Data["perpage"]=perpage
	nextpage:=pageindex+1
	if nextpage>int(pagecount) {
		nextpage=int(pagecount)
	}
	g.Data["nextpage"]=nextpage
	sort:=g.GetString("sort")
	if sort=="" {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",id).Limit(pagesize,start).All(&goods)
		g.Data["sort"]=""
		g.Data["goods"]=goods
	}else if sort=="price" {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",id).OrderBy("Price").Limit(pagesize,start).All(&goods)
		g.Data["sort"]="price"
		g.Data["goods"]=goods
	}else {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",id).OrderBy("Sales").Limit(pagesize,start).All(&goods)
		g.Data["sort"]="sale"
		g.Data["goods"]=goods
	}
	g.TplName="list.html"
}

func (g *Goods)Handlesearch()  {
	goodsname:=g.GetString("goodsname")
	o:=orm.NewOrm()
	var goods []models.GoodsSKU
	if goodsname=="" {
		o.QueryTable("GoodsSKU").All(&goods)
		g.Data["goods"]=goods
		Showlayout(&g.Controller)
		g.TplName="search.html"
		return
	}
	o.QueryTable("GoodsSKU").Filter("Name__icontains",goodsname).All(&goods)
	g.Data["goods"]=goods
	Showlayout(&g.Controller)
	g.TplName="search.html"
}