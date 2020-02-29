package models

import "fmt"

//用户提交过来的数据
type WarehouseReq struct {
	IdMin int `form:"idmin"` //okid界于[idmin 和 idmax] 之间的数据
	IdMax int `form:"idmax"` //ok
	//本页面的搜索字段 sf固定等于customer_name， key的值为用户提交过来的客户名关键字
	Key  string `form:"key"`              //用户提交过来的模糊搜索关键字
	Sf   string `form:"sf"`               //用户模糊搜索的字段  search field
	Page int64  `json:"page" form:"page"` //ok用户查询的是哪一页的数据
	Size int64  `json:"size" form:"size"` //ok用户希望每页展现多少条数据
	OrdF string `json:"ordf" form:"ordf"` //ok用户排序字段 order field
	Ord  string `json:"ord" form:"ord"`   //ok顺序还是倒序排列  ord=desc 倒序，ord = asc 升序
	TMin int    `form:"tmin"`             //时间最小值[tmin,tmax)
	TMax int    `form:"tmax"`             //时间最大值
	//本页面定制的搜索字段
	ID      int64  `json:"warehouse_id" form:"warehouse_id"`
	Name    string `json:"warehouse_name" form:"warehouse_name"`       //模糊搜索
	Address string `json:"warehouse_address" form:"warehouse_address"` //模糊搜索
	Manager int64  `json:"manager" form:"manager"`                     // 仓库管理员id
	Stuff   int64  `json:"stuff" form:"stuff"`                         //仓库职员id
}

type ResponseWarehouseData struct {
	Warehouses []Warehouse `json:"warehouses"`
	//Stuffs []Stuff `json:"wh_stuff"`
	Total       int `json:"total"`
	Pages       int `json:"pages"`
	Size        int `json:"size"`
	CurrentPage int `json:"current_page"`
}

func (w *Warehouse) UpdateStock() {
	fmt.Println("hello this is my warehouse: ", w.Name)
}
