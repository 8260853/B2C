package controllers

import (
	"github.com/astaxie/beego"
	"regexp"
	"github.com/astaxie/beego/orm"
	"daydayshengxian/models"
	"github.com/astaxie/beego/utils"
	"strconv"
	"encoding/base64"
	"github.com/gomodule/redigo/redis"
)

type Register struct {
	beego.Controller
}

func (r *Register)Showregister()  {
	r.TplName="register.html"
}

func (r *Register)Handleregister()  {
	username:=r.GetString("user_name")
	pwd:=r.GetString("pwd")
	cpwd:=r.GetString("cpwd")
	email:=r.GetString("email")
	if username==""||pwd==""||cpwd==""||email=="" {
		r.Data["errmsg"]="信息不完整"
		r.TplName="register.html"
		return
	}
	if pwd!=cpwd {
		r.Data["errmsg"]="密码不一致"
		r.TplName="register.html"
		return
	}
	reg,_:=regexp.Compile("^[A-Za-z0-9\u4e00-\u9fa5]+@[a-zA-Z0-9_-]+(\\.[a-zA-Z0-9_-]+)+$")
	res:=reg.FindString(email)
	if res=="" {
		r.Data["errmsg"]="邮箱格式不正确"
		r.TplName="register.html"
		return
	}
	o:=orm.NewOrm()
	var user models.User
	user.Name=username
	user.PassWord=pwd
	user.Email=email
	_,err:=o.Insert(&user)
	if err!=nil {
		r.Data["errmsg"]="用户重名"
		r.TplName="register.html"
		return
	}
	emailConfig:=`{"username":"563364657@qq.com","password":"cgapyzgkkczubdea","host":"smtp.qq.com","port":587}`
	emailConn:=utils.NewEMail(emailConfig)
	emailConn.From="563364657@qq.com"
	emailConn.To=[]string{email}
	emailConn.Subject="天天生鲜账户激活"
	emailConn.Text="192.168.104.142:8080/active?id="+strconv.Itoa(user.Id)
	emailConn.Send()
	r.Ctx.WriteString("请先激活账户")
}

func (r *Register)Active()  {
	id,err:=r.GetInt("id")
	if err!=nil {
		r.Data["errmsg"]="用户不存在"
		r.TplName="register.html"
		return
	}
	o:=orm.NewOrm()
	var user models.User
	user.Id=id
	err=o.Read(&user)
	if err!=nil {
		r.Data["errmsg"]="用户不存在"
		r.TplName="register.html"
		return
	}
	user.Active=true
	o.Update(&user)
	r.Redirect("/login",302)
}

func (r *Register)Showlogin()  {
	username:=r.Ctx.GetCookie("username")
	temp,_:=base64.StdEncoding.DecodeString(username)
	if string(temp)=="" {
		r.Data["username"]=""
		r.Data["checked"]=""
	}else {
		r.Data["username"]=string(temp)
		r.Data["checked"]="checked"
	}
	r.TplName="login.html"
}

func (r *Register)Handlelogin()  {
	username:=r.GetString("username")
	pwd:=r.GetString("pwd")
	if username==""||pwd=="" {
		r.Data["errmsg"]="用户名或密码错误"
		r.TplName="login.html"
		return
	}
	o:=orm.NewOrm()
	var user models.User
	user.Name=username
	err:=o.Read(&user,"Name")
	if err!=nil {
		r.Data["errmsg"]="用户不存在"
		r.TplName="login.html"
		return
	}
	if user.PassWord!=pwd {
		r.Data["errmsg"]="用户名或密码错误"
		r.TplName="login.html"
		return
	}
	if user.Active!=true {
		r.Data["errmsg"]="账号未激活"
		r.TplName="login.html"
		return
	}
	remenber:=r.GetString("remenber")
	if remenber=="on" {
		temp:=base64.StdEncoding.EncodeToString([]byte(username))
		r.Ctx.SetCookie("username",temp,24*3600*30)
	}else {
		r.Ctx.SetCookie("username",username,-1)
	}
	r.SetSession("username",username)
	r.Redirect("/",302)
}

func (r *Register)Logout()  {
	r.DelSession("username")
	r.Redirect("/login",302)
}

func (r *Register)Showcenterinfo()  {
	username:=Getuser(&r.Controller)
	r.Data["username"]=username
	o:=orm.NewOrm()
	var addr models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("user__name",username).Filter("Isdefault",true).One(&addr)
	if addr.Id==0 {
		r.Data["addr"]=""
	}else {
		r.Data["addr"]=addr
	}
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	defer conn.Close()
	if err!=nil {
		beego.Info("连接失败")
	}
	var user models.User
	user.Name=username
	o.Read(&user,"Name")
	rep,err:=conn.Do("lrange","history_"+strconv.Itoa(user.Id),0,4)
	goodsids,_:=redis.Ints(rep,err)
	var goodsskus []models.GoodsSKU
	for _,v:=range goodsids{
		var goodssku models.GoodsSKU
		goodssku.Id=v
		o.Read(&goodssku)
		goodsskus=append(goodsskus,goodssku)
	}
	r.Data["goodsskus"]=goodsskus
	r.Layout="usercenterlayout.html"
	r.TplName="user_center_info.html"
}

func (r *Register)Showcenterorder()  {
	username:=Getuser(&r.Controller)
	o:=orm.NewOrm()
	var user models.User
	user.Name=username
	o.Read(&user,"Name")
	var orderinfos []models.OrderInfo
	o.QueryTable("OrderInfo").RelatedSel("User").Filter("User__Id",user.Id).All(&orderinfos)
	goodsbuffer:=make([]map[string]interface{},len(orderinfos))
	for i,orderinfo:=range orderinfos{
		var ordergoods []models.OrderGoods
		o.QueryTable("OrderGoods").RelatedSel("OrderInfo","GoodsSKU").Filter("OrderInfo__Id",orderinfo.Id).All(&ordergoods)
		temp:=make(map[string]interface{})
		temp["orderinfo"]=orderinfo
		temp["ordergoods"]=ordergoods
		goodsbuffer[i]=temp
	}
	r.Data["goodsbuffer"]=goodsbuffer
	r.Layout="usercenterlayout.html"
	r.TplName="user_center_order.html"
}

func (r *Register)Showcentersite()  {
	username:=Getuser(&r.Controller)
	r.Data["username"]=username
	o:=orm.NewOrm()
	var addr models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name",username).Filter("Isdefault",true).One(&addr)
	r.Data["addr"]=addr
	r.Layout="usercenterlayout.html"
	r.TplName="user_center_site.html"
}

func (r *Register)Handlecentersite()  {
	receiver:=r.GetString("receiver")
	addr:=r.GetString("addr")
	zipcode:=r.GetString("zipcode")
	phone:=r.GetString("phone")
	if receiver=="" || addr=="" || zipcode=="" || phone==""{
		r.Redirect("/user/centersite",302)
		return
	}
	o:=orm.NewOrm()
	var add models.Address
	add.Isdefault=true
	err:=o.Read(&add,"Isdefault")
	if err==nil {
		add.Isdefault=false
		o.Update(&add)
	}
	username:=r.GetSession("username")
	var user models.User
	user.Name=username.(string)
	o.Read(&user,"Name")
	var addnew models.Address
	addnew.Receiver=receiver
	addnew.Addr=addr
	addnew.Zipcode=zipcode
	addnew.Phone=phone
	addnew.Isdefault=true
	addnew.User=&user
	o.Insert(&addnew)
	r.Redirect("/user/centersite",302)
}