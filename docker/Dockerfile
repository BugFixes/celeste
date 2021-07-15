# Build
FROM golang:1.16 AS builder

RUN mkdir -p /home/builder
WORKDIR /home/builder

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
ARG build
ARG version

RUN go build -ldflags "-w -s -X github.com/bugfixes/celeste/internal/app.version=$(version) -X github.com/bugfixes/celeste/internal/app.build=$(build)" -race -o ./bin/celeste -v ./cmd/local/local.go
RUN cp ./bin/celeste /

# Runner
FROM alpine
RUN apk update
RUN apk upgrade
RUN apk add ca-certificates && update-ca-certificates
RUN apk add --update tzdata
RUN rm -rf /var/cache/apk/*

COPY --from=builder /celeste /home

ENV TZ=Europe/London

WORKDIR /home
ENV _SERVICENAME=celeste
RUN echo "#!/bin/bash" >> ./entrypoint.sh
RUN echo "./celeste" >> ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

ENTRYPOINT ["sh", "./entrypoint.sh"]

HEALTHCHECK --interval=5s --timeout=2s --retries=12 CMD curl --silent --fail localhost/probe || exit 1

EXPOSE 80
