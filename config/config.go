package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	GROQ_API_KEY string `env:"GROQ_API_KEY" env-required:"true"`
}

var Cfg Config

func MustLoad() error {
	if err := godotenv.Load("./.env"); err != nil {
		return err
	}
	if err := cleanenv.ReadEnv(&Cfg); err != nil {
		return err
	}
	return nil
}
