package auth

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	collection *mongo.Collection
}

func New(uri, dbName, collectionName string) (*Storage, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	db := client.Database(dbName)
	return &Storage{collection: db.Collection(collectionName)}, nil
}
