package models

// import "go.mongodb.org/mongo-driver/bson"



func (u User) CheckPassword(password string) bool {
	if u.Password != password {
		return false
	}
	return true
}

// 登录日志数据结构
type LoginLogData struct {
	LogId   string `json:"log_id"`
	Ip      string `json:"ip"`
	UserId  string `json:"user_id"`
	Message string `json:"message"`
}

