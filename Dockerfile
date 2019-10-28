FROM golang:1.13-alpine
RUN apk add git
RUN mkdir /raindrops
WORKDIR /raindrops
COPY go.mod .
COPY go.sum .
ENV GO111MODULE=on
RUN go mod download
COPY . ./
RUN rm -rf ./testdata &&  find ./ |grep ".*_test.go" |xargs rm
RUN go build -o raindrops ./cmd  && mv raindrops /bin/raindrops
RUN chmod +x docker-entrypoint.sh && mv docker-entrypoint.sh /bin/ && rm -rf ./*
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["raindrops"]
