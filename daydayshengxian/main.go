package main

import (
	_ "daydayshengxian/routers"
	"github.com/astaxie/beego"
	_"github.com/go-sql-driver/mysql"
	_"daydayshengxian/models"
)

func main() {
	beego.Run()
}

