package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	RedisPort                int     `env:"RedisPort"`
	RedisHost                string  `env:"RedisHost"`
	VLLMPort                 int     `env:"VLLM_PORT"`
	VLLMBaseURL              string  `env:"VLLM_BASE_URL"`
	VLLMTimeout              string  `env:"VLLM_TIMEOUT"`
	VLLMModes                string  `env:"VLLM_MODELS"`
	VLLMGPUMemoryUtilization float64 `env:"VLLM_GPU_MEMORY_UTILIZATION"`
	VLLMMaxModelLen          int     `env:"VLLM_MAX_MODEL_LEN"`
}

func LoadConfig() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
