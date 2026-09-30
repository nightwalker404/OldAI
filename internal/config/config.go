package config

import (
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	RedisPort                int     `env:"RedisPort"`
	RedisHost                string  `env:"RedisHost"`
	RedisSession             string  `env:"RedisSession"`
	RedisPassword            string  `env:"RedisPassword"`
	VLLMPort                 int     `env:"VLLM_PORT"`
	VLLMBaseURL              string  `env:"VLLM_BASE_URL"`
	VLLMTimeout              string  `env:"VLLM_TIMEOUT"`
	VLLMModel                string  `env:"VLLM_MODEL"`
	VLLMGPUMemoryUtilization float64 `env:"VLLM_GPU_MEMORY_UTILIZATION"`
	VLLMMaxModelLen          int     `env:"VLLM_MAX_MODEL_LEN"`
	MongoURI                 string  `env:"MONGO_URI"`
	MongoDBName              string  `env:"MONGO_DB_NAME"`
	ServerPort               int     `env:"ServerPort"`
	ServerHost               string  `env:"ServerHost"`
}

var (
	cfg  Config
	once sync.Once
	err  error
)

func LoadConfig() (Config, error) {
	once.Do(func() {
		_ = godotenv.Load()
		err = env.Parse(&cfg)
	})
	return cfg, err
}

func Get() Config {
	return cfg
}
