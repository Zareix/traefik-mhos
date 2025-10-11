FROM oven/bun:1.3.0 AS bun_builder

WORKDIR /app

COPY package.json bun.lock ./

RUN bun install

COPY internal/web/static/css/input.css internal/web/templates/ ./

RUN bun run --bun tailwindcss -i input.css -o style.css --minify


FROM golang:1.25.2-alpine3.21 AS go_builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
COPY --from=bun_builder /app/style.css /app/internal/web/static/css/style.css

RUN go build -o /app/traefik-mhos


FROM gcr.io/distroless/static-debian12 AS runner

COPY --from=go_builder /app/traefik-mhos /app/traefik-mhos

HEALTHCHECK --interval=30s \
  --timeout=5s \
  CMD ["/app/traefik-mhos", "healthcheck"]

CMD ["/app/traefik-mhos"]