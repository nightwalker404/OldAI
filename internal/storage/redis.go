package storage

import (
	"context"
	"encoding/json"

	"github.com/nightwalker404/OldAI/internal/vllmclient"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type Storage struct {
	rdb *redis.Client
}

func New(addr string) *Storage {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return &Storage{rdb: rdb}
}

func (s *Storage) AppendMessage(SessionID string, message vllmclient.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return s.rdb.RPush(ctx, "chat:"+SessionID, data).Err()
}

func (s *Storage) Load(SessionID string) ([]vllmclient.Message, error) {
	data, err := s.rdb.LRange(ctx, "chat:"+SessionID, 0, -1).Result()
	if err == redis.Nil {
		return []vllmclient.Message{}, redis.Nil
	} else if err != nil {
		return nil, err
	}

	history := make([]vllmclient.Message, 0, len(data))
	for _, d := range data {
		var msg vllmclient.Message
		if err := json.Unmarshal([]byte(d), &msg); err != nil {
			return nil, err
		}
		history = append(history, msg)
	}

	return history, nil
}
