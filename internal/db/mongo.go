package db

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/rx3lixir/zdoroviebot/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RecipeRepo struct {
	collection *mongo.Collection
}

func NewRecipeRepo(client *mongo.Client, dbname, collectionName string) *RecipeRepo {
	return &RecipeRepo{
		collection: client.Database(dbname).Collection(collectionName),
	}
}

func ConnectMongo(uri string, logger *log.Logger) (*mongo.Client, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		logger.Error("Unable to establish connection to MongoDB client")
		return nil, err
	}

	if err = client.Ping(context.TODO(), nil); err != nil {
		logger.Error("Mongo client is not responding")
		return nil, err
	}

	logger.Info("Successfully connected to MondoDB")

	return client, nil
}

func (r *RecipeRepo) GetAllRecipes(logger *log.Logger) ([]models.Recipe, error) {
	logger.Info("Запрос к MongoDB...")

	var recipes []models.Recipe

	logger.Info("Запрос к MongoDB в коллекции:", r.collection.Name())

	cursor, err := r.collection.Find(context.TODO(), bson.M{})
	if err != nil {
		logger.Error("Ошибка запроса:", err)
		return nil, err
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &recipes); err != nil {
		logger.Error("Ошибка десериализации:", err)
		return nil, err
	}

	logger.Info("Получено рецептов:", len(recipes))
	return recipes, nil
}

func (r *RecipeRepo) GetRecipeByID(id primitive.ObjectID, logger *log.Logger) (*models.Recipe, error) {
	var recipe models.Recipe

	err := r.collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&recipe)
	if err != nil {
		logger.Error("Ошибка запроса рецепта по ID", id)
		return nil, err
	}

	return &recipe, nil
}

// Получение всех базовых ингредиентов
func (r *RecipeRepo) GetAllBaseIngredients(logger *log.Logger) ([]string, error) {
	var results []struct {
		BaseIngredients []string `bson:"base_ingredients"`
	}

	cursor, err := r.collection.Find(context.TODO(), bson.M{})
	if err != nil {
		logger.Error("Ошибка получения списка ингредиентов", err)
		return nil, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var res struct {
			BaseIngredients []string `bson:"base_ingredients"`
		}
		if err := cursor.Decode(&res); err != nil {
			continue
		}
		results = append(results, res)
	}

	// Собираем уникальные ингредиенты
	uniqueIngredients := make(map[string]struct{})
	for _, entry := range results {
		for _, ing := range entry.BaseIngredients {
			uniqueIngredients[ing] = struct{}{}
		}
	}

	var ingredientList []string
	for ing := range uniqueIngredients {
		ingredientList = append(ingredientList, ing)
	}

	return ingredientList, nil
}

// Поиск рецептов по ингредиентам
func (r *RecipeRepo) FindRecipesByIngredients(ingredients []string, logger *log.Logger) ([]models.Recipe, error) {
	var recipes []models.Recipe

	filter := bson.M{"base_ingredients": bson.M{"$all": ingredients}} // Рецепты, содержащие все указанные ингредиенты

	cursor, err := r.collection.Find(context.TODO(), filter)
	if err != nil {
		logger.Error("Ошибка поиска рецептов", err)
		return nil, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var recipe models.Recipe
		if err := cursor.Decode(&recipe); err == nil {
			recipes = append(recipes, recipe)
		}
	}

	return recipes, nil
}
