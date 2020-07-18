package models

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// mongo中数据表字段维护在代码中
// 售价管理表
type CustomerProductPrice struct {
	ComID        int64   `json:"com_id" bson:"com_id"`
	CustomerID   int64   `json:"customer_id" bson:"customer_id"`
	CustomerName string  `json:"customer_name" bson:"customer_name"`
	ProductID    int64   `json:"product_id" bson:"product_id"`
	Product      string  `json:"product" bson:"product_name"`
	Price        float64 `json:"price" bson:"price"`
	CreateAt     int64   `json:"create_at" bson:"create_at"`
	IsValid      bool    `json:"is_valid" bson:"is_valid"`
	DefaultPrice float64 `json:"default_price" bson:"omitempty"`
}

func getCustomerProductPriceCollection() *mongo.Collection {
	return Client.Collection("customer_product_price")
}

func (c *CustomerProductPrice) Add() error{
	_, err := getCustomerProductPriceCollection().InsertOne(context.TODO(), c)
	return err
}

type CustomerProductPriceReq struct {
	BaseReq
	Product      string `json:"product"`
	ProductID    int64  `json:"product_id"`
	CustomerID   int64  `json:"customer_id"`
	CustomerName string `json:"customer_name"`
}

type ResponseCustomerProductPriceData struct {
	PriceTable  interface{} `json:"price_table"`
	Total       int         `json:"total"`
	Pages       int         `json:"pages"`
	Size        int         `json:"size"`
	CurrentPage int         `json:"current_page"`
}

// 接收查找到的对应客户的价格
type CustomerOrderProductPrice struct {
	ProductID   int64   `json:"product_id" bson:"product_id"`
	ProductName string  `json:"product_name" bson:"product_name"`
	Price       float64 `json:"price" bson:"price"`
}

type CustomerOrderProductPriceResponse struct {
	CustomerOrderProductPriceList []CustomerOrderProductPrice `json:"customer_order_product_price_list"`
}

type CustomerProductPriceResponse struct {
	CustomerProductPriceList []CustomerProductPrice `json:"customer_product_price_list"`
}

func SelectCustomerProductPriceByCondition(filter bson.M) (*CustomerProductPrice, error) {
	var product CustomerProductPrice
	err := getCustomerProductPriceCollection().FindOne(context.TODO(), filter).Decode(&product)
	return &product, err
}

func SelectMultiplyCustomerOrderProductPriceByConditoin(filter bson.M) (*CustomerOrderProductPriceResponse, error) {
	cur, err := getCustomerProductPriceCollection().Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	var resp = new(CustomerOrderProductPriceResponse)
	for cur.Next(context.TODO()) {
		var result CustomerOrderProductPrice
		if err := cur.Decode(&result); err != nil {
			fmt.Println("error while decoding: ", err)
			return nil, err
		}
		resp.CustomerOrderProductPriceList = append(resp.CustomerOrderProductPriceList, result)
	}
	return resp, nil
}

func SelectMultiplyCustomerProductPriceByConditoin(filter bson.M) (*CustomerProductPriceResponse, error) {
	cur, err := getCustomerProductPriceCollection().Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	var resp = new(CustomerProductPriceResponse)
	for cur.Next(context.TODO()) {
		var result CustomerProductPrice
		if err := cur.Decode(&result); err != nil {
			fmt.Println("error while decoding: ", err)
			return nil, err
		}
		resp.CustomerProductPriceList = append(resp.CustomerProductPriceList, result)
	}
	return resp, nil
}

func SelectCustomerProductPriceByComIDAndProductID(comID int64, productID int64, isValid bool) (*CustomerProductPrice, error) {
	filter := bson.M{}
	filter["com_id"] = comID
	filter["product_id"] = productID
	filter["is_valid"] = isValid
	var record CustomerProductPrice
	err := getCustomerProductPriceCollection().FindOne(context.TODO(), filter).Decode(&record)
	return &record, err
}

func SelectCustomerProductPriceByComIDAndCustomerIDAndProductID(comID int64, customerID int64, productID int64, isValid bool) (*CustomerProductPrice, error) {
	filter := bson.M{}
	filter["com_id"] = comID
	filter["customer_id"] = customerID
	filter["product_id"] = productID
	filter["is_valid"] = isValid
	var record CustomerProductPrice
	err := getCustomerProductPriceCollection().FindOne(context.TODO(), filter).Decode(&record)
	return &record, err
}

func UpdateCustomerProductPriceValidStatus(comID int64, productID int64, oldValid bool, newValid bool) (*mongo.UpdateResult, error) {
	filter := bson.M{}
	filter["com_id"] = comID
	filter["product_id"] = productID
	filter["is_valid"] = oldValid
	return getCustomerProductPriceCollection().UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"is_valid": newValid}})
}

func UpdateCustomerProductPriceValidStatusWithCustomerID(comID int64, productID int64, customerID int64, oldValid bool, newValid bool) (*mongo.UpdateResult, error) {
	filter := bson.M{}
	filter["com_id"] = comID
	filter["product_id"] = productID
	filter["customer_id"] = customerID
	filter["is_valid"] = oldValid
	return getCustomerProductPriceCollection().UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"is_valid": newValid}})
}

