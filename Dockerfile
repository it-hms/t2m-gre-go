# syntax=docker/dockerfile:1

FROM golang:1.22 as build-stage

WORKDIR /work
RUN apt-get update && apt-get -y install libpcap-dev
COPY go.mod go.sum ./
RUN go mod download
COPY dcp ./dcp
COPY gre ./gre
COPY iface ./iface
COPY t2mgre ./t2mgre
COPY *.go ./
RUN CGO_ENABLED=1 GOOS=linux go build -C t2mgre


FROM debian:bookworm-slim as release-stage 

# Install openvpn
RUN apt-get update && \ 
    apt-get -y upgrade &&\
    apt-get install -y libpcap-dev



COPY --from=build-stage /work/t2mgre/t2mgre /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/t2mgre", "-server" ]