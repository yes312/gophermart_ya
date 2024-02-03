package main

import (
	"context"
	"flag"

	"gophermart/internal/app"
	"gophermart/internal/config"
	"gophermart/pkg/logger"

	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var f config.Flags

func init() {

	// #ВопросМентору: стоит ли строчки ниже спрятать в функцию в пакет config  или в отдельный пакет?

	flag.StringVar(&f.A, "a", "localhost:8081", "IP adress")
	// flag.StringVar(&f.D, "d", "postgresql://postgres:12345@localhost/gmtest?sslmode=disable", "database uri")
	flag.StringVar(&f.R, "r", "http://127.0.0.1:8080", "ACCRUAL_SYSTEM_ADDRESS")
	flag.StringVar(&f.D, "d", "postgresql://postgres:12345@localhost/gmtest?sslmode=disable", "database uri")
}

func main() {

	flag.Parse()

	newConfig, err := config.NewConfig(f)

	if err == config.ErrFileNotFound {
		log.Println("Файл конфигурации не найден.Будут использованы значения по умолчанию.")
	} else {
		if err != nil {
			log.Fatal(err)
		}
	}

	logger, err := logger.NewLogger(newConfig.LoggerLevel)

	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {

		<-c
		cancel()
		logger.Info("Завершение по сигналу с клавиатуры. ")
		os.Exit(0)

	}()

	s := app.New(ctx, newConfig)
	wg := &sync.WaitGroup{}
	defer func() {
		wg.Wait()
		if err := s.Close(); err != nil {
			logger.Info("ошибка при закрытии сервера:", err)
		} else {
			logger.Info("работа сервера успешно завершена")
		}

	}()

	if err := s.Start(ctx, logger, wg); err != nil {
		logger.Error(err)
		os.Exit(1)
	}

}
