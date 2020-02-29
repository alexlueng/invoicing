package service

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
)

// 公司信息更新规则
type CompanyService struct {
	// 配送方式 可以为空
	// 商品量词 可以为空 最长不超过35
	// 结算方式 可以为空
	// 公司名 可以为空，最长不超过35
	// 开发商名称 可以为空，最长不超过35
	// 绑定域名 不能为空 最少要有一个

	Delivery  []string `from:"delivery" json:"delivery"`
	Units     string   `from:"units" json:"units" binding:"max=35"`
	Payment   string   `from:"payment" json:"payment"`
	ComName   string   `from:"com_name" json:"com_name" binding:"max=35"`
	Developer string   `from:"developer" json:"developer" binding:"max=35"`
	Domains   []string `from:"domains" json:"domains" binding:"min=1"`
}

// 查找到对应的公司
func FindCompany(com_id int64) (models.CompanyData, error) {
	company := models.CompanyData{}
	collection := models.Client.Collection("company")
	err := collection.FindOne(context.TODO(), bson.D{{"com_id", com_id}}).Decode(&company)
	if err != nil {
		// com_id错误，没有找到对应的数据
		return company, errors.New("")
	}
	return company, nil
}

// 通过用户信息查找对应的公司
func UidFindCompany(user_id string) (models.CompanyData, error) {
	company := models.CompanyData{}
	collection := models.Client.Collection("company")
	err := collection.FindOne(context.TODO(), bson.D{{"user_id", user_id}}).Decode(&company)
	if err != nil {
		// com_id错误，没有找到对应的数据
		return company, errors.New("")
	}
	return company, nil
}

// 更新公司
func UpdateCompany(com_id string, com models.CompanyData) {
	collection := models.Client.Collection("company")

	if _, err := collection.UpdateOne(context.TODO(), bson.M{"com_id": com_id}, bson.M{"$set": com}); err != nil {
		checkErr(err)
	}
	//fmt.Printf("UpdateOne的数据:%d\n", updateRes)

}

// 查找费送方式
func FindOneDelivery(deliveryId int64, comId int64) (*models.Delivery, error) {
	collection := models.Client.Collection("company")
	var delivery models.Delivery
	filter := bson.M{}
	filter["comid"] = comId
	filter["delivery_id"] = deliveryId
	err := collection.FindOne(context.TODO(), filter).Decode(&delivery)
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}
