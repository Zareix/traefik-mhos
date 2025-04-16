package main

import (
	"context"
	"traefik-multi-hosts/cmd/mhos"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/docker"
	"traefik-multi-hosts/internal/logging"
	"traefik-multi-hosts/internal/redis"
	"traefik-multi-hosts/internal/web"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()

	logging.Init()
	config.Init()
	zerolog.SetGlobalLevel(config.LogLevel())

	dockerClient, err := docker.New(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create docker client")
		return
	}
	defer dockerClient.Close()

	redisClient, err := redis.New(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create redis client")
		return
	}
	defer redisClient.Close()

	go func() {
		web.Serve(*dockerClient, *redisClient)
	}()
	go func() {
		mhos.Run(ctx, *dockerClient, *redisClient)
	}()

	select {}
}
