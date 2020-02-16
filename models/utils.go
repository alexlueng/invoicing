package models

import (
	"context"
	bson "go.mongodb.org/mongo-driver/bson")

const THIS_MODULE  = 1

func GetComIDAndModuleByDomain(domain string) (DomainData, error) {

	var com DomainData
	collection := Client.Collection("domain")
	filter := bson.D{{"domain", domain}}
	err := collection.FindOne(context.TODO(), filter).Decode(&com)
	if err != nil {
		return DomainData{}, err
	}
	return com, nil
}

