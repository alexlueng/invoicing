package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"jxc/auth"
	"jxc/models"
	"jxc/serializer"
	"jxc/util"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// 微信公众号，小程序处理接口

//type VerifyFileService struct {
//	PlatformType string `json:"type"`
//	Dst          string `json:"dst"` // 文件存放路径
//}

type VerifyFileService struct {
	UploadKey    *multipart.FileHeader `form:"file"`
	PlatformType string                `form:"type"`
	Dst          string                `form:"dst"` // 文件存放路径
}

// 返回上传的目录
func FileUploadPath(c *gin.Context) {
	// /attachment/com_id/
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

	//uploadFile, err := c.FormFile("file")
	//if err != nil {
	//	c.JSON(http.StatusOK, serializer.Response{
	//		Code: serializer.CodeError,
	//		Msg:  "file upload error",
	//	})
	//	return
	//}

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
//
//func WeChatTest(c *gin.Context) {
//
//	resp, err := wxofficial.GetAccessToken(, appsecret)
//	if err != nil {
//		util.Log().Error("Can't get token: ", err)
//	}
//	fmt.Println("Get access token: ", resp.AccessToken)
//	c.JSON(http.StatusOK, serializer.Response{
//		Code: serializer.CodeSuccess,
//		Msg:  "Upload file succeed",
//		Data: resp,
//	})
//}

// 微信认证域名
func WechatVerify(c *gin.Context) {
	/*
		要求用户上传文件到指定的目录
		在nginx中配置转发规则 如 weqi.exechina.com/MP_verify_faghaklgaa34r1n41.txt
		的请求转发到这个路由
		这个路由返回请求到的信息

		代码实现：获取get请求参数，到指定目录找到这个文件，
		a_path := c.Request.URL
	*/

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

	//c.JSON(http.StatusOK, serializer.Response{
	//	Code: serializer.CodeSuccess,
	//	Msg:  "Get ok",
	//	Data: string(buffer),
	//})
	c.String(http.StatusOK, string(buffer))
}
