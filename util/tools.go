package util

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"gitee.com/xiaochengtech/wechat/wxpay"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 这个文件存放公用的工具方法

// 获取weqiid返回数据格式
type WeqiidResponse struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data WeqiidResponseData
}

// data 字段格式
type WeqiidResponseData struct {
	Table string `json:"table"`
	Count int64  `json:"count"`
}

// 某个数组是否包含某个字符串
//同PHP in_array
func InArray(arr []string, str string) bool {
	for _, val := range arr {
		if val == str {
			return true
		}
	}
	return false
}

func InArrayInt64(arr []int64, _int int64) bool {
	for _, val := range arr {
		if val == _int {
			return true
		}
	}
	return false
}

//作为客户端发送post 请求
//postForm 需要发送的数据
//server_url 服务端地址
func SendPost(postForm map[string]string, server_url string) (body []byte, err error) {
	urlValues := url.Values{}
	//装填请求的参数
	for key, val := range postForm {
		urlValues.Add(key, val)
	}
	//发送请求
	resp, sendErr := http.PostForm(server_url, urlValues)
	if sendErr != nil {
		return nil, sendErr
	}
	//接收请求参数
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	return body, nil
}

// 向weqiid 获取表id
func GetTableId(table string) (int64, error) {
	postForm := map[string]string{
		"table": table,
	}
	// 如果读取不到配置则返回提示信息
	fmt.Println("weqi_id url: ", os.Getenv("GET_TABLE_ID"))
	url := os.Getenv("GET_TABLE_ID")
	if url == "" {
		return 0, errors.New("获取id失败")
	}
	data, err := SendPost(postForm, url)
	if err != nil {
		return 0, err
	}
	res := WeqiidResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return 0, errors.New("获取id失败")
	}
	if res.Code != 200 {
		return 0, errors.New("获取id失败")
	}

	return res.Data.Count, nil
}

// 生成密码摘要
func PasswordBcrypt(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

//密码验证
func PasswordVerify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// 数组去重
func RemoveRepeatedElement(arr []string) (newArr []string) {
	newArr = make([]string, 0)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return
}

func RemoveRepeatedElementInt64(arr []int64) (newArr []int64) {
	newArr = make([]int64, 0)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return
}

// 生成订单号，规则：日期(20200101) + com_id + 递增id
func GetOrderSN(com_id int64) (string, error) {
	currentDate := time.Now().Format("20060102")
	var orderSn string
	orderHead := currentDate + strconv.FormatInt(com_id, 10)
	orderId, err := GetTableId(orderHead)
	if err != nil {
		return "", err
	}
	orderIdStr := strconv.FormatInt(orderId, 10)
	orderSn = orderHead + orderIdStr
	return orderSn, nil
}

//将float64转成精确的int64
func Wrap(num float64, retain int) int64 {
	return int64(num * math.Pow10(retain))
}

//将int64恢复成正常的float64
func Unwrap(num int64, retain int) float64 {
	return float64(num) / math.Pow10(retain)
}

//精准float64
func WrapToFloat64(num float64, retain int) float64 {
	return num * math.Pow10(retain)
}

//精准int64
func UnwrapToInt64(num int64, retain int) int64 {
	return int64(Unwrap(num, retain))
}

// 文件保存路径
// 公司id，存储内容，
func GetFileSavePath(comId int64, fileType string) string {
	// 文件目录 upload/comId/fileType/日期/   避免目录存储过多文件
	comIdStr := strconv.FormatInt(comId, 10)
	currentDate := time.Now().Format("20060102")
	path := "/upload/" + comIdStr + "/" + fileType + "/" + currentDate
	return path
}

func StructToMap(item interface{}) map[string]interface{} {
	res := make(map[string]interface{})

	if item == nil {
		return res
	}

	v := reflect.TypeOf(item)
	reflectValue := reflect.ValueOf(item)
	reflectValue = reflect.Indirect(reflectValue)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		tag := v.Field(i).Tag.Get("json")
		field := reflectValue.Field(i).Interface()

		if tag != "" && tag != "-" {
			if v.Field(i).Type.Kind() == reflect.Struct {
				res[tag] = StructToMap(field)
			} else {
				res[tag] = field
			}
		}
	}
	return res
}

// 微信支付计算签名的函数
func WxpaySign(body wxpay.UnifiedOrderBody, key string) (sign string) {

	data := StructToMap(body)
	//STEP 1, 对key进行升序排序.
	sorted_keys := make([]string, 0)
	for k, _ := range data {
		sorted_keys = append(sorted_keys, k)
	}

	sort.Strings(sorted_keys)

	//STEP2, 对key=value的键值对用&连接起来，略过空值
	var signStrings string
	for _, k := range sorted_keys {
		value := fmt.Sprintf("%v", data[k])
		if value != "" {
			signStrings = signStrings + k + "=" + value + "&"
		}
	}

	//STEP3, 在键值对的最后加上key=API_KEY
	if key != "" {
		signStrings = signStrings + "key=" + key
	}
	fmt.Println("sign string: ", signStrings)
	//STEP4, 进行MD5签名并且将所有字符转为大写.
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(signStrings))
	cipherStr := md5Ctx.Sum(nil)
	upperSign := strings.ToUpper(hex.EncodeToString(cipherStr))

	fmt.Println("Get wxpay sign: ", upperSign)

	return upperSign
}

func GetIpAddress(c *gin.Context) string {
	realip := c.GetHeader("X-Real-Ip")
	if realip != "" {
		return realip
	}
	remoteAddr := c.Request.RemoteAddr
	if remoteAddr != "" {
		idx := strings.Index(remoteAddr, ":")
		return remoteAddr[:idx]
	}
	ips := c.GetHeader("X-Forwarded-For")
	fmt.Println("X-Forwarded-For", ips)
	if ips != "" {
		iplist := strings.Split(ips, ",")
		return strings.TrimSpace(iplist[0])
	}
	return ""
}


























