package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	GROQ_MODEL   string `env:"GROQ_MODEL"`
	GROQ_API_KEY string `env:"GROQ_API_KEY" env-required:"true"`
	Event        chan string
}

var Cfg Config

func MustLoad() error {
	if err := godotenv.Load("./.env"); err != nil {
		return err
	}
	if err := cleanenv.ReadEnv(&Cfg); err != nil {
		return err
	}
	Cfg.Event = make(chan string, 100)
	Cfg.Event <- ":)"
	if Cfg.GROQ_MODEL == "" {
		Cfg.GROQ_MODEL = "llama-3.1-8b-instant"
	}
	return nil
}
