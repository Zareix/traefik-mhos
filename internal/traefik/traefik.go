package traefik

import (
	"context"
	"fmt"
	"strings"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/redis"

	docker "github.com/docker/docker/client"
)

func AddContainerToTraefik(ctx context.Context, dockerClient *docker.Client, containerId string) {
	container, err := dockerClient.ContainerInspect(ctx, containerId)
	if err != nil {
		panic(err)
	}

	kv := make(map[string]string)

	var serviceName string
	var servicePort string
	for labelKey, labelValue := range container.Config.Labels {
		if !strings.HasPrefix(labelKey, "traefik.http.services") && !strings.HasPrefix(labelKey, "traefik.http.routers") && !strings.HasPrefix(labelKey, "traefik.tcp.routers") {
			continue
		}

		if strings.HasSuffix(labelKey, "loadbalancer.server.port") {
			servicePort = labelValue // TODO This is not the good port
			continue
		}

		if strings.HasPrefix(labelKey, "traefik.http.routers.") && strings.HasSuffix(labelKey, ".rule") {
			serviceName = strings.Split(labelKey, ".")[3]
		}

		labelKey = strings.ReplaceAll(labelKey, ".", "/")
		labelKey = strings.ReplaceAll(labelKey, "[", "/")
		labelKey = strings.ReplaceAll(labelKey, "]", "")
		kv[labelKey] = labelValue
	}

	fmt.Printf("Adding service %s to point to http://%s:%s\n", serviceName, config.AppConfig.HostIP, servicePort)
	kv["traefik/http/routers/"+serviceName+"/service"] = serviceName
	kv["traefik/http/services/"+serviceName+"/loadbalancer/servers/0/url"] = fmt.Sprintf("http://%s:%s", config.AppConfig.HostIP, servicePort)

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

	fmt.Printf("Removing service %s\n", serviceName)
	redis.RemoveService(ctx, serviceName)
}
