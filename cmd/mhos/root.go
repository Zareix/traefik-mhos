package mhos

import (
	"context"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/docker"
	"traefik-multi-hosts/internal/listeners"
	"traefik-multi-hosts/internal/log"
	"traefik-multi-hosts/internal/redis"
	"traefik-multi-hosts/internal/traefik"

	"github.com/rs/zerolog"
)

func Run() {
	ctx := context.Background()

	zerolog.SetGlobalLevel(config.LogLevel())

	log.Info().Msg("Starting traefik-mhos")

	FreshScan(ctx)

	if config.ListenEvents() {
		listeners.ListenForNewContainers(ctx)
		listeners.ListenForStoppedContainers(ctx)
	}
}

func FreshScan(ctx context.Context) error {
	log.Info().Msg("Running first start with existing containers")

	redis.CleanCurrentServices(ctx)

	containers, err := docker.ListContainers()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list running containers")
		return err
	}
	log.Debug().Int("containers", len(containers)).Msg("Found running containers")

	for _, container := range containers {
		err := traefik.AddContainerToTraefik(ctx, container.ID)
		if err != nil {
			log.Error().Err(err).Str("container", container.ID).Msg("Failed to add container to traefik")
			return err
		}
	}

	redis.Cleanup(ctx)
	return nil
}
