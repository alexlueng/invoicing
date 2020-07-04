package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"time"
)

const (
	TOBECONFIRMED = iota + 1
	TOBEDELIVERED
	TOBEPAID
	CONFIRM
	FAILED
	INVALID
)

type CustomerOrder struct {
	ComID                 int64                       `json:"com_id" bson:"com_id"`
	OrderId               int64                       `json:"order_id" bson:"order_id"`
	OrderSN               string                      `json:"order_sn" bson:"order_sn"`
	WarehouseID           int64                       `json:"warehouse_id" bson:"warehouse_id"`
	SupplierOrderID       int64                       `json:"supplier_order_id" bson:"supplier_order_id"`
	CustomerID            int64                       `json:"customer_id" bson:"customer_id"`
	CustomerName          string                      `json:"customer_name" bson:"customer_name"`
	Contacts              string                      `json:"contacts" bson:"contacts"`
	Receiver              string                      `json:"receiver" bson:"receiver"`
	Phone                 string                      `json:"receiver_phone" bson:"receiver_phone"`
	TotalPrice            float64                     `json:"total_price" bson:"price"` // 订单总价
	Amount                int64                       `json:"amount" bson:"amount"`     // 订单总数量
	ExtraAmount           float64                     `json:"extra_amount" bson:"extra_amount"`
	Delivery              string                      `json:"delivery" bson:"delivery"`
	DeliveryCode          string                      `json:"delivery_code" bson:"delivery_code"`
	OrderTime             int64                       `json:"order_time" bson:"order_time"` // 所有的时间都是以int64的类型插入到mongodb中
	ShipTime              int64                       `json:"ship_time" bson:"ship_time"`
	ConfirmTime           int64                       `json:"confirm_time" bson:"confirm_time"`
	PayTime               int64                       `json:"pay_time" bson:"pay_time"`
	FinishTime            int64                       `json:"finish_time" bson:"finish_time"`
	Status                int64                       `json:"status" bson:"status"`         // 订单状态，1：待发货 2：配货中 3：配货完成 4：审核通过（已打款） 5：审核不通过 6: 失效
	Products              []CustomerOrderProductsInfo `json:"products"`                     // 订单中的商品列表
	SubOrders             []CustomerSubOrder          `json:"sub_orders"`                   // 子订单
	OperatorID            int64                       `json:"operator_id"`                  // 本次订单的操作人，对应user_id
	TransportationExpense float64                     `json:"transportation_expense"`       // 邮费 此项如果为0，则为包邮，此字段不能为负数，应该对它进行检查，或者设为无符号数
	IsPrepare             bool                        `json:"is_prepare" bson:"is_prepare"` // 备货完成
}

// 销售订单实例表结构
// 销售子订单
type CustomerSubOrder struct {
	SubOrderId      int64   `json:"sub_order_id" bson:"sub_order_id"` // 子订单id
	SubOrderSn      string  `json:"sub_order_sn" bson:"sub_order_sn"` // 子订单号
	ComID           int64   `json:"com_id" bson:"com_id"`             // 公司id
	OrderSN         string  `json:"order_sn" bson:"order_sn"`         // 订单号
	OrderId         int64   `json:"order_id" bson:"order_id"`         // 订单id
	CustomerID      int64   `json:"customer_id" bson:"customer_id"`
	CustomerName    string  `json:"customer_name" bson:"customer_name"`
	ProductID       int64   `json:"product_id" bson:"product_id"`
	Product         string  `json:"product" bson:"product"`                   // 商品名称
	Contacts        string  `json:"contacts" bson:"contacts"`                 //客户的联系人
	Receiver        string  `json:"receiver" bson:"receiver"`                 //本单的收货人
	ReceiverPhone   string  `json:"receiver_phone" bson:"receiver_phone"`     //本单的收货人电话
	Price           float64 `json:"price" bson:"price"`                       //本项价格
	Amount          int64   `json:"amount" bson:"amount"`                     //本项购买总数量
	WarehouseAmount int64   `json:"warehouse_amount" bson:"warehouse_amount"` // 仓库发货的数量
	SupplierAmount  int64   `json:"supplier_amount" bson:"supplier_amount"`   // 供应商发货的数量
	ExtraAmount     float64 `json:"extra_amount" bson:"extra_amount"`         //本单优惠或折扣金额
	Delivery        string  `json:"delivery" bson:"delivery"`                 // 快递方式
	DeliveryCode    string  `json:"delivery_code" bson:"delivery_code"`       // 快递号
	OrderTime       int64   `json:"order_time" bson:"order_time"`             // 下单时间
	ShipTime        int64   `json:"ship_time" bson:"ship_time"`               // 发货时间
	ConfirmTime     int64   `json:"confirm_time" bson:"confirm_time"`         // 确认订单时间
	PayTime         int64   `json:"pay_time" bson:"pay_time"`                 // 订单结算时间
	FinishTime      int64   `json:"finish_time" bson:"finish_time"`           // 供应结束时间
	Status          int64   `json:"status" bson:"status"`                     // 订单状态
	IsPrepare       bool    `json:"is_prepare" bson:"is_prepare"`             // 是否备货完成
}

func (c *CustomerOrder) FindAll(filter bson.M, options *options.FindOptions) ([]CustomerOrder, error) {
	var result []CustomerOrder
	cur, err := Client.Collection("customer_order").Find(context.TODO(), filter, options)
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var res CustomerOrder
		if err := cur.Decode(&res); err != nil {
			return nil, err
		}
		result = append(result, res)
	}
	return result, nil
}

func (c *CustomerOrder) Insert() error {
	_, err := Client.Collection("customer_order").InsertOne(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

type CustomerOrderReq struct {
	BaseReq
	OrderSN        string  `json:"order_sn" form:"order_sn"`
	CustomerName   string  `json:"customer_name" form:"customer_name"` //模糊搜索
	Contacts       string  `json:"contacts" form:"contacts"`           //模糊搜索
	Receiver       string  `json:"receiver" form:"receiver"`           //模糊搜索
	Delivery       string  `json:"delivery" form:"delivery"`
	ExtraAmount    float64 `json:"extra_amount" form:"extra_amount"`
	Status         int64   `json:"status" form:"status"`
	StartOrderTime string  `json:"start_order_time" form:"start_order_time"`
	EndOrderTime   string  `json:"end_order_time" form:"end_order_time"`
	StartPayTime   string  `json:"start_pay_time" form:"start_pay_time"`
	EndPayTime     string  `json:"end_pay_time" form:"end_pay_time"`
	StartShipTime  string  `json:"start_ship_time" form:"start_ship_time"`
	EndShipTime    string  `json:"end_ship_time" form:"end_ship_time"`
}

// 订单中包含的商品信息
type CustomerOrderProductsInfo struct {
	SubOrderSN string  `json:"sub_order_sn" bson:"sub_order_sn"`
	SubOrderId int64   `json:"sub_order_id" bson:"sub_order_id"`
	ProductID  int64   `json:"product_id" bson:"product_id"`
	Product    string  `json:"product"`
	Quantity   int64   `json:"quantity" bson:"amount"` //数量
	Price      float64 `json:"price"`
}

// 用来查找选中商品后对应客户的价格
type OrderProducts struct {
	ProductsID   []int64 `json:"products_id"`
	CustomerID   int64   `json:"customer_id"`
	CustomerName string  `json:"customer_name"`
}

// 接收查找到的对应客户的价格
type CustomerOrderProductPrice struct {
	ProductID   int64   `json:"product_id" bson:"product_id"`
	ProductName string  `json:"product_name" bson:"product_name"`
	Price       float64 `json:"price" bson:"price"`
}

type ResponseCustomerOrdersData struct {
	Result      interface{} `json:"result"`
	Total       int64       `json:"total"`
	Pages       int64       `json:"pages"`
	Size        int64       `json:"size"`
	CurrentPage int64       `json:"current_page"`
}

func GetCustomerOrderParam(req CustomerOrderReq, com_id int64) bson.M {
	filter := bson.M{}
	if req.OrderSN != "" {
		filter["order_sn"] = bson.M{"$regex": req.OrderSN}
	}
	if req.CustomerName != "" {
		filter["customer_name"] = bson.M{"$regex": req.CustomerName}
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
		stime, _ := strconv.Atoi(req.StartOrderTime)
		startOrderTime := int64(stime)
		if req.EndOrderTime != "" {
			stime, _ := strconv.Atoi(req.EndOrderTime)
			endOrderTime := int64(stime)
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
		stime, _ := strconv.Atoi(req.StartPayTime)
		startPayTime := int64(stime)
		if req.EndPayTime != "" {
			stime, _ := strconv.Atoi(req.StartPayTime)
			endOrderTime := int64(stime)
			filter["pay_time"] = bson.M{"$gte": startPayTime, "$lte": endOrderTime}
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
		stime, _ := strconv.Atoi(req.StartShipTime)
		startShipTime := int64(stime)
		if req.EndPayTime != "" {
			stime, _ := strconv.Atoi(req.StartShipTime)
			endShipTime := int64(stime)
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
