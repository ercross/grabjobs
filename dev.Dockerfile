FROM golang:1.17-bullseye

RUN apt update && apt install make && apt install psmisc

WORKDIR /app

COPY . /app

EXPOSE 4045

CMD ["make", "run_app"]