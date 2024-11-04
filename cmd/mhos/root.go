package mhos

import (
	"context"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/docker"
	"traefik-multi-hosts/internal/listeners"
	"traefik-multi-hosts/internal/redis"
	"traefik-multi-hosts/internal/traefik"

	"github.com/rs/zerolog/log"
)

func Run(ctx context.Context, dockerClient docker.DockerClient, redisClient redis.RedisClient) {
	log.Info().Msg("Starting traefik-mhos")
	FreshScan(dockerClient, redisClient)

	if config.ListenEvents() {
		listeners.ListenForContainersEvent(ctx, dockerClient, redisClient)
	}
}

func FreshScan(dockerClient docker.DockerClient, redisClient redis.RedisClient) error {
	log.Info().Msg("Running first start with existing containers")

	redisClient.CleanCurrentServices()

	containers, err := dockerClient.ListContainers()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list running containers")
		return err
	}
	log.Debug().Int("containers", len(containers)).Msg("Found running containers")

	for _, container := range containers {
		err := traefik.AddContainerToTraefik(dockerClient, redisClient, container.ID)
		if err != nil {
			log.Error().Err(err).Str("container", container.ID).Msg("Failed to add container to traefik")
			return err
		}
	}

	redisClient.Cleanup()
	return nil
}
