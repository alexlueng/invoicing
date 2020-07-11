package models

import "go.mongodb.org/mongo-driver/mongo"

// 库存实例表
// 类型 = 0 凭空多出的商品，采购订单号、销售订单号均为空 +
// 类型 = 1 退货 销售订单退货造成库存增多，订单号记录到销售订单号 +
// 类型 = 2 销售 销售后发货造成仓库库存减少，订单号记录到销售订单号 —
// 类型 = 3 损耗 盘点仓库，商品损耗造成库存减少 -
// 类型 = 4 采购 采购商品造成库存增多，订单号记录到采购订单号 +
type WosInstance struct {
	ComID            int64   `json:"com_id" bson:"com_id"`                         // 公司id
	Type             int64   `json:"type" bson:"type"`                             // 类型
	WarehouseID      int64   `json:"warehouse_id" bson:"warehouse_id"`             // 仓库id
	WarehouseName    string  `json:"warehouse_name" bson:"warehouse_name"`         // 仓库名，这是冗余字段
	PurchaseOrderSn  string  `json:"purchase_order_sn" bson:"purchase_order_sn"`   // 采购订单号
	SalesOrderSn     string  `json:"sales_order_sn" bson:"sales_order_sn"`         // 销售订单号
	ProductID        int64   `json:"product_id" bson:"product_id"`                 // 商品id
	ProductName      string  `json:"product_name" bson:"product_name"`             // 商品名
	Units            string  `json:"units" bson:"units"`                           // 商品量词，这是冗余字段
	ProductNum       int64   `json:"product_num" bson:"product_num"`               // 商品数量
	ProductUnitPrice float64 `json:"product_unit_price" bson:"product_unit_price"` // s
	CreateAt         int64   `json:"create_at"  bson:"create_at"`                  // 创建时间戳
	CreateBy         int64   `json:"create_by" bson:"create_by"`                   // 创建者id
	ShipTime         int64   `json:"ship_time" bson:"ship_time"`                   // 发货时间戳
	Shipper          int64   `json:"shipper" bson:"shipper"`                       // 发货者id
	ConfirmAt        int64   `json:"confirm_at" bson:"confirm_at"`                 // 确认收货时间戳
	ConfirmBy        int64   `json:"confirm_by" bson:"confirm_by"`                 // 确认收货者id
	CheckAt          int64   `json:"check_at" bson:"check_at"`                     // 盘点时间
	CheckBy          int64   `json:"check_by" bson:"check_by"`                     // 盘点操作者id
}

func getWosInstanceCollection() *mongo.Collection {
	return Client.Collection("wos_instance")
}
