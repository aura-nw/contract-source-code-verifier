FROM ubuntu:20.04

SHELL ["/bin/bash", "-c"]

# WORKDIR /root

RUN apt-get update
RUN apt-get install wget git gcc -y
RUN apt update && apt upgrade -y
RUN apt install curl make bash -y

# RUN mkdir -p /etc/apt/keyrings
# RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
# RUN echo \
#   "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
#   $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
# RUN apt-get update
# RUN apt-get install docker-ce docker-ce-cli containerd.io docker-compose-plugin -y

RUN wget -P /tmp https://dl.google.com/go/go1.17.5.linux-amd64.tar.gz

RUN tar -C /usr/local -xzf /tmp/go1.17.5.linux-amd64.tar.gz
RUN rm /tmp/go1.17.5.linux-amd64.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

RUN curl https://sh.rustup.rs -sSf | bash -s -- -y

ENV PATH /root/.cargo/bin:$PATH
RUN rustup target list --installed
RUN rustup target add wasm32-unknown-unknown

ARG PORT=8080

RUN mkdir -p /usr/src/app

WORKDIR /usr/src/app

COPY . .

RUN git clone https://github.com/aura-nw/aura.git
WORKDIR /usr/src/app/aura
RUN make

WORKDIR /usr/src/app

RUN go mod download

RUN go build

EXPOSE $PORT

CMD [ "go", "run", "main.go" ]