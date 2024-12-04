FROM golang:1.23.4-alpine3.19 as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o /app/traefik-mhos


FROM alpine:3.20 as runner

ENV GIN_MODE=release
ENV PORT=8888

COPY --from=builder /app/traefik-mhos /app/traefik-mhos

HEALTHCHECK CMD wget -qO- http://localhost:$PORT/api/health || exit 1

CMD ["/app/traefik-mhos"]