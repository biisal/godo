package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	GROQ_MODEL   string `env:"GROQ_MODEL"`
	GROQ_API_KEY string `env:"GROQ_API_KEY"`
	Event        chan string
}

var (
	Cfg          Config
	StartTime    time.Time
	Ping         = make(chan string, 250)
	HomeDIR      string
	AppName      = "logs"
	LogDIR       string
	TodoFilePath string
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

func MustLoad() error {
	var err error
	HomeDIR, err = os.UserHomeDir()

	if err != nil {
		return fmt.Errorf("Failed to get user home directory for writing logs and todos: %s", err.Error())
	}

	if err := godotenv.Load("./.env"); err != nil {
		fmt.Println("Warning : .env file not found")
	}
	if err := cleanenv.ReadEnv(&Cfg); err != nil {
		return err
	}
	StartTime = time.Now()
	Cfg.Event = make(chan string, 100)
	Cfg.Event <- ":)"
	if Cfg.GROQ_MODEL == "" {
		Cfg.GROQ_MODEL = "llama-3.1-8b-instant"
		fmt.Println("GROQ_MODEL is not set, using default value:", Cfg.GROQ_MODEL)
	}
	if Cfg.GROQ_API_KEY == "" {
		Cfg.GROQ_API_KEY = os.Getenv("GROQ_API_KEY")
	}
	if Cfg.GROQ_API_KEY == "" {
		return fmt.Errorf("GROQ_API_KEY is not set")
	}

	LogDIR = HomeDIR + "/.local/share/godo"
	LogDIR = HomeDIR + "/local/share/godo"
	TodoFilePath = HomeDIR + "/local/share/godo/agentTodos.json"
	return nil
}
