package wxapp

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/api"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"net/http"
)

type UserAddressService struct {
	CustomerID int64 `json:"customer_id" bson:"customer_id"`
	AddressIDs []int64 `json:"address_ids" bson:"address_ids"`
}

// 用户收货地址管理

func ListUserAddress(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var userAddrSrv UserAddressService
	if err := c.ShouldBindJSON(&userAddrSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Params error",
		})
		return
	}

	collection := models.Client.Collection("address")
	var addresses []models.Address
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["customer_id"] = userAddrSrv.CustomerID
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "user address error",
		})
		return
	}
	for cur.Next(context.TODO()) {
		var res models.Address
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "decode address error",
			})
			return
		}
		addresses = append(addresses, res)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "User address",
		Data: addresses,
	})
}

func AddUserAddress(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var address models.Address
	if err := c.ShouldBindJSON(&address); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	address.ComID = claims.ComId
	address.AddressID = api.GetLastID("address")

	collection := models.Client.Collection("address")
	_, err := collection.InsertOne(context.TODO(), address)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Create user address error",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Create user address",
	})
}

// 可以同时删除多个商品
func DeleteUserAddress(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var userAddrSrv UserAddressService
	if err := c.ShouldBindJSON(&userAddrSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	collection := models.Client.Collection("address")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["customer_id"] = userAddrSrv.CustomerID
	if len(userAddrSrv.AddressIDs) < 0 {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "No address to delete",
		})
		return
	}
	filter["address_id"] = bson.M{"$in" : userAddrSrv.AddressIDs}
	_, err := collection.DeleteMany(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Delete user address error",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Delete user address",
	})
}

func UpdateAddress(c *gin.Context) {

	var address models.Address
	if err := c.ShouldBindJSON(&address); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "params error",
		})
		return
	}

	collection := models.Client.Collection("address")
	filter := bson.M{}
	filter["com_id"] = address.ComID
	filter["address_id"] = address.AddressID
	filter["customer_id"] = address.CustomerID
	_, err := collection.UpdateOne(context.TODO(), filter, bson.M{"$set" : address})

	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Update user address error",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Update user address",
	})
}