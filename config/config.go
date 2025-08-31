package config

import (
	"database/sql"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	GEMINI_MODEL   string `env:"GEMINI_MODEL"`
	GEMINI_API_KEY string `env:"GEMINI_API_KEY"`
	ENVIRONMENT    string `env:"ENVIRONMENT"`
	DB_PATH        string
	DB_NAME        string
	DB             *sql.DB
}

var (
	EnvProduction  = "prod"
	EnvDevelopment = "dev"
	Cfg            Config
	StartTime      time.Time
	Ping           = make(chan string, 250)
	HomeDIR        string
	AppName        = "logs"
	LogDIR         string
	TodoFilePath   string
)

func loadEnv(paths []string) error {
	for _, path := range paths {
		if err := godotenv.Load(path); err == nil {
			return nil
		}
	}
	return fmt.Errorf("No env file found, Checked : %s", strings.Join(slices.Compact(paths), ", "))
}

func MustLoad() error {
	var err error
	HomeDIR, err = os.UserHomeDir()

	if err != nil {
		return fmt.Errorf("Failed to get user home directory for writing logs and todos: %s", err.Error())
	}

	if err := loadEnv([]string{HomeDIR + "/.local/share/godo/.env", "./.env"}); err != nil {
		return err
	}

	if err := cleanenv.ReadEnv(&Cfg); err != nil {
		return err
	}
	StartTime = time.Now()
	if Cfg.GEMINI_MODEL == "" {
		Cfg.GEMINI_MODEL = "gemini-2.5-flash"
		fmt.Println("GEMINI_MODEL is not set, using default value:", Cfg.GEMINI_MODEL)
	}
	if Cfg.GEMINI_API_KEY == "" {
		Cfg.GEMINI_API_KEY = os.Getenv("GEMINI_API_KEY")
	}
	if Cfg.GEMINI_API_KEY == "" {
		return fmt.Errorf("GEMINI_API_KEY is not set")
	}
	if Cfg.ENVIRONMENT == "" {
		Cfg.ENVIRONMENT = EnvProduction
		fmt.Println("Env is not set, using default value:", Cfg.ENVIRONMENT)
	}
	if Cfg.ENVIRONMENT == EnvDevelopment {
		LogDIR = "."
		TodoFilePath = "./agentTodos.json"
	} else {
		LogDIR = HomeDIR + "/local/share/godo"
		TodoFilePath = HomeDIR + "/local/share/godo/agentTodos.json"
	}
	if Cfg.DB_PATH == "" {
		if Cfg.DB_NAME == "" {
			Cfg.DB_NAME = "todo.db"
		}
		Cfg.DB_PATH = HomeDIR + "/local/share/godo/" + Cfg.DB_NAME
	}
	if err = initDb(); err != nil {
		return err
	}
	return nil
}

func initDb() error {
	db, err := sql.Open("sqlite3", Cfg.DB_PATH)
	if err != nil {
		return err
	}
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS todos (
		Id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		Title TEXT NOT NULL,
		Description TEXT NOT NULL,
		Done BOOLEAN NOT NULL DEFAULT FALSE
	)`
	if _, err = db.Exec(sqlStmt); err != nil {
		return err
	}
	Cfg.DB = db
	return nil
}
