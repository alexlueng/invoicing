package models

type SupplierOrderReq struct {
	BaseReq
	//本页面定制的搜索字段
	OrderSN string `json:"order_sn" form:"order_sn"`
	SupplierName      string `json:"supplier_name" form:"supplier_name"` //模糊搜索
	Contacts string `json:"contacts" form:"contacts"` //模糊搜索
	Receiver string `json:"receiver" form:"receiver"` //模糊搜索
	Delivery string `json:"delivery" form:"delivery"`
	ExtraAmount float64 `json:"extra_amount" form:"extra_amount"`
	Status string `json:"status" form:"status"`
	StartOrderTime string `json:"start_order_time" form:"start_order_time"`
	EndOrderTime string `json:"end_order_time" form:"end_order_time"`
	StartPayTime string `json:"start_pay_time" form:"start_pay_time"`
	EndPayTime string `json:"end_pay_time" form:"end_pay_time"`
	StartShipTime string `json:"start_ship_time" form:"start_ship_time"`
	EndShipTime string `json:"end_ship_time" form:"end_ship_time"`
}

type ResponseSupplierOrdersData struct {
	SupplierOrders   interface{} `json:"customer_orders"`
	Total       int        `json:"total"`
	Pages       int        `json:"pages"`
	Size        int        `json:"size"`
	CurrentPage int        `json:"current_page"`
}
