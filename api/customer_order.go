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
	"jxc/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

	var orders []models.CustomerOrder
	var req models.CustomerOrderReq

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

	option.Projection = bson.M{"products": 0}
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
		stime, _ := strconv.Atoi(req.StartOrderTime)
		startOrderTime := int64(stime)
		if req.EndOrderTime != "" {
			stime, _ := strconv.Atoi(req.StartOrderTime)
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
	//StartPayTime time.Time `json:"start_pay_time" form:"start_pay_time"`
	//EndPayTime time.Time `json:"end_pay_time" form:"end_pay_time"`
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
	//StartShipTime time.Time `json:"start_ship_time" form:"start_ship_time"`
	//EndShipTime time.Time `json:"end_ship_time" form:"end_ship_time"`
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

	//filter["com_id"] = com.ComId
	filter["com_id"] = 1
	fmt.Println("filter: ", filter)

	collection := models.Client.Collection("customer_order")
	sub_collection := models.Client.Collection("customer_suborder")

	cur, err := collection.Find(context.TODO(), filter, option)
	if err != nil {
		fmt.Println("error found finding customer orders: ", err)
		return
	}
	for cur.Next(context.TODO()) {
		var result models.CustomerOrder
		//var products []models.CustomerOrderProductsInfo
		err := cur.Decode(&result)
		if err != nil {
			fmt.Println("error found decoding customer order: ", err)
			return
		}
		f := bson.M{}
		f["com_id"] = 1
		f["order_sn"] = result.OrderSN
		sub_cur, err := sub_collection.Find(context.TODO(), f)
		for sub_cur.Next(context.TODO()) {
			var res models.CustomerOrderProductsInfo
			if err := sub_cur.Decode(&res); err != nil {
				fmt.Println("error found decoding product info: ", err)
				return
			}
			result.Products = append(result.Products, res)
		}
		orders = append(orders, result)
	}

	//查询的总数
	var total int64
	cur, _ = models.Client.Collection("customer_order").Find(context.TODO(),filter)
	for cur.Next(context.TODO()) {
		total++
	}


	// 返回查询到的总数，总页数
	resData := models.ResponseCustomerOrdersData{}
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

func AddCustomerOrder(c *gin.Context) {

	// 生成订单的流程：
	// 1. 获取选择的商品列表，数量
	// 2. 获取选择的客户
	// 3. 获取商品的单价
	// 4. 检查包邮字段是否有值，如果有，则获取相应的值
	// 5. 订单状态设置为待确认，下单时间设置为当前时间，
	// 6. 算出总价
	// 7. 组合好前端需要的数据并返回

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

	order := models.CustomerOrder{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Println("Get customer_order data: ", string(data))
	err = json.Unmarshal(data, &order)
	if err != nil {
		fmt.Println("unmarshall error: ", err)
	}

	// 计算订单总价并与前端传过来的值做对比，如果不相等，则下单失败
	var checkPrice float64
	var totalAmount int64
	for _, product := range order.Products {
		checkPrice += product.Price * float64(product.Quantity)
		totalAmount +=product.Quantity
	}
	checkPrice += order.TransportationExpense
	fmt.Println("Counting price: ", checkPrice)
	if checkPrice != order.TotalPrice {
		fmt.Println("the price that posted does not match.")
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "订单价格错误",
		})
		return
	}
	order.Amount = totalAmount
	fmt.Printf("count amount: %d, order amount: %d\n", totalAmount, order.Amount)

	//这里需要一个订单号生成方法，日期加上6位数的编号,这个订单编号应该是全局唯一的
	order.OrderSN = GetTempOrderSN()
	order.OrderId = getLastID("customer_order")
	//order.ComID = com.ComId

	order.ComID = 1

	// 创建订单的时间，以int64的类型插入到mongodb
	// TODO: 把这个方法独立出来
	current_time := time.Now()
	order.OrderTime = current_time.Unix()
	fmt.Println("order_time: ", order.OrderTime)

	//设置订单状态
	order.Status = models.TOBECONFIRMED

	SmartPrint(order)


	collection := models.Client.Collection("customer_order")
	insertResult, err := collection.InsertOne(context.TODO(), order)
	if err != nil {
		fmt.Println("Error while inserting mongo: ", err)
		return
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)

	collection = models.Client.Collection("customer_suborder")
	// 把订单中的每个子项插入到客户订单实例表中
	for _, item := range order.Products {
		var result models.CustomerSubOrder
		//result.ComID = com.ComId
		result.ComID = 1

		subSn_str, _ := util.GetOrderSN(result.ComID)
		//fmt.Println("get subID_str: ", subSn_str)
		result.SubOrderId =getLastID("sub_order")
		result.SubOrderSn = subSn_str
		//fmt.Println("get subID: ", subId)
		result.OrderId = order.OrderId

		result.CustomerID = order.CustomerID
		result.OrderSN = order.OrderSN
		result.Product = item.Product
		result.Amount = item.Quantity
		result.Price = item.Price
		result.ProductID = item.ProductID
		result.Receiver = order.Receiver
		result.ReceiverPhone = order.Phone
		result.OrderTime = order.OrderTime
		result.Status = order.Status

		_, err = collection.InsertOne(context.TODO(), result)
		if err != nil {
			fmt.Println("Error while inserting mongo: ", err)
			return
		}
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Customer order create succeeded",
	})
}


func CheckCustomerPrice(c *gin.Context) {
	data, _ :=ioutil.ReadAll(c.Request.Body)
	var orderProducts models.OrderProducts
	err := json.Unmarshal(data, &orderProducts)
	if err != nil {
		fmt.Println("error while unmarshaling: ", err)
		return
	}
	// 从客户商品价格表中找到对应商品的客户价格
	// 如果没有找到这条记录，则从商品表中找到该商品默认价格，并将这条记录插入到客户商品价格表，并将它的价格设置为商品的默认价格

	var price []float64 // 需要返回给前端的价格数组
	filter := bson.M{}
	filter["com_id"] = 1 // need to get com id from middleware
	filter["customer_id"] = orderProducts.CustomerID

	collection := models.Client.Collection("customer_product_price")
	for _, product_id := range orderProducts.ProductsID {
		filter["product_id"] = product_id
		var product models.CustomerProductPrice
		err := collection.FindOne(context.TODO(), filter).Decode(&product)
		if err != nil {
			// 说明客户商品价格表中没有这条记录,需要从商品表中找到默认价格
			var p models.Product
			c := models.Client.Collection("product")
			err := c.FindOne(context.TODO(), bson.D{{"product_id", product_id}}).Decode(&p)
			if err != nil {
				fmt.Println("Can't find default product price: ", err)
			}

			price = append(price, p.DefaultPrice)

			var cpp models.CustomerProductPrice
			cpp.ComID = 1
			cpp.ProductID = product_id
			cpp.Product = p.Product
			cpp.CustomerID = orderProducts.CustomerID
			cpp.CustomerName = orderProducts.CustomerName
			cpp.Price = p.DefaultPrice
			cpp.IsValid = true
			cpp.CreateAt = time.Now().Unix()

			cppCollection := models.Client.Collection("customer_product_price")
			_, err = cppCollection.InsertOne(context.TODO(), cpp)
			if err != nil {
				fmt.Println("err while insert result: ", err)
				return
			}

			// 更新商品客户列表，把客户id追加到cus_price数组中
			collection = models.Client.Collection("product")
			insertProduct := bson.M{"product_id": cpp.ProductID}

			pushToArray := bson.M{"$addToSet": bson.M{"cus_price": cpp.CustomerID}}
			_, err = collection.UpdateOne(context.TODO(), insertProduct, pushToArray)
			if err != nil {
				fmt.Println("err while insert result: ", err)
				return
			}

			continue
		}
		price = append(price, product.Price)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get customer price succeeded",
		Data: price,
	})
}


func CustomerPrice(c *gin.Context) {

	data, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Println("get raw data: ", string(data))
	var orderProducts models.OrderProducts
	err := json.Unmarshal(data, &orderProducts)
	if err != nil {
		fmt.Println("error while unmarshaling: ", err)
		return
	}
	SmartPrint(orderProducts)

	var prices []models.CustomerOrderProductPrice

	collection := models.Client.Collection("customer_product_price")
	filter := bson.M{}

	//filter["com_id"] = c.GetInt64("com_id")

	// 从客户商品价格表中找到对应商品的客户价格
	// 如果没有找到，则将价格设置为商品的默认价格


	filter["product_id"] = bson.M{"$in": orderProducts.ProductsID}
	filter["customer_id"] = bson.M{"$eq": orderProducts.CustomerID}
	fmt.Println("filter: ", filter)
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("err while finding record: ", err)
		return
	}
	for cur.Next(context.TODO()) {

		var result models.CustomerOrderProductPrice
		if err := cur.Decode(&result); err != nil {
			fmt.Println("error while decoding: ", err)
			return
		}
		prices = append(prices, result)
	}
	fmt.Println("get data: ", prices)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get customer price succeeded",
		Data: prices,
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

	updateCustomerOrder := models.CustomerOrder{}
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
			"contacts":       updateCustomerOrder.Contacts,
			"receiver_phone": updateCustomerOrder.Phone,
			"amount":         updateCustomerOrder.Amount,
			"Delivery":       updateCustomerOrder.Delivery,
			"warehouse_id":   updateCustomerOrder.WarehouseID,
			"receiver":       updateCustomerOrder.Receiver,
			"price":          updateCustomerOrder.TotalPrice,
			"extra_amount":   updateCustomerOrder.ExtraAmount,
			"delivery_code":  updateCustomerOrder.DeliveryCode,}})
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
	filter["order_sn"] = order_sn.OrderSN

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

type OrderDetail struct {
	Order models.CustomerOrder `json:"order"`
	SubOrders []models.CustomerSubOrder `json:"sub_orders"`
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
	order := models.CustomerOrder{}
	order_sn := GetCustomerOrderSNService{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	err = json.Unmarshal(data, &order_sn)
	if err != nil {
		fmt.Println("error found: ", err)
	}

	filter := bson.M{}
	//filter["com_id"] = com.ComId
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

	var suborders []models.CustomerSubOrder
	collection = models.Client.Collection("customer_suborder")
	cur, err := collection.Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		var res models.CustomerSubOrder
		if err := cur.Decode(&res); err != nil {
			fmt.Println("error decoding suborder")
			return
		}
		suborders = append(suborders, res)
	}

	resData := OrderDetail{}
	resData.Order = order
	resData.SubOrders = suborders

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Customer order detail response",
		Data: resData,
	})
}
