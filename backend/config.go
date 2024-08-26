package main

import (
	"log"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Env struct {
	SERVER_PORT string `env:"SERVER_PORT,required"`
	API_KEY     string `env:"API_KEY,required"`

	// PostgreSQL connection
	DB_HOST     string `env:"DB_HOST,required"`
	DB_NAME     string `env:"DB_NAME,required"`
	DB_USER     string `env:"DB_USER,required"`
	DB_PASSWORD string `env:"DB_PASSWORD,required"`
	DB_SSLMODE  string `env:"DB_SSLMODE,required"`
}

// Init all the env variables
func EnvConfig() *Env {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Unable to load .env file: %e", err)
	}

	config := &Env{}

	if err := env.Parse(config); err != nil {
		log.Fatalf("Unable to load varialbes from the .env file: %e", err)
	}

	return config
}
