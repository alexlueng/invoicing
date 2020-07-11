package models

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Auth struct {
	ID int64 `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func getAuthCollection() *mongo.Collection {
	return Client.Collection("auth")
}