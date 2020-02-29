package service

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
)

// 销售订单相关方法 (客户订单)

// 查找客户订单
func FindSalesOrder(order_sn string,com_id int64) (*models.CustomerOrder, error) {
	collection := models.Client.Collection("customer_order")
	var salesOrder models.CustomerOrder
	filter := bson.M{}
	filter["order_sn"] = order_sn
	filter["com_id"] = com_id

	err := collection.FindOne(context.TODO(), filter).Decode(&salesOrder)
	if err != nil {
		return nil, err
	}
	return &salesOrder, nil
}
