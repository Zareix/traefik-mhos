package redis

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/log"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var client *redis.Client

func NewClient(address string, password string, db int) *redis.Client {
	client = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to redis")
	}

	return client
}

func SaveService(ctx context.Context, serviceName string, kv map[string]string) {
	for key, value := range kv {
		log.Debug().Str("key", key).Str("value", value).Msg("Saving key-value pair")
		client.Set(ctx, key, value, 0)
	}

	client.SAdd(ctx, "mhos:"+config.AppConfig.HostIP, serviceName)
}

func RemoveService(ctx context.Context, serviceName string) {
	client.SRem(ctx, config.AppConfig.HostIP, serviceName)

	keys, err := client.Keys(ctx, "*").Result()
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
	hosts, err := client.Keys(ctx, "mhos:*").Result()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all redis mhos:* stored keys")
		return
	}
	for _, key := range hosts {
		hostsCurrent, err := client.SMembers(ctx, key).Result()
		if err != nil {
			log.Error().Err(err).Str("key", key).Msg("Failed to get members of key")
			return
		}
		current = append(current, hostsCurrent...)
	}

	sort.Strings(current)

	log.Info().Strs("services", current).Msg("Current services")

	keys, err := client.Keys(ctx, "*").Result()
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
