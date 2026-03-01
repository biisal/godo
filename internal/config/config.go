package config

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

type Config struct {
	OPENAI_API_KEY  string `env:"OPENAI_API_KEY"`
	OPENAI_MODEL    string `env:"OPENAI_MODEL"`
	OPENAI_BASE_URL string `env:"OPENAI_BASE_URL"`
	ENVIRONMENT     string `env:"ENVIRONMENT"`
	MODE            string `env:"MODE"`
	DB_PATH         string
	DB_NAME         string
	DB              *sql.DB
}

var (
	EnvProduction  = "prod"
	EnvDevelopment = "dev"
	ModeTodo       = "todo"
	ModeAgent      = "agent"
	Cfg            Config
	StartTime      time.Time
	Ping           = make(chan string, 250)
	HomeDIR        string
	AppDIR         = "/.godo/"
	AppName        = "logs"
	LogDIR         string
	EroorType      = "error"
	MessageType    = "message"
)

func loadEnv(paths []string) error {
	for _, path := range paths {
		if err := godotenv.Load(path); err == nil {
			return nil
		}
	}
	return fmt.Errorf("no env file found, checked: %s", strings.Join(slices.Compact(paths), ", "))
}

func MustLoad() error {
	var err error
	HomeDIR, err = os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory for writing logs and todos: %s", err.Error())
	}
	if err = os.MkdirAll(HomeDIR+AppDIR, os.ModePerm); err != nil {
		return err
	}

	_ = loadEnv([]string{"./.env", filepath.Join(HomeDIR, AppDIR, ".env")})

	if err = cleanenv.ReadEnv(&Cfg); err != nil {
		return err
	}
	StartTime = time.Now()
	if err = getApiKey(); err != nil {
		return err
	}
	if Cfg.OPENAI_MODEL == "" {
		Cfg.OPENAI_MODEL = "gpt-4o-mini"
		fmt.Println("OPENAI_MODEL is not set, using default value:", Cfg.OPENAI_MODEL)
	}
	if Cfg.OPENAI_BASE_URL == "" {
		Cfg.OPENAI_BASE_URL = "https://api.openai.com/v1"
	}
	if Cfg.ENVIRONMENT == "" {
		Cfg.ENVIRONMENT = EnvProduction
	}
	if Cfg.ENVIRONMENT == EnvDevelopment {
		LogDIR = "."
	} else {
		LogDIR = HomeDIR + AppDIR
	}
	if Cfg.MODE == "" {
		Cfg.MODE = "agent"
	}
	if Cfg.DB_PATH == "" {
		if Cfg.DB_NAME == "" {
			Cfg.DB_NAME = "todo.db"
		}
		Cfg.DB_PATH = HomeDIR + AppDIR + Cfg.DB_NAME
	}
	if err = SaveCfg(); err != nil {
		return err
	}
	if err = initDb(); err != nil {
		return err
	}
	return nil
}

func SaveCfg() error {
	filepath := HomeDIR + AppDIR + ".env"
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	content := ""
	content += "OPENAI_API_KEY=" + Cfg.OPENAI_API_KEY + "\n"
	content += "OPENAI_MODEL=" + Cfg.OPENAI_MODEL + "\n"
	content += "OPENAI_BASE_URL=" + Cfg.OPENAI_BASE_URL + "\n"
	content += "MODE=" + Cfg.MODE + "\n"

	_, err = f.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func initDb() error {
	db, err := sql.Open("sqlite", Cfg.DB_PATH)
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
	CREATE TABLE IF NOT EXISTS memories (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL UNIQUE,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
	Cfg.OPENAI_API_KEY = os.Getenv("OPENAI_API_KEY")
	if Cfg.OPENAI_API_KEY == "" {
		fmt.Println("Enter your OpenAI API key (or compatible API key):")
		input := ""
		_, _ = fmt.Scanln(&input)
		Cfg.OPENAI_API_KEY = input
		if err := SaveCfg(); err != nil {
			return err
		}
	}
	if Cfg.OPENAI_API_KEY == "" {
		return fmt.Errorf("OPENAI_API_KEY is not set")
	}
	return nil
}
