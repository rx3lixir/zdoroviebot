package bot

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/rx3lixir/zdoroviebot/internal/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/telebot.v3"
)

func SetupHandlers(b *telebot.Bot, repo *db.RecipeRepo, logger *log.Logger) error {
	b.Handle("/recipes", func(c telebot.Context) error {
		recipes, err := repo.GetAllRecipes(logger)
		if err != nil {
			logger.Error("Error processing '/recipes' handler", err)
			return c.Send("Ошибка при получении рецептов")
		}
		if len(recipes) == 0 {
			return c.Send("Рецептов пока нет")
		}

		// Создаем разметку
		markup := &telebot.ReplyMarkup{}

		// Создаем кнопки и добавляем их в разметку по 2 в ряд
		var currentRow []telebot.InlineButton

		for i, r := range recipes {
			button := telebot.InlineButton{
				Text: r.Name,
				Data: "recipe_" + r.ID.Hex(),
			}

			currentRow = append(currentRow, button)

			// Если в текущем ряду уже 2 кнопки или это последний рецепт, добавляем ряд в инлайн-клавиатуру
			if len(currentRow) == 2 || i == len(recipes)-1 {
				markup.InlineKeyboard = append(markup.InlineKeyboard, currentRow)
				currentRow = []telebot.InlineButton{} // Сбрасываем текущий ряд
			}
		}

		return c.Send("Выберите рецепт:", markup)
	})

	b.Handle(telebot.OnCallback, func(c telebot.Context) error {
		data := c.Callback().Data // Получаем callback дату

		logger.Info("Received callback data:", data) // Логируем

		recipeID := strings.TrimPrefix(data, "recipe_")

		logger.Info("Extracted recipe ID:", recipeID) // Логируем после обрезки префикса

		// Проверяем, пуст ли ID
		if recipeID == "" {
			logger.Error("Empty recipe ID in callback", recipeID)
			return c.Respond()
		}

		// Ищем рецепт в базе по ID
		objID, err := primitive.ObjectIDFromHex(recipeID)
		if err != nil {
			logger.Error("Ошибка конвертации ID", err)
			return c.Respond()
		}

		recipe, err := repo.GetRecipeByID(objID, logger)
		if err != nil {
			logger.Error("Ошибка получения рецепта", recipe)
			return c.Respond()
		}

		// Формируем сообщение с полной информацией
		detailsRecipe := fmt.Sprintf(
			"*%s*\n\n🍽 *Ингредиенты*:\n%s\n\n🧂 *Дополнительные ингредиенты*:\n%s\n\n🔥 *Способ приготовления*:\n%s\n\n⚡ *Пищевая ценность*:\nКалории: %d ккал\nБелки: %.1f г\nЖиры: %.1f г\nУглеводы: %.1f г",
			recipe.Name,
			strings.Join(recipe.BaseIngredients, ", "),
			strings.Join(recipe.ExtraIngredients, ", "),
			recipe.Procedure,
			recipe.Nutrition.Calories,
			recipe.Nutrition.Proteins,
			recipe.Nutrition.Fats,
			recipe.Nutrition.Carbs,
		)

		return c.Edit(detailsRecipe, telebot.ModeMarkdown)

	})

	b.Handle("/start", func(c telebot.Context) error {
		return c.Send("Все работает")
	})
	return nil
}
