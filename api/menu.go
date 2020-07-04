package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
	"jxc/serializer"
	"net/http"
)

type MenuConfig struct {
	MenuID   int64  `json:"menu_id" bson:"menu_id"`
	MenuName string `json:"menu_name" bson:"menu_name"`
}

// 设置产品管理菜单
func GetProductMenuName(c *gin.Context) {
	var menu MenuConfig
	collection := models.Client.Collection("menu_config")
	err := collection.FindOne(context.TODO(), bson.D{{"menu_id", 1}}).Decode(&menu)
	if err != nil {
		fmt.Println("Can't get menu name: ", err)
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Hello",
		Data: menu,
	})
}

func SetProductMenuName(c *gin.Context) {
	var menu MenuConfig
	if err := c.ShouldBindJSON(&menu); err != nil {

		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Params error",
		})
		return
	}
	collection := models.Client.Collection("menu_config")
	updateResult, err := collection.UpdateOne(context.TODO(), bson.D{{"menu_id", menu.MenuID}}, bson.M{
		"$set": bson.M{"menu_name": menu.MenuName}})
	if err != nil {
		fmt.Println("Can't update menu name: ", err)
		return
	}
	fmt.Println("Update result: ", updateResult.UpsertedID)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Update Menu Succeed",
	})
}

