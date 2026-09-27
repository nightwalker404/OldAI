package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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
	session := d.collection.Database().Collection("session")
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

func sessionFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".oldai", "session.json"), nil
}

type localSession struct {
	Token string `json:"token"`
}

func SaveLocalToken(token string) error {
	session := localSession{Token: token}
	path, err := sessionFilePath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func LoadLocalToken() (string, error) {
	path, err := sessionFilePath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("session.json still missing after first run")
		}
		return "", err
	}

	var session localSession
	if err := json.Unmarshal(data, &session); err != nil {
		return "", err
	}

	return session.Token, nil
}

func (d *DataBase) IsFirstRun() (bool, error) {
	path, err := sessionFilePath()
	if err != nil {
		return false, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return false, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return true, nil
	}
	return false, nil
}

func (d *DataBase) CreateFirstRunSession() error {
	token, err := generateToken()
	if err != nil {
		return err
	}

	if err := SaveLocalToken(token); err != nil {
		return err
	}

	return nil
}
