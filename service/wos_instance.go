package service

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
	"time"
)

// 仓库中的存储情况
type Wos struct {
	WosID              int64  `json:"wos_id"`               //仓库id
	Wos                string `json:"wos"`                  //仓库名
	WosAddress         string `json:"wos_address"`          //仓库地址
	WarehouseAdminName string `json:"warehouse_admin_name"` //仓库管理员
	ProductID          int64  `json:"product_id"`           // 商品id
	ProductName        string `json:"product_name"`         //商品名
	Units              string `json:"units"`                //商品量词
	Num                int64  `json:"num"`                  // 库存

	InWos      int64 `json:"in_wos"`      // 已在库
	NotWosed   int64 `json:"not_wosed"`   // 未入库
	Shipped    int64 `json:"shipped"`     // 出库
	NotShipped int64 `json:"not_shipped"` // 未发货
}

// 库存商品数据格式
type WosProduct struct {
	ProductID   int64         `json:"product_id"`   // 商品id
	ProductName string        `json:"product_name"` // 商品名
	Num         int64         `json:"num"`          // 库存 （已在库+未入库）
	Wos         map[int64]Wos `json:"wos"`
	//
}

// 创建库存实例提交的数据
type WosExamplesData struct {
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
func GetProductWos(product_ids []int64, com_id int64, warehouse_ids []int64) (map[int64]WosProduct, error) {
	//var wosInstance models.GoodsInstance
	// 商品的直接统计信息放在这 map[product_id]WosProduct
	productCount := make(map[int64]WosProduct)

	var instance models.GoodsInstance
	// 指定数据集
	collection := models.Client.Collection("goods_instance")
	// 查询条件
	// com_id = com_id  src_type = 3 or dest_type = 3,
	filter := bson.M{}
	if warehouse_ids != nil {
		filter["place_id"] = bson.M{"$in": warehouse_ids}
	}
	// 获取仓库信息
	warehouse, err := FindWarehouse(warehouse_ids, com_id)

	filter["com_id"] = com_id
	if len(product_ids) > 0 {
		filter["product_id"] = bson.M{"$in": product_ids}
	}

	filter["place_type"] = bson.M{"$ne": 0}
	//filter["$or"] = []bson.M{bson.M{"src_type":3},bson.M{"dest_type":3}}

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
		if productCount[instance.ProductID].Wos == nil {
			productCount[instance.ProductID] = WosProduct{
				ProductID:   instance.ProductID,
				ProductName: instance.Product,
				Num:         productCount[instance.ProductID].Num,
				Wos: map[int64]Wos{
					instance.PlaceId: {
						WosID:              instance.PlaceId,
						Wos:                "",
						WarehouseAdminName: "",
						WosAddress:         "",
						ProductID:          instance.ProductID,
						Units:              "计量单位",
						ProductName:        instance.Product,
						Num:                0,
						InWos:              0,
						NotWosed:           0,
						Shipped:            0,
						NotShipped:         0,
					},
				},
			}
		}

		switch instance.PlaceType {
		case 1: // 销售-待发货
			productCount[instance.ProductID] = WosProduct{
				ProductID:   instance.ProductID,
				ProductName: instance.Product,
				Num:         productCount[instance.ProductID].Num - instance.Amount,
				Wos:         productCount[instance.ProductID].Wos,
			}
			productCount[instance.ProductID].Wos[instance.PlaceId] = Wos{
				WosID:              instance.PlaceId,
				Wos:                instance.SrcTitle,
				WosAddress:         warehouse[instance.PlaceId].Address,
				WarehouseAdminName: warehouse[instance.PlaceId].WarehouseAdminName,
				ProductID:          instance.ProductID,
				ProductName:        instance.Product,
				Units:              instance.Units,
				Num:                productCount[instance.ProductID].Wos[instance.PlaceId].Num - instance.Amount,        // 总库存
				InWos:              productCount[instance.ProductID].Wos[instance.PlaceId].InWos - instance.Amount,      // 已在库
				NotWosed:           productCount[instance.ProductID].Wos[instance.PlaceId].NotWosed,                     // 已入库
				Shipped:            productCount[instance.ProductID].Wos[instance.PlaceId].Shipped,                      // 已发货
				NotShipped:         productCount[instance.ProductID].Wos[instance.PlaceId].NotShipped + instance.Amount, // 未发货
			}
			break
		case 2: // 销售-已发货
			productCount[instance.ProductID] = WosProduct{
				ProductID:   instance.ProductID,
				ProductName: instance.Product,
				Num:         productCount[instance.ProductID].Num,
				Wos:         productCount[instance.ProductID].Wos,
			}
			productCount[instance.ProductID].Wos[instance.PlaceId] = Wos{
				WosID:              instance.PlaceId,
				Wos:                instance.SrcTitle,
				WosAddress:         warehouse[instance.PlaceId].Address,
				WarehouseAdminName: warehouse[instance.PlaceId].WarehouseAdminName,
				ProductID:          instance.ProductID,
				ProductName:        instance.Product,
				Units:              instance.Units,
				Num:                productCount[instance.ProductID].Wos[instance.PlaceId].Num,                       // 总库存
				InWos:              productCount[instance.ProductID].Wos[instance.PlaceId].InWos,                     // 已在库
				NotWosed:           productCount[instance.ProductID].Wos[instance.PlaceId].NotWosed,                  // 已入库
				Shipped:            productCount[instance.ProductID].Wos[instance.PlaceId].Shipped + instance.Amount, // 已发货
				NotShipped:         productCount[instance.ProductID].Wos[instance.PlaceId].NotShipped,                // 未发货
			}

			break
		case 3: // 销售-确认收货

			break
		case 4: // 采购-待收货
			productCount[instance.ProductID] = WosProduct{
				ProductID:   instance.ProductID,
				ProductName: instance.Product,
				Num:         productCount[instance.ProductID].Num + instance.Amount,
				Wos:         productCount[instance.ProductID].Wos,
			}
			productCount[instance.ProductID].Wos[instance.PlaceId] = Wos{
				WosID:              instance.PlaceId,
				Wos:                instance.DestTitle,
				WosAddress:         warehouse[instance.PlaceId].Address,
				WarehouseAdminName: warehouse[instance.PlaceId].WarehouseAdminName,
				ProductID:          instance.ProductID,
				ProductName:        instance.Product,
				Units:              instance.Units,
				Num:                productCount[instance.ProductID].Wos[instance.PlaceId].Num + instance.Amount,      // 总库存
				InWos:              productCount[instance.ProductID].Wos[instance.PlaceId].InWos,                      // 已在库
				NotWosed:           productCount[instance.ProductID].Wos[instance.PlaceId].NotWosed + instance.Amount, // 已入库
				Shipped:            productCount[instance.ProductID].Wos[instance.PlaceId].Shipped,                    // 已发货
				NotShipped:         productCount[instance.ProductID].Wos[instance.PlaceId].NotShipped,                 // 未发货
			}
			break
		case 5: // 采购-已收货
			productCount[instance.ProductID] = WosProduct{
				ProductID:   instance.ProductID,
				ProductName: instance.Product,
				Num:         productCount[instance.ProductID].Num + instance.Amount,
				Wos:         productCount[instance.ProductID].Wos,
			}
			productCount[instance.ProductID].Wos[instance.PlaceId] = Wos{
				WosID:              instance.PlaceId,
				Wos:                instance.DestTitle,
				WosAddress:         warehouse[instance.PlaceId].Address,
				WarehouseAdminName: warehouse[instance.PlaceId].WarehouseAdminName,
				ProductID:          instance.ProductID,
				ProductName:        instance.Product,
				Units:              instance.Units,
				Num:                productCount[instance.ProductID].Wos[instance.PlaceId].Num + instance.Amount,   // 总库存
				InWos:              productCount[instance.ProductID].Wos[instance.PlaceId].InWos + instance.Amount, // 已在库
				NotWosed:           productCount[instance.ProductID].Wos[instance.PlaceId].NotWosed,                // 已入库
				Shipped:            productCount[instance.ProductID].Wos[instance.PlaceId].Shipped,                 // 已发货
				NotShipped:         productCount[instance.ProductID].Wos[instance.PlaceId].NotShipped,              // 未发货
			}
			break
		case 6: // 无端损耗
			productCount[instance.ProductID] = WosProduct{
				ProductID:   instance.ProductID,
				ProductName: instance.Product,
				Num:         productCount[instance.ProductID].Num - instance.Amount,
				Wos:         productCount[instance.ProductID].Wos,
			}
			productCount[instance.ProductID].Wos[instance.PlaceId] = Wos{
				WosID:              instance.PlaceId,
				Wos:                instance.SrcTitle,
				WosAddress:         warehouse[instance.PlaceId].Address,
				WarehouseAdminName: warehouse[instance.PlaceId].WarehouseAdminName,
				ProductID:          instance.ProductID,
				ProductName:        instance.Product,
				Units:              instance.Units,
				Num:                productCount[instance.ProductID].Wos[instance.PlaceId].Num - instance.Amount,   // 总库存
				InWos:              productCount[instance.ProductID].Wos[instance.PlaceId].InWos - instance.Amount, // 已在库
				NotWosed:           productCount[instance.ProductID].Wos[instance.PlaceId].NotWosed,                // 已入库
				Shipped:            productCount[instance.ProductID].Wos[instance.PlaceId].Shipped,                 // 已发货
				NotShipped:         productCount[instance.ProductID].Wos[instance.PlaceId].NotShipped,              // 未发货
			}
			break
		case 7: // 凭空增加
			productCount[instance.ProductID] = WosProduct{
				ProductID:   instance.ProductID,
				ProductName: instance.Product,
				Num:         productCount[instance.ProductID].Num + instance.Amount,
				Wos:         productCount[instance.ProductID].Wos,
			}
			productCount[instance.ProductID].Wos[instance.PlaceId] = Wos{
				WosID:              instance.PlaceId,
				Wos:                instance.DestTitle,
				WosAddress:         warehouse[instance.PlaceId].Address,
				WarehouseAdminName: warehouse[instance.PlaceId].WarehouseAdminName,
				ProductID:          instance.ProductID,
				ProductName:        instance.Product,
				Units:              instance.Units,
				Num:                productCount[instance.ProductID].Wos[instance.PlaceId].Num + instance.Amount,   // 总库存
				InWos:              productCount[instance.ProductID].Wos[instance.PlaceId].InWos + instance.Amount, // 已在库
				NotWosed:           productCount[instance.ProductID].Wos[instance.PlaceId].NotWosed,                // 已入库
				Shipped:            productCount[instance.ProductID].Wos[instance.PlaceId].Shipped,                 // 已发货
				NotShipped:         productCount[instance.ProductID].Wos[instance.PlaceId].NotShipped,              // 未发货
			}
			break

		}
	}

	return productCount, nil
}

// 查询某商品在某仓库中的数量
func FindOneProductWos(product, warehouse_id, com_id int64) (int64, error) {
	var products []int64
	var arr []int64
	products = append(products, product)
	wosProduct, err := GetProductWos(products, com_id, arr)
	if err != nil {
		return 0, err
	}
	if wosProduct[product].Num == 0 {
		return 0, nil
	}
	wos, ok := wosProduct[product].Wos[warehouse_id]
	if !ok {
		return 0, nil
	}
	return wos.Num, nil
}

// 创建库存实例
func CreateWosExamples(wosExamplesData WosExamplesData, user_id, com_id int64) (error) {
	// 验证仓库id
	_, err := FindOneWarehouse(wosExamplesData.WarehouseId, com_id)
	if err != nil {
		return nil
	}

	// 验证商品
	_, err = FindOneProduct(wosExamplesData.Product, com_id)
	if err != nil {
		return nil
	}
	var wos_instance models.WosInstance

	// 商品id、商品数量、仓库id，商品价格 必填
	wos_instance.ProductID = wosExamplesData.Product
	wos_instance.ProductNum = wosExamplesData.Num
	//wos_instance.ProductUnitPrice = util.Unwrap(wosExamplesData.ProductUnitPrice,0)
	wos_instance.WarehouseID = wosExamplesData.WarehouseId
	// 设置创建者和创建时间
	wos_instance.CreateAt = user_id
	wos_instance.CreateBy = time.Now().Unix()

	// 填充添加数据
	switch wosExamplesData.Type {
	case 0:
		// 无故添加 +
		wos_instance.Type = 0
		wos_instance.CheckAt = user_id            // 盘点者
		wos_instance.CreateBy = time.Now().Unix() //盘点时间
		break
	case 1:
		// 退货 +
		// 订单号（销售订单号）
		// 查询销售订单是否存在
		_, err := FindSalesOrder(wosExamplesData.OrderSn, com_id)
		if err != nil {
			return nil
		}
		wos_instance.ConfirmAt = user_id           // 仓库确认收货人
		wos_instance.ConfirmBy = time.Now().Unix() // 仓库确认收货时间
		wos_instance.SalesOrderSn = wosExamplesData.OrderSn
		break
	case 2:
		// 销售 -
		// 需要参数，订单号（销售订单号）
		// 查询销售订单是否存在
		// 销售数量不能大于当前仓库的库存
		_, err := FindSalesOrder(wosExamplesData.OrderSn, com_id)
		if err != nil {
			return err
		}
		num, err := FindOneProductWos(wosExamplesData.Product, wosExamplesData.WarehouseId, com_id)
		if err != nil {
			return err
		}
		if num < wosExamplesData.Num {
			return errors.New("销售数量不能大于库存！")
		}
		wos_instance.SalesOrderSn = wosExamplesData.OrderSn
		break
	case 3:
		// 损耗 -
		// 损耗数量不能大于当前仓库的库存
		num, err := FindOneProductWos(wosExamplesData.Product, wosExamplesData.WarehouseId, com_id)
		if err != nil {
			return nil
		}
		if num < wosExamplesData.Num {
			return errors.New("损耗数量不能大于库存！")
		}

		break
	case 4:
		// 采购 +
		// 需要参数 采购订单号
		// 查询采购订单是否存在
		_, err = FindPurchaseOrder(wosExamplesData.OrderSn, com_id)
		if err != nil {
			return err
		}
		wos_instance.PurchaseOrderSn = wosExamplesData.OrderSn
		break
	default:
		// 未定义的类型
		return errors.New("未定义的类型")
	}
	collection := models.Client.Collection("wos_examples")
	_, err = collection.InsertOne(context.TODO(), wos_instance)
	if err != nil {
		return err
	}
	return nil
}

// 添加库存实例
func AddWosInstance(wosInstance []interface{}) (error) {
	collection := models.Client.Collection("instance")
	_, err := collection.InsertMany(context.TODO(), wosInstance)
	if err != nil {
		return err
	}
	return nil
}
