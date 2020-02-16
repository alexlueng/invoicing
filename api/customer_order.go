package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"jxc/models"
	"jxc/serializer"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type CustomerOrder struct {
	ComID int64 `json:"com_id" bson:"com_id"`
	OrderSN string `json:"order_sn" bson:"order_sn"`
	WarehouseID int64 `json:"warehouse_id" bson:"warehouse_id"`
	SupplierOrderID int64 `json:"supplier_order_id" bson:"supplier_order_id"`
	CustomerID int64 `json:"customer_id" bson:"customer_id"`
	CustomerName string `json:"customer_name" bson:"customer_name"`
	Contacts string `json:"contacts" bson:"contacts"`
	Receiver string  `json:"receiver" bson:"receiver"`
	Phone string `json:"receiver_phone" bson:"receiver_phone"`
	Price float64 `json:"price" bson:"price"`
	Amount int64 `json:"amount" bson:"amount"`
	ExtraAmount float64 `json:"extra_amount" bson:"extra_amount"`
	Delivery string `json:"delivery" bson:"delivery"`
	DeliveryCode string `json:"delivery_code" bson:"delivery_code"`
	OrderTime int64 `json:"order_time" bson:"order_time"` // 所有的时间都是以int64的类型插入到mongodb中
	ShipTime int64 `json:"ship_time" bson:"ship_time"`
	ConfirmTime int64 `json:"confirm_time" bson:"confirm_time"`
	PayTime int64 `json:"pay_time" bson:"pay_time"`
	FinishTime int64 `json:"finish_time" bson:"finish_time"`
	Status string `json:"status" bson:"status"`
}

type CustomerOrderReq struct {
	IdMin int `form:"idmin"` //okid界于[idmin 和 idmax] 之间的数据
	IdMax int `form:"idmax"` //ok
	//本页面的搜索字段 sf固定等于customer_name， key的值为用户提交过来的客户名关键字
	Key  string `form:"key"`              //用户提交过来的模糊搜索关键字
	Sf   string `form:"sf"`               //用户模糊搜索的字段  search field
	Page int64  `json:"page" form:"page"` //ok用户查询的是哪一页的数据
	Size int64  `json:"size" form:"size"` //ok用户希望每页展现多少条数据
	OrdF string `json:"ordf" form:"ordf"` //ok用户排序字段 order field
	Ord  string `json:"ord" form:"ord"`   //ok顺序还是倒序排列  ord=desc 倒序，ord = asc 升序
	TMin int    `form:"tmin"`             //时间最小值[tmin,tmax)
	TMax int    `form:"tmax"`             //时间最大值
	//本页面定制的搜索字段
	OrderSN string `json:"order_sn" form:"order_sn"`
	CustomerName      string `json:"customer_name" form:"customer_name"` //模糊搜索
	Contacts string `json:"contacts" form:"contacts"` //模糊搜索
	Receiver string `json:"receiver" form:"receiver"` //模糊搜索
	Delivery string `json:"delivery" form:"delivery"`
	ExtraAmount float64 `json:"extra_amount" form:"extra_amount"`
	Status string `json:"status" form:"status"`
	StartOrderTime string `json:"start_order_time" form:"start_order_time"`
	EndOrderTime string `json:"end_order_time" form:"end_order_time"`
	StartPayTime string `json:"start_pay_time" form:"start_pay_time"`
	EndPayTime string `json:"end_pay_time" form:"end_pay_time"`
	StartShipTime string `json:"start_ship_time" form:"start_ship_time"`
	EndShipTime string `json:"end_ship_time" form:"end_ship_time"`
}

type ResponseCustomerOrdersData struct {
	CustomerOrders   []CustomerOrder `json:"customer_orders"`
	Total       int        `json:"total"`
	Pages       int        `json:"pages"`
	Size        int        `json:"size"`
	CurrentPage int        `json:"current_page"`
}


func AllCustomerOrders(c *gin.Context) {

	//根据域名得到com_id
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	var orders []CustomerOrder
	var req CustomerOrderReq

	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	// 设置分页的默认值
	if req.Size < 11 {
		req.Size = 10
	}
	if req.Page < 2 {
		req.Page = 1
	}

	// 设置分页的默认值
	if req.Size < 11 {
		req.Size = 10
	}
	if req.Page < 2 {
		req.Page = 1
	}
	// 设置排序主键
	orderField := []string{"OrderSN", "price"}
	exist := false
	fmt.Println("order field: ", req.OrdF)
	for _, v := range orderField {
		if req.OrdF == v {
			exist = true
			break
		}
	}
	if !exist {
		req.OrdF = "OrderSN"
	}
	// 设置排序顺序 desc asc
	order := 1
	fmt.Println("order: ", req.Ord)
	if req.Ord == "desc" {
		order = -1
		req.Ord = "desc"
	} else {
		order = 1
		req.Ord = "asc"
	}

	option := options.Find()
	option.SetLimit(int64(req.Size))
	option.SetSkip((int64(req.Page) - 1) * int64(req.Size))
	option.SetSort(bson.D{{req.OrdF, order}})

	//设置搜索规则
	filter := bson.M{}
	//OrderSN string `json:"order_sn" form:"order_sn"`
	if req.OrderSN != "" {
		filter["order_sn"] = bson.M{"$regex": req.OrderSN}
	}
	//CustomerName      string `json:"customer_name" form:"customer_name"` //模糊搜索
	if req.CustomerName != "" {
		filter["customer_name"] = bson.M{"$regex": req.CustomerName}
	}
	//Contacts string `json:"contacts" form:"contacts"` //模糊搜索
	if req.Contacts != "" {
		filter["contacts"] = bson.M{"$regex": req.Contacts}
	}
	//Receiver string `json:"receiver" form:"receiver"` //模糊搜索
	if req.Receiver != "" {
		filter["receiver"] = bson.M{"$regex": req.Receiver}
	}
	//Delivery string `json:"delivery" form:"delivery"`
	if req.Delivery != "" {
		filter["delivery"] = bson.M{"$regex": req.Delivery}
	}
	//ExtraAmount float64 `json:"extra_amount" form:"extra_amount"`
	if req.ExtraAmount != 0.0 {
		filter["extra_amount"] = bson.M{"$eq": req.ExtraAmount}
	}
	//Status string `json:"status" form:"status"`
	if req.Status != "" {
		filter["status"] = bson.M{"$regex": req.Status}
	}
	//db.getCollection('collection_name').find({"create_time":{"$gte":ISODate("2019-09-30 00:00:00"), "$lt": ISODate("2019-09-30 23:59:59")}})
	//StartOrderTime time.Time `json:"start_order_time" form:"start_order_time"`
	//EndOrderTime time.Time `json:"end_order_time" form:"end_order_time"`
	// 根据时间来查找订单的几个条件：
	// 1.开始、结束时间都传了 2. 只有开始时间，没有结束时间 3. 只有结束时间，没有开始时间
	if req.StartOrderTime != "" {
		time, _ := strconv.Atoi(req.StartOrderTime)
		startOrderTime := int64(time)
		if req.EndOrderTime != "" {
			time, _ := strconv.Atoi(req.StartOrderTime)
			endOrderTime := int64(time)
			filter["order_time"] = bson.M{"$gte":startOrderTime, "$lte": endOrderTime}
		} else {
			filter["order_time"] = bson.M{"$gte":startOrderTime}
		}
	} else {
		if req.EndOrderTime != "" {
			current_time := time.Now()
			filter["order_time"] = bson.M{"$lte": current_time.UTC().UnixNano()}
		}
	}
	//StartPayTime time.Time `json:"start_pay_time" form:"start_pay_time"`
	//EndPayTime time.Time `json:"end_pay_time" form:"end_pay_time"`
	if req.StartPayTime != "" {
		time, _ := strconv.Atoi(req.StartPayTime)
		startPayTime := int64(time)
		if req.EndPayTime != "" {
			time, _ := strconv.Atoi(req.StartPayTime)
			endOrderTime := int64(time)
			filter["pay_time"] = bson.M{"$gte":startPayTime, "$lte": endOrderTime}
		} else {
			filter["pay_time"] = bson.M{"$gte":startPayTime}
		}
	} else {
		if req.EndPayTime != "" {
			current_time := time.Now()
			filter["pay_time"] = bson.M{"$lte": current_time.UTC().UnixNano()}
		}
	}
	//StartShipTime time.Time `json:"start_ship_time" form:"start_ship_time"`
	//EndShipTime time.Time `json:"end_ship_time" form:"end_ship_time"`
	if req.StartShipTime != "" {
		time, _ := strconv.Atoi(req.StartShipTime)
		startShipTime := int64(time)
		if req.EndPayTime != "" {
			time, _ := strconv.Atoi(req.StartShipTime)
			endShipTime := int64(time)
			filter["ship_time"] = bson.M{"$gte":startShipTime, "$lte": endShipTime}
		} else {
			filter["ship_time"] = bson.M{"$gte":startShipTime}
		}
	} else {
		if req.EndPayTime != "" {
			current_time := time.Now()
			filter["ship_time"] = bson.M{"$lte": current_time.UTC().UnixNano()}
		}
	}

	filter["com_id"] = com.ComId
	fmt.Println("filter: ", filter)

	collection := models.Client.Collection("customer_order")
	cur, err := collection.Find(context.TODO(), filter, option)
	if err != nil {
		fmt.Println("error found finding customer orders: ", err)
		return
	}
	for cur.Next(context.TODO()) {
		var result CustomerOrder
		err := cur.Decode(&result)
		if err != nil {
			fmt.Println("error found decoding customer order: ", err)
			return
		}
		orders = append(orders, result)
	}

	//查询的总数
	var total int64
	cur, _ = models.Client.Collection("customer_order").Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		total++
	}

	// 返回查询到的总数，总页数
	resData := ResponseCustomerOrdersData{}
	resData.CustomerOrders = orders
	//	total, _ = models.Client.Collection("customer").CountDocuments(context.Background(), bson.D{})
	resData.Total = int(total)
	resData.Pages = int(total)/int(req.Size) + 1
	resData.Size = int(req.Size)
	resData.CurrentPage = int(req.Page)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get all customer orders",
		Data: resData,
	})
}

func SmartPrint(i interface{}){
	var kv = make(map[string]interface{})
	vValue := reflect.ValueOf(i)
	vType :=reflect.TypeOf(i)
	for i:=0; i < vValue.NumField(); i++{
		kv[vType.Field(i).Name] = vValue.Field(i)
	}
	fmt.Println("获取到数据:")
	for k,v := range kv {
		fmt.Print(k)
		fmt.Print(":")
		fmt.Print(v)
		fmt.Println()
	}
}


func AddCustomerOrder(c *gin.Context) {

	//根据域名得到com_id
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	order := CustomerOrder{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Println("Get customer_order data: ", string(data))
	err = json.Unmarshal(data, &order)
	if err != nil {
		fmt.Println("unmarshall error: ", err)
	}

	//这里需要一个订单号生成方法，日期加上6位数的编号,这个订单编号应该是全局唯一的
	order.OrderSN = GetTempOrderSN()
	order.ComID = com.ComId

	SmartPrint(order)

	// 创建订单的时间，以int64的类型插入到mongodb
	// TODO: 把这个方法独立出来
	current_time := time.Now()
	order.OrderTime = current_time.UTC().UnixNano()
	fmt.Println("order_time: ", order.OrderTime)

	collection := models.Client.Collection("customer_order")
	insertResult, err := collection.InsertOne(context.TODO(), order)
	if err != nil {
		fmt.Println("Error while inserting mongo: ", err)
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Customer order create succeeded",
	})
}

func UpdateCustomerOrder(c *gin.Context) {
	// 根据域名得到com_id
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	updateCustomerOrder := CustomerOrder{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	err = json.Unmarshal(data, &updateCustomerOrder)
	if err != nil {
		fmt.Println("unmarshall error: ", err)
	}
	collection := models.Client.Collection("customer_order")

	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["order_sn"] = updateCustomerOrder.OrderSN
	// 更新记录
	result, err := collection.UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"customer_name": updateCustomerOrder.CustomerName,
			"contacts": updateCustomerOrder.Contacts,
			"receiver_phone": updateCustomerOrder.Phone,
			"amount": updateCustomerOrder.Amount,
			"Delivery": updateCustomerOrder.Delivery,
			"warehouse_id": updateCustomerOrder.WarehouseID,
			"receiver": updateCustomerOrder.Receiver,
			"price": updateCustomerOrder.Price,
			"extra_amount": updateCustomerOrder.ExtraAmount,
			"delivery_code": updateCustomerOrder.DeliveryCode,}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "更新失败",
		})
		return
	}
	fmt.Println("Update result: ", result.UpsertedID)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Customer order update succeeded",
	})

}

type GetCustomerOrderSNService struct {
	OrderSN string `json:"order_sn" form:"order_sn"`
}

func DeleteCustomerOrder(c *gin.Context) {
	// 根据域名得到com_id
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}
	order_sn := GetCustomerOrderSNService{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &order_sn)

	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["order_sn"] =order_sn.OrderSN

	collection := models.Client.Collection("customer_order")
	deleteResult, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "删除客户订单失败",
		})
		return
	}
	fmt.Println("Delete a single document: ", deleteResult.DeletedCount)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Customer order delete succeeded",
	})

}


func CustomerOrderDetail(c *gin.Context) {

	// 根据域名得到com_id
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}
	order := CustomerOrder{}
	order_sn := GetCustomerOrderSNService{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	err = json.Unmarshal(data, &order_sn)
	if err != nil {
		fmt.Println("error found: ", err)
	}
	SmartPrint(order_sn)
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["order_sn"] = order_sn.OrderSN

	collection := models.Client.Collection("customer_order")
	err = collection.FindOne(context.TODO(), filter).Decode(&order)
	if err != nil {
		fmt.Println("error found while find order detail: ", err)
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "查看订单详情错误",
		})
		return
	}
	fmt.Println("Find a customer order")
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Customer order detail response",
		Data: order,
	})
}



