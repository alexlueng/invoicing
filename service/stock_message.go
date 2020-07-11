package service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
	"time"
)


func StockMessage(product models.Product, comID int64) {

	collection := models.Client.Collection("users")
	filter := bson.M{}
	filter["com_id"] = comID

	// 不需要再找仓库管理员了
	//filter["position"] = "仓库管理员"

	messageType := 1
	collection = models.Client.Collection("message_type")
	var t models.MessageType
	err := collection.FindOne(context.TODO(), bson.D{{"id", messageType}}).Decode(&t)
	if err != nil {
		fmt.Println("Can't find message type: ", err)
		return
	}

	collection = models.Client.Collection("message")

	// 只生成一条提醒消息，让有权限的人查看即可
	var message models.Message
	message.ID = GetLastID("message")
	message.ComID = comID
	message.Type = int64(messageType)
	message.Title = "库存不足提醒"
	message.Message = fmt.Sprintf(t.Template, product.Product, product.MinAlert)
	message.IsRead = false
	message.NotifyWay = "web"
	message.CreateAt = time.Now().Unix()

	_, err = collection.InsertOne(context.TODO(), message)
	if err != nil {
		fmt.Println("Can't insert stock message: ", err)
		return
	}
	_ = SetLastID("message")
}

type Counts struct {
	NameField string
	Count     int64
}

func GetLastID(field_name string) int64 {
	var c Counts
	collection := models.Client.Collection("counters")
	err := collection.FindOne(context.TODO(), bson.D{{"name", field_name}}).Decode(&c)
	if err != nil {
		fmt.Println("can't get message ID")
		return 0
	}
	//collection.UpdateOne(context.TODO(), bson.M{"name": field_name}, bson.M{"$set": bson.M{"count": c.Count + 1}})
	//fmt.Printf("%s count: %d", field_name, c.Count)
	return c.Count + 1
}

func SetLastID(field_name string) error {
	collection := models.Client.Collection("counters")
	updateResult, err := collection.UpdateOne(context.TODO(), bson.D{{"name", field_name}}, bson.M{"$inc": bson.M{"count": 1}})
	if err != nil {
		return err
	}
	fmt.Println("Update result: ", updateResult.UpsertedID)
	return nil
}