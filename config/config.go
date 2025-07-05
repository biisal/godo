package config

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	GEMINI_MODEL   string `env:"GEMINI_MODEL"`
	GEMINI_API_KEY string `env:"GEMINI_API_KEY"`
	ENVIRONMENT    string `env:"ENVIRONMENT"`
	Event          chan string
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

func checkErr(err error) bool {
	if err != nil {
		fmt.Println("Log error:", err)
		return true // error occurred
	}
	return false // no error
}
func WriteLog(DEBUG bool, msg ...any) {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Print("Error detected logging:", r)
		}
	}()
	if DEBUG {
		fmt.Println(msg...)
	} else {
		logfile, err := os.OpenFile(LogDIR+"/"+AppName+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if !checkErr(err) {
			log.SetOutput(logfile)
			log.Println(msg...)
		}
		defer logfile.Close()
	}
}

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

	if err := loadEnv([]string{"./.env", HomeDIR + "/.local/share/godo/.env"}); err != nil {
		return err
	}

	if err := cleanenv.ReadEnv(&Cfg); err != nil {
		return err
	}
	StartTime = time.Now()
	if Cfg.GEMINI_MODEL == "" {
		Cfg.GEMINI_MODEL = "llama-3.1-8b-instant"
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
	Cfg.Event = make(chan string, 100)
	Cfg.Event <- ":)"

	return nil
}
