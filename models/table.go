package models

// 所有的数据表格式都会存放在这里

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

// 公司表数据结构
type Company struct {
	ComId               int64       `json:"com_id" bson:"com_id"`
	ComName             string      `json:"com_name" bson:"com_name"`
	ExpireAt            int64       `json:"expire_at" bson:"expire_at"`                         // 到期时间
	CreateAt            int64       `json:"create_at" bson:"create_at"`                         // 创建时间
	Units               interface{} `json:"units" bson:"units"`                                 //计量单位
	Payment             interface{} `json:"payment" bson:"payment"`                             //结算方式
	Module              string      `json:"module" bson:"module"`                               //平台名称
	Developer           string      `json:"developer" bson:"developer"`                         //开发名称
	Position            interface{} `bson:"position" json:"position"`                           //职务
	DefaultProfitMargin float64     `json:"default_profit_margin" bson:"default_profit_margin"` //默认利润率
	Admin               string      `json:"admin" bson:"admin"`
	Telephone           string      `json:"phone" bson:"phone"`
	Password            string      `json:"password" bson:"password"`
}

// 支付方式
type Payment struct {
	ComId       int64  `json:"com_id" bson:"com_id"`
	PaymentId   int64  `json:"payment_id" bson:"payment_id"`     // 支付方式id
	PaymentName string `json:"payment_name" bson:"payment_name"` // 支付方式名称
	Days        int64  `json:"days" bson:"days"`                 // 天数
}

// 域名表
type Domain struct {
	ComId    int64  `bson:"comid" json:"comid"`       // 公司id
	Domain   string `bson:"domain" json:"domain"`     //
	ModuleId int64  `bson:"moduleid" json:"moduleid"` //
	Status   bool   `bson:"status" json:"status"`     // 域名可用状态，false情况下无法登录
}

// 职位表
type Position struct {
	ComId      int64  `json:"comid" bson:"comid"`
	PositionId int64  `json:"position_id" bson:"position_id"`
	Position   string `json:"position" bson:"position"`
}

// 权限节点
// 本模块的所有公司通用这段数据，所以不添加comid
type AuthNote struct {
	AuthId  int64    `json:"auth_id" bson:"auth_id"`   // 节点id
	Note    string   `json:"note" bson:"note"`         // 节点名
	Group   string   `json:"group" bson:"group"`       // 组名
	GroupId int64    `json:"group_id" bson:"group_id"` // 权限节点组id，5为仓库权限
	Urls    []string `json:"urls" bson:"urls"`         // 这里记录了这个节点所有的路由
}

// 路由信息
type Router struct {
	RouterId   int64  `json:"router_id" bson:"router_id"`     // 路由id
	RouterName string `json:"router_name" bson:"router_name"` // 路由名
	Url        string `json:"url" bson:"url"`                 //访问路径
}

// 获取最新的主键ID
type CustomerOrderCount struct {
	NameField string
	Count     int
}

// 库存实例表
// 类型 = 0 凭空多出的商品，采购订单号、销售订单号均为空 +
// 类型 = 1 退货 销售订单退货造成库存增多，订单号记录到销售订单号 +
// 类型 = 2 销售 销售后发货造成仓库库存减少，订单号记录到销售订单号 —
// 类型 = 3 损耗 盘点仓库，商品损耗造成库存减少 -
// 类型 = 4 采购 采购商品造成库存增多，订单号记录到采购订单号 +
type WosInstance struct {
	ComID            int64   `json:"com_id" bson:"com_id"`                         // 公司id
	Type             int64   `json:"type" bson:"type"`                             // 类型
	WarehouseID      int64   `json:"warehouse_id" bson:"warehouse_id"`             // 仓库id
	WarehouseName    string  `json:"warehouse_name" bson:"warehouse_name"`         // 仓库名，这是冗余字段
	PurchaseOrderSn  string  `json:"purchase_order_sn" bson:"purchase_order_sn"`   // 采购订单号
	SalesOrderSn     string  `json:"sales_order_sn" bson:"sales_order_sn"`         // 销售订单号
	ProductID        int64   `json:"product_id" bson:"product_id"`                 // 商品id
	ProductName      string  `json:"product_name" bson:"product_name"`             // 商品名
	Units            string  `json:"units" bson:"units"`                           // 商品量词，这是冗余字段
	ProductNum       int64   `json:"product_num" bson:"product_num"`               // 商品数量
	ProductUnitPrice float64 `json:"product_unit_price" bson:"product_unit_price"` // s
	CreateAt         int64   `json:"create_at"  bson:"create_at"`                  // 创建时间戳
	CreateBy         int64   `json:"create_by" bson:"create_by"`                   // 创建者id
	ShipTime         int64   `json:"ship_time" bson:"ship_time"`                   // 发货时间戳
	Shipper          int64   `json:"shipper" bson:"shipper"`                       // 发货者id
	ConfirmAt        int64   `json:"confirm_at" bson:"confirm_at"`                 // 确认收货时间戳
	ConfirmBy        int64   `json:"confirm_by" bson:"confirm_by"`                 // 确认收货者id
	CheckAt          int64   `json:"check_at" bson:"check_at"`                     // 盘点时间
	CheckBy          int64   `json:"check_by" bson:"check_by"`                     // 盘点操作者id
}
