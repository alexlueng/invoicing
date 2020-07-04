package models

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

// 库存数量不足时，通知的是仓库管理员

type MessageType struct {
	ID       int64  `json:"id" bson:"id"`             // 1 商品库存 2 仓库库存 3 订单 4 结算单（可能会拆分为客户和供应商两种）
	Name     string `json:"name" bson:"name"`         // 类型名字
	Template string `json:"template" bson:"template"` // 类型模板
}

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
