package models

// 域名表
type DomainData struct {
	ComId    int64  `bson:"comid" json:"comid"`       // 公司id
	Domain   string `bson:"domain" json:"domain"`     //
	ModuleId int64  `bson:"moduleid" json:"moduleid"` //
}
