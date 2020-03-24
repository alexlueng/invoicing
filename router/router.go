package router

import (
	"github.com/gin-gonic/gin"
	"jxc/api"
	//	"invoicing/api/company"
	"jxc/middleware"
)

func InitRouter() *gin.Engine {
	gin.SetMode(gin.DebugMode)

	r := gin.Default()

	// 服务之间的通信，不需要权限验证
	r.POST("/add_superadmin", api.AddSuperAdmin)
	r.POST("/adminpasswd_update", api.UpdateAdminPasswd)
	r.POST("/changedomainstatus", api.ChangeDomainStatus)

	// 中间件, 顺序不能改
	//r.Use(middleware.Session(os.Getenv("SESSION_SECRET")))
	// 允许跨域
	r.Use(middleware.Cors())
	// 鉴权
	//r.Use(middleware.CheckAuth())

	r.Use(middleware.CurrentUser())

	//设置默认路由当访问一个错误网站时返回
	r.NoRoute(api.NotFound)



	//使用以下gin提供的Group函数为不同的API进行分组
	//r.GET("auth", api.GetAuth)
	v1 := r.Group("/api/v1")
	{
		//v1.GET("/register", api.Register)
		v1.POST("/login", api.Login)
		{
			v1.GET("/", api.Index)

			v1.POST("/upload_images", api.UploadImages)

			// 公司管理
			v1.GET("/company/list", api.AllCompanies)
			//v1.POST("/company/create", api.CreateCompany)
			v1.GET("/company/detail", api.CompanyDetail)              // 获取公司详细信息
			v1.POST("/company/update", api.UpdateCompany)             // 更新公司信息
			v1.GET("/delivery", api.DeliveryList)                     // 获取配送方式列表
			v1.GET("/units", api.UnitsList)                           // 获取计量单位
			v1.GET("/payment", api.PaymentList)                       // 获取配送结算方式
			v1.GET("/default_profit_margin", api.DefaultProfitMargin) // 获取默认利润率



			// 人员管理
			v1.GET("/employee/list", api.AllEmployees)      // 获取人员列表
			v1.POST("/employee/list", api.AllEmployees)     // 获取人员列表
			v1.POST("/employee/create", api.AddEmployee)    // 添加人员
			v1.POST("/employee/update", api.UpdateEmployee) // 更新人员信息
			v1.POST("/employee/delete", api.DeleteEmployee) // 删除人员
			v1.GET("/position/list", api.AllPositions)      //职位列表
			v1.GET("/auth/list", api.AllAuthNote)           // 权限列表
			v1.POST("/employee/auth", api.UpdateAuthority)  // 修改人员权限

			// 仓库管理
			v1.POST("/warehouse/list", middleware.GetComIDAndModuleID(), api.AllWarehouses) // 获取仓库列表
			v1.POST("/warehouse/create", api.AddWarehouse)                                  // 添加仓库
			v1.POST("/warehouse/update", api.UpdateWarehouse)                               // 更新仓库
			v1.POST("/warehouse/delete", api.DeleteWarehouse)                               // 删除仓库
			v1.POST("/warehouse/detail", api.WarehouseDetail)                               //仓库详情

			// 客户管理
			v1.POST("/customer/list", api.ListCustomers)    // 客户列表
			v1.POST("/customer/create", api.AddCustomer)    // 添加客户
			v1.POST("/customer/update", api.UpdateCustomer) // 更新客户信息
			v1.POST("/customer/delete", api.DeleteCustomer) // 删除公司

			// 客户订单管理
			v1.POST("/customer_order/list", api.AllCustomerOrders) //  获取所有客户订单列表
			v1.POST("/customer_order/create", api.AddCustomerOrder) // 创建客户订单
			v1.POST("/customer_order/update", api.UpdateCustomerOrder) // 修改客户订单
			v1.POST("/customer_order/delete", api.DeleteCustomerOrder) // 删除客户订单
			v1.POST("/customer_order/detail", api.CustomerOrderDetail) // 客户订单详情
			v1.POST("/customer_order/customerprice", api.CustomerPrice) //
			v1.POST("/customer_order/checkcustomerprice", api.CheckCustomerPrice) //

			v1.GET("/customer_order/sub_order_instance_list", api.AllCustomerSubOrderInstance)     //获取子订单实例列表
			v1.POST("/customer_order/customer_order_shipped", api.CustomerSubOrderInstanceShipped) //子订单实例发货
			v1.POST("/customer_order/customer_order_confirm", api.CustomerSubOrderInstanceConfirm) //客户子订单实例确认收货
			v1.POST("/customer_order/customer_order_check", api.CustomerSubOrderInstanceCheck)     //客户子订单实例审核通过
			v1.POST("/supplier_order/supplier_order_confirm", api.SupplierSubOrderInstanceConfirm) // 采购订单/确认收货
			v1.POST("/supplier_order/supplier_order_check", api.SupplierSubOrderInstanceCheck)     // 采购订单/审核

			// 供应商管理
			v1.POST("/supplier/list", api.ListSuppliers) // 供应商列表
			v1.POST("/supplier/create", api.AddSuppliers) // 添加供应商
			v1.POST("/supplier/update", api.UpdateSuppliers) // 更新供应商信息
			v1.POST("/supplier/delete", api.DeleteSuppliers) // 删除供应商

			// 供应商订单管理
			v1.POST("/supplier_order/list", api.AllSupplierOrders) // 采购订单列表
			v1.POST("/supplier_order/create", api.AddSupplierOrder) // 添加采购订单
			v1.POST("/supplier_order/update", api.UpdateSupplierOrder) // 更新采购订单
			v1.POST("/supplier_order/delete", api.DeleteSupplierOrder) // 删除采购订单
			v1.POST("/supplier_order/detail", api.SupplierOrderDetail) // 采购订单详情
			//v1.POST("/supplier_order/customerprice", middleware.GetComIDAndModuleID(), api.CustomerPrice)
			v1.POST("/supplier_order/customer_purchase", api.AddCustomerPurchaseOrder) //客户订单/客户采购
			v1.POST("/supplier_order/purchase_order", api.AddPurchaseOrder)            //添加采购订单

			v1.POST("/customer_order/checksupplierprice", api.CheckSupplierPrice) //

			// 商品管理
			v1.POST("/product/list", api.AllProducts) // 商品列表
			v1.POST("/product/create", api.AddProduct) // 添加商品
			v1.POST("/product/update", api.UpdateProduct) // 更新商品
			v1.POST("/product/delete", api.DeleteProduct) // 删除商品
			v1.POST("/product/addprice", api.AddPrice) //添加供应商价格
			v1.POST("/product/detail", middleware.GetComIDAndModuleID(), api.ProductDetail) // 商品详情
			v1.POST("/product/supplierlist", api.SupplierListOfProducts)
			v1.GET("/product/img_load_sign", api.GetYpyunSign) //获取又拍云上传签名

			// 库存实例
			v1.GET("/wos_examples/wos_product", api.ProductWos) // 获取商品的库存
			v1.POST("/wos/add_product", api.CreateWosInstance)  // 补充商品
			v1.POST("/wos/wos_log", api.AllWosInstance)         // 获取库存实例日志
			// 客户价格管理(售价管理)
			v1.POST("/customer_price/list", api.ListCustomerPrice)
			v1.POST("/customer_price/create", api.AddCustomerPrice) // 添加采购价
			v1.POST("/customer_price/delete", api.DeleteCustomerPrice) // 删除采购价
			// 供应商价格管理(进价管理)
			v1.POST("/supplier_price/list", api.ListSupplierPrice)
			v1.POST("/supplier_price/create", api.AddSupplierPrice)
			v1.POST("/supplier_price/delete", api.DeleteSupplierPrice)

			// 客户结算管理
			v1.POST("/customer_settlement/list", api.ListCustomerSettlement)
			v1.POST("/customer_settlement/create", api.GenSettlement)
			//v1.POST("/auth_a",api.AuthA)
			v1.POST("/customer_settlement/getcustomer", api.FindSettlementCustomers)
			v1.POST("/customer_settlement/getsettlement", api.FindOneSettlements)

			v1.POST("/customer_settlement/detail", api.SettlementDetail)
			v1.POST("/customer_settlement/confirm", api.SettlementConfirm)
			v1.POST("/customer_settlement/download", api.CustomerSettlementDownload)

			// 供应商结算管理
			v1.POST("/supplier_settlement/list", api.ListSupplierSettlement)
			v1.POST("/supplier_settlement/create", api.GenSupSettlement)

			v1.POST("/supplier_settlement/getsupplier", api.FindSettlementSuppliers)
			v1.POST("/supplier_settlement/getsettlement", api.FindOneSupSettlements)

			v1.POST("/supplier_settlement/detail", api.SupSettlementDetail)
			v1.POST("/supplier_settlement/confirm", api.SupSettlementConfirm)
		}

	}

	return r
}
