FROM golang:1.25.5 AS builder
WORKDIR /app

# cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# copy project
COPY ./pkg ./pkg
COPY "./cmd" "./cmd"
COPY ./internal ./internal

# build
RUN go build -o bot cmd/main.go 

FROM debian:bookworm-20251020-slim
WORKDIR /app

# installing certificates
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# copy from previous stage
COPY --from=builder /app/bot ./bot
COPY ./config ./config
COPY ./config ./data

# creating new user for security
RUN useradd -m docker
USER docker

# expose port and run
ENTRYPOINT ["./bot"]

