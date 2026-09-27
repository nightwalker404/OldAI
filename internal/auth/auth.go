package auth

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserExists = errors.New("user Username already exists")
var ErrUserNotFound = errors.New("invalid Username or Password")

func (s *db) Register(username, password string) error {
	var existingUser User
	err := s.collection.FindOne(context.Background(), bson.M{"username": username}).Decode(&existingUser)
	if err == nil {
		return ErrUserExists
	} else if err != mongo.ErrNoDocuments {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	newUser := User{
		Username: username,
		Password: string(hashedPassword),
	}

	_, err = s.collection.InsertOne(context.Background(), newUser)
	return err
}

func (s *db) Login(username, password string) error {
	var user User
	err := s.collection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return ErrUserNotFound
	} else if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return ErrUserNotFound
	}
	return nil
}
