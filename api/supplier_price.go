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
// 主要有两个地方使用：1.进价管理页面 2.商品管理页面中指定供应商
func AddSupplierPrice(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var supplierProductPrice models.SupplierProductPrice
	data, _ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(data, &supplierProductPrice)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "参数解释错误",
		})
		return
	}

	if supplierProductPrice.Price < 0 {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "商品价格不能为负",
		})
		return
	}

	supplierProductPrice.ComID = claims.ComId
	timestamp := time.Now().Unix() // 加上一个时间戳，以及一个有效值
	supplierProductPrice.CreateAt = timestamp
	supplierProductPrice.IsValid = true

	collection := models.Client.Collection("supplier_product_price")

	// 找到此商品上一个有效价格记录，如果有，则把它设置为无效
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["product_id"] = supplierProductPrice.ProductID
	filter["supplier_id"] = supplierProductPrice.SupplierID
	filter["is_valid"] = true

	var result models.SupplierProductPrice
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		// 说明没有找到记录
		// 在商品表中的sup_price中更新
		// 把记录插入到supplier_product_price表中
		fmt.Println("no document found, this is a new record")

		_, err := collection.InsertOne(context.TODO(), supplierProductPrice)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "添加进价记录错误",
			})
			return
		}

		// 更新商品客户列表，把客户id追加到sup_price数组中
		collection = models.Client.Collection("product")
		insertProduct := bson.M{"product_id": supplierProductPrice.ProductID}
		pushToArray := bson.M{"$addToSet": bson.M{"sup_price": supplierProductPrice.SupplierID}}
		_, err = collection.UpdateOne(context.TODO(), insertProduct, pushToArray)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Failed to update supplier to sup_price",
			})
			return
		}

		collection := models.Client.Collection("supplier") // 更新供应商表中的supply_list字段，把商品id追加进去
		filter = bson.M{}
		filter["com_id"] = claims.ComId
		filter["supplier_id"] = supplierProductPrice.SupplierID
		pushToArray = bson.M{"$addToSet": bson.M{"supply_list": supplierProductPrice.ProductID}}
		_, err = collection.UpdateOne(context.TODO(), filter, pushToArray)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Failed to add pruduct to supplier",
			})
			return
		}

	} else {
		// 找到了旧记录
		// 把旧记录的is_valid字段更新为false,然后插入这条记录
		_, err = collection.UpdateOne(context.TODO(), filter, bson.M{
			"$set": bson.M{"is_valid": false}})
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "修改进价记录错误",
			})
			return
		}
		_, err := collection.InsertOne(context.TODO(), supplierProductPrice)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "修改进价记录错误",
			})
			return
		}

	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Insert record succeeded",
		Data: supplierProductPrice,
	})
	return
}

type SupplierList struct {
	ID       int64   `bson:"product_id"`
	Product  string  `bson:"product"`
	SupPrice []int64 `bson:"sup_price"`
}

func ListSupplierPrice(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req models.SupplierProductPriceReq

	var allProducts []SupplierList
	var allProductsID []int64

	responseData := make(map[string]map[string]interface{})

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
	option.Projection = bson.M{"product_id": 1, "product": 1, "sup_price": 1, "_id": 0}

	filter := bson.M{}
	filter["com_id"] = claims.ComId

	// mongodb中返回指定字段的写法
	//opts := options.FindOne()
	//opts.Projection = bson.M{"cus_price":1, "_id": 0}
	// 按商品名字去搜索
	// TODO: 可以优化这个流程，因为这里只选择一种商品，所以不用循环整个product表了

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
		var result SupplierList
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

	for _, product := range allProducts {
		allProductsID = append(allProductsID, product.ID)
		responseData[product.Product] = make(map[string]interface{})
		responseData[product.Product]["product_id"] = product.ID
	}

	// 根据商品id得到客户名和售价
	// 在商品表中维护一个售价客户id,刚可省去一次循环查找数据库的工作
	// 可以直接从商品表中的sup_price字段中得到已有售价记录的客户id
	filter = bson.M{}
	filter["com_id"] = claims.ComId
	if req.Product != "" {
		filter["product"] = bson.M{"$regex": req.Product}
	}
	if req.SupplierName != "" {
		filter["supplier"] = bson.M{"$regex": req.SupplierName}
	}
	if len(allProductsID) > 0 {
		filter["product_id"] = bson.M{"$in": allProductsID}
	}
	filter["is_valid"] = true

	collection = models.Client.Collection("supplier_product_price")
	cur, err = collection.Find(context.TODO(), filter)

	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find supplier_product_price",
		})
		return
	}

	for cur.Next(context.TODO()) {
		var res models.SupplierProductPrice
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't decode supplier product price",
			})
			return
		}

		responseData[res.Product]["product_id"] = res.ProductID
		if responseData[res.Product]["supplier_price"] == nil {
			responseData[res.Product]["supplier_price"] = []models.SupplierProductPrice{}
		}

		responseData[res.Product]["supplier_price"] = append(responseData[res.Product]["supplier_price"].([]models.SupplierProductPrice), res)
	}

	var total int64
	total, _ = models.Client.Collection("product").CountDocuments(context.TODO(), bson.D{{"com_id", claims.ComId}})

	res := models.ResponseSupplierProductPriceData{}
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

// 删除商品的供应商
func DeleteSupplierPrice(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req models.SupplierProductPriceReq

	data, _ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(data, &req)

	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "参数解释错误",
		})
		return
	}

	// supplier_product_price表中找出该商品供应商价格，置为false
	collection := models.Client.Collection("supplier_product_price")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["product_id"] = req.ProductID
	filter["supplier_id"] = req.SupplierID
	filter["is_valid"] = true

	_, err = collection.UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"is_valid": false,}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't set false to supplier product price",
		})
		return
	}

	// 在商品列表中的sup_price这个数组中去掉这个供应商id
	proCollects := models.Client.Collection("product")
	_, err = proCollects.UpdateOne(context.TODO(), bson.D{{"product_id", req.ProductID}}, bson.M{"$pull": bson.M{"sup_price": req.SupplierID}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't drop supplier from supplier product price",
		})
		return
	}

	// 在该供应商中去掉这个商品id
	supCollects := models.Client.Collection("supplier")
	_, err = supCollects.UpdateOne(context.TODO(), bson.D{{"supplier_id", req.SupplierID}}, bson.M{"$pull": bson.M{"supply_list": req.ProductID}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't set false to supplier product price",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Delete supplier price success",
	})

}
