FROM golang:1.21.5-alpine as base

FROM base as dev

RUN mkdir /app
WORKDIR /app
COPY . /app

RUN go install github.com/cosmtrek/air@latest

CMD ["air", "-c", ".air.toml"]