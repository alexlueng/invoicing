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
	"jxc/util"
	"net/http"
	"time"
)

// 所有客户订单
func AllCustomerOrders(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req models.CustomerOrderReq

	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	page, size := SetDefaultPageAndSize(req.Page, req.Size)

	// 设置排序主键
	orderFields := []string{"OrderSN", "price", "order_id"}
	option := SetPaginationAndOrder("order_id", orderFields, req.Ord, page, size)

	option.Projection = bson.M{"products": 0}
	//设置搜索规则
	filter := models.GetCustomerOrderParam(req, claims.ComId)

	collection := models.Client.Collection("customer_order")

	orderIDs := []int64{}

	order := models.CustomerOrder{}
	orders, err := order.FindAll(filter, option)

	for _, order := range orders {
		orderIDs = append(orderIDs, order.OrderId)
	}

	allSubOrders := []models.CustomerSubOrder{}
	collection = models.Client.Collection("customer_suborder")
	subOrderFilter := bson.M{}
	subOrderFilter["com_id"] = claims.ComId
	if len(orderIDs) > 0 {
		subOrderFilter["order_id"] = bson.M{"$in": orderIDs} // TODO: 要判断orderIDs里面是否有值，不然程序会报错
	}
	cur, err := collection.Find(context.TODO(), subOrderFilter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	for cur.Next(context.TODO()) {
		var res models.CustomerSubOrder
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  err.Error(),
			})
			return
		}
		allSubOrders = append(allSubOrders, res)
	}
	for _, subItem := range allSubOrders {
		for key, item := range orders {
			if subItem.OrderId == item.OrderId {
				orders[key].SubOrders = append(orders[key].SubOrders, subItem)
			}
		}
	}

	//查询的总数
	total, _ := models.Client.Collection("customer_order").CountDocuments(context.TODO(), filter)

	// 返回查询到的总数，总页数
	resData := models.ResponseCustomerOrdersData{}
	resData.Result = orders
	resData.Total = total
	resData.Pages = total/size + 1
	resData.Size = size
	resData.CurrentPage = page

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get all customer orders",
		Data: resData,
	})
}

// 生成订单的流程：
// 1. 获取选择的商品列表，数量
// 2. 获取选择的客户
// 3. 获取商品的单价
// 4. 检查包邮字段是否有值，如果有，则获取相应的值
// 5. 订单状态设置为待确认，下单时间设置为当前时间，
// 6. 算出总价
// 7. 组合好前端需要的数据并返回
func AddCustomerOrder(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	order := models.CustomerOrder{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(data, &order)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "创建订单失败",
		})
		return
	}

	// TODO: 需要解决浮点数与整数相乘精度丢失的问题

	var totalAmount int64
	for _, product := range order.Products {
		totalAmount += product.Quantity
	}

	order.Amount = totalAmount

	//这里需要一个订单号生成方法，日期加上6位数的编号,这个订单编号应该是全局唯一的
	order.OrderSN = GetTempOrderSN()
	order.OrderId = GetLastID("customer_order")
	order.ComID = claims.ComId

	// 创建订单的时间，以int64的类型插入到mongodb
	current_time := time.Now()
	order.OrderTime = current_time.Unix()

	//设置订单状态
	order.Status = models.TOBECONFIRMED
	order.IsPrepare = false // 备货状态

	err = order.Insert()

	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "创建订单失败",
		})
		return
	}
	err = SetLastID("customer_order")
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "创建订单失败",
		})
		return
	}
	var subOrders []interface{}

	collection := models.Client.Collection("customer_suborder")
	// 把订单中的每个子项插入到客户订单实例表中
	for _, item := range order.Products {
		var result models.CustomerSubOrder
		result.ComID = claims.ComId

		subSn_str, _ := util.GetOrderSN(result.ComID)
		result.SubOrderId =GetLastID("sub_order")
		result.SubOrderSn = subSn_str
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
		SetLastID("sub_order")
	}

	_, err = collection.InsertMany(context.TODO(), subOrders)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "创建订单失败",
		})
		return
	}

	// 修改商品的出售数量
	collection = models.Client.Collection("product")
	for _, item := range order.Products {
		_, err := collection.UpdateOne(context.TODO(),
			bson.D{{"com_id", claims.ComId}, {"product_id", item.ProductID}},
			bson.M{"$inc": bson.M{"num": item.Quantity}})
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "创建订单失败",
			})
			return
		}
	}

	responseData := make(map[string]interface{})
	responseData["order"] = order
	responseData["sub_orders"] = subOrders

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Customer order create succeeded",
		Data: responseData,
	})
}

// 查询该商品的客户售价
func CheckCustomerPrice(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	data, _ :=ioutil.ReadAll(c.Request.Body)
	var orderProducts models.OrderProducts
	err := json.Unmarshal(data, &orderProducts)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "创建订单失败",
		})
		return
	}
	// 从客户商品价格表中找到对应商品的客户价格
	// 如果没有找到这条记录，则从商品表中找到该商品默认价格，并将这条记录插入到客户商品价格表，并将它的价格设置为商品的默认价格

	var price []float64 // 需要返回给前端的价格数组
	filter := bson.M{}
	filter["com_id"] = claims.ComId // need to get com id from middleware
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
			cpp.ComID = claims.ComId
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

// 查找该客户的该商品的价格
func CustomerPrice(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	data, _ := ioutil.ReadAll(c.Request.Body)

	var orderProducts models.OrderProducts
	err := json.Unmarshal(data, &orderProducts)
	if err != nil {
		fmt.Println("CustomerPrice error while unmarshaling: ", err)
		return
	}

	var prices []models.CustomerOrderProductPrice

	collection := models.Client.Collection("customer_product_price")
	filter := bson.M{}

	//filter["com_id"] = c.GetInt64("com_id")

	// 从客户商品价格表中找到对应商品的客户价格
	// 如果没有找到，则将价格设置为商品的默认价格

	filter["com_id"] = claims.ComId
	filter["product_id"] = bson.M{"$in": orderProducts.ProductsID}
	filter["customer_id"] = bson.M{"$eq": orderProducts.CustomerID}
	//fmt.Println("filter: ", filter)
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
	//fmt.Println("get data: ", prices)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get customer price succeeded",
		Data: prices,
	})
}

func UpdateCustomerOrder(c *gin.Context) {
	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)

	updateCustomerOrder := models.CustomerOrder{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(data, &updateCustomerOrder)
	if err != nil {
		fmt.Println("unmarshall error: ", err)
	}
	collection := models.Client.Collection("customer_order")

	filter := bson.M{}
	filter["com_id"] = claims.ComId
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
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)

	order_sn := GetCustomerOrderSNService{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &order_sn)

	filter := bson.M{}
	filter["com_id"] = claims.ComId
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
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)

	order := models.CustomerOrder{}
	order_sn := GetCustomerOrderSNService{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(data, &order_sn)
	if err != nil {
		fmt.Println("error found: ", err)
	}

	filter := bson.M{}
	//filter["com_id"] = com.ComId
	filter["com_id"] = claims.ComId
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
	if err != nil {
		fmt.Println("Can't not find customer suborder: ", err)
		return
	}
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

type OrderStatus struct {
	Status int64 `json:"status"`
}

// 列出各种未处理的订单
func UnDealOrders(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var status OrderStatus
	if err := c.ShouldBindJSON(&status); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "params error",
		})
		return
	}

	collection := models.Client.Collection("goods_instance")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["dest_id"] = 1
	filter["status"] = status.Status

	var result []models.GoodsInstance
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Can't find instances: ", err)
		return
	}
	for cur.Next(context.TODO()) {
		var res models.GoodsInstance
		if err := cur.Decode(&res); err != nil {
			fmt.Println("Can't decode instance: ", err)
			return
		}
		result = append(result, res)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Customer order detail response",
		Data: result,
	})
}

// 是否开票
func OrderInvoicing(c *gin.Context) {

}

type PrepareStockService struct {
	SubOrderID int64 `json:"sub_order_id"`
}

// 备货完成
func PrepareStock(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		fmt.Println("Can't get com_id")
		return
	}
	//claims.(auth.Claims).ComId
	var ps PrepareStockService
	if err := c.ShouldBindJSON(&ps); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "params error",
		})
		return
	}
	collection := models.Client.Collection("customer_suborder")
	filter := bson.M{}
	filter["com_id"] = claims.(*auth.Claims).ComId
	filter["sub_order_id"] = ps.SubOrderID
	updateResult, err := collection.UpdateOne(context.TODO(), filter, bson.M{"$set" : bson.M{"is_prepare": true}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: 200,
			Msg:  "备货完成失败",
		})
		return
	}
	fmt.Println("Prepare stock update result: ", updateResult.UpsertedID)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "备货完成",
	})
}
