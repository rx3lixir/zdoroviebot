package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rx3lixir/zdoroviebot/config"
	"github.com/rx3lixir/zdoroviebot/internal/bot"
	"github.com/rx3lixir/zdoroviebot/internal/db"
	"github.com/rx3lixir/zdoroviebot/internal/lib/logger"
)

func main() {
	logger := logger.InitLogger()
	ctx := context.Background()

	logger.Info("Loading config...")
	appConfig, err := config.LoadConfig()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Info("Connecting to MongoDB...")
	client, err := db.ConnectMongo(ctx, appConfig.MongoURI)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	// Важно: освобождаем ресурсы при завершении
	defer func() {
		if err := db.DisconnectMongo(ctx, client); err != nil {
			logger.Error("Failed to disconnect from MongoDB", "error", err)
		}
	}()

	repo := db.NewMongoRecipeRepo(client, appConfig.DataBaseName, appConfig.CollectionName)

	logger.Info("Initializing bot...")
	telegramBot, err := bot.NewBot(appConfig.BotToken, repo, logger)
	if err != nil {
		logger.Error("Failed to initialize bot", "error", err)
		os.Exit(1)
	}

	// Запускаем бота в отдельной горутине
	go telegramBot.Start()
	logger.Info("Bot started successfully!")

	// Обрабатываем сигналы завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down...")
	telegramBot.Stop()
}

