package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
)

// 仓库库存管理表
// 入库管理 出库管理
// 这个表没有删除 只有修改

/*

	操作类型：
		采购订单：
			用户下采购订单时，如果该仓库没有这个商品，则新建一条记录，增加未入库数量
			采购订单确认收货时，减少未入库。增加当前库存
			暂时不考虑无故增加或无故减少的情况
		销售订单
			用户下销售订单,并从仓库发货的时候，减少当前库存，增加待发货数量
			销售订单确认发货，减少待发货，增加已发货

*/


type WarehouseProduct struct {
	ComID        int64  `json:"com_id" bson:"com_id"`
	WarehouseID  int64  `json:"warehouse_id" bson:"warehouse_id"`
	InstanceID   int64  `json:"instance_id" bson:"instance_id"` // 实例ID
	ProductID    int64  `json:"product_id" bson:"product_id"`
	ProductName  string `json:"product_name" bson:"product_name"`
	ProductUnit  string `json:"product_unit" bson:"product_unit"`   // 商品单位
	CurrentStock int64  `json:"current_stock" bson:"current_stock"` // 当前库存
	UnStock      int64  `json:"un_stock" bson:"un_stock"`           // 未入库/待入库
	Shipped      int64  `json:"sent" bson:"sent"`                   // 已发货
	ToBeShipped  int64  `json:"to_be_sent" bson:"to_be_sent"`       // 待发货
	Operation    string `json:"operation" bson:"operation"`         // 操作
	IsDelete     bool   `json:"is_delete" bson:"is_delete"`
	CreateAt     int64  `json:"create_at" bson:"create_at"`
}


func WarehouseProductDetail(com_id, warehouse_id int64) (result []WarehouseProduct, err error) {
	collection := Client.Collection("warehouse_product")
	filter := bson.M{}
	filter["com_id"] = com_id
	filter["warehouse_id"] = warehouse_id
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return
	}
	for cur.Next(context.TODO()) {
		var res WarehouseProduct
		if err := cur.Decode(&res); err != nil {
			return nil, err
		}
		result = append(result, res)
	}
	return
}

func ProductInfoOfWarehouse(product_id int64, com_id int64) (result []WarehouseProduct, err error) {

	collection := Client.Collection("warehouse_product")
	filter := bson.M{}
	filter["com_id"] = com_id
	filter["product_id"] = product_id

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return
	}

	for cur.Next(context.TODO()) {
		var res WarehouseProduct
		if err := cur.Decode(&res); err != nil {
			return nil, err
		}
		result = append(result, res)
	}
	return
}