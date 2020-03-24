package models

// 客户结算单表结构
type CustomerSettlement struct {
	ID           int64  `json:"id"`
	SettlementSN string `json:"settlement_sn" bson:"settlement_sn"` //结算单号
	ComID        int64  `json:"com_id" bson:"com_id"`               // 公司id

	CustomerInstance []int64 `json:"cus_instance" bson:"cus_instance"` // 包含的子项 存放商品实例的id

	CustomerName string `json:"customer_name" bson:"customer_name"`
	CustomerID   int64  `json:"customer_id" bson:"customer_id"`
	//	ProductID     int64   `json:"product_id" bson:"product_id"`
	//	Product       string  `json:"product" bson:"product"`               // 商品名称
	Contacts string `json:"contacts" bson:"contacts"` //客户的联系人
	//Receiver      string  `json:"receiver" bson:"receiver"`             //本单的收货人
	//ReceiverPhone string  `json:"receiver_phone" bson:"receiver_phone"` //本单的收货人电话
	//	Price         float64 `json:"price" bson:"price"`                   //本项价格
	//	Amount        int64   `json:"amount" bson:"amount"`                 //本项购买总数量
	//	ExtraAmount   float64 `json:"extra_amount" bson:"extra_amount"`     //本单优惠或折扣金额
	//	Delivery      string  `json:"delivery" bson:"delivery"`             // 快递方式
	//	DeliveryCode  string  `json:"delivery_code" bson:"delivery_code"`   // 快递号
	CreateTime int64 `json:"create_time" bson:"create_time"` // 创建时间
	CreateBy   int64 `json:"create_by" bson:"create_by"`     // 创建这个结算单的人 可以是user_id
	//ShipTime      int64   `json:"ship_time" bson:"ship_time"`           // 发货时间
	//ConfirmTime   int64   `json:"confirm_time" bson:"confirm_time"`     // 确认订单时间
	//PayTime       int64   `json:"pay_time" bson:"pay_time"`             // 订单结算时间
	FinishTime int64 `json:"finish_time" bson:"finish_time"` // 结束时间，就是这个结算单付清的时间

	SettlementAmount float64 `json:"settlement_amount" bson:"settlement"` // 结算单总金额
	PaidAmount       float64 `json:"paid_amount" bson:"paid_amount"`      // 已付金额
	UnpaidAmount     float64 `json:"unpaid_amount" bson:"unpaid_amount"`  //未付金额

	Status int64 `json:"status" bson:"status"` // 结算单状态
	// 1 表示正在结算，2表示结算完成
	// operator_id操作人
	// 如何拆分子订单
}

// 客户付款后应该修改customer表中的已付款字段

type CustomerSettlementReq struct {
	BaseReq

	View string `json:"view"` // 查看结算单的视图，分为订单模式，客户模式，结算单模式
}

type ResponseCustomerSettlementData struct {
	Result      interface{} `json:"result"`
	Total       int         `json:"total"`
	Pages       int         `json:"pages"`
	Size        int         `json:"size"`
	CurrentPage int         `json:"current_page"`
}
