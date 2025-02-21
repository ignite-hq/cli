# syntax = docker/dockerfile:1.2
# WARNING! Use `DOCKER_BUILDKIT=1` with `docker build` to enable --mount feature.

## prep the base image.
#
FROM golang as base

RUN apt update && \
    apt-get install -y \
    build-essential \
    ca-certificates \
    curl

# enable faster module downloading.
ENV GOPROXY https://proxy.golang.org

## builder stage.
#
FROM base as builder

WORKDIR /ignite

# cache dependencies.
COPY ./go.mod . 
COPY ./go.sum . 
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build go install -v ./...

## prep the final image.
#
FROM base

RUN useradd -ms /bin/bash tendermint
USER tendermint

COPY --from=builder /go/bin/ignite /usr/bin

WORKDIR /apps

# see docs for exposed ports:
#   https://docs.ignite.com/kb/config.html#host
EXPOSE 26657
EXPOSE 26656
EXPOSE 6060 
EXPOSE 9090 
EXPOSE 1317 

ENTRYPOINT ["ignite"]
