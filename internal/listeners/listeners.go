package listeners

import (
	"context"
	"sync"
	"time"
	"traefik-multi-hosts/internal/docker"
	"traefik-multi-hosts/internal/redis"
	"traefik-multi-hosts/internal/traefik"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/rs/zerolog/log"
)

func ProcessEvents(dockerClient docker.DockerClient, redisClient redis.RedisClient, eventType events.Action, containerId string) {
	switch eventType {
	case events.ActionStart:
		log.Debug().Str("containerId", containerId).Msg("Container started")
		traefik.AddContainerToTraefik(dockerClient, redisClient, containerId)
	case events.ActionStop:
		log.Debug().Str("containerId", containerId).Msg("Container stopped")
		traefik.RemoveContainerFromTraefik(dockerClient, redisClient, containerId)
	}
	time.Sleep(2 * time.Second)
}

func ListenForContainersEvent(ctx context.Context, dockerClient docker.DockerClient, redisClient redis.RedisClient) {
	for {
		err := listenForContainersEvent(ctx, dockerClient, redisClient)
		if err != nil {
			log.Error().Err(err).Msg("Error listening for containers event")
			continue
		}
		return
	}
}

func listenForContainersEvent(ctx context.Context, dockerClient docker.DockerClient, redisClient redis.RedisClient) error {
	eventsFilters := filters.NewArgs()
	eventsFilters.Add("type", "container")
	eventsFilters.Add("event", "stop")
	eventsFilters.Add("event", "start")
	eventsFilters.Add("label", "traefik.enable=true")
	containersEventsStream, errors := dockerClient.Events(types.EventsOptions{
		Filters: eventsFilters,
	})

	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)

	log.Debug().Msg("Listening for new containers")
	for {
		select {
		case event := <-containersEventsStream:
			log.Debug().Interface("action", event.Action).Str("containerId", event.Actor.ID).Msg("New event")
			sem <- struct{}{}
			wg.Add(1)
			go func() {
				defer func() {
					<-sem
					wg.Done()
				}()
				ProcessEvents(dockerClient, redisClient, event.Action, event.Actor.ID)
			}()
		case err := <-errors:
			return err
		case <-ctx.Done():
			log.Debug().Msg("Context cancelled, waiting for processing to complete...")
			wg.Wait()
			return nil
		}
	}
}
