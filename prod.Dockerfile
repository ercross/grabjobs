FROM golang:1.17-bullseye

WORKDIR /app

COPY . /app

RUN mkdir "bin"

RUN go build -o ./bin/grabjobs ./cmd/api/*.go

EXPOSE 4045