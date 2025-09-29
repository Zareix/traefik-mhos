package docker

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
)

type Client interface {
	InspectContainer(containerId string) (container.InspectResponse, error)
}

type ClientImpl struct {
	ctx context.Context
	API *client.Client
}

func New(ctx context.Context) (*ClientImpl, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create docker dockerClient")
		return nil, err
	}
	return &ClientImpl{
		ctx: ctx,
		API: dockerClient,
	}, nil
}

func (c *ClientImpl) ListContainers() ([]container.Summary, error) {
	filtersArgs := filters.NewArgs()
	filtersArgs.Add("status", "running")
	filtersArgs.Add("label", "traefik.enable=true")
	containers, err := c.API.ContainerList(c.ctx, container.ListOptions{
		Filters: filtersArgs,
	})
	return containers, err
}

func (c *ClientImpl) InspectContainer(containerId string) (container.InspectResponse, error) {
	return c.API.ContainerInspect(c.ctx, containerId)
}

func (c *ClientImpl) Events(options events.ListOptions) (<-chan events.Message, <-chan error) {
	return c.API.Events(c.ctx, options)
}

func (c *ClientImpl) Close() {
	if c.API != nil {
		_ = c.API.Close()
	}
}
