package models

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Customer represent the customer
// 需要加上com_id, 每个公司都有自己的ID
type Customer struct {
	ID        int64  `json:"customer_id" bson:"customer_id"`
	ComID     int64  `json:"com_id" bson:"com_id"`
	Name      string `json:"customer_name" form:"customer_name"`
	Level     int64  `json:"level" form:"level"`
	Payment   string `json:"payment" form:"payment"`
	PayAmount float64 `json:"paid" form:"paid" bson:"paid"`
	Receiver  string `json:"receiver" form:"receiver"`
	Address   string `json:"receiver_address" form:"address" bson:"receiver_address"`
	Phone     string `json:"receiver_phone" form:"phone" bson:"receiver_phone"`
	//due string
}

func (c Customer) FindAll(filter bson.M, options *options.FindOptions) ([]Customer, error) {
	var result []Customer
	cur, err := Client.Collection("customer").Find(context.TODO(), filter, options)
	if err != nil {
		//logrus.Error("Can't get customer list")
		fmt.Println("Can't get customer list")
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var r Customer
		if err := cur.Decode(&r); err != nil {
			fmt.Println("Can't decode into customer")
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func (c Customer) Total(filter bson.M) (int64, error) {
	total, err := Client.Collection("customer").CountDocuments(context.TODO(), filter)
	return total, err
}

func (c Customer) CheckExist() bool {
	filter := bson.M{}
	filter["com_id"] = c.ComID
	filter["name"] = c.Name

	err := Client.Collection("customer").FindOne(context.TODO(), filter).Err()
	if err != nil {
		// 说明没有存在重名
		return false
	}
	return true
}

func (c Customer) Insert() error {
	_, err := Client.Collection("customer").InsertOne(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

// false: 检查不通过
func (c Customer) UpdateCheck() bool {

	filter := bson.M{}
	filter["com_id"] = c.ComID
	filter["name"] = c.Name
	cur, err := Client.Collection("customer").Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("error found decoding customer: ", err)
		return false
	}
	for cur.Next(context.TODO()) {
		var tempRes Customer
		err := cur.Decode(&tempRes)
		if err != nil {
			fmt.Println("error found decoding customer: ", err)
			return false
		}
		if tempRes.ID != c.ID {
			return false
		}
	}
	return true
}



func (c Customer) Update() error {
	fmt.Println(c)
	filter := bson.M{}
	filter["com_id"] = c.ComID
	filter["customer_id"] = c.ID
	// 更新记录
	_, err := Client.Collection("customer").UpdateOne(context.TODO(), filter, bson.M{
		"$set": bson.M{"name": c.Name,
			"receiver": c.Receiver,
			"receiver_phone": c.Phone,
			"receiver_address": c.Address,
			"payment": c.Payment,
			"level": c.Level}})
	if err != nil {
		return err
	}
	return nil
}

func (c Customer) Delete() error {
	filter := bson.M{}
	filter["com_id"] = c.ComID
	filter["customer_id"] = c.ID
	collection := Client.Collection("customer")
	_, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {

		return err
	}
	return nil
}

//用户提交过来的数据
type CustReq struct {
	IdMin int `form:"idmin"` //okid界于[idmin 和 idmax] 之间的数据
	IdMax int `form:"idmax"` //ok
	//本页面的搜索字段 sf固定等于customer_name， key的值为用户提交过来的客户名关键字
	Key  string `form:"key"`              //用户提交过来的模糊搜索关键字
	Sf   string `form:"sf"`               //用户模糊搜索的字段  search field
	Page int64  `json:"page" form:"page"` //ok用户查询的是哪一页的数据
	Size int64  `json:"size" form:"size"` //ok用户希望每页展现多少条数据
	OrdF string `json:"ordf" form:"ordf"` //ok用户排序字段 order field
	Ord  string `json:"ord" form:"ord"`   //ok顺序还是倒序排列  ord=desc 倒序，ord = asc 升序
	TMin int    `form:"tmin"`             //时间最小值[tmin,tmax)
	TMax int    `form:"tmax"`             //时间最大值
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
	Total       int64        `json:"total"`
	Pages       int64        `json:"pages"`
	Size        int64        `json:"size"`
	CurrentPage int64        `json:"current_page"`
}
