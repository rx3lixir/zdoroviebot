package db

import (
	"context"
	"time"

	"github.com/rx3lixir/zdoroviebot/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RecipeRepository определяет интерфейс для работы с рецептами
type RecipeRepository interface {
	GetAllRecipes(ctx context.Context) ([]models.Recipe, error)
	GetRecipeByID(ctx context.Context, id primitive.ObjectID) (*models.Recipe, error)
	GetAllBaseIngredients(ctx context.Context) ([]string, error)
	FindRecipesByIngredients(ctx context.Context, ingredients []string) ([]models.Recipe, error)
}

// MongoRecipeRepo имплементирует интерфейс RecipeRepository
type MongoRecipeRepo struct {
	collection *mongo.Collection
}

// NewMongoRecipeRepo создает новый репозиторий рецептов
func NewMongoRecipeRepo(client *mongo.Client, dbname, collectionName string) *MongoRecipeRepo {
	return &MongoRecipeRepo{
		collection: client.Database(dbname).Collection(collectionName),
	}
}

// ConnectMongo устанавливает соединение с MongoDB
func ConnectMongo(ctx context.Context, uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}

// DisconnectMongo закрывает соединение с MongoDB
func DisconnectMongo(ctx context.Context, client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.Disconnect(ctx)
}

// GetAllRecipes возвращает все рецепты из базы данных
func (r *MongoRecipeRepo) GetAllRecipes(ctx context.Context) ([]models.Recipe, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var recipes []models.Recipe

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &recipes); err != nil {
		return nil, err
	}

	return recipes, nil
}

// GetRecipeByID возвращает рецепт по его ID
func (r *MongoRecipeRepo) GetRecipeByID(ctx context.Context, id primitive.ObjectID) (*models.Recipe, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var recipe models.Recipe

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&recipe)
	if err != nil {
		return nil, err
	}

	return &recipe, nil
}

// GetAllBaseIngredients возвращает все уникальные базовые ингредиенты
func (r *MongoRecipeRepo) GetAllBaseIngredients(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Агрегация для получения уникальных ингредиентов
	pipeline := []bson.M{
		{"$unwind": "$base_ingredients"},
		{"$group": bson.M{"_id": "$base_ingredients"}},
		{"$sort": bson.M{"_id": 1}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID string `bson:"_id"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Преобразуем результаты в срез строк
	ingredients := make([]string, 0, len(results))
	for _, r := range results {
		ingredients = append(ingredients, r.ID)
	}

	return ingredients, nil
}

// FindRecipesByIngredients ищет рецепты, содержащие все указанные ингредиенты
func (r *MongoRecipeRepo) FindRecipesByIngredients(ctx context.Context, ingredients []string) ([]models.Recipe, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var recipes []models.Recipe

	filter := bson.M{"base_ingredients": bson.M{"$all": ingredients}}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &recipes); err != nil {
		return nil, err
	}

	return recipes, nil
}

