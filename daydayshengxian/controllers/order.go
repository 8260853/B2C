package controllers

import (
	"github.com/astaxie/beego"
	"strconv"
	"github.com/astaxie/beego/orm"
	"daydayshengxian/models"
	"github.com/gomodule/redigo/redis"
	"time"
	"strings"
	"github.com/smartwalle/alipay"
	"fmt"
)

type Order struct {
	beego.Controller
}

func (o *Order)Showorder()  {
	skuids:=o.GetStrings("skuid")
	beego.Info(skuids)
	if len(skuids)==0 {
		o.Redirect("/user/cart",302)
		return
	}
	orm:=orm.NewOrm()
	conn,_:=redis.Dial("tcp","127.0.0.1:6379")
	defer conn.Close()
	var user models.User
	username:=o.GetSession("username")
	user.Name=username.(string)
	orm.Read(&user,"Name")
	goodsbuffer:=make([]map[string]interface{},len(skuids))
	totalprice:=0
	totalcount:=0
	for i,skuid:=range skuids{
		temp:=make(map[string]interface{})
		id,_:=strconv.Atoi(skuid)
		var goodssku models.GoodsSKU
		goodssku.Id=id
		orm.Read(&goodssku)
		temp["goods"]=goodssku
		count,_:=redis.Int(conn.Do("hget","cart_"+strconv.Itoa(user.Id),id))
		temp["count"]=count
		amount:=goodssku.Price*count
		temp["amount"]=amount
		totalcount+=count
		totalprice+=amount
		goodsbuffer[i]=temp
	}
	o.Data["goodsbuffer"]=goodsbuffer
	var add []models.Address
	orm.QueryTable("Address").RelatedSel("User").Filter("User__Id",user.Id).All(&add)
	o.Data["add"]=add
	o.Data["totalprice"]=totalprice
	o.Data["totalcount"]=totalcount
	transferprice:=10
	o.Data["transferprice"]=transferprice
	o.Data["realyprice"]=totalprice+transferprice
	o.Data["skuids"]=skuids
	o.TplName="place_order.html"
}

func (o *Order)Addorder()  {
	addrid,_:=o.GetInt("addrid")
	payid,_:=o.GetInt("payid")
	skuid:=o.GetString("skuids")
	ids:=skuid[1:len(skuid)-1]
	skuids:=strings.Split(ids," ")
	totalCount,_:=o.GetInt("totalCount")
	transferPrice,_:=o.GetInt("transferPrice")
	realyPrice,_:=o.GetInt("realyPrice")
	resp:=make(map[string]interface{})
	defer o.ServeJSON()
	if len(skuids)==0 {
		resp["code"]=1
		resp["msg"]="数据库连接失败"
		o.Data["json"]=resp
		return
	}
	oo:=orm.NewOrm()
	oo.Begin()
	username:=o.GetSession("username")
	var user models.User
	user.Name=username.(string)
	oo.Read(&user,"Name")
	var order models.OrderInfo
	order.OrderId=time.Now().Format("2006010215030405")+strconv.Itoa(user.Id)
	order.User=&user
	order.Orderstatus=1
	order.PayMethod=payid
	order.TotalCount=totalCount
	order.TotalPrice=realyPrice
	order.TransitPrice=transferPrice
	var addr models.Address
	addr.Id=addrid
	oo.Read(&addr)
	order.Address=&addr
	oo.Insert(&order)
	conn,_:=redis.Dial("tcp","127.0.0.1:6379")
	defer conn.Close()
	for _,v:=range skuids{
		id,_:=strconv.Atoi(v)
		var goods models.GoodsSKU
		goods.Id=id
		i:=3
		for i>0 {
			oo.Read(&goods)
			var ordergoods models.OrderGoods
			ordergoods.GoodsSKU=&goods
			ordergoods.OrderInfo=&order
			count,_:=redis.Int(conn.Do("hget","cart_"+strconv.Itoa(user.Id),id))
			if count>goods.Stock {
				resp["code"]=2
				resp["msg"]="库存不足"
				o.Data["json"]=resp
				oo.Rollback()
				return
			}
			precount:=goods.Stock
			ordergoods.Count=count
			ordergoods.Price=count*goods.Price
			oo.Insert(&ordergoods)
			goods.Stock-=count
			goods.Sales+=count
			updatecount,_:=oo.QueryTable("GoodsSKU").Filter("Id",goods.Id).Filter("Stock",precount).Update(orm.Params{"Stock":goods.Stock,"Sales":goods.Sales})
			if updatecount==0 {
				if i>0 {
					i-=1
					continue
				}
				resp["code"]=3
				resp["msg"]="库存改变，提交失败"
				o.Data["json"]=resp
				oo.Rollback()
				return
			}else {
				conn.Do("hdel","cart_"+strconv.Itoa(user.Id),goods.Id)
				break
			}
		}
	}
	oo.Commit()
	resp["code"]=5
	resp["msg"]="ok"
	o.Data["json"]=resp
}

func (o *Order)Handlepay()  {
	var aliPublicKey = "" // 可选，支付宝提供给我们用于签名验证的公钥，通过支付宝管理后台获取
	var privateKey = "xxx" // 必须，上一步中使用 RSA签名验签工具 生成的私钥
	var appId="2016092200569649"
	var client = alipay.New(appId, aliPublicKey, privateKey, false)
	orderid:=o.GetString("orderid")
	totalprice:=o.GetString("totalprice")
	var p = alipay.AliPayTradePagePay{}
	p.NotifyURL = "http://xxx"
	p.ReturnURL = "http://192.168.104.112:8082/user/payok"
	p.Subject = "天天生鲜"
	p.OutTradeNo =orderid
	p.TotalAmount = totalprice
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	var url, err = client.TradePagePay(p)
	if err != nil {
		fmt.Println(err)
	}

	var payURL = url.String()
	o.Redirect(payURL,302)
}

func (o *Order)Payok()  {
	orderid:=o.GetString("out_trade_no")
	if orderid=="" {
		o.Redirect("/user/centerorder",302)
		return
	}
	oo:=orm.NewOrm()
	count,_:=oo.QueryTable("OrderInfo").Filter("OrderId",orderid).Update(orm.Params{"Orderstatus":2})
	if count==0 {
		o.Redirect("/user/centerorder",302)
		return
	}
	o.Redirect("/user/centerorder",302)
}