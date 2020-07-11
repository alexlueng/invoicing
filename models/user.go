package models

import "go.mongodb.org/mongo-driver/mongo"

// 人员表数据格式
type User struct {
	ComId     int64       `bson:"com_id" json:"com_id"`        //
	UserID    int64       `bson:"user_id" json:"user_id"`      //
	Password  string      `bson:"password" json:"password"`    //
	Username  string      `bson:"username" json:"username"`    //
	Phone     string      `bson:"phone" json:"phone"`          //
	Authority interface{} `bson:"authority" json:"authority"`  // 普通权限
	Warehouse interface{} `bson:"warehouse" json:"warehouse"`  // 仓库权限
	Position  string      `bson:"position" json:"position"`    //职务
	CreateAt  int64       `json:"create_at"  bson:"create_at"` // 创建时间戳
	CreateBy  int64       `json:"create_by" bson:"create_by"`  // 创建者id
	ModifyAt  int64       `json:"modify_at" bson:"modify_at"`  // 最后修改时间戳
	ModifyBy  int64       `json:"modify_by" bson:"modify_by"`  // 最后修改者id

	Urls []string `json:"urls"` // 权限路由，不在数据库中存储
}

func getUserCollection() *mongo.Collection {
	return Client.Collection("users")
}

func (u User) CheckPassword(password string) bool {
	if u.Password != password {
		return false
	}
	return true
}

// 登录日志数据结构
type LoginLogData struct {
	LogId   string `json:"log_id"`
	Ip      string `json:"ip"`
	UserId  string `json:"user_id"`
	Message string `json:"message"`
}

