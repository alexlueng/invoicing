package models

type Product struct {
	ComID            int64             `json:"com_id" bson:"com_id"`
	ProductID        int64             `json:"product_id" bson:"product_id"`
	Product          string            `json:"product" bson:"product" form:"product"`
	Units            string            `json:"units" bson:"units"`
	URLS             []string          `json:"urls" bson:"urls"`
	Num              int64             `json:"num" bson:"num"`
	//PriceOfSuppliers []PriceOfSupplier `json:"price_of_supplier" bson:"price_of_supplier"`
	Stock            int64             `json:"stock" bson:"stock"` //库存
	DefaultPrice     float64           `json:"default_price" bson:"default_price" form:"default_price"`
	CusPrice         []int64           `json:"cus_price" bson:"cus_price"`
	SupPrice         []int64           `json:"sup_price" bson:"sup_price"`

	CatID            int64             `json:"cat_id" bson:"cat_id"` // 商品分类

	MinAlert         int64             `json:"min_alert" bson:"min_alert"` // 最低库存预警
}

type PriceOfSupplier struct {
	SupplierID int64   `json:"supplier_id"`
	Supplier   string  `json:"supplier"`
	Price      float64 `json:"price"`
}

type ProductSupplier struct {
	ID   int    `json:"supplier_id" bson:"supplier_id"`
	Name string `json:"supplier_name" bson:"supplier_name"`
}

type ProductReq struct {
	BaseReq
	Product          string            `json:"product" form:"product"`
	Units            string            `json:"units" form:"units"`
	PriceOfSuppliers []PriceOfSupplier `json:"price_of_supplier"`
	InStock          int64             `json:"in_stock"`
}

type ResponseProductData struct {
	Products    []Product         `json:"products"`
	//Suppliers   []ProductSupplier `json:"product_suppliers"`
	Total       int               `json:"total"`
	Pages       int               `json:"pages"`
	Size        int               `json:"size"`
	CurrentPage int               `json:"current_page"`
}
