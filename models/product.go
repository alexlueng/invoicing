package models

type Product struct {
	ComID int64 `json:"com_id" bson:"com_id"`
	ProductID int64 `json:"product_id" bson:"product_id"`
	Product string `json:"product" bson:"product"`
	Units string `json:"units" bson:"units"`
	URL string `json:"url" bson:"url"`
	Num int64 `json:"num" bson:"num"`
	PriceOfSuppliers []PriceOfSupplier `json:"price_of_supplier"`
}

type PriceOfSupplier struct {
	Supplier string `json:"supplier"`
	Price float64 `json:"price"`
}

type ProductReq struct {
	BaseReq
	Product string `json:"product" form:"product"`
	Units string `json:"units" form:"units"`
	PriceOfSuppliers []PriceOfSupplier `json:"price_of_supplier"`
	InStock int64 `json:"in_stock"`
}

type ResponseProductData struct {
	Products   []Product `json:"products"`
	Total       int        `json:"total"`
	Pages       int        `json:"pages"`
	Size        int        `json:"size"`
	CurrentPage int        `json:"current_page"`
}