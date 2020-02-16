package router

import (
	"jxc/api"
	//	"invoicing/api/company"
	"jxc/middleware"
	"os"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	gin.SetMode(gin.DebugMode)

	r := gin.Default()

	// 中间件, 顺序不能改
	r.Use(middleware.Session(os.Getenv("SESSION_SECRET")))
	r.Use(middleware.Cors())
	r.Use(middleware.CurrentUser())
	r.Use(middleware.AuthRequired())
	//设置默认路由当访问一个错误网站时返回
	r.NoRoute(api.NotFound)

	//使用以下gin提供的Group函数为不同的API进行分组

	v1 := r.Group("/api/v1")
	{
		//v1.GET("/register", api.Register)
		v1.POST("/login", api.Login)
		auth := v1.Group("")
		auth.Use(middleware.AuthRequired())
		{
			v1.GET("/", api.Index)

			// 公司管理
			v1.GET("/company/list", api.AllCompanies)
			//v1.POST("/company/create", api.CreateCompany)
			v1.GET("/company/detail", api.CompanyDetail)
			v1.POST("/company/update", api.UpdateCompany)

			// 人员管理
			v1.GET("/employee/list", api.AllEmployees)
			v1.POST("/employee/list", api.AllEmployees)
			v1.POST("/employee/create", api.AddEmployee)
			v1.POST("/employee/update/:id", api.UpdateEmployee)
			v1.POST("employee/delete/:id", api.DeleteEmployee)

			// 客户管理
			v1.POST("/customer/list", api.ListCustomers)
			v1.POST("/customer/create", api.AddCustomer)
			v1.POST("/customer/update", api.UpdateCustomer)
			v1.POST("/customer/delete", api.DeleteCustomer)

			// 仓库管理

			v1.POST("/warehouse/list", api.AllWarehouses)
			v1.POST("/warehouse/create", api.AddWarehouse)
			v1.POST("/warehouse/update", api.UpdateWarehouse)
			v1.POST("/warehouse/delete", api.DeleteWarehouse)

			// 供应商结算管理
			v1.POST("/customer_order/list", api.AllCustomerOrders)
			v1.POST("/customer_order/create", api.AddCustomerOrder)
			v1.POST("/customer_order/update", api.UpdateCustomerOrder)
			v1.POST("/customer_order/delete", api.DeleteCustomerOrder)
			v1.POST("/customer_order/detail", api.CustomerOrderDetail)

		}

	}

	return r
}
