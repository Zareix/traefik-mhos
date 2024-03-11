package cmd

import (
	"context"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/listeners"
	"traefik-multi-hosts/internal/log"
	"traefik-multi-hosts/internal/redis"
	"traefik-multi-hosts/internal/traefik"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/rs/zerolog"
)

func Run() {
	ctx := context.Background()

	zerolog.SetGlobalLevel(config.LogLevel())

	log.Info().Msg("Starting traefik-mhos")

	redisClient := redis.NewClient(ctx, config.RedisAddress(), config.RedisPassword(), config.RedisDB())
	redisClient.Del(ctx, "mhos:"+config.HostIP())

	dockerClient, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create docker client")
	}

	log.Info().Msg("Running first start with existing containers")
	filters := filters.NewArgs()
	filters.Add("status", "running")
	filters.Add("label", "traefik.enable=true")
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{
		Filters: filters,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list running containers")
	}
	log.Debug().Int("containers", len(containers)).Msg("Found running containers")

	for _, container := range containers {
		traefik.AddContainerToTraefik(ctx, dockerClient, container.ID)
	}

	redis.Cleanup(ctx)

	listeners.ListenForNewContainers(ctx, dockerClient)
	listeners.ListenForStoppedContainers(ctx, dockerClient)
}
