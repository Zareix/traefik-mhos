package config

import (
	"os"
	"strconv"
	"traefik-multi-hosts/internal/log"

	"github.com/rs/zerolog"
)

type Config struct {
	RedisAddress  string
	RedisPassword string
	RedisDB       int

	HostIP string

	LogLevel zerolog.Level
}

var AppConfig *Config

func Init() {
	log.Info().Msg("Initializing config")

	AppConfig = &Config{
		RedisAddress:  "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
		HostIP:        "localhost",
		LogLevel:      zerolog.InfoLevel,
	}

	if redisAddress := os.Getenv("REDIS_ADDRESS"); redisAddress != "" {
		log.Info().Str("REDIS_ADDRESS", redisAddress).Msg("Using REDIS_ADDRESS env var")
		AppConfig.RedisAddress = redisAddress
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		log.Info().Str("REDIS_PASSWORD", "***").Msg("Using REDIS_PASSWORD env var")
		AppConfig.RedisPassword = redisPassword
	}
	if redisDB := os.Getenv("REDIS_DB"); redisDB != "" {
		db, err := strconv.Atoi(redisDB)
		if err != nil {
			log.Fatal().Err(err).Msg("Error converting REDIS_DB to int")
		} else {
			log.Info().Int("REDIS_DB", db).Msg("Using REDIS_DB env var")
			AppConfig.RedisDB = db
		}
	}

	if hostIP := os.Getenv("HOST_IP"); hostIP != "" {
		log.Info().Str("HOST_IP", hostIP).Msg("Using HOST_IP env var")
		AppConfig.HostIP = hostIP
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		level, err := zerolog.ParseLevel(logLevel)
		if err != nil {
			log.Fatal().Err(err).Msg("Error parsing LOG_LEVEL")
		} else {
			log.Info().Str("LOG_LEVEL", level.String()).Msg("Using LOG_LEVEL env var")
			AppConfig.LogLevel = level
		}
	}

}
