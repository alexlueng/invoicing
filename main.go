package main

import (
	"jxc/conf"
	"jxc/router"
)

func main() {
	// 从配置文件读取配置
	conf.Init()

	// 设置一个定时任务
	// 扫描全部的订单实例，找出需要提醒的订单
	// 每天执行一次
	// 这个做法无法更改提醒时间
	//go func() {
	//	s := gocron.NewScheduler()
	//	// 需要提醒时长应该是用户输入，保存到数据库中，现先从配置文件中读出来
	//	// TODO：每个独立的模块需要一个go协程来启动定时任务
	//	s.Every(1).Days().Do(api.OrderNotify, 1, 3, 2, "order_time", 0) // 发货提醒
	//	//s.Every(1).Days().Do(api.OrderNotify, 4, 3, 3, "confirm_time", 0) // 审核提醒
	//	//s.Every(1).Days().Do(api.OrderNotify, 3, 3, 4, "check_time", 1) // 客户结算提醒
	//	//s.Every(1).Days().Do(api.OrderNotify, 3, 3, 4, "check_time", 2) // 供应商结算提醒
	//	<- s.Start()
	//}()

	// 装载路由
	r := router.InitRouter()
	r.Run(":3000")

	// 支持热更新

}
