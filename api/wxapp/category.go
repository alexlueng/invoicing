package wxapp

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"net/http"
)

// 根据商品分类获取商品列表

type CatProductService struct {
	CatID int64 `json:"cat_id"`
}

func CatProducts(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var catProductSrv CatProductService
	if err := c.ShouldBindJSON(&catProductSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	var products []models.Product
	collection := models.Client.Collection("product")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["cat_id"] = catProductSrv.CatID
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find product",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var res models.Product
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't decode product",
			})
			return
		}
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Get products",
		Data: products,
	})


}
