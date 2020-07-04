package service

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
)

// 客户子订单

func FindOneCustomerSubOrder(sub_order_sn string, com_id int64) (*models.CustomerSubOrder, error) {
	collection := models.Client.Collection("customer_suborder")
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

// 修改子订单供应商、仓库发货数量
func UpdateSupplierAndWarehouseAmount(subOrderSn string, amountType int64, amount int64, comId int64) error {
	collection := models.Client.Collection("customer_suborder")
	filter := bson.M{}
	update := bson.M{}
	filter["com_id"] = comId
	filter["sub_order_sn"] = subOrderSn
	if amountType == 1 {
		update["warehouse_amount"] = amount
	} else {
		update["supplier_amount"] = amount
	}
	// update
	_, err := collection.UpdateOne(context.TODO(), filter, bson.M{"$inc": update})
	if err != nil {
		return err
	}
	return nil
}
