package models

// mongo中数据表字段维护在代码中
// 售价管理表
type CustomerProductPrice struct {
	ComID        int64   `json:"com_id" bson:"com_id"`
	CustomerID   int64   `json:"customer_id" bson:"customer_id"`
	CustomerName string  `json:"customer_name" bson:"customer_name"`
	ProductID    int64   `json:"product_id" bson:"product_id"`
	Product      string  `json:"product" bson:"product_name"`
	Price        float64 `json:"price" bson:"price"`
	CreateAt     int64   `json:"create_at" bson:"create_at"`
	IsValid      bool    `json:"is_valid" bson:"is_valid"`
	DefaultPrice float64 `json:"default_price" bson:"omitempty"`
}

type CustomerProductPriceReq struct {
	BaseReq
	Product      string `json:"product"`
	ProductID    int64  `json:"product_id"`
	CustomerID   int64  `json:"customer_id"`
	CustomerName string `json:"customer_name"`
}

type ResponseCustomerProductPriceData struct {
	PriceTable  interface{} `json:"price_table"`
	Total       int         `json:"total"`
	Pages       int         `json:"pages"`
	Size        int         `json:"size"`
	CurrentPage int         `json:"current_page"`
}
