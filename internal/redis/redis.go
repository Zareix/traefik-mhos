package redis

import (
	"context"
	"fmt"
	"strings"
	"traefik-multi-hosts/internal/config"

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
		panic(err)
	}

	return client
}

func SaveService(ctx context.Context, serviceName string, kv map[string]string) {
	for key, value := range kv {
		client.Set(ctx, key, value, 0)
	}

	client.SAdd(ctx, config.AppConfig.HostIP, serviceName)
}

func RemoveService(ctx context.Context, serviceName string) {
	client.SRem(ctx, config.AppConfig.HostIP, serviceName)

	keys, err := client.Keys(ctx, "*").Result()
	if err != nil {
		panic(err)
	}

	for _, key := range keys {
		if strings.Contains(key, fmt.Sprintf("/%s/", serviceName)) {
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
	current, err := client.SMembers(ctx, config.AppConfig.HostIP).Result()
	if err != nil {
		panic(err)
	}

	fmt.Println("Current services:", current)

	keys, err := client.Keys(ctx, "*").Result()
	if err != nil {
		panic(err)
	}

	for _, key := range keys {
		if key != config.AppConfig.HostIP && !contains(current, key) {
			client.Del(ctx, key)
		}
	}
}
