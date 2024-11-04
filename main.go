package main

import (
	"context"
	"log"
	"traefik-multi-hosts/cmd/mhos"
	"traefik-multi-hosts/internal/docker"
	"traefik-multi-hosts/internal/redis"
	"traefik-multi-hosts/internal/web"
)

func main() {
	ctx := context.Background()

	dockerClient, err := docker.New(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer dockerClient.Close()

	redisClient, err := redis.New(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer redisClient.Close()

	go func() {
		web.Serve(*dockerClient, *redisClient)
	}()
	go func() {
		mhos.Run(ctx, *dockerClient, *redisClient)
	}()

	select {}
}
