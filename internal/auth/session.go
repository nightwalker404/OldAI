package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

var ErrSessionNotFound = errors.New("session invalid or expired")

type Session struct {
	Token     string    `bson:"token"`
	Username  string    `bson:"username"`
	ExpiresAt time.Time `bson:"expires_at"`
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (d *DataBase) CreateSession(username string, duration time.Duration) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}
	session := d.collection.Database().Collection("sessions")
	_, err = session.InsertOne(context.Background(), Session{
		Token:     token,
		Username:  username,
		ExpiresAt: time.Now().Add(duration),
	})
	if err != nil {
		return "", err
	}

	return token, nil
}

func (d *DataBase) ValidateSession(token string) (string, error) {
	session := d.collection.Database().Collection("sessions")
	var s Session
	err := session.FindOne(context.Background(), map[string]interface{}{"token": token}).Decode(&s)
	if err == mongo.ErrNoDocuments {
		return "", ErrSessionNotFound
	} else if err != nil {
		return "", ErrSessionNotFound
	}

	if time.Now().After(s.ExpiresAt) {
		return "", ErrSessionNotFound
	}

	return s.Username, nil
}
