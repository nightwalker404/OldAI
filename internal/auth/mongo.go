package auth

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type db struct {
	collection *mongo.Collection
}

func NewDb(uri, dbName, collectionName string) (*db, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	database := client.Database(dbName)
	return &db{collection: database.Collection(collectionName)}, nil
}
