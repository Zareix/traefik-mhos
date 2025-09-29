package listeners

import (
	"context"
	"time"
	"traefik-multi-hosts/internal/docker"
	"traefik-multi-hosts/internal/redis"
	"traefik-multi-hosts/internal/traefik"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/rs/zerolog/log"
)

func ListenForContainersEvent(ctx context.Context, dockerClient *docker.ClientImpl, redisClient *redis.ClientImpl) {
	retries := 0
	for {
		err := listenForContainersEvent(ctx, dockerClient, redisClient)
		if err != nil {
			if retries > 3 {
				log.Fatal().Err(err).Msg("Failed to listen for containers event")
				panic(err)
			}
			retries++
			log.Error().Err(err).Msg("Could not listen for containers event, retrying")
			time.Sleep(time.Second * 5)
			continue
		}
		return
	}
}

func listenForContainersEvent(ctx context.Context, dockerClient *docker.ClientImpl, redisClient *redis.ClientImpl) error {
	eventsFilters := filters.NewArgs()
	eventsFilters.Add("type", "container")
	eventsFilters.Add("event", "stop")
	eventsFilters.Add("event", "start")
	eventsFilters.Add("label", "traefik.enable=true")
	containersEventsStream, errors := dockerClient.Events(events.ListOptions{
		Filters: eventsFilters,
	})

	log.Debug().Msg("Listening for new containers")
	for {
		select {
		case event := <-containersEventsStream:
			log.Debug().Interface("action", event.Action).Str("containerId", event.Actor.ID).Msg("New event")
			switch event.Action {
			case events.ActionStart:
				log.Debug().Str("containerId", event.Actor.ID).Msg("Container started")
				_ = traefik.AddContainerToTraefik(dockerClient, redisClient, event.Actor.ID)
			case events.ActionStop:
				log.Debug().Str("containerId", event.Actor.ID).Msg("Container stopped")
				traefik.RemoveContainerFromTraefik(dockerClient, redisClient, event.Actor.ID)
			}
		case err := <-errors:
			return err
		case <-ctx.Done():
			log.Debug().Msg("Context cancelled")
			return nil
		}
	}
}
