package models

import (
	"context"
	"jxc/util"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Collection *mongo.Collection
	Client *mongo.Database
	size int64
)

func Database(connString, dbname string) {
	ctx := context.Background()
	clientOpts := options.Client().ApplyURI(connString)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		util.Log().Panic("连接mongodb不成功", err)
	}
	db := client.Database(dbname)
	Client = db
	util.Log().Info(db.Name())
}
