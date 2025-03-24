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

// Константы для префиксов callback-данных
const (
	recipePrefix     = "recipe_"
	ingredientPrefix = "ingr_"
	searchPrefix     = "search_"
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

// Храним выбранные ингредиенты для каждого пользователя
// В реальном приложении лучше использовать Redis или другое хранилище
var userIngredients = make(map[int64]map[string]bool)

// RegisterHandlers регистрирует все обработчики команд
func (h *BotHandler) RegisterHandlers() {
	h.bot.Handle("/start", h.handleStart)
	h.bot.Handle("/recipes", h.handleRecipes)
	h.bot.Handle("/ingredients", h.handleIngredients) // Новый обработчик
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
	userID := c.Sender().ID

	// -- ОБРАБОТКА ВЫБОРА ИНГРИДИЕНТОВ -- \\
	if strings.HasPrefix(data, ingredientPrefix) {
		ingredient := strings.TrimPrefix(data, ingredientPrefix)

		// Переключаем состояние выбора ингридиента
		if userIngredients[userID] == nil {
			userIngredients[userID] = make(map[string]bool)
		}

		userIngredients[userID][ingredient] = !userIngredients[userID][ingredient]

		// Получаем все ингридиенты для обновления клавиатуры
		ingredients, err := h.repo.GetAllBaseIngredients(ctx)
		if err != nil {
			h.logger.Error("Failed to get all base ingredients", "error", err)
			return c.Respond()
		}

		markup := h.createIngredientsKeyboard(ingredients, userID)
		return c.Edit("Выберите ингридиенты, которые у вас есть", markup)
	}

	// -- ОБРАБОТКА ПО ВЫБРАННЫМ ИГРИДИЕНТАМ -- \\
	if strings.HasPrefix(data, searchPrefix) {

		// Собираемм выбранные ингридиенты
		var selectedIngredients []string

		for ingredient, selected := range userIngredients[userID] {
			if selected {
				selectedIngredients = append(selectedIngredients, ingredient)
			}
		}

		if len(selectedIngredients) == 0 {
			return c.Respond(&telebot.CallbackResponse{
				Text:      "Выберите хотя бы один ингридиент",
				ShowAlert: true,
			})
		}

		recipes, err := h.repo.FindRecipesByIngredients(ctx, selectedIngredients)
		if err != nil {
			h.logger.Error("Failed to find recipes by ingredients", "error", err)
			return c.Respond(&telebot.CallbackResponse{
				Text:      "Ошибка при поиске рецептов",
				ShowAlert: true,
			})
		}

		if len(recipes) == 0 {
			return c.Edit(fmt.Sprintf("К сожалению, рецептов с ингридиентами %s не найдено.", strings.Join(selectedIngredients, ", ")))
		}

		// Показываем найденные рецепты
		resultMessage := fmt.Sprintf("Найдено %d рецептов с ингридиентами: \n%s",
			len(recipes),
			strings.Join(selectedIngredients, ", "))

		markup := h.createRecipesKeyboard(recipes)
		return c.Edit(resultMessage, markup)
	}

	// -- Обработка для рецептов-- \\
	if strings.HasPrefix(data, recipePrefix) {

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

	h.logger.Warn("Unknown callback data", "data", data)
	return c.Respond()
}

func (h *BotHandler) handleIngredients(c telebot.Context) error {
	ctx := context.Background()

	ingredients, err := h.repo.GetAllBaseIngredients(ctx)
	if err != nil {
		h.logger.Error("Failed to get base ingredients", "error", err)
		return c.Send("Не удалось получить список ингредиентов. Попробуйте позже.")
	}

	if len(ingredients) == 0 {
		return c.Send("Ингридиентов пока нет. Загляните позже!")
	}

	// Инициализация карты выбарнных ингридиентов для пользователя (костыль пиздец)
	userID := c.Sender().ID
	userIngredients[userID] = make(map[string]bool)

	markup := h.createIngredientsKeyboard(ingredients, userID)
	return c.Send("Выберите ингридиенты, которые у вас есть:", markup)
}

func (h *BotHandler) createIngredientsKeyboard(ingredients []string, userID int64) *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}

	// Две кнопки в ряд
	const buttonsPerRow = 2
	var currentRow []telebot.InlineButton

	// Создаем кнопки для каждого ингридиента
	for i, ingredient := range ingredients {
		selected := userIngredients[userID][ingredient]

		// Галка для выбранных ингридиентов
		buttonText := ingredient
		if selected {
			buttonText = "✅ " + ingredient
		}

		button := telebot.InlineButton{
			Text: buttonText,
			Data: fmt.Sprintf("%s%s", ingredientPrefix, ingredient),
		}

		currentRow = append(currentRow, button)

		// Если ряд заполнен или это последний ингридиент, добавляем ряд в клавиатуру
		if len(currentRow) == buttonsPerRow || i == len(ingredients)-1 {
			markup.InlineKeyboard = append(markup.InlineKeyboard, currentRow)
			currentRow = []telebot.InlineButton{}
		}

	}
	searchButton := telebot.InlineButton{
		Text: "🔍 Найти рецепты",
		Data: fmt.Sprintf("%s%d", searchPrefix, userID),
	}

	// Отдельный ряд только для кнопки поиска рецептов
	searchRow := []telebot.InlineButton{searchButton}
	markup.InlineKeyboard = append(markup.InlineKeyboard, searchRow)

	return markup
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
			"🧂 *Ингридиенты:*\n%s, %s\n\n"+
			"🔥 *Способ приготовления*:\n%s\n\n"+
			"⚡ *Пищевая ценность*:\n"+
			"Калории: %d ккал\n"+
			"Белки: %.1f г\n"+
			"Жиры: %.1f г\n"+
			"Углеводы: %.1f г",
		recipe.Name,
		strings.Join(recipe.ExtraIngredients, ", "),
		strings.Join(recipe.BaseIngredients, ", "),
		recipe.Procedure,
		recipe.Nutrition.Calories,
		recipe.Nutrition.Proteins,
		recipe.Nutrition.Fats,
		recipe.Nutrition.Carbs,
	)
}
