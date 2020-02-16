package models

// 所有的数据表格式都会存放在这里

// 人员表数据格式
type User struct {
	ComId     int64  `bson:"com_id" json:"com_id"`       //
	UserID    int64 `bson:"user_id" json:"user_id"`     //
	Password  string `bson:"password" json:"password"`   //
	Username  string `bson:"username" json:"username"`   //
	Phone     string `bson:"phone" json:"phone"`         //
	Authority string `bson:"authority" json:"authority"` //权限
	Position  string `bson:"position" json:"position"`   //职务
}

// 获取最新的主键ID
type CustomerOrderCount struct {
	NameField string
	Count     int
}

