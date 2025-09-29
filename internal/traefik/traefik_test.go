package traefik

import (
	"testing"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/docker"
	"traefik-multi-hosts/internal/redis"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

func TestGetFirstExposedPort(t *testing.T) {
	inspectResponse := container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			HostConfig: &container.HostConfig{
				PortBindings: nat.PortMap{
					"80/tcp": {
						{
							HostPort: "8080",
						},
					},
				},
			},
		},
	}

	port := GetFirstExposedPort(inspectResponse)

	if port != "8080" {
		t.Errorf("Expected port 8080, got %s", port)
	}
}

func TestGetFirstExposedPortMultiple(t *testing.T) {
	inspectResponse := container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			HostConfig: &container.HostConfig{
				PortBindings: nat.PortMap{
					"80/tcp": {
						{
							HostPort: "8080",
						},
					},
					"443/tcp": {
						{
							HostPort: "8443",
						},
					},
				},
			},
		},
	}

	port := GetFirstExposedPort(inspectResponse)

	if port != "8080" {
		t.Errorf("Expected port 8080, got %s", port)
	}
}

type DockerClientMock struct {
	InspectContainerCallCount int
}

func (m *DockerClientMock) InspectContainer(string) (container.InspectResponse, error) {
	m.InspectContainerCallCount++
	return container.InspectResponse{
		Config: &container.Config{
			Labels: map[string]string{
				"traefik.http.routers.test.rule":                      "Host(`test.local`)",
				"traefik.http.services.test.loadbalancer.server.port": "8080",
			},
		},
		ContainerJSONBase: &container.ContainerJSONBase{
			Name: "test-container",
			HostConfig: &container.HostConfig{
				PortBindings: nat.PortMap{
					"8080/tcp": {
						{
							HostPort: "8080",
						},
					},
				},
			},
		},
	}, nil
}

func (m *DockerClientMock) GetInspectContainerCallCount() int {
	return m.InspectContainerCallCount
}

type RedisClientMock struct {
	SaveServiceCallCount   int
	RemoveServiceCallCount int

	SaveServiceNames []string
	SaveServiceKVs   []map[string]string
}

func (m *RedisClientMock) SaveService(serviceName string, kv map[string]string) {
	m.SaveServiceCallCount++
	m.SaveServiceNames = append(m.SaveServiceNames, serviceName)
	// Copier la map pour éviter les effets de bord
	kvCopy := make(map[string]string, len(kv))
	for k, v := range kv {
		kvCopy[k] = v
	}
	m.SaveServiceKVs = append(m.SaveServiceKVs, kvCopy)
}

func (m *RedisClientMock) RemoveService(string) {
	m.RemoveServiceCallCount++
}

func (m *RedisClientMock) GetSaveServiceCallCount() int {
	return m.SaveServiceCallCount
}

func (m *RedisClientMock) GetRemoveServiceCallCount() int {
	return m.RemoveServiceCallCount
}

func TestAddContainerToTraefik(t *testing.T) {
	config.Init()
	dockerClientMock := &DockerClientMock{}
	redisClientMock := &RedisClientMock{}
	var dockerClient docker.Client = dockerClientMock
	var redisClient redis.Client = redisClientMock
	containerId := "test-container-id"

	err := AddContainerToTraefik(dockerClient, redisClient, containerId)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if dockerClientMock.GetInspectContainerCallCount() != 1 {
		t.Errorf("Expected InspectContainer to be called once, got %d", dockerClientMock.GetInspectContainerCallCount())
	}
	if redisClientMock.GetSaveServiceCallCount() != 1 {
		t.Errorf("Expected SaveService to be called once, got %d", redisClientMock.GetSaveServiceCallCount())
	}
	if len(redisClientMock.SaveServiceNames) != 1 || redisClientMock.SaveServiceNames[0] != "test" {
		t.Errorf("Expected SaveService to be called with service name 'test', got %v", redisClientMock.SaveServiceNames)
	}
	if len(redisClientMock.SaveServiceKVs) != 1 {
		t.Errorf("Expected SaveService to be called with one set of key-value pairs, got %d", len(redisClientMock.SaveServiceKVs))
	}
	if redisClientMock.SaveServiceKVs[0]["traefik/http/routers/test/service"] != "test" {
		t.Errorf("Expected key 'traefik/http/routers/test/service' to be 'test', got %s", redisClientMock.SaveServiceKVs[0]["traefik/http/routers/test/service"])
	}
}
