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

const Version = "1.3.0"

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

	redisClient, err := redis.New(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create redis client")
		return
	}
	defer redisClient.Close()
	log.Info().Msg("Starting traefik-mhos")

	if len(config.DockerHosts()) > 0 {
		log.Info().Msgf("Monitoring docker hosts: %v", config.DockerHosts())

		if config.Mode() == config.PullMode {
			log.Info().Msg("Operating in PULL mode")
			redisClient.CleanAllServices()
		} else {
			log.Info().Msg("Operating in PUSH mode")
		}

		for _, host := range config.DockerHosts() {
			go func(host string) {
				dockerClient, err := docker.New(ctx, host)
				if err != nil {
					log.Fatal().Err(err).Msg("Failed to create docker client")
					return
				}
				defer dockerClient.Close()
				redisClient.CleanCurrentServices(config.HostIP())
				err = mhos.FreshScan(dockerClient, redisClient)
				if config.ListenEvents() {
					listeners.ListenForContainersEvent(ctx, dockerClient, redisClient)
				}
			}(host)
		}
		// TODO web server
		if config.ListenEvents() {
			go func() {
				web.Serve(redisClient)
			}()
		}
		select {}
	} else {
		log.Warn().Msg("No docker hosts configured, exiting")
	}

}
