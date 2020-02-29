package service

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
)

// 客户子订单

func FindOneCustomerSubOrder(sub_order_sn string, com_id int64) (*models.CustomerSubOrder, error) {
	// TODO 订单实例和订单信息未分开
	collection := models.Client.Collection("customer_sub_order")
	filter := bson.M{}
	var customerSubOrder models.CustomerSubOrder
	filter["com_id"] = com_id
	filter["sub_order_sn"] = sub_order_sn
	err := collection.FindOne(context.TODO(), filter).Decode(&customerSubOrder)
	if err != nil {
		return nil, err
	}
	return &customerSubOrder, nil
}
