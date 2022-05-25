FROM golang:1.16-alpine as build-stage

ARG PORT=8080

RUN mkdir -p /usr/src/app

WORKDIR /usr/src/app

COPY . .

RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
RUN rustup target list --installed
RUN rustup target add wasm32-unknown-unknown

RUN go mod download

RUN swag init

RUN go build

EXPOSE $PORT

CMD [ "go", "run", "main.go" ]