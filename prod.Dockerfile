FROM golang:1.17-bullseye

WORKDIR /app

RUN mkdir "bin"

COPY . /app

RUN ls -al

EXPOSE 4045