package api

import (
	"github.com/gin-gonic/gin"
	"jxc/models"
)

func AuthA(c *gin.Context) {
	authNote := []models.AuthNote{
		{
			AuthId:  1,
			Note:    "系统设置",
			Group:   "管理",
			GroupId: 1,
			Urls: []string{
				"/api/v1/company/detail",
				"/api/v1/company/update",
			},
		},
		{
			AuthId:  2,
			Note:    "人员管理",
			Group:   "管理",
			GroupId: 1,
			Urls: []string{
				"/api/v1/employee/list",
				"/api/v1/employee/create",
				"/api/v1/employee/update",
				"/api/v1/employee/delete",
				"/api/v1/position/list",
				"/api/v1/auth/list",
				"/api/v1/employee/auth",
			},
		},
		{
			AuthId:  3,
			Note:    "商品管理",
			Group:   "管理",
			GroupId: 1,
			Urls: []string{
				"/api/v1/product/list",
				"/api/v1/product/create",
				"/api/v1/product/update",
				"/api/v1/product/delete",
				"/api/v1/product/addprice",
				"/api/v1/product/detail",
				"/api/v1/product/supplierlist",
				"/api/v1/upload_images",
				"/api/v1/units",
			},
		},
		{
			AuthId:  4,
			Note:    "仓库管理",
			Group:   "管理",
			GroupId: 1,
			Urls: []string{
				"/api/v1/warehouse/list",
				"/api/v1/warehouse/create",
				"/api/v1/warehouse/update",
				"/api/v1/warehouse/delete",
				"/api/v1/warehouse/detail",
				"/api/v1/wos/wos_log",
				"/api/v1/delivery",
				"/api/v1/customer_order/customer_order_shipped",
				"/api/v1/customer_order/customer_order_confirm",
				"/api/v1/customer_order/customer_order_check",
				"/api/v1/warehouse/detail",
			},
		},
		{
			AuthId:  5,
			Note:    "进价管理",
			Group:   "管理",
			GroupId: 1,
			Urls: []string{
				"/api/v1/supplier_price/list",
				"/api/v1/supplier/list",
				"/api/v1/supplier_price/create",
				"/api/v1/supplier_price/delete",
			},
		},
		{
			AuthId:  6,
			Note:    "售价管理",
			Group:   "管理",
			GroupId: 1,
			Urls: []string{
				"/api/v1/customer_price/list",
				"/api/v1/customer/list",
				"/api/v1/customer_price/create",
				"/api/v1/customer_price/delete",
			},
		},
		{
			AuthId:  7,
			Note:    "客户管理",
			Group:   "管理",
			GroupId: 1,
			Urls: []string{
				"/api/v1/customer/list",
				"/api/v1/customer/update",
				"/api/v1/customer/create",
				"/api/v1/customer/delete",
			},
		},
		{
			AuthId:  8,
			Note:    "供应商管理",
			Group:   "管理",
			GroupId: 1,
			Urls: []string{
				"/api/v1/supplier/list",
				"/api/v1/supplier/create",
				"/api/v1/supplier/update",
				"/api/v1/supplier/delete",
			},
		},
		{
			AuthId:  9,
			Note:    "查看系统设置",
			Group:   "查看",
			GroupId: 2,
			Urls: []string{
				"/api/v1/company/detail",
			},
		},
		{
			AuthId:  10,
			Note:    "查看人员",
			Group:   "查看",
			GroupId: 2,
			Urls: []string{
				"/api/v1/employee/list",
				"/api/v1/position/list",
				"/api/v1/auth/list",
			},
		},
		{
			AuthId:  11,
			Note:    "查看商品",
			Group:   "查看",
			GroupId: 2,
			Urls: []string{
				"/api/v1/product/list",
				"/api/v1/product/detail",
				"/api/v1/product/supplierlist",
				"/api/v1/units",
			},
		},
		{
			AuthId:  12,
			Note:    "查看进价",
			Group:   "查看",
			GroupId: 2,
			Urls: []string{
				"/api/v1/supplier_price/list",
				"/api/v1/supplier/list",
			},
		},
		{
			AuthId:  13,
			Note:    "查看售价",
			Group:   "查看",
			GroupId: 2,
			Urls: []string{
				"/api/v1/customer_price/list",
				"/api/v1/customer/list",
			},
		},
		{
			AuthId:  14,
			Note:    "查看销售订单",
			Group:   "查看",
			GroupId: 2,
			Urls: []string{
				"/api/v1/customer_order/list",
				"/api/v1/customer_order/sub_order_instance_list",
			},
		},
		{
			AuthId:  15,
			Note:    "查看采购订单",
			Group:   "查看",
			GroupId: 2,
			Urls: []string{
				"/api/v1/supplier_order/list",
			},
		},
		{
			AuthId:  16,
			Note:    "查看客户",
			Group:   "查看",
			GroupId: 2,
			Urls: []string{
				"/api/v1/customer/list",
			},
		},
		{
			AuthId:  17,
			Note:    "查看供应商",
			Group:   "查看",
			GroupId: 2,
			Urls: []string{
				"/api/v1/supplier/list",
			},
		},
		{
			AuthId:  18,
			Note:    "查看仓库",
			Group:   "查看",
			GroupId: 2,
			Urls: []string{
				"/api/v1/warehouse/list",
			},
		},
		{
			AuthId:  19,
			Note:    "创建采购订单",
			Group:   "采购",
			GroupId: 3,
			Urls: []string{
				"/api/v1/product/list",
				"/api/v1/supplier/list",
				"/api/v1/warehouse/list",
				"/api/v1/supplier_order/purchase_order",
			},
		},
		/*		{
				AuthId:   20,
				Note:     "查看进货价",
				Group:    "采购",
				GroupId:  3,
			},*/
		{
			AuthId:  21,
			Note:    "审核采购订单",
			Group:   "采购",
			GroupId: 3,
			Urls: []string{
				"/api/v1/supplier_order/supplier_order_check",
			},
		},
		{
			AuthId:  22,
			Note:    "结算采购订单",
			Group:   "采购",
			GroupId: 3,
			Urls: []string{
				"/api/v1/units",
			},
		},
		{
			AuthId:  23,
			Note:    "创建销售订单",
			Group:   "销售",
			GroupId: 4,
			Urls: []string{
				"/api/v1/product/list",
				"/api/v1/customer/list",
				"/api/v1/customer_order/checkcustomerprice",
				"/api/v1/customer_order/create",
			},
		},
		{
			AuthId:  24,
			Note:    "查看销售价格",
			Group:   "销售",
			GroupId: 4,
			Urls: []string{
				"/api/v1/units",
			},
		},
		{
			AuthId:  25,
			Note:    "结算销售订单",
			Group:   "销售",
			GroupId: 4,
			Urls: []string{
				"/api/v1/customer_settlement/list",
			},
		},
		{
			// 备货需要看到销售订单、仓库库存、采购价、已备货的实例
			AuthId:  26,
			Note:    "备货",
			Group:   "销售",
			GroupId: 4,
			Urls: []string{
				"/api/v1/supplier_order/list",
				"/api/v1/wos_examples/wos_product",
				"/api/v1/product/supplierlist",
				"/api/v1/customer_order/sub_order_instance_list",
				"/api/v1/supplier_order/customer_purchase",
			},
		},
		{
			// 确认销售订单需要看到销售订单列表、确认销售订单
			AuthId:  27,
			Note:    "确认销售订单",
			Group:   "销售",
			GroupId: 4,
			Urls: []string{
				"/api/v1/supplier_order/list",
				"/api/v1/supplier_order/supplier_order_confirm",
			},
		},
	}

	var authNotes []interface{}
	for _, val := range authNote {
		authNotes = append(authNotes, val)
	}

	var a = models.AuthNote{}
	a.InsertMany(authNotes)
}
