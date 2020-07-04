package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"net/http"
)


// 系统基本配置中新增配送方式，但是还没有使用
func AddDelivery(c *gin.Context) {
	claims, _ := c.Get("claims")

	var delivery models.Delivery
	if err := c.ShouldBindJSON(&delivery); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "params error",
		})
		return
	}

	delivery.ComId = claims.(*auth.Claims).ComId
	delivery.IsUsing = false

	collection := models.Client.Collection("delivery")
	insertResult, err := collection.InsertOne(context.TODO(), delivery)
	if err != nil {
		fmt.Println("Can't add new delivery")
		return
	}
	fmt.Println("insert delivery: ", insertResult.InsertedID)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Insert delivery Succeed",
	})
}
