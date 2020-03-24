package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"jxc/util"
	"net/http"
)

const ENABLESAMEPRODUCT = false

func AllProducts(c *gin.Context) {
	//根据域名得到com_id
	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)


	var products []models.Product
	var req models.ProductReq

	err := c.ShouldBind(&req)
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
	filter["com_id"] = claims.ComId
	//filter["com_id"] = 1
	//fmt.Println("filter: ", filter)

	collection := models.Client.Collection("product")
	cur, err := collection.Find(context.TODO(), filter, option)
	if err != nil {
		fmt.Println("can't find products")
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var result models.Product
		err := cur.Decode(&result)
		if err != nil {
			fmt.Println("can't decoding products: ", err)
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "参数解释错误",
			})
			return
		}
		products = append(products, result)
	}

	var total int64
	cur, _ = models.Client.Collection("product").Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		total++
	}

	// 返回所有的供应商
	var productSuppliers []models.ProductSupplier
	cur, _ = models.Client.Collection("supplier").Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		var result models.ProductSupplier
		if err := cur.Decode(&result); err != nil {
			fmt.Println("can't decoding supplier")
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "参数解释错误",
			})
			return
		}
		productSuppliers = append(productSuppliers, result)
	}

	// 返回查询到的总数，总页数
	resData := models.ResponseProductData{}
	resData.Products = products
	//resData.Suppliers = productSuppliers
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
	//com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	//if err != nil || models.THIS_MODULE != int(com.ModuleId) {
	//	c.JSON(http.StatusOK, serializer.Response{
	//		Code: -1,
	//		Msg:  "Domain error",
	//	})
	//	return
	//}
	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)


	// 商品应该和供应商联系在一起
	// 获取供应商以及供应商的价格
	// 在供应商表中添加对应的商品信息

	data, _ := ioutil.ReadAll(c.Request.Body)
	var product models.Product

	err := json.Unmarshal(data, &product)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "添加商品失败",
		})
		return
	}

	//var product models.Product
	//if err := c.ShouldBindQuery(&product); err != nil {
	//	c.JSON(http.StatusOK, serializer.Response{
	//		Code: -1,
	//		Msg:  "添加商品失败",
	//	})
	//	return
	//}
	//file, err := c.FormFile("images")
	//if err != nil {
	//	c.JSON(http.StatusOK, serializer.Response{
	//		Code: -1,
	//		Msg:  "添加商品图片失败",
	//	})
	//	return
	//}
	//
	//fmt.Println("file name: ", file.Filename)
	//save_path := os.Getenv("UPLOAD_PATH")
	//fmt.Println("save path: ", save_path)
	//_, err = os.Stat(save_path)
	//if err != nil {
	//	fmt.Println("filepath not exist: ", err)
	//	return
	//}
	//
	//if err = c.SaveUploadedFile(file, save_path+file.Filename); err != nil {
	//	c.String(http.StatusBadRequest, fmt.Sprintf("上传错误: %s", err.Error()))
	//	return
	//}

	SmartPrint(product)


	var result models.Product
	collection := models.Client.Collection("product")
	if !ENABLESAMEPRODUCT { // 不允许重名的情况，先查找数据库是否已经存在记录，如果有，则返回错误码－1
		filter := bson.M{}
		//filter["com_id"] = com.ComId
		filter["com_id"] = claims.ComId
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
	//product.ComID = com.ComId
	product.ComID = claims.ComId

	product.SupPrice = []int64{}
	product.CusPrice = []int64{}

	// 默认价格是default价格乘以公司的利润率
	//var company models.Company
	//_ = models.Client.Collection("company").FindOne(context.TODO(),bson.D{{"com_id", 1}}).Decode(&company)
	//default_margin := company.DefaultProfitMargin
	//product.DefaultPrice += product.DefaultPrice * float64(default_margin) / 100

	SmartPrint(product)

	_, err = collection.InsertOne(context.TODO(), product)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	// 把商品的默认记录存到客户商品价格表中
	cusProPrice := models.CustomerProductPrice{}
	cusProPrice.IsValid = true
	cusProPrice.ComID = claims.ComId
	cusProPrice.Product = product.Product
	cusProPrice.ProductID = product.ProductID
	cusProPrice.Price = product.DefaultPrice
	cusProPrice.CustomerName = "default"

	collection = models.Client.Collection("customer_product_price")
	insertResult, err := collection.InsertOne(context.TODO(), cusProPrice)

	if err != nil {
		fmt.Println("Can't insert cus pro price: ", err)
		return
	}
	fmt.Println("insert default price success: ", insertResult.InsertedID)




	//fmt.Println("Inserted a single document: ", insertResult.InsertedID)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Product create succeeded",
		Data: product.ProductID,
	})
}

// 更新商品需要修改三张表：product customer_product_price supplier_product_price
func UpdateProduct(c *gin.Context) {
	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)


	updateProduct := models.Product{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &updateProduct)

	// 更新的条件：更改的时候如果有同名的记录，则要判断是否有与要修改的记录的customer_id相等,如果有不相等的，则返回
	// 如果只有相等的customer_id, 则允许修改
	filter := bson.M{}

	filter["com_id"] = claims.ComId
	filter["product"] = updateProduct.Product
	collection := models.Client.Collection("product")

	cur, err := collection.Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		var tempRes models.Product
		err := cur.Decode(&tempRes)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "参数解释错误",
			})
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

	var oldRecord models.Product
	err = collection.FindOne(context.TODO(), filter).Decode(&oldRecord)
	if err != nil {
		fmt.Print("Can not find old product: ", err)
		return
	}

	// 上传的价格和原来的价格不一样，说明更新了价格
	if oldRecord.DefaultPrice != updateProduct.DefaultPrice {
		cusProPriceCollects := models.Client.Collection("customer_product_price")
		// 将旧记录设置为false
		var oldRecord models.CustomerProductPrice

		filter := bson.M{}
		filter["com_id"] = claims.ComId
		filter["product_id"] = updateProduct.ProductID
		filter["customer_id"] = 0
		filter["is_valid"] = true

		err := cusProPriceCollects.FindOne(context.TODO(), filter).Decode(&oldRecord)
		if err != nil {
			fmt.Println("can't find old record: ", err)
			return
		}

		newRecord := oldRecord
		newRecord.Price = updateProduct.DefaultPrice

		// 将旧记录设为false
		updateResult, err := cusProPriceCollects.UpdateOne(context.TODO(), filter, bson.M{
			"$set": bson.M{"is_valid": false}})
		if err != nil {
			fmt.Println("Can't update old record: ", err)
			return
		}
		fmt.Println("update old record: ", updateResult.UpsertedID)

		// 加入一条新记录
		insertResult, err := cusProPriceCollects.InsertOne(context.TODO(), newRecord)
		if err != nil {
			fmt.Println("Can't not insert new record: ", err)
			return
		}
		fmt.Println("insert new record: ", insertResult.InsertedID)
	}


	filter = bson.M{}
	filter["com_id"] = claims.ComId
	filter["product_id"] = updateProduct.ProductID
	// 更新记录
	_, err = collection.UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"product": updateProduct.Product,
			"num": updateProduct.Num,
			"units": updateProduct.Units,
			"urls": updateProduct.URLS,
			"default_price": updateProduct.DefaultPrice}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "更新失败",
		})
		return
	}

	// update customer_product_price
	collection = models.Client.Collection("customer_product_price")
	collection.UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{
			"product_name": updateProduct.Product,}})

	// update supplier_product_price
	collection = models.Client.Collection("supplier_product_price")
	collection.UpdateMany(context.TODO(), filter, bson.M{
		"$set": bson.M{
			"product": updateProduct.Product,}})




	//fmt.Println("Update result: ", result)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Product update succeeded",
	})
}

type ProductService struct {
	ID int64 `json:"product_id"`
}


// 删除商品要更新四个表：1, product 2, customer_product_price和supplier_product_price表中将此商品的is_valid字段置为false 3，supplier表中supply_list中有这个商品id的值去掉
func DeleteProduct(c *gin.Context) {
	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)


	var d ProductService

	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &d)

	filter := bson.M{}

	filter["com_id"] = claims.ComId
	filter["product_id"] = d.ID
	collection := models.Client.Collection("product")

	_, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "删除商品失败",
		})
		return
	}

	filter = bson.M{}

	filter["com_id"] = claims.ComId
	filter["product_id"] = d.ID
	filter["is_valid"] = true
	// update customer_product_price
	collection = models.Client.Collection("customer_product_price")
	collection.UpdateMany(context.TODO(), filter, bson.M{
		"$set": bson.M{
			"is_valid": false,}})

	// update supplier_product_price
	collection = models.Client.Collection("supplier_product_price")
	collection.UpdateMany(context.TODO(), filter, bson.M{
		"$set": bson.M{
			"is_valid": false,}})

	// $pull删除子数组中的元素

	//fmt.Println("Delete a single document: ", deleteResult.DeletedCount)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Product delete succeeded",
	})
}


// 供应商价格表
type ProductPrice struct {
	ProductID int64 `json:"product_id" bson:"product_id"`
	Product string `json:"product" bson:"product"`
	SupplierID int64 `json:"supplier_id" bson:"supplier_id"`
	Supplier string `json:"supplier" bson:"supplier"`
	Price float64 `json:"price" bson:"price"`
}


// 添加此商品的供应商价格
// 把商品id保存到供应商列表的字段中，以便快速查询此供应商能供应的产品
func AddPrice(c *gin.Context) {

	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)


	var supplierPrice models.SupplierProductPrice

	data, _ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(data, &supplierPrice)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	//SmartPrint(supplierPrice)

	//Who := bson.M{"_id" : "57437f271e02f4bae78788af"}
	//PushToArray := bson.M{"$push": bson.M{"loginattempts": bson.M{"timestamp": 1464045212, "ip": "195.0.0.200"}}}
	//collection.Update(Who, PushToArray)

	var productPrice ProductPrice

	err = json.Unmarshal(data, &productPrice)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	SmartPrint(productPrice)

	//把这个记录插入供应商商品价格表
	collection := models.Client.Collection("supplier_product_price")
	//insertResult, err := collection.InsertOne(context.TODO(), productPrice)
	_, err = collection.InsertOne(context.TODO(), productPrice)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}
	//fmt.Println("Insert result: ", insertResult.InsertedID)


	//在商品表中也更新这个记录
	collection = models.Client.Collection("product")
	insertProduct := bson.M{"product_id": productPrice.ProductID} //

	// mongoDB往记录的数组添加数据的两个方法：$addToSet, $push
	// 两者的区别在于是否会判断重复数据，$push不会做此检查
	//pushToArray := bson.M{"$push": bson.M{"price_of_supplier": supplierPrice}}
	pushToArray := bson.M{"$addToSet": bson.M{"price_of_supplier": bson.M{"supplier": supplierPrice.Supplier, "price": supplierPrice.Price}}}
	_, err = collection.UpdateOne(context.TODO(), insertProduct, pushToArray)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}
	//fmt.Println("updateReult: ", updateResult.ModifiedCount)

	//在供应商表中也更新这个记录
	collection = models.Client.Collection("supplier")
	insertProduct = bson.M{"supplier_id": productPrice.SupplierID} //

	// mongoDB往记录的数组添加数据的两个方法：$addToSet, $push
	// 两者的区别在于是否会判断重复数据，$push不会做此检查
	//pushToArray := bson.M{"$push": bson.M{"price_of_supplier": supplierPrice}}
	pushToArray = bson.M{"$addToSet": bson.M{"supply_list": bson.M{"supply_list": productPrice.ProductID}}}
	_, err = collection.UpdateOne(context.TODO(), insertProduct, pushToArray)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}
	//fmt.Println("updateReult: ", updateResult.ModifiedCount)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Product price add to supply list succeeded",
	})
}

func ProductDetail(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	data, _ := ioutil.ReadAll(c.Request.Body)
	var p ProductService
	err := json.Unmarshal(data, &p)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	// 去supplier_product_price表中查找出所有product_id相等，且is_valid字段为true的记录
	var result []models.SupplierProductPrice
	collection := models.Client.Collection("supplier_product_price")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["product_id"] = p.ID
	filter["is_valid"] = true
	cur, err := collection.Find(context.TODO(), filter)

	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	// sup_price_list
	var supPriceList []int64

	for cur.Next(context.TODO()) {
		var r models.SupplierProductPrice
		if err := cur.Decode(&r); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "参数解释错误",
			})
			return
		}
		result = append(result, r)
		supPriceList = append(supPriceList, r.SupplierID)
	}

	// 给出不供应此商品的供应商列表
	// 使用sup_price和所有供应商id做一个差集
	var allSuppliersID []int64
	collection = models.Client.Collection("supplier")
	filter = bson.M{}
	filter["com_id"] = claims.ComId
	option := options.Find()
	option.Projection = bson.M{"supplier_id":1, "_id": 0}
	cur, err = collection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Can't get all suppliers: ", err)
		return
	}
	for cur.Next(context.TODO()) {
		var res models.Supplier
		if err := cur.Decode(&res); err != nil {
			fmt.Println("can't get supplier id: ", err)
			return
		}
		allSuppliersID = append(allSuppliersID, res.ID)
	}



	fmt.Println("all supplier id: ", allSuppliersID)
	fmt.Println("sup price: ", supPriceList)
	var unSuppliersID []int64

	for _, item := range allSuppliersID {
		exist := false
		for _, sub_item := range supPriceList {
			if item == sub_item {
				exist = true
				break
			}
		}
		if !exist {
			unSuppliersID = append(unSuppliersID, item)
		}
	}
	fmt.Println(unSuppliersID)

	if len(unSuppliersID) == 0 {
		responseData := make(map[string]interface{})
		responseData["supplier"] = result
		c.JSON(http.StatusOK, serializer.Response{
			Code: 200,
			Msg: "没有另外提供此商品的供应商",
			Data: responseData,
		})
	}


	var unSuppliers []models.Supplier
	collection = models.Client.Collection("supplier")
	filter = bson.M{}
	filter["com_id"] = claims.ComId
	filter["supplier_id"] = bson.M{"$in": unSuppliersID}
	cur, err = collection.Find(context.TODO(), filter)

	if err != nil {
		fmt.Println("Can't find un suppliers: ", err)
		return
	}

	for cur.Next(context.TODO()) {
		var res models.Supplier
		if err := cur.Decode(&res); err != nil {
			fmt.Println("Can't find supplier: ", err)
			return
		}
		unSuppliers = append(unSuppliers, res)
	}

	fmt.Println(unSuppliers)

	responseData := make(map[string]interface{})
	responseData["supplier"] = result
	responseData["un_supplier"] = unSuppliers
	//SmartPrint(product)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg: "Get product detail",
		Data: responseData,
	})
}

type SupplierListService struct {
	ProductList []int64 `json:"products"`
}

type SupplierPriceOfProduct struct {
	Supplier models.Supplier `json:"supplier"`
	Price float64 `json:"price"`
}

// 采购接口：点击采购后提供拥有这几个商品的供应商
func  SupplierListOfProducts (c *gin.Context) {

	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)


	//遍历供应商列表中supply_list这个字段，比对参数是否包含在这个数组中

	var  ss SupplierListService
	data, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Println("get data: ", string(data))
	_ = json.Unmarshal(data, &ss)
	//if err := c.ShouldBind(&ss); err != nil {
	//	fmt.Println("err while decoding array: ", err)
	//	c.JSON(http.StatusOK, serializer.Response{
	//		Code: -1,
	//		Msg:  "参数解释错误",
	//	})
	//	return
	//}

	fmt.Println("list: ", ss.ProductList)

	var list []models.Supplier
	collection := models.Client.Collection("supplier")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter = bson.M{"supply_list": bson.M{"$all": ss.ProductList}}
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var result models.Supplier
		err := cur.Decode(&result)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "参数解释错误",
			})
			return
		}
		list = append(list, result)
	}

	//collection = models.Client.Collection("supplier_product_price")
	//var supplierPriceOfProduct []models.SupplierProductPrice
	//for _, p_id := range ss.ProductList {
	//	for _, supplier := range list {
	//		var res models.SupplierProductPrice
	//		filter := bson.M{}
	//		filter["com_id"] = com.ComId
	//		filter["product_id"] = p_id
	//		filter["supplier_id"] = supplier.ID
	//		filter["is_valid"] = true
	//		err = collection.FindOne(context.TODO(), filter).Decode(&res)
	//		if err != nil {
	//			fmt.Println("Can't find supplier product price: ", err)
	//			return
	//		}
	//		supplierPriceOfProduct = append(supplierPriceOfProduct, res)
	//	}
	//}
	//
	//for _, item := range supplierPriceOfProduct {
	//	fmt.Println(item)
	//}

//	SmartPrint(list)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg: "Get supplier list of this product",
		Data: list,
	})
}

// 获取又拍云图片上传 policy  signature
func GetYpyunSign(c *gin.Context) {
	// 根据域名得到com_id
	// 根据域名得到com_id
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)
	fmt.Println("ComID: ", claims.ComId)

	type ReqYpyunSign struct {
		FileName string `json:"file_name" form:"file_name"`
		FileHash string `json:"file_hash" form:"file_hash"`
		FileSize string `json:"file_size" form:"file_size"`
	}

	var req ReqYpyunSign

	// 获取请求数据
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	SmartPrint(req)

	path, _ := util.GetYpyunPath(req.FileName)
	sign := util.GetYpyunSign(path,req.FileHash,req.FileSize)
	c.JSON(http.StatusOK,serializer.Response{
		Code:  200,
		Data:  sign,
		Msg:   "",
		Error: "",
	})
}

