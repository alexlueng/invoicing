package service

import (
	"context"
	//"errors"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
	//"time"
)

// 仓库中的存储情况
type Stock struct {
	WarehouseID        int64  `json:"warehouse_id"`         // 仓库id
	WarehouseName      string `json:"warehuose_name"`       // 仓库名
	WarehouseAddress   string `json:"warehouse_address"`    // 仓库地址
	WarehouseAdminName string `json:"warehouse_admin_name"` // 仓库管理员
	ProductID          int64  `json:"product_id"`           // 商品id
	ProductName        string `json:"product_name"`         // 商品名
	Units              string `json:"units"`                // 商品量词
	Num                int64  `json:"num"`                  // 库存
	InWarehouse        int64  `json:"in_warehouse"`         // 已在库
	NotInWarehouse     int64  `json:"not_in_warehouse"`     // 未入库
	Shipped            int64  `json:"shipped"`              // 出库
	NotShipped         int64  `json:"not_shipped"`          // 未发货
}

// 库存商品数据格式
type WarehouseProduct struct {
	ProductID   int64           `json:"product_id"`   // 商品id
	ProductName string          `json:"product_name"` // 商品名
	Num         int64           `json:"num"`          // 库存 （已在库+未入库）
	Stock       map[int64]Stock `json:"stock"`
}

// 创建库存实例提交的数据
type WarehouseData struct {
	Type             int64  `json:"type" form:"type"`                             // 记录类型
	Product          int64  `json:"product" form:"product"`                       // 商品id
	ProductUnitPrice string `json:"product_unit_price" form:"product_unit_price"` // 商品单价
	Num              int64  `json:"num" form:"num"`                               // 商品数量
	WarehouseId      int64  `json:"warehouse_id" form:"warehouse_id"`             // 仓库id
	OrderSn          string `json:"order_sn" form:"order_sn"`                     // 订单号
}

//
/**
* 获取商品库存
* @param product_ids []int64 商品数组
* @param comd_id int64 公司id
* @param warehouse_id int64 仓库id ，0 则是所有仓库
* @return map[商品id]WosProduct，error
* @deprecated
 */
func GetProductInfoOfWarehouse(product_id int64, com_id int64, warehouse_id int64) (map[int64]WarehouseProduct, error) {
	// 商品的直接统计信息放在这 map[product_id]WosProduct
	productCount := make(map[int64]WarehouseProduct)

	var instance models.GoodsInstance
	collection := models.Client.Collection("goods_instance")
	// com_id = com_id  src_type = 3 or dest_type = 3,

	filter := bson.M{}
	filter["com_id"] = com_id
	//filter["place_id"] = warehouse_id
	filter["product_id"] = product_id
	filter["place_type"] = bson.M{"$ne": 0}

	// 获取仓库信息
	warehouse, err := FindWarehouse(warehouse_id, com_id)

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {
		err = cur.Decode(&instance)
		if err != nil {
			continue
		}
		// 如果是第一次出现这个仓库，则在productCount 分配空间
		if productCount[instance.ProductID].Stock == nil {
			productCount[instance.ProductID] = WarehouseProduct{
				ProductID:   instance.ProductID,
				ProductName: instance.Product,
				Num:         productCount[instance.ProductID].Num,
				Stock: map[int64]Stock{
					instance.PlaceId: {
						WarehouseID: instance.PlaceId,
						ProductID:   instance.ProductID,
						Units:       "计量单位",
						ProductName: instance.Product,
					},
				},
			}
		}

		switch instance.PlaceType {
		case 1: // 销售-待发货
			productCount[instance.ProductID] = WarehouseProduct{
				ProductID:   instance.ProductID,
				ProductName: instance.Product,
				Num:         productCount[instance.ProductID].Num - instance.Amount,
				Stock:       productCount[instance.ProductID].Stock,
			}
			productCount[instance.ProductID].Stock[instance.PlaceId] = Stock{
				WarehouseID:        instance.PlaceId,
				WarehouseName:      instance.SrcTitle,
				WarehouseAddress:   warehouse[instance.PlaceId].Address,
				WarehouseAdminName: warehouse[instance.PlaceId].WarehouseAdminName,
				ProductID:          instance.ProductID,
				ProductName:        instance.Product,
				Units:              instance.Units,
				Num:                productCount[instance.ProductID].Stock[instance.PlaceId].Num - instance.Amount,         // 总库存
				InWarehouse:        productCount[instance.ProductID].Stock[instance.PlaceId].InWarehouse - instance.Amount, // 已在库
				NotInWarehouse:     productCount[instance.ProductID].Stock[instance.PlaceId].NotInWarehouse,                // 已入库
				Shipped:            productCount[instance.ProductID].Stock[instance.PlaceId].Shipped,                       // 已发货
				NotShipped:         productCount[instance.ProductID].Stock[instance.PlaceId].NotShipped + instance.Amount,  // 未发货
			}
			break
		case 2: // 销售-已发货
			productCount[instance.ProductID] = WarehouseProduct{
				ProductID:   instance.ProductID,
				ProductName: instance.Product,
				Num:         productCount[instance.ProductID].Num,
				Stock:       productCount[instance.ProductID].Stock,
			}
			productCount[instance.ProductID].Stock[instance.PlaceId] = Stock{
				WarehouseID:        instance.PlaceId,
				WarehouseName:      instance.SrcTitle,
				WarehouseAddress:   warehouse[instance.PlaceId].Address,
				WarehouseAdminName: warehouse[instance.PlaceId].WarehouseAdminName,
				ProductID:          instance.ProductID,
				ProductName:        instance.Product,
				Units:              instance.Units,
				Num:                productCount[instance.ProductID].Stock[instance.PlaceId].Num,                       // 总库存
				InWarehouse:        productCount[instance.ProductID].Stock[instance.PlaceId].InWarehouse,               // 已在库
				NotInWarehouse:     productCount[instance.ProductID].Stock[instance.PlaceId].NotInWarehouse,            // 已入库
				Shipped:            productCount[instance.ProductID].Stock[instance.PlaceId].Shipped + instance.Amount, // 已发货
				NotShipped:         productCount[instance.ProductID].Stock[instance.PlaceId].NotShipped,                // 未发货
			}

			break
		case 3: // 销售-确认收货

			break
		case 4: // 采购-待收货
			productCount[instance.ProductID] = WarehouseProduct{
				ProductID:   instance.ProductID,
				ProductName: instance.Product,
				Num:         productCount[instance.ProductID].Num + instance.Amount,
				Stock:       productCount[instance.ProductID].Stock,
			}
			productCount[instance.ProductID].Stock[instance.PlaceId] = Stock{
				WarehouseID:        instance.PlaceId,
				WarehouseName:      instance.DestTitle,
				WarehouseAddress:   warehouse[instance.PlaceId].Address,
				WarehouseAdminName: warehouse[instance.PlaceId].WarehouseAdminName,
				ProductID:          instance.ProductID,
				ProductName:        instance.Product,
				Units:              instance.Units,
				Num:                productCount[instance.ProductID].Stock[instance.PlaceId].Num + instance.Amount,            // 总库存
				InWarehouse:        productCount[instance.ProductID].Stock[instance.PlaceId].InWarehouse,                      // 已在库
				NotInWarehouse:     productCount[instance.ProductID].Stock[instance.PlaceId].NotInWarehouse + instance.Amount, // 已入库
				Shipped:            productCount[instance.ProductID].Stock[instance.PlaceId].Shipped,                          // 已发货
				NotShipped:         productCount[instance.ProductID].Stock[instance.PlaceId].NotShipped,                       // 未发货
			}
			break
		case 5: // 采购-已收货
			productCount[instance.ProductID] = WarehouseProduct{
				ProductID:   instance.ProductID,
				ProductName: instance.Product,
				Num:         productCount[instance.ProductID].Num + instance.Amount,
				Stock:       productCount[instance.ProductID].Stock,
			}
			productCount[instance.ProductID].Stock[instance.PlaceId] = Stock{
				WarehouseID:        instance.PlaceId,
				WarehouseName:      instance.DestTitle,
				WarehouseAddress:   warehouse[instance.PlaceId].Address,
				WarehouseAdminName: warehouse[instance.PlaceId].WarehouseAdminName,
				ProductID:          instance.ProductID,
				ProductName:        instance.Product,
				Units:              instance.Units,
				Num:                productCount[instance.ProductID].Stock[instance.PlaceId].Num + instance.Amount,         // 总库存
				InWarehouse:        productCount[instance.ProductID].Stock[instance.PlaceId].InWarehouse + instance.Amount, // 已在库
				NotInWarehouse:     productCount[instance.ProductID].Stock[instance.PlaceId].NotInWarehouse,                // 已入库
				Shipped:            productCount[instance.ProductID].Stock[instance.PlaceId].Shipped,                       // 已发货
				NotShipped:         productCount[instance.ProductID].Stock[instance.PlaceId].NotShipped,                    // 未发货
			}
			break
		case 6: // 无端损耗
			productCount[instance.ProductID] = WarehouseProduct{
				ProductID:   instance.ProductID,
				ProductName: instance.Product,
				Num:         productCount[instance.ProductID].Num - instance.Amount,
				Stock:       productCount[instance.ProductID].Stock,
			}
			productCount[instance.ProductID].Stock[instance.PlaceId] = Stock{
				WarehouseID:        instance.PlaceId,
				WarehouseName:      instance.SrcTitle,
				WarehouseAddress:   warehouse[instance.PlaceId].Address,
				WarehouseAdminName: warehouse[instance.PlaceId].WarehouseAdminName,
				ProductID:          instance.ProductID,
				ProductName:        instance.Product,
				Units:              instance.Units,
				Num:                productCount[instance.ProductID].Stock[instance.PlaceId].Num - instance.Amount,         // 总库存
				InWarehouse:        productCount[instance.ProductID].Stock[instance.PlaceId].InWarehouse - instance.Amount, // 已在库
				NotInWarehouse:     productCount[instance.ProductID].Stock[instance.PlaceId].NotInWarehouse,                // 已入库
				Shipped:            productCount[instance.ProductID].Stock[instance.PlaceId].Shipped,                       // 已发货
				NotShipped:         productCount[instance.ProductID].Stock[instance.PlaceId].NotShipped,                    // 未发货
			}
			break
		case 7: // 凭空增加
			productCount[instance.ProductID] = WarehouseProduct{
				ProductID:   instance.ProductID,
				ProductName: instance.Product,
				Num:         productCount[instance.ProductID].Num + instance.Amount,
				Stock:       productCount[instance.ProductID].Stock,
			}
			productCount[instance.ProductID].Stock[instance.PlaceId] = Stock{
				WarehouseID:        instance.PlaceId,
				WarehouseName:      instance.DestTitle,
				WarehouseAddress:   warehouse[instance.PlaceId].Address,
				WarehouseAdminName: warehouse[instance.PlaceId].WarehouseAdminName,
				ProductID:          instance.ProductID,
				ProductName:        instance.Product,
				Units:              instance.Units,
				Num:                productCount[instance.ProductID].Stock[instance.PlaceId].Num + instance.Amount,         // 总库存
				InWarehouse:        productCount[instance.ProductID].Stock[instance.PlaceId].InWarehouse + instance.Amount, // 已在库
				NotInWarehouse:     productCount[instance.ProductID].Stock[instance.PlaceId].NotInWarehouse,                // 已入库
				Shipped:            productCount[instance.ProductID].Stock[instance.PlaceId].Shipped,                       // 已发货
				NotShipped:         productCount[instance.ProductID].Stock[instance.PlaceId].NotShipped,                    // 未发货
			}
			break

		}
	}

	return productCount, nil
}
