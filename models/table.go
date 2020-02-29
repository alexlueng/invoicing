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
}

// 公司表数据结构
type Company struct {
	ComId          int64  `json:"com_id" bson:"com_id"`
	ComName        string `json:"com_name" bson:"com_name"`
	ExpirationDate int64  `json:"expiration_date" bson:"expiration_date"` // 到期时间
	//Delivery       interface{} `json:"delivery" bson:"delivery"`               // 配送方式
	Units               interface{} `json:"units" bson:"units"`                                 //计量单位
	Payment             interface{} `json:"payment" bson:"payment"`                             //结算方式
	Module              string      `json:"module" bson:"module"`                               //平台名称
	Developer           string      `json:"developer" bson:"developer"`                         //开发名称
	Position            interface{} `bson:"position" json:"position"`                           //职务
	DefaultProfitMargin int64       `json:"default_profit_margin" bson:"default_profit_margin"` //默认利润率
}

// 配送方式数据格式
type Delivery struct {
	DeliveryId     int64  `json:"delivery_id" bson:"delivery_id"`        //配送方式id
	ComId          int64  `json:"com_id" bson:"comid"`                   //公司id
	DeliveryCom    string `json:"delivery_com" bson:"deliverycom"`       // 配送公司
	DeliveryPerson string `json:"delivery_person" bson:"deliveryperson"` // 配送员
	Phone          string `json:"phone" bson:"phone"`                    // 配送员电话
	Config         string `json:"config" bson:"config"`                  // 配置参数
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
	AuthId  int64  `json:"authid" bson:"authid"` // 节点id
	Note    string `json:"note" bson:"note"`     // 节点名
	Group   string `json:"group" bson:"group"`   // 组名
	GroupId int64  `json:"groupid" bson:"groupid"`
}

// 路由信息
type Router struct {
	RouterId   int64  `json:"routerid" bson:""` // 路由id
	RouterName string `json:"router_name"`      // 路由名
	Router     string `json:"router"`           // 访问路径
}

// 获取最新的主键ID
type CustomerOrderCount struct {
	NameField string
	Count     int
}

type Warehouse struct {
	ID                 int64  `json:"warehouse_id" bson:"warehouse_id"`
	ComID              int64  `json:"com_id" bson:"com_id"`
	Name               string `json:"warehouse_name" bson:"warehouse_name"`
	Address            string `json:"warehouse_address" bson:"warehouse_address"`
	WarehouseAdminId   int64  `json:"warehouse_admin_id" bson:"warehouse_admin_id"`     //仓库管理员id
	WarehouseAdminName string `json:"warehouse_admin_name" bson:"warehouse_admin_name"` //仓库管理员
	Phone              string `json:"phone" bson:"phone"`
	Config             string `json:"config" bson:"config"`

	WarehouseStuff interface{} `json:"warehouse_stuff"` // 仓库职员，不插入数据库

	CreateAt int64   `json:"create_at"  bson:"create_at"` // 创建时间戳
	CreateBy int64   `json:"create_by" bson:"create_by"`  // 创建者id
	ModifyAt int64   `json:"modify_at" bson:"modify_at"`  // 最后修改时间戳
	ModifyBy int64   `json:"modify_by" bson:"modify_by"`  // 最后修改者id
	Product  []int64 `json:"product" bson:"product"`      // 有新商品进入仓库，则在这个字段字段商品id
}

// 仓库职员表
type WarehouseStuff struct {
	ComID         int64  `json:"com_id" bson:"com_id"`                 //公司id
	UserId        int64  `json:"user_id" bson:"user_id"`               // 用户id
	Username      string `json:"username" bson:"username"`             // 用户名
	WarehouseId   int64  `json:"warehouse_id" bson:"warehouse_id"`     // 仓库id
	WarehouseName string `json:"warehouse_name" bson:"warehouse_name"` // 仓库名
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

	// 审核时间

	//  如果是仓库发货给客户，则记录发货信息
}

// 通用的实例表
// 在仓库实例基础上增加字段
// 所有的订单实例都会来这
//
// 类型 仓库发、供应商发、退货（客户发）、
// common
// 销售订单流程
//
//
//
//

// 采购订单表结构 (供应商订单)
type SupplierOrder struct {
	ComID         int64   `json:"com_id" bson:"com_id"`                 // 公司id
	OrderId       int64   `json:"order_id" bson:"order_id"`             // 订单id
	OrderSN       string  `json:"order_sn" bson:"order_sn"`             // 订单号
	SalesOrderSn  string  `json:"sales_order_sn" bson:"sales_order_sn"` // 销售单号
	WarehouseID   int64   `json:"warehouse_id" bson:"warehouse_id"`     // 仓库id
	WarehouseName string  `json:"warehouse_name" bson:"warehouse_name"` // 仓库名
	SupplierID    int64   `json:"supplier_id" bson:"supplier_id"`       // 供应商id
	Contacts      string  `json:"contacts" bson:"contacts"`             //供应商的联系人
	Receiver      string  `json:"receiver" bson:"receiver"`             //本单的收货人
	ReceiverPhone string  `json:"receiver_phone" bson:"receiver_phone"` //本单的收货人电话
	Price         float64 `json:"price" bson:"price"`                   //本单总价格
	Amount        int64   `json:"amount" bson:"amount"`                 //本单购买总数量
	ExtraAmount   float64 `json:"extra_amount" bson:"extra_amount"`     //本单优惠或折扣金额
	Delivery      string  `json:"delivery" bson:"delivery"`             // 快递方式
	DeliveryCode  string  `json:"delivery_code" bson:"delivery_code"`   // 快递号
	OrderTime     int64   `json:"order_time" bson:"order_time"`         // 下单时间
	CreateBy      int64   `json:"create_by" bson:"create_by"`           // 创建者id
	ShipTime      int64   `json:"ship_time" bson:"ship_time"`           // 发货时间
	Shipper       int64   `json:"shipper" bson:"shipper"`               // 发货者id
	ConfirmTime   int64   `json:"confirm_time" bson:"confirm_time"`     // 确认订单时间
	ConfirmBy     int64   `json:"confirm_by" bson:"confirm_by"`         // 确认收货者id
	PayTime       int64   `json:"pay_time" bson:"pay_time"`             // 订单结算时间
	PayBy         int64   `json:"pay_by" bson:"pay_by"`                 // 确认支付者id
	FinishTime    int64   `json:"finish_time" bson:"finish_time"`       // 供应结束时间
	Status        int64   `json:"status" bson:"status"`                 // 订单状态
}

// 采购订单实例（供应商订单实例）
// 采购订单子订单
type SupplierSubOrder struct {
	SubOrderId int64  `json:"order_sub_id" bson:"order_sub_id"` // 子订单id
	SubOrderSn string `json:"order_sub_sn" bson:"order_sub_sn"` // 子订单号
	ComID      int64  `json:"com_id" bson:"com_id"`             // 公司id
	OrderId    int64  `json:"order_id" bson:"order_id"`         // 采购订单id
	OrderSn    string `json:"order_sn" bson:"order_sn"`         // 订单号

	ProductID        int64   `json:"product_id" bson:"product_id"`                 // 商品id
	ProductName      string  `json:"product_name" bson:"product_name"`             // 商品名 这是冗余字段
	ProductNum       int64   `json:"product_num" bson:"product_num"`               // 商品数量
	ProductUnitPrice float64 `json:"product_unit_price" bson:"product_unit_price"` // 商品单价
	Units            string  `json:"units" bson:"units"`                           // 商品量词，这是冗余字段
	CreateAt         int64   `json:"create_at"  bson:"create_at"`                  // 创建时间戳
	CreateBy         int64   `json:"create_by" bson:"create_by"`                   // 创建者id

	ShipTime  int64 `json:"ship_time" bson:"ship_time"`   // 发货时间戳
	Shipper   int64 `json:"shipper" bson:"shipper"`       // 发货者id
	ConfirmAt int64 `json:"confirm_at" bson:"confirm_at"` // 确认收货时间戳
	ConfirmBy int64 `json:"confirm_by" bson:"confirm_by"` // 确认收货者id
	CheckAt   int64 `json:"check_at" bson:"check_at"`     // 盘点时间
	CheckBy   int64 `json:"check_by" bson:"check_by"`     // 盘点操作者id
	State     int64 `json:"state" bson:"state"`           // 订单状态
}

// 销售订单实例表结构
// 销售子订单
type CustomerSubOrder struct {
	SubOrderId int64  `json:"sub_order_id" bson:"sub_order_id"` // 子订单id
	SubOrderSn string `json:"sub_order_sn" bson:"sub_order_sn"` // 子订单号
	ComID      int64  `json:"com_id" bson:"com_id"`             // 公司id
	OrderSN    string `json:"order_sn" bson:"order_sn"`         // 订单号
	OrderId    int64  `json:"order_id" bson:"order_id"`         // 订单id

	CustomerID    int64   `json:"customer_id" bson:"customer_id"`
	CustomerName    string   `json:"customer_name" bson:"customer_name"`
	ProductID     int64   `json:"product_id" bson:"product_id"`
	Product       string  `json:"product" bson:"product"`               // 商品名称
	Contacts      string  `json:"contacts" bson:"contacts"`             //客户的联系人
	Receiver      string  `json:"receiver" bson:"receiver"`             //本单的收货人
	ReceiverPhone string  `json:"receiver_phone" bson:"receiver_phone"` //本单的收货人电话
	Price         float64 `json:"price" bson:"price"`                   //本项价格
	Amount        int64   `json:"amount" bson:"amount"`                 //本项购买总数量
	ExtraAmount   float64 `json:"extra_amount" bson:"extra_amount"`     //本单优惠或折扣金额
	Delivery      string  `json:"delivery" bson:"delivery"`             // 快递方式
	DeliveryCode  string  `json:"delivery_code" bson:"delivery_code"`   // 快递号
	OrderTime     int64   `json:"order_time" bson:"order_time"`         // 下单时间
	ShipTime      int64   `json:"ship_time" bson:"ship_time"`           // 发货时间
	ConfirmTime   int64   `json:"confirm_time" bson:"confirm_time"`     // 确认订单时间
	PayTime       int64   `json:"pay_time" bson:"pay_time"`             // 订单结算时间
	FinishTime    int64   `json:"finish_time" bson:"finish_time"`       // 供应结束时间
	Status        int64   `json:"status" bson:"status"`                 // 订单状态

	// operator_id操作人
	// 如何拆分子订单
}

type SupplierOrder_bak struct {
	ComID       int64  `json:"com_id" bson:"com_id"`
	OrderSN     string `json:"order_sn" bson:"order_sn"`
	WarehouseID int64  `json:"warehouse_id" bson:"warehouse_id"`
	//SupplierOrderID int64 `json:"supplier_order_id" bson:"supplier_order_id"`
	SupplierID int64 `json:"supplier_id" bson:"supplier_id"`
	//SupplierName string `json:"supplier_name" bson:"supplier_name"`
	Contacts     string  `json:"contacts" bson:"contacts"`
	Receiver     string  `json:"receiver" bson:"receiver"`
	Phone        string  `json:"receiver_phone" bson:"receiver_phone"`
	Price        float64 `json:"price" bson:"price"`
	Amount       int64   `json:"amount" bson:"amount"`
	ExtraAmount  float64 `json:"extra_amount" bson:"extra_amount"`
	Delivery     string  `json:"delivery" bson:"delivery"`
	DeliveryCode string  `json:"delivery_code" bson:"delivery_code"`
	OrderTime    int64   `json:"order_time" bson:"order_time"` // 所有的时间都是以int64的类型插入到mongodb中
	ShipTime     int64   `json:"ship_time" bson:"ship_time"`
	ConfirmTime  int64   `json:"confirm_time" bson:"confirm_time"`
	PayTime      int64   `json:"pay_time" bson:"pay_time"`
	FinishTime   int64   `json:"finish_time" bson:"finish_time"`
	Status       string  `json:"status" bson:"status"`
}
