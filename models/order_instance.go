package models

// 商品实例表
// 客户 1
// 供应商 2
// 仓库 3

// 从仓库发往客户 销售订单 src_type = 3  dest_type = 1 库存 -
// 从供应商发往客户 销售订单 src_type = 2 dest_type = 1
// 从供应商到仓库 采购订单 src_type = 2 dest_type = 3 库存 +
// 客户退货，从客户到仓库 销售订单  src_type = 1 dest_type = 3 库存 +
// 客户退货，从客户到供应商 销售订单 src_type = 1 dest_type = 2
// 凭空增加，没有来源到仓库  src_type = 0 dest_type = 3 库存 +
// 损耗，有来源没有去向 src_type = 3 dest_type = 0 库存 -
type GoodsInstance struct {
	InstanceId int64 `json:"instance_id" bson:"instance_id"` // 实例id
	ComID      int64 `json:"com_id" bson:"com_id"`           // 公司id
	Type       int64 `json:"type" bson:"type"`               // 订单类型：1、客户订单 2、采购订单

	SrcType       int64  `json:"src_type" bson:"src_type"`                 // 来源类型
	SrcId         int64  `json:"src_id" bson:"src_id"`                     // 来源id
	SrcTitle      string `json:"src_title"  bson:"src_title"`              // 来源名称 相当于原 Supplier
	SrcOrderId    int64  `json:"src_order_id" bson:"src_order_id"`         // 来源订单id
	SrcOrderSn    string `json:"src_order_sn" bson:"src_order_sn"`         // 来源订单号
	SrcSubOrderId int64  `json:"src_sub_order_id" bson:"src_sub_order_id"` // 来源订单子id
	SrcSubOrderSn string `json:"src_sub_order_sn" bson:"src_sub_order_sn"` // 来源订单子号

	DestType       int64  `json:"dest_type"  bson:"dest_type"`  // 去向类型
	DestId         int64  `json:"dest_id"  bson:"dest_id"`      // 去向id
	DestTitle      string `json:"dest_title" bson:"dest_title"` // 去向名称 相当于原 Customer
	DestOrderId    int64  `json:"dest_order_id" bson:"dest_order_id"`        // 去向订单id
	DestOrderSn    string `json:"dest_order_sn" bson:"dest_order_sn"`        // 去向订单号
	DestSubOrderId int64  `json:"dest_sub_order_id" bson:"dest_sub_order_id"`    // 去向订单子id
	DestSubOrderSn string  `json:"dest_sub_order_sn" bson:"dest_sub_order_sn"`   // 去向订单子号

	// 销售-待发货 销售-已发货 销售-确认收货 采购-待收货 采购-已收货 无端损耗 凭空增加
	// placeType 1 销售-待发货
	// placeType 2 销售-已发货
	// placeType 3 销售-确认收货
	// placeType 4 采购-待收货
	// placeType 5 采购-已收货
	// placeType 6 无端损耗
	// placeType 7 凭空增加

	PlaceType  int64 `json:"place_type" bson:"place_type"`     // 当前此商品的位置类型
	PlaceId    int64 `json:"place_id" bson:"place_id"`         // 当前此商品的位置
	SubPlaceId int64 `json:"sub_place_id" bson:"sub_place_id"` // 当前此商品的货架ID

	ProductID     int64   `json:"product_id" bson:"product_id"`         //商品id
	Product       string  `json:"product" bson:"product"`               // 商品名称
	Contacts      string  `json:"contacts" bson:"contacts"`             //客户的联系人
	Receiver      string  `json:"receiver" bson:"receiver"`             //本单的收货人
	ReceiverPhone string  `json:"receiver_phone" bson:"receiver_phone"` //本单的收货人电话
	CustomerPrice float64 `json:"customer_price" bson:"customer_price"` // 客户价格，对应售价
	SupplierPrice float64 `json:"supplier_price" bson:"supplier_price"` // 供应商价格，对应采购价
	Amount        int64   `json:"amount" bson:"amount"`                 //本项购买数量
	ExtraAmount   float64 `json:"extra_amount" bson:"extra_amount"`     //本单优惠或折扣金额
	DeliveryCom   int64   `json:"delivery_com" bson:"delivery_com"`     //配送方式id
	Delivery      string  `json:"delivery" bson:"delivery"`             // 快递方式
	DeliveryCode  string  `json:"delivery_code" bson:"delivery_code"`   // 快递号
	OrderTime     int64   `json:"order_time" bson:"order_time"`         // 下单时间
	CreateBy      int64   `json:"create_by" bson:"create_by"`           // 创建者id
	ShipTime      int64   `json:"ship_time" bson:"ship_time"`           // 发货时间
	ConfirmTime   int64   `json:"confirm_time" bson:"confirm_time"`     // 确认订单时间
	CheckTime     int64   `json:"check_time" bson:"check_time"`         // 审核时间
	PayTime       int64   `json:"pay_time" bson:"pay_time"`             // 订单结算时间
	FinishTime    int64   `json:"finish_time" bson:"finish_time"`       // 供应结束时间
	Status        int64   `json:"status" bson:"status"`                 // 订单状态
	// 订单状态：1、未发货 2、确认收货 3、审核 4、已确认 5、已发货

	Units string `json:"units" bson:"units"` // 计量单位

	SettlementOrderSN string `json:"settlement_order_sn" bson:"settlement_order_sn"` // 结算单号
	Settlement        int64  `json:"settlement"`                                     // 是否结算
}
