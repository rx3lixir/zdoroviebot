package main

import (
	"os"

	"github.com/rx3lixir/zdoroviebot/config"
	"github.com/rx3lixir/zdoroviebot/internal/bot"
	"github.com/rx3lixir/zdoroviebot/internal/db"
	"github.com/rx3lixir/zdoroviebot/internal/lib/logger"
)

func main() {
	logger := logger.InitLogger()

	logger.Info("Getting config file...")

	appConfig, err := config.LoadConfig()
	if err != nil {
		logger.Error("Error loading config file", err)
		os.Exit(1)
	}

	logger.Info("Success!")

	logger.Info("Connecting to MongoDB...")

	client, err := db.ConnectMongo(appConfig.MongoURI, logger)
	if err != nil {
		logger.Error("Error connecting to MongoDB", err)
		os.Exit(1)
	}

	logger.Info("Success!")

	repo := db.NewRecipeRepo(client, appConfig.DataBaseName, appConfig.CollectionName)

	logger.Info("Ititializing bot...")

	telegramBot, err := bot.InitBot(appConfig.BotToken, repo, logger)
	if err != nil {
		logger.Error("Error initializing telegram bot", err)
		os.Exit(1)
	}

	logger.Info("Bot has been started!")

	telegramBot.Start()
}
