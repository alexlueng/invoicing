package models

import "go.mongodb.org/mongo-driver/mongo"

// 采购订单实例（供应商订单实例）
// 采购订单子订单
type SupplierSubOrder struct {
	SubOrderId       int64   `json:"order_sub_id" bson:"order_sub_id"`             // 子订单id
	SubOrderSn       string  `json:"order_sub_sn" bson:"order_sub_sn"`             // 子订单号
	ComID            int64   `json:"com_id" bson:"com_id"`                         // 公司id
	OrderId          int64   `json:"order_id" bson:"order_id"`                     // 采购订单id
	OrderSn          string  `json:"order_sn" bson:"order_sn"`                     // 订单号
	ProductID        int64   `json:"product_id" bson:"product_id"`                 // 商品id
	ProductName      string  `json:"product_name" bson:"product_name"`             // 商品名 这是冗余字段
	ProductNum       int64   `json:"product_num" bson:"product_num"`               // 商品数量
	ProductUnitPrice float64 `json:"product_unit_price" bson:"product_unit_price"` // 商品单价
	Units            string  `json:"units" bson:"units"`                           // 商品量词，这是冗余字段
	CreateAt         int64   `json:"create_at"  bson:"create_at"`                  // 创建时间戳
	CreateBy         int64   `json:"create_by" bson:"create_by"`                   // 创建者id
	ShipTime         int64   `json:"ship_time" bson:"ship_time"`                   // 发货时间戳
	Shipper          int64   `json:"shipper" bson:"shipper"`                       // 发货者id
	ConfirmAt        int64   `json:"confirm_at" bson:"confirm_at"`                 // 确认收货时间戳
	ConfirmBy        int64   `json:"confirm_by" bson:"confirm_by"`                 // 确认收货者id
	CheckAt          int64   `json:"check_at" bson:"check_at"`                     // 盘点时间
	CheckBy          int64   `json:"check_by" bson:"check_by"`                     // 盘点操作者id
	State            int64   `json:"state" bson:"state"`                           // 订单状态
	ActualAmount     int64   `json:"actual_amount" bson:"actual_amount"`           // 实发数量
	FailReason       string  `json:"fail_reason" bson:"fail_reason"`               // 审核不通过的理由
}

func getSupplierSubOrderCollection() *mongo.Collection {
	return Client.Collection("supplier_sub_order")
}
