package listeners

import (
	"context"
	"traefik-multi-hosts/internal/log"
	"traefik-multi-hosts/internal/traefik"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
)

func ListenForNewContainers(ctx context.Context, dockerClient *docker.Client) {
	eventsFilters := filters.NewArgs()
	eventsFilters.Add("type", "container")
	eventsFilters.Add("event", "start")
	eventsFilters.Add("label", "traefik.enable=true")
	startedContainersEventsStream, errors := dockerClient.Events(ctx, types.EventsOptions{
		Filters: eventsFilters,
	})

	go func() {
		log.Debug().Msg("Listening for new containers")
		for {
			select {
			case event := <-startedContainersEventsStream:
				log.Debug().Interface("action", event.Action).Str("containerId", event.Actor.ID).Msg("New event")
				traefik.AddContainerToTraefik(ctx, dockerClient, event.Actor.ID)
			case err := <-errors:
				if err != nil {
					log.Error().Err(err).Msg("Event error")
				}
			}
		}
	}()
}

func ListenForStoppedContainers(ctx context.Context, dockerClient *docker.Client) {
	eventsFilters := filters.NewArgs()
	eventsFilters.Add("type", "container")
	eventsFilters.Add("event", "stop")
	eventsFilters.Add("label", "traefik.enable=true")
	stoppedContainersEventsStream, errors := dockerClient.Events(ctx, types.EventsOptions{
		Filters: eventsFilters,
	})

	go func() {
		log.Debug().Msg("Listening for stopped containers")
		for {
			select {
			case event := <-stoppedContainersEventsStream:
				log.Debug().Interface("action", event.Action).Str("container", event.Actor.ID).Msg("New event")
				traefik.RemoveContainerFromTraefik(ctx, dockerClient, event.Actor.ID)
			case err := <-errors:
				if err != nil {
					log.Error().Err(err).Msg("Event error")
				}
			}
		}
	}()
}
