package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type Config struct {
	RedisAddress  string
	RedisPassword string
	RedisDB       int

	HostIP string
}

var AppConfig *Config

func Init() {
	fmt.Println("Initializing config")

	AppConfig = &Config{
		RedisAddress:  "localhost:6379",
		RedisPassword: "password",
		RedisDB:       0,
		HostIP:        "localhost",
	}

	if redisAddress := os.Getenv("REDIS_ADDRESS"); redisAddress != "" {
		fmt.Printf("Using REDIS_ADDRESS=%s\n", redisAddress)
		AppConfig.RedisAddress = redisAddress
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		fmt.Printf("Using REDIS_PASSWORD=%s\n", redisPassword)
		AppConfig.RedisPassword = redisPassword
	}
	if redisDB := os.Getenv("REDIS_DB"); redisDB != "" {
		db, err := strconv.Atoi(redisDB)
		if err != nil {
			log.Printf("Error converting REDIS_DB to int: %v", err)
		} else {
			fmt.Printf("Using REDIS_DB=%d\n", db)
			AppConfig.RedisDB = db
		}
	}

	if hostIP := os.Getenv("HOST_IP"); hostIP != "" {
		fmt.Printf("Using HOST_IP=%s\n", hostIP)
		AppConfig.HostIP = hostIP
	}

}
