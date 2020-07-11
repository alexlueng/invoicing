package wxapp

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
	"jxc/serializer"
	"net/http"
)

func SeckillProductList(c *gin.Context) {
	var products []models.Product
	collection := models.Client.Collection("product")

	cur, err := collection.Find(context.TODO(), bson.D{})
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
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeError,
		Msg:  "Product list",
		Data: products,
	})
}
