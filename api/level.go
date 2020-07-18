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
// 客户关联客户等级 进而在计算商品售价时与客户等级挂钩
type LevelReq struct {
	LevelID         int64   `json:"level_id"`
	LevelName       string  `json:"level_name"`
	Discount        int64   `json:"discount"`
	LevelClass      int64   `json:"level_class"`
	AutoUpgrade     bool    `json:"auto_upgrade"`
	ConsumeAmount   float64 `json:"consume_amount"`
	ConsumeUsing    bool    `json:"consume_using"`
	RecommandPeople int64   `json:"recommand_people"`
	RecommandUsing  bool    `json:"recommand_using"`
}

func LevelList(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	collection := models.Client.Collection("level")
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

	var level models.Level
	level.ComID = claims.ComId
	level.LevelID = GetLastID("level")
	level.LevelName = req.LevelName
	level.Discount = float64(req.Discount)
	level.LevelClass = req.LevelClass
	level.AutoUpgrade = req.AutoUpgrade
	level.ConsumeAmount = req.ConsumeAmount
	level.ConsumeUsing = req.ConsumeUsing
	level.RecommandPeople = req.RecommandPeople
	level.RecommandUsing = req.RecommandUsing

	collection := models.Client.Collection("level")
	_, err := collection.InsertOne(context.TODO(), level)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Can't insert level",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Insert level succeed",
		Data: level,
	})

}

func UpdateLevel(c *gin.Context) {

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
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["level_id"] = req.LevelID
	_, err := collection.UpdateOne(context.TODO(), filter, bson.M{"$set": bson.M{"level_name": req.LevelName,
			"discount": req.Discount,
			"level_class": req.LevelClass,
			"auto_upgrade": req.AutoUpgrade,
			"consume_amount": req.ConsumeAmount,
			"consume_using": req.ConsumeUsing,
			"recommand_people": req.RecommandPeople,
			"recommand_using": req.RecommandUsing,}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Update level failed",
		})
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
