package service

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
)

// 子订单id查找采购子订单
func SubOrderIdFindOneSupplierSubOrder(SubOrderId int64, comId int64) (*models.SupplierSubOrder, error) {
	collection := models.Client.Collection("supplier_sub_order")
	var subOrder models.SupplierSubOrder
	filter := bson.M{}
	filter["com_id"] = comId
	filter["order_sub_id"] = SubOrderId
	err := collection.FindOne(context.TODO(), filter).Decode(&subOrder)
	if err != nil {
		return nil, err
	}
	return &subOrder, err
}

// 子订单号查找采购子订单
func SubOrderSnFindOneSupplierSubOrder(subOrderSn string, comId int64) (*models.SupplierSubOrder, error) {
	collection := models.Client.Collection("supplier_sub_order")
	var subOrder models.SupplierSubOrder
	filter := bson.M{}
	filter["com_id"] = comId
	filter["sub_order_sn"] = subOrderSn
	err := collection.FindOne(context.TODO(), filter).Decode(&subOrder)
	if err != nil {
		return nil, err
	}
	return &subOrder, err
}
