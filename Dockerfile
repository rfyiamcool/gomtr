FROM golang:alpine as builder

RUN apk add --no-cache make git
WORKDIR /gomtr-src
COPY . /gomtr-src
RUN go mod download && \
    go build . && \
    mv gomtr /gomtr

FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /gomtr /
ENTRYPOINT ["/gomtr"]