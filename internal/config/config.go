package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env                      string        `yaml:"env" env-default:"local"`
	ConnectionStringPostgres string        `yaml:"connection-string-postgres-sso"`
	AccessTokenTTL           time.Duration `yaml:"access_token_ttl" env-required:"true"`
	RefreshTokenTTL          time.Duration `yaml:"refresh_token_ttl" env-required:"true"`
	Http                     HTTPConfig    `yaml:"http"`
}

type HTTPConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(fmt.Sprintf("failed to parse config: %s", err.Error()))
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
