package models

type CustomerProductPrice struct {
	ComID        int64   `json:"com_id" bson:"com_id"`
	CustomerID   int64   `json:"customer_id" bson:"customer_id"`
	CustomerName string  `json:"customer_name" bson:"customer_name"`
	ProductID    int64   `json:"product_id" bson:"product_id"`
	Product      string  `json:"product" bson:"product_name"`
	Price        float64 `json:"price" bson:"price"`
	//DefaultPrice float64 `json:"default_price" bson:"default_price"`
	CreateAt int64 `json:"create_at" bson:"create_at"`
	IsValid  bool  `json:"is_valid" bson:"is_valid"`

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
	//DefaultPrice interface{} `json:"default_price"`
	PriceTable interface{} `json:"price_table"`
	//Products       []Product       `json:"product"`
	//Customers   []Customer `json:"customer"`
	Total       int `json:"total"`
	Pages       int `json:"pages"`
	Size        int `json:"size"`
	CurrentPage int `json:"current_page"`
}
