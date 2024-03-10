package traefik

import (
	"context"
	"fmt"
	"strings"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/log"
	"traefik-multi-hosts/internal/redis"

	docker "github.com/docker/docker/client"
)

func AddContainerToTraefik(ctx context.Context, dockerClient *docker.Client, containerId string) {
	container, err := dockerClient.ContainerInspect(ctx, containerId)
	if err != nil {
		log.Error().Err(err).Str("containerId", containerId).Msg("Failed to inspect container")
	}

	kv := make(map[string]string)

	var serviceName string
	var servicePort string
	var routerRule string
	for labelKey, labelValue := range container.Config.Labels {
		if !strings.HasPrefix(labelKey, "traefik.http.services") && !strings.HasPrefix(labelKey, "traefik.http.routers") && !strings.HasPrefix(labelKey, "traefik.tcp.routers") {
			continue
		}

		if strings.HasPrefix(labelKey, "traefik.http.routers.") && strings.HasSuffix(labelKey, ".rule") {
			routerRule = labelValue
		}

		if strings.HasSuffix(labelKey, "loadbalancer.server.port") {
			servicePort = labelValue // TODO This is not the good port
			log.Debug().Str("servicePort", servicePort).Msg("Found service port")
			continue
		}

		if strings.HasPrefix(labelKey, "traefik.http.routers.") && strings.HasSuffix(labelKey, ".rule") {
			serviceName = strings.Split(labelKey, ".")[3]
			log.Debug().Str("serviceName", serviceName).Msg("Found service name")
		}

		labelKey = strings.ReplaceAll(labelKey, ".", "/")
		labelKey = strings.ReplaceAll(labelKey, "[", "/")
		labelKey = strings.ReplaceAll(labelKey, "]", "")
		kv[labelKey] = labelValue
		log.Debug().Str("key", labelKey).Str("value", labelValue).Msg("Adding key-value pair")
	}

	if serviceName == "" {
		log.Error().Str("containerId", containerId).Msg("Container has no traefik labels")
		return
	}

	if servicePort == "" {
		log.Error().Str("serviceName", serviceName).Msg("Service has no port")
		return
	}

	log.Debug().Str("serviceName", serviceName).Str("servicePort", servicePort).Msg("Adding service to traefik")
	kv["traefik/http/routers/"+serviceName+"/service"] = serviceName
	kv["traefik/http/services/"+serviceName+"/loadbalancer/servers/0/url"] = fmt.Sprintf("http://%s:%s", config.AppConfig.HostIP, servicePort)

	log.Info().Str("serviceName", serviceName).Str("rule", routerRule).Str("target", fmt.Sprintf("http://%s:%s", config.AppConfig.HostIP, servicePort)).Msg("Adding service to traefik")
	redis.SaveService(ctx, serviceName, kv)
}

func RemoveContainerFromTraefik(ctx context.Context, dockerClient *docker.Client, containerId string) {
	container, err := dockerClient.ContainerInspect(ctx, containerId)
	if err != nil {
		panic(err)
	}

	var serviceName string
	for labelKey := range container.Config.Labels {
		if strings.HasPrefix(labelKey, "traefik.http.routers.") && strings.HasSuffix(labelKey, ".rule") {
			serviceName = strings.Split(labelKey, ".")[3]
			break
		}
	}

	if serviceName == "" {
		log.Error().Str("containerId", containerId).Msg("Container has no traefik labels")
		return
	}

	log.Info().Str("serviceName", serviceName).Msg("Removing service from traefik")
	redis.RemoveService(ctx, serviceName)
}
