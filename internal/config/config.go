package config

import (
	"errors"
	"gophermart/utils"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

var ErrFileNotFound = errors.New("file not found")

type Flags struct {
	A string // RUN_ADDRESS
	D string // DATABASE_URI
	R string // ACCRUAL_SYSTEM_ADDRESS
}

type Config struct {
	RunAdress                string
	AccrualSysremAdress      string
	AccrualRequestInterval   int
	AccuralPuttingDBInterval int
	DatabaseURI              string
	LoggerLevel              string
	Key                      string
	TokenExp                 time.Duration
	MigrationsPath           string
	NumberOfWorkers          int
}

func NewConfig(flag Flags) (*Config, error) {
	log.Println("NewConfig=================")

	c := Config{}
	if buf, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		c.RunAdress = buf
	} else {
		var err error
		if c.RunAdress, err = utils.GetValidURL(flag.A); err != nil {
			return &Config{}, utils.ErrorWrongURLFlag
		}
	}

	if buf, ok := os.LookupEnv("DATABASE_URI"); ok {
		c.DatabaseURI = buf
	} else {
		c.DatabaseURI = flag.D
	}

	if buf, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		c.AccrualSysremAdress = buf
	} else {
		c.AccrualSysremAdress = flag.R
	}
	rootDir, err := utils.FindProjectRoot()
	if err != nil {
		log.Println("Не удалось определить корневой каталог проекта. Будут использованы значения по умолчанию.")
	}
	filepathStr := filepath.Join(rootDir, "configs", "config.toml")
	_, err = toml.DecodeFile(filepathStr, &c)
	if err != nil {
		c.MigrationsPath = "migrations"
		c.TokenExp = time.Hour * 999
		c.AccrualRequestInterval = 1
		c.AccuralPuttingDBInterval = 1
		c.NumberOfWorkers = 3
		c.LoggerLevel = "Info"
		return &c, ErrFileNotFound
	}

	c.TokenExp = c.TokenExp * time.Hour

	return &c, nil

}
