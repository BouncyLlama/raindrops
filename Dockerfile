FROM golang:1.13-alpine as builder
RUN apk add git
RUN mkdir /raindrops
WORKDIR /raindrops
COPY go.mod .
COPY go.sum .
ENV GO111MODULE=on
RUN go mod download
COPY . ./
RUN go build -o raindrops ./cmd

FROM alpine:latest
WORKDIR /bin/
COPY docker-entrypoint.sh /bin/docker-entrypoint.sh
RUN chmod +x docker-entrypoint.sh
COPY --from=builder /raindrops/raindrops /bin/raindrops

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["raindrops"]
