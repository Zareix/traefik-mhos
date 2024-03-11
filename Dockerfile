FROM golang:1.22.1-alpine3.19 as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o /app/traefik-mhos


FROM alpine:3.19 as runner

ENV GIN_MODE=release
ENV PORT=8888

COPY --from=builder /app/traefik-mhos /app/traefik-mhos

CMD ["/app/traefik-mhos"]