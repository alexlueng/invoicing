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
	"jxc/service"
	"jxc/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 返回的供应商列表数据
type responseSupOrders struct {
	models.SupplierOrder
	Products []models.SupplierSubOrder `json:"products"`
}

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
	//var orders []responseSupOrders
	var orderSns []string
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
	//StartPayTime time.Time `json:"start_pay_time" form:"start_pay_time"`
	//EndPayTime time.Time `json:"end_pay_time" form:"end_pay_time"`
	if req.StartPayTime != "" {
		time, _ := strconv.Atoi(req.StartPayTime)
		startPayTime := int64(time)
		if req.EndPayTime != "" {
			time, _ := strconv.Atoi(req.StartPayTime)
			endOrderTime := int64(time)
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
	//StartShipTime time.Time `json:"start_ship_time" form:"start_ship_time"`
	//EndShipTime time.Time `json:"end_ship_time" form:"end_ship_time"`
	if req.StartShipTime != "" {
		time, _ := strconv.Atoi(req.StartShipTime)
		startShipTime := int64(time)
		if req.EndPayTime != "" {
			time, _ := strconv.Atoi(req.StartShipTime)
			endShipTime := int64(time)
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
		//var result responseSupOrders
		err := cur.Decode(&result)
		if err != nil {
			fmt.Println("error found decoding supplier order: ", err)
			return
		}
		orders = append(orders, result)
		orderSns = append(orderSns, result.OrderSN)
	}

	//查询的总数
	var total int64
	total, err = models.Client.Collection("supplier_order").CountDocuments(context.TODO(), filter)
	if err != nil {
		return
	}

	// 获取相关子订单,拼接成map[order_sn][]models.SupplierSubOrder
	orderInstance := make(map[string][]models.SupplierSubOrder)
	var resultOrderInstance models.SupplierSubOrder
	filter = bson.M{}
	filter["com_id"] = com.ComId
	filter["order_sn"] = bson.M{"$in": orderSns}
	cur, err = models.Client.Collection("supplier_sub_order").Find(context.TODO(), filter)
	if err != nil {
		return
	}
	for cur.Next(context.TODO()) {
		err := cur.Decode(&resultOrderInstance)
		if err != nil {
			return
		}
		orderInstance[resultOrderInstance.OrderSn] = append(orderInstance[resultOrderInstance.OrderSn], resultOrderInstance)
	}

	var data []responseSupOrders
	// 组装返回的数据
	for _, val := range orders {
		//orders[key].Products = orderInstance[val.OrderSN]
		data = append(data, responseSupOrders{
			SupplierOrder: val,
			Products:      orderInstance[val.OrderSN],
		})
	}

	// 返回查询到的总数，总页数
	resData := models.ResponseSupplierOrdersData{}
	resData.SupplierOrders = data
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

// 添加仓库采购订单
// 发货方是客户
// 收货方是仓库
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
	/*
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
				"contacts":       updateSupplierOrder.Contacts,
				"receiver_phone": updateSupplierOrder.Phone,
				"amount":         updateSupplierOrder.Amount,
				"Delivery":       updateSupplierOrder.Delivery,
				"warehouse_id":   updateSupplierOrder.WarehouseID,
				"receiver":       updateSupplierOrder.Receiver,
				"price":          updateSupplierOrder.Price,
				"extra_amount":   updateSupplierOrder.ExtraAmount,
				"delivery_code":  updateSupplierOrder.DeliveryCode,}})
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
	*/
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
	filter["order_sn"] = order_sn.OrderSN

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

/*
* 在供应商订单中，
* 发货方可以是供应商、也可以是仓库
* 收货方可以是仓库，也可以是客户
* 针对场景进行划分：
* 1.拆分销售订单（客户订单）时，提交的供应商订单为客户采购订单
* 发货方是仓库或供应商、收货只能是客户
* 一个订单包含一个商品
* 2. 仓库添加采购库存时，提交的供应商订单为仓库采购订单
* 发货方是供应商、收货方是仓库
* 一个订单包含多个商品
 */

// 添加订单提交的参数
type Product struct {
	ProductId    int64 `json:"product_id"`
	Num          int64 `json:"num"`
	WarehousesId int64 `json:"warehouses_id" form:"warehouses_id"`
	SupplierID   int64 `json:"supplier_id" form:"supplier_id"` // 供应商id
}

type ReqCustomerPurchaseOrder struct {
	Product     []Product `json:"product" form:"product"`           // 提交的商品数据
	Type        int64     `json:"type" form:"type"`                 // 1:仓库发,2:供应商发
	OrderSn     string    `json:"order_sn" form:"order_sn"`         // type=2，订单号
	ExtraAmount float64   `json:"extra_amount" form:"extra_amount"` //本单优惠或折扣金额
	SupplierID  int64     `json:"supplier_id" form:"supplier_id"`   // 供应商id
}

// 添加采购订单数据格式
type ReqAddPurchaseOrder struct {
	SupplierID   int64     `json:"supplier_id" form:"supplier_id"` // 供应商id
	WarehousesId int64     `json:"warehouse_id" form:"warehouse_id"`
	Product      []Product `json:"product" form:"product"` // 提交的商品数据，公用相同的数据格式
}

// 添加客户采购订单
// 发货方可以是仓库和供应商
// 收货方只能是客户
// 如果选择了仓库发货，可以是多个仓库同时发货
// 如果选择了供应商发货，则创建供应商订单
// 只能是一个供应商
func AddCustomerPurchaseOrder(c *gin.Context) {
	// 根据域名获取comid
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	type warehouses struct {
		ProductId    int64 `json:"product_id" form:"product_id"`
		Num          int64 `json:"num" form:"num"`
		WarehousesId int64 `json:"warehouses_id" form:"warehouses_id"`
		SupplierID   int64 `json:"supplier_id" form:"supplier_id"` // 供应商id
	}
	type supplier struct {
		SupplierId int64 `json:"supplier_id" form:"supplier_id"`
		Num        int64 `json:"num" form:"num"`
	}
	type reqData struct {
		SubOrderId int64        `json:"sub_order_id" form:"sub_order_id"`
		SubOrderSn string       `json:"sub_order_sn" form:"sub_order_sn"`
		Type       int64        `json:"type" form:"type"`
		Warehouses []warehouses `json:"warehouses" form:"supplier"`
		Supplier   supplier     `json:"supplier" form:"supplier"`
	}

	var req reqData
	var unit_price float64 // 商品单价
	var product_ids []int64
	var orderSn string
	user_id := int64(1) // TODO
	data, _ := ioutil.ReadAll(c.Request.Body)
	err = json.Unmarshal(data, &req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	// 获取销售订单信息
	// 主要是收货信息
	order, err := service.FindOneCustomerSubOrder(req.SubOrderSn, com.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	// 获取提交的商品进货价
	purchasePrice, err := service.FindProductPurchasePrice(product_ids, com.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	// 判断是供应商发货还是仓库发货，针对创建实例和订单
	switch req.Type {
	case 1:
		// 仓库发
		// 库存实例
		var instance models.GoodsInstance
		var instanceArr []interface{}
		// 获取商品的库存  TODO 暂不判断库存
		/*productWos, err := service.GetProductWos(product_ids, com.ComId, 0)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  err.Error(),
			})
			return
		}
		// 每个仓库是否有足够的库存
		for _, val := range req.Product {
			data, ok := productWos[val.ProductId]
			if !ok {
				//仓库中没有这个商品的库存
				c.JSON(http.StatusOK, serializer.Response{
					Code: -1,
					Data: map[string]int64{"product": val.ProductId},
					Msg:  "仓库中没有这个商品的库存!",
				})
				return
			}
			wos, ok := data.Wos[val.WarehousesId]
			if !ok {
				// 这个仓库中没有这个商品的库存
				c.JSON(http.StatusOK, serializer.Response{
					Code: -1,
					Data: map[string]int64{"warehouses": val.WarehousesId, "product": val.ProductId},
					Msg:  "这个仓库中没有这个商品的库存!",
				})
				return
			}
			if wos.Num < val.Num {
				// 这个仓库中的数量不足
				c.JSON(http.StatusOK, serializer.Response{
					Code: -1,
					Data: map[string]int64{"warehouses": val.WarehousesId, "product": val.ProductId, "wos_num": wos.Num, "num": val.Num},
					Msg:  "这个仓库中的库存数量不足!",
				})
				return
			}
		}*/
		// 获取仓库id
		var warehouses_ids []int64
		for _, val := range req.Warehouses {
			warehouses_ids = append(warehouses_ids, val.WarehousesId)
		}
		warehousesInfo, _ := service.FindWarehouse(warehouses_ids, com.ComId)
		var instanceId int64
		for _, val := range req.Warehouses {
			instanceId, _ = util.GetTableId("instance")
			// 组装库存实例数据
			instance = models.GoodsInstance{
				InstanceId:        instanceId,
				ComID:             com.ComId,
				Type:              1,
				SrcType:           3, // 从仓库发
				SrcId:             val.WarehousesId,
				SrcTitle:          warehousesInfo[val.WarehousesId].WarehouseAdminName,
				DestType:          1, // 接收方是客户
				DestId:            order.CustomerID,
				DestTitle:         order.CustomerName,
				DestOrderId:       order.OrderId,
				DestOrderSn:       order.OrderSN,
				DestSubOrderId:    order.SubOrderId,
				DestSubOrderSn:    order.SubOrderSn,
				PlaceType:         1, //1 销售-待发货
				PlaceId:           val.WarehousesId,
				SubPlaceId:        0,
				ProductID:         order.ProductID,
				Product:           order.Product,
				Contacts:          warehousesInfo[val.WarehousesId].WarehouseAdminName,
				Receiver:          order.Receiver,
				ReceiverPhone:     order.ReceiverPhone,
				CustomerPrice:     order.Price,
				SupplierPrice:     0,
				Amount:            order.Amount,
				ExtraAmount:       0,
				DeliveryCom:       0,
				Delivery:          "",
				DeliveryCode:      "",
				OrderTime:         time.Now().Unix(),
				CreateBy:          0,
				ShipTime:          0,
				ConfirmTime:       0,
				CheckTime:         0,
				PayTime:           0,
				FinishTime:        0,
				Status:            1,
				Units:             "",
				SettlementOrderSN: "",
				Settlement:        0,
			}
			instanceArr = append(instanceArr, instance)
		}
		err = service.AddGoodsInstance(instanceArr)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  err.Error(),
			})
		}

		break
	case 2:
		// 供应商发
		// 对采购价map进行验证，如有提交的供应商没有供货该商品的记录，则提示

		data, ok := purchasePrice[order.ProductID]
		if !ok {
			// 没有找到这个商品的价格
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Data: map[string]int64{"product_id": order.ProductID},
				Msg:  "没有找到这个商品的价格",
			})
			return
		}
		_, ok = data.SupplierPrices[req.Supplier.SupplierId]
		if !ok {
			// 这个供应商没有提供这个商品
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Data: map[string]int64{"supplier_id": req.Supplier.SupplierId, "product_id": order.ProductID},
				Msg:  "这个供应商没有提供这个商品",
			})
			return
		}

		// 组装供应商订单数据，组装子订单数据，组装实例数据

		var supplierOrder models.SupplierOrder
		var supplierSubOrder models.SupplierSubOrder

		// 获取供应商商品价格
		unit_price = purchasePrice[order.ProductID].SupplierPrices[req.Supplier.SupplierId].SupplierPrice

		// 获取订单号
		orderSn, _ = util.GetOrderSN(com.ComId)
		orderId, _ := util.GetTableId("supplier_order")

		supplierOrder = models.SupplierOrder{
			ComID:         com.ComId,
			OrderId:       orderId,
			OrderSN:       orderSn,
			SalesOrderSn:  order.OrderSN,
			WarehouseID:   0,
			WarehouseName: "",
			SupplierID:    req.Supplier.SupplierId,
			Contacts:      order.Contacts,
			Receiver:      order.Receiver,
			ReceiverPhone: order.ReceiverPhone,
			Price:         unit_price * util.Unwrap(req.Supplier.Num, 0),
			Amount:        req.Supplier.Num,
			ExtraAmount:   0,
			Delivery:      "",
			DeliveryCode:  "",
			OrderTime:     time.Now().Unix(),
			CreateBy:      user_id,
			ShipTime:      0,
			Shipper:       0,
			ConfirmTime:   0,
			ConfirmBy:     0,
			PayTime:       0,
			PayBy:         0,
			FinishTime:    0,
			Status:        1,
		}
		subOrderId, _ := util.GetTableId("sub_order")
		subOrderSn, _ := util.GetOrderSN(com.ComId)
		// 组装子订单数据
		supplierSubOrder = models.SupplierSubOrder{
			SubOrderId:  subOrderId, // 子订单id
			SubOrderSn:  subOrderSn, // 子订单号
			OrderId:     orderId,    // 订单id
			OrderSn:     orderSn,    // 订单号
			ProductName: "ProductName",
			Units:       "",

			ComID: com.ComId,

			ProductID:        order.ProductID,
			ProductNum:       req.Supplier.Num,
			ProductUnitPrice: unit_price,
			CreateAt:         user_id,
			CreateBy:         time.Now().Unix(),
			ShipTime:         0,
			Shipper:          0,
			ConfirmAt:        0,
			ConfirmBy:        0,
			CheckAt:          0,
			CheckBy:          0,
			State:            1,
		}
		var insertArr []interface{}
		insertArr = append(insertArr, supplierSubOrder)

		err = service.CreateSupplierOrder(supplierOrder, insertArr)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  err.Error(),
			})
			return
		}
		instanceId, _ := util.GetTableId("instance")
		// 组装库存实例数据
		instance := models.GoodsInstance{
			InstanceId:        instanceId,
			ComID:             com.ComId,
			Type:              1,
			SrcType:           2, // 从仓库发
			SrcId:             req.Supplier.SupplierId,
			SrcTitle:          "",
			SrcOrderId:        0, // 从仓库发，所以来源订单id为空
			SrcOrderSn:        "",
			SrcSubOrderId:     0,
			SrcSubOrderSn:     "",
			DestType:          1, // 接收方是客户
			DestId:            order.CustomerID,
			DestTitle:         order.CustomerName,
			DestOrderId:       order.OrderId,
			DestOrderSn:       order.OrderSN,
			DestSubOrderId:    order.SubOrderId,
			DestSubOrderSn:    order.SubOrderSn,
			PlaceType:         0,
			PlaceId:           0,
			SubPlaceId:        0,
			ProductID:         order.ProductID,
			Product:           order.Product,
			Contacts:          "",
			Receiver:          order.Receiver,
			ReceiverPhone:     order.ReceiverPhone,
			CustomerPrice:     order.Price,
			SupplierPrice:     0,
			Amount:            order.Amount,
			ExtraAmount:       0,
			DeliveryCom:       0,
			Delivery:          "",
			DeliveryCode:      "",
			OrderTime:         time.Now().Unix(),
			CreateBy:          0,
			ShipTime:          0,
			ConfirmTime:       0,
			CheckTime:         0,
			PayTime:           0,
			FinishTime:        0,
			Status:            1,
			Units:             "",
			SettlementOrderSN: "",
			Settlement:        0,
		}
		var instanceArr []interface{}
		instanceArr = append(instanceArr, instance)

		err = service.AddGoodsInstance(instanceArr)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  err.Error(),
			})
		}

		break
	default:
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "未定义的类型",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "添加供应商订单成功！",
	})
	return

}

// 添加仓库采购订单
// 发货方为供应商
// 收货方为仓库
func AddPurchaseOrder(c *gin.Context) {
	// 根据域名获取comid
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	var req ReqAddPurchaseOrder
	var amount int64                             //商品总数量
	var price, unit_price float64                // 订单总价,商品单价
	var supplierSubOrder models.SupplierSubOrder // 供应商订单实例
	var supplierSubOrderArr []interface{}        //
	var instance models.GoodsInstance            // 商品实例
	var productName string                       //商品名
	var instanceArr []interface{}

	user_id := int64(1) // TODO
	data, _ := ioutil.ReadAll(c.Request.Body)
	err = json.Unmarshal(data, &req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	var product_ids []int64
	var orderSn string

	// 获取提交的商品进货价
	for _, val := range req.Product {
		product_ids = append(product_ids, val.ProductId)
	}
	purchasePrice, err := service.FindProductPurchasePrice(product_ids, com.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	// 对采购价map进行验证，如有提交的供应商没有供货该商品的记录，则提示
	for _, val := range req.Product {
		data, ok := purchasePrice[val.ProductId]
		if !ok {
			// 没有找到这个商品的价格
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Data: map[string]int64{"product_id": val.ProductId},
				Msg:  "没有找到这个商品的价格",
			})
			return
		}
		_, ok = data.SupplierPrices[req.SupplierID]
		if !ok {
			// 这个供应商没有提供这个商品
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Data: map[string]int64{"supplier_id": val.SupplierID, "product_id": val.ProductId},
				Msg:  "这个供应商没有提供这个商品",
			})
			return
		}
	}

	// 获取供应商信息
	Supplier, err := service.FindOneSupplier(req.SupplierID, com.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	// 获取仓库信息
	warehouses, err := service.FindOneWarehouse(req.WarehousesId, com.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	// 添加一条供应商订单，N条订单实例
	// 添加N条库存实例

	orderSn, _ = util.GetOrderSN(com.ComId)
	orderId, _ := util.GetTableId("supplier_sub_order")
	supplier, _ := service.FindOneSupplier(req.SupplierID, com.ComId)

	// 组装供应商子订单,累加价格
	for _, val := range req.Product {
		unit_price = purchasePrice[val.ProductId].SupplierPrices[req.SupplierID].SupplierPrice
		productName = purchasePrice[val.ProductId].ProductName
		price += unit_price * util.Unwrap(val.Num, 0)
		amount += val.Num
		subOrderSn, _ := util.GetOrderSN(com.ComId)
		subOrderId, _ := util.GetTableId("supplier_sub_order")

		supplierSubOrder = models.SupplierSubOrder{
			SubOrderId:       subOrderId,
			SubOrderSn:       subOrderSn,
			OrderId:          orderId,
			OrderSn:          "",
			Units:            "",
			ComID:            com.ComId,
			ProductID:        val.ProductId,
			ProductNum:       val.Num,
			ProductName:      productName,
			ProductUnitPrice: unit_price,
			CreateAt:         user_id,
			CreateBy:         time.Now().Unix(),
			ShipTime:         0,
			Shipper:          0,
			ConfirmAt:        0,
			ConfirmBy:        0,
			CheckAt:          0,
			CheckBy:          0,
			State:            1,
		}

		supplierSubOrderArr = append(supplierSubOrderArr, supplierSubOrder)
	}

	// 组装供应商订单
	supplierOrder := models.SupplierOrder{
		ComID:         com.ComId,
		OrderSN:       orderSn,
		OrderId:       orderId,
		SalesOrderSn:  "",
		WarehouseID:   req.WarehousesId,
		WarehouseName: warehouses.Name,
		SupplierID:    req.SupplierID,
		Contacts:      Supplier.Contacts,
		Receiver:      warehouses.WarehouseAdminName,
		ReceiverPhone: warehouses.Phone,
		Price:         price,
		Amount:        amount,
		ExtraAmount:   0,
		Delivery:      "",
		DeliveryCode:  "",
		OrderTime:     time.Now().Unix(),
		CreateBy:      user_id,
		ShipTime:      0,
		Shipper:       0,
		ConfirmTime:   0,
		ConfirmBy:     0,
		PayTime:       0,
		PayBy:         0,
		FinishTime:    0,
		Status:        1,
	}
	//
	instanceId, _ := util.GetTableId("instance")
	// 组装实例

	for _, val := range supplierSubOrderArr {

		// 从供应商发到仓库 src_type = 2 ,dest_type = 3
		data := val.(models.SupplierSubOrder)
		fmt.Print("data:", data)
		instance = models.GoodsInstance{
			InstanceId:    instanceId,
			ComID:         com.ComId,
			Type:          2,
			SrcType:       2,
			SrcId:         req.SupplierID,
			SrcTitle:      supplier.SupplierName,
			SrcOrderId:   data.OrderId,
			SrcOrderSn:    data.OrderSn,
			SrcSubOrderId: data.SubOrderId,
			SrcSubOrderSn: data.SubOrderSn,
			DestType:      3,
			DestId:        req.WarehousesId, //接收方为仓库
			DestTitle:     warehouses.Name,
			DestOrderId:0, // 接收方是仓库，所以没有接收方订单id
			DestOrderSn:"",
			DestSubOrderId:0,
			DestSubOrderSn:"",
			PlaceType:     4, // 采购-待收货
			PlaceId:       req.WarehousesId,
			SubPlaceId:    0,
			ProductID:     data.ProductID,
			Product:       data.ProductName,
			Contacts:      supplier.SupplierName,
			Receiver:      warehouses.Name,
			ReceiverPhone: warehouses.Phone,
			//Price:             data.ProductUnitPrice,
			Amount:            data.ProductNum,
			ExtraAmount:       0,
			Delivery:          "",
			DeliveryCode:      "",
			OrderTime:         time.Now().Unix(),
			CreateBy:          user_id,
			ShipTime:          0,
			ConfirmTime:       0,
			PayTime:           0,
			FinishTime:        0,
			Status:            1,
			SettlementOrderSN: "",
			//Settlement:        false,
		}
		instanceArr = append(instanceArr, instance)
	}

	// 添加供应商订单
	err = service.CreateSupplierOrder(supplierOrder, supplierSubOrderArr)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	// 添加库存实例记录
	err = service.AddGoodsInstance(instanceArr)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: map[string]string{"order_sn": orderSn,},
		Msg:  "创建采购订单成功",
	})

}

//
//// 采购订单实例发货
//func PurchaseOrderShipped(c *gin.Context) {
//	// 传入参数
//	// 订单号
//	// 商品id
//	// 配送方式
//	// 快递单号
//
//	// 获取实例信息
//
//}
//
//// 采购订单实例确认
//
//// 传入参数
//// 订单号
//// 商品id
//
//// 采购订单审核
//
//// 传入参数
//// 订单号
//// 商品id

func CheckSupplierPrice(c *gin.Context) {
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

}
