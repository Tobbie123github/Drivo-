package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	PORT             string
	DB_HOST          string
	DB_USER          string
	DB_PASSWORD      string
	DB_PORT          string
	DB_NAME          string
	JWTSecret        string
	CLOUDINARY_URL   string
	REDIS_URL        string
	MailjetAPIKey    string
	MailjetSecretKey string
	MailjetFromEmail string
	MailjetFromName  string
}

func Load() (Config, error) {

	godotenv.Load()

	port, err := extractText("PORT")

	if err != nil {
		return Config{}, fmt.Errorf("Port cant be empty")
	}

	db_host, err := extractText("DB_HOST")

	if err != nil {
		return Config{}, fmt.Errorf("cannot find host")
	}
	db_port, err := extractText("DB_PORT")

	if err != nil {
		return Config{}, fmt.Errorf("cannot find port")
	}
	db_password, err := extractText("DB_PASSWORD")

	if err != nil {
		return Config{}, fmt.Errorf("cannot find password")
	}

	db_name, err := extractText("DB_NAME")

	if err != nil {
		return Config{}, fmt.Errorf("cannot find databae name")
	}

	db_user, err := extractText("DB_user")

	if err != nil {
		return Config{}, fmt.Errorf("cannot find host")
	}

	jwtSecret, err := extractText("JWTSecret")

	if err != nil {
		return Config{}, fmt.Errorf("cannot find host")
	}

	cloudinary, err := extractText("CLOUDINARY_URL")

	if err != nil {
		return Config{}, fmt.Errorf("cloudinary not found")
	}

	redisURL, err := extractText("REDIS_URL")

	if err != nil {
		return Config{}, fmt.Errorf("redis url not found")
	}

	mailjetAPIKey, err := extractText("MAILJET_API_KEY")

	if err != nil {
		return Config{}, fmt.Errorf("Mailjet API key not found")
	}

	mailjetSecretKey, err := extractText("MAILJET_SECRET_KEY")

	if err != nil {
		return Config{}, fmt.Errorf("Mailjet secret key not found")
	}

	mailjetFromEmail, err := extractText("MAILJET_FROM_EMAIL")

	if err != nil {
		return Config{}, fmt.Errorf("Mailjet from email not found")
	}

	mailJetFromName, err := extractText("MAILJET_FROM_NAME")

	if err != nil {
		return Config{}, fmt.Errorf("Mailjet from name not found")
	}

	return Config{
		PORT:             port,
		DB_HOST:          db_host,
		DB_USER:          db_user,
		DB_PASSWORD:      db_password,
		DB_PORT:          db_port,
		DB_NAME:          db_name,
		JWTSecret:        jwtSecret,
		CLOUDINARY_URL:   cloudinary,
		REDIS_URL:        redisURL,
		MailjetAPIKey:    mailjetAPIKey,
		MailjetSecretKey: mailjetSecretKey,
		MailjetFromEmail: mailjetFromEmail,
		MailjetFromName:  mailJetFromName,
	}, nil
}

func extractText(key string) (string, error) {

	cf := os.Getenv(key)

	if cf == "" {
		return "", fmt.Errorf("Cant be empty")
	}

	cfg := strings.TrimSpace(cf)

	return cfg, nil
}
