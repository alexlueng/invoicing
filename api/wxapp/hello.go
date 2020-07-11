package wxapp

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
	"jxc/serializer"
	"jxc/util"
	"net/http"
	"strings"
)

// 微信小程序商城api
// 商品，分类等列表的返回

func Hello(c *gin.Context) {
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "hello world",
	})
}

type WxLogin struct {
	JsCode string `json:"jscode"`
}

func Login(c *gin.Context) {
	//jscode := c.Query("jsCode")
	var jscode WxLogin
	if err := c.ShouldBindJSON(&jscode); err != nil {
		fmt.Println("Miniapp error: ", err)
		return
	}

	fmt.Println("jsCode: ", jscode.JsCode)
}


// 首页优选商品列表
func PreferredProductList(c *gin.Context) {

	// 获取请求的域名，可以得知所属公司
	domain := c.Request.Header.Get("Origin")
	domain = strings.Split(domain, ":")[1]
	com, err := models.GetComIDAndModuleByDomain(domain[len("//"):])
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}
	util.Log().Info("get com id: ", com.ComId)

	var products []models.Product
	collection := models.Client.Collection("product")
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["preferred"] = true
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Find preferred product error",
		})
		return
	}

	for cur.Next(context.TODO()) {
		var res models.Product
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Decode preferred product error",
			})
			return
		}
		products = append(products, res)
	}

	// 返回商品图片
	productImages := make(map[int64][]models.Image)
	collection = models.Client.Collection("image")
	for _, product := range products {

		cur, err := collection.Find(context.TODO(), bson.D{{"com_id", com.ComId}, {"product_id", product.ProductID}})
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

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Product list",
		Data: map[string]interface{}{
			"products" : products,
			"images" : productImages,
		},
	})
}

// 首页推荐商品列表
func RecommandProductList(c *gin.Context) {

	// 获取请求的域名，可以得知所属公司
	domain := c.Request.Header.Get("Origin")
	domain = strings.Split(domain, ":")[1]
	com, err := models.GetComIDAndModuleByDomain(domain[len("//"):])
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	util.Log().Info("get com id: ", com.ComId)

	var products []models.Product
	collection := models.Client.Collection("product")
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["recommand"] = true
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Find recommand product error",
		})
		return
	}

	for cur.Next(context.TODO()) {
		var res models.Product
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Decode recommand product error",
			})
			return
		}
		products = append(products, res)
	}

	// 返回商品图片
	productImages := make(map[int64][]models.Image)
	collection = models.Client.Collection("image")
	for _, product := range products {

		cur, err := collection.Find(context.TODO(), bson.D{{"com_id", com.ComId}, {"product_id", product.ProductID}})
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

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Product list",
		Data: map[string]interface{}{
			"products" : products,
			"images" : productImages,
		},
	})
}

func CategoryList(c *gin.Context) {
	var categories []models.Category
	collection := models.Client.Collection("category")

	cur, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Find category error",
		})
		return
	}

	for cur.Next(context.TODO()) {
		var res models.Category
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Decode category error",
			})
			return
		}
		categories = append(categories, res)
	}

	var images []models.CategoryImage
	collection = models.Client.Collection("category_image")
	cur, err = collection.Find(context.TODO(), bson.D{})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find category images",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var res models.CategoryImage
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "decode image error",
			})
			return
		}
		images = append(images, res)
	}


	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Category list",
		Data: map[string]interface{}{
			"categories" : categories,
			"images" : images,
		},
	})
}
