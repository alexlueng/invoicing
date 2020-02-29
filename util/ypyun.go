package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/upyun/go-sdk/upyun"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

// URL = weqijxc.test.upcdn.net

// 又拍云异步上传步骤
// signature 和 policy 算法

// signature
// 1 .将所需的元信息键值对按照键的字典顺序排列，并连接成为字符串
// 2 将第 1 步中所得字符串与您的表单API
// 3 将第 2 步中所的的字符串计算 md5，所得即为 signature

// policy
// 将请求所需的元信息键值对转换为 JSON 字符串
// 将第 1 步中所得字符串进行 base64 编码，所得即为 policy

// 键	值
// path	/demo.png
// expiration	1409200758
// file_blocks	1
// file_hash	b1143cbc07c8e768d517fa5e73cb79ca
// file_size	653252

// 又拍云的元参数
type Parameter struct {
	Path       string `json:"path"`        // 上传后的路径
	Expiration string `json:"expiration"`  // 有效时间，这是一个未来时间戳
	FileBlocks string `json:"file_blocks"` // 文件桶
	FileHash   string `json:"file_hash"`   // 文件hash值 php版未提供此参数
	FileSize   int64  `json:"file_size"`   // 文件大小 php版未提供此参数
}

func md5Str(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func makeRFC1123Date(d time.Time) string {
	utc := d.UTC().Format(time.RFC1123)
	return strings.Replace(utc, "UTC", "GMT", -1)
}
func base64ToStr(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// 获取signature
func getSignature(path, expiration, file_hash, file_size string) string {
	apiKey := os.Getenv("YPYUNAPIKEY")
	file_blocks := os.Getenv("YPYUNOPERATORNAME")

	// 拼接数据
	str := path + expiration + file_blocks + file_hash + file_size
	// 加上apikey
	str = str + apiKey
	// md5 得到的字符串
	return md5Str(str)
}

// 获取 policy
func getPolicy(path, expiration, file_hash, file_size string) string {
	file_blocks := os.Getenv("YPYUNOPERATORNAME")

	parameter := Parameter{
		Path:       path,
		Expiration: expiration,
		FileBlocks: file_blocks,
		FileHash:   file_hash,
		FileSize:   0,
	}

	parameterStr, _ := json.Marshal(parameter)

	return base64ToStr(parameterStr)
}

// 获取又拍云上传图片签名
func GetYpyunSign(path, file_hash, file_size string) map[string]string {
	// 设置过期时间
	expiration := time.Now().Unix() + int64(600)
	expirationStr := strconv.FormatInt(expiration, 10)

	signature := getSignature(path, expirationStr, file_hash, file_size)
	policy := getPolicy(path, expirationStr, file_hash, file_size)

	return map[string]string{"signature": signature, "policy": policy}
}

// 生成上传又拍云路径 product_img/日期/时间戳+4位随机码.后缀名
func GetYpyunPath(fileName string) string {

	up := upyun.NewUpYun(&upyun.UpYunConfig{
		Bucket:   "demo",
		Operator: "op",
		Password: "password",
	})

	// 上传文件
	fmt.Println(up.Put(&upyun.PutObjectConfig{
		Path:      "/demo.log",
		LocalPath: "/tmp/upload",
	}))

	fileType := strings.Split(fileName, ".")
	currentDate := time.Now().Format("20060102")
	timeStr := strconv.FormatInt(time.Now().Unix(), 10)
	var randStr string
	for i := 0; i < 4; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(100))
		randStr += strconv.FormatInt(n.Int64(), 10)
	}
	path := "product_img/" + currentDate + "/" + timeStr + randStr + "." + fileType[1]
	return path
}

func NewUpYun() *upyun.UpYun {
	bucket := os.Getenv("YPYUNSERVICENAME")        //空间名
	Operator := os.Getenv("YPYUNOPERATORNAME")     // 操作员名
	Password := os.Getenv("YPYUNOPERATORPASSWORD") // 操作密码
	up := upyun.NewUpYun(&upyun.UpYunConfig{
		Bucket:   bucket,
		Operator: Operator,
		Password: Password,
	})
	return up
}

// 又拍云上传
func UpYunPut(path, localPath string) error {
	up := NewUpYun()
	err := up.Put(&upyun.PutObjectConfig{
		Path:      "/demo.log",
		LocalPath: "/tmp/upload",
	})
	return err
}


