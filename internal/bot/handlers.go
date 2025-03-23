package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/rx3lixir/zdoroviebot/internal/db"
	"github.com/rx3lixir/zdoroviebot/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/telebot.v3"
)

// BotHandler обрабатывает команды для телеграм-бота
type BotHandler struct {
	bot        *telebot.Bot
	repo       db.RecipeRepository
	logger     *log.Logger
	msgBuilder *MessageBuilder
}

// NewBotHandler создает новый обработчик команд
func NewBotHandler(bot *telebot.Bot, repo db.RecipeRepository, logger *log.Logger) *BotHandler {
	return &BotHandler{
		bot:        bot,
		repo:       repo,
		logger:     logger,
		msgBuilder: NewMessageBuilder(),
	}
}

// RegisterHandlers регистрирует все обработчики команд
func (h *BotHandler) RegisterHandlers() {
	h.bot.Handle("/start", h.handleStart)
	h.bot.Handle("/recipes", h.handleRecipes)
	h.bot.Handle(telebot.OnCallback, h.handleCallback)
}

// handleStart обрабатывает команду /start
func (h *BotHandler) handleStart(c telebot.Context) error {
	welcomeMsg := "👋 Добро пожаловать в ПП Бот Алтайские продукты +100 к здоровью!\n\n" +
		"🌱 Здесь вы найдете полезные рецепты с использованием натуральных ингредиентов.\n" +
		"🔍 Воспользуйтесь командой /recipes, чтобы просмотреть все доступные рецепты."

	return c.Send(welcomeMsg)
}

// handleRecipes обрабатывает команду /recipes
func (h *BotHandler) handleRecipes(c telebot.Context) error {
	ctx := context.Background()

	recipes, err := h.repo.GetAllRecipes(ctx)
	if err != nil {
		h.logger.Error("failed to get recipes", "error", err)
		return c.Send("⚠️ Не удалось получить список рецептов. Попробуйте позже.")
	}

	if len(recipes) == 0 {
		return c.Send("📭 Рецептов пока нет. Загляните позже!")
	}

	markup := h.createRecipesKeyboard(recipes)
	return c.Send("📋 Выберите рецепт:", markup)
}

// handleCallback обрабатывает нажатия на инлайн-кнопки
func (h *BotHandler) handleCallback(c telebot.Context) error {
	ctx := context.Background()
	data := c.Callback().Data

	// Используем константы для префиксов
	const recipePrefix = "recipe_"

	if !strings.HasPrefix(data, recipePrefix) {
		h.logger.Warn("unknown callback data", "data", data)
		return c.Respond()
	}

	recipeID := strings.TrimPrefix(data, recipePrefix)

	objID, err := primitive.ObjectIDFromHex(recipeID)
	if err != nil {
		h.logger.Error("invalid recipe ID format", "id", recipeID, "error", err)
		return c.Respond(&telebot.CallbackResponse{
			Text:      "⚠️ Неверный формат ID рецепта",
			ShowAlert: true,
		})
	}

	recipe, err := h.repo.GetRecipeByID(ctx, objID)
	if err != nil {
		h.logger.Error("failed to get recipe by ID", "id", recipeID, "error", err)
		return c.Respond(&telebot.CallbackResponse{
			Text:      "⚠️ Рецепт не найден",
			ShowAlert: true,
		})
	}

	detailsMsg := h.msgBuilder.BuildRecipeDetails(recipe)
	return c.Edit(detailsMsg, telebot.ModeMarkdown)
}

// createRecipesKeyboard создает инлайн-клавиатуру с рецептами
func (h *BotHandler) createRecipesKeyboard(recipes []models.Recipe) *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}

	// Добавляем кнопки по 2 в ряд
	const buttonsPerRow = 2
	var currentRow []telebot.InlineButton

	for i, recipe := range recipes {
		button := telebot.InlineButton{
			Text: recipe.Name,
			Data: fmt.Sprintf("recipe_%s", recipe.ID.Hex()),
		}

		currentRow = append(currentRow, button)

		if len(currentRow) == buttonsPerRow || i == len(recipes)-1 {
			markup.InlineKeyboard = append(markup.InlineKeyboard, currentRow)
			currentRow = []telebot.InlineButton{}
		}
	}

	return markup
}

// MessageBuilder формирует сообщения для пользователя
type MessageBuilder struct{}

// NewMessageBuilder создает новый билдер сообщений
func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{}
}

// BuildRecipeDetails формирует детальную информацию о рецепте
func (mb *MessageBuilder) BuildRecipeDetails(recipe *models.Recipe) string {
	return fmt.Sprintf(
		"*%s*\n\n"+
			"🍽 *Ингредиенты*:\n%s\n\n"+
			"🧂 *Дополнительные ингредиенты*:\n%s\n\n"+
			"🔥 *Способ приготовления*:\n%s\n\n"+
			"⚡ *Пищевая ценность*:\n"+
			"Калории: %d ккал\n"+
			"Белки: %.1f г\n"+
			"Жиры: %.1f г\n"+
			"Углеводы: %.1f г",
		recipe.Name,
		strings.Join(recipe.BaseIngredients, ", "),
		strings.Join(recipe.ExtraIngredients, ", "),
		recipe.Procedure,
		recipe.Nutrition.Calories,
		recipe.Nutrition.Proteins,
		recipe.Nutrition.Fats,
		recipe.Nutrition.Carbs,
	)
}
