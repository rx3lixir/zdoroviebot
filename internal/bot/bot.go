package bot

import (
	"github.com/charmbracelet/log"
	"github.com/rx3lixir/zdoroviebot/internal/db"
	"gopkg.in/telebot.v3"
)

func InitBot(botToken string, repo *db.RecipeRepo, logger *log.Logger) (*telebot.Bot, error) {
	logger.Info("Creating bot entity...")
	b, err := telebot.NewBot(telebot.Settings{
		Token: botToken,
	})
	if err != nil {
		logger.Error("Error creating bot entity")
		return nil, err
	}

	SetupHandlers(b, repo, logger)

	logger.Info("Bot is successfully initialized")

	return b, nil
}
