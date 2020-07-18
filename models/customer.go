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
	ID             int64     `json:"customer_id" bson:"customer_id"`
	ComID          int64     `json:"com_id" bson:"com_id"`
	Name           string    `json:"customer_name" bson:"name"`
	LevelID        int64     `json:"level" bson:"level"`     //用户等级
	Payment        string    `json:"payment" bson:"payment"` // 支付方式
	PayAmount      float64   `json:"paid" bson:"paid"`
	Receiver       string    `json:"receiver" bson:"receiver"`                 // TODO：弃用 放在addresses中
	Address        string    `json:"receiver_address" bson:"receiver_address"` // TODO：弃用 使用地址数组
	Phone          string    `json:"receiver_phone" bson:"receiver_phone"`
	LastSettlement int64     `json:"last_settlement" bson:"last_settlement"` // 上次结算时间
	Addresses      []Address `json:"addresses" bson:"addresses"`             // 用户收货地址
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

// 微信小程序用户
type MiniAppUser struct {
	ComID          int64   `json:"com_id" bson:"com_id"`
	UserID         int64   `json:"user_id" bson:"user_id"`
	OpenID         string  `json:"open_id" bson:"open_id"`     // 小程序用户唯一标识
	AvatarURL      string  `json:"avatarurl" bson:"avatarurl"` // 头像地址
	City           string  `json:"city" bson:"city"`
	Gender         int64   `json:"gender" bson:"gender"`       // 性别 1 男 2 女
	Language       string  `json:"language" bson:"language"`   // 语言
	NickName       string  `json:"nickname" bson:"nickname"`   // 昵称
	Province       string  `json:"province" bson:"province"`   // 省份
	Telephone      string  `json:"telephone" bson:"telephone"` // 手机号
	CreateAt       int64   `json:"create_at" bson:"create_at"`
	Verify         int64   `json:"verify" bson:"verify"`                   // 是否通过验证 0 待验证 1 已经验证 2 验证不通过
	SessionKey     string  `json:"session_key" bson:"session_key"`         // 微信服务器传回来的key
	Salt           string  `json:"salt" bson:"salt"`                       // 用于加密的随机字符串
	PurchaseAmount float64 `json:"purchase_amount" bson:"purchase_amount"` // 消费总额
}

// 用户上传验证资料
type UserVerifyMaterial struct {
	ComID           int64  `json:"com_id" bson:"com_id"`
	MaterialID      int64  `json:"material_id" bson:"material_id"`
	UserID          int64  `json:"user_id" bson:"user_id"`
	BusinessLicense string `json:"business_license" bson:"business_license"` // 营业执照图片
	LicenseNo       string `json:"license_no" bson:"license_no"`             // 营业执照号
	CompanyType     int64  `json:"company_type" bson:"company_type"`         // 公司类型 1 配送公司 2 档口
	CompanyName     string `json:"company_name" bson:"company_name"`         // 公司名称
	Contact         string `json:"contact" bson:"contact"`                   // 联系人
	Telephone       string `json:"telephone" bson:"telephone"`
	Address         string `json:"address" bson:"address"`
	VerifyCode      string `json:"verify_code" bson:"verify_code"` // 验证码
}
