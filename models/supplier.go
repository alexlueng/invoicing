package models

type Supplier struct {
	ID int64 `json:"supplier_id" bson:"supplier_id"`
	ComID int64 `json:"com_id" bson:"com_id"`
	Phone string `json:"phone" bson:"phone"`
	SupplierName string `json:"supplier_name" bson:"supplier_name"`
	Address string `json:"address" bson:"address"`
	Contacts string `json:"contacts" bson:"contacts"`
	TransactionNum int64 `json:"transaction_num" bson:"transaction_num"`
	Payment string `json:"payment" bson:"payment"`
	// due string
	Level int64 `json:"level" bson:"level"`
}

type SupplierReq struct {
	BaseReq
	// 本页订制的搜索条件
	ID int64 `json:"supplier_id" form:"supplier_id"`
	Phone string `json:"phone" form:"phone"`
	SupplierName string `json:"supplier_name" form:"supplier_name"`
	Address string `json:"address" form:"address"`
	Level int64 `json:"level" form:"level"`
	Payment string `json:"payment" form:"payment"`
}

type ResponseSupplierData struct {
	Suppliers   []Supplier `json:"suppliers"`
	Total       int        `json:"total"`
	Pages       int        `json:"pages"`
	Size        int        `json:"size"`
	CurrentPage int        `json:"current_page"`
}
