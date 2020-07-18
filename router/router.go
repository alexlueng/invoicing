package router

import (
	"github.com/gin-gonic/gin"
	"jxc/api"
	"jxc/api/wxapp"
	"jxc/middleware"
)

func InitRouter() *gin.Engine {
	gin.SetMode(gin.DebugMode)

	r := gin.Default()

	// 服务之间的通信，不需要权限验证
	r.POST("/add_superadmin", api.AddSuperAdmin)
	r.POST("/adminpasswd_update", api.UpdateAdminPasswd)
	r.POST("/changedomainstatus", api.ChangeDomainStatus)
	r.POST("/update_expire_time", api.UpdateExpireTime)
	r.POST("/getconfig", api.GetConfig) //页面信息配置

	r.GET("/wx/:filename", api.WechatVerify) // 微信验证接口

	// 中间件, 顺序不能改
	r.Use(middleware.Cors())      // 允许跨域
	r.Use(middleware.CheckAuth()) // 鉴权
	r.NoRoute(api.NotFound)       //设置默认路由当访问一个错误网站时返回

	//使用以下gin提供的Group函数为不同的API进行分组
	v1 := r.Group("/api/v1")
	{
		v1.POST("/login", api.Login)
		v1.POST("/supplier/mobile/login", api.SupplierLogin) // 供应商登录
		{
			// 供应商平台操作接口

			v1.POST("/supplier/mobile/product", api.SupplierProducts)                // 供应商产品
			v1.POST("/supplier/mobile/orders", api.SupplierOrders)                   // 供应商订单
			v1.POST("/supplier/mobile/exec_order", api.SupplierExecOrder)            // 供应商发货
			v1.POST("/supplier/mobile/settlements", api.SupplierSettlements)         // 供应商结算单
			v1.POST("/supplier/mobile/reset_password", api.SupplierResetPassword)    // 重置密码
			v1.POST("/supplier/mobile/upload_images", api.SupplierUploadCertificate) // 上传凭证

			// 图片处理接口
			v1.POST("/upload_images", api.UploadImages) // 上传图片
			v1.POST("/delete_images", api.DeleteImages) // 删除图片

			//系统设置
			v1.POST("/sysconfig/set_expire_date", middleware.GetComID(), api.SetSysExpireDate) // 设置系统发货提示时间
			v1.POST("/sysconfig/get_expire_date", middleware.GetComID(), api.GetExpireDate)    // 获取系统设置

			// 公司管理
			v1.GET("/company/detail", api.CompanyDetail)                     // 获取公司详细信息
			v1.POST("/company/update", api.UpdateCompany)                    // 更新公司信息
			v1.GET("/delivery", api.DeliveryList)                            // 获取配送方式列表
			v1.GET("/units", api.UnitsList)                                  // 获取计量单位
			v1.GET("/payment", api.PaymentList)                              // 获取配送结算方式
			v1.GET("/default_profit_margin", api.DefaultProfitMargin)        // 获取默认利润率
			v1.GET("/product_menu_name", api.GetProductMenuName)             // 获取菜单名字
			v1.POST("/set_product_menu_name", api.SetProductMenuName)        // 设置菜单名字
			v1.POST("/add_delivery", middleware.GetComID(), api.AddDelivery) // 增加配送方式

			// 人员管理
			v1.GET("/employee/list", api.AllEmployees)      // 获取人员列表
			v1.POST("/employee/list", api.AllEmployees)     // 获取人员列表
			v1.POST("/employee/create", api.AddEmployee)    // 添加人员
			v1.POST("/employee/update", api.UpdateEmployee) // 更新人员信息
			v1.POST("/employee/delete", api.DeleteEmployee) // 删除人员
			v1.GET("/position/list", api.AllPositions)      // 职位列表
			v1.GET("/auth/list", api.AllAuthNote)           // 权限列表
			v1.POST("/employee/auth", api.UpdateAuthority)  // 修改人员权限

			// 仓库管理
			v1.POST("/warehouse/list", middleware.GetComIDAndModuleID(), api.AllWarehouses) // 获取仓库列表
			v1.POST("/warehouse/create", api.AddWarehouse)                                  // 添加仓库
			v1.POST("/warehouse/update", api.UpdateWarehouse)                               // 更新仓库
			v1.POST("/warehouse/delete", api.DeleteWarehouse)                               // 删除仓库
			v1.POST("/warehouse/detail", api.WarehouseDetail)                               // 仓库详情
			//v1.POST("/warehouse/download", api.WarehouseDownload)                           // 库存导出

			// 客户管理
			v1.POST("/customer/list", api.ListCustomers)    // 客户列表
			v1.POST("/customer/create", api.AddCustomer)    // 添加客户
			v1.POST("/customer/update", api.UpdateCustomer) // 更新客户信息
			v1.POST("/customer/delete", api.DeleteCustomer) // 删除公司

			v1.POST("/customer/level/list", api.LevelList)     // 等级列表
			v1.POST("/customer/level/add", api.AddLevel)       // 增加等级
			v1.POST("/customer/level/update", api.UpdateLevel) // 更新等级
			v1.POST("/customer/level/delete", api.DeleteLevel) // 删除等级

			// 客户订单管理
			v1.POST("/customer_order/list", api.AllCustomerOrders)                                 //  获取所有客户订单列表
			v1.POST("/customer_order/create", api.AddCustomerOrder)                                // 创建客户订单
			v1.POST("/customer_order/update", api.UpdateCustomerOrder)                             // 修改客户订单
			v1.POST("/customer_order/delete", api.DeleteCustomerOrder)                             // 删除客户订单
			v1.POST("/customer_order/detail", api.CustomerOrderDetail)                             // 客户订单详情
			v1.POST("/customer_order/customerprice", api.CustomerPrice)                            //
			v1.POST("/customer_order/checkcustomerprice", api.CheckCustomerPrice)                  //
			v1.GET("/customer_order/sub_order_instance_list", api.AllCustomerSubOrderInstance)     //获取子订单实例列表
			v1.POST("/customer_order/customer_order_shipped", api.CustomerSubOrderInstanceShipped) //子订单实例发货
			v1.POST("/customer_order/customer_order_confirm", api.CustomerSubOrderInstanceConfirm) //客户子订单实例确认收货
			v1.POST("/customer_order/customer_order_check", api.CustomerSubOrderInstanceCheck)     //客户子订单实例审核通过
			v1.POST("/customer_order/isprepare", middleware.GetComID(), api.PrepareStock)          //客户子订单实例审核通过

			// 供应商管理
			v1.POST("/supplier/list", api.ListSuppliers)     // 供应商列表
			v1.POST("/supplier/create", api.AddSuppliers)    // 添加供应商
			v1.POST("/supplier/update", api.UpdateSuppliers) // 更新供应商信息
			v1.POST("/supplier/delete", api.DeleteSuppliers) // 删除供应商

			// 供应商订单管理
			v1.POST("/supplier_order/list", api.AllSupplierOrders)                                 // 采购订单列表
			v1.POST("/supplier_order/create", api.AddSupplierOrder)                                // 添加采购订单
			v1.POST("/supplier_order/update", api.UpdateSupplierOrder)                             // 更新采购订单
			v1.POST("/supplier_order/delete", api.DeleteSupplierOrder)                             // 删除采购订单
			v1.POST("/supplier_order/detail", api.SupplierOrderDetail)                             // 采购订单详情
			v1.POST("/supplier_order/customer_purchase", api.AddCustomerPurchaseOrder)             //客户订单/客户采购
			v1.POST("/supplier_order/purchase_order", api.AddPurchaseOrder)                        //添加采购订单
			v1.POST("/supplier_order/supplier_order_confirm", api.SupplierSubOrderInstanceConfirm) // 采购订单/确认收货
			v1.POST("/supplier_order/supplier_order_check", api.SupplierSubOrderInstanceCheck)     // 采购订单/审核
			v1.POST("/supplier_order/split_supplier_sub_order", api.SplitSupplierSubOrder)         // 部分接受订单

			// 商品管理
			v1.POST("/product/list", api.AllProducts)                                       // 商品列表
			v1.POST("/product/create", api.AddProduct)                                      // 添加商品
			v1.POST("/product/update", api.UpdateProduct)                                   // 更新商品
			v1.POST("/product/delete", api.DeleteProduct)                                   // 删除商品
			v1.POST("/product/addprice", api.AddPrice)                                      //添加供应商价格
			v1.POST("/product/detail", middleware.GetComIDAndModuleID(), api.ProductDetail) // 商品详情
			v1.POST("/product/supplierlist", api.SupplierListOfProducts)
			v1.GET("/product/img_load_sign", api.GetYpyunSign)  //获取又拍云上传签名
			v1.POST("/product/preferred", api.PreferredProduct) // 设置为优选商品
			v1.POST("/product/recommand", api.RecommandProduct) // 设置为推荐商品

			// 商品分类
			v1.POST("/product/category/create", api.AddCategory)
			v1.POST("/product/category/list", api.ListCategory)
			v1.POST("/product/category/find_category", api.FindOneCategory)
			v1.POST("/product/category/delete", api.DeleteCategory)
			v1.POST("/product/category/update", api.UpdateCategory)

			// 库存实例
			v1.POST("/stock/product_stocks", api.GetProductStock) // 获取商品的库存 ProductStocks
			v1.POST("/wos/add_product", api.CreateWosInstance)    // 补充商品
			v1.POST("/wos/wos_log", api.AllWosInstance)           // 获取库存实例日志

			// 客户价格管理(售价管理)
			v1.POST("/customer_price/list", api.ListCustomerPrice)
			v1.POST("/customer_price/create", api.AddCustomerPrice)    // 添加采购价
			v1.POST("/customer_price/delete", api.DeleteCustomerPrice) // 删除采购价

			// 供应商价格管理(进价管理)
			v1.POST("/supplier_price/list", api.ListSupplierPrice)
			v1.POST("/supplier_price/create", api.AddSupplierPrice)
			v1.POST("/supplier_price/delete", api.DeleteSupplierPrice)

			// 客户结算管理
			v1.POST("/customer_settlement/list", api.ListCustomerSettlement)
			v1.POST("/customer_settlement/create", api.GenSettlement)
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
			v1.POST("/supplier_settlement/download", api.SupplierSettlementDownload)

			// 用户提醒
			message := v1.Group("/message")
			{
				message.POST("/list", api.MessageList)
				message.POST("/read", api.ReadMessage)
				message.POST("/find_expire_orders", api.OrderMessages)
			}

			// 微信小程序接口，商品列表，购物车，分类，
			wxApp := v1.Group("/wxapp")
			{
				wxApp.POST("/login", wxapp.Login)                           // 登录
				wxApp.POST("/user_info", wxapp.GetUserInfo)                 // 得到用户信息
				wxApp.POST("/category", wxapp.CategoryList)                 // 分类
				wxApp.POST("/recommandproduct", wxapp.RecommandProductList) // 首页推荐
				wxApp.POST("/preferredproduct", wxapp.PreferredProductList) // 首页优选
				wxApp.POST("/cat_products", wxapp.CatProducts)              // 分类商品
				wxApp.POST("/verify", wxapp.UserVerify)                     // 用户上传验证资料
				wxApp.POST("/cat_product", wxapp.ProductListByCategoryID)   // 根据分类id获取商品列表
				wxApp.POST("/product_detail", wxapp.ProductDetail)          // 商品详情
				wxApp.POST("/add_cart", wxapp.AddToCart)                    // 添加购物车
				wxApp.POST("/remove_cart", wxapp.RemoveFromCart)            // 购物车移除商品
				wxApp.POST("/clear_cart", wxapp.ClearCart)                  // 清空购物车
				wxApp.POST("/list_cart", wxapp.ListCart)                    // 购物车列表
				wxApp.POST("/delete_cart", wxapp.DeleteCartItem)            // 删除购物车项
				wxApp.POST("/summit_order", wxapp.SummitOrder)              // 提交订单
				wxApp.POST("/address", wxapp.UserAddresses)                 // 用户地址

			}

			// 后台微信设置接口
			wechat := v1.Group("/wechat")
			{
				wechat.GET("/upload_path", api.FileUploadPath)
				wechat.POST("/upload", api.UploadVerifyFile)
			}

			// 微信支付接口

		}

	}

	return r
}
