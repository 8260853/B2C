package controllers

import (
	"github.com/astaxie/beego"
	"github.com/KenmyZhang/aliyun-communicate"
)

type SMS struct {
	beego.Controller
}

func (s *SMS)Showsms()  {
	var (
		gatewayUrl      = "http://dysmsapi.aliyuncs.com/"
		accessKeyId     = "LTAIh83X7bYYTIXw"
		accessKeySecret = "fYSLqA3BI8jNviNhURKT9T9TmHeOuP"
		phoneNumbers    = "18379363028"
		signName        = "天天生鲜"
		templateCode    = "SMS_149101793"
		templateParam   = "{\"code\":\"34242\"}"
	)
	smsClient := aliyunsmsclient.New(gatewayUrl)
	result, err := smsClient.Execute(accessKeyId, accessKeySecret, phoneNumbers, signName, templateCode, templateParam)
	if err != nil {
		panic("Failed to send Message: " + err.Error())
	}

	if err != nil {
		panic(err)
	}
	if result.IsSuccessful() {
		s.Data["result"] = "发送成功"
	} else {
		s.Data["result"] = "发送失败"
	}
	s.TplName = "SMS.html"
}