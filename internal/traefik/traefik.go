package traefik

import (
	"errors"
	"fmt"
	"strings"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/docker"
	"traefik-multi-hosts/internal/redis"

	"github.com/docker/docker/api/types"
	"github.com/rs/zerolog/log"
)

func GetFirstExposedPort(container types.ContainerJSON) string {
	for _, port := range container.HostConfig.PortBindings {
		log.Debug().Str("port", port[0].HostPort).Str("container", container.Name).Msg("Found exposed port")
		return port[0].HostPort
	}
	return ""
}

func AddContainerToTraefik(dockerClient docker.DockerClient, redisClient redis.RedisClient, containerId string) error {
	container, err := dockerClient.InspectContainer(containerId)
	if err != nil {
		log.Error().Err(err).Str("containerId", containerId).Msg("Failed to inspect container")
		return err
	}

	log.Debug().Str("id", containerId).Str("name", container.Name).Msg("Adding container to traefik")

	kv := make(map[string]string)
	var serviceName string
	servicePort := GetFirstExposedPort(container)
	var routerRule string
	for labelKey, labelValue := range container.Config.Labels {
		if !strings.HasPrefix(labelKey, "traefik.http.services") && !strings.HasPrefix(labelKey, "traefik.http.routers") && !strings.HasPrefix(labelKey, "traefik.tcp.routers") {
			continue
		} else if serviceName == "" {
			serviceName = strings.Split(labelKey, ".")[3]
			log.Debug().Str("serviceName", serviceName).Msg("Found service name")
		}

		if strings.HasPrefix(labelKey, "traefik.http.routers.") && strings.HasSuffix(labelKey, ".rule") {
			routerRule = labelValue
		}

		if strings.HasSuffix(labelKey, "loadbalancer.server.port") {
			servicePort = labelValue
			log.Debug().Str("servicePort", servicePort).Msg("Found service port")
			continue
		}

		labelKey = strings.ReplaceAll(labelKey, ".", "/")
		labelKey = strings.ReplaceAll(labelKey, "[", "/")
		labelKey = strings.ReplaceAll(labelKey, "]", "")
		kv[labelKey] = labelValue
		log.Debug().Str("key", labelKey).Str("value", labelValue).Msg("Adding key-value pair")
	}

	if serviceName == "" {
		err := errors.New("Container has no traefik labels, id: " + containerId)
		log.Error().Str("containerId", containerId).Msg("Container has no traefik labels")
		return err
	}

	if servicePort == "" {
		err := errors.New("Service has no port: " + serviceName)
		log.Error().Str("serviceName", serviceName).Msg("Service has no port")
		return err
	}

	log.Debug().Str("serviceName", serviceName).Str("servicePort", servicePort).Msg("Adding service to traefik")
	kv["traefik/http/routers/"+serviceName+"/service"] = serviceName
	kv["traefik/http/services/"+serviceName+"/loadbalancer/servers/0/url"] = fmt.Sprintf("http://%s:%s", config.HostIP(), servicePort)

	log.Info().Str("serviceName", serviceName).Str("rule", routerRule).Str("target", fmt.Sprintf("http://%s:%s", config.HostIP(), servicePort)).Msg("Adding service to traefik")
	redisClient.SaveService(serviceName, kv)
	return nil
}

func RemoveContainerFromTraefik(dockerClient docker.DockerClient, redisClient redis.RedisClient, containerId string) {
	container, err := dockerClient.InspectContainer(containerId)
	if err != nil {
		log.Error().Err(err).Str("containerId", containerId).Msg("Failed to inspect container")
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
	redisClient.RemoveService(serviceName)
}
