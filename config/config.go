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
	GROQ_API_KEY string `env:"GROQ_API_KEY" env-required:"true"`
	Event        chan string
}
type AgentResModel struct {
	Text string
	Done bool
}

var (
	Cfg       Config
	StartTime time.Time
	AgentRes  = make(chan AgentResModel, 2)
	LOGDIR    = "./"
	APPNAME   = "logs"
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
		logfile, err := os.OpenFile(LOGDIR+"/"+APPNAME+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if !checkErr(err) {
			log.SetOutput(logfile)
			log.Println(msg...)
		}
		defer logfile.Close()
	}
}

func MustLoad() error {
	if err := godotenv.Load("./.env"); err != nil {
		return err
	}
	if err := cleanenv.ReadEnv(&Cfg); err != nil {
		return err
	}
	StartTime = time.Now()
	Cfg.Event = make(chan string, 100)
	Cfg.Event <- ":)"
	if Cfg.GROQ_MODEL == "" {
		Cfg.GROQ_MODEL = "llama-3.1-8b-instant"
	}
	return nil
}
