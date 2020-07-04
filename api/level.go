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

// 客户等级
type LevelReq struct {
	Levels []models.Level `json:"levels"`
}

func LevelList(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	collection := models.Client.Collection("livel")
	filter := bson.M{}
	filter["com_id"] = claims.ComId

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't find customer level",
		})
		return
	}
	var result []models.Level
	for cur.Next(context.TODO()) {
		var res models.Level
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't decode level",
			})
			return
		}
		result = append(result, res)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Customer level",
		Data: result,
	})
}


func AddLevel(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var req LevelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Params error",
		})
		return
	}

	collection := models.Client.Collection("level")

	for _, level := range req.Levels {
		level.ComID = claims.ComId
		level.LevelID = getLastID("level")
		setLastID("level")
		_, err := collection.InsertOne(context.TODO(), level)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't decode level",
			})
			return
		}
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Insert level succeed",
	})

}

func UpdateLevel(c *gin.Context) {

	var req LevelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Params error",
		})
		return
	}

	collection := models.Client.Collection("level")
	for _, level := range req.Levels {
		_, err := collection.UpdateOne(context.TODO(), bson.D{{"level_id", level.LevelID}}, bson.M{"$set" : level})
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Params error",
			})
			return
		}
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Update level succeed",
	})
}

func DeleteLevel(c *gin.Context) {

	var level models.Level
	if err := c.ShouldBindJSON(&level); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Params error",
		})
		return
	}

	collection := models.Client.Collection("level")
	_, err := collection.DeleteOne(context.TODO(), bson.D{{"level_id", level.LevelID}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Delete level error",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Update level succeed",
	})
}