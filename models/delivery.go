package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// 配送方式数据格式
type Delivery struct {
	DeliveryId     int64  `json:"delivery_id" bson:"delivery_id"`        //配送方式id 这个id由前端生成
	ComId          int64  `json:"com_id" bson:"comid"`                   //公司id
	DeliveryCom    string `json:"delivery_com" bson:"deliverycom"`       // 配送公司
	DeliveryPerson string `json:"delivery_person" bson:"deliveryperson"` // 配送员
	Phone          string `json:"phone" bson:"phone"`                    // 配送员电话
	Config         string `json:"config" bson:"config"`                  // 配置参数
	IsUsing        bool   `json:"is_using" bson:"is_using"`              // 是否启用
}

type DeliveryResult struct {
	Delivery []Delivery `json:"delivery"`
}

func getDeliveryCollection() *mongo.Collection {
	return Client.Collection("delivery")
}

func SelectDeliveryByComID(comID int64)(*DeliveryResult, error) {
	cur, err := getDeliveryCollection().Find(context.TODO(), bson.M{"comid": comID})
	if err != nil {
		return nil , err
	}
	var rs = new(DeliveryResult)
	for cur.Next(context.TODO()){
		var delivery Delivery
		err = cur.Decode(&delivery)
		rs.Delivery = append(rs.Delivery, delivery)
	}
	return rs, nil
}

func UpdateDeliveryIsUsingFlag(comID int64, Delivery []int64, isUsing bool) error {
	_, err := getDeliveryCollection().UpdateMany(context.TODO(),
		bson.M{"delivery_id": bson.M{"$in": Delivery}, "comid": comID},
		bson.M{"$set": bson.M{"is_using": isUsing}})
	return err
}


