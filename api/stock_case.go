package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
	"jxc/serializer"
	"net/http"
	"strings"
)


// 商品库存列表
func ProductStock2(c *gin.Context) {
	// 根据域名获取comid
	com, err := models.GetComIDAndModuleByDomain(strings.Split(c.Request.RemoteAddr, ":")[0])
	//moduleID, _ := strconv.Atoi(com.ModuleId)
	if err != nil || models.THIS_MODULE != com.ModuleId {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "域名错误",
		})
		return
	}

	req := ReqProductStock{}
	// 验证提交过来的数据
	err = c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	// 指定数据集
	collection := models.Client.Collection("wos_examples")

	// 查询条件
	// com_id = com_id,productID in req.Products
	filter := bson.M{}
	filter["com_id"] = com.ComId
	filter["product_id"] = bson.M{"$in": req.Products}

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return
	}

	// 返回的数据结构
	// 商品id ： {仓库id：库存},总库存

	// 统计仓库间的库存
	var wosExamples models.WosInstance

	// 库存分布
	type warehouseStock struct {
		// 仓库id
		WarehouseId int64 `json:"warehouse_id"`
		// 商品库存
		ProductNum int64 `json:"product_num"`
	}
	// 商品库存
	type ProductStock struct {
		// 商品id
		ProductID int64 `json:"product_id"`
		// 库存
		Stock int64 `json:"stock"`
		// 库存分布
		WarehouseStock []warehouseStock `json:"warehouse_stock"`
	}
	type StockA struct {
		ProductID int64 `json:"product_id"` // 商品id
		Num       int64 `json:"num"`        // 库存
	}

	// 各商品数量 商品id
	productStock := make(map[int64]StockA)
	var warehouse []int64

	// 各仓库商品数量 仓库id 商品id
	wosProduct1 := make(map[int64]StockA)
	wosProduct2 := make(map[int64]map[int64]StockA)


	// 商品id ： {商品id： ，数量：}

	// 共查出多少个仓库

	// 从数据中拼出数据格式 仓库id : 商品，数量

	for cur.Next(context.TODO()) {
		err = cur.Decode(&wosExamples)
		if err != nil {
			return
		}
		//warehouse = append(warehouse, stockCase.WarehouseID)
		//product = append(product, stockCase.ProductID)

		//

		// 统计各商品数量
		switch wosExamples.Type {
		case 0: // 凭空多出的数量 +
			if wosProduct2[wosExamples.WarehouseID] ==  nil{
				wosProduct1[wosExamples.ProductID] = StockA{
					ProductID: wosExamples.ProductID,
					Num:       0,
				}
				wosProduct2[wosExamples.WarehouseID] = wosProduct1
			}
			// 统计各仓库商品库存
			wosProduct2[wosExamples.WarehouseID][wosExamples.ProductID] = StockA{
				ProductID: wosExamples.ProductID,
				Num:       wosExamples.ProductNum + wosProduct2[wosExamples.WarehouseID][wosExamples.ProductID].Num,
			}
			// 统计各商品数量
			productStock[wosExamples.ProductID] = StockA{
				ProductID: wosExamples.ProductID,
				Num:       wosExamples.ProductNum + productStock[wosExamples.ProductID].Num,
			}
			break
		case 1: // 退货多出的数量 +
			if wosProduct2[wosExamples.WarehouseID] ==  nil{
				wosProduct1[wosExamples.ProductID] = StockA{
					ProductID: wosExamples.ProductID,
					Num:       0,
				}
				wosProduct2[wosExamples.WarehouseID] = wosProduct1
			}
			// 统计各仓库商品库存
			wosProduct2[wosExamples.WarehouseID][wosExamples.ProductID] = StockA{
				ProductID: wosExamples.ProductID,
				Num:       wosExamples.ProductNum + wosProduct2[wosExamples.WarehouseID][wosExamples.ProductID].Num,
			}

			wosProduct2[wosExamples.WarehouseID][wosExamples.ProductID] = StockA{
				ProductID: wosExamples.ProductID,
				Num:       wosExamples.ProductNum + productStock[wosExamples.ProductID].Num,
			}

			productStock[wosExamples.ProductID] = StockA{
				ProductID: wosExamples.ProductID,
				Num:       wosExamples.ProductNum + productStock[wosExamples.ProductID].Num,
			}
			break
		case 2: // 销售减少的数量 -
			if wosProduct2[wosExamples.WarehouseID] ==  nil{
				wosProduct1[wosExamples.ProductID] = StockA{
					ProductID: wosExamples.ProductID,
					Num:       0,
				}
				wosProduct2[wosExamples.WarehouseID] = wosProduct1
			}
			// 统计各仓库商品库存
			wosProduct2[wosExamples.WarehouseID][wosExamples.ProductID] = StockA{
				ProductID: wosExamples.ProductID,
				Num:       wosProduct2[wosExamples.WarehouseID][wosExamples.ProductID].Num - wosExamples.ProductNum,
			}

			productStock[wosExamples.ProductID] = StockA{
				ProductID: wosExamples.ProductID,
				Num:       productStock[wosExamples.ProductID].Num - wosExamples.ProductNum,
			}
			break
		case 3: // 损耗减少的数量 -
			if wosProduct2[wosExamples.WarehouseID] ==  nil{
				wosProduct1[wosExamples.ProductID] = StockA{
					ProductID: wosExamples.ProductID,
					Num:       0,
				}
				wosProduct2[wosExamples.WarehouseID] = wosProduct1
			}
			// 统计各仓库商品库存
			wosProduct2[wosExamples.WarehouseID][wosExamples.ProductID] = StockA{
				ProductID: wosExamples.ProductID,
				Num:       wosProduct2[wosExamples.WarehouseID][wosExamples.ProductID].Num - wosExamples.ProductNum,
			}

			productStock[wosExamples.ProductID] = StockA{
				ProductID: wosExamples.ProductID,
				Num:       productStock[wosExamples.ProductID].Num - wosExamples.ProductNum,
			}
			break
		case 4: // 采购多出的数量 +
			if wosProduct2[wosExamples.WarehouseID] == nil{
				wosProduct1[wosExamples.ProductID] = StockA{
					ProductID: wosExamples.ProductID,
					Num:       0,
				}
				wosProduct2[wosExamples.WarehouseID] = wosProduct1
			}
			// 统计各仓库商品库存
			wosProduct2[wosExamples.WarehouseID][wosExamples.ProductID] = StockA{
				ProductID: wosExamples.ProductID,
				Num:       wosProduct2[wosExamples.WarehouseID][wosExamples.ProductID].Num + wosExamples.ProductNum,
			}

			productStock[wosExamples.ProductID] = StockA{
				ProductID: wosExamples.ProductID,
				Num:       wosExamples.ProductNum + productStock[wosExamples.ProductID].Num,
			}
			break
		default:
			return
		}

		// 统计各商品在仓库中的分布

	}

	/*

	 */

	fmt.Println(warehouse)

	// 如果有商品没有库存也在结果提现

	// 返回所有商品数量
}


// 模拟修改库存实例
func ProductStockA(c *gin.Context) {
	stockCase := []models.WosInstance{
		{
			ComID:            1,
			Type:             0,
			WarehouseID:      10,
			PurchaseOrderSn:  "",
			SalesOrderSn:     "",
			ProductID:        2,
			ProductNum:       100,
			//ProductUnitPrice: "",
			CreateAt:         1582096973,
			CreateBy:         1,
			ShipTime:         0,
			Shipper:          0,
			ConfirmAt:        0,
			ConfirmBy:        0,
			CheckAt:          0,
			CheckBy:          0,
		},
		{
			ComID:            1,
			Type:             1,
			WarehouseID:      10,
			PurchaseOrderSn:  "",
			SalesOrderSn:     "",
			ProductID:        2,
			ProductNum:       50,
			//ProductUnitPrice: "",
			CreateAt:         1582096973,
			CreateBy:         1,
			ShipTime:         0,
			Shipper:          0,
			ConfirmAt:        0,
			ConfirmBy:        0,
			CheckAt:          0,
			CheckBy:          0,
		},
		{
			ComID:            1,
			Type:             2,
			WarehouseID:      10,
			PurchaseOrderSn:  "",
			SalesOrderSn:     "",
			ProductID:        2,
			ProductNum:       10,
			//ProductUnitPrice: "",
			CreateAt:         1582096973,
			CreateBy:         1,
			ShipTime:         0,
			Shipper:          0,
			ConfirmAt:        0,
			ConfirmBy:        0,
			CheckAt:          0,
			CheckBy:          0,
		},
		{
			ComID:            1,
			Type:             3,
			WarehouseID:      10,
			PurchaseOrderSn:  "",
			SalesOrderSn:     "",
			ProductID:        2,
			ProductNum:       20,
			//ProductUnitPrice: "",
			CreateAt:         1582096973,
			CreateBy:         1,
			ShipTime:         0,
			Shipper:          0,
			ConfirmAt:        0,
			ConfirmBy:        0,
			CheckAt:          0,
			CheckBy:          0,
		},
		{
			ComID:            1,
			Type:             4,
			WarehouseID:      10,
			PurchaseOrderSn:  "",
			SalesOrderSn:     "",
			ProductID:        2,
			ProductNum:       30,
			//ProductUnitPrice: "",
			CreateAt:         1582096973,
			CreateBy:         1,
			ShipTime:         0,
			Shipper:          0,
			ConfirmAt:        0,
			ConfirmBy:        0,
			CheckAt:          0,
			CheckBy:          0,
		},
	}

	collection := models.Client.Collection("wos_examples")
	var stockCases []interface{}
	for _, val := range stockCase {
		stockCases = append(stockCases, val)
	}

	collection.InsertMany(context.TODO(), stockCases)
}

// 库存管理，盘点（凭空多出的、损耗的）需要人工操作

// 销售自动进行

// 退货自动进行
