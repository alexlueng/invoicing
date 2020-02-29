package service

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
)

// 采购订单相关方法（供应商订单）

// 查找采购订单
func FindPurchaseOrder(order_sn string, com_id int64) (*models.SupplierOrder, error) {
	collection := models.Client.Collection("supplier_order")
	var purchaseOrder models.SupplierOrder
	filter := bson.M{}
	filter["order_sn"] = order_sn
	filter["com_id"] = com_id

	err := collection.FindOne(context.TODO(), filter).Decode(&purchaseOrder)
	if err != nil {
		return nil, err
	}
	return &purchaseOrder, nil
}
