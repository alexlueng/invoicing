package api

import (
	"context"
	"encoding/json"
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

type ProductService struct {
	ID int64 `json:"product_id"`
}

// 供应商价格表
type ProductPrice struct {
	ProductID  int64   `json:"product_id" bson:"product_id"`
	Product    string  `json:"product" bson:"product"`
	SupplierID int64   `json:"supplier_id" bson:"supplier_id"`
	Supplier   string  `json:"supplier" bson:"supplier"`
	Price      float64 `json:"price" bson:"price"`
}

type SupplierListService struct {
	ProductList []int64 `json:"products"`
}

type SupplierPriceOfProduct struct {
	Supplier models.Supplier `json:"supplier"`
	Price    float64         `json:"price"`
}

type ProductReq struct {
	ProductID    int64      `json:"product_id"`
	Product      string     `json:"product"`
	Units        string     `json:"units"`
	DefaultPrice float64    `json:"default_price"`
	Category     int64      `json:"cat_id"`
	URLs         []ImageURL `json:"urls"`
	Tags         []string   `json:"tags"`      // 商品标签
	Preferred    bool       `json:"preferred"` // 是否优选
	Recommand    bool       `json:"recommand"` // 是否推荐
}

func AllProducts(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var (
		products []models.Product
		req      models.ProductReq
	)
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "参数解释错误",
		})
		return
	}

	req.Page, req.Size = SetDefaultPageAndSize(req.Page, req.Size)
	// 设置排序主键
	orderField := []string{"product_id", "com_id", "product"}
	option := SetPaginationAndOrder(req.OrdF, orderField, req.Ord, req.Page, req.Size)

	filter := bson.M{}

	if req.Product != "" {
		filter["product"] = bson.M{"$regex": req.Product}
	}

	filter["com_id"] = claims.ComId

	collection := models.Client.Collection("product")
	cur, err := collection.Find(context.TODO(), filter, option)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find products",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var result models.Product
		err := cur.Decode(&result)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't decode products",
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

	// 返回商品图片
	productImages := make(map[int64][]models.Image)
	collection = models.Client.Collection("image")
	for _, product := range products {

		cur, err := collection.Find(context.TODO(), bson.D{{"com_id", claims.ComId}, {"product_id", product.ProductID}})
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't find images",
			})
			return
		}
		for cur.Next(context.TODO()) {
			var image models.Image
			if err := cur.Decode(&image); err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: serializer.CodeError,
					Msg:  "Can't decode images",
				})
				return
			}
			productImages[product.ProductID] = append(productImages[product.ProductID], image)
		}
	}

	// 返回所有的供应商
	var productSuppliers []models.ProductSupplier
	cur, _ = models.Client.Collection("supplier").Find(context.TODO(), filter)
	for cur.Next(context.TODO()) {
		var result models.ProductSupplier
		if err := cur.Decode(&result); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't decode suppliers",
			})
			return
		}
		productSuppliers = append(productSuppliers, result)
	}

	// 返回查询到的总数，总页数
	resData := models.ResponseProductData{}
	resData.Products = products
	resData.ProductImages = productImages
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

// 商品应该和供应商联系在一起
// 获取供应商以及供应商的价格
// 在供应商表中添加对应的商品信息
func AddProduct(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req ProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	var result models.Product
	collection := models.Client.Collection("product")
	if !ENABLESAMEPRODUCT { // 不允许重名的情况，先查找数据库是否已经存在记录，如果有，则返回错误码－1
		filter := bson.M{}
		filter["com_id"] = claims.ComId
		filter["product"] = result.Product
		_ = collection.FindOne(context.TODO(), filter).Decode(&result)
		if result.Product != "" {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "该商品已经存在",
			})
			return
		}
	}
	var product models.Product
	product.ProductID = GetLastID("product")
	product.Product = req.Product
	product.ComID = claims.ComId
	product.Units = req.Units
	product.Tags = req.Tags
	product.CatID = req.Category
	product.DefaultPrice = req.DefaultPrice

	product.SupPrice = []int64{}
	product.CusPrice = []int64{}

	// 默认价格是default价格乘以公司的利润率
	//var company models.Company
	//_ = models.Client.Collection("company").FindOne(context.TODO(),bson.D{{"com_id", 1}}).Decode(&company)
	//default_margin := company.DefaultProfitMargin
	//product.DefaultPrice += product.DefaultPrice * float64(default_margin) / 100

	_, err := collection.InsertOne(context.TODO(), product)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "添加商品失败",
		})
		return
	}

	// 将图片保存到图片表中
	collection = models.Client.Collection("image")
	for _, url := range req.URLs {
		image := models.Image{
			ComID:     claims.ComId,
			ProductID: product.ProductID,
			ImageID:   GetLastID("image"),
			LocalPath: url.LocalURL,
			CloudPath: url.CloudURL,
			IsDelete:  false,
		}
		_, err := collection.InsertOne(context.TODO(), image)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Insert image error",
			})
			return
		}
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
	_, err = collection.InsertOne(context.TODO(), cusProPrice)

	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "添加商品默认价格失败",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Product create succeeded",
		Data: product.ProductID,
	})
}

// 更新商品需要修改三张表：product customer_product_price supplier_product_price
func UpdateProduct(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req ProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	var updateProduct models.Product

	updateProduct.ProductID = req.ProductID
	updateProduct.Product = req.Product
	updateProduct.DefaultPrice = req.DefaultPrice
	updateProduct.Units = req.Units
	updateProduct.Tags = req.Tags
	updateProduct.CatID = req.Category

	// 更新的条件：更改的时候如果有同名的记录，则要判断是否有与要修改的记录的customer_id相等,如果有不相等的，则返回
	// 如果只有相等的customer_id, 则允许修改
	filter := bson.M{}

	filter["com_id"] = claims.ComId
	filter["product"] = updateProduct.Product
	collection := models.Client.Collection("product")

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "添加商品失败",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var tempRes models.Product
		err := cur.Decode(&tempRes)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "decode product error",
			})
			return
		}
		if tempRes.ProductID != updateProduct.ProductID {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "要修改的商品已经存在",
			})
			return
		}
	}

	var oldRecord models.Product
	err = collection.FindOne(context.TODO(), filter).Decode(&oldRecord)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "decode product error",
		})
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
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "decode product error",
			})
			return
		}

		newRecord := oldRecord
		newRecord.Price = updateProduct.DefaultPrice

		// 将旧记录设为false
		_, err = cusProPriceCollects.UpdateOne(context.TODO(), filter, bson.M{
			"$set": bson.M{"is_valid": false}})
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't find old record",
			})
			return
		}

		// 加入一条新记录
		_, err = cusProPriceCollects.InsertOne(context.TODO(), newRecord)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't insert new record",
			})
			return
		}
	}

	filter = bson.M{}
	filter["com_id"] = claims.ComId
	filter["product_id"] = updateProduct.ProductID
	// 更新记录
	_, err = collection.UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"product": updateProduct.Product,
			"num":           updateProduct.Num,
			"units":         updateProduct.Units,
			"urls":          updateProduct.URLS,
			"default_price": updateProduct.DefaultPrice,
			"tags":          updateProduct.Tags,
			"cat_id":        updateProduct.CatID,}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "更新失败",
		})
		return
	}

	// update customer_product_price
	collection = models.Client.Collection("customer_product_price")
	_, err = collection.UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{
			"product_name": updateProduct.Product,}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't update customer price",
		})
		return
	}

	// update supplier_product_price
	collection = models.Client.Collection("supplier_product_price")
	_, err = collection.UpdateMany(context.TODO(), filter, bson.M{
		"$set": bson.M{
			"product": updateProduct.Product,}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't update supplier price",
		})
		return
	}

	// update images
	collection = models.Client.Collection("image")
	for _, url := range req.URLs {
		if url.ProductID == 0 {
			image := models.Image{
				ComID:     claims.ComId,
				ProductID: req.ProductID,
				ImageID:   GetLastID("image"),
				LocalPath: url.LocalURL,
				CloudPath: url.CloudURL,
				IsDelete:  false,
			}
			_, err := collection.InsertOne(context.TODO(), image)
			if err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: serializer.CodeError,
					Msg:  "Can't insert image",
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Product update succeeded",
	})
}

// 删除商品要更新四个表：
// 1, product
// 2, customer_product_price和supplier_product_price表中将此商品的is_valid字段置为false
// 3，supplier表中supply_list中有这个商品id的值去掉
// 4, 将上传的图片删掉
func DeleteProduct(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var (
		d       ProductService
		product models.Product
	)
	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &d)

	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["product_id"] = d.ID
	collection := models.Client.Collection("product")

	err := collection.FindOne(context.TODO(), filter).Decode(&product)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find product",
		})
		return
	}

	//删除商品图片（云上的以及本地的）
	collection = models.Client.Collection("image")
	filter = bson.M{}
	filter["com_id"] = claims.ComId
	filter["product_id"] = product.ProductID
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find product",
		})
		return
	}

	for cur.Next(context.TODO()) {
		var res models.Image
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't decode image",
			})
			return
		}
		if err := deleteImage(res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't delete image",
			})
			return
		}
	}

	_, err = collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
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
	_, err = collection.UpdateMany(context.TODO(), filter, bson.M{
		"$set": bson.M{
			"is_valid": false,}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "更新商品售价表失败",
		})
		return
	}

	// update supplier_product_price
	collection = models.Client.Collection("supplier_product_price")
	_, err = collection.UpdateMany(context.TODO(), filter, bson.M{
		"$set": bson.M{
			"is_valid": false,}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "更新商品进价表失败",
		})
		return
	}

	// TODO: $pull删除子数组中的元素

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Product delete succeeded",
	})
}

// 添加此商品的供应商价格
// 把商品id保存到供应商列表的字段中，以便快速查询此供应商能供应的产品
func AddPrice(c *gin.Context) {

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

	var productPrice ProductPrice

	err = json.Unmarshal(data, &productPrice)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}
	//把这个记录插入供应商商品价格表
	collection := models.Client.Collection("supplier_product_price")
	_, err = collection.InsertOne(context.TODO(), productPrice)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	//在商品表中也更新这个记录
	collection = models.Client.Collection("product")
	insertProduct := bson.M{"product_id": productPrice.ProductID}

	pushToArray := bson.M{"$addToSet": bson.M{"price_of_supplier": bson.M{"supplier": supplierPrice.Supplier, "price": supplierPrice.Price}}}
	_, err = collection.UpdateOne(context.TODO(), insertProduct, pushToArray)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

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
			Code: serializer.CodeError,
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
			Code: serializer.CodeError,
			Msg:  "参数解释错误",
		})
		return
	}

	var supPriceList []int64

	for cur.Next(context.TODO()) {
		var r models.SupplierProductPrice
		if err := cur.Decode(&r); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
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
	option.Projection = bson.M{"supplier_id": 1, "_id": 0}
	cur, err = collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't get all suppliers",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var res models.Supplier
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't decode supplier",
			})
			return
		}
		allSuppliersID = append(allSuppliersID, res.ID)
	}

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

	if len(unSuppliersID) == 0 {
		responseData := make(map[string]interface{})
		responseData["supplier"] = result
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeSuccess,
			Msg:  "没有另外提供此商品的供应商",
			Data: responseData,
		})
		return
	}

	var unSuppliers []models.Supplier
	collection = models.Client.Collection("supplier")
	filter = bson.M{}
	filter["com_id"] = claims.ComId
	filter["supplier_id"] = bson.M{"$in": unSuppliersID}
	cur, err = collection.Find(context.TODO(), filter)

	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	for cur.Next(context.TODO()) {
		var res models.Supplier
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't find supplier",
			})
			return
		}
		unSuppliers = append(unSuppliers, res)
	}

	responseData := make(map[string]interface{})
	responseData["supplier"] = result
	responseData["un_supplier"] = unSuppliers
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Get product detail",
		Data: responseData,
	})
}

// 采购接口：点击采购后提供拥有这几个商品的供应商
func SupplierListOfProducts(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	//遍历供应商列表中supply_list这个字段，比对参数是否包含在这个数组中

	var ss SupplierListService
	data, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(data, &ss)

	var list []models.Supplier
	collection := models.Client.Collection("supplier")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter = bson.M{"supply_list": bson.M{"$all": ss.ProductList}}
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find suppliers",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var result models.Supplier
		err := cur.Decode(&result)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't decode supplier",
			})
			return
		}
		list = append(list, result)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get supplier list of this product",
		Data: list,
	})
}

// 获取又拍云图片上传 policy  signature
func GetYpyunSign(c *gin.Context) {

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
			Code: serializer.CodeError,
			Msg:  "参数解释错误",
		})
		return
	}

	path, _ := util.GetYpyunPath(req.FileName)
	sign := util.GetYpyunSign(path, req.FileHash, req.FileSize)
	c.JSON(http.StatusOK, serializer.Response{
		Code:  200,
		Data:  sign,
		Msg:   "",
		Error: "",
	})
}

// 设置是否优选商品
func PreferredProduct(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req ProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	collection := models.Client.Collection("product")

	var product models.Product

	err := collection.FindOne(context.TODO(), bson.D{{"com_id", claims.ComId}, {"product_id", req.ProductID}}).Decode(&product)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "find product error",
		})
		return
	}

	_, err = collection.UpdateOne(context.TODO(), bson.D{{"com_id", claims.ComId}, {"product_id", req.ProductID}}, bson.M{
		"$set": bson.M{"preferred": req.Preferred}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "update preferred product error",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "update preferred succeed",
		Data: product,
	})
}

// 设置是否推荐商品
func RecommandProduct(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req ProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	collection := models.Client.Collection("product")

	var product models.Product

	err := collection.FindOne(context.TODO(), bson.D{{"com_id", claims.ComId}, {"product_id", req.ProductID}}).Decode(&product)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "find product error",
		})
		return
	}

	_, err = collection.UpdateOne(context.TODO(), bson.D{{"com_id", claims.ComId}, {"product_id", req.ProductID}}, bson.M{
		"$set": bson.M{"recommand": req.Recommand}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "update recommand product error",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "update recommand succeed",
		Data: product,
	})
}
