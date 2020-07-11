package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// 商品详情
type ProductDetail struct {
	ComID       int64  `json:"com_id" bson:"com_id"`
	ProductID   int64  `json:"product_id" bson:"product_id"`
	ProductName string `json:"product_name" bson:"product_name"`
	Attr        string `json:"attr" bson:"attr"`         // 商品属性 以json的方式保存
	Discribe    string `json:"discribe" bson:"discribe"` // 商品描述
	/*	DetailID int64 `json:"detail_id" bson:"detail_id"`
		ProductName string `json:"product_name" bson:"product_name"`
		Sales int64 `json:"sales" bson:"sales"` // 销量
		Origin string `json:"origin" bson:"origin"` // 产地
		Size string `json:"size" bson:"size"` // 规格
		Unit int64 `json:"unit" bson:"unit"` // 商品单位
		Weight float64 `json:"weight" bson:"weight"` // 重量
		Packaging string `json:"packaging" bson:"packaging"` // 包装
		Quality string `json:"quality" bson:"quality"` // 保质期
		Storage string `json:"storage" bson:"storage"` // 贮存方式*/
}

func getProductDetailCollection() *mongo.Collection {
	return Client.Collection("product_detail")
}

func GetProductDetailByID(com_id, product_id int64) (detail *ProductDetail, err error) {

	filter := bson.M{}
	filter["com_id"] = com_id
	filter["product_id"] = product_id

	err = getProductDetailCollection().FindOne(context.TODO(), filter).Decode(&detail)
	if err != nil {
		return nil, err
	}
	return
}
