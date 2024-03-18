package redis

import (
	"context"
	"fmt"
	"strings"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/log"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

var client *redis.Client

func init() {
	ctx := context.Background()

	client = redis.NewClient(&redis.Options{
		Addr:     config.RedisAddress(),
		Password: config.RedisPassword(),
		DB:       config.RedisDB(),
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to redis")
	}

	CleanCurrentServices(ctx)
}

func SaveService(ctx context.Context, serviceName string, kv map[string]string) {
	for key, value := range kv {
		log.Debug().Str("key", key).Str("value", value).Msg("Saving key-value pair")
		client.Set(ctx, key, value, 0)
	}

	client.ZAdd(ctx, "mhos:"+config.HostIP(), redis.Z{
		Score:  0,
		Member: serviceName,
	})
}

func RemoveService(ctx context.Context, serviceName string) {
	client.ZRem(ctx, "mhos:"+config.HostIP(), serviceName)

	keys, err := client.Keys(ctx, "*").Result() // TODO: use scan
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all redis stored keys")
		return
	}

	for _, key := range keys {
		if strings.Contains(key, fmt.Sprintf("/%s/", serviceName)) {
			log.Debug().Str("key", key).Msg("Removing key")
			client.Del(ctx, key)
		}
	}
}

func contains(slice []string, item string) bool {
	for _, i := range slice {
		if strings.Contains(item, fmt.Sprintf("/%s/", i)) {
			return true
		}
	}
	return false
}

func Cleanup(ctx context.Context) {
	var current []string
	hosts, err := client.Keys(ctx, "mhos:*").Result() // TODO: use scan
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all redis mhos:* stored keys")
		return
	}
	for _, key := range hosts {
		hostsCurrent, err := client.ZRange(ctx, key, 0, -1).Result()
		if err != nil {
			log.Error().Err(err).Str("key", key).Msg("Failed to get members of key")
			return
		}
		current = append(current, hostsCurrent...)
	}

	log.Info().Strs("services", current).Msg("Current services")

	keys, err := client.Keys(ctx, "*").Result() // TODO: use scan
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all redis stored keys")
		return
	}

	for _, key := range keys {
		if !strings.HasPrefix(key, "mhos") && !contains(current, key) {
			client.Del(ctx, key)
		}
	}
}

func getAllLabels(ctx context.Context) (map[string]string, error) {
	keys, err := scanKeys(ctx, "traefik/*")
	if err != nil {
		return nil, err
	}

	labels := make(map[string]string)
	for _, key := range keys {
		label, err := client.Get(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		labels[key] = label
	}
	return labels, nil
}

func filterLabelsOfService(allLabels map[string]string, serviceName string) map[string]string {
	serviceLabels := make(map[string]string)
	for key, value := range allLabels {
		if strings.Contains(key, fmt.Sprintf("/%s/", serviceName)) {
			serviceLabels[key] = value
		}
	}
	return serviceLabels
}

func scanKeys(ctx context.Context, pattern string) ([]string, error) {
	var allKeys []string
	var cursor uint64
	for {
		var err error
		var keys []string
		keys, cursor, err = client.Scan(ctx, cursor, pattern, 500).Result()
		if err != nil {
			log.Error().Err(err).Str("pattern", pattern).Msg("Failed to get all scan for pattenr")
			return nil, err
		}
		allKeys = append(allKeys, keys...)
		if cursor == 0 {
			break
		}
	}
	return allKeys, nil
}

func GetAllHostsWithServices(ctx context.Context) (map[string][]Service, error) {
	hosts := make(map[string][]Service)
	hostsKeys, err := scanKeys(ctx, "mhos:*")
	if err != nil {
		return nil, err
	}

	allTraefikLabels, err2 := getAllLabels(ctx)
	if err2 != nil {
		return nil, err2
	}
	for _, hostKey := range hostsKeys {
		hostServices, err := client.ZRange(ctx, hostKey, 0, -1).Result()
		if err != nil {
			return nil, err
		}
		hostKey = strings.TrimPrefix(hostKey, "mhos:")
		for _, serviceName := range hostServices {
			var service Service
			service.Name = serviceName
			service.Labels = filterLabelsOfService(allTraefikLabels, serviceName)
			hosts[hostKey] = append(hosts[hostKey], service)
		}
	}
	return hosts, nil
}

func CleanCurrentServices(ctx context.Context) {
	client.Del(ctx, "mhos:"+config.HostIP())
}
