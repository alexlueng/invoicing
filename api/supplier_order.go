package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"jxc/service"
	"jxc/util"
	"net/http"
	"time"
)

// 返回的供应商列表数据
type responseSupOrders struct {
	models.SupplierOrder
	Products []models.SupplierSubOrder `json:"products"`
}

func AllSupplierOrders(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var orders []models.SupplierOrder
	var orderIds []int64
	var req models.SupplierOrderReq

	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "参数解释错误",
		})
		return
	}

	page, size := SetDefaultPageAndSize(req.Page, req.Size)

	// 设置排序主键
	orderFields := []string{"OrderSN", "price", "order_id"}
	option := SetPaginationAndOrder("order_id", orderFields, req.Ord, page, size)

	//设置搜索规则
	filter := models.GetSupplierOrderParam(req, claims.ComId)

	collection := models.Client.Collection("supplier_order")
	cur, err := collection.Find(context.TODO(), filter, option)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeSuccess,
			Msg:  "Get all supplier orders",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var result models.SupplierOrder
		err := cur.Decode(&result)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeSuccess,
				Msg:  "Get all supplier orders",
			})
			return
		}
		orders = append(orders, result)
		orderIds = append(orderIds, result.OrderId)
	}

	//查询的总数
	var total int64
	total, err = models.Client.Collection("supplier_order").CountDocuments(context.TODO(), filter)
	if err != nil {
		return
	}

	// 获取相关子订单,拼接成map[order_id][]models.SupplierSubOrder
	orderInstance := make(map[int64][]models.SupplierSubOrder)
	var resultOrderInstance models.SupplierSubOrder
	filter = bson.M{}
	filter["com_id"] = claims.ComId
	if len(orderIds) > 0 {
		filter["order_id"] = bson.M{"$in": orderIds}
		cur, err = models.Client.Collection("supplier_sub_order").Find(context.TODO(), filter)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "没有找到相应的子订单",
			})
			return
		}
		for cur.Next(context.TODO()) {
			err := cur.Decode(&resultOrderInstance)
			if err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: -1,
					Msg:  "Can't decode supplier sub order",
				})
				return
			}
			orderInstance[resultOrderInstance.OrderId] = append(orderInstance[resultOrderInstance.OrderId], resultOrderInstance)
		}
	}

	var data []responseSupOrders
	// 组装返回的数据
	for _, val := range orders {
		//orders[key].Products = orderInstance[val.OrderSN]
		data = append(data, responseSupOrders{
			SupplierOrder: val,
			Products:      orderInstance[val.OrderId],
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
	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)

	order := models.SupplierOrder{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Println("Get supplier_order data: ", string(data))
	err := json.Unmarshal(data, &order)
	if err != nil {
		fmt.Println("unmarshall error: ", err)
	}

	//这里需要一个订单号生成方法，日期加上6位数的编号,这个订单编号应该是全局唯一的
	order.OrderSN = GetTempOrderSN()
	order.ComID = claims.ComId

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
		return
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)

	// 更新供应商交易次数
	collection = models.Client.Collection("supplier")
	updateResult, err := collection.UpdateOne(context.TODO(), bson.D{{"supplier_id", order.SupplierID}, {"com_id", claims.ComId}}, bson.M{
		"$inc": bson.M{"transaction_num": 1}})
	if err != nil {
		fmt.Println("Can't update supplier transaction num: ", err)
		return
	}
	fmt.Println("update result: ", updateResult.UpsertedID)

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
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)

	order_sn := GetSupplierOrderSNService{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &order_sn)

	filter := bson.M{}
	filter["com_id"] = claims.ComId
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
	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)

	order := models.SupplierOrder{}
	order_sn := GetSupplierOrderSNService{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(data, &order_sn)
	if err != nil {
		fmt.Println("error found: ", err)
	}
	SmartPrint(order_sn)
	filter := bson.M{}
	filter["com_id"] = claims.ComId
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

type warehouses struct {
	ProductId    int64 `json:"product_id" form:"product_id"`
	Num          int64 `json:"num" form:"num"`
	WarehousesId int64 `json:"warehouse_id" form:"warehouse_id"`
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
	Warehouses []warehouses `json:"warehouse" form:"warehouse"`
	Supplier   supplier     `json:"supplier" form:"supplier"`
}

// 添加客户采购订单
// 发货方可以是仓库和供应商
// 收货方只能是客户
// 如果选择了仓库发货，可以是多个仓库同时发货
// 如果选择了供应商发货，则创建供应商订单
// 只能是一个供应商
func AddCustomerPurchaseOrder(c *gin.Context) {
	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var (
		req              reqData
		unit_price       float64 // 商品单价
		product_ids      []int64
		orderSn          string
		stockNum, amount int64
	)
	//user_id := int64(1) // TODO
	data, _ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(data, &req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	// 获取销售订单信息
	// 主要是收货信息
	order, err := service.FindOneCustomerSubOrder(req.SubOrderSn, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	product, err := service.FindOneProduct(order.ProductID, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "未能找到这个商品的信息！",
		})
		return
	}

	// 该商品的库存
	stockNum = product.Stock

	// 判断是供应商发货还是仓库发货，针对创建实例和订单
	switch req.Type {
	case 1:
		// 仓库发 库存实例
		var instance models.GoodsInstance
		var instanceArr []interface{}

		// 获取仓库id
		var warehouses_id = req.Warehouses[0].WarehousesId
		// 获取商品的库存
		productStock, err := service.GetProductInfoOfWarehouse(order.ProductID, claims.ComId, warehouses_id)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  err.Error(),
			})
			return
		}

		warehousesInfo, _ := service.FindWarehouse(warehouses_id, claims.ComId)
		var instanceId int64
		for _, val := range req.Warehouses {
			//仓库中是否有足够的库存
			data, ok := productStock[order.ProductID]
			if !ok {
				//仓库中没有这个商品的库存
				c.JSON(http.StatusOK, serializer.Response{
					Code: -1,
					Data: map[string]int64{"product": order.ProductID},
					Msg:  "仓库中没有这个商品的库存!",
				})
				return
			}
			wos, ok := data.Stock[val.WarehousesId]
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

			instanceId, _ = util.GetTableId("instance")
			// 组装库存实例数据
			instance = models.GoodsInstance{
				InstanceId:     instanceId,
				ComID:          claims.ComId,
				Type:           1,
				SrcType:        3, // 从仓库发
				SrcId:          val.WarehousesId,
				SrcTitle:       warehousesInfo[val.WarehousesId].Name,
				DestType:       1, // 接收方是客户
				DestId:         order.CustomerID,
				DestTitle:      order.CustomerName,
				DestOrderId:    order.OrderId,
				DestOrderSn:    order.OrderSN,
				DestSubOrderId: order.SubOrderId,
				DestSubOrderSn: order.SubOrderSn,
				PlaceType:      1, //1 销售-待发货
				PlaceId:        val.WarehousesId,
				SubPlaceId:     0,
				ProductID:      order.ProductID,
				Product:        order.Product,
				Contacts:       warehousesInfo[val.WarehousesId].WarehouseAdminName,
				Receiver:       order.Receiver,
				ReceiverPhone:  order.ReceiverPhone,
				CustomerPrice:  order.Price,
				SupplierPrice:  0,
				Amount:         val.Num,
				ExtraAmount:    0,
				DeliveryCom:    0,
				Delivery:       "",
				DeliveryCode:   "",
				OrderTime:      time.Now().Unix(),
				CreateBy:       claims.UserId,
				ShipTime:       0,
				ConfirmTime:    0,
				CheckTime:      0,
				PayTime:        0,
				FinishTime:     0,
				Status:         1,
				Units:          "",
				//SettlementOrderSN: "",
				//Settlement:        0,
			}
			instanceArr = append(instanceArr, instance)
			stockNum = stockNum - val.Num
			amount += val.Num
		}

		if (order.Amount - order.WarehouseAmount - order.SupplierAmount) < amount {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "购买总数不可大于需求量",
			})
			return
		}

		// 将这个订单实例插入订单实例表
		err = service.AddGoodsInstance(instanceArr)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  err.Error(),
			})
			return
		}

		// 修改子订单已发货数量
		err = service.UpdateSupplierAndWarehouseAmount(order.SubOrderSn, 1, amount, claims.ComId)
		if err != nil {
			fmt.Println("Can't update sub order: ", err)
			return
		}
		// 修改商品表中的库存
		err = service.UpdateProductStock(order.ProductID, stockNum, claims.ComId)
		if err != nil {
			fmt.Println("Can't update product stock: ", err)
			return
		}

		// 检测库存数量，如果低于库存的最低预警值，则向消息表中写入一条message，message中的User字段为仓库管理员
		var product models.Product
		collection := models.Client.Collection("product")
		err = collection.FindOne(context.TODO(), bson.D{{"product_id", order.ProductID}}).Decode(&product)
		if err != nil {
			fmt.Println("Can't find product")
			return
		}

		if product.Stock < 0 {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "商品库存小于0，请查看原因！",
			})
			return
		}

		// 检查系统是否开启了库存提醒 getSystemConfig()
		if product.Stock < product.MinAlert { // 生成n条提醒消息,消息的接收者是仓库管理员
			service.StockMessage(product, claims.ComId)
		}

		break
	case 2:
		// 供应商发

		// 获取供应商信息
		supplier, err := service.FindOneSupplier(req.Supplier.SupplierId, claims.ComId)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "未能找到该供应商！",
			})
			return
		}

		// 获取提交的商品进货价
		purchasePrice, err := service.FindProductPurchasePrice(product_ids, claims.ComId)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  err.Error(),
			})
			return
		}

		// 对采购价map进行验证，如有提交的供应商没有供货该商品的记录，则提示
		supplierPrice, ok := purchasePrice[order.ProductID].SupplierPrices[order.ProductID]
		if !ok {
			// 没有找到这个商品的价格
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Data: map[string]int64{"product_id": order.ProductID},
				Msg:  "没有找到这个商品的价格",
			})
			return
		}

		// 获取供应商商品价格
		unit_price = purchasePrice[order.ProductID].SupplierPrices[req.Supplier.SupplierId].SupplierPrice
		if unit_price < 0 {
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

		// 获取订单号
		orderSn, _ = util.GetOrderSN(claims.ComId)
		orderId, _ := util.GetTableId("supplier_order")

		// 组装采购订单
		supplierOrder = models.SupplierOrder{
			ComID:         claims.ComId,
			OrderId:       orderId,
			OrderSN:       orderSn,
			SalesOrderSn:  order.OrderSN,
			WarehouseID:   0,
			WarehouseName: "",
			SupplierID:    req.Supplier.SupplierId,
			Supplier:      supplier.SupplierName,
			Contacts:      order.Contacts,
			Receiver:      order.Receiver,
			ReceiverPhone: order.ReceiverPhone,
			Price:         supplierPrice.SupplierPrice * util.Unwrap(req.Supplier.Num, 0),
			Amount:        req.Supplier.Num,
			ExtraAmount:   0,
			Delivery:      "",
			DeliveryCode:  "",
			OrderTime:     time.Now().Unix(),
			CreateBy:      claims.UserId,
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
		subOrderSn, _ := util.GetOrderSN(claims.ComId)
		// 组装子订单数据
		supplierSubOrder = models.SupplierSubOrder{
			SubOrderId:       subOrderId, // 子订单id
			SubOrderSn:       subOrderSn, // 子订单号
			OrderId:          orderId,    // 订单id
			OrderSn:          orderSn,    // 订单号
			Units:            product.Units,
			ComID:            claims.ComId,
			ProductID:        order.ProductID,
			ProductNum:       req.Supplier.Num,
			ProductName:      product.Product,
			ProductUnitPrice: supplierPrice.SupplierPrice,
			CreateAt:         time.Now().Unix(),
			CreateBy:         claims.UserId,
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

		instanceId, _ := util.GetTableId("instance")
		// 组装库存实例数据
		instance := models.GoodsInstance{
			InstanceId:     instanceId,
			ComID:          claims.ComId,
			Type:           1,
			SrcType:        2, // 从供应商发
			SrcId:          req.Supplier.SupplierId,
			SrcTitle:       supplier.SupplierName,
			SrcOrderId:     orderId, //
			SrcOrderSn:     orderSn,
			SrcSubOrderId:  subOrderId,
			SrcSubOrderSn:  subOrderSn,
			DestType:       1, // 接收方是客户
			DestId:         order.CustomerID,
			DestTitle:      order.CustomerName,
			DestOrderId:    order.OrderId,
			DestOrderSn:    order.OrderSN,
			DestSubOrderId: order.SubOrderId,
			DestSubOrderSn: order.SubOrderSn,
			PlaceType:      0,
			PlaceId:        0,
			SubPlaceId:     0,
			ProductID:      order.ProductID,
			Product:        order.Product,
			Contacts:       supplier.Phone,
			Receiver:       order.Receiver,
			ReceiverPhone:  order.ReceiverPhone,
			CustomerPrice:  order.Price,
			SupplierPrice:  supplierPrice.SupplierPrice,
			Amount:         req.Supplier.Num,
			ExtraAmount:    0,
			DeliveryCom:    0,
			Delivery:       "",
			DeliveryCode:   "",
			OrderTime:      time.Now().Unix(),
			CreateBy:       claims.UserId,
			ShipTime:       0,
			ConfirmTime:    0,
			CheckTime:      0,
			PayTime:        0,
			FinishTime:     0,
			Status:         1,
			Units:          "",
			//SettlementOrderSN: "",
			//Settlement:        0,
		}
		if (order.Amount - order.WarehouseAmount - order.SupplierAmount) < req.Supplier.Num {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "购买总数不可大于需求量",
			})
			return
		}

		// 创建采购订单
		err = service.CreateSupplierOrder(supplierOrder, insertArr)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  err.Error(),
			})
			return
		}

		// 创建商品实例
		var instanceArr []interface{}
		instanceArr = append(instanceArr, instance)

		err = service.AddGoodsInstance(instanceArr)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  err.Error(),
			})
			return
		}

		//stockNum = stockNum - req.Supplier.Num

		// 修改子订单已发货数量
		if err = service.UpdateSupplierAndWarehouseAmount(order.SubOrderSn, 2, req.Supplier.Num+order.SupplierAmount, claims.ComId); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  err.Error(),
			})
			return
		}
		break
	default:
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "未定义的类型",
		})
		return
	}

	// TODO：更新总订单的状态为配货中
	// 如果已配货数量等于总数，则将状态改为配货完成

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

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var (
		req                 ReqAddPurchaseOrder
		amount              int64                   //商品总数量
		price, unit_price   float64                 // 订单总价,商品单价
		supplierSubOrder    models.SupplierSubOrder // 供应商订单实例
		supplierSubOrderArr []interface{}           //
		instance            models.GoodsInstance    // 商品实例
		productName         string                  //商品名
		instanceArr         []interface{}
		product_ids         []int64
		orderSn             string
	)

	data, _ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(data, &req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	// 获取提交的商品进货价
	for _, val := range req.Product {
		product_ids = append(product_ids, val.ProductId)
	}
	// 1
	purchasePrice, err := service.FindProductPurchasePrice(product_ids, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	// 获取仓库信息
	//2
	warehouses, err := service.FindOneWarehouse(req.WarehousesId, claims.ComId)
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
		warehouses.Product = append(warehouses.Product, val.ProductId)
	}
	// 去重
	warehouses.Product = util.RemoveRepeatedElementInt64(warehouses.Product)

	// 获取商品信息
	//3
	product, err := service.FindProduct(product_ids, claims.ComId)

	// 获取供应商信息
	//4
	Supplier, err := service.FindOneSupplier(req.SupplierID, claims.ComId)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	// 添加一条供应商订单，N条订单实例
	// 添加N条库存实例
	orderSn, err = util.GetOrderSN(claims.ComId)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	orderId, _ := util.GetTableId("supplier_sub_order")
	//5
	supplier, _ := service.FindOneSupplier(req.SupplierID, claims.ComId)

	// 组装供应商子订单,累加价格
	for _, val := range req.Product {
		unit_price = purchasePrice[val.ProductId].SupplierPrices[req.SupplierID].SupplierPrice
		productName = purchasePrice[val.ProductId].ProductName
		price += unit_price * util.Unwrap(val.Num, 0)
		amount += val.Num
		subOrderSn, _ := util.GetOrderSN(claims.ComId)
		subOrderId, _ := util.GetTableId("supplier_sub_order")

		supplierSubOrder = models.SupplierSubOrder{
			SubOrderId:       subOrderId,
			SubOrderSn:       subOrderSn,
			OrderId:          orderId,
			OrderSn:          orderSn,
			Units:            product[val.ProductId].Units,
			ComID:            claims.ComId,
			ProductID:        val.ProductId,
			ProductNum:       val.Num,
			ProductName:      productName,
			ProductUnitPrice: unit_price,
			CreateAt:         time.Now().Unix(),
			CreateBy:         claims.UserId,
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
		ComID:         claims.ComId,
		OrderSN:       orderSn,
		OrderId:       orderId,
		SalesOrderSn:  "",
		WarehouseID:   req.WarehousesId,
		WarehouseName: warehouses.Name,
		SupplierID:    req.SupplierID,
		Supplier:      supplier.SupplierName,
		Contacts:      Supplier.Contacts,
		Receiver:      warehouses.WarehouseAdminName,
		ReceiverPhone: warehouses.Phone,
		Price:         price,
		Amount:        amount,
		ExtraAmount:   0,
		Delivery:      "",
		DeliveryCode:  "",
		OrderTime:     time.Now().Unix(),
		CreateBy:      claims.UserId,
		ShipTime:      0,
		Shipper:       0,
		ConfirmTime:   0,
		ConfirmBy:     0,
		PayTime:       0,
		PayBy:         0,
		FinishTime:    0,
		Status:        1, // 未发货
	}
	// 组装实例
	for _, val := range supplierSubOrderArr {

		instanceId, _ := util.GetTableId("instance")

		// 从供应商发到仓库 src_type = 2 ,dest_type = 3
		data := val.(models.SupplierSubOrder)
		fmt.Print("data:", data)
		instance = models.GoodsInstance{
			InstanceId:     instanceId,
			ComID:          claims.ComId,
			Type:           2,
			SrcType:        2,
			SrcId:          req.SupplierID,
			SrcTitle:       supplier.SupplierName,
			SrcOrderId:     data.OrderId,
			SrcOrderSn:     data.OrderSn,
			SrcSubOrderId:  data.SubOrderId,
			SrcSubOrderSn:  data.SubOrderSn,
			DestType:       3,
			DestId:         req.WarehousesId, //接收方为仓库
			DestTitle:      warehouses.Name,
			DestOrderId:    0, // 接收方是仓库，所以没有接收方订单id
			DestOrderSn:    "",
			DestSubOrderId: 0,
			DestSubOrderSn: "",
			PlaceType:      4, // 采购-待收货
			PlaceId:        req.WarehousesId,
			SubPlaceId:     0,
			ProductID:      data.ProductID,
			Product:        data.ProductName,
			Contacts:       supplier.SupplierName,
			Units:          data.Units,
			Receiver:       warehouses.Name,
			ReceiverPhone:  warehouses.Phone,

			CustomerPrice: 0,
			SupplierPrice: data.ProductUnitPrice,
			//Price:             data.ProductUnitPrice,
			Amount:       data.ProductNum,
			ExtraAmount:  0,
			Delivery:     "",
			DeliveryCode: "",
			OrderTime:    time.Now().Unix(),
			CreateBy:     claims.UserId,
			ShipTime:     0,
			ConfirmTime:  0,
			PayTime:      0,
			FinishTime:   0,
			Status:       1,
			//SettlementOrderSN: "",
			//Settlement:        false,
		}
		instanceArr = append(instanceArr, instance)

		// 添加采购商品订单的时候不应该更新商品库存，未入库的商品不能算是库存
		//err = service.UpdateProductStock(data.ProductID, product[data.ProductID].Stock+data.ProductNum, claims.ComId)
		//	if err != nil {
		//		fmt.Println("Update product stock err: ", err)
		//}
	}

	// 添加供应商订单
	//6
	err = service.CreateSupplierOrder(supplierOrder, supplierSubOrderArr)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	// 添加库存实例记录
	//7
	err = service.AddGoodsInstance(instanceArr)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	// 更新仓库商品种类
	//8
	err = service.UpdateWarehouseProduct(warehouses.ID, claims.ComId, warehouses.Product)
	if err != nil {
		fmt.Println("Update Warehouse err: ", err)
		return
	}

	//// 更新商品库存数量
	//err = service.UpdateProductStock(product.ProductID, (product.Stock + req.Num), claims.ComId)
	//if err != nil {
	//	fmt.Println("Can't update product stock: ", err)
	//	return
	//}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: map[string]string{"order_sn": orderSn,},
		Msg:  "创建采购订单成功",
	})

}

func SupplierMessage() {
	// TODO：生成一条供应商订单消息，插入到消息表中，通过微信公众号发送给供应商
	/*	var message models.MessageForClient
		message.ID = GetLastID("client_message")
		message.ComID = claims.ComId
		message.Client = 2 // 2代表的是供应商
		message.ClientID = supplier.ID
		message.Telephone = supplier.Phone
		message.CreateAt = time.Now().Unix()
		message.Title = "新的采购订单"
		message.Content = "向" + supplier.SupplierName + "下了一个订单"
		message.IsRead = false

		collection := models.Client.Collection("client_message")
		insertResult, err := collection.InsertOne(context.TODO(), message)
		if err != nil {
			fmt.Println("Can't insert client message")
			return
		}
		fmt.Println("insert result: ", insertResult.InsertedID)*/
}

// 拆分子订单
// 一批货物可能是分批次到达的，所以当这一单的部分货物到货的时候，
// 需要将这个订单进行拆分；
// 需要客户端传过来的参数：订单实例id 拆分数量
// 还需要更新这个订单的状态

type SplitOrderService struct {
	SubOrderID int64 `json:"order_sub_id"` // 子订单号
	Num        int64 `json:"num"`          // 已经接收货物的数量
}

// 采购子订单部分收货 拆单
func SplitSupplierSubOrder(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var sos SplitOrderService
	if err := c.ShouldBindJSON(&sos); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	// 原始订单
	oldSubOrder := models.SupplierSubOrder{}
	collection := models.Client.Collection("supplier_sub_order")
	err := collection.FindOne(context.TODO(), bson.D{{"com_id", claims.ComId},
		{"order_sub_id", sos.SubOrderID}}).Decode(&oldSubOrder)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "找不到原子订单",
		})
		return
	}

	// TODO：需要增加一条新的订单实例？

	// 如果前端传过来的值大于原子订单的值，则报错，提示返回
	if sos.Num > oldSubOrder.ProductNum {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "接收数量不能大于总数",
		})
		return
	}

	// 先创建新的订单后更新原订单
	newSubOrder := oldSubOrder
	newSubOrder.CreateAt = time.Now().Unix()
	newSubOrder.ProductNum = sos.Num
	newSubOrder.SubOrderId = GetLastID("supplier_sub_order")
	subSn_str, _ := util.GetOrderSN(newSubOrder.ComID)
	newSubOrder.SubOrderSn = subSn_str
	newSubOrder.State = 2 // 拆出来的单的状态是已收货
	newSubOrder.ActualAmount = sos.Num
	_, err = collection.InsertOne(context.TODO(), newSubOrder)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "接收数量不能大于总数",
		})
		return
	}
	SetLastID("supplier_sub_order")

	// 更新原订单未接收的数量
	_, err = collection.UpdateOne(context.TODO(), bson.D{{"com_id", claims.ComId},
		{"order_sub_id", sos.SubOrderID}},
		bson.M{"$set": bson.M{"product_num": oldSubOrder.ProductNum - sos.Num,
			"actual_amount": oldSubOrder.ActualAmount - sos.Num}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't update old order",
		})
		return
	}

	// 修改总订单状态，根据子订单的状态来确定是全部收货还是部分收货
	collection = models.Client.Collection("supplier_order")
	var supplierOrder models.SupplierOrder
	ordFilter := bson.M{}
	ordFilter["com_id"] = claims.ComId
	ordFilter["order_id"] = oldSubOrder.OrderId
	err = collection.FindOne(context.TODO(), ordFilter).Decode(&supplierOrder)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}
	// 找出这个订单的所有子订单
	subOrderCollection := models.Client.Collection("supplier_sub_order")
	var subOrders []models.SupplierSubOrder
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["order_id"] = supplierOrder.OrderId
	cur, err := subOrderCollection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}
	for cur.Next(context.TODO()) {
		var res models.SupplierSubOrder
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  err.Error(),
			})
			return
		}
		subOrders = append(subOrders, res)
	}
	var allRecieve = false // 判断总订单是部分收货还是全部收货的标志位
	for _, subOrder := range subOrders {
		if subOrder.State == 3 {
			allRecieve = true
			break
		}
	}
	if allRecieve {
		_, err := collection.UpdateOne(context.TODO(), ordFilter, bson.M{"$set": bson.M{"status": 4}})
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  err.Error(),
			})
			return
		}
	} else {
		_, err := collection.UpdateOne(context.TODO(), ordFilter, bson.M{"$set": bson.M{"status": 2}})
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  err.Error(),
			})
			return
		}
	}

	// 原子订单号相同的订单都要返回
	var respData []models.SupplierSubOrder
	collection = models.Client.Collection("supplier_sub_order")
	cur, err = collection.Find(context.TODO(), bson.D{{"order_id", oldSubOrder.OrderId}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find all sub order",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var res models.SupplierSubOrder
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't decode sub order",
			})
			return
		}
		respData = append(respData, res)
	}

	// 将拆分出来的订单返回给前端
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Customer order detail response",
		Data: map[string]interface{}{
			"sub_orders": respData,
			"order":      supplierOrder,
		},
	})
}
