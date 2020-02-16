package api

import (
	"fmt"
	"jxc/models"
	"jxc/serializer"
	"jxc/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Company struct {
	ComID     string              `json:"com_id"`
	ComName   string              `json:"com_name"`
	Delivery  string              `json:"delivery"`
	Domain    string              `json:"domain"`
	Units     string              `json:"units"`
	Developer string              `json:"developer"`
	Domains   []models.DomainData `json:"domains"`
}

// 获取所有配送方式
func AllCompanies(c *gin.Context) {

	var companies []Company

	com1 := Company{
		ComID:     "1",
		ComName:   "huazhi01",
		Delivery:  "shunfeng",
		Domain:    "www.huazhi01.com",
		Units:     "pounds",
		Developer: "alex",
	}
	companies = append(companies, com1)
	com2 := Company{
		ComID:     "2",
		ComName:   "huazhi02",
		Delivery:  "yunda",
		Domain:    "www.huazhi02.com",
		Units:     "pounds",
		Developer: "bob",
	}
	companies = append(companies, com2)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Hello",
		Data: companies,
	})
}

// 获取公司信息
func CompanyDetail(c *gin.Context) {
	// 获取请求的域名，可以得知所属公司
	host := c.Request.Host
	com, err := service.FindDomain(host)
	if err != nil {
		// 用户id找不到对应的公司
	}

	//标准post
	fmt.Println("p1=",c.Request.FormValue("p1"))
	fmt.Println("p2=",c.Request.FormValue("p2"))
	fmt.Println("p3=",c.PostForm("p3"))

	//RESTful
	//c.PostForm("id")接收的是标准POST的id字段的值
	//c.Query("id")接收的是标准GET的id字段的值jsonstr
	//c.Request.FormValue("id")接收的是标准GET的id字段的值
	buf := make([]byte, 1024)
	n, _ := c.Request.Body.Read(buf)
	jsonstr := string(buf[0:n])
	fmt.Println("body=",jsonstr)



	fmt.Println("p12=",c.Request.FormValue("p1"))
	fmt.Println("p22=",c.Request.FormValue("p2"))
	fmt.Println("p32=",c.PostForm("p3"))


	// 查找公司相应的信息
	company2, err := service.FindCompany(com.ComId)

	company := Company{
		ComID:   company2.ComId,
		ComName: company2.ComName,
	}

	// 找到公司下配置的所有域名
	domains, err := service.AllDomain(com.ComId)
	company.Domains = domains

	c.JSON(http.StatusOK, serializer.Response{
		Code:  200,
		Data:  company,
		Msg:   "",
		Error: "",
	})

}

// 更新公司信息
func UpdateCompany(c *gin.Context) {
	// 通过域名获取comid
	host := c.Request.Host
	com, err := service.FindDomain(host)
	if err != nil {
		// 用户id找不到对应的公司
	}
	fmt.Println(com)

	// 查找公司相应的信息
/*	company2, err := service.FindCompany(com.ComId)
	if err != nil {
		// 未找到公司信息
	}*/

	reqService := service.CompanyService{}

	// 提交的数据格式
	type delivery struct {
		Idx string `json:"idx"`
		EnTitle string `json:"en_title"`
		Title string `json:"title"`
	}

	type paymentType struct {
		PaymentId string `json:"payment_id"`
		PaymentTitle string `json:"payment_title"`
	}

	type requestData struct {
		ComName string `json:"com_name"`
		Delivery []delivery `json:"delivery"`
		Units []string `json:"units"`
		Payment []paymentType `json:"payment"`
		Developer string `json:"developer"`
		Domain []string `json:"domain"`
	}


	// 对提交上来的数据进行验证
	if err := c.ShouldBind(&reqService); err == nil {
		// 处理好需要更新的数据格式

		// 提取出域名，验证域名是否被注册


		// 提交给数据库保存
		req := requestData{}
		var domains  []interface{}
		for _,val := range req.Domain {
			domains = append(domains, models.DomainData{
				ComId:    com.ComId,
				Domain:   val,
				ModuleId: 1,
			})
		}





		// 保存域名


		// 返回更新结果

		//c.JSON(200, res)
	} else {
		c.JSON(200, serializer.Response{
			Code: -1,
			Msg:  err.Error(),
		})
	}

}
