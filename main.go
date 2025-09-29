package main

import (
	"context"
	"os"
	"traefik-multi-hosts/cmd/mhos"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/docker"
	"traefik-multi-hosts/internal/listeners"
	"traefik-multi-hosts/internal/logging"
	"traefik-multi-hosts/internal/redis"
	"traefik-multi-hosts/internal/web"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const Version = "1.2.0"

func main() {
	ctx := context.Background()

	logging.Init()
	config.Init()
	zerolog.SetGlobalLevel(config.LogLevel())

	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		mhos.Healthcheck()
		return
	}

	log.Info().Msgf("Starting traefik-mhos v%s", Version)

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
	log.Info().Msg("Starting traefik-mhos")

	redisClient.CleanCurrentServices()
	err = mhos.FreshScan(dockerClient, redisClient)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to perform initial fresh scan")
		return
	}
	if config.ListenEvents() {
		go func() {
			listeners.ListenForContainersEvent(ctx, dockerClient, redisClient)
		}()
		go func() {
			web.Serve(dockerClient, redisClient)
		}()
		select {}
	}

}
