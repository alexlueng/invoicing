package models

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Supplier struct {
	ID             int64    `json:"supplier_id" bson:"supplier_id"`
	ComID          int64    `json:"com_id" bson:"com_id"`
	Phone          string   `json:"phone" bson:"phone"`
	SupplierName   string   `json:"supplier_name" bson:"supplier_name"`
	Address        string   `json:"address" bson:"address"`
	Contacts       string   `json:"contacts" bson:"contacts"`
	TransactionNum int64    `json:"transaction_num" bson:"transaction_num"` //交易次数
	Payment        string   `json:"payment" bson:"payment"`
	Level          int64    `json:"level" bson:"level"`             // 应该有一个默认级别，然后有一个机制会提升这个级别
	SupplyList     []int64  `json:"supply_list" bson:"supply_list"` // 此供应商供应的商品列表
	LastSettlement int64    `json:"last_settlement" bson:"last_settlement"`
	Password       string   `json:"password" bson:"password"`
	NotifyWay      []Notify `json:"notify_way" bson:"notify_way"`
}

type Notify struct {
	Name  string `json:"name" bson:"name"`
	Using bool   `json:"using" bson:"using"`
	Extend string `json:"extend" bson:"extend"`
}

func getSupplierCollection() *mongo.Collection {
	return Client.Collection("supplier")
}

func (c *Supplier) FindAll(filter bson.M, options *options.FindOptions) ([]Supplier, error) {
	var result []Supplier
	cur, err := getSupplierCollection().Find(context.TODO(), filter, options)
	if err != nil {
		//logrus.Error("Can't get customer list")
		fmt.Println("Can't get supplier list")
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var r Supplier
		if err := cur.Decode(&r); err != nil {
			fmt.Println("Can't decode into supplier")
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func (c *Supplier) Total(filter bson.M) (int64, error) {
	total, err := getSupplierCollection().CountDocuments(context.TODO(), filter)
	return total, err
}

func (c *Supplier) CheckExist() bool {
	filter := bson.M{}
	filter["com_id"] = c.ComID
	filter["supplier_name"] = c.SupplierName

	err := getSupplierCollection().FindOne(context.TODO(), filter).Err()
	if err != nil {
		// 说明没有存在重名
		return false
	}
	return true
}

func (c *Supplier) Insert() error {
	_, err := getSupplierCollection().InsertOne(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Supplier) Delete() error {
	filter := bson.M{}
	filter["com_id"] = c.ComID
	filter["supplier_id"] = c.ID
	_, err := getSupplierCollection().DeleteOne(context.TODO(), filter)
	if err != nil {

		return err
	}
	return nil
}

// false: 检查不通过
func (c *Supplier) UpdateCheck() bool {

	filter := bson.M{}
	filter["com_id"] = c.ComID
	filter["supplier_name"] = c.SupplierName
	cur, err := getSupplierCollection().Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("error found decoding supplier: ", err)
		return false
	}
	for cur.Next(context.TODO()) {
		var tempRes Supplier
		err := cur.Decode(&tempRes)
		if err != nil {
			fmt.Println("error found decoding supplier: ", err)
			return false
		}
		if tempRes.ID != c.ID {
			return false
		}
	}
	return true
}

func (c *Supplier) Update() error {
	filter := bson.M{}
	filter["com_id"] = c.ComID
	filter["supplier_id"] = c.ID
	// 更新记录
	_, err := getSupplierCollection().UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"supplier_name": c.SupplierName,
			"contacts": c.Contacts,
			"phone":    c.Phone,
			"address":  c.Address,
			"payment":  c.Payment,
			"level":    c.Level}})
	if err != nil {
		return err
	}
	return nil
}

func (s *Supplier) UpdatePassword(pass string) error {
	collection := getSupplierCollection()
	_, err := collection.UpdateOne(context.TODO(), bson.D{{"supplier_id", s.ID}}, bson.M{
		"$set" : bson.M{"password" : pass}})
	if err != nil {
		return err
	}
	return nil
}

func (c *Supplier) FindByID(id int64) (*Supplier, error) {
	filter := bson.M{}
	filter["com_id"] = c.ComID
	filter["supplier_id"] = id
	err := getSupplierCollection().FindOne(context.TODO(), filter).Decode(c)
	if err != nil {
		fmt.Println("inner error: ", err)
		return nil, err
	}
	return c, nil
}

type SupplierReq struct {
	BaseReq
	// 本页订制的搜索条件
	ID       int64  `json:"supplier_id" form:"supplier_id"`
	Phone    string `json:"phone" form:"phone"`
	Contacts string `json:"contacts" form:"contacts"`
	Name     string `json:"supplier_name" form:"supplier_name"`
}

type ResponseSupplierData struct {
	Suppliers   []Supplier `json:"suppliers"`
	Total       int64      `json:"total"`
	Pages       int64      `json:"pages"`
	Size        int64      `json:"size"`
	CurrentPage int64      `json:"current_page"`
}

// 供应商设备
type SupplierDevice struct {
	ID           int64  `json:"id" bson:"id"`
	ComID        int64  `json:"com_id" bson:"com_id"`
	SupplierID   int64  `json:"supplier_id" bson:"supplier_id"`
	DeviceName   string `json:"device_name" bson:"device_name"`
	DeviceNumber string `json:"device_number" bson:"device_number"` // 设备序列号
}

//
//type SupplierNotify struct {
//	ID        int64  `json:"id" bson:"id"`
//	ComID     int64  `json:"com_id" bson:"com_id"`
//	NotifyWay string `json:"notify_way" bson:"notify_way"` // 通知方式：邮件 公众号 短信
//	Active    bool   `json:"active" bson:"active"`
//}
