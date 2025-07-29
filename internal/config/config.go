package config

import (
	"os"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type config struct {
	redisAddress  string
	redisPassword string
	redisDB       int

	hostIP string
	port   string

	logLevel zerolog.Level

	listenEvents bool
}

var appConfig *config

func Init() {
	log.Info().Msg("Initializing config")

	appConfig = &config{
		redisAddress:  "localhost:6379",
		redisPassword: "",
		redisDB:       0,
		hostIP:        "localhost",
		port:          "8888",
		logLevel:      zerolog.InfoLevel,
		listenEvents:  true,
	}

	if redisAddress := os.Getenv("REDIS_ADDRESS"); redisAddress != "" {
		log.Info().Str("REDIS_ADDRESS", redisAddress).Msg("Using REDIS_ADDRESS env var")
		appConfig.redisAddress = redisAddress
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		log.Info().Str("REDIS_PASSWORD", "***").Msg("Using REDIS_PASSWORD env var")
		appConfig.redisPassword = redisPassword
	}
	if redisDB := os.Getenv("REDIS_DB"); redisDB != "" {
		db, err := strconv.Atoi(redisDB)
		if err != nil {
			log.Fatal().Err(err).Msg("Error converting REDIS_DB to int")
		} else {
			log.Info().Int("REDIS_DB", db).Msg("Using REDIS_DB env var")
			appConfig.redisDB = db
		}
	}

	if hostIP := os.Getenv("HOST_IP"); hostIP != "" {
		log.Info().Str("HOST_IP", hostIP).Msg("Using HOST_IP env var")
		appConfig.hostIP = hostIP
	}

	if port := os.Getenv("PORT"); port != "" {
		log.Info().Str("PORT", port).Msg("Using PORT env var")
		appConfig.port = port
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		level, err := zerolog.ParseLevel(logLevel)
		if err != nil {
			log.Fatal().Err(err).Msg("Error parsing LOG_LEVEL")
		} else {
			log.Info().Str("LOG_LEVEL", level.String()).Msg("Using LOG_LEVEL env var")
			appConfig.logLevel = level
		}
	}

	if listenEvents := os.Getenv("LISTEN_EVENTS"); listenEvents != "" {
		listen, err := strconv.ParseBool(listenEvents)
		if err != nil {
			log.Fatal().Err(err).Msg("Error parsing LISTEN_EVENTS")
		} else {
			log.Info().Bool("LISTEN_EVENTS", listen).Msg("Using LISTEN_EVENTS env var")
			appConfig.listenEvents = listen
		}
	}

}

func RedisAddress() string {
	return appConfig.redisAddress
}

func RedisPassword() string {
	return appConfig.redisPassword
}

func RedisDB() int {
	return appConfig.redisDB
}

func HostIP() string {
	return appConfig.hostIP
}

func Port() string {
	return appConfig.port
}

func LogLevel() zerolog.Level {
	return appConfig.logLevel
}

func ListenEvents() bool {
	return appConfig.listenEvents
}
