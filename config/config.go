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

type StreamMsg struct {
	Text   string
	IsUser bool
}

var (
	EnvProduction  = "prod"
	EnvDevelopment = "dev"
	Cfg            Config
	StartTime      time.Time
	Ping           = make(chan string, 250)
	StreamResponse = make(chan StreamMsg, 1)
	HomeDIR        string
	AppDIR         = "/.local/share/godo/"
	AppName        = "logs"
	LogDIR         string
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

	if err := loadEnv([]string{HomeDIR + AppDIR + ".env", "./.env"}); err != nil {

		return err
	}

	if err := cleanenv.ReadEnv(&Cfg); err != nil {
		return err
	}
	StartTime = time.Now()
	if err = getApiKey(); err != nil {
		return err
	}
	if Cfg.GEMINI_MODEL == "" {
		Cfg.GEMINI_MODEL = "gemini-2.5-flash"
		fmt.Println("GEMINI_MODEL is not set, using default value:", Cfg.GEMINI_MODEL)
	}
	if Cfg.ENVIRONMENT == "" {
		Cfg.ENVIRONMENT = EnvProduction
	}
	if Cfg.ENVIRONMENT == EnvDevelopment {
		LogDIR = "."
	} else {
		LogDIR = HomeDIR + AppDIR
	}
	if Cfg.DB_PATH == "" {
		if Cfg.DB_NAME == "" {
			Cfg.DB_NAME = "todo.db"
		}
		Cfg.DB_PATH = HomeDIR + AppDIR + Cfg.DB_NAME
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
	BEGIN;
	CREATE TABLE IF NOT EXISTS todos (
		Id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		Title TEXT NOT NULL,
		Description TEXT NOT NULL,
		Done BOOLEAN NOT NULL DEFAULT FALSE
	);
	CREATE TABLE IF NOT EXISTS chats(
		Id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		chat TEXT
	);
	COMMIT;
	`
	if _, err = db.Exec(sqlStmt); err != nil {
		return err
	}
	Cfg.DB = db
	return nil
}

func getApiKey() error {
	Cfg.GEMINI_API_KEY = os.Getenv("GEMINI_API_KEY")
	if Cfg.GEMINI_API_KEY == "" {
		fmt.Println("Enter your gemini api key")
		input := ""
		fmt.Scanln(&input)
		Cfg.GEMINI_API_KEY = input
		if err := saveApiKey(Cfg.GEMINI_API_KEY); err != nil {
			return err
		}
	}
	if Cfg.GEMINI_API_KEY == "" {
		return fmt.Errorf("GEMINI_API_KEY is not set")
	}
	return nil
}

func saveApiKey(key string) error {
	file := HomeDIR + AppDIR + ".env"
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("GEMINI_API_KEY=" + key + "\n")
	return err

}
