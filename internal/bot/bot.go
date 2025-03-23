package bot

import (
	"github.com/charmbracelet/log"
	"github.com/rx3lixir/zdoroviebot/internal/db"
	"gopkg.in/telebot.v3"
)

// Bot представляет основную структуру бота
type Bot struct {
	handler *BotHandler
	telebot *telebot.Bot
	logger  *log.Logger
}

// NewBot создает экземпляр бота
func NewBot(token string, repo db.RecipeRepository, logger *log.Logger) (*Bot, error) {
	b, err := telebot.NewBot(telebot.Settings{
		Token: token,
		// Добавляем полинг и другие настройки
		Poller: &telebot.LongPoller{Timeout: 10},
	})

	if err != nil {
		return nil, err
	}

	handler := NewBotHandler(b, repo, logger)

	return &Bot{
		telebot: b,
		handler: handler,
		logger:  logger,
	}, nil
}

// Start запускает бота
func (b *Bot) Start() {
	b.handler.RegisterHandlers()
	b.logger.Info("Bot started and listening for messages")
	b.telebot.Start()
}

// Stop останавливает бота
func (b *Bot) Stop() {
	b.logger.Info("Stopping bot")
	b.telebot.Stop()
}

