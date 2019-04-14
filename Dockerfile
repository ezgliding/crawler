FROM golang:1.12-stretch AS builder
WORKDIR /build
COPY . /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o crawler .

FROM alpine:latest
WORKDIR /root
RUN apk --no-cache add ca-certificates
COPY --from=builder /build/crawler .
