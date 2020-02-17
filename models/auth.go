package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
)

type Auth struct {
	ID int64 `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func CheckAuth(username, password string) bool {
	var auth Auth
	filter := bson.M{"username": username, "password": password}
	_ = Client.Collection("auth").FindOne(context.TODO(), filter).Decode(&auth)
	if auth.ID > 0 {
		return true
	}
	return false
}