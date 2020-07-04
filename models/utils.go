package models

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
)

const THIS_MODULE  = 1

type DomainError struct {
	s string
}

func (de *DomainError) Error() string {
	return de.s
}

func GetComIDAndModuleByDomain(domain string) (*DomainData, error) {

	fmt.Println("domain string: ", domain)

	var com DomainData
	collection := Client.Collection("domain")
	filter := bson.D{{"domain", domain}}
	err := collection.FindOne(context.TODO(), filter).Decode(&com)
	if err != nil {
		return nil, &DomainError{"域名未注册"}
	}
	if com.Status == false {
		return nil, &DomainError{"域名已停用"}
	}
	return &com, nil
}

