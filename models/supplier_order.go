package models

type SupplierOrder struct {
	ComID int64 `json:"com_id" bson:"com_id"`
	OrderSN string `json:"order_sn" bson:"order_sn"`
	WarehouseID int64 `json:"warehouse_id" bson:"warehouse_id"`
	//SupplierOrderID int64 `json:"supplier_order_id" bson:"supplier_order_id"`
	SupplierID int64 `json:"supplier_id" bson:"supplier_id"`
	SupplierName string `json:"supplier_name" bson:"supplier_name"`
	Contacts string `json:"contacts" bson:"contacts"`
	Receiver string  `json:"receiver" bson:"receiver"`
	Phone string `json:"receiver_phone" bson:"receiver_phone"`
	Price float64 `json:"price" bson:"price"`
	Amount int64 `json:"amount" bson:"amount"`
	ExtraAmount float64 `json:"extra_amount" bson:"extra_amount"`
	Delivery string `json:"delivery" bson:"delivery"`
	DeliveryCode string `json:"delivery_code" bson:"delivery_code"`
	OrderTime int64 `json:"order_time" bson:"order_time"` // 所有的时间都是以int64的类型插入到mongodb中
	ShipTime int64 `json:"ship_time" bson:"ship_time"`
	ConfirmTime int64 `json:"confirm_time" bson:"confirm_time"`
	PayTime int64 `json:"pay_time" bson:"pay_time"`
	FinishTime int64 `json:"finish_time" bson:"finish_time"`
	Status string `json:"status" bson:"status"`
}

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
	SupplierOrders   []SupplierOrder `json:"customer_orders"`
	Total       int        `json:"total"`
	Pages       int        `json:"pages"`
	Size        int        `json:"size"`
	CurrentPage int        `json:"current_page"`
}