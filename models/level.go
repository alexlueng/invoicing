package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// 客户等级表
type Level struct {
	ComID           int64   `json:"com_id" bson:"com_id"`
	LevelID         int64   `json:"level_id" bson:"level_id"`
	LevelClass      int64   `json:"level_class" bson:"level_class"` // 客户等级
	LevelName       string  `json:"level_name" bson:"level_name"`
	Discount        float64 `json:"discount" bson:"discount"`
	AutoUpgrade     bool    `json:"auto_upgrade" bson:"auto_upgrade"`
	ConsumeAmount   float64 `json:"consume_amount" bson:"consume_amount"`     // 消费总额
	ConsumeUsing    bool    `json:"consume_using" bson:"consume_using"`       // 是否使用
	RecommandPeople int64   `json:"recommand_people" bson:"recommand_people"` // 推荐人数
	RecommandUsing  bool    `json:"recommand_using" bson:"recommand_using"`   // 是否使用
}

func getLevelCollection() *mongo.Collection {
	return Client.Collection("level")
}

type LevelResult struct {
	Level []Level `json:"level"`
}

func SelectLevelByComID(comID int64) (*LevelResult, error) {
	cur,err := getLevelCollection().Find(context.TODO(), bson.M{"com_id": comID})
	if err != nil {
		return nil, err
	}

	var rs = new(LevelResult)
	for cur.Next(context.TODO()) {
		var level Level
		err := cur.Decode(&level)
		if err != nil {
			return nil, err
		}
		rs.Level = append(rs.Level, level)
	}
	return rs, nil
}


