package models


type OrderInstance2 struct {
	ComID    int64  `json:"com_id" bson:"com_id"`     // 公司id
	SupplierOrderSN  string `json:"order_sn" bson:"order_sn"` // 采购订单号
	CustomerOrderSN  string  `json:"sales_order_sn" bson:"sales_order_sn"` // 销售单号
	WarehouseID int64 `json:"warehouse_id" bson:"warehouse_id"` // 仓库id
	SupplierID    int64   `json:"supplier_id" bson:"supplier_id"`       // 供应商id
	Supplier string `json:"supplier_name" bson:"supplier_name"`
	CustomerID    int64   `json:"customer_id" bson:"customer_id"`
	Customer string `json:"customer_name" bson:"customer_name"`
	ProductID     int64   `json:"product_id" bson:"product_id"`
	Product       string  `json:"product" bson:"product"`               // 商品名称
	Contacts      string  `json:"contacts" bson:"contacts"`             //客户的联系人
	Receiver      string  `json:"receiver" bson:"receiver"`             //本单的收货人
	ReceiverPhone string  `json:"receiver_phone" bson:"receiver_phone"` //本单的收货人电话
	Price         float64 `json:"price" bson:"price"`                   //本项价格
	Amount        int64   `json:"amount" bson:"amount"`                 //本项购买数量
	ExtraAmount   float64 `json:"extra_amount" bson:"extra_amount"`     //本单优惠或折扣金额
	Delivery      string  `json:"delivery" bson:"delivery"`             // 快递方式
	DeliveryCode  string  `json:"delivery_code" bson:"delivery_code"`   // 快递号
	OrderTime     int64   `json:"order_time" bson:"order_time"`         // 下单时间
	ShipTime      int64   `json:"ship_time" bson:"ship_time"`           // 发货时间
	ConfirmTime   int64   `json:"confirm_time" bson:"confirm_time"`     // 确认订单时间
	PayTime       int64   `json:"pay_time" bson:"pay_time"`             // 订单结算时间
	FinishTime    int64   `json:"finish_time" bson:"finish_time"`       // 供应结束时间
	Status        int64   `json:"status" bson:"status"`                 // 订单状态
	// 订单状态：1、未发货 2、已发货 3、未确认 4、已确认

	SettlementOrderSN string `json:"settlement_order_sn" bson:"settlement_order_sn"` // 结算单号
	Settlement bool `json:"settlement"` // 是否结算
}
