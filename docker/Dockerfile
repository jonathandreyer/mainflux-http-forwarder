FROM golang:1.13-alpine AS builder
ARG SVC
ARG GOARCH
ARG GOARM

WORKDIR /go/src/github.com/jonathandreyer/mainflux-http-forwarder
COPY . .
RUN apk update \
    && apk add make\
    && make \
    && mv build/mainflux-$SVC /exe

FROM scratch
COPY --from=builder /exe /
ENTRYPOINT ["/exe"]
