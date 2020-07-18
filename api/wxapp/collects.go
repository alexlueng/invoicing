package wxapp

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/api"
	"jxc/auth"
	"jxc/models"
	"jxc/models/wxapp"
	"jxc/serializer"
	"net/http"
)

// 用户关注商品列表接口

type UserCollectionService struct {
	CustomerID int64 `json:"customer_id"`
	ProductID  int64 `json:"product_id"`
	CollectID  int64 `json:"collect_id"`
}

// 加入收藏
func AddToCollection(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var userCollectionSrv UserCollectionService
	if err := c.ShouldBindJSON(&userCollectionSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Params error",
		})
		return
	}

	userCollection := wxapp.UserCollection{
		ComID:      claims.ComId,
		CustomerID: userCollectionSrv.CustomerID,
		ProductID:  userCollectionSrv.ProductID,
		Status:     1,
		CollectID:  api.GetLastID("user_collection"),
	}

	collection := models.Client.Collection("user_collection")
	_, err := collection.InsertOne(context.TODO(), userCollection)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't add user collection",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeError,
		Msg:  "Add user collection",
	})
}

// 移出收藏夹
func RemoveUserCollection(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var userCollectionSrv UserCollectionService
	if err := c.ShouldBindJSON(&userCollectionSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Params error",
		})
		return
	}

	collection := models.Client.Collection("user_collection")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["collect_id"] = userCollectionSrv.CollectID
	_, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Delete collect error",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Delete user collect",
	})
}

// 用户收藏列表
func UserCollection(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var userCollectionSrv UserCollectionService
	if err := c.ShouldBindJSON(&userCollectionSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Params error",
		})
		return
	}

	collection := models.Client.Collection("user_collection")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["customer_id"] = userCollectionSrv.CustomerID
	filter["status"] = 1 // 产品为可用状态
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Get collect error",
		})
		return
	}
	//var collects []wxapp.UserCollection
	var productIDs []int64
	for cur.Next(context.TODO()) {
		var res wxapp.UserCollection
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Decode collect error",
			})
			return
		}
		productIDs = append(productIDs, res.ProductID)
	}

	if len(productIDs) < 1 {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeSuccess,
			Msg:  "No user collections",
			Data: nil,
		})
		return
	}

	var products []models.Product
	collection = models.Client.Collection("product")
	filter = bson.M{}
	filter["com_id"] = claims.ComId
	filter["product_id"] = bson.M{"$in": productIDs}
	cur, err = collection.Find(context.TODO(), filter)
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
		Code: serializer.CodeSuccess,
		Msg:  "User collection products",
		Data: products,
	})
}

