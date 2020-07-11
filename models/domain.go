package models

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"jxc/util"
)

// 域名表
type Domain struct {
	ComId    int64  `bson:"comid" json:"comid"`       // 公司id
	Domain   string `bson:"domain" json:"domain"`     //
	ModuleId int64  `bson:"moduleid" json:"moduleid"` //
	Status   bool   `bson:"status" json:"status"`     // 域名可用状态，false情况下无法登录
}

// 域名表
type DomainData struct {
	ComId    int64  `bson:"comid" json:"comid"`       // 公司id
	Domain   string `bson:"domain" json:"domain"`     //
	ModuleId int64  `bson:"moduleid" json:"moduleid"` //
	Status   bool   `bson:"status" json:"status"`     // 域名可用状态，false情况下无法登录
}

func getDomainCollection() *mongo.Collection {
	return Client.Collection("domain")
}

func (d *Domain) Add() error {
	_, err := getDomainCollection().InsertOne(context.TODO(), d)
	if err != nil {
		util.Log().Error("add domain error, err:", err)
	}
	return err
}

type DomainResult struct {
	Domain []Domain `json:"domain"`
}

//根据公司查找域名
func SelectDomainByComID(comID int64) (*DomainResult, error) {
	cur, err := getDomainCollection().Find(context.TODO(), bson.D{{"comid", comID}})
	if err != nil {
		return nil, err
	}
	var rs = new(DomainResult)
	for cur.Next(context.TODO()) {
		var d Domain
		err = cur.Decode(&d)
		if err != nil {
			return nil , err
		}
		rs.Domain = append(rs.Domain, d)
	}
	if nil == rs.Domain || len(rs.Domain) < 1 {
		return nil, &DomainError{"未获取到域名数据"}
	}
	return rs, nil
}

//根据域名查找
func GetComIDAndModuleByDomain(domain string) (*DomainData, error) {

	fmt.Println("domain string: ", domain)

	var com DomainData
	filter := bson.D{{"domain", domain}}
	err := getDomainCollection().FindOne(context.TODO(), filter).Decode(&com)
	if err != nil {
		return nil, &DomainError{"域名未注册"}
	}
	if com.Status == false {
		return nil, &DomainError{"域名已停用"}
	}
	return &com, nil
}

//更新公司的状态
func UpdateDomainStatusByComID(comID int64, status bool) error {
	_, err := getDomainCollection().UpdateOne(context.TODO(), bson.D{{"comid", comID}}, bson.M{"$set": bson.M{"status": status}})
	return err
}


