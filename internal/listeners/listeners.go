package listeners

import (
	"context"
	"sync"
	"time"
	"traefik-multi-hosts/internal/docker"
	"traefik-multi-hosts/internal/log"
	"traefik-multi-hosts/internal/traefik"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
)

func ProcessEvents(ctx context.Context, eventType events.Action, containerId string) {
	switch eventType {
	case events.ActionStart:
		log.Debug().Str("containerId", containerId).Msg("Container started")
		traefik.AddContainerToTraefik(ctx, containerId)
	case events.ActionStop:
		log.Debug().Str("containerId", containerId).Msg("Container stopped")
		traefik.RemoveContainerFromTraefik(ctx, containerId)
	}
	time.Sleep(2 * time.Second)
}

func ListenForContainersEvent(ctx context.Context) {
	eventsFilters := filters.NewArgs()
	eventsFilters.Add("type", "container")
	eventsFilters.Add("event", "stop")
	eventsFilters.Add("event", "start")
	eventsFilters.Add("label", "traefik.enable=true")
	startedContainersEventsStream, errors := docker.Events(ctx, types.EventsOptions{
		Filters: eventsFilters,
	})

	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)

	go func() {
		log.Debug().Msg("Listening for new containers")
		for {
			select {
			case event := <-startedContainersEventsStream:
				log.Debug().Interface("action", event.Action).Str("containerId", event.Actor.ID).Msg("New event")
				sem <- struct{}{}
				wg.Add(1)
				go func() {
					defer func() {
						<-sem
						wg.Done()
					}()
					ProcessEvents(ctx, event.Action, event.Actor.ID)
				}()
			case err := <-errors:
				log.Error().Err(err).Msg("Event error")
			case <-ctx.Done():
				log.Debug().Msg("Context cancelled, waiting for processing to complete...")
				wg.Wait()
				return
			}
		}
	}()
}
