package docker

import (
	"context"
	"traefik-multi-hosts/internal/log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	dockerSDK "github.com/docker/docker/client"
)

var dockerClient *dockerSDK.Client

func init() {
	var err error
	dockerClient, err = dockerSDK.NewClientWithOpts(dockerSDK.FromEnv, dockerSDK.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create docker client")
	}
}

func ListContainers() ([]types.Container, error) {
	ctx := context.Background()

	filters := filters.NewArgs()
	filters.Add("status", "running")
	filters.Add("label", "traefik.enable=true")
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{
		Filters: filters,
	})
	return containers, err
}

func InspectContainer(ctx context.Context, containerId string) (types.ContainerJSON, error) {
	return dockerClient.ContainerInspect(ctx, containerId)
}

func Events(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {
	return dockerClient.Events(ctx, options)
}
