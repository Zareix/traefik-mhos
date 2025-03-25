FROM oven/bun:1.2.6 AS bun_builder

WORKDIR /app

COPY package.json bun.lock ./

RUN bun install

COPY internal/web/static/css/input.css internal/web/templates/ ./

RUN bun run --bun tailwindcss -i input.css -o style.css --minify


FROM golang:1.23.4-alpine3.19 AS go_builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
COPY --from=bun_builder /app/style.css /app/internal/web/static/css/style.css

RUN go build -o /app/traefik-mhos


FROM alpine:3.21 AS runner

ENV GIN_MODE=release
ENV PORT=8888

COPY --from=go_builder /app/traefik-mhos /app/traefik-mhos

HEALTHCHECK CMD wget -qO- http://localhost:$PORT/api/health || exit 1

CMD ["/app/traefik-mhos"]