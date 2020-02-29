package api

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"jxc/models"
	"jxc/serializer"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)



//设置默认路由当访问一个错误网站时返回
func NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"status": 404,
		"error":  "404 ,page not exists!",
	})
}

func UploadImages(c *gin.Context) {

	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("err: %s", err.Error()))
		return
	}

	fmt.Println("r.PostForm:     ",c.PostForm)

	fmt.Println("r.MultiPartForm:",c.MultipartForm)


	files := form.File["images"]
	fmt.Println("form files: ", files)

	for _, file := range files {
		fmt.Println(file.Filename)
		c.SaveUploadedFile(file, "src/assets/"+file.Filename)
	}


	c.String(http.StatusCreated, "upload successful \n")
}

// 根据请求的域名，确定是哪家公司
//func GetCompany(host string) (models.CompanyData, error) {
//	//host := c.Request.Host //这里可能获取到的是 http://host/ 的结构，需要做进一步拆分
//	// 在域名表中找到公司id
//	domain, err := service.FindDomain(host)
//	if err != nil {
//		// 域名没有注册，数据库中没有记录
//		return models.CompanyData{}, errors.New("域名未注册！")
//	}
//	models.Client.Collection("domain")
//	company, err := service.FindCompany(domain.ComID)
//	if err != nil {
//		// 在库中没找到对应的公司
//		return company, errors.New("未找到对应的公司")
//	}
//	return company, nil
//}

func Index(c *gin.Context) {
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Hello",
	})
}

func SystemConfig(c *gin.Context) {

}

// 顺序生成一个6位数的字符串，然后与日期拼接，得到当前的order_sn
func GetTempOrderSN() string {
	// 000001 000002 000003
	// 先顺序生成数字 然后转成字符串，不足6位的用0补齐
	var coc models.CustomerOrderCount
	collection := models.Client.Collection("counters")
	err := collection.FindOne(context.TODO(), bson.D{{"name", "customer_order"}}).Decode(&coc)
	if err != nil {
		fmt.Println("can't get OrderSN")
		return ""
	}
	sn := strconv.Itoa(coc.Count+1)
	if len(sn) < 6 {
		fmt.Printf("len of sn: %d\n", len(sn))
		step := 6-len(sn)
		for i := 0; i < step; i++ {
			sn = "0" + sn
		}
	}
	current_date := time.Now().Format("20060102")
	sn = current_date + sn
	fmt.Println("Current OrderSN: ", sn)

	_, _  = collection.UpdateOne(context.TODO(), bson.M{"name": "customer_order"}, bson.M{"$set": bson.M{"count": coc.Count + 1}})

	return sn
}

func SetDefaultPageAndSize(page, size int64) (int64, int64) {
	s := size
	p := page
	if s < 11 {
		s = 10
	}
	if p < 2 {
		p = 1
	}
	return p, s
}

func SmartPrint(i interface{}){
	var kv = make(map[string]interface{})
	vValue := reflect.ValueOf(i)
	vType :=reflect.TypeOf(i)
	for i:=0; i < vValue.NumField(); i++{
		kv[vType.Field(i).Name] = vValue.Field(i)
	}
	fmt.Println("获取到数据:")
	for k,v := range kv {
		fmt.Print(k)
		fmt.Print(":")
		fmt.Print(v)
		fmt.Println()
	}
}

// 需要传一个自定义的数组，里面的元素是可以排序的字段
func SetPaginationAndOrder(ordF string, ordFields []string,  ord string, page, size int64) *options.FindOptions {

	exist := false
//	fmt.Println("order field: ", req.OrdF)
	for _, v := range ordFields {
		if ordF == v {
			exist = true
			break
		}
	}
	if !exist {
		ordF = ordFields[0]
	}
	order := 1
	if ord == "desc" {
		order = -1
		//req.Ord = "desc"
	}

	option := options.Find()
	option.SetLimit(size)
	option.SetSkip((page - 1) * size)

	//1从小到大,-1从大到小
	option.SetSort(bson.D{{ordF, order}})

	return option
}

type Counts struct {
	NameField string
	Count     int64
}
// 因mongodb不允许自增方法，所以要生成新增客户的id
// 这是极度不安全的代码，因为本程序是分布式的，本程序可能放在多台服务器上同时运行的。
// 需要在交付之前修改正确
func getLastID(field_name string) int64 {
	var c Counts
	collection := models.Client.Collection("counters")
	err := collection.FindOne(context.TODO(), bson.D{{"name", field_name}}).Decode(&c)
	if err != nil {
		fmt.Println("can't get ID")
		return 0
	}
	collection.UpdateOne(context.TODO(), bson.M{"name": field_name}, bson.M{"$set": bson.M{"count": c.Count + 1}})
	fmt.Printf("%s count: %d", field_name, c.Count)
	return c.Count + 1
}

// 本地文件上传到又拍云



