# traefik-mhos

## Introduction

Traefik-mhos (Multi HOSts) helps you use a single traefik for proxying multiple docker hosts (withtout swarm or k8s).

Inspired by [jittering/traefik-kop](https://github.com/jittering/traefik-kop).

## Usage

Create a Redis database (here using docker) along with traefik

```yaml
services:
  db:
    image: redis:latest
    command: redis-server --requirepass password
    ports:
      - '6379:6379'
    volumes:
      - ./redis_data:/data

  traefik:
    image: traefik
    command:
      - --providers.redis.endpoints=db:6379
      - --providers.redis.password=password
    ports:
      - '80:80'
      - '443:443'
```

Then setup traefik-mhos on another docker host:

```yaml
services:
  traefik-mhos:
    image: ghcr.io/zareix/traefik-mhos
    environment:
      - REDIS_ADDRESS=[ip]:6379
      - REDIS_PASSWORD=password
      - REDIS_DB=0
      - HOST_IP=[ip]
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
```
