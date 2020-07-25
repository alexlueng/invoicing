package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	//"jxc/util"
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

// 预付订单 可存储在Redis中，待订单确认后再写入mongodb
type PreOrder struct {
	ComID        int64                       `json:"com_id" bson:"com_id"`
	OrderID      int64                       `json:"order_id" bson:"order_id"`
	OrderSN      string                      `json:"order_sn" bson:"order_sn"` // 订单号
	CustomerID   int64                       `json:"customer_id" bson:"customer_id"`
	CustomerName string                      `json:"customer_name" bson:"customer_name"`
	AddressID    int64                       `json:"address_id" bson:"address_id"`     // 收货地址
	DeliveryID   int64                       `json:"delivery_id" bson:"delivery_id"`   // 配送方式
	DeliveryFee  float64                     `json:"delivery_fee" bson:"delivery_fee"` // 配送费
	Items        []CustomerOrderProductsInfo `json:"items" bson:"items"`               // 订单商品项
	TotalPrice   float64                     `json:"total_price" bson:"total_price"`   // 总价
	PayWay       int64                       `json:"pay_way" bson:"pay_way"`           // 支付方式
	Comment      string                      `json:"comment" bson:"comment"`           // 备注
	FeedBack     string                      `json:"evaluation" bson:"evaluation"`     // 订单评价
	Status       int64                       `json:"status" bson:"status"`             // 订单状态
	IsPay        bool                        `json:"is_pay" bson:"is_pay"`             // 是否支付
	IsCancel     bool                        `json:"is_cancel" bson:"is_cancel"`       // 是否取消
	IsDelete     bool                        `json:"is_delete" bson:"is_delete"`       // 是否删除
	CreateAt     int64                       `json:"create_at" bson:"create_at"`       // 创建时间
	ExpireAt     int64                       `json:"create_at" bson:"expire_at"`       // 过期时间
	PayAt        int64                       `json:"pay_at" bson:"pay_at"`             // 支付时间
	ShipAt       int64                       `json:"delivery_at" bson:"delivery_at"`   // 发货时间
	FinishAt     int64                       `json:"finish_at" bson:"finish_at"`       // 完成时间
}

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
	ReceiverAddress       string                      `json:"receiver_address" bson:"receiver_address"` // 收货地址
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
	Status                int64                       `json:"status" bson:"status"`         // 订单状态，1：待发货 2:已发货 (2：配货中 3：配货完成) 4：审核通过（已打款） 5：审核不通过 6: 失效
	Products              []CustomerOrderProductsInfo `json:"products"`                     // 订单中的商品列表
	SubOrders             []CustomerSubOrder          `json:"sub_orders"`                   // 子订单
	OperatorID            int64                       `json:"operator_id"`                  // 本次订单的操作人，对应user_id
	TransportationExpense float64                     `json:"transportation_expense"`       // 邮费 此项如果为0，则为包邮，此字段不能为负数，应该对它进行检查，或者设为无符号数
	IsPrepare             bool                        `json:"is_prepare" bson:"is_prepare"` // 备货完成
	IsPay                 bool                        `json:"is_pay" bson:"is_pay"`         // 是否支付
	OrderType             int64                       `json:"order_type" bson:"order_type"` // 订单类型 1 微信小程序订单 2 平台下单
}

// 订单中包含的商品信息
type CustomerOrderProductsInfo struct {
	SubOrderSN string  `json:"sub_order_sn" bson:"sub_order_sn"`
	SubOrderId int64   `json:"sub_order_id" bson:"sub_order_id"`
	ProductID  int64   `json:"product_id" bson:"product_id"`
	Product    string  `json:"product" bson:"product"`
	Quantity   int64   `json:"quantity" bson:"amount"` //数量
	Price      float64 `json:"price" bson:"price"`
	Thumbnail  string  `json:"thumbnail" bson:"thumbnail"`
}

func PreOrderToCustomerOrderAndSubOrder(pre PreOrder) (*CustomerOrder, []interface{}, error) {
	order := CustomerOrder{
		ComID:                 pre.ComID,
		OrderId:               pre.OrderID,
		OrderSN:               pre.OrderSN,
		CustomerID:            pre.CustomerID,
		CustomerName:          pre.CustomerName,
		Contacts:              pre.CustomerName,
		Receiver:              pre.CustomerName,
		Phone:                 "",
		TotalPrice:            pre.TotalPrice,
		Amount:                0,
		OrderTime:             pre.CreateAt,
		Status:                pre.Status,
		TransportationExpense: pre.DeliveryFee,
		IsPay:                 pre.IsPay,
		Products:              pre.Items,
		OrderType:             1,
	}

	// 用户收货地址
	var address Address
	collection := Client.Collection("address")
	filter := bson.M{}
	filter["com_id"] = pre.ComID
	filter["address_id"] = pre.AddressID
	err := collection.FindOne(context.TODO(), filter).Decode(&address)
	if err != nil {
		return nil, nil, err
	}

	order.ReceiverAddress = address.Location
	order.Phone = address.Telephone

	var subOrders []interface{}
	var amount int64
	// 把订单中的每个子项插入到客户订单实例表中
	for _, item := range order.Products {
		var result CustomerSubOrder
		result.ComID = order.ComID

		//result.SubOrderId = util.GetLastID("sub_order")
		//result.SubOrderSn = conf.IdWorker.GetOrderSN(order.ComID, order.CustomerID)
		result.OrderId = order.OrderId
		result.CustomerID = order.CustomerID
		result.CustomerName = order.CustomerName
		result.OrderSN = order.OrderSN
		result.Product = item.Product
		result.Amount = item.Quantity
		result.Price = item.Price
		result.ProductID = item.ProductID
		result.Receiver = order.Receiver
		result.ReceiverPhone = order.Phone
		result.OrderTime = order.OrderTime
		result.Status = order.Status
		result.IsPrepare = false // 备货未完成

		subOrders = append(subOrders, result)
		// 订单总数量
		amount += item.Quantity
	}
	order.Amount = amount

	return &order, subOrders, nil
}

// 获取最新的主键ID
type CustomerOrderCount struct {
	NameField string
	Count     int
}

func getCustomerOrderCollection() *mongo.Collection {
	return Client.Collection("customer_order")
}

func (c *CustomerOrder) FindAll(filter bson.M, options *options.FindOptions) ([]CustomerOrder, error) {
	var result []CustomerOrder
	cur, err := getCustomerOrderCollection().Find(context.TODO(), filter, options)
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
	_, err := getCustomerOrderCollection().InsertOne(context.TODO(), c)
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

// 用来查找选中商品后对应客户的价格
type OrderProducts struct {
	ProductsID   []int64 `json:"products_id"`
	CustomerID   int64   `json:"customer_id"`
	CustomerName string  `json:"customer_name"`
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

func CountCustomerOrder(filter bson.M) (int64, error) {
	return getCustomerOrderCollection().CountDocuments(context.TODO(), filter)
}

func UpdateCustomerOrderByCondition(filter, updates bson.M) (*mongo.UpdateResult, error) {
	return getCustomerOrderCollection().UpdateOne(context.TODO(), filter, updates);
}

func DeleteCustomerOrderByComIDAndOrderSN(comID int64, orderSN string) (*mongo.DeleteResult, error) {
	filter := bson.M{}
	filter["com_id"] = comID
	filter["order_sn"] = orderSN
	return getCustomerOrderCollection().DeleteOne(context.TODO(), filter)
}

func SelectCustomerOrderByComIDAndOrderSN(comID int64, orderSN string) (CustomerOrder, error) {
	order := CustomerOrder{}
	filter := bson.M{}
	filter["com_id"] = comID
	filter["order_sn"] = orderSN
	err := getCustomerOrderCollection().FindOne(context.TODO(), filter).Decode(&order)
	return order, err
}
