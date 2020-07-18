package models

import "go.mongodb.org/mongo-driver/mongo"

// 商品图片存储表
type Image struct {
	ImageID   int64  `json:"image_id" bson:"image_id"`
	ComID     int64  `json:"com_id" bson:"com_id"`
	ProductID int64  `json:"product_id" bson:"product_id"`
	LocalPath string `json:"local_path" bson:"local_path"` // 本地存放路径
	CloudPath string `json:"cloud_path" bson:"cloud_path"` // 云上存储路径
	IsDelete  bool   `json:"is_delete" bson:"is_delete"`
	Order     int64  `json:"order" bson:"order"`
}

func getImageCollection() *mongo.Collection {
	return Client.Collection("image")
}
