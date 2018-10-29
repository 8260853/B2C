package routers

import (
	"daydayshengxian/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {
	beego.InsertFilter("/user/*",beego.BeforeExec,filtfunc)
    beego.Router("/register",&controllers.Register{},"get:Showregister;post:Handleregister")
    beego.Router("/active",&controllers.Register{},"get:Active")
    beego.Router("/login",&controllers.Register{},"get:Showlogin;post:Handlelogin")
	beego.Router("/",&controllers.Goods{},"get:Showindex")
	beego.Router("/user/logout",&controllers.Register{},"get:Logout")
	beego.Router("/user/centerinfo",&controllers.Register{},"get:Showcenterinfo")
	beego.Router("/user/centerorder",&controllers.Register{},"get:Showcenterorder")
	beego.Router("/user/centersite",&controllers.Register{},"get:Showcentersite;post:Handlecentersite")
	beego.Router("/goodsdetail",&controllers.Goods{},"get:Showgoodsdetail")
	beego.Router("/goodslist",&controllers.Goods{},"get:Showlist")
	beego.Router("/goodssearch",&controllers.Goods{},"post:Handlesearch")
	beego.Router("/user/addcart",&controllers.Cart{},"post:Handleaddcart")
	beego.Router("/user/cart",&controllers.Cart{},"get:Showcart")
	beego.Router("/user/updatecart",&controllers.Cart{},"post:Handleupdatecart")
	beego.Router("/user/deletecart",&controllers.Cart{},"post:Deletecart")
	beego.Router("/user/showorder",&controllers.Order{},"post:Showorder")
	beego.Router("/user/addorder",&controllers.Order{},"post:Addorder")
	beego.Router("/user/pay",&controllers.Order{},"get:Handlepay")
	beego.Router("/user/payok",&controllers.Order{},"get:Payok")
	beego.Router("/sendsms",&controllers.SMS{},"get:Showsms")
}

var filtfunc = func(ctx *context.Context) {
	username:=ctx.Input.Session("username")
	if username==nil {
		ctx.Redirect(302,"/login")
		return
	}
}