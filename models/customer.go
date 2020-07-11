package models

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Customer represent the customer
// 需要加上com_id, 每个公司都有自己的ID
type Customer struct {
	ID             int64   `json:"customer_id" bson:"customer_id"`
	ComID          int64   `json:"com_id" bson:"com_id"`
	Name           string  `json:"customer_name" bson:"name"`
	LevelID        int64   `json:"level" bson:"level"`
	Payment        string  `json:"payment" bson:"payment"`
	PayAmount      float64 `json:"paid" bson:"paid"`
	Receiver       string  `json:"receiver" bson:"receiver"`
	Address        string  `json:"receiver_address" bson:"receiver_address"`
	Phone          string  `json:"receiver_phone" bson:"receiver_phone"`
	LastSettlement int64   `json:"last_settlement" bson:"last_settlement"` // 上次结算时间
}

func getCustomerCollection() *mongo.Collection {
	return Client.Collection("customer")
}

func (c *Customer) FindAll(filter bson.M, options *options.FindOptions) ([]Customer, error) {
	var result []Customer
	cur, err := getCustomerCollection().Find(context.TODO(), filter, options)
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var r Customer
		if err := cur.Decode(&r); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func (c *Customer) Total(filter bson.M) (int64, error) {
	total, err := getCustomerCollection().CountDocuments(context.TODO(), filter)
	return total, err
}

func (c *Customer) CheckExist() bool {
	filter := bson.M{}
	filter["com_id"] = c.ComID
	filter["name"] = c.Name

	err := getCustomerCollection().FindOne(context.TODO(), filter).Err()
	if err != nil {
		// 说明没有存在重名
		return false
	}
	return true
}

func (c *Customer) Insert() error {
	_, err := getCustomerCollection().InsertOne(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

// false: 检查不通过
func (c *Customer) UpdateCheck() bool {

	filter := bson.M{}
	filter["com_id"] = c.ComID
	filter["name"] = c.Name
	cur, err := getCustomerCollection().Find(context.TODO(), filter)
	if err != nil {
		return false
	}
	for cur.Next(context.TODO()) {
		var tempRes Customer
		err := cur.Decode(&tempRes)
		if err != nil {
			return false
		}
		if tempRes.ID != c.ID {
			return false
		}
	}
	return true
}

func (c *Customer) Update() error {
	fmt.Println(c)
	filter := bson.M{}
	filter["com_id"] = c.ComID
	filter["customer_id"] = c.ID
	// 更新记录
	_, err := getCustomerCollection().UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"name": c.Name,
			"receiver":         c.Receiver,
			"receiver_phone":   c.Phone,
			"receiver_address": c.Address,
			"payment":          c.Payment,
			"level":            c.LevelID}})
	if err != nil {
		return err
	}
	return nil
}

func (c *Customer) Delete() error {
	filter := bson.M{}
	filter["com_id"] = c.ComID
	filter["customer_id"] = c.ID
	_, err := getCustomerCollection().DeleteOne(context.TODO(), filter)
	if err != nil {

		return err
	}
	return nil
}

func (c *Customer) FindByID(id int64) (*Customer, error) {
	filter := bson.M{}
	filter["com_id"] = c.ComID
	filter["customer_id"] = id
	err := getCustomerCollection().FindOne(context.TODO(), filter).Decode(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

//用户提交过来的数据
type CustReq struct {
	BaseReq
	//本页面定制的搜索字段
	Name     string `json:"customer_name" form:"customer_name"`
	Level    string `json:"level" form:"level"`
	Payment  string `json:"payment" form:"payment"`
	Receiver string `json:"receiver" form:"receiver"` //模糊搜索
	Address  string `json:"address" form:"address"`   //模糊搜索
	Phone    string `json:"phone" form:"phone"`       //模糊搜索
}

type ResponseCustomerData struct {
	Customers   []Customer `json:"customers"`
	Levels      []Level    `json:"levels"`
	Total       int64      `json:"total"`
	Pages       int64      `json:"pages"`
	Size        int64      `json:"size"`
	CurrentPage int64      `json:"current_page"`
}
