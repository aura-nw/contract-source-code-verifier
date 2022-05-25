FROM ubuntu:latest

RUN apt-get update
RUN apt-get install -y wget git gcc
RUN apt update && apt upgrade -y
RUN apt install curl -y

RUN wget -P /tmp https://dl.google.com/go/go1.17.5.linux-amd64.tar.gz

RUN tar -C /usr/local -xzf /tmp/go1.17.5.linux-amd64.tar.gz
RUN rm /tmp/go1.17.5.linux-amd64.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

ARG PORT=8080

RUN mkdir -p /usr/src/app

WORKDIR /usr/src/app

COPY . .

RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs -y | sh
RUN rustup target list --installed
RUN rustup target add wasm32-unknown-unknown

RUN go mod download

RUN swag init

RUN go build

EXPOSE $PORT

CMD [ "go", "run", "main.go" ]