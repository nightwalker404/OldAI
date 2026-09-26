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

func (s *Storage) Save(SessionID string, history []vllmclient.Message) error {
	data, err := json.Marshal(history)
	if err != nil {
		return err
	}

	return s.rdb.Set(ctx, "chat:"+SessionID, data, 0).Err()
}

func (s *Storage) Load(SessionID string) ([]vllmclient.Message, error) {
	data, err := s.rdb.Get(ctx, "chat:"+SessionID).Result()
	if err == redis.Nil {
		return []vllmclient.Message{}, redis.Nil
	} else if err != nil {
		return nil, err
	}

	var history []vllmclient.Message
	if err := json.Unmarshal([]byte(data), &history); err != nil {
		return nil, err
	}

	return history, nil
}
