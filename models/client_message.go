package models

import "go.mongodb.org/mongo-driver/mongo"

// TODO：消息模板上加上一个链接来让用户点击就可以去到相应的未处理的订单
// 供应商，客户通知消息
type MessageForClient struct {
	ID        int64  `json:"id" bson:"id"`
	ComID     int64  `json:"com_id" bson:"com_id"`
	Client    int64  `json:"client" bson:"client"`       // 1 客户 2 供应商
	ClientID  int64  `json:"client_id" bson:"client_id"` //
	Telephone string `json:"telephone" bson:"telephone"`
	Title     string `json:"title" bson:"title"`     // 消息标题
	Content   string `json:"content" bson:"content"` // 消息内容
	CreateAt  int64  `json:"create_at" bson:"create_at"`
	IsRead    bool   `json:"is_read" bson:"is_read"`
}

func getMessageForClientTypeCollection() *mongo.Collection {
	return Client.Collection("client_message")
}
