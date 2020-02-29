package api

import (
	"context"
	"encoding/json"
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
// 主要有两个地方使用：1.售价管理页面 2.商品管理页面中指定供应商
func AddSupplierPrice(c *gin.Context) {
	//com_id customer_id customer_name product_id product_name price

	//com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	//if err != nil || models.THIS_MODULE != int(com.ModuleId) {
	//	c.JSON(http.StatusOK, serializer.Response{
	//		Code: -1,
	//		Msg:  "域名错误",
	//	})
	//	return
	//}

	var supplierProductPrice models.SupplierProductPrice
	data, _ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(data, &supplierProductPrice)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	//supplierProductPrice.ComID = com.ComId
	supplierProductPrice.ComID = 1
	// 加上一个时间戳，以及一个有效值
	timestamp := time.Now().Unix()
	supplierProductPrice.CreateAt = timestamp
	supplierProductPrice.IsValid = true

	//SmartPrint(supplierProductPrice)

	collection := models.Client.Collection("supplier_product_price")

	// 找到此商品上一个有效价格记录，如果有，则把它设置为无效
	filter := bson.M{}
	filter["com_id"] = 1
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

		//SmartPrint(customerProductPrice)

		_, err := collection.InsertOne(context.TODO(), supplierProductPrice)
		if err != nil {
			fmt.Println("Update supplier price failed: ", err)
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "添加记录错误",
			})
			return
		}

		// 更新商品客户列表，把客户id追加到sup_price数组中
		collection = models.Client.Collection("product")
		insertProduct := bson.M{"product_id": supplierProductPrice.ProductID}

		pushToArray := bson.M{"$addToSet": bson.M{"sup_price": supplierProductPrice.SupplierID}}
		_, err = collection.UpdateOne(context.TODO(), insertProduct, pushToArray)
		if err != nil {
			fmt.Println("update sup_price err: ", err)
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "参数解释错误",
			})
			return
		}

		// 更新供应商表中的supply_list字段，把商品id追加进去
		collection := models.Client.Collection("supplier")
		filter = bson.M{}
		filter["com_id"] = 1
		filter["supplier_id"] = supplierProductPrice.SupplierID
		pushToArray = bson.M{"$addToSet": bson.M{"supply_list": supplierProductPrice.ProductID}}
		_, err = collection.UpdateOne(context.TODO(), filter, pushToArray)
		if err != nil {
			fmt.Println("update supply_list err: ", err)
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "参数解释错误",
			})
			return
		}

	} else {
		// 找到了旧记录
		// 把旧记录的is_valid字段更新为false,然后插入这条记录
		_, err = collection.UpdateOne(context.TODO(), filter, bson.M{
			"$set": bson.M{"is_valid": false}})
		if err != nil {
			fmt.Println("Update supplier price failed: ", err)
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "添加记录错误",
			})
			return
		}
		_, err := collection.InsertOne(context.TODO(), supplierProductPrice)
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
type SupplierList struct {
	ID int64 `bson:"product_id"`
	Product string `bson:"product"`
	SupPrice []int64 `bson:"sup_price"`
}

func ListSupplierPrice(c *gin.Context) {
	// 得到所有的商品id
	// 得到所有的客户

	// 可以分页，搜索

	var req models.SupplierProductPriceReq

	var allProducts []SupplierList

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
	//	option.SetSort(bson.D{{req.OrdF, order}})
	option.Projection = bson.M{"product_id": 1, "product": 1, "sup_price": 1, "_id": 0}

	//option.S

	filter := bson.M{}
	filter["com_id"] = 1

	// mongodb中返回指定字段的写法
	//opts := options.FindOne()
	//opts.Projection = bson.M{"cus_price":1, "_id": 0}
	// 按商品名字去搜索
	// TODO: 可以优化这个流程，因为这里只选择一种商品，所以不用循环整个product表了
	if req.Product != "" {
		filter["product"] = bson.M{"$regex": req.Product}
	}

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
		var result SupplierList
		err := cur.Decode(&result)
		if err != nil {
			//fmt.Println("error found decoding product: ", err)
			return
		}
		//fmt.Println("product name: ", result.Product)
		allProducts = append(allProducts, result)
	}

	responseData := make(map[string]map[string]interface{})

	// 根据商品id得到客户名和售价
	// 在商品表中维护一个售价客户id,刚可省去一次循环查找数据库的工作

	// 可以直接从商品表中的sup_price字段中得到已有售价记录的客户id
	// product.cus_price
	filter = bson.M{}
	filter["com_id"] = 1
	collection = models.Client.Collection("supplier_product_price")
	//filter["product_id"] = product_id
	for _, product := range allProducts {

		var supplierList []models.SupplierProductPrice
		for _, id := range product.SupPrice {
			var result models.SupplierProductPrice
		//	filter = bson.M{"supplier_id": id, "product_id": product.ID,  "is_valid": true}

			if req.SupplierName != "" {
				fmt.Println("exec here")
				filter = bson.M{"supplier_id": id, "product":product.Product, "is_valid": true, "supplier": req.SupplierName}
			} else {
				filter = bson.M{"supplier_id": id, "product":product.Product, "is_valid": true}
			}

			err := collection.FindOne(context.TODO(), filter).Decode(&result)
			if err != nil {
				fmt.Println("error found finding suppliers price: ", err)
				//没有记录
				//return
				continue
			}
			supplierList = append(supplierList, result)
		}
		if responseData[product.Product] == nil {
			responseData[product.Product] = make(map[string]interface{})
		}
		responseData[product.Product]["product_id"] = product.ID
		responseData[product.Product]["supplier_price"] = supplierList

	}

	var total int64
	cur, _ = models.Client.Collection("product").Find(context.TODO(), bson.D{})
	for cur.Next(context.TODO()) {
		total++
	}

	res := models.ResponseSupplierProductPriceData{}
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
