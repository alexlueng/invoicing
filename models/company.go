package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

//

// 公司管理表数据结构
type CompanyData struct {
	ComId          int64  `json:"com_id" bson:"com_id"`
	ComName        string `json:"com_name" bson:"com_name"`
	ExpirationDate string `json:"expiration_date" bson:"expiration_date"`
	//Delivery       []DeliveryData `json:"delivery" bson:"delivery"`
	Units     string   `json:"units" bson:"units"`
	Payment   []string `json:"payment" bson:"payment"`     //结算方式
	Module    string   `json:"module" bson:"module"`       //平台名称
	Developer string   `json:"developer" bson:"developer"` //开发名称
	Admin     string   `json:"admin" bson:"admin"`
	Telephone string   `json:"telephone" bson:"telephone"`
}

// 公司表数据结构
type Company struct {
	ComId               int64       `json:"com_id" bson:"com_id"`
	ComName             string      `json:"com_name" bson:"com_name"`
	ExpireAt            int64       `json:"expire_at" bson:"expire_at"`                         // 到期时间
	CreateAt            int64       `json:"create_at" bson:"create_at"`                         // 创建时间
	Units               interface{} `json:"units" bson:"units"`                                 //计量单位
	Payment             interface{} `json:"payment" bson:"payment"`                             //结算方式
	Module              string      `json:"module" bson:"module"`                               //平台名称
	Developer           string      `json:"developer" bson:"developer"`                         //开发名称
	Position            interface{} `bson:"position" json:"position"`                           //职务
	DefaultProfitMargin float64     `json:"default_profit_margin" bson:"default_profit_margin"` //默认利润率
	Admin               string      `json:"admin" bson:"admin"`
	Telephone           string      `json:"phone" bson:"phone"`
	Password            string      `json:"password" bson:"password"`
}

func getCompanyCollection() *mongo.Collection {
	return Client.Collection("company")
}

func (c *Company) Add() error {
	_, err := getCompanyCollection().InsertOne(context.TODO(), c)
	return err
}

func SelectCompanyByComID(comID int64) (*Company, error) {
	var c Company
	err := getCompanyCollection().FindOne(context.TODO(), bson.M{"com_id": comID}).Decode(&c)
	if err != nil{
		return nil, err
	}
	return &c, nil
}

func UpdateCompanyByComID(comID int64, updateData bson.M) error {
	_, err := getCompanyCollection().UpdateOne(context.TODO(), bson.M{"com_id": comID}, bson.M{
		"$set": updateData,
	})
	return err
}

func UpdateAdminPwdByTelPhone(telephone, password string) error {
	_, err := getCompanyCollection().UpdateMany(context.TODO(), bson.D{{"telephone", telephone}}, bson.M{"$set": bson.M{"password": password}})
	return err
}

func UpdateCompanyExpireTime(comID, expireTime int64) error {
	_, err := getCompanyCollection().UpdateOne(context.TODO(), bson.D{{"com_id", comID}}, bson.M{"$set": bson.M{"expire_at": expireTime}})
	return err
}


