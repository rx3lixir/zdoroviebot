package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken       string
	MongoURI       string
	DataBaseName   string
	CollectionName string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	return &Config{
		BotToken:       os.Getenv("BOT_TOKEN"),
		MongoURI:       os.Getenv("MONGO_URI"),
		DataBaseName:   os.Getenv("DB_NAME"),
		CollectionName: os.Getenv("COLLECTION_NAME"),
	}, nil
}
