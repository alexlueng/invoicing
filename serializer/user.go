package serializer

import "jxc/models"

// User 用户序列化器
type User struct {
	ID        uint   `json:"id"`
	UserName  string `json:"user_name"`
	Nickname  string `json:"nickname"`
	Status    string `json:"status"`
	Avatar    string `json:"avatar"`
	CreatedAt int64  `json:"created_at"`
}

// BuildUser 序列化用户
func BuildUser(user models.User) User {
	return User{
		UserName: user.Username,
	}
}

// BuildUserResponse 序列化用户响应
func BuildUserResponse(user models.User) Response {
	return Response{
		Data: BuildUser(user),
		Code: 200,
		Msg:  "Login success",
	}
}
