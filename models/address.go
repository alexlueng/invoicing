package models

// 用户地址信息管理
type Address struct {
	ComID          int64  `json:"com_id" bson:"com_id"`
	AddressID      int64  `json:"address_id" bson:"address_id"`
	CustomerID     int64  `json:"customer_id" bson:"customer_id"`
	Reciever       string `json:"reciever" bson:"reciever"`               // 收货人
	Gender         int64  `json:"gender" bson:"gender"`                   // 性别
	Telephone      string `json:"telephone"`                              // 手机
	Location       string `json:"location" bson:"location"`               // 收货地址,小区，写字楼
	LocationDetail string `json:"location_detail" bson:"location_detail"` // 楼号，门牌号
	Type           string `json:"tpye" bson:"type"`                       // 地址类型
	IsDefault      bool   `json:"is_default" bson:"is_default"`           // 是否默认地址
}
