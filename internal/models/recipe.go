package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Nutrition struct {
	Calories int     `bson:"calories"`
	Proteins float64 `bson:"proteins"`
	Fats     float64 `bson:"fats"`
	Carbs    float64 `bson:"carbs"`
}

type Recipe struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	Name             string             `bson:"name"`
	BaseIngredients  []string           `bson:"base_ingredients"`
	ExtraIngredients []string           `bson:"extra_ingredients"`
	Procedure        string             `bson:"cooking_procedure"`
	Nutrition        Nutrition          `bson:"nutrition"`
}
