package models


// Customer represent the customer
// 需要加上com_id, 每个公司都有自己的ID
type Customer struct {
	ID        int64  `json:"customer_id" bson:"customer_id"`
	ComID     int32 `json:"com_id" bson:"com_id"`
	Name      string `json:"customer_name" form:"customer_name"`
	Level     int64  `json:"level" form:"level"`
	Payment   string `json:"payment" form:"payment"`
	PayAmount string `json:"paid" form:"paid"`
	Receiver  string `json:"receiver" form:"receiver"`
	Address   string `json:"receiver_address" form:"address"`
	Phone     string `json:"receiver_phone" form:"phone"`
	//due string
}

//用户提交过来能数据
type CustReq struct {
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
	Name      string `json:"customer_name" form:"customer_name"`
	Level    string `json:"level" form:"level"`
	Payment  string `json:"payment" form:"payment"`
	Receiver string `json:"receiver" form:"receiver"` //模糊搜索
	Address  string `json:"address" form:"address"`   //模糊搜索
	Phone    string `json:"phone" form:"phone"`       //模糊搜索

}

type ResponseCustomerData struct {
	Customers   []Customer `json:"customers"`
	Total       int        `json:"total"`
	Pages       int        `json:"pages"`
	Size        int        `json:"size"`
	CurrentPage int        `json:"current_page"`
}