package redis

import (
	"context"
	"fmt"
	"strings"
	"traefik-multi-hosts/internal/config"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Interface pour permettre le mock
type Client interface {
	SaveService(serviceName string, kv map[string]string)
	RemoveService(serviceName string)
	// Ajouter d'autres méthodes si besoin pour les tests
}

type ClientImpl struct {
	ctx context.Context
	API *redis.Client
}

type Service struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

func New(ctx context.Context) (*ClientImpl, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddress(),
		Password: config.RedisPassword(),
		DB:       config.RedisDB(),
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to redis")
	}

	redisClient := &ClientImpl{ctx, client}

	redisClient.CleanCurrentServices()

	return redisClient, nil
}

func (r *ClientImpl) SaveService(serviceName string, kv map[string]string) {
	for key, value := range kv {
		log.Debug().Str("key", key).Str("value", value).Msg("Saving key-value pair")
		r.API.Set(r.ctx, key, value, 0)
	}

	r.API.ZAdd(r.ctx, "mhos:"+config.HostIP(), redis.Z{
		Score:  0,
		Member: serviceName,
	})
}

func (r *ClientImpl) RemoveService(serviceName string) {
	r.API.ZRem(r.ctx, "mhos:"+config.HostIP(), serviceName)

	keys, err := r.scanKeys(fmt.Sprintf("*/%s/*", serviceName))
	if err != nil {
		log.Error().Err(err).Msg("Failed to scan redis keys for service removal")
		return
	}

	for _, key := range keys {
		log.Debug().Str("key", key).Msg("Removing key")
		r.API.Del(r.ctx, key)
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

func (r *ClientImpl) Cleanup() {
	var current []string
	hosts, err := r.scanKeys("mhos:*")
	if err != nil {
		log.Error().Err(err).Msg("Failed to scan redis 'mhos:*' keys")
		return
	}
	for _, key := range hosts {
		hostsCurrent, err := r.API.ZRange(r.ctx, key, 0, -1).Result()
		if err != nil {
			log.Error().Err(err).Str("key", key).Msg("Failed to get members of key")
			return
		}
		current = append(current, hostsCurrent...)
	}

	log.Info().Strs("services", current).Msg("Current services")

	keys, err := r.scanKeys("*")
	if err != nil {
		log.Error().Err(err).Msg("Failed to scan all redis keys")
		return
	}

	for _, key := range keys {
		if !strings.HasPrefix(key, "mhos") && !contains(current, key) {
			r.API.Del(r.ctx, key)
		}
	}
}

func (r *ClientImpl) GetAllHostsWithServices() (map[string][]Service, error) {
	hosts := make(map[string][]Service)
	hostsKeys, err := r.scanKeys("mhos:*")
	if err != nil {
		return nil, err
	}

	allTraefikLabels, err2 := r.getAllLabels()
	if err2 != nil {
		return nil, err2
	}
	for _, hostKey := range hostsKeys {
		hostServices, err := r.API.ZRange(r.ctx, hostKey, 0, -1).Result()
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

func (r *ClientImpl) CleanCurrentServices() {
	r.API.Del(r.ctx, "mhos:"+config.HostIP())
}

func (r *ClientImpl) Close() {
	if r.API != nil {
		_ = r.API.Close()
	}
}

func (r *ClientImpl) getAllLabels() (map[string]string, error) {
	keys, err := r.scanKeys("traefik/*")
	if err != nil {
		return nil, err
	}

	labels := make(map[string]string)
	for _, key := range keys {
		label, err := r.API.Get(r.ctx, key).Result()
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

func (r *ClientImpl) scanKeys(pattern string) ([]string, error) {
	const batchSize = 1000
	allKeys := make([]string, 0, 1000)
	var cursor uint64

	for {
		keys, nextCursor, err := r.API.Scan(r.ctx, cursor, pattern, batchSize).Result()
		if err != nil {
			log.Error().Err(err).Str("pattern", pattern).Msg("Failed to scan keys with pattern")
			return nil, err
		}

		allKeys = append(allKeys, keys...)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	return allKeys, nil
}
