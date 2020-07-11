package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"net/http"
	"time"
)

// 消息提醒的接口
type UserMessageService struct {
	UserID int64 `json:"user_id"`
}

type MessageSrv struct {
	MessageID int64 `json:"message_id"`
}

// 用户一登录系统就会去查message表，返回消息列表
// TODO：应该按时间倒序来排序
// TODO：消息列表应该可以分页，可以选择未读

func MessageList(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	collection := models.Client.Collection("message")
	var lists []models.Message

	cur, err := collection.Find(context.TODO(), bson.D{{"com_id", claims.ComId}})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeMessageErr,
			Msg:  err.Error(),
		})
		return
	}
	for cur.Next(context.TODO()) {
		var res models.Message
		if err := cur.Decode(&res); err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeMessageErr,
				Msg:  err.Error(),
			})
			return
		}
		lists = append(lists, res)
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Get user messages",
		Data: lists,
	})
	return
}

// message details
// 用户阅读消息后将该消息设置为已读
func ReadMessage(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var messSrv MessageSrv
	if err := c.ShouldBindJSON(&messSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	collection := models.Client.Collection("message")
	updateResult, err := collection.UpdateOne(context.TODO(), bson.D{{"id", messSrv.MessageID}, {"com_id", claims.ComId}}, bson.M{
		"$set": bson.M{"is_read": true, "read_at" : time.Now().Unix()}})
	if err != nil {
		fmt.Println("Can't update message: ", err)
		return
	}
	fmt.Println("update message: ", updateResult.UpsertedID)

	collection = models.Client.Collection("system_config")
	updateResult, err = collection.UpdateOne(context.TODO(), bson.D{{"com_id", claims.ComId}}, bson.M{
		"$inc" : bson.M{"unread_message" : -1}})
	if err != nil {
		fmt.Println("Can't update unread message: ", err)
		return
	}
	fmt.Println("update unread message: ", updateResult.UpsertedID)

	var sysConf SystemConfig
	collection = models.Client.Collection("system_config")
	err = collection.FindOne(context.TODO(), bson.D{{"com_id", claims.ComId}}).Decode(&sysConf)
	if err != nil {
		fmt.Println("Can't find system config: ", err)
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Update user messages",
		Data: sysConf.UnreadMessage,
	})
}

// 提醒消息分两类，一类是数量类的，如商品，仓库库存，这种情况在减库存后算一下便可触发
// 另一类是时间类的提醒，如订单，结算单的过期或者逾期时间，这种需要定时任务来触发

// @title    OrderNotify
// @description   通用订单提醒函数，生成提醒消息记录到消息表中
// @auth      alex             时间（2020/4/2   18:30 ）
// @param     status        int         实例状态
// @param     expireDay     int         过期提醒天数
// @param     messType      int         消息类型
// @param     filterTime    string      订单查找条件
// @param     orderType     int         订单类型 （客户，供应商）
// @return

// 函数单一职责功能
// 1 找出需要提醒的记录（未发货，未审核，未生成结算单）
// 2 将这些记录的ID放到一个数组中
// 3 生成提醒消息插入数据表中

// 消息列表的阅读权限
// 既然难以区分提醒消息的应该接收的用户，那么就让系统管理员为每个用户设置消息提醒权限
// 只有拥有消息列表阅读权限的人才能查看消息

type QueryLimit struct {
	ComID    int64 `json:"com_id" bson:"com_id"`
	ExpireAt int64 `json:"expire_at" bson:"expire_at"`
}

// 从系统设置中获取订单超时提醒的时间
// 遍历good_instance表，找出需要提醒的订单实例
// 生成消息，存到订单消息表中
// 这个接口限制用户每天访问一次
// 增加一个接口查询表 query_limit表
func OrderMessages(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, err := auth.ParseToken(token)

	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	expire_date, err := getExpireDateFromSystem(claims.ComId) // 过期时间，超过这个时间的话就生成提醒消息
	if err != nil { // 没有设置过期时间，返回空列表
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeMessageErr,
			Msg:  "No user messages",
		})
		return
	}

	var ql QueryLimit
	collection := models.Client.Collection("query_limit")
	err = collection.FindOne(context.TODO(), bson.D{{"com_id", claims.ComId}}).Decode(&ql)
	if err != nil { // 没有找到expire_date 说明这个用户是第一次登录
		qLimit := QueryLimit{
			ComID:    claims.ComId, // 为该com_id生成一个这样的记录
			ExpireAt: time.Now().Unix(),
		}
		_, err := collection.InsertOne(context.TODO(), qLimit)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't insert expire date",
			})
			return
		}
	}

	if time.Now().Unix() < ql.ExpireAt {
		var sysConf SystemConfig
		collection = models.Client.Collection("system_config")
		err := collection.FindOne(context.TODO(), bson.D{{"com_id", claims.ComId}}).Decode(&sysConf)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Can't get system config",
			})
			return
		}
		// 说明还没到时间
		c.JSON(http.StatusOK, serializer.Response{
			Code: 200,
			Msg:  "Update user messages",
			Data: sysConf.UnreadMessage,
		})
		return
	}

	// 发货提醒

	collection = models.Client.Collection("goods_instance")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["type"] = 1   // 客户订单
	filter["status"] = 1 // 未发货
	filter["order_time"] = bson.M{"$lte": time.Now().Unix() - int64(expire_date)*86400}

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Can't find order instances: ", err)
		return
	}
	var instances []models.GoodsInstance
	for cur.Next(context.TODO()) {
		var res models.GoodsInstance
		if err := cur.Decode(&res); err != nil {
			fmt.Println("Can't decode good instance: ", err)
			return
		}
		instances = append(instances, res)
	}

	messageType := 2
	collection = models.Client.Collection("message_type")
	var t models.MessageType
	err = collection.FindOne(context.TODO(), bson.D{{"id", messageType}}).Decode(&t)
	if err != nil {
		fmt.Println("Can't find message type: ", err)
		return
	}

	var messageList []interface{}

	for _, instance := range instances {
		var message models.Message
		message.ID = GetLastID("message")
		message.ComID = instance.ComID
		message.Type = int64(messageType)
		message.Title = t.Name

		real_expire := (time.Now().Unix() - instance.OrderTime) / 86400 + 1

		message.Message = fmt.Sprintf(t.Template, instance.InstanceId, real_expire)
		message.IsRead = false
		message.NotifyWay = "web"
		message.CreateAt = time.Now().Unix()

		SetLastID("message")
		messageList = append(messageList, message)

	}
	if len(messageList) > 0 {
		collection = models.Client.Collection("message")
		_, err = collection.InsertMany(context.TODO(), messageList)
		if err != nil {
			fmt.Println("Can't insert message: ", err)
			return
		}
	}


	// 更新系统未读消息
	collection = models.Client.Collection("system_config")
	updateResult, err := collection.UpdateOne(context.TODO(), bson.D{{"com_id", claims.ComId}}, bson.M{
		"$inc" : bson.M{"unread_message" : len(messageList)}})
	if err != nil {
		fmt.Println("Can't update unread message: ", err)
		return
	}
	fmt.Println("update unread message: ", updateResult.UpsertedID)

	// 最后要更新query_limit表中的expireat
	collection = models.Client.Collection("query_limit")
	updateResult, err = collection.UpdateOne(context.TODO(), bson.D{{"com_id", claims.ComId}}, bson.M{
		"$set" : bson.M{"expire_at" : time.Now().Unix() + 86400}})

	if err != nil {
		fmt.Println("Can't update query limit")
		return
	}
	fmt.Println("update result: ", updateResult.UpsertedID)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Update user messages",
		Data: len(messageList),
	})

}

func OrderNotify(status int, expireDay int, messType int, filterTime string, orderType int) {
	// 查找需要提醒的订单
	collection := models.Client.Collection("goods_instance")
	filter := bson.M{}

	filter["status"] = status // 订单状态
	if status == 3 { // 如果是结算提醒，还需要分客户还是供应商
		if orderType == 1 { // 客户
			filter["cussettle"] = 0
		}
		if orderType == 2 { //供应商
			filter["supsettle"] = 0
		}
	}
	// 计算过期时间
	// 这个时间是当前生成消息提醒记录的分界值，小于这个时间的且没有发货的订单将会被提醒，
	expireTime := time.Now().Unix() - int64(expireDay)*86400
	filter[filterTime] = bson.M{"$lte": expireTime} //  小于 time.Now().Unix() - 3天
	var instances []models.GoodsInstance
	//var instanceIDs []int64
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Can't find instance: ", err)
		return
	}
	for cur.Next(context.TODO()) {
		var res models.GoodsInstance
		if err := cur.Decode(&res); err != nil {
			fmt.Println("Can't decode instance: ", err)
			return
		}
		instances = append(instances, res)
	}

	// 找出了需要提醒的订单实例，生成提醒消息，应该是向一个特定的部门发送一条消息，先暂时是发给超级管理员
	// 对于查询到的每一条记录生成一条提醒消息
	// 在记录中有com_id, 有user_id

	messageType := messType
	collection = models.Client.Collection("message_type")
	var t models.MessageType
	err = collection.FindOne(context.TODO(), bson.D{{"id", messageType}}).Decode(&t)
	if err != nil {
		fmt.Println("Can't find message type: ", err)
		return
	}

	var messageList []interface{}

	for _, instance := range instances {
		var message models.Message
		message.ID = GetLastID("message")
		// 这是错误的代码
		// 超级管理员创建的订单的create_by也会有值，就区分不了是哪一个用户
		// 如果是超级管理员创建的订单，create_by置为0
		// 还是采取另一种方法，不区分管理员还是普通用户，
		// 把生成的消息都存到系统消息表中
		// 把这个消息发送给这个模块的所有用户

		//message.User = instance.CreateBy // com_id 加上 user_id可以确定唯一用户

		message.ComID = instance.ComID
		message.Type = int64(messageType)
		message.Message = fmt.Sprintf(t.Template, instance.InstanceId)
		message.IsRead = false
		message.NotifyWay = "web"
		message.CreateAt = time.Now().Unix()

		SetLastID("message")
		messageList = append(messageList, message)

	}

	collection = models.Client.Collection("message")
	_, err = collection.InsertMany(context.TODO(), messageList)
	if err != nil {
		fmt.Println("Can't insert message: ", err)
		return
	}
}

// 消息列表
func MessageForClients(c *gin.Context) {
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	collection := models.Client.Collection("client_message")
	cur, err := collection.Find(context.TODO(), bson.D{{"com_id", claims.ComId}})
	if err != nil {
		fmt.Println("Can't find client message: ", err)
		return
	}
	var messages []models.MessageForClient
	for cur.Next(context.TODO()) {
		var res models.MessageForClient
		if err := cur.Decode(&res); err != nil {
			fmt.Println("Can't decode client message: ", err)
			return
		}
		messages = append(messages, res)
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Update user messages",
		Data: messages,
	})
}
