package docker

import (
	"context"
	"net/url"
	"traefik-multi-hosts/internal/config"

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
	ctx    context.Context
	API    *client.Client
	HostIp string
}

func NewDefault(ctx context.Context) (*ClientImpl, error) {
	return New(ctx, "")
}

func New(ctx context.Context, dockerHost string) (*ClientImpl, error) {
	var dockerClient *client.Client
	var err error
	if dockerHost != "" {
		dockerClient, err = client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
	} else {
		dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	}
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create docker dockerClient")
		return nil, err
	}

	var hostIp string
	if dockerHost != "" && dockerHost != "unix:///var/run/docker.sock" {
		parsedUrl, err := url.Parse(dockerHost)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to parse docker host URL")
			return nil, err
		}
		hostIp = parsedUrl.Hostname()
	} else {
		hostIp = config.HostIP()
	}
	return &ClientImpl{
		ctx:    ctx,
		API:    dockerClient,
		HostIp: hostIp,
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
