package service

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"jxc/models"
)

// 根据域名获取公司id
func FindDomain(domain string) (models.DomainData, error) {
	domainInfo := models.DomainData{}
	collection := models.Client.Collection("domain")
	err := collection.FindOne(context.TODO(), bson.D{{"domain", domain}}).Decode(&domainInfo)
	if err != nil {
		// 这个域名还没注册，找不到对应的公司
		return domainInfo, errors.New("")
	} else {
		return domainInfo, nil
	}
}

// 获取公司下所有的域名
func AllDomain(com_id int64) ([]models.DomainData, error) {

	domain := models.DomainData{}
	domainArrayEmpty := []models.DomainData{}

	collection := models.Client.Collection("domain")

	var cursor *mongo.Cursor

	//一次查询多条数据
	// 查询com_id = com_id
	// 限制取100条
	// createtime从大到小排序的数据
	cursor, err := collection.Find(context.TODO(), bson.M{"comid": bson.M{"$eq": com_id}}, options.Find().SetLimit(100), options.Find().SetSort(bson.M{"com_id": -1}))
	if err != nil {
		checkErr(err)
	}
	if err := cursor.Err(); err != nil {
		checkErr(err)
	}
	for cursor.Next(context.Background()) {
		if err := cursor.Decode(&domain); err != nil {
			checkErr(err)
		}
		domainArrayEmpty = append(domainArrayEmpty, domain)
	}

	return domainArrayEmpty, nil
}

// 域名是否被别的公司注册，需要提供公司id、域名
func checkUnique(com_id string, domains []models.DomainData) {
	collection := models.Client.Collection("domain")

	domain := models.DomainData{}

	for _, val := range domains {
		if err := collection.FindOne(context.TODO(), bson.D{{"domain", val.Domain}}).Decode(&domain); err != nil {
			checkErr(err)
		}
		//if domain.ComId != com_id {
		//	// 域名已被注册，且是另一家公司
		//	return
		//}
	}

}

func checkErr(err error) {
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("没有查到数据")
			return
		} else {
			fmt.Println(err)
			return
		}

	}
}
