FROM golang:1.26 AS build-env

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o api ./cmd/agentapi


# target image
FROM debian:12-slim

# Install curl and install/updates certificates
RUN apt-get update \
    && apt-get install -y -q --no-install-recommends \
    ca-certificates \
    curl \
    && apt-get clean

COPY --from=build-env /app/api /usr/bin/api

EXPOSE 80

HEALTHCHECK CMD curl -f http://localhost/status || exit 1

CMD ["api"]
