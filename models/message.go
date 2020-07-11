package models

import "go.mongodb.org/mongo-driver/mongo"

// 消息提醒功能
// 要设计好的数据结构，不要浪费每一位内存
// 库存提醒，时间提醒
// 使用websocket来通信？
type Message struct {
	ID        int64  `json:"id" bson:"id"`
	ComID     int64  `json:"com_id" bson:"com_id"`
	Title     string `json:"title" bson:"title"`     // 标题
	Message   string `json:"message" bson:"message"` // 内容
	Type      int64  `json:"type" bson:"type"`       // 通知类型
	IsRead    bool   `json:"is_read" bson:"is_read"` // 是否已读
	NotifyWay string `json:"notify_way" bson:"notify_way"`
	CreateAt  int64  `json:"create_at" bson:"create_at"`
	ReadAt    int64  `json:"read_at" bson:"read_at"` // 阅读时间
	//User       int64 `json:"user" bson:"user"` // 接收通知的人
	//SuperAdmin bool  `json:"super_admin" bson:"super_admin"` // 是否发给超级管理员
	//DepartmentID int64 `json:"department_id" bson:"department_id"` // 部门ID
}

func getMessageCollection() *mongo.Collection {
	return Client.Collection("message")
}

