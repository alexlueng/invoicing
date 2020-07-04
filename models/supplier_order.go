package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"time"
)

// 采购订单与子订单的状态
//1、未发货
//2、已收货
//3、已发货
//4、部分收货
//5，审核通过
//6，审核不通过

// 采购订单表结构 (供应商订单)
type SupplierOrder struct {
	ComID         int64    `json:"com_id" bson:"com_id"`                 // 公司id
	OrderId       int64    `json:"order_id" bson:"order_id"`             // 订单id
	OrderSN       string   `json:"order_sn" bson:"order_sn"`             // 订单号
	SalesOrderSn  string   `json:"sales_order_sn" bson:"sales_order_sn"` // 销售单号
	WarehouseID   int64    `json:"warehouse_id" bson:"warehouse_id"`     // 仓库id
	WarehouseName string   `json:"warehouse_name" bson:"warehouse_name"` // 仓库名
	SupplierID    int64    `json:"supplier_id" bson:"supplier_id"`       // 供应商id
	Supplier      string   `json:"supplier" bson:"supplier"`             // 供应商名
	Contacts      string   `json:"contacts" bson:"contacts"`             // 供应商的联系人
	Receiver      string   `json:"receiver" bson:"receiver"`             // 本单的收货人
	ReceiverPhone string   `json:"receiver_phone" bson:"receiver_phone"` // 本单的收货人电话
	Price         float64  `json:"price" bson:"price"`                   // 本单总价格
	Amount        int64    `json:"amount" bson:"amount"`                 // 本单购买总数量 应发数量
	ExtraAmount   float64  `json:"extra_amount" bson:"extra_amount"`     // 本单优惠或折扣金额
	Delivery      string   `json:"delivery" bson:"delivery"`             // 快递方式
	DeliveryCode  string   `json:"delivery_code" bson:"delivery_code"`   // 快递号
	OrderTime     int64    `json:"order_time" bson:"order_time"`         // 下单时间
	CreateBy      int64    `json:"create_by" bson:"create_by"`           // 创建者id
	ShipTime      int64    `json:"ship_time" bson:"ship_time"`           // 发货时间
	Shipper       int64    `json:"shipper" bson:"shipper"`               // 发货者id
	ConfirmTime   int64    `json:"confirm_time" bson:"confirm_time"`     // 确认订单时间
	ConfirmBy     int64    `json:"confirm_by" bson:"confirm_by"`         // 确认收货者id
	PayTime       int64    `json:"pay_time" bson:"pay_time"`             // 订单结算时间
	PayBy         int64    `json:"pay_by" bson:"pay_by"`                 // 确认支付者id
	FinishTime    int64    `json:"finish_time" bson:"finish_time"`       // 供应结束时间
	Status        int64    `json:"status" bson:"status"`                 // 订单状态
	OrderURLs     []string `json:"order_urls" bson:"order_urls"`         // 订单凭证图片地址
}

// 采购订单实例（供应商订单实例）
// 采购订单子订单
type SupplierSubOrder struct {
	SubOrderId       int64   `json:"order_sub_id" bson:"order_sub_id"`             // 子订单id
	SubOrderSn       string  `json:"order_sub_sn" bson:"order_sub_sn"`             // 子订单号
	ComID            int64   `json:"com_id" bson:"com_id"`                         // 公司id
	OrderId          int64   `json:"order_id" bson:"order_id"`                     // 采购订单id
	OrderSn          string  `json:"order_sn" bson:"order_sn"`                     // 订单号
	ProductID        int64   `json:"product_id" bson:"product_id"`                 // 商品id
	ProductName      string  `json:"product_name" bson:"product_name"`             // 商品名 这是冗余字段
	ProductNum       int64   `json:"product_num" bson:"product_num"`               // 商品数量
	ProductUnitPrice float64 `json:"product_unit_price" bson:"product_unit_price"` // 商品单价
	Units            string  `json:"units" bson:"units"`                           // 商品量词，这是冗余字段
	CreateAt         int64   `json:"create_at"  bson:"create_at"`                  // 创建时间戳
	CreateBy         int64   `json:"create_by" bson:"create_by"`                   // 创建者id
	ShipTime         int64   `json:"ship_time" bson:"ship_time"`                   // 发货时间戳
	Shipper          int64   `json:"shipper" bson:"shipper"`                       // 发货者id
	ConfirmAt        int64   `json:"confirm_at" bson:"confirm_at"`                 // 确认收货时间戳
	ConfirmBy        int64   `json:"confirm_by" bson:"confirm_by"`                 // 确认收货者id
	CheckAt          int64   `json:"check_at" bson:"check_at"`                     // 盘点时间
	CheckBy          int64   `json:"check_by" bson:"check_by"`                     // 盘点操作者id
	State            int64   `json:"state" bson:"state"`                           // 订单状态
	ActualAmount     int64   `json:"actual_amount" bson:"actual_amount"`           // 实发数量
	FailReason       string  `json:"fail_reason" bson:"fail_reason"`               // 审核不通过的理由
}

type SupplierOrderReq struct {
	BaseReq
	//本页面定制的搜索字段
	OrderSN        string  `json:"order_sn"`
	SupplierName   string  `json:"supplier_name"` //模糊搜索
	Contacts       string  `json:"contacts"`           //模糊搜索
	Receiver       string  `json:"receiver"`           //模糊搜索
	Delivery       string  `json:"delivery"`
	ExtraAmount    float64 `json:"extra_amount"`
	Status         int64   `json:"status"`
	StartOrderTime string  `json:"start_order_time"`
	EndOrderTime   string  `json:"end_order_time"`
	StartPayTime   string  `json:"start_pay_time"`
	EndPayTime     string  `json:"end_pay_time"`
	StartShipTime  string  `json:"start_ship_time"`
	EndShipTime    string  `json:"end_ship_time"`
}

type ResponseSupplierOrdersData struct {
	SupplierOrders interface{} `json:"customer_orders"`
	Total          int         `json:"total"`
	Pages          int         `json:"pages"`
	Size           int         `json:"size"`
	CurrentPage    int         `json:"current_page"`
}

func getSupplierOrderCollection() *mongo.Collection {
	return Client.Collection("supplier_order")
}

func getSupplierSubOrderCollection() *mongo.Collection {
	return Client.Collection("supplier_sub_order")
}

// 根据订单状态找出相应的订单
func (so *SupplierOrder) FindOrders(filter bson.M) (result []SupplierOrder, err error) {

	collection := getSupplierOrderCollection()
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var res SupplierOrder
		if err := cur.Decode(&res); err != nil {
			return nil, err
		}
		result = append(result, res)
	}
	return
}

// 根据订单状态找出相应的订单
func (so *SupplierOrder) FindOneByOrderID(filter bson.M) (*SupplierOrder, error) {

	collection := getSupplierOrderCollection()

	err := collection.FindOne(context.TODO(), filter).Decode(&so)
	if err != nil {
		return nil, err
	}
	return so, nil
}

// 根据订单状态找出相应的订单
func (so *SupplierOrder) UpdateOneByOrderID(filter, update bson.M) error {

	collection := getSupplierOrderCollection()

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

// 根据总订单ID查找出所有子订单
func (so *SupplierOrder) FindSubOrderByOrderID(filter bson.M) ([]SupplierSubOrder, error) {
	collection := getSupplierOrderCollection()
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	var result []SupplierSubOrder

	for cur.Next(context.TODO()) {
		var res SupplierSubOrder
		if err := cur.Decode(&res); err != nil {
			return nil, err
		}
		result = append(result, res)
	}
	return result, nil
}

func GetSupplierOrderParam(req SupplierOrderReq, com_id int64) bson.M {
	filter := bson.M{}
	if req.OrderSN != "" {
		filter["order_sn"] = bson.M{"$regex": req.OrderSN}
	}
	if req.SupplierName != "" {
		filter["supplier"] = bson.M{"$regex": req.SupplierName}
	}
	if req.Contacts != "" {
		filter["contacts"] = bson.M{"$regex": req.Contacts}
	}
	if req.Receiver != "" {
		filter["receiver"] = bson.M{"$regex": req.Receiver}
	}
	if req.Delivery != "" {
		filter["delivery"] = bson.M{"$regex": req.Delivery}
	}
	if req.ExtraAmount != 0.0 {
		filter["extra_amount"] = bson.M{"$eq": req.ExtraAmount}
	}
	if req.Status > 0 {
		filter["status"] = req.Status
	}
	if req.StartOrderTime != "" {
		timestamp, _ := strconv.Atoi(req.StartOrderTime)
		startOrderTime := int64(timestamp)
		if req.EndOrderTime != "" {
			timestamp, _ := strconv.Atoi(req.EndOrderTime)
			endOrderTime := int64(timestamp)
			filter["order_time"] = bson.M{"$gte": startOrderTime, "$lte": endOrderTime}
		} else {
			filter["order_time"] = bson.M{"$gte": startOrderTime}
		}
	} else {
		if req.EndOrderTime != "" {
			current_time := time.Now()
			filter["order_time"] = bson.M{"$lte": current_time.UTC().UnixNano()}
		}
	}
	if req.StartPayTime != "" {
		timestamp, _ := strconv.Atoi(req.StartPayTime)
		startPayTime := int64(timestamp)
		if req.EndPayTime != "" {
			timestamp, _ := strconv.Atoi(req.EndPayTime)
			endShipTime := int64(timestamp)
			filter["pay_time"] = bson.M{"$gte": startPayTime, "$lte": endShipTime}
		} else {
			filter["pay_time"] = bson.M{"$gte": startPayTime}
		}
	} else {
		if req.EndPayTime != "" {
			current_time := time.Now()
			filter["pay_time"] = bson.M{"$lte": current_time.UTC().UnixNano()}
		}
	}
	if req.StartShipTime != "" {
		timestamp, _ := strconv.Atoi(req.StartShipTime)
		startShipTime := int64(timestamp)
		if req.EndPayTime != "" {
			timestamp, _ := strconv.Atoi(req.EndOrderTime)
			endShipTime := int64(timestamp)
			filter["ship_time"] = bson.M{"$gte": startShipTime, "$lte": endShipTime}
		} else {
			filter["ship_time"] = bson.M{"$gte": startShipTime}
		}
	} else {
		if req.EndPayTime != "" {
			current_time := time.Now()
			filter["ship_time"] = bson.M{"$lte": current_time.UTC().UnixNano()}
		}
	}

	filter["com_id"] = com_id

	return filter

}
