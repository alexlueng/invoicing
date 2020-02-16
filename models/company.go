package models

//

// 公司管理表数据结构
type CompanyData struct {
	ComId          string `json:"com_id" bson:"com_id"`
	ComName        string `json:"com_name" bson:"com_name"`
	ExpirationDate string `json:"expiration_date" bson:"expiration_date"`
	DeliveryData
	Units     []string `json:"units" bson:"units"`
	Payment   []string `json:"payment" bson:"payment"`     //结算方式
	Module    string   `json:"module" bson:"module"`       //平台名称
	Developer string   `json:"developer" bson:"developer"` //开发名称
}

// 配送方式
type DeliveryData struct {
	Idx     string
	EnTitle string
	Title   string
}


