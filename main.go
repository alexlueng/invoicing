package main

import (
	"jxc/conf"
	"jxc/router"
)

func main() {
	// 从配置文件读取配置
	conf.Init()


	//if err := models.Client.Ping(context.Background(), readpref.Primary()); err == nil {
	//	util.Log().Println("hello mongo")
	//}

	// 装载路由
	r := router.InitRouter()
	r.Run(":3000")

	// 支持热更新

}
