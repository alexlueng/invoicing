package wxapp

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
	"jxc/serializer"
	"net/http"
	"strings"
)

type DetailService struct {
	ProductID int64 `json:"product_id"`
}

// 商品详情
func ProductDetail(c *gin.Context) {

	//token := c.GetHeader("Access-Token")
	//claims, _ := auth.ParseToken(token)
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

	var detailSrv DetailService
	if err := c.ShouldBindJSON(&detailSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "参数解释错误",
		})
		return
	}

	product, err := models.GetProductByID(com.ComId, detailSrv.ProductID)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find product",
		})
		return
	}

	detail, err := models.GetProductDetailByID(com.ComId, detailSrv.ProductID)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find product detail",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeError,
		Msg:  "Product detail",
		Data: map[string]interface{}{
			"product": product,
			"detail":  detail,
		},
	})
}

type ProductCategoryService struct {
	CategoryID int64 `json:"cat_id"`
}

// 根据分类id获取商品列表
func ProductListByCategoryID(c *gin.Context) {

	var srv ProductCategoryService
	if err := c.ShouldBindJSON(&srv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	var products []models.Product
	var productIDs []int64
	collection := models.Client.Collection("product")

	cur, err := collection.Find(context.TODO(), bson.D{{"cat_id", srv.CategoryID}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Find product error",
		})
		return
	}

	for cur.Next(context.TODO()) {
		var res models.Product
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Decode product error",
			})
			return
		}
		products = append(products, res)
		productIDs = append(productIDs, res.ProductID)
	}

	// 商品图片
	var images []models.Image
	if len(productIDs) > 0 {

		collection = models.Client.Collection("image")
		filter := bson.M{"product_id" : bson.M{"$in" : productIDs}}
		cur, err := collection.Find(context.TODO(), filter)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Find product image error",
			})
			return
		}
		for cur.Next(context.TODO()) {
			var res models.Image
			if err := cur.Decode(&res); err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: serializer.CodeError,
					Msg:  "Decode product error",
				})
				return
			}
			images = append(images, res)
		}
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Product list",
		Data: map[string]interface{}{
			"products" : products,
			"images" : images,
		},
	})
}