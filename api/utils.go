package api

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"jxc/util"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
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

type ImageURL struct {
	ProductID int64  `json:"product_id"`
	LocalURL  string `json:"local_path"`
	CloudURL  string `json:"cloud_path"`
}

type CategoryImageURL struct {
	CategoryID int64  `json:"category_id"`
	LocalURL   string `json:"local_path"`
	CloudURL   string `json:"cloud_path"`
}

func UploadImages(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("err: %s", err.Error()))
		return
	}

	files := form.File["file"]

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	save_path := dir

	urls := []ImageURL{}

	for _, file := range files {

		url := ImageURL{}

		path, filename := util.GetYpyunPath(file.Filename)
		upload_path := save_path + path
		_, err = os.Stat(upload_path)

		if os.IsNotExist(err) {
			err = os.MkdirAll(upload_path, os.ModePerm)

			if err != nil {
				c.JSON(http.StatusOK, serializer.Response{
					Code: serializer.CodeError,
					Msg:  "Create directory error",
				})
				return
			}
		}
		err := c.SaveUploadedFile(file, upload_path+filename)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Save image error",
			})
			return
		}

		url.LocalURL = upload_path + filename

		//ypyunURL := "http://img.jxc.weqi.exechina.com/upload/" + strconv.Itoa(int(claims.ComId)) + "/product_img/" + filename
		//fmt.Println("url: ", ypyunURL)

		ypyunURL1 := "/upload/" + strconv.Itoa(int(claims.ComId)) + "/product_img/" + filename

		err = util.UpYunPut(ypyunURL1, upload_path+filename)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Upload image to net error",
			})
			return
		}
		ret_url := "http://img.jxc.weqi.exechina.com" + ypyunURL1
		url.CloudURL = ret_url

		urls = append(urls, url)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Get supplier instance",
		Data: urls,
	})
}

type DeleteImageService struct {
	ImageID int64 `json:"image_id"`
}

// 删除图片接口
func DeleteImages(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var delImageSrv DeleteImageService
	if err := c.ShouldBindJSON(&delImageSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeSuccess,
			Msg:  "Params error",
		})
		return
	}

	collection := models.Client.Collection("image")
	filter := bson.M{}
	filter["com_id"] = claims.ComId
	filter["image_id"] = delImageSrv.ImageID
	var image models.Image
	err := collection.FindOne(context.TODO(), filter).Decode(&image)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Decode image error",
		})
		return
	}

	if err := deleteImage(image); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Delete image success",
	})
}

func deleteImage(image models.Image) error {

	// 删除本地的照片
	util.Log().Info("local dir: ", image.LocalPath)
	if _, err := os.Stat(image.LocalPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		util.Log().Info("image does not exist: ", image.LocalPath)
	} else {
		err := os.Remove(image.LocalPath)
		if err != nil {
			return err
		}
	}

	// 删除云端的图片
	if err := util.UpYunDelete(image.CloudPath, true); err != nil {
		return err
	}

	collection := models.Client.Collection("image")
	_, err := collection.UpdateOne(context.TODO(), bson.D{{"com_id", image.ComID}, {"image_id", image.ImageID}}, bson.M{
		"$set": bson.M{"is_delete": true}})
	if err != nil {
		return err
	}
	return nil
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
	sn := strconv.Itoa(coc.Count + 1)
	if len(sn) < 6 {
		fmt.Printf("len of sn: %d\n", len(sn))
		step := 6 - len(sn)
		for i := 0; i < step; i++ {
			sn = "0" + sn
		}
	}
	current_date := time.Now().Format("20060102")
	sn = current_date + sn
	fmt.Println("Current OrderSN: ", sn)

	_, _ = collection.UpdateOne(context.TODO(), bson.M{"name": "customer_order"}, bson.M{"$set": bson.M{"count": coc.Count + 1}})

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

func SmartPrint(i interface{}) {
	var kv = make(map[string]interface{})
	vValue := reflect.ValueOf(i)
	vType := reflect.TypeOf(i)
	for i := 0; i < vValue.NumField(); i++ {
		kv[vType.Field(i).Name] = vValue.Field(i)
	}
	fmt.Println("获取到数据:")
	for k, v := range kv {
		fmt.Print(k)
		fmt.Print(":")
		fmt.Print(v)
		fmt.Println()
	}
}

// 需要传一个自定义的数组，里面的元素是可以排序的字段
func SetPaginationAndOrder(ordF string, ordFields []string, ord string, page, size int64) *options.FindOptions {

	exist := false
	for _, v := range ordFields {
		if ordF == v {
			exist = true
			break
		}
	}
	if !exist {
		ordF = ordFields[0]
	}
	order := -1
	if ord == "asc" {
		order = 1
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
func GetLastID(field_name string) int64 {
	collection := models.Client.Collection("counters")
	updateResult, err := collection.UpdateOne(context.TODO(), bson.D{{"name", field_name}}, bson.M{"$inc": bson.M{"count": 1}})
	if err != nil || updateResult.ModifiedCount < 1 {
		collection.InsertOne(context.TODO(), bson.D{{"name", field_name}, {"count", 1}})
		return 1
	}

	var c Counts
	err = collection.FindOne(context.TODO(), bson.D{{"name", field_name}}).Decode(&c)
	if err != nil {
		return 0
	}
	return c.Count
}

type Config struct {
	ProductMenu string `json:"product_menu"`
}

func GetConfig(c *gin.Context) {

	// TODO：需要一个灵活的方法
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	jsonFile, err := os.Open(dir + "\\config.json")
	if err != nil {
		fmt.Println("Can't read json file: ", err)
		return
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var config Config
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Println("Can't get config: ", err)
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get Config",
		Data: config,
	})
}

var (
	BeijingLocation = time.FixedZone("Asia/Shanghai", 8*60*60)
)

func RandomString(n int) string {

	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	length := len(letter)

	for i := range b {
		b[i] = letter[rand.Intn(length)]
	}

	return string(b)
}

func NonceStr() string {

	return RandomString(32)
}

// FormatTime time.Time => "yyyyMMddHHmmss."
func FormatTime(t time.Time) string {

	return t.In(BeijingLocation).Format("20060102150405")
}

// ParseTime "yyyyMMddHHmmss" => time.Time.
func ParseTime(value string) (time.Time, error) {

	return time.ParseInLocation("20060102150405", value, BeijingLocation)
}

func TimeStamp() string  {

	return strconv.FormatInt(time.Now().Unix(), 10)
}