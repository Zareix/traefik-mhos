# Introduction

Traefik-mhos (Multi HOSts) helps you use a single traefik for proxying multiple docker hosts (withtout swarm or k8s).

Inspired by [jittering/traefik-kop](https://github.com/jittering/traefik-kop).

- [Introduction](#introduction)
  - [Usage](#usage)
  - [Environment Variables](#environment-variables)
  - [Port discovery](#port-discovery)
  - [API/Frontend](#apifrontend)

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

| Name           | Description                                                   | Default                                                  |
| -------------- | ------------------------------------------------------------- | -------------------------------------------------------- |
| REDIS_ADDRESS  | The Redis db address                                          | `localhost:6379`                                         |
| REDIS_PASSWORD | The Redis db password                                         | `<empty string>`                                         |
| REDIS_DB       | The Redis db name                                             | `0`                                                      |
| HOST_IP        | The current host IP (where the traefik routers will point to) | `localhost` (but you should change it to an IP/hostname) |
| LOG_LEVEL      | Minimum log level (debug, info, warn, error, fatal)           | `info`                                                   |
| LISTEN_EVENTS  | Listen to docker events                                       | `true`                                                   |

## Port discovery

Right now, traefik-mhos only reads the `traefik.http.services.<service-name>.loadbalancer.server.port` label.
Be careful, this port needs to point to the host's port, not to the internal container's port (contrary to the traefik version of this label).

If this label is not found, traefik-mhos will default to the first exposed port of the container.

## API/Frontend

Traefik-mhos also exposes a simple API and frontend to list current hosts and their services.

The API and the frontend are available on port `8888` by default. You can change the port by setting the `PORT` environment variable.

Routes:

- `GET /api/heath`: Health check
- `GET /api/hosts`: List all hosts with their services
- `GET /`: Frontend
