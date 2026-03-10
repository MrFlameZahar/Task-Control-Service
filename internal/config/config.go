package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string `yaml:"app_port" env:"APP_PORT" env-default:"8080"`

	MySQL MySQL `yaml:"mysql"`
	Redis Redis `yaml:"redis"`

	JWTSecret string `yaml:"jwt_secret" env:"JWT_SECRET"`
}

type MySQL struct {
	Host     string `yaml:"host" env:"MYSQL_HOST"`
	Port     string `yaml:"port" env:"MYSQL_PORT"`
	User     string `yaml:"user" env:"MYSQL_USER"`
	Password string `yaml:"password" env:"MYSQL_PASSWORD"`
	DB       string `yaml:"db" env:"MYSQL_DB"`
}

type Redis struct {
	Host string `yaml:"host" env:"REDIS_HOST"`
	Port string `yaml:"port" env:"REDIS_PORT"`
}

func MustLoad() (*Config, error) {
	var cfg Config

	_ = godotenv.Load()

	if _, err := os.Stat("config.yaml"); err == nil {
		if err := cleanenv.ReadConfig("config.yaml", &cfg); err != nil {
			return nil, err
		}
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}