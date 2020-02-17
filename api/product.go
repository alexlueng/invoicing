package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"jxc/models"
	"jxc/serializer"
	"net/http"
	"strings"
)

func AllProducts(c *gin.Context) {
	// 根据域名得到com_id
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	var products []models.Product
	var req models.ProductReq

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
	orderField := []string{"product_id", "com_id", "product"}
	exist := false
	fmt.Println("order field: ", req.OrdF)
	for _, v := range orderField {
		if req.OrdF == v {
			exist = true
			break
		}
	}
	if !exist {
		req.OrdF = "product_id"
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

	//1从小到大,-1从大到小
	option.SetSort(bson.D{{req.OrdF, order}})

	//1从小到大,-1从大到小
	option.SetSort(bson.D{{req.OrdF, order}})

	//IdMin,IdMax
	if req.IdMin > req.IdMax {
		t := req.IdMax
		req.IdMax = req.IdMin
		req.IdMin = t
	}
	filter := bson.M{}
	if (req.IdMin == req.IdMax) && (req.IdMin != 0) {
		//filter["id"] = bson.M{"$gte":0}
		filter["id"] = bson.M{"$eq": req.IdMin}
	} else {
		if req.IdMin > 0 {
			filter["id"] = bson.M{"$gte": req.IdMin}
		}
		if req.IdMax > 0 {
			filter["id"] = bson.M{"$lt": req.IdMax}
		}
	}
	// Product string `form:"Product"` //模糊搜索
	if req.Product != "" {
		filter["product"] = bson.M{"$regex": req.Product}
	}

	// 每个查询都要带着com_id去查
	filter["com_id"] = com.ComId
	fmt.Println("filter: ", filter)

	collection := models.Client.Collection("product")
	cur, err := collection.Find(context.TODO(), filter, option)
	if err != nil {
		fmt.Println("error while setting findoptions: ", err)
		return
	}
	for cur.Next(context.TODO()) {
		var result models.Product
		err := cur.Decode(&result)
		if err != nil {
			fmt.Println("error while decoding product")
			return
		}
		products = append(products, result)
	}

	var total int64
	cur, _ = models.Client.Collection("product").Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		total++
	}

	// 返回查询到的总数，总页数
	resData := models.ResponseProductData{}
	resData.Products = products
	resData.Total = int(total)
	resData.Pages = int(total)/int(req.Size) + 1
	resData.Size = int(req.Size)
	resData.CurrentPage = int(req.Page)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get products",
		Data: resData,
	})

}

func AddProduct(c *gin.Context) {
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Domain error",
		})
		return
	}

	data, _ := ioutil.ReadAll(c.Request.Body)
	var product models.Product

	err = json.Unmarshal(data, &product)
	if err != nil {
		fmt.Println("err found while decoding into product: ", err)
	}

	var result models.Product
	collection := models.Client.Collection("product")
	if !ENABLESAMECUSTOMER { // 不允许重名的情况，先查找数据库是否已经存在记录，如果有，则返回错误码－1
		filter := bson.M{}
		filter["com_id"] = com.ComId
		filter["product"] = result.Product
		_ = collection.FindOne(context.TODO(), filter).Decode(&result)
		if result.Product != "" {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "该商品已经存在",
			})
			return
		}
	}

	product.ProductID = int64(getLastID("product"))
	product.ComID = com.ComId

	insertResult, err := collection.InsertOne(context.TODO(), product)
	if err != nil {
		fmt.Println("Error while inserting mongo: ", err)
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Product create succeeded",
	})
}

func UpdateProduct(c *gin.Context) {
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != int(com.ModuleId) {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Domain error",
		})
		return
	}

	updateProduct := models.Product{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &updateProduct)

	// 更新的条件：更改的时候如果有同名的记录，则要判断是否有与要修改的记录的customer_id相等,如果有不相等的，则返回
	// 如果只有相等的customer_id, 则允许修改
	filter := bson.M{}

	filter["com_id"] = com.ComId
	filter["product"] = updateProduct.Product
	collection := models.Client.Collection("product")

	cur, err := collection.Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		var tempRes models.Product
		err := cur.Decode(&tempRes)
		if err != nil {
			fmt.Println("error found decoding customer: ", err)
			return
		}
		if tempRes.ProductID != updateProduct.ProductID {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "要修改的商品已经存在",
			})
			return
		}
	}

	filter = bson.M{}
	filter["com_id"] = com.ComId
	filter["product_id"] = updateProduct.ProductID
	// 更新记录
	result, err := collection.UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"product": updateProduct.Product,
			"num": updateProduct.Num,
			"price_of_suppliers": updateProduct.PriceOfSuppliers,
			"units": updateProduct.Units,
			"url": updateProduct.URL}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "更新失败",
		})
		return
	}
	fmt.Println("Update result: ", result)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Product update succeeded",
	})
}

type DeleteProductService struct {
	ID int64 `json:"product_id"`
}


func DeleteProduct(c *gin.Context) {
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	if ( (err != nil) || (models.THIS_MODULE != com.ModuleId) ){
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Domain error",
		})
		return
	}

	var d DeleteProductService

	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &d)

	filter := bson.M{}

	filter["com_id"] = com.ComId
	filter["product_id"] = d.ID
	collection := models.Client.Collection("product")
	deleteResult, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "删除商品失败",
		})
		return
	}
	fmt.Println("Delete a single document: ", deleteResult.DeletedCount)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Product delete succeeded",
	})
}
