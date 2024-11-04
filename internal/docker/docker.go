package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
)

type DockerClient struct {
	ctx context.Context
	API *client.Client
}

func New(ctx context.Context) (*DockerClient, error) {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create docker client")
		return nil, err
	}
	return &DockerClient{
		ctx: ctx,
		API: client,
	}, nil
}

func (c *DockerClient) ListContainers() ([]types.Container, error) {
	filters := filters.NewArgs()
	filters.Add("status", "running")
	filters.Add("label", "traefik.enable=true")
	containers, err := c.API.ContainerList(c.ctx, container.ListOptions{
		Filters: filters,
	})
	return containers, err
}

func (c *DockerClient) InspectContainer(containerId string) (types.ContainerJSON, error) {
	return c.API.ContainerInspect(c.ctx, containerId)
}

func (c *DockerClient) Events(options types.EventsOptions) (<-chan events.Message, <-chan error) {
	return c.API.Events(c.ctx, options)
}

func (c *DockerClient) Close() {
	if c.API != nil {
		_ = c.API.Close()
	}
}
