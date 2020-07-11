package models

import "go.mongodb.org/mongo-driver/mongo"

// 供应商价格表
type SupplierProductPrice struct {
	ComID      int64   `json:"com_id" bson:"com_id"`
	ProductID  int64   `json:"product_id" bson:"product_id"`
	Product    string  `json:"product" bson:"product"`
	SupplierID int64   `json:"supplier_id" bson:"supplier_id"`
	Supplier   string  `json:"supplier_name" bson:"supplier"`
	Price      float64 `json:"price" bson:"price"`
	CreateAt   int64   `json:"create_at" bson:"create_at"`
	IsValid    bool    `json:"is_valid" bson:"is_valid"`
}

func getSupplierProductPriceCollection() *mongo.Collection {
	return Client.Collection("supplier_product_price")
}

type SupplierProductPriceReq struct {
	BaseReq
	Product      string `json:"product"`
	SupplierName string `json:"supplier_name"`
	ProductID    int64  `json:"product_id"`
	SupplierID   int64  `json:"supplier_id"`
}

type ResponseSupplierProductPriceData struct {
	PriceTable  interface{} `json:"price_table"`
	Total       int         `json:"total"`
	Pages       int         `json:"pages"`
	Size        int         `json:"size"`
	CurrentPage int         `json:"current_page"`
}
