package service

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
)

// 供应商相关方法

// 获取供应商信息
func FindOneSupplier(supplier_id, com_id int64) (*models.Supplier, error) {
	collection := models.Client.Collection("supplier")
	filter := bson.M{}
	var supplier models.Supplier
	filter["com_id"] = com_id
	filter["supplier_id"] = supplier_id
	err := collection.FindOne(context.TODO(), filter).Decode(&supplier)
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}
