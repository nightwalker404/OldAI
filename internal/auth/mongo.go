package auth

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	client *mongo.Client
}

func New(uri string) (*Storage, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return &Storage{client: client}, nil
}

func (s *Storage) Close() error {
	return s.client.Disconnect(context.Background())
}
