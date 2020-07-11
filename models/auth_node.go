package models

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

// 权限节点
// 本模块的所有公司通用这段数据，所以不添加comid
type AuthNote struct {
	AuthId  int64    `json:"auth_id" bson:"auth_id"`   // 节点id
	Note    string   `json:"note" bson:"note"`         // 节点名
	Group   string   `json:"group" bson:"group"`       // 组名
	GroupId int64    `json:"group_id" bson:"group_id"` // 权限节点组id，5为仓库权限
	Urls    []string `json:"urls" bson:"urls"`         // 这里记录了这个节点所有的路由
}

func getAuthNoteCollection() *mongo.Collection {
	return Client.Collection("auth_note")
}

func (a *AuthNote) InsertMany(authNotes []interface{}) {
	getAuthNoteCollection().InsertMany(context.TODO(), authNotes)
}
