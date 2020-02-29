package router

import (
	"jxc/api"
	"net/http"

	//	"invoicing/api/company"
	"jxc/middleware"
	"os"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	gin.SetMode(gin.DebugMode)

	r := gin.Default()

	r.StaticFS("/upload/images", http.Dir("F:\\projects\\jxc\\src\\assets"))

	// 中间件, 顺序不能改
	r.Use(middleware.Session(os.Getenv("SESSION_SECRET")))
	r.Use(middleware.Cors())
	r.Use(middleware.CurrentUser())
	//r.Use(static.Serve("/", static.LocalFile("/tmp", false)))
	//r.Use(middleware.AuthRequired())
	//设置默认路由当访问一个错误网站时返回
	r.NoRoute(api.NotFound)

	// 设置静态文件路径
	//r.MaxMultipartMemory = 8 << 20 // 8 MiB
	//r.Static("/assets", "./assets")

	//使用以下gin提供的Group函数为不同的API进行分组
	r.GET("auth", api.GetAuth)
	// 对于特定路由需要特定中间件的写法
	//v1.POST("/company/create", middleware.JWT(), api.CreateCompany)

	v1 := r.Group("/api/v1")
	{
		//v1.GET("/register", api.Register)
		v1.POST("/login", api.Login)
		//auth := v1.Group("")
		// 用户是否要带着token来访问路由
		//v1.Use(middleware.JWT())
		{
			v1.GET("/", api.Index)

			v1.POST("/upload_images", api.UploadImages)

			// 公司管理
			v1.GET("/company/list", api.AllCompanies)
			//v1.POST("/company/create", api.CreateCompany)
			v1.GET("/company/detail", api.CompanyDetail)
			v1.POST("/company/update", api.UpdateCompany)
			v1.GET("/delivery", api.DeliveryList)
			v1.GET("/units", api.UnitsList)
			v1.GET("/payment", api.PaymentList)
			v1.GET("/default_profit_margin", api.DefaultProfitMargin)

			// 人员管理
			v1.GET("/employee/list", api.AllEmployees)
			v1.POST("/employee/list", api.AllEmployees)
			v1.POST("/employee/create", api.AddEmployee)
			v1.POST("/employee/update", api.UpdateEmployee)
			v1.POST("/employee/delete", api.DeleteEmployee)
			v1.GET("/position/list", api.AllPositions)
			v1.GET("/auth/list", api.AllAuthNote)
			v1.POST("/employee/auth", api.UpdateAuthority)

			// 仓库管理
			v1.POST("/warehouse/list", middleware.GetComIDAndModuleID(), api.AllWarehouses)
			v1.POST("/warehouse/create", api.AddWarehouse)
			v1.POST("/warehouse/update", api.UpdateWarehouse)
			v1.POST("/warehouse/delete", api.DeleteWarehouse)
			v1.POST("/warehouse/detail", api.WarehouseDetail)

			// 客户管理
			v1.POST("/customer/list", api.ListCustomers)
			v1.POST("/customer/create", api.AddCustomer)
			v1.POST("/customer/update", api.UpdateCustomer)
			v1.POST("/customer/delete", api.DeleteCustomer)

			// 客户订单管理
			v1.POST("/customer_order/list", api.AllCustomerOrders)
			v1.POST("/customer_order/create", api.AddCustomerOrder)
			v1.POST("/customer_order/update", api.UpdateCustomerOrder)
			v1.POST("/customer_order/delete", api.DeleteCustomerOrder)
			v1.POST("/customer_order/detail", api.CustomerOrderDetail)
			v1.POST("/customer_order/customerprice", api.CustomerPrice)
			v1.POST("/customer_order/checkcustomerprice", api.CheckCustomerPrice)

			// AllCustomerSubOrderInstance
			v1.GET("/customer_order/sub_order_info",api.AllCustomerSubOrderInstance)//获取子订单实例列表

			// 供应商管理
			v1.POST("/supplier/list", api.ListSuppliers)
			v1.POST("/supplier/create", api.AddSuppliers)
			v1.POST("/supplier/update", api.UpdateSuppliers)
			v1.POST("/supplier/delete", api.DeleteSuppliers)

			// 供应商订单管理
			v1.POST("/supplier_order/list", api.AllSupplierOrders)
			v1.POST("/supplier_order/create", api.AddSupplierOrder)
			v1.POST("/supplier_order/update", api.UpdateSupplierOrder)
			v1.POST("/supplier_order/delete", api.DeleteSupplierOrder)
			v1.POST("/supplier_order/detail", api.SupplierOrderDetail)
			//v1.POST("/supplier_order/customerprice", middleware.GetComIDAndModuleID(), api.CustomerPrice)
			v1.POST("/supplier_order/customer_purchase", api.AddCustomerPurchaseOrder) //客户订单/客户采购
			v1.POST("/supplier_order/purchase_order", api.AddPurchaseOrder)            //添加采购订单

			v1.POST("/customer_order/checksupplierprice", api.CheckSupplierPrice)

			// 商品管理
			v1.POST("/product/list", api.AllProducts)
			v1.POST("/product/create", api.AddProduct)
			v1.POST("/product/update", api.UpdateProduct)
			v1.POST("/product/delete", api.DeleteProduct)
			v1.POST("/product/addprice", api.AddPrice) //添加供应商价格
			v1.POST("/product/detail", middleware.GetComIDAndModuleID(), api.ProductDetail)
			v1.POST("/product/supplierlist", api.SupplierListOfProducts)
			v1.GET("/product/img_load_sign", api.GetYpyunSign) //获取又拍云上传签名

			// 库存实例
			v1.GET("/wos_examples/wos_product", api.ProductWos) // 获取商品的库存
			v1.POST("/wos/add_product", api.CreateWosInstance) // 补充商品
			v1.POST("/wos/wos_log", api.AllWosInstance) // 获取库存实例日志
			// 客户价格管理(售价管理)
			v1.POST("/customer_price/list", api.ListCustomerPrice)
			v1.POST("/customer_price/create", api.AddCustomerPrice)
			// 供应商价格管理(售价管理)
			v1.POST("/supplier_price/list", api.ListSupplierPrice)
			v1.POST("/supplier_price/create", api.AddSupplierPrice)

			// 客户结算管理
			v1.POST("/customer_settlement/list", api.ListCustomerSettlement)
		}

	}

	return r
}
