package models

import "go.mongodb.org/mongo-driver/mongo"

// 库存数量不足时，通知的是仓库管理员
type MessageType struct {
	ID       int64  `json:"id" bson:"id"`             // 1 商品库存 2 仓库库存 3 订单 4 结算单（可能会拆分为客户和供应商两种）
	Name     string `json:"name" bson:"name"`         // 类型名字
	Template string `json:"template" bson:"template"` // 类型模板
}

func getMessageTypeCollection() *mongo.Collection {
	return Client.Collection("message_type")
}
