package cmd

import (
	"context"
	"fmt"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/listeners"
	"traefik-multi-hosts/internal/redis"
	"traefik-multi-hosts/internal/traefik"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
)

func Run() {
	ctx := context.Background()

	redisClient := redis.NewClient(config.AppConfig.RedisAddress, config.AppConfig.RedisPassword, config.AppConfig.RedisDB)
	defer redisClient.Close()
	// redisClient.FlushDB(ctx)

	dockerClient, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	fmt.Println("Running first start with existing containers")
	filters := filters.NewArgs()
	filters.Add("status", "running")
	filters.Add("label", "traefik.enable=true")
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{
		Filters: filters,
	})
	if err != nil {
		panic("Initial container list failed: " + err.Error())
	}

	for _, container := range containers {
		traefik.AddContainerToTraefik(ctx, dockerClient, container.ID)
	}

	redis.Cleanup(ctx)

	listeners.ListenForNewContainers(ctx, dockerClient)
	listeners.ListenForStoppedContainers(ctx, dockerClient)

	select {}
}
