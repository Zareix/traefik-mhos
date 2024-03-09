package listeners

import (
	"context"
	"fmt"
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
		fmt.Println("Listening for new containers")
		for {
			select {
			case event := <-startedContainersEventsStream:
				fmt.Printf("Event: %s of container %s\n", event.Action, event.Actor.ID[:10])
				traefik.AddContainerToTraefik(ctx, dockerClient, event.Actor.ID)
			case err := <-errors:
				if err != nil {
					fmt.Printf("Error: %s\n", err)
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
		fmt.Println("Listening for stopped containers")
		for {
			select {
			case event := <-stoppedContainersEventsStream:
				fmt.Printf("Event: %s of container %s\n", event.Action, event.Actor.ID[:10])
				traefik.RemoveContainerFromTraefik(ctx, dockerClient, event.Actor.ID)
			case err := <-errors:
				if err != nil {
					fmt.Printf("Error: %s\n", err)
				}
			}
		}
	}()
}
