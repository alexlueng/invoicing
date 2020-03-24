package models

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	ComID           int64   `json:"com_id" bson:"com_id"`
	OrderId         int64   `json:"order_id" bson:"order_id"`
	OrderSN         string  `json:"order_sn" bson:"order_sn"`
	WarehouseID     int64   `json:"warehouse_id" bson:"warehouse_id"`
	SupplierOrderID int64   `json:"supplier_order_id" bson:"supplier_order_id"`
	CustomerID      int64   `json:"customer_id" bson:"customer_id"`
	CustomerName    string  `json:"customer_name" bson:"customer_name"`
	Contacts        string  `json:"contacts" bson:"contacts"`
	Receiver        string  `json:"receiver" bson:"receiver"`
	Phone           string  `json:"receiver_phone" bson:"receiver_phone"`
	TotalPrice      float64 `json:"total_price" bson:"price"` // 订单总价
	Amount          int64   `json:"amount" bson:"amount"`     // 订单总数量
	ExtraAmount     float64 `json:"extra_amount" bson:"extra_amount"`
	Delivery        string  `json:"delivery" bson:"delivery"`
	DeliveryCode    string  `json:"delivery_code" bson:"delivery_code"`
	OrderTime       int64   `json:"order_time" bson:"order_time"` // 所有的时间都是以int64的类型插入到mongodb中
	ShipTime        int64   `json:"ship_time" bson:"ship_time"`
	ConfirmTime     int64   `json:"confirm_time" bson:"confirm_time"`
	PayTime         int64   `json:"pay_time" bson:"pay_time"`
	FinishTime      int64   `json:"finish_time" bson:"finish_time"`
	Status          int64   `json:"status" bson:"status"` // 订单状态，1：待发货 2：待确认（已发货） 3：待付款（已确认） 4：审核通过（已打款） 5：审核不通过 6: 失效

	Products   []CustomerOrderProductsInfo `json:"products"`    // 订单中的商品列表
	SubOrders  []CustomerSubOrder          `json:"sub_orders"`  // 子订单
	OperatorID int64                       `json:"operator_id"` // 本次订单的操作人，对应user_id

	TransportationExpense float64 `json:"transportation_expense"` // 邮费 此项如果为0，则为包邮，此字段不能为负数，应该对它进行检查，或者设为无符号数
}

func (c *CustomerOrder) FindAll(filter bson.M, options *options.FindOptions) ([]CustomerOrder, error) {
	var result []CustomerOrder
	cur, err := Client.Collection("customer_order").Find(context.TODO(), filter, options)
	if err != nil {
		fmt.Println("Can't get customer order list")
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var res CustomerOrder
		if err := cur.Decode(&res); err != nil {
			fmt.Println("Can't decode into customer order")
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
	//本页面定制的搜索字段
	OrderSN        string  `json:"order_sn" form:"order_sn"`
	CustomerName   string  `json:"customer_name" form:"customer_name"` //模糊搜索
	Contacts       string  `json:"contacts" form:"contacts"`           //模糊搜索
	Receiver       string  `json:"receiver" form:"receiver"`           //模糊搜索
	Delivery       string  `json:"delivery" form:"delivery"`
	ExtraAmount    float64 `json:"extra_amount" form:"extra_amount"`
	Status         string  `json:"status" form:"status"`
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
	//CustomerOrders []CustomerOrder `json:"customer_orders"`
	//Products       []Product       `json:"product"`
	//Customers   []Customer `json:"customer"`
	Result      interface{} `json:"result"`
	Total       int64       `json:"total"`
	Pages       int64       `json:"pages"`
	Size        int64       `json:"size"`
	CurrentPage int64       `json:"current_page"`
}
