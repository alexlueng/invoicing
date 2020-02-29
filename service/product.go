package service

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"jxc/models"
)

// 商品相关方法

// 查找商品
func FindOneProduct(product_id, com_id int64) (*models.Product, error) {
	collection := models.Client.Collection("product")
	var product models.Product
	filter := bson.M{}
	filter["product_id"] = product_id
	filter["com_id"] = com_id

	err := collection.FindOne(context.TODO(), filter).Decode(&product)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// 商品采购价格
type ProductPurchasePrice struct {
	ProductId      int64                    `json:"product_id"`             // 商品id
	ProductName    string                   `json:"product_name"`           // 商品名
	SupplierPrices map[int64]SupplierPrices `json:"supplier_product_price"` // 各供应商处采购价
}
type SupplierPrices struct {
	SupplierId    int64   `json:"supplier_id"`    // 供应商id
	SupplierName  string  `json:"supplier_name"`  //供应商名
	SupplierPrice float64 `json:"supplier_price"` // 供应商价格
}

// 查找某商品的采购价格
func FindOneProductPurchasePrice(product_id, com_id int64) (*ProductPurchasePrice, error) {
	/*collection := models.Client.Collection("supplier_product_price")
	filter := bson.M{}
	var productPurchasePrice ProductPurchasePrice
	var supplierProductPriceTMP models.SupplierProductPrice
	supplierProductPrice := make(map[int64]models.SupplierProductPrice)

	filter["comd_id"] = com_id
	filter["product_id"] = product_id
	filter["is_valid"] = true
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	// supplier_id 为0，是这个商品的默认进货价
	for cur.Next(context.TODO()) {
		err = cur.Decode(&supplierProductPriceTMP)
		if err != nil {
			return nil, err
		}
		if supplierProductPriceTMP.SupplierID == 0 {
			productPurchasePrice.DefaultPrice = supplierProductPriceTMP.Price
		} else {
			supplierProductPrice[supplierProductPriceTMP.ProductID] = supplierProductPriceTMP
		}
	}
	productPurchasePrice.ProductId = product_id
	productPurchasePrice.Prices = supplierProductPrice
	return &productPurchasePrice, nil
	*/
	return nil, nil
}

// 查找一组商品的采购价
func FindProductPurchasePrice(product_id []int64, com_id int64) (map[int64]ProductPurchasePrice, error) {
	collection := models.Client.Collection("supplier_product_price")
	filter := bson.M{}
	var productPriceTMP models.SupplierProductPrice // 用来接从数据库中取出的数据
	ProductPrice := make(map[int64]ProductPurchasePrice)
	filter["com_id"] = com_id
	filter["is_valid"] = true
	filter["product_id"] = bson.M{"$in": product_id}
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {
		err = cur.Decode(&productPriceTMP)
		if err != nil {
			return nil, err
		}
		// 第一次出现的产品为他分配内存
		_, ok := ProductPrice[productPriceTMP.ProductID]
		if !ok {
			ProductPrice[productPriceTMP.ProductID] = ProductPurchasePrice{
				ProductId:   productPriceTMP.ProductID,
				ProductName: productPriceTMP.Product,
				SupplierPrices: map[int64]SupplierPrices{
					productPriceTMP.ProductID: {
						SupplierId:    productPriceTMP.SupplierID,
						SupplierName:  productPriceTMP.Supplier,
						SupplierPrice: productPriceTMP.Price,
					},
				},
			}
		}
		ProductPrice[productPriceTMP.ProductID].SupplierPrices[productPriceTMP.SupplierID] = SupplierPrices{
			SupplierId:    productPriceTMP.SupplierID,
			SupplierName:  productPriceTMP.Supplier,
			SupplierPrice: productPriceTMP.Price,
		}

	}
	return ProductPrice, nil

}