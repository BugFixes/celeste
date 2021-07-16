# Build
FROM golang:1.16 AS builder

RUN mkdir -p /home/builder
WORKDIR /home/builder

## Download dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

## Copy files
COPY . .
ARG build
ARG version

RUN go build -ldflags "-w -s -X github.com/bugfixes/celeste/internal/app.version=$(version) -X github.com/bugfixes/celeste/internal/app.build=$(build)" -race -o ./bin/celeste -v ./cmd/local/local.go
RUN cp ./bin/celeste /

# Runner
FROM alpine as Runner
RUN apk update
RUN apk upgrade
RUN apk add ca-certificates && update-ca-certificates
RUN apk add --update tzdata
RUN rm -rf /var/cache/apk/*

COPY --from=builder /celeste /home

## Set the timezone
ENV TZ=Europe/London

WORKDIR /home

## Set the entrypoint
RUN echo "#!/bin/bash" >> ./entrypoint.sh
RUN echo "./celeste" >> ./entrypoint.sh
RUN chmod +x ./entrypoint.sh
ENTRYPOINT ["sh", "./entrypoint.sh"]

## Healthcheck to keep system alive
HEALTHCHECK --interval=5s --timeout=2s --retries=12 CMD curl --silent --fail localhost/probe || exit 1

## Expose the main ports
EXPOSE 80 3000
