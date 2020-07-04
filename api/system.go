package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"net/http"
)

// 系统设置方面的接口

type SystemConfig struct {
	ComID         int64 `json:"com_id" bson:"com_id"`
	ExpireDate    int64 `json:"expire_date" bson:"expire_date"`
	UnreadMessage int64 `json:"unread_message" bson:"unread_message"`
}

func GetExpireDate(c *gin.Context) {

	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Token error",
		})
		return
	}
	collection := models.Client.Collection("system_config")
	var sysConf SystemConfig
	err := collection.FindOne(context.TODO(), bson.D{{"com_id", claims.(*auth.Claims).ComId}}).Decode(&sysConf)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't get system config",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Get system config",
		Data: sysConf,
	})
}

func getExpireDateFromSystem(com_id int64) (int64, error) {
	collection := models.Client.Collection("system_config")
	var sysConf SystemConfig
	err := collection.FindOne(context.TODO(), bson.D{{"com_id", com_id}}).Decode(&sysConf)
	if err != nil {
		return 0, err
	}
	return sysConf.ExpireDate, nil
}

type SystemExpireDate struct {
	ExpireDate int64 `json:"expire_date"`
}

// 设置发货提醒时间
func SetSysExpireDate(c *gin.Context) {
	claims, _ := c.Get("claims")

	var sed SystemExpireDate

	if err := c.ShouldBindJSON(&sed); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	collection := models.Client.Collection("system_config")
	_, err := collection.UpdateOne(context.TODO(), bson.D{{"com_id", claims.(*auth.Claims).ComId}}, bson.M{
		"$set": bson.M{"expire_date": sed.ExpireDate}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Update user messages",
	})
}
