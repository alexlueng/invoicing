package service

import (
	"context"
	"errors"
	"jxc/models"

	"go.mongodb.org/mongo-driver/bson"
)

// 用户验证器
type UserRules struct {

	UserName string `form:"username" json:"username" binding:"required,min=5,max=20"`
	Password string `form:"password" json:"password" binding:"required,min=8,max=20"`
	Phone    string `form:"phone" json:"phone" binding:"required,min=6,max=20"`
}

// 用户登录
func Login(comId, userName, password string) (models.User, error) {
	User := models.User{}
	collection := models.Client.Collection("users")
	err := collection.FindOne(context.TODO(), bson.D{{"com_id", comId}, {"username", userName}}).Decode(&User)
	if err != nil {
		// 无此用户
		return models.User{}, errors.New("无此用户")
	}
	if User.Password != password {
		// 密码错误
		return models.User{}, errors.New("密码错误")
	}

	return User, nil
}

// 创建一条登录日志
func CreateLoginLog(user_id, ip, msg string) {

	loginLog := models.LoginLogData{
		UserId:  user_id,
		Ip:      ip,
		Message: msg,
	}

	_, err := models.Collection.InsertOne(context.TODO(), loginLog)
	if err != nil {
		// 添加日志失败
		//checkErr(err)
	}
}

// 查找用户
func FindUser(user_id []int64, com_id int64) (map[int64]models.User, error) {
	var user models.User
	users := make(map[int64]models.User) // map[user_id]user
	filter := bson.M{}
	filter["user_id"] = bson.M{"$in": user_id}
	filter["com_id"] = com_id
	collection := models.Client.Collection("users")
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		err = cur.Decode(&user)
		if err != nil {
			return nil, err
		}
		users[user.UserID] = user
	}
	return users, nil
}

// 查找一条用户信息
func FindOneUser(userId int64, comId int64) (*models.User, error) {
	var user models.User
	filter := bson.M{}
	filter["user_id"] = userId
	filter["com_id"] = comId
	collection := models.Client.Collection("users")
	err := collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	// 获取这个用户的所有权限路由节点id，在根据节点id获取所有路由
	filter = bson.M{}
	auth_note := models.AuthNote{}
	 urls:= []string{"/api/v1/customer_settlement/getcustomer",
	 	"/api/v1/customer_settlement/getsettlement",
	 	"/api/v1/customer_settlement/create",
		 "/api/v1/customer_settlement/detail",
		 "/api/v1/customer_settlement/confirm",
		 "/api/v1/supplier_settlement/list",
		 "/api/v1/supplier_settlement/getsupplier",
		 "/api/v1/supplier_settlement/getsettlement",
		 "/api/v1/supplier_settlement/create",
		 "/api/v1/supplier_settlement/detail",
		 "/api/v1/supplier_settlement/confirm",
	 	 "/upload_images"}
	filter["auth_id"] = bson.M{"$in": user.Authority}
	cur, err := models.Client.Collection("auth_note").Find(context.TODO(), filter)
	if err != nil {
		// 没有找到对应的数据，返回空
		return &user,nil
	}
	//defaultUrl := []string{"/api/v1/units"}
	for cur.Next(context.TODO()) {
		err = cur.Decode(&auth_note)
		if err != nil {
			continue
		}
		for _, val := range auth_note.Urls {
			urls = append(urls, val)
		}
	}
	user.Urls = urls
	return &user, nil
}

// 更新用户信息
func UpdateUser(user models.User, user_id string) error {
	_, err := models.Collection.UpdateOne(context.TODO(), bson.M{"user_id": user_id}, bson.M{"$set": bson.M{"password": user.Password}})
	if err != nil {
		return errors.New("修改密码失败")
	}
	return nil
}

// 添加用户
func AddUser(user models.User) (string, error) {
	// 指定数据库 invoicing ，数据集 users
	collection := models.Client.Collection("users")
	_, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		return "", err
	}

	return "", errors.New("")
}
