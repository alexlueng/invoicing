package api

import (
	"context"
	"encoding/json"
	"strings"

	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"

	"github.com/gin-gonic/gin"
	"io/ioutil"
	"jxc/models"
	"jxc/serializer"
	"net/http"
)

// AddCustomerPrice 操作的是customer_product_price这张表
// 主要有两个地方使用：1.售价管理页面 2.客户下订单时没有对应的售价
func AddCustomerPrice(c *gin.Context) {
	//com_id customer_id customer_name product_id product_name price

	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	var customerProductPrice models.CustomerProductPrice
	data, _ := ioutil.ReadAll(c.Request.Body)
	err = json.Unmarshal(data, &customerProductPrice)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	//customerProductPrice.ComID = com.ComId
	customerProductPrice.ComID = com.ComId
	// 加上一个时间戳，以及一个有效值
	timestamp := time.Now().Unix()
	customerProductPrice.CreateAt = timestamp
	customerProductPrice.IsValid = true



	collection := models.Client.Collection("customer_product_price")

	// 找到此商品上一个有效价格记录，如果有，则把它设置为无效
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["product_id"] = customerProductPrice.ProductID
	filter["customer_id"] = customerProductPrice.CustomerID
	filter["is_valid"] = true

	var result models.CustomerProductPrice
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		//没有找到这个记录，说明这个客户价格是新增的
		//保存这条记录，更新product表中的cus_product字段
		fmt.Println("no document found, this is a new record")

		_, err := collection.InsertOne(context.TODO(), customerProductPrice)
		if err != nil {
			fmt.Println("Update customer price failed: ", err)
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "添加记录错误",
			})
			return
		}

		// 更新商品客户列表，把客户id追加到cus_price数组中
		collection = models.Client.Collection("product")
		insertProduct := bson.M{"product_id": customerProductPrice.ProductID}

		pushToArray := bson.M{"$addToSet": bson.M{"cus_price": customerProductPrice.CustomerID}}
		_, err = collection.UpdateOne(context.TODO(), insertProduct, pushToArray)
		if err != nil {
			fmt.Println("update cus_price err: ", err)
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "参数解释错误",
			})
			return
		}

	} else {
		// 找到了旧记录
		// 把旧记录的is_valid字段更新为false,然后插入这条记录
		fmt.Println("change old record: ", filter)
		_, err = collection.UpdateOne(context.TODO(), filter, bson.M{
			"$set": bson.M{"is_valid": false}})
		if err != nil {
			fmt.Println("Update customer price failed: ", err)
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "添加记录错误",
			})
			return
		}
		_, err := collection.InsertOne(context.TODO(), customerProductPrice)
		if err != nil {
			fmt.Println("Update customer price failed: ", err)
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "添加记录错误",
			})
			return
		}
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Insert record succeeded",
	})
	return
}

type ProductList struct {
	ID int64 `bson:"product_id"`
	Product string `bson:"product"`
	CusPrice []int64 `bson:"cus_price"`
	DefaultPrice float64 `bson:"default_price"`
}

func ListCustomerPrice(c *gin.Context) {

	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}


	// 得到所有的商品id
	// 得到所有的客户

	// 可以分页，搜索

	// get com_id

	var req models.CustomerProductPriceReq

	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	fmt.Println("customer name: ", req.CustomerName)

	req.Page, req.Size = SetDefaultPageAndSize(req.Page, req.Size)

	option := options.Find()
	option.SetLimit(int64(req.Size))
	option.SetSkip((int64(req.Page) - 1) * int64(req.Size))
	//	option.SetSort(bson.D{{req.OrdF, order}})
	option.Projection = bson.M{"product_id": 1, "product": 1, "cus_price": 1, "default_price":1, "_id": 0}

	//option.S

	filter := bson.M{}
	filter["com_id"] = com.ComId

	// 按商品名字去搜索
	// TODO: 可以优化这个流程，因为这里只选择一种商品，所以不用循环整个product表了
	if req.Product != "" {
		filter["product"] = bson.M{"$regex": req.Product}
	}

	// mongodb中返回指定字段的写法
	//opts := options.FindOne()
	//opts.Projection = bson.M{"cus_price":1, "_id": 0}
	var allProducts []ProductList
	// 得到当前分页中的商品列表
	collection := models.Client.Collection("product")
	cur, err := collection.Find(context.TODO(), filter, option)

	// 对于cur, file等操作，都需要defer close掉，防止内存泄露
	defer cur.Close(context.TODO())

	if err != nil {
		fmt.Println("error found finding products: ", err)
		return
	}
	for cur.Next(context.TODO()) {
		var result ProductList
		err := cur.Decode(&result)
		if err != nil {
			fmt.Println("error found decoding product: ", err)
			return
		}
		//fmt.Println("product name: ", result.Product)

		allProducts = append(allProducts, result)
	}

	responseData := make(map[string]map[string]interface{})

	// 根据商品id得到客户名和售价
	// 在商品表中维护一个售价客户id,刚可省去一次循环查找数据库的工作

	// 可以直接从商品表中的cus_price字段中得到已有售价记录的客户id
	// product.cus_price
	filter = bson.M{}
	filter["com_id"] = com.ComId
	if req.CustomerName != "" {
		filter["customer_name"] = req.CustomerName
	}
	collection = models.Client.Collection("customer_product_price")

	for _, product := range allProducts {

		var customerList []models.CustomerProductPrice
		for _, id := range product.CusPrice {

			var result models.CustomerProductPrice
			if req.CustomerName != "" {
				filter = bson.M{"customer_id": id, "product_name":product.Product, "is_valid": true, "customer_name": req.CustomerName}
			} else {
				filter = bson.M{"customer_id": id, "product_name":product.Product, "is_valid": true}
			}



			err := collection.FindOne(context.TODO(), filter).Decode(&result)

			if err != nil {
				//fmt.Println("error found finding Customers price: ", err)
				//没有记录
				//return
				continue
			}
			//fmt.Println("appending: ", result.CustomerName, result.Product, result.Price)
			customerList = append(customerList, result)
		}
		//fmt.Println("CustomerList: ", customerList)
		if responseData[product.Product] == nil {
			responseData[product.Product] = make(map[string]interface{})
		}
		responseData[product.Product]["default_price"] = product.DefaultPrice
		responseData[product.Product]["product_id"] = product.ID
		responseData[product.Product]["customer_price"] = customerList

	}


	var total int64
	cur, _ = models.Client.Collection("product").Find(context.TODO(), bson.D{})
	for cur.Next(context.TODO()) {
		total++
	}

	res := models.ResponseCustomerProductPriceData{}
	//res.DefaultPrice = allProducts
	res.PriceTable = responseData
	res.Size = int(req.Size)
	res.Pages = int(req.Page)
	res.CurrentPage = int(total)/int(req.Size) + 1
	res.Total = int(total)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get all products",
		Data: res,
	})

}

func DeleteCustomerPrice(c *gin.Context) {
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	var req models.CustomerProductPriceReq

	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	collection := models.Client.Collection("customer_product_price")
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["product_id"] = req.ProductID
	filter["customer_id"] = req.CustomerID
	filter["is_valid"] = true

	collection.UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"is_valid": false,}})

	//collection = models.Client.Collection("product")
	//updateProduct := bson.M{"product_id": req.ProductID}
	//
	//pushToArray := bson.M{"$addToSet": bson.M{"cus_price": customerProductPrice.CustomerID}}
	//

}