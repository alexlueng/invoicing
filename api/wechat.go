package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"jxc/util"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 微信公众号，小程序处理接口

//type VerifyFileService struct {
//	PlatformType string `json:"type"`
//	Dst          string `json:"dst"` // 文件存放路径
//}


//https://developers.weixin.qq.com/miniprogram/dev/api/wx.getUserInfo.html

type VerifyFileService struct {
	UploadKey    *multipart.FileHeader `form:"file"`
	PlatformType string                `form:"type"`
	Dst          string                `form:"dst"` // 文件存放路径
}

// 微信登录返回
type WXLoginResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type Watermark struct {
	AppID     string `json:"appid"`
	TimeStamp int64  `json:"timestamp"`
}

type WXUserInfo struct {
	OpenID    string    `json:"openId,omitempty"`
	NickName  string    `json:"nickName"`
	AvatarUrl string    `json:"avatarUrl"`
	Gender    int       `json:"gender"`
	Country   string    `json:"country"`
	Province  string    `json:"province"`
	City      string    `json:"city"`
	UnionID   string    `json:"unionId,omitempty"`
	Language  string    `json:"language"`
	Watermark Watermark `json:"watermark,omitempty"`
}

// 返回上传的目录
func FileUploadPath(c *gin.Context) {
	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	upload_path := "/opt/attachment/" + strconv.Itoa(int(claims.ComId)) + "/"
	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Upload path",
		Data: upload_path,
	})
}

// 上传微信验证文件到服务器根目录
func UploadVerifyFile(c *gin.Context) {

	token := c.GetHeader("Access-Token")
	claims, _ := auth.ParseToken(token)

	var verifyFileSrv VerifyFileService
	if err := c.ShouldBind(&verifyFileSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Params error",
		})
		return
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "get upload dir error",
		})
		return
	}

	savePath := dir + "\\upload\\attachment\\" + strconv.Itoa(int(claims.ComId)) + "\\wx\\"
	util.Log().Info("Save upload path: ", savePath)

	_, err = os.Stat(savePath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(savePath, os.ModePerm)

		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: serializer.CodeError,
				Msg:  "Create directory error",
			})
			return
		}
	}

	if err := c.SaveUploadedFile(verifyFileSrv.UploadKey, savePath+verifyFileSrv.UploadKey.Filename); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "save file error",
		})
		return
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: serializer.CodeSuccess,
		Msg:  "Upload file succeed",
	})

}

/*
	要求用户上传文件到指定的目录
	在nginx中配置转发规则 如 weqi.exechina.com/MP_verify_faghaklgaa34r1n41.txt
	的请求转发到这个路由
	这个路由返回请求到的信息

	代码实现：获取get请求参数，到指定目录找到这个文件，
	a_path := c.Request.URL
*/
// 微信认证域名
func WechatVerify(c *gin.Context) {


	// 获取请求的域名，可以得知所属公司
	domain := c.Request.Header.Get("Origin")
	com, err := models.GetComIDAndModuleByDomain(domain[len("http://"):])
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  err.Error(),
		})
		return
	}

	filename := c.Request.URL.Path
	fmt.Println("Wechat get request: ", filename)
	filename = filename[1:]

	filename = strings.Replace(filename, "/", "\\", -1)

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "get upload dir error",
		})
		return
	}

	savePath := dir + "\\upload\\attachment\\" + strconv.Itoa(int(com.ComId)) + "\\" + filename
	util.Log().Info("Save upload path: ", savePath)

	file, err := os.Open(savePath)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Read file error",
		})
		return
	}

	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Read file info error",
		})
		return
	}

	fmt.Println("File info: ", fileInfo.Name())

	buffer := make([]byte, fileInfo.Size())
	_, err = file.Read(buffer) // 文件内容读取到buffer中
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusOK, serializer.Response{
			Code: serializer.CodeError,
			Msg:  "Read file content error",
		})
		return
	}

	c.String(http.StatusOK, string(buffer))
}

type PayClient struct {
	httpClient *http.Client
}

func NewPayClient(httpClient *http.Client) *PayClient {

	if httpClient == nil {

		httpClient = http.DefaultClient
		httpClient.Timeout = time.Second * 5
	}

	return &PayClient{
		httpClient: httpClient,
	}
}

// 获取支付相关参数
func (pc *PayClient) WechatLogin(jscode string) (*models.PayData, error) {
	loginUrl := "https://api.weixin.qq.com/sns/jscode2session?appid=" + url.QueryEscape(models.PayConfigInstance.AppId()) +
		"&secret=" + url.QueryEscape(models.PayConfigInstance.AppSecret()) +
		"&js_code=" + url.QueryEscape(jscode) +
		"&grant_type=authorization_code"

	httpResp, err := pc.httpClient.Get(loginUrl)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http.Status: %s", httpResp.Status)
	}

	respData := models.NewPayData()
	err = respData.FromJson(httpResp.Body)
	if err != nil {
		return nil, err
	}

	return respData, nil
}

// 按照微信支付的接口格式提交XML类型的参数
func (pc *PayClient) PostXML(url string, pdata *models.PayData) (*models.PayData, error) {
	reqSignType := pdata.Get("sign")
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(pdata.ToXml()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	req = req.WithContext(ctx)
	httpResp, err := pc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http.Status: %s", httpResp.Status)
	}

	respData := models.NewPayData()
	err = respData.FromXml(httpResp.Body)
	if err != nil {
		return nil, err
	}

	err = respData.CheckSign(models.PayConfigInstance.ApiKey(), reqSignType)
	if err != nil {
		return nil, err
	}

	return respData, nil
}
