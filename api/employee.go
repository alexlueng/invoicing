package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	UserId    string `form:"" json:"user_id"`
	Password  string `form:"" json:"password"`  // 登录密码
	Username  string `form:"" json:"username"`  // 用户名
	Phone     string `form:"" json:"phone"`     // 电话
	Authority string `form:"" json:"authority"` // 权限
	Position  string `form:"" json:"position"`  // 职务
}

// 公用返回的数据格式
type ResponseData struct {
	Total       int         `json:"total"`
	Pages       int         `json:"pages"`
	Size        int         `json:"size"`
	CurrentPage int         `json:"current_page"`
}

// 人员管理返回的数据格式
type EmplResponseData struct {
	Users interface{} `json:"users"`
	ResponseData
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

	// 每页显示数量 min = 10 , max = 100
	// 页码 min = 1
	if req.Size > 100 {
		req.Size = 100
	}
	if req.Size < 10 {
		req.Size = 10
	}
	if req.Page < 1 {
		req.Page = 1
	}
	// 设置排序主键
	orderField := []string{"user_id", "username", "phone", "authority", "position"}
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
		filter["name"] = bson.M{"$regex": req.Username}
	}
	if req.Phone != "" {
		filter["phone"] = bson.M{"$regex": req.Phone}
	}
	if req.Authority != "" {
		filter["authority"] = bson.M{"$eq": req.Authority}
	}

	// 每个查询都要带着com_id去查
	//com_id, _ := strconv.Atoi(com.ComId)
	filter["com_id"] = com.ComId

	// 当前请求页面的数据
	cur, err := models.Client.Collection("user").Find(ctx, filter, option)
	if err != nil {
		fmt.Println("error while setting findoptions: ", err)
		return
	}

	for cur.Next(context.TODO()) {
		var result models.User
		err := cur.Decode(&result)
		if err != nil {
			fmt.Println("error found decoding customer: ", err)
			return
		}
		users = append(users, result)
	}

	//查询的总数
	var total int64
	cur, _ = models.Client.Collection("user").Find(ctx, filter)
	for cur.Next(context.TODO()) {
		total++
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
		Msg:  "Get customers",
		Data: resData,
	})
}

func AddEmployee(c *gin.Context) {
	// 根据域名获取comid

	// 用户是否有对应的权限

	// 验证提交的数据

	// 判断用户名是否重复

	// 插入数据库

	// 返回结果
}

func UpdateEmployee(c *gin.Context) {
	// 根据域名获取comid

	// 用户是否有对应的权限

	// 验证提交的数据

	// 需要更新的用户记录是否存在

	// 如果修改了用户名，判断是否重名

	// 插入数据库

	// 返回结果
}

func DeleteEmployee(c *gin.Context) {
	// 根据域名获取comid

	// 用户是否有对应的权限

	// 验证提交的数据

	// 需要删除的用户记录是否存在

	// 删除数据

	// 返回结果
}
