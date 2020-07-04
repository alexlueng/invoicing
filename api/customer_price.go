package api

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"jxc/auth"
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

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var customerProductPrice models.CustomerProductPrice
	data, _ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(data, &customerProductPrice)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	if customerProductPrice.DefaultPrice > 0 {
		// 修改默认价格
		// 在商品表，客户商品价格表中都要修改
		// 在客户商品价格表中，把原来的默认价格记为false,再插入一条新的记录
		fmt.Println("execute here")

		fmt.Println("customer id is: ", customerProductPrice.CustomerID)
		fmt.Println("product id is: ", customerProductPrice.ProductID)
		filter := bson.M{}
		//filter["com_id"] = com.ComId
		filter["com_id"] = claims.ComId
		filter["product_id"] = customerProductPrice.ProductID
		//filter["customer_id"] = customerProductPrice.CustomerID
		filter["is_valid"] = true
		collection := models.Client.Collection("customer_product_price")

		var oldRecord models.CustomerProductPrice

		err := collection.FindOne(context.TODO(), filter).Decode(&oldRecord)
		if err != nil {
			fmt.Println("can't find old record: ", err)
			return
		}

		newRecord := oldRecord
		newRecord.Price = customerProductPrice.Price

		// 将旧记录设为false
		updateResult, err := collection.UpdateOne(context.TODO(), filter, bson.M{
			"$set": bson.M{"is_valid": false}})
		if err != nil {
			fmt.Println("Can't update old record: ", err)
			return
		}
		fmt.Println("update old record: ", updateResult.UpsertedID)

		// 加入一条新记录
		insertResult, err := collection.InsertOne(context.TODO(), newRecord)
		if err != nil {
			fmt.Println("Can't not insert new record: ", err)
			return
		}
		fmt.Println("insert new record: ", insertResult.InsertedID)

		filter = bson.M{}
		filter["product_id"] = customerProductPrice.ProductID
		collection = models.Client.Collection("product")
		updateResult, err = collection.UpdateOne(context.TODO(), filter, bson.M{
			"$set": bson.M{"default_price": customerProductPrice.DefaultPrice}})
		if err != nil {
			fmt.Println("Can't update default price: ", err)
			return
		}
		fmt.Println("update default product price: ", updateResult.UpsertedID)

		c.JSON(http.StatusOK, serializer.Response{
			Code: 200,
			Msg:  "update default price success",
		})
		return
	}

	//customerProductPrice.ComID = com.ComId
	customerProductPrice.ComID = claims.ComId
	// 加上一个时间戳，以及一个有效值
	timestamp := time.Now().Unix()
	customerProductPrice.CreateAt = timestamp
	customerProductPrice.IsValid = true

	collection := models.Client.Collection("customer_product_price")

	// 找到此商品上一个有效价格记录，如果有，则把它设置为无效
	filter := bson.M{}
	//filter["com_id"] = com.ComId
	filter["com_id"] = claims.ComId
	filter["product_id"] = customerProductPrice.ProductID
	filter["customer_id"] = customerProductPrice.CustomerID
	filter["is_valid"] = true
	// 找到这个商品的默认价格
	proCollects := models.Client.Collection("product")
	var res models.Product
	err = proCollects.FindOne(context.TODO(), bson.D{{"product_id", customerProductPrice.ProductID}}).Decode(&res)
	if err != nil {
		fmt.Println("Can't get default product price: ", err)
		return
	}

	customerProductPrice.DefaultPrice = res.DefaultPrice

	var result models.CustomerProductPrice
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		//没有找到这个记录，说明这个客户价格是新增的
		//保存这条记录，更新product表中的cus_product字段
		fmt.Println("no document found, this is a new record")

		_, err = collection.InsertOne(context.TODO(), customerProductPrice)
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
	ID           int64   `bson:"product_id"`
	Product      string  `bson:"product"`
	CusPrice     []int64 `bson:"cus_price"`
	DefaultPrice float64 `bson:"default_price"`
}

func ListCustomerPrice(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	// 得到所有的商品id
	// 得到所有的客户

	// 可以分页，搜索

	// get com_id

	var req models.CustomerProductPriceReq

	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	req.Page, req.Size = SetDefaultPageAndSize(req.Page, req.Size)

	option := options.Find()
	option.SetLimit(int64(req.Size))
	option.SetSkip((int64(req.Page) - 1) * int64(req.Size))
	option.Projection = bson.M{"product_id": 1, "product": 1, "cus_price": 1, "default_price": 1, "_id": 0}

	filter := bson.M{}
	filter["com_id"] = claims.ComId

	// mongodb中返回指定字段的写法
	//opts := options.FindOne()
	//opts.Projection = bson.M{"cus_price":1, "_id": 0}
	var allProducts []ProductList
	// 得到当前分页中的商品列表
	collection := models.Client.Collection("product")
	cur, err := collection.Find(context.TODO(), filter, option)

	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't get products",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var result ProductList
		err := cur.Decode(&result)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't decode supplier list",
			})
			return
		}
		allProducts = append(allProducts, result)
	}

	responseData := make(map[string]map[string]interface{})

	// 根据商品id得到客户名和售价
	// 在商品表中维护一个售价客户id,刚可省去一次循环查找数据库的工作
	// 可以直接从商品表中的cus_price字段中得到已有售价记录的客户id
	filter = bson.M{}
	filter["com_id"] = claims.ComId

	allProductsID := []int64{}
	for _, product := range allProducts {
		allProductsID = append(allProductsID, product.ID)
		responseData[product.Product] = make(map[string]interface{})
	}

	// 按商品名字去搜索
	// TODO: 可以优化这个流程，因为这里只选择一种商品，所以不用循环整个product表了
	if req.Product != "" {
		filter["product_name"] = bson.M{"$regex": req.Product}
	}
	if req.CustomerName != "" {
		filter["customer_name"] = bson.M{"$regex": req.CustomerName,}
	}
	if len(allProductsID) > 0 {
		filter["product_id"] = bson.M{"$in": allProductsID}
	}
	filter["is_valid"] = true
	collection = models.Client.Collection("customer_product_price")

	cur, err = collection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Can't not use or like that: ", err)
		return
	}
	for cur.Next(context.TODO()) {
		var res models.CustomerProductPrice
		if err := cur.Decode(&res); err != nil {
			fmt.Println("err: ", err)
			return
		}

		responseData[res.Product]["product_id"] = res.ProductID
		if responseData[res.Product]["customer_price"] == nil {
			responseData[res.Product]["customer_price"] = []models.CustomerProductPrice{} //make(map[string]models.CustomerProductPrice)
		}
		if res.CustomerID == 0 {
			if responseData[res.Product]["default_price"] == nil {
				responseData[res.Product]["default_price"] = models.CustomerProductPrice{}
			}
			responseData[res.Product]["default_price"] = res
			continue
		}
		responseData[res.Product]["customer_price"] = append(responseData[res.Product]["customer_price"].([]models.CustomerProductPrice), res)
	}

	if req.CustomerName != "" {
		filter["customer_name"] = bson.M{"$eq": "default"}
		cur, err = collection.Find(context.TODO(), filter)
		if err != nil {
			fmt.Println("Can't not use or like that: ", err)
			return
		}
		for cur.Next(context.TODO()) {
			var res models.CustomerProductPrice
			if err := cur.Decode(&res); err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: serializer.CodeError,
					Msg:  "Can't decode customer product price",
				})
				return
			}
			if responseData[res.Product]["default_price"] == nil {
				responseData[res.Product]["default_price"] = models.CustomerProductPrice{} //make(map[string]models.CustomerProductPrice)
			}
			responseData[res.Product]["default_price"] = res
		}

	}

	var total int
	//total, _ = models.Client.Collection("customer_product_price").CountDocuments(context.TODO(), filter)
	total = len(responseData)

	res := models.ResponseCustomerProductPriceData{}
	res.PriceTable = responseData
	res.Size = int(req.Size)
	res.Pages = int(req.Page)
	res.CurrentPage = int(total)/int(req.Size) + 1
	res.Total = int(total)

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Get all products",
		Data: res,
	})

}

func DeleteCustomerPrice(c *gin.Context) {
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)

	var req models.CustomerProductPriceReq

	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &req)

	collection := models.Client.Collection("customer_product_price")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
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
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Delete customer price success",
	})

}
