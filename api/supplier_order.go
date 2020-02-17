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
	"strconv"
	"strings"
	"time"
)

func AllSupplierOrders(c *gin.Context) {
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

	var orders []models.SupplierOrder
	var req models.SupplierOrderReq

	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	req.Page, req.Size = SetDefaultPageAndSize(req.Page, req.Size)

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
	//supplierName      string `json:"supplier_name" form:"supplier_name"` //模糊搜索
	if req.SupplierName != "" {
		filter["supplier_name"] = bson.M{"$regex": req.SupplierName}
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

	collection := models.Client.Collection("supplier_order")
	cur, err := collection.Find(context.TODO(), filter, option)
	if err != nil {
		fmt.Println("error found finding supplier orders: ", err)
		return
	}
	for cur.Next(context.TODO()) {
		var result models.SupplierOrder
		err := cur.Decode(&result)
		if err != nil {
			fmt.Println("error found decoding supplier order: ", err)
			return
		}
		orders = append(orders, result)
	}

	//查询的总数
	var total int64
	cur, _ = models.Client.Collection("supplier_order").Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		total++
	}

	// 返回查询到的总数，总页数
	resData := models.ResponseSupplierOrdersData{}
	resData.SupplierOrders = orders
	//	total, _ = models.Client.Collection("supplier").CountDocuments(context.Background(), bson.D{})
	resData.Total = int(total)
	resData.Pages = int(total)/int(req.Size) + 1
	resData.Size = int(req.Size)
	resData.CurrentPage = int(req.Page)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get all supplier orders",
		Data: resData,
	})
}
func AddSupplierOrder(c *gin.Context) {
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

	order := models.SupplierOrder{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Println("Get supplier_order data: ", string(data))
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

	collection := models.Client.Collection("supplier_order")
	insertResult, err := collection.InsertOne(context.TODO(), order)
	if err != nil {
		fmt.Println("Error while inserting mongo: ", err)
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Supplier order create succeeded",
	})
}
func UpdateSupplierOrder(c *gin.Context) {
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

	updateSupplierOrder := models.SupplierOrder{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	err = json.Unmarshal(data, &updateSupplierOrder)
	if err != nil {
		fmt.Println("unmarshall error: ", err)
	}
	collection := models.Client.Collection("supplier_order")

	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["order_sn"] = updateSupplierOrder.OrderSN
	// 更新记录
	result, err := collection.UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"supplier_name": updateSupplierOrder.SupplierName,
			"contacts": updateSupplierOrder.Contacts,
			"receiver_phone": updateSupplierOrder.Phone,
			"amount": updateSupplierOrder.Amount,
			"Delivery": updateSupplierOrder.Delivery,
			"warehouse_id": updateSupplierOrder.WarehouseID,
			"receiver": updateSupplierOrder.Receiver,
			"price": updateSupplierOrder.Price,
			"extra_amount": updateSupplierOrder.ExtraAmount,
			"delivery_code": updateSupplierOrder.DeliveryCode,}})
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
		Msg:  "Supplier order update succeeded",
	})
}

type GetSupplierOrderSNService struct {
	OrderSN string `json:"order_sn" form:"order_sn"`
}

func DeleteSupplierOrder(c *gin.Context) {
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
	order_sn := GetSupplierOrderSNService{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &order_sn)

	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["order_sn"] =order_sn.OrderSN

	collection := models.Client.Collection("supplier_order")
	deleteResult, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "删除供应商订单失败",
		})
		return
	}
	fmt.Println("Delete a single document: ", deleteResult.DeletedCount)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Supplier order delete succeeded",
	})
}
func SupplierOrderDetail(c *gin.Context) {
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
	order := models.SupplierOrder{}
	order_sn := GetSupplierOrderSNService{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	err = json.Unmarshal(data, &order_sn)
	if err != nil {
		fmt.Println("error found: ", err)
	}
	SmartPrint(order_sn)
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["order_sn"] = order_sn.OrderSN

	collection := models.Client.Collection("supplier_order")
	err = collection.FindOne(context.TODO(), filter).Decode(&order)
	if err != nil {
		fmt.Println("error found while find order detail: ", err)
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "查看订单详情错误",
		})
		return
	}
	fmt.Println("Find a supplier order")
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Suppliers order detail response",
		Data: order,
	})
}
