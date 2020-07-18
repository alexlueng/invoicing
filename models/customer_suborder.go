package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

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

func getCustomerSubOrderCollection() *mongo.Collection {
	return Client.Collection("customer_suborder")
}

type CustomerSubOrderResult struct {
	CustomerSubOrder []CustomerSubOrder `json:"customer_sub_order"`
}

func SelectCustomerSubOrderWithCondition(subOrderFilter bson.M) (*CustomerSubOrderResult, error) {
	cur, err := getCustomerSubOrderCollection().Find(context.TODO(), subOrderFilter)
	if err != nil {
		return nil, err
	}

	var res = new(CustomerSubOrderResult)
	for cur.Next(context.TODO()) {
		var c CustomerSubOrder
		if err := cur.Decode(&c); err != nil {
			return nil, err
		}
		res.CustomerSubOrder = append(res.CustomerSubOrder, c)
	}
	return res, nil
}

func MultiplyInsertCustomerSubOrder(subOrders []interface{}) error {
	_, err := getCustomerSubOrderCollection().InsertMany(context.TODO(), subOrders)
	return err
}

func SelectCustomerSubOrderByComIDAndOrderSN(comID int64, orderSN string)(*CustomerSubOrderResult, error) {
	filter := bson.M{}
	filter["com_id"] = comID
	filter["order_sn"] = orderSN
	return SelectCustomerSubOrderWithCondition(filter);
}

func UpdateCustomerSubOrderPrepared(comID, subOrderID int64) (*mongo.UpdateResult, error) {
	filter := bson.M{}
	filter["com_id"] = comID
	filter["sub_order_id"] = subOrderID
	return getCustomerSubOrderCollection().UpdateOne(context.TODO(), filter, bson.M{"$set" : bson.M{"is_prepare": true}})
}


