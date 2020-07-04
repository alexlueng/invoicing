package models

type Warehouse struct {
	ID                 int64       `json:"warehouse_id" bson:"warehouse_id"`
	ComID              int64       `json:"com_id" bson:"com_id"`
	Name               string      `json:"warehouse_name" bson:"warehouse_name"`
	Address            string      `json:"warehouse_address" bson:"warehouse_address"`
	WarehouseAdminId   int64       `json:"warehouse_admin_id" bson:"warehouse_admin_id"`     //仓库管理员id
	WarehouseAdminName string      `json:"warehouse_admin_name" bson:"warehouse_admin_name"` //仓库管理员
	Phone              string      `json:"phone" bson:"phone"`
	Config             string      `json:"config" bson:"config"`
	WarehouseStuff     interface{} `json:"warehouse_stuff"`             // 仓库职员，不插入数据库
	CreateAt           int64       `json:"create_at"  bson:"create_at"` // 创建时间戳
	CreateBy           int64       `json:"create_by" bson:"create_by"`  // 创建者id
	ModifyAt           int64       `json:"modify_at" bson:"modify_at"`  // 最后修改时间戳
	ModifyBy           int64       `json:"modify_by" bson:"modify_by"`  // 最后修改者id
	Product            []int64     `json:"product" bson:"product"`      // 有新商品进入仓库，则在这个字段字段商品id
}

// 仓库职员表
type WarehouseStuff struct {
	ComID         int64  `json:"com_id" bson:"com_id"`                 //公司id
	UserId        int64  `json:"user_id" bson:"user_id"`               // 用户id
	Username      string `json:"username" bson:"username"`             // 用户名
	WarehouseId   int64  `json:"warehouse_id" bson:"warehouse_id"`     // 仓库id
	WarehouseName string `json:"warehouse_name" bson:"warehouse_name"` // 仓库名
}

//用户提交过来的数据
type WarehouseReq struct {
	BaseReq
	ID      int64  `json:"warehouse_id" form:"warehouse_id"`
	Name    string `json:"warehouse_name" form:"warehouse_name"`       //模糊搜索
	Address string `json:"warehouse_address" form:"warehouse_address"` //模糊搜索
	Manager int64  `json:"manager" form:"manager"`                     // 仓库管理员id
	Stuff   int64  `json:"stuff" form:"stuff"`                         //仓库职员id
}

type ResponseWarehouseData struct {
	Warehouses  []Warehouse `json:"warehouses"`
	Total       int         `json:"total"`
	Pages       int         `json:"pages"`
	Size        int         `json:"size"`
	CurrentPage int         `json:"current_page"`
}
