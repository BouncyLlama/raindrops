FROM golang:1.13-alpine
RUN mkdir /app
WORKDIR /app
COPY ./* ./
RUN rm -rf testdata && rm *_test.go && go build -o raindrops && mv raindrops /bin/raindrops
RUN chmod +x docker-entrypoint.sh && mv docker-entrypoint.sh /bin/ && rm -rf ./*
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["raindrops"]
