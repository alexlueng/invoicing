package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"jxc/models"
	"jxc/serializer"
	"jxc/util"
	"net/http"
	"strings"
	"time"
)

//用户提交过来能数据
type EmplReq struct {
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
	UserId    int64    `form:"user_id" json:"user_id"`
	Password  string   `form:"password" json:"password"`     // 登录密码
	Username  string   `form:"username" json:"username"`     // 用户名
	Phone     string   `form:"phone" json:"phone"`           // 电话
	Authority []string `form:"authority[]" json:"authority"` // 权限
	Position  string   `form:"position" json:"position"`     // 职务
}

// 公用返回的数据格式
type ResponseData struct {
	Total       int `json:"total"`
	Pages       int `json:"pages"`
	Size        int `json:"size"`
	CurrentPage int `json:"current_page"`
}

// 人员管理返回的数据格式
type EmplResponseData struct {
	Users interface{} `json:"users"`
	ResponseData
}

// 获取提交权限数据
type ReqAuthNote struct {
	Authid  int64 `json:"authid" form:"authid"`
	Groupid int64 `json:"groupid" form:"groupid"`
}

//
type ReqAuth struct {
	UserId int64         `json:"user_id" form:"user_id"`
	Auth   []ReqAuthNote `json:"auth" form:"auth"`
}

// 获取所有人员
func AllEmployees(c *gin.Context) {
	// 根据域名获取comid
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}
	// 收集提交过来的参数，如页码、搜索关键字
	var req EmplReq
	var users []models.User

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// 验证提交过来的数据
	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	// 整理搜索条件

	req.Page, req.Size = SetDefaultPageAndSize(req.Page, req.Size)

	// 设置排序主键
	orderField := []string{"user_id", "username", "phone", "authority", "position", "create_at"}
	// 提交的排序主键是否合法
	if !util.InArray(orderField, req.OrdF) {
		req.OrdF = "user_id"
	}
	// 设置排序规则
	// 设置排序顺序
	// 降序 desc -1
	// 升序 asc 1
	order := 1
	if req.Ord == "desc" {
		order = -1
		req.Ord = "desc"
	} else {
		order = 1
		req.Ord = "asc"
	}

	// 查询数据库

	option := options.Find()
	option.SetLimit(int64(req.Size))
	option.SetSkip((int64(req.Page) - 1) * int64(req.Size))

	option.SetSort(bson.D{{req.OrdF, order}})

	filter := bson.M{}
	if req.Username != "" {
		filter["username"] = bson.M{"$regex": req.Username}
	}
	if req.Phone != "" {
		filter["phone"] = bson.M{"$regex": req.Phone}
	}
	if req.Position != ""{
		filter["position"] = req.Position
	}
	//if req.Authority != "" {
	//	filter["authority"] = bson.M{"$eq": req.Authority}
	//}

	// 每个查询都要带着com_id去查
	//com_id, _ := strconv.Atoi(com.ComId)
	filter["com_id"] = com.ComId

	// 当前请求页面的数据
	cur, err := models.Client.Collection("users").Find(ctx, filter, option)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "获取人员列表失败",
		})
		fmt.Println("error while setting findoptions: ", err)
		return
	}

	for cur.Next(context.TODO()) {
		var result models.User
		err := cur.Decode(&result)
		if err != nil {
			fmt.Println("error found decoding user: ", err)
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "获取人员列表失败",
			})
			return
		}
		result.Password = ""
		users = append(users, result)
	}

	total, err := models.Client.Collection("users").CountDocuments(context.TODO(), filter)
	if err != nil {
		fmt.Println("error found decoding user: ", err)
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "获取人员列表失败",
		})
		return
	}

	// 返回查询到的总数，总页数
	resData := EmplResponseData{}
	resData.Users = users
	resData.Total = int(total)
	resData.Pages = int(total)/int(req.Size) + 1
	resData.Size = int(req.Size)
	resData.CurrentPage = int(req.Page)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get users",
		Data: resData,
	})
}

// 添加人员信息
func AddEmployee(c *gin.Context) {
	// 根据域名获取comid
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}
	// 收集提交过来的参数，如页码、搜索关键字
	var req EmplReq
	var user models.User

	// 用户是否有对应的权限 TODO

	// 验证提交的数据
	// 验证提交过来的数据
	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	// 用户名可以重复，手机号不能重复

	createBy := int64(1) // TODO

	// 指定数据集
	collection := models.Client.Collection("users")
	filter := bson.M{}
	filter["phone"] = req.Phone
	filter["com_id"] = com.ComId

	_ = collection.FindOne(context.TODO(), filter).Decode(&user)
	if user.Username != "" {
		c.JSON(http.StatusOK, serializer.Response{
			Code:  -1,
			Data:  nil,
			Msg:   "手机号不能重复！",
			Error: "",
		})
		return
	}

	// 获取用户id
	user_id, err := util.GetTableId("users")
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code:  -1,
			Data:  nil,
			Msg:   "添加用户出错！",
			Error: "",
		})
		return
	}

	pssword, _ := util.PasswordBcrypt(req.Password)

	// 插入数据库
	user.Username = req.Username
	user.ComId = com.ComId
	user.Password = pssword
	user.Phone = req.Phone
	user.UserID = user_id
	user.Authority = req.Authority
	user.Position = req.Position
	user.CreateAt = time.Now().Unix()
	user.CreateBy = createBy

	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code:  -1,
			Data:  nil,
			Msg:   "添加用户出错！",
			Error: "",
		})
		return
	}

	// 添加成功，把创建的用户信息返回

	user.Password = ""
	c.JSON(http.StatusOK, serializer.Response{
		Code:  200,
		Data:  user,
		Msg:   "添加用户成功！",
		Error: "",
	})
	return
}

// 更新人员信息
func UpdateEmployee(c *gin.Context) {
	// 根据域名获取comid
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}
	// 收集提交过来的参数
	var req EmplReq
	var user, user2 models.User
	user_id := int64(1) // TODO

	// 用户是否有对应的权限 TODO

	// 验证提交的数据
	// 验证提交过来的数据
	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	// 判断数据id是否正确
	// 判断用户名是否重复
	// com_id = com_id,username = username,user_id != user_id

	// 指定数据集
	collection := models.Client.Collection("users")
	filter := bson.M{}
	filter["user_id"] = req.UserId
	err = collection.FindOne(context.TODO(), filter).Decode(&user)
	// 查询不到数据，传入的id有误
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code:  -1,
			Data:  nil,
			Msg:   "传入的id错误，没有对应的数据",
			Error: "",
		})
		return
	}

	filter["user_id"] = bson.M{"$ne": req.UserId}
	filter["phone"] = req.Phone
	filter["com_id"] = com.ComId

	_ = collection.FindOne(context.TODO(), filter).Decode(&user2)
	if user2.Username != "" {
		c.JSON(http.StatusOK, serializer.Response{
			Code:  -1,
			Data:  nil,
			Msg:   "手机号不能重复！",
			Error: "",
		})
		return
	}

	update := bson.M{}
	update["username"] = req.Username
	update["phone"] = req.Phone
	//update["authority"] = req.Authority
	update["position"] = req.Position
	update["modify_at"] = time.Now().Unix()
	update["modify_by"] = user_id

	// 如果修改了密码
	if req.Password != "" {
		password, _ := util.PasswordBcrypt(req.Password)
		update["password"] = password
	}

	// 更新数据
	_, err = collection.UpdateOne(context.TODO(), bson.M{"user_id": req.UserId}, bson.M{
		"$set": update,
	})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code:  -1,
			Data:  nil,
			Msg:   "用户更新失败！",
			Error: "",
		})
		return
	}

	// 返回结果
	user.Password = ""
	c.JSON(http.StatusOK, serializer.Response{
		Code:  200,
		Data:  user,
		Msg:   "修改成功！",
		Error: "",
	})
}

// 删除人员
func DeleteEmployee(c *gin.Context) {
	// 根据域名获取comid
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}
	// 收集提交过来的参数
	var req EmplReq
	var user models.User

	// 用户是否有对应的权限 TODO

	// 验证提交的数据
	// 验证提交过来的数据
	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	// 需要删除的用户记录是否存在

	// 指定数据集
	collection := models.Client.Collection("users")
	filter := bson.M{}
	filter["user_id"] = req.UserId
	err = collection.FindOne(context.TODO(), filter).Decode(&user)
	// 查询不到数据，传入的id有误
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code:  -1,
			Data:  nil,
			Msg:   "传入的id错误，没有对应的数据",
			Error: "",
		})
		return
	}

	// 删除数据
	_, err = collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code:  -1,
			Data:  nil,
			Msg:   "删除失败",
			Error: "",
		})
		return
	}

	// 返回结果
	user.Password = ""
	c.JSON(http.StatusOK, serializer.Response{
		Code:  200,
		Data:  user,
		Msg:   "删除成功！",
		Error: "",
	})
	return
}

// 获取职位列表
func AllPositions(c *gin.Context) {
	// 根据域名获取comid
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	// 指定数据集
	collection := models.Client.Collection("company")
	company := models.Company{}
	filter := bson.M{}
	filter["com_id"] = com.ComId
	err = collection.FindOne(context.TODO(), filter).Decode(&company)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "获取职位列表失败！",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: company.Position,
		Msg:  "获取职位列表成功！",
	})
}

// 获取所有权限节点列表
func AllAuthNote(c *gin.Context) {
	// 根据域名获取comid
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	// 指定数据集
	collection := models.Client.Collection("auth_note")
	warehouseCollection := models.Client.Collection("warehouse")
	var warehouse models.Warehouse
	var auth_note models.AuthNote
	var auth_notes []models.AuthNote
	// 获取权限节点
	cur, err := collection.Find(context.TODO(), bson.M{}, options.Find().SetSort(bson.M{"authid": 1}))
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "获取权限节点错误",
		})
		return
	}
	for cur.Next(context.TODO()) {
		err = cur.Decode(&auth_note)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "获取权限节点错误",
			})
			return
		}
		auth_notes = append(auth_notes, auth_note)
	}
	// 获取仓库
	cur, err = warehouseCollection.Find(context.TODO(), bson.M{"com_id": com.ComId})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "获取权限节点错误",
		})
		return
	}
	for cur.Next(context.TODO()) {
		err = cur.Decode(&warehouse)
		if err != nil {
			return
		}
		auth_notes = append(auth_notes, models.AuthNote{
			AuthId:  warehouse.ID,
			Note:    warehouse.Name,
			Group:   "仓库管理",
			GroupId: 5,
		})
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Data: auth_notes,
		Msg:  "获取权限节点成功！",
	})

}

// 更新用户权限
func UpdateAuthority(c *gin.Context) {
	// 根据域名获取comid
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	// 接收数据
	var req ReqAuth
	data, _ := ioutil.ReadAll(c.Request.Body)
	err = json.Unmarshal(data, &req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "更新权限失败",
		})
		return
	}

	// 把普通的权限和仓库权限分拣开
	// 仓库权限 Groupid = 5
	var warehouse_ids, auth_ids []int64

	for _, val := range req.Auth {
		if val.Groupid == 5 {
			warehouse_ids = append(warehouse_ids, val.Authid)
		} else {
			auth_ids = append(auth_ids, val.Authid)
		}
	}
	// 去重
	warehouse_ids = util.RemoveRepeatedElementInt64(warehouse_ids)
	auth_ids = util.RemoveRepeatedElementInt64(auth_ids)

	//指定数据集
	collection := models.Client.Collection("users")

	// 更新条件
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["user_id"] = req.UserId

	// 更新数据内容
	update := bson.M{}
	update["warehouse"] = warehouse_ids
	update["authority"] = auth_ids
	// 更新数据
	_, err = collection.UpdateOne(context.TODO(), filter, bson.M{
		"$set": update,
	})
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "更新权限失败",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "更新权限成功",
	})
	return
}
