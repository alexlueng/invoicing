package service

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
)

// 客户订单相关接口

// 获取客户订单
func FindOneCustomerOrder(order_sn string, com_id int64) (*models.CustomerOrder, error) {
	// TODO 订单实例和订单信息未分开
	collection := models.Client.Collection("customer_order")
	filter := bson.M{}
	var customerOrder models.CustomerOrder
	filter["com_id"] = com_id
	filter["order_sn"] = order_sn
	err := collection.FindOne(context.TODO(), filter).Decode(&customerOrder)
	if err != nil {
		return nil, err
	}
	return &customerOrder, nil
}
