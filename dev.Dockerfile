FROM golang:1.17-bullseye

RUN apt update && apt install make && apt install psmisc

WORKDIR /app

COPY . /app

EXPOSE 4046

CMD ["make", "run_app"]