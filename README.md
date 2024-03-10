# Introduction

Traefik-mhos (Multi HOSts) helps you use a single traefik for proxying multiple docker hosts (withtout swarm or k8s).

Inspired by [jittering/traefik-kop](https://github.com/jittering/traefik-kop).

- [Introduction](#introduction)
  - [Usage](#usage)
  - [Environment Variables](#environment-variables)
  - [Port discovery](#port-discovery)

## Usage

Create a Redis database alongside your Traefik instance and add it as a provider:

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

| You can also add it through Traefik static configuration.

Then setup traefik-mhos on another docker host:

```yaml
services:
  traefik-mhos:
    image: ghcr.io/zareix/traefik-mhos
    environment:
      - REDIS_ADDRESS=${REDIS_ADDRESS}
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - REDIS_DB=${REDIS_DB}
      - HOST_IP=${HOST_IP}
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
```

## Environment Variables

| Name           | Description                                                     | Default          |
| -------------- | --------------------------------------------------------------- | ---------------- |
| REDIS_ADDRESS  | The Redis db address                                            | `localhost:6379` |
| REDIS_PASSWORD | The Redis db password                                           | `""`             |
| REDIS_DB       | The Redis db name                                               | `0`              |
| HOST_IP        | The current host IP (where the traefik routers's will point to) | `"localhost"`    |
| LOG_LEVEL      | Mininum log level (debug, info, warn, error, fatal)             | `info`           |

## Port discovery

Right now, traefik-mhos only reads the `traefik.http.services.<service-name>.loadbalancer.server.port` label.
Be careful, this port needs to point to the host's port, not to the internal container's port (contrary to the traefik version of this label).

If this label is not found, traefik-mhos will default to the first exposed port of the container.
