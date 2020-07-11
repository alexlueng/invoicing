package models

import "go.mongodb.org/mongo-driver/mongo"

// 仓库职员表
type WarehouseStuff struct {
	ComID         int64  `json:"com_id" bson:"com_id"`                 //公司id
	UserId        int64  `json:"user_id" bson:"user_id"`               // 用户id
	Username      string `json:"username" bson:"username"`             // 用户名
	WarehouseId   int64  `json:"warehouse_id" bson:"warehouse_id"`     // 仓库id
	WarehouseName string `json:"warehouse_name" bson:"warehouse_name"` // 仓库名
}

func getWarehouseStuffCollection() *mongo.Collection {
	return Client.Collection("warehouse_stuffs")
}
