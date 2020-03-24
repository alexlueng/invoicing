package models

//

// 公司管理表数据结构
type CompanyData struct {
	ComId          int64  `json:"com_id" bson:"com_id"`
	ComName        string `json:"com_name" bson:"com_name"`
	ExpirationDate string `json:"expiration_date" bson:"expiration_date"`
	//Delivery       []DeliveryData `json:"delivery" bson:"delivery"`
	Units     string   `json:"units" bson:"units"`
	Payment   []string `json:"payment" bson:"payment"`     //结算方式
	Module    string   `json:"module" bson:"module"`       //平台名称
	Developer string   `json:"developer" bson:"developer"` //开发名称
	Admin     string   `json:"admin" bson:"admin"`
	Telephone string   `json:"telephone" bson:"telephone"`
}
